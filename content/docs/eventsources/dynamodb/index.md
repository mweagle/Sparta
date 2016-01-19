+++
author = "Matt Weagle"
date = "2015-11-29T06:50:17"
title = "Event Source - DynamoDB"
tags = ["sparta"]
type = "doc"
+++

In this section we'll walkthrough how to trigger your lambda function in response to DynamoDB stream events.  This overview is based on the [SpartaApplication](https://github.com/mweagle/SpartaApplication) sample code if you'd rather jump to the end result.

## <a href="{{< relref "#goal" >}}">Goal</a>

Assume that we're given a DynamoDB stream.  See [below](http://localhost:1313/docs/eventsources/dynamodb/#creatingDynamoDBStream:d680e8a854a7cbad6d490c445cba2eba) for details on how to create the stream.  We've been asked to write a lambda function that logs when operations are performed to the table so that we can perform offline analysis.

## <a href="{{< relref "#gettingStarted" >}}">Getting Started</a>

We'll start with an empty lambda function and build up the needed functionality.

{{< highlight go >}}

func echoDynamoDBEvent(event *json.RawMessage,
                       context *sparta.LambdaContext,
                       w http.ResponseWriter,
                      logger *logrus.Logger)
{
  logger.WithFields(logrus.Fields{
    "RequestID": context.AWSRequestID,
  }).Info("Request received")
}
{{< /highlight >}}

## <a href="{{< relref "#unmarshalDynamoDBEvent" >}}">Unmarshalling the DynamoDB Event</a>

Since the `echoDynamoDBEvent` is expected to be triggered by DynamoDB events, we will unmarshal the `*json.RawMessage` data into an DynamoDB-specific event provided by Sparta via:

{{< highlight go >}}

var lambdaEvent spartaDynamoDB.Event
err := json.Unmarshal([]byte(*event), &lambdaEvent)
if err != nil {
  logger.Error("Failed to unmarshal event data: ", err.Error())
  http.Error(w, err.Error(), http.StatusInternalServerError)
}
{{< /highlight >}}   

DynamoDB events are delivered in batches, via lists of [EventRecords](https://godoc.org/github.com/mweagle/Sparta/aws/dynamodb#EventRecord
), so we'll need to process each record.

{{< highlight go >}}
for _, eachRecord := range lambdaEvent.Records {
  logger.WithFields(logrus.Fields{
    "Keys":     eachRecord.DynamoDB.Keys,
    "NewImage": eachRecord.DynamoDB.NewImage,
  }).Info("DynamoDb Event")
}
{{< /highlight >}}   

That's enough to get the data into CloudWatch Logs.

## <a href="{{< relref "#spartaIntegration" >}}">Sparta Integration</a>

With the core of the `echoDynamoDBEvent` complete, the next step is to integrate the *Go* function with Sparta.  This is performed by the [appendDynamoDBLambda](https://github.com/mweagle/SpartaApplication/blob/master/application.go#L114) function.  Since the `echoDynamoDBEvent` function doesn't access any additional services (Sparta enables CloudWatch Logs privileges by default), the integration is pretty straightforward:

{{< highlight go >}}
lambdaFn = sparta.NewLambda(sparta.IAMRoleDefinition{}, echoDynamoDBEvent, nil)
{{< /highlight >}}   

## <a href="{{< relref "#eventSourceMapping" >}}">Event Source Mappings</a>

If we were to deploy this Sparta application, the `echoDynamoDBEvent` function would have the ability to log DynamoDB stream events, but would not be invoked in response to events published by the stream.  To register for notifications, we need to configure the lambda's [EventSourceMappings](http://docs.aws.amazon.com/lambda/latest/dg/intro-core-components.html#intro-core-components-event-sources):

{{< highlight go >}}
  lambdaFn.EventSourceMappings = append(lambdaFn.EventSourceMappings, &lambda.CreateEventSourceMappingInput{
    EventSourceArn:   aws.String(dynamoTestStream),
    StartingPosition: aws.String("TRIM_HORIZON"),
    BatchSize:        aws.Int64(10),
    Enabled:          aws.Bool(true),
  })
lambdaFunctions = append(lambdaFunctions, lambdaFn)
{{< /highlight >}}  

The `dynamoTestStream` param is the ARN of the Dynamo stream that that your lambda function will [poll](http://docs.aws.amazon.com/lambda/latest/dg/intro-invocation-modes.html) (eg: _arn:aws:dynamodb:us-west-2:000000000000:table/myDynamoDBTable/stream/2015-12-05T16:28:11.869_).  

The `EventSourceMappings` field is transformed into the appropriate [CloudFormation Resource](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-lambda-eventsourcemapping.html) which enables automatic polling of the DynamoDB stream.

## <a href="{{< relref "#wrappingUp" >}}">Wrapping Up</a>

With the `lambdaFn` fully defined, we can provide it to `sparta.Main()` and deploy our service.  The workflow below is shared by all DynamoDB stream based lambda functions:

  * Define the lambda function (`echoDynamoDBEvent`).
  * If needed, create the required [IAMRoleDefinition](https://godoc.org/github.com/mweagle/Sparta*IAMRoleDefinition) with appropriate privileges if the lambda function accesses other AWS services.
  * Provide the lambda function & IAMRoleDefinition to `sparta.NewLambda()`
  * Add the necessary [EventSourceMappings](https://godoc.org/github.com/aws/aws-sdk-go/service/lambda#CreateEventSourceMappingInput) to the `LambdaAWSInfo` struct so that the lambda function is properly configured.

## <a href="{{< relref "#otherResources" >}}">Other Resources</a>

  * [Using Triggers for Cross Region DynamoDB Replication](https://aws.amazon.com/blogs/aws/dynamodb-update-triggers-streams-lambda-cross-region-replication-app/)

<hr />
## Appendix
### <a href="{{< relref "#creatingDynamoDBStream" >}}">Creating a DynamoDB Stream</a>

To create a DynamoDB stream for a given table, follow the steps below:

#### Select Table

![Select Table](/images/eventsources/dynamodb/DynamoDB_ManageStream.png)

#### Enable Stream

![Enable Stream](/images/eventsources/dynamodb/DynamoDB_Enable.png)

#### Copy ARN
![Copy ARN](/images/eventsources/dynamodb/DynamoDB_StreamARN.png)

The **Latest stream ARN** value is the value that should be provided as the `EventSourceArn` in to the [Event Source Mappings](http://localhost:1313/docs/eventsources/dynamodb/#eventSourceMapping:d680e8a854a7cbad6d490c445cba2eba).
