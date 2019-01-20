---
date: 2016-03-09T19:56:50+01:00
title: SNS
weight: 10
---

In this section we'll walkthrough how to trigger your lambda function in response to SNS events.  This overview is based on the [SpartaApplication](https://github.com/mweagle/SpartaApplication/blob/master/application.go#L79) sample code if you'd rather jump to the end result.

# Goal

Assume that we have an SNS topic that broadcasts notifications.  We've been asked to write a lambda function that logs the _Subject_ and _Message_ text to CloudWatch logs for later processing.

# Getting Started

We'll start with an empty lambda function and build up the needed functionality.

```go
import (
	awsLambdaEvents "github.com/aws/aws-lambda-go/events"
)
func echoSNSEvent(ctx context.Context, snsEvent awsLambdaEvents.SNSEvent) (*awsLambdaEvents.SNSEvent, error) {
	logger, _ := ctx.Value(sparta.ContextKeyRequestLogger).(*logrus.Entry)
	logger.WithFields(logrus.Fields{
		"Event": snsEvent,
	}).Info("Event received")
	return &snsEvent, nil
}
```

# Unmarshalling the SNS Event

SNS events are delivered in batches, via lists of [SNSEventRecords](https://godoc.org/github.com/aws/aws-lambda-go/events#SNSEventRecord
), so we'll need to process each record.

```go
for _, eachRecord := range lambdaEvent.Records {
	logger.WithFields(logrus.Fields{
		"Subject": eachRecord.Sns.Subject,
		"Message": eachRecord.Sns.Message,
	}).Info("SNS Event")
}
```

That's enough to get the data into CloudWatch Logs.

# Sparta Integration

With the core of the `echoSNSEvent` complete, the next step is to integrate the **go** function with Sparta.  This is performed by
the [appendSNSLambda](https://github.com/mweagle/SpartaApplication/blob/master/application.go#L79) function.  Since the `echoSNSEvent`
function doesn't access any additional services (Sparta enables CloudWatch Logs privileges by default), the integration is
pretty straightforward:

```go
lambdaFn, _ := sparta.NewAWSLambda(sparta.LambdaName(echoSNSEvent),
  echoSNSEvent,
  sparta.IAMRoleDefinition{})
```

# Event Source Registration

If we were to deploy this Sparta application, the `echoSNSEvent` function would have the ability to log SNS events, but would not be invoked in response to messages published to that topic.  To register for notifications, we need to configure the lambda's [Permissions](http://docs.aws.amazon.com/lambda/latest/dg/intro-permission-model.html):

```go
lambdaFn.Permissions = append(lambdaFn.Permissions, sparta.SNSPermission{
  BasePermission: sparta.BasePermission{
    SourceArn: snsTopic,
  },
})
lambdaFunctions = append(lambdaFunctions, lambdaFn)
```

The `snsTopic` param is the ARN of the SNS topic that will notify your lambda function (eg: _arn:aws:sns:us-west-2:000000000000:myTopicName).

See the [S3 docs](http://gosparta.io/docs/eventsources/s3/#eventSourceRegistration) for more information on how the _Permissions_ data is processed.

# Wrapping Up

With the `lambdaFn` fully defined, we can provide it to `sparta.Main()` and deploy our service.  The workflow below is shared by all SNS-triggered lambda function:

  * Define the lambda function (`echoSNSEvent`).
  * If needed, create the required [IAMRoleDefinition](https://godoc.org/github.com/mweagle/Sparta*IAMRoleDefinition) with appropriate privileges if the lambda function accesses other AWS services.
  * Provide the lambda function & IAMRoleDefinition to `sparta.NewAWSLambda()`
  * Add the necessary [Permissions](https://godoc.org/github.com/mweagle/Sparta#LambdaAWSInfo) to the `LambdaAWSInfo` struct so that the lambda function is triggered.

# Other Resources

  * TBD
