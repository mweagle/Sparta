---
date: 2016-03-09T19:56:50+01:00
title: DynamoDB
weight: 10
---

In this section we'll walkthrough how to trigger your lambda function in response to DynamoDB stream events.  This overview is based on the [SpartaApplication](https://github.com/mweagle/SpartaApplication) sample code if you'd rather jump to the end result.

# Goal

Assume that we're given a DynamoDB stream.  See [below](http://localhost:1313/docs/eventsources/dynamodb/#creatingDynamoDBStream:d680e8a854a7cbad6d490c445cba2eba) for details on how to create the stream.  We've been asked to write a lambda function that logs when operations are performed to the table so that we can perform offline analysis.

# Getting Started

We'll start with an empty lambda function and build up the needed functionality.

{{< highlight go >}}
import (
	awsLambdaEvents "github.com/aws/aws-lambda-go/events"
)
func echoDynamoDBEvent(ctx context.Context, ddbEvent awsLambdaEvents.DynamoDBEvent) (*awsLambdaEvents.DynamoDBEvent, error) {
	logger, _ := ctx.Value(sparta.ContextKeyRequestLogger).(*logrus.Entry)
	logger.WithFields(logrus.Fields{
		"Event": ddbEvent,
	}).Info("Event received")
	return &ddbEvent, nil
}
{{< /highlight >}}

Since the `echoDynamoDBEvent` is triggered by Dynamo events, we can leverage the AWS Go Lambda SDK [event types](https://godoc.org/github.com/aws/aws-lambda-go/events)
to access the record.

# <a href="{{< relref "#spartaIntegration" >}}">Sparta Integration</a>

With the core of the `echoDynamoDBEvent` complete, the next step is to integrate the **go** function with Sparta.  This is performed by the [appendDynamoDBLambda](https://github.com/mweagle/SpartaApplication/blob/master/application.go#L114) function.  Since the `echoDynamoDBEvent` function doesn't access any additional services (Sparta enables CloudWatch Logs privileges by default), the integration is pretty straightforward:

{{< highlight go >}}
lambdaFn := sparta.HandleAWSLambda(
  sparta.LambdaName(echoDynamoDBEvent),
  echoDynamoDBEvent,
  sparta.IAMRoleDefinition{})
{{< /highlight >}}

# Event Source Mappings

If we were to deploy this Sparta application, the `echoDynamoDBEvent` function would have the ability to log DynamoDB stream events, but would not be invoked in response to events published by the stream.  To register for notifications, we need to configure the lambda's [EventSourceMappings](http://docs.aws.amazon.com/lambda/latest/dg/intro-core-components.html#intro-core-components-event-sources):

{{< highlight go >}}
  lambdaFn.EventSourceMappings = append(lambdaFn.EventSourceMappings,
    &lambda.CreateEventSourceMappingInput{
      EventSourceArn:   aws.String(dynamoTestStream),
      StartingPosition: aws.String("TRIM_HORIZON"),
      BatchSize:        aws.Int64(10),
      Enabled:          aws.Bool(true),
  })
lambdaFunctions = append(lambdaFunctions, lambdaFn)
{{< /highlight >}}

The `dynamoTestStream` param is the ARN of the Dynamo stream that that your lambda function will [poll](http://docs.aws.amazon.com/lambda/latest/dg/intro-invocation-modes.html) (eg: _arn:aws:dynamodb:us-west-2:000000000000:table/myDynamoDBTable/stream/2015-12-05T16:28:11.869_).

The `EventSourceMappings` field is transformed into the appropriate [CloudFormation Resource](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-lambda-eventsourcemapping.html) which enables automatic polling of the DynamoDB stream.

# Wrapping Up

With the `lambdaFn` fully defined, we can provide it to `sparta.Main()` and deploy our service.  The workflow below is shared by all DynamoDB stream based lambda functions:

  * Define the lambda function (`echoDynamoDBEvent`).
  * If needed, create the required [IAMRoleDefinition](https://godoc.org/github.com/mweagle/Sparta*IAMRoleDefinition) with appropriate privileges if the lambda function accesses other AWS services.
  * Provide the lambda function & IAMRoleDefinition to `sparta.HandleAWSLambda()`
  * Add the necessary [EventSourceMappings](https://godoc.org/github.com/aws/aws-sdk-go/service/lambda#CreateEventSourceMappingInput) to the `LambdaAWSInfo` struct so that the lambda function is properly configured.

# Other Resources

  * [Using Triggers for Cross Region DynamoDB Replication](https://aws.amazon.com/blogs/aws/dynamodb-update-triggers-streams-lambda-cross-region-replication-app/)

<hr />
# Appendix

## Creating a DynamoDB Stream

To create a DynamoDB stream for a given table, follow the steps below:

### Select Table

![Select Table](/images/eventsources/dynamodb/DynamoDB_ManageStream.png)

### Enable Stream

![Enable Stream](/images/eventsources/dynamodb/DynamoDB_Enable.png)

### Copy ARN
![Copy ARN](/images/eventsources/dynamodb/DynamoDB_StreamARN.png)

The **Latest stream ARN** value is the value that should be provided as the `EventSourceArn` in to the [Event Source Mappings](http://localhost:1313/docs/eventsources/dynamodb/#eventSourceMapping:d680e8a854a7cbad6d490c445cba2eba).
