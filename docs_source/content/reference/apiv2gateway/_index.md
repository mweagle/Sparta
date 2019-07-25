---
date: 2016-03-09T19:56:50+01:00
pre: "<b>API V2 Gateway</b>"
alwaysopen: false
weight: 100
---

# API V2 Gateway

The API V2 Gateway service provides a way to expose a WebSocket API that is 
supported by a set of Lambda functions. The AWS [blog post](https://aws.amazon.com/blogs/compute/announcing-websocket-apis-in-amazon-api-gateway/) supplies an excellent overview of
the pros and cons of this approach that enables a near real time, pushed-based
application. This section will provide an overview of how to configure a WebSocket
API using Sparta. It is based on the [SpartaWebSocket](https://github.com/mweagle/SpartaWebSocket) sample project.

## Payload

Similar to the AWS blog post, our WebSocket API will transmit messages of the form

```json
{
    "message":"sendmessage",
    "data":"hello world !"
}
```

We'll use [ws](https://github.com/hashrocket/ws) to test the API from the command line.

## Routes

The Sparta service consists of three lambda functions:

* `connectWorld(context.Context, awsEvents.APIGatewayWebsocketProxyRequest) (*wsResponse, error)`
* `disconnectWorld(context.Context, awsEvents.APIGatewayWebsocketProxyRequest) (*wsResponse, error)`
* `sendMessage(context.Context, awsEvents.APIGatewayWebsocketProxyRequest) (*wsResponse, error)`

Our functions will use the __PROXY__ style integration and therefore accept an instance of the [APIGatewayWebsocketProxyRequest](https://godoc.org/github.com/aws/aws-lambda-gevents#APIGatewayWebsocketProxyRequest) 

Each function returns a `*wsResponse` instance that satisfies the __PROXY__ mapping:

```go
type wsResponse struct {
  StatusCode int    `json:"statusCode"`
  Body       string `json:"body"`
}
```

### connectWorld

The `connectWorld` AWS Lambda function is responsible for saving the incoming _connectionID_ into a dynamically provisioned DynamoDB database so that subsequent _sendMessage_ requests can broadcast to all subscribed parties.

The table name is advertised in the Lambda function via a user-defined environment variable. The specifics of how that table is provisioned will be addressed in a section below.

```go
...
// Operation
putItemInput := &dynamodb.PutItemInput{
  TableName: aws.String(os.Getenv(envKeyTableName)),
  Item: map[string]*dynamodb.AttributeValue{
    ddbAttributeConnectionID: &dynamodb.AttributeValue{
      S: aws.String(request.RequestContext.ConnectionID),
     },
  },
}
_, putItemErr := dynamoClient.PutItem(putItemInput)
...
```

### disconnectWorld

The complement to `connectWorld` is `disconnectWorld` which is responsible for removing the _connectionID_ from the list of registered connections:

```go
  delItemInput := &dynamodb.DeleteItemInput{
    TableName: aws.String(os.Getenv(envKeyTableName)),
    Key: map[string]*dynamodb.AttributeValue{
      ddbAttributeConnectionID: &dynamodb.AttributeValue{
        S: aws.String(connectionID),
      },
    },
  }
  _, delItemErr := ddbService.DeleteItem(delItemInput)
```

### sendMessage

With the `connectWorld` and `disconnectWorld` connection management functions created, the core of the WebSocket API is `sendMessage`. This function is responsible for scanning over the set of registered _connectionIDs_ and forwarding a request to [PostConnectionWithContext](https://godoc.org/github.com/aws/aws-sdk-go/service/apigatewaymanagementapi#ApiGatewayManagementApi.PostToConnectionWithContext). This function sends the message to the registered connections.

The `sendMessage` function has a few distinct sections that can be understood as follows.

The first requirement is to setup the API Gateway Management service instance using the proper endpoint:

```go
  endpointURL := fmt.Sprintf("%s/%s",
    request.RequestContext.DomainName,
    request.RequestContext.Stage)
  logger.WithField("Endpoint", endpointURL).Info("API Gateway Endpoint")
  dynamoClient := dynamodb.New(sess)
    apigwMgmtClient := apigwManagement.New(sess, aws.NewConfig().WithEndpoint(endpointURL))
```

The new step is to unmarshal and validate the incoming JSON request body:

```go
  // Get the input request...
  var objMap map[string]*json.RawMessage
  unmarshalErr := json.Unmarshal([]byte(request.Body), &objMap)
  if unmarshalErr != nil || objMap["data"] == nil {
    return &wsResponse{
      StatusCode: 500,
      Body:       "Failed to unmarshal request: " + unmarshalErr.Error(),
    }, nil
  }
```

Once the incoming `data` property is validated, the next step is to scan the DynamoDB table for the registered connections and post a message to each one. Note that the scan callback also attempts to cleanup connections that are no longer valid, but which haven't been cleanly removed.

```go
  scanCallback := func(output *dynamodb.ScanOutput, lastPage bool) bool {
    // Send the message to all the clients
    for _, eachItem := range output.Items {
      receiverConnection := ""
      if eachItem[ddbAttributeConnectionID].S != nil {
        receiverConnection = *eachItem[ddbAttributeConnectionID].S
      }
      postConnectionInput := &apigwManagement.PostToConnectionInput{
        ConnectionId: aws.String(receiverConnection),
        Data:         *objMap["data"],
      }
      _, respErr := apigwMgmtClient.PostToConnectionWithContext(ctx, postConnectionInput)
      if respErr != nil {
        if receiverConnection != "" &&
          strings.Contains(respErr.Error(), apigwManagement.ErrCodeGoneException) {
          // Async clean it up...
          go deleteConnection(receiverConnection, dynamoClient)
        } else {
          logger.WithField("Error", respErr).Warn("Failed to post to connection")
        }
      }
      return true
    }
    return true
  }

  // Scan the connections table
  scanInput := &dynamodb.ScanInput{
    TableName: aws.String(os.Getenv(envKeyTableName)),
  }
  scanItemErr := dynamoClient.ScanPagesWithContext(ctx,
    scanInput,
    scanCallback)
    ...
```

These three functions are the core of the WebSocket service.

## API V2 Gateway Decorator

The next step is to create the API V2 API which is comprised of:

* [Stage](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-apigatewayv2-stage.html)
* [API](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-apigatewayv2-api.html)
* [Routes](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-apigatewayv2-route.html)

There is one Stage and one API per service, but a given service (including this one) may include multiple Routes:

```go
// APIv2 Websockets
stage, _ := sparta.NewAPIV2Stage("v1")
stage.Description = "New deploy!"

apiGateway, _ := sparta.NewAPIV2(sparta.Websocket,
  "sample",
  "$request.body.message",
   stage)
```

The `NewAPIV2` creation function requires:

* The protocol to use (`sparta.Websocket`)
* The name of the API (`sample`)
* The [route selection expression](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-websocket-api-selection-expressions.html#apigateway-websocket-api-route-selection-expressions) that represents a JSONPath selection expression to map input data to the corresponding lambda function.
* The stage

Once the API is defined, each route is associated with the API as in:

```go
apiv2ConnectRoute, _ := apiGateway.NewAPIV2Route("$connect",
    lambdaConnect)
apiv2ConnectRoute.OperationName = "ConnectRoute"
...
apiv2SendRoute, _ := apiGateway.NewAPIV2Route("sendmessage",
    lambdaSend)
apiv2SendRoute.OperationName = "SendRoute"
...
```

The `$connect` routeKey is a special [route key value](https://aws.amazon.com/blogs/compute/announcing-websocket-apis-in-amazon-api-gateway/) that is sent when a client first connects to the WebSocket API. 

The `sendmessage` routeKey value means that a payload of the form:

```json
{
    "message":"sendmessage",
    "data":"hello world !"
}
```

will trigger the `lambdaSend` function given the parent API's route selection expression of `$request.body.message`.

### Additional Privileges

Because the `lambdaSend` function also needs to invoke the API Gateway Management APIs an additional IAM Privilege must be enabled:

```go
  var apigwPermissions = []sparta.IAMRolePrivilege{
    {
      Actions: []string{"execute-api:ManageConnections"},
      Resource: gocf.Join("",
        gocf.String("arn:aws:execute-api:"),
        gocf.Ref("AWS::Region"),
        gocf.String(":"),
        gocf.Ref("AWS::AccountId"),
        gocf.String(":"),
        gocf.Ref(apiGateway.LogicalResourceName()),
        gocf.String("/*")),
    },
  }
  lambdaSend.RoleDefinition.Privileges = append(lambdaSend.RoleDefinition.Privileges, apigwPermissions...)
```

## Annotating Lambda Functions

The final configuration step is to use the API gateway to create an instance of the `APIV2GatewayDecorator`. This decorator is responsible for:

* Provisioning the DynamoDB table.
* Ensuring DynamoDB "CRUD" permissions for all the AWS Lambda functions.
* Publishing the table name into the Lambda function's Environment block.
* Adding the WebSocket `wss://...` URL to the Stack's Outputs.

```go
  decorator, _ := apiGateway.NewConnectionTableDecorator(envKeyTableName /* ENV key to use for DDB table name*/,
    ddbAttributeConnectionID /* DDB attr name for connectionID */,
    5 /* readCapacity */,
    5 /* writeCapacity */)
  var lambdaFunctions []*sparta.LambdaAWSInfo
  lambdaFunctions = append(lambdaFunctions,
    lambdaConnect,
    lambdaDisconnect,
    lambdaSend)
  decorator.AnnotateLambdas(lambdaFunctions)
```

## Provision

With everything defined, provide the API V2 Decorator as a Workflow hook as in:

```go
// Set everything up and run it...
  workflowHooks := &sparta.WorkflowHooks{
    ServiceDecorators: []sparta.ServiceDecoratorHookHandler{decorator},
  }
  err := sparta.MainEx(awsName,
    "Sparta application that demonstrates API v2 Websocket support",
    lambdaFunctions,
    apiGateway,
    nil,
    workflowHooks,
    false)
```

and then provision the application:

```bash
go run main.go provision --s3Bucket $S3_BUCKET --noop
INFO[0000] ════════════════════════════════════════════════
INFO[0000] ╔═╗╔═╗╔═╗╦═╗╔╦╗╔═╗   Version : 1.9.4
INFO[0000] ╚═╗╠═╝╠═╣╠╦╝ ║ ╠═╣   SHA     : cfd44e2
INFO[0000] ╚═╝╩  ╩ ╩╩╚═ ╩ ╩ ╩   Go      : go1.12.6
INFO[0000] ════════════════════════════════════════════════
INFO[0000] Service: SpartaWebSocket-123412341234         LinkFlags= Option=provision UTC="2019-07-25T05:26:57Z"
INFO[0000] ════════════════════════════════════════════════
INFO[0000] Using `git` SHA for StampedBuildID            Command="git rev-parse HEAD" SHA=6b26f8e645e9d58c1b678e46576e19bbc29886c0
INFO[0000] Provisioning service                          BuildID=6b26f8e645e9d58c1b678e46576e19bbc29886c0 CodePipelineTrigger= InPlaceUpdates=false NOOP=false Tags=
INFO[0000] Verifying IAM Lambda execution roles
INFO[0000] IAM roles verified                            Count=3
INFO[0000] Checking S3 versioning                        Bucket=weagle VersioningEnabled=true
INFO[0000] Checking S3 region                            Bucket=weagle Region=us-west-2
INFO[0000] Running `go generate`
INFO[0000] Compiling binary                              Name=Sparta.lambda.amd64
INFO[0002] Creating code ZIP archive for upload          TempName=./.sparta/SpartaWebSocket_123412341234-code.zip
INFO[0002] Lambda code archive size                      Size="23 MB"
INFO[0002] Uploading local file to S3                    Bucket=weagle Key=SpartaWebSocket-123412341234/SpartaWebSocket_123412341234-code.zip Path=./.sparta/SpartaWebSocket_123412341234-code.zip Size="23 MB"
INFO[0011] Calling WorkflowHook                          ServiceDecoratorHook= WorkflowHookContext="map[]"
INFO[0011] Uploading local file to S3                    Bucket=weagle Key=SpartaWebSocket-123412341234/SpartaWebSocket_123412341234-cftemplate.json Path=./.sparta/SpartaWebSocket_123412341234-cftemplate.json Size="14 kB"
INFO[0011] Creating stack                                StackID="arn:aws:cloudformation:us-west-2:123412341234:stack/SpartaWebSocket-123412341234/d8a405b0-ae9c-11e9-a05a-0a1528792fce"
INFO[0122] CloudFormation Metrics ▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬
...
INFO[0122] Stack Outputs ▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬
INFO[0122]     APIGatewayURL                             Description="API Gateway Websocket URL" Value="wss://gu4vmnia27.execute-api.us-west-2.amazonaws.com/v1"
INFO[0122] Stack provisioned                             CreationTime="2019-07-25 05:27:08.687 +0000 UTC" StackId="arn:aws:cloudformation:us-west-2:123412341234:stack/SpartaWebSocket-123412341234/d8a405b0-ae9c-11e9-a05a-0a1528792fce" StackName=SpartaWebSocket-123412341234
INFO[0122] ════════════════════════════════════════════════
INFO[0122] SpartaWebSocket-123412341234 Summary
INFO[0122] ════════════════════════════════════════════════
INFO[0122] Verifying IAM roles                           Duration (s)=0
INFO[0122] Verifying AWS preconditions                   Duration (s)=0
INFO[0122] Creating code bundle                          Duration (s)=1
INFO[0122] Uploading code                                Duration (s)=9
INFO[0122] Ensuring CloudFormation stack                 Duration (s)=112
INFO[0122] Total elapsed time                            Duration (s)=122
```

## Test

With the API Gateway deployed, the last step is to test it. Download and install the [ws](go get -u github.com/hashrocket/ws
) tool:

```bash
go get -u github.com/hashrocket/ws
```

then connect to your new API and send a message as in:

```bash
22:31 $ ws wss://gu4vmnia27.execute-api.us-west-2.amazonaws.com/v1
> {"message":"sendmessage", "data":"hello world !"}
< "hello world !"
```

You can also send messages with [Firecamp](https://chrome.google.com/webstore/detail/firecamp-a-campsite-for-d/eajaahbjpnhghjcdaclbkeamlkepinbl?hl=en), a Chrome extension, and send messages between your `ws` session and the web (or vice versa).

## Conclusion

While a production ready application would likely need to include authentication and authorization, this is the beginnings of a full featured WebSocket service.

Remember to terminate the stack when you're done to avoid any unintentional costs.


