---
date: 2016-03-09T19:56:50+01:00
title: Echo Event
weight: 10
---

To start, we'll create a HTTPS accessible lambda function that simply echoes back the contents of the Lambda event.  The source for this is the [SpartaHTML](https://github.com/mweagle/SpartaHTML).

For reference, the `helloWorld` function is below.

```go
import (
	awsLambdaEvents "github.com/aws/aws-lambda-go/events"
)
func helloWorld(ctx context.Context,
	gatewayEvent spartaAWSEvents.APIGatewayRequest) (interface{}, error) {
	/*
		 To return an error back to the client using a standard HTTP status code:

			errorResponse := spartaAPIG.NewErrorResponse(http.StatusInternalError,
			"Something failed inside here")
			return errorResponse, nil

			You can also create custom error response types, so long as they
			include `"code":HTTP_STATUS_CODE` somewhere in the response body.
			This reserved expression is what Sparta uses as a RegExp to determine
			the Integration Mapping value
	*/

	logger, loggerOk := ctx.Value(sparta.ContextKeyLogger).(*logrus.Logger)
	if loggerOk {
		logger.Info("Hello world structured log message")
	}
	// Return a message, together with the incoming input...
	return &helloWorldResponse{
		Message: fmt.Sprintf("Hello world 🌏"),
		Request: gatewayEvent,
	}, nil
}
```

# Create the API Gateway

The first requirement is to create a new [API](https://godoc.org/github.com/mweagle/Sparta#API) instance via `sparta.NewAPIGateway()`

```go
stage := sparta.NewStage("prod")
apiGateway := sparta.NewAPIGateway("MySpartaAPI", stage)
```

In the example above, we're also including a [Stage](https://godoc.org/github.com/mweagle/Sparta#Stage) value.  A non-`nil` Stage value will cause the registered API to be deployed.  If the Stage value is `nil`, a REST API will be created, but it will not be [deployed](http://docs.aws.amazon.com/apigateway/latest/developerguide/how-to-deploy-api.html) (and therefore not publicly accessible).

# Create a Resource

The next step is to associate a URL path with the `sparta.LambdaAWSInfo` struct that encapsulates the **go** function:

```go
func spartaHTMLLambdaFunctions(api *sparta.API) []*sparta.LambdaAWSInfo {
	var lambdaFunctions []*sparta.LambdaAWSInfo
	lambdaFn := sparta.HandleAWSLambda(sparta.LambdaName(helloWorld),
		helloWorld,
		sparta.IAMRoleDefinition{})

	if nil != api {
		apiGatewayResource, _ := api.NewResource("/hello", lambdaFn)

		// We only return http.StatusOK
		apiMethod, apiMethodErr := apiGatewayResource.NewMethod("GET",
			http.StatusOK,
			http.StatusInternalServerError)
		if nil != apiMethodErr {
			panic("Failed to create /hello resource: " + apiMethodErr.Error())
		}
		// The lambda resource only supports application/json Unmarshallable
		// requests.
		apiMethod.SupportedRequestContentTypes = []string{"application/json"}
	}
	return append(lambdaFunctions, lambdaFn)
}
```

Our `helloWorld` only supports `GET`.  We'll see how a single lambda function can support multiple HTTP methods shortly.

# Provision

The final step is to to provide the API instance to `Sparta.Main()`

```go
// Register the function with the API Gateway
apiStage := sparta.NewStage("v1")
apiGateway := sparta.NewAPIGateway("SpartaHTML", apiStage)
```

Once the service is successfully provisioned, the `Outputs` key will include the API Gateway Deployed URL (sample):

```text
INFO[0096] ────────────────────────────────────────────────
INFO[0096] Stack Outputs
INFO[0096] ────────────────────────────────────────────────
INFO[0096] S3SiteURL                                     Description="S3 Website URL" Value="http://spartahtml-mweagle-s3site89c05c24a06599753eb3ae4e-1w6rehqu6x04c.s3-website-us-west-2.amazonaws.com"
INFO[0096] APIGatewayURL                                 Description="API Gateway URL" Value="https://w2tefhnt4b.execute-api.us-west-2.amazonaws.com/v1"
INFO[0096] ────────────────────────────────────────────────
```

Combining the _API Gateway URL_ `OutputValue` with our resource path (_/hello/world/test_), we get the absolute URL to our lambda function: _https://w2tefhnt4b.execute-api.us-west-2.amazonaws.com/v1/hello_

# Querying

Let's query the lambda function and see what the `event` data is at execution time. The
snippet below is pretty printed by piping the response through [jq](https://stedolan.github.io/jq/).

```nohighlight
$ curl -vs https://3e7ux226ga.execute-api.us-west-2.amazonaws.com/v1/hello | jq .
*   Trying 52.84.237.220...
* TCP_NODELAY set
* Connected to 3e7ux226ga.execute-api.us-west-2.amazonaws.com (52.84.237.220) port 443 (#0)
* TLS 1.2 connection using TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
* Server certificate: *.execute-api.us-west-2.amazonaws.com
* Server certificate: Amazon
* Server certificate: Amazon Root CA 1
* Server certificate: Starfield Services Root Certificate Authority - G2
> GET /v1/hello HTTP/1.1
> Host: 3e7ux226ga.execute-api.us-west-2.amazonaws.com
> User-Agent: curl/7.54.0
> Accept: */*
>
< HTTP/1.1 200 OK
< Content-Type: application/json
< Content-Length: 1137
< Connection: keep-alive
< Date: Mon, 29 Jan 2018 14:15:28 GMT
< x-amzn-RequestId: db7f5734-04fe-11e8-b264-c70ecab3a032
< Access-Control-Allow-Origin: http://spartahtml-mweagle-s3site89c05c24a06599753eb3ae4e-419zo4dp8n2d.s3-website-us-west-2.amazonaws.com
< Access-Control-Allow-Headers: Content-Type,X-Amz-Date,Authorization,X-Api-Key
< Access-Control-Allow-Methods: *
< X-Amzn-Trace-Id: sampled=0;root=1-5a6f2c80-efb0f84554384252abca6d15
< X-Cache: Miss from cloudfront
< Via: 1.1 570a1979c411cb4529fa1e711db52490.cloudfront.net (CloudFront)
< X-Amz-Cf-Id: -UsCegiR1K3vJUFyAo9IMrWGdH8rKW6UBrtJLjxZqke19r0cxMl1NA==
<
{ [1137 bytes data]
* Connection #0 to host 3e7ux226ga.execute-api.us-west-2.amazonaws.com left intact
{
  "Message": "Hello world 🌏",
  "Request": {
    "method": "GET",
    "body": {},
    "headers": {
      "Accept": "*/*",
      "CloudFront-Forwarded-Proto": "https",
      "CloudFront-Is-Desktop-Viewer": "true",
      "CloudFront-Is-Mobile-Viewer": "false",
      "CloudFront-Is-SmartTV-Viewer": "false",
      "CloudFront-Is-Tablet-Viewer": "false",
      "CloudFront-Viewer-Country": "US",
      "Host": "3e7ux226ga.execute-api.us-west-2.amazonaws.com",
      "User-Agent": "curl/7.54.0",
      "Via": "1.1 570a1979c411cb4529fa1e711db52490.cloudfront.net (CloudFront)",
      "X-Amz-Cf-Id": "vAFNTV5uAMeTG9JN6IORnA7LYJhZyB3jHV7vh-7lXn2uZQUR6eHQUw==",
      "X-Amzn-Trace-Id": "Root=1-5a6f2c80-2b48a9c86a30b0162d8ab1f1",
      "X-Forwarded-For": "73.118.138.121, 205.251.214.60",
      "X-Forwarded-Port": "443",
      "X-Forwarded-Proto": "https"
    },
    "queryParams": {},
    "pathParams": {},
    "context": {
      "appId": "",
      "method": "GET",
      "requestId": "db7f5734-04fe-11e8-b264-c70ecab3a032",
      "resourceId": "401s9n",
      "resourcePath": "/hello",
      "stage": "v1",
      "identity": {
        "accountId": "",
        "apiKey": "",
        "caller": "",
        "cognitoAuthenticationProvider": "",
        "cognitoAuthenticationType": "",
        "cognitoIdentityId": "",
        "cognitoIdentityPoolId": "",
        "sourceIp": "73.118.138.121",
        "user": "",
        "userAgent": "curl/7.54.0",
        "userArn": ""
      }
    }
  }
}
```

While this demonstrates that our lambda function is publicly accessible, it's not immediately obvious where the `*event` data is being populated.

# Mapping Templates

The event data that's actually supplied to `echoS3Event` is the complete HTTP request body.  This content is what the API Gateway sends to our lambda function, which is defined by the integration mapping.  This event data also includes the values of any whitelisted parameters.  When the API Gateway Method is defined, it optionally includes any whitelisted query params and header values that should be forwarded to the integration target.  For this example, we're not whitelisting any params, so those fields (`queryParams`, `pathParams`) are empty.  Then for each integration target (which can be AWS Lambda, a mock, or a HTTP Proxy), it's possible to transform the API Gateway request data and whitelisted arguments into a format that's more amenable to the target.

Sparta uses a pass-through template that passes all valid data, with minor **Body** differences based on the inbound _Content-Type_:

  * [application/json](https://github.com/mweagle/Sparta/blob/master/resources/provision/apigateway/inputmapping_json.vtl)
  * [*](https://github.com/mweagle/Sparta/blob/master/resources/provision/apigateway/inputmapping_default.vtl)

The `application/json` template is copied below:

```nohighlight
#*
Provide an automatic pass through template that transforms all inputs
into the JSON payload sent to a golang function. The JSON behavior attempts to parse
the incoming HTTP body as JSON assign it to the `body` field.

See
  https://forums.aws.amazon.com/thread.jspa?threadID=220274&tstart=0
  http://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-mapping-template-reference.html
*#
{
  "method": "$context.httpMethod",
  "body" : $input.json('$'),
  "headers": {
    #foreach($param in $input.params().header.keySet())
    "$param": "$util.escapeJavaScript($input.params().header.get($param))" #if($foreach.hasNext),#end
    #end
  },
  "queryParams": {
    #foreach($param in $input.params().querystring.keySet())
    "$param": "$util.escapeJavaScript($input.params().querystring.get($param))" #if($foreach.hasNext),#end

    #end
  },
  "pathParams": {
    #foreach($param in $input.params().path.keySet())
    "$param": "$util.escapeJavaScript($input.params().path.get($param))" #if($foreach.hasNext),#end

    #end
  },
  "context" : {
    "apiId" : "$util.escapeJavaScript($context.apiId)",
    "method" : "$util.escapeJavaScript($context.httpMethod)",
    "requestId" : "$util.escapeJavaScript($context.requestId)",
    "resourceId" : "$util.escapeJavaScript($context.resourceId)",
    "resourcePath" : "$util.escapeJavaScript($context.resourcePath)",
    "stage" : "$util.escapeJavaScript($context.stage)",
    "identity" : {
      "accountId" : "$util.escapeJavaScript($context.identity.accountId)",
      "apiKey" : "$util.escapeJavaScript($context.identity.apiKey)",
      "caller" : "$util.escapeJavaScript($context.identity.caller)",
      "cognitoAuthenticationProvider" : "$util.escapeJavaScript($context.identity.cognitoAuthenticationProvider)",
      "cognitoAuthenticationType" : "$util.escapeJavaScript($context.identity.cognitoAuthenticationType)",
      "cognitoIdentityId" : "$util.escapeJavaScript($context.identity.cognitoIdentityId)",
      "cognitoIdentityPoolId" : "$util.escapeJavaScript($context.identity.cognitoIdentityPoolId)",
      "sourceIp" : "$util.escapeJavaScript($context.identity.sourceIp)",
      "user" : "$util.escapeJavaScript($context.identity.user)",
      "userAgent" : "$util.escapeJavaScript($context.identity.userAgent)",
      "userArn" : "$util.escapeJavaScript($context.identity.userArn)"
    }
  }
}
```

This template forwards all whitelisted data & body to the lambda function.  You can see by switching on the `method` field would permit a single function to service multiple HTTP method names.

The next example will show how to unmarshal this data and perform request-specific actions.

# Proxying Envelope

Because the integration request returned a successful response, the API Gateway response body contains only our lambda's output (`$input.json('$')`).

To return an error that API Gateway can properly translate into an HTTP
status code, use an [apigateway.NewErrorResponse](https://godoc.org/github.com/mweagle/Sparta/aws/apigateway#NewErrorResponse) type. This
custom error type includes fields that trigger integration mappings based on the
inline [HTTP StatusCode](https://golang.org/src/net/http/status.go). If you look at the **Integration Response** section of the _/hello/world/test_ resource in the Console, you'll see a list of Regular Expression matches:

![API Gateway](/images/apigateway/IntegrationMapping.png)

The regular expressions are used to translate the integration response, which is just a blob of text provided to the underlying AWS Go API, into API Gateway Method responses.
If you use the `apigateway.Error` type, the marshalled string version of your
error text will trigger the proper Method Response.

Sparta annotates your lambda functions response with **go**'s values based on the HTTP status code your lambda function produced.  By default, Sparta also provides a corresponding Method Response entry for all valid HTTP codes:

![API Gateway](/images/apigateway/MethodResponse.png)

These mappings are defaults, and it's possible to override either one by providing a non-zero length values to either:

  * [Integration.Responses](https://godoc.org/github.com/mweagle/Sparta#Integration).  See the [DefaultIntegrationResponses](https://github.com/mweagle/Sparta/blob/master/apigateway.go#L60) for the default values.
  * [Method.Responses](https://godoc.org/github.com/mweagle/Sparta#Method).  See the [DefaultMethodResponses](https://godoc.org/github.com/mweagle/Sparta#DefaultMethodResponses) for the default method response mappings.

# Cleaning Up

Before moving on, remember to decommission the service via:

```bash
go run application.go delete
```

# Wrapping Up

Now that we know what data is actually being sent to our API Gateway-connected Lambda function, we'll move on to performing a more complex operation, including returning a custom HTTP response body.

# Other Resources

  * [Mapping Template Reference](http://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-mapping-template-reference.html)
