---
date: 2016-03-09T19:56:50+01:00
title: Kinesis
weight: 10
---

In this section we'll walkthrough how to trigger your lambda function in response to [Amazon Kinesis](https://aws.amazon.com/kinesis/) streams. This overview is based on the [SpartaApplication](https://github.com/mweagle/SpartaApplication/blob/master/application.go#L130) sample code if you'd rather jump to the end result.

# Goal

The goal of this example is to provision a Sparta lambda function that logs Amazon Kinesis events to CloudWatch logs.

## Getting Started

We'll start with an empty lambda function and build up the needed functionality.

```go
import (
	awsLambdaEvents "github.com/aws/aws-lambda-go/events"
)
func echoKinesisEvent(ctx context.Context, kinesisEvent awsLambdaEvents.KinesisEvent) (*awsLambdaEvents.KinesisEvent, error) {
  logger, _ := ctx.Value(sparta.ContextKeyRequestLogger).(*zerolog.Logger)

  logger.Info().
    Interface("Event", kinesisEvent).
    Msg("Event received")

	return &kinesisEvent, nil
}
```

For this sample all we're going to do is transparently unmarshal the Kinesis event to an AWS Lambda [event](https://godoc.org/github.com/aws/aws-lambda-go/events), log
it, and return the value.

With the function defined let's register it with Sparta.

## Sparta Integration

First we wrap the **go** function in a [LambdaAWSInfo](https://godoc.org/github.com/mweagle/Sparta#LambdaAWSInfo) struct:

```go
lambdaFn, _ := sparta.NewAWSLambda(sparta.LambdaName(echoKinesisEvent),
	echoKinesisEvent,
	sparta.IAMRoleDefinition{})
```

Since our lambda function doesn't access any other AWS Services, we can use an empty IAMRoleDefinition (`sparta.IAMRoleDefinition{}`).

## Event Source Registration

Then last step is to configure our AWS Lambda function with Kinesis as the [EventSource](http://docs.aws.amazon.com/lambda/latest/dg/intro-core-components.html)

```go
lambdaFn.EventSourceMappings = append(lambdaFn.EventSourceMappings,
  &lambda.CreateEventSourceMappingInput{
    EventSourceArn:   aws.String(kinesisTestStream),
    StartingPosition: aws.String("TRIM_HORIZON"),
    BatchSize:        aws.Int64(100),
    Enabled:          aws.Bool(true),
  })
```

The `kinesisTestStream` parameter is the Kinesis stream ARN (eg: _arn:aws:kinesis:us-west-2:123412341234:stream/kinesisTestStream_) whose events will trigger lambda execution.

# Wrapping Up

With the `lambdaFn` fully defined, we can provide it to `sparta.Main()` and deploy our service. The workflow below is shared by all Kinesis-triggered lambda functions:

- Define the lambda function (`echoKinesisEvent`).
- If needed, create the required [IAMRoleDefinition](https://godoc.org/github.com/mweagle/Sparta*IAMRoleDefinition) with appropriate privileges if the lambda function accesses other AWS services.
- Provide the lambda function & IAMRoleDefinition to `sparta.NewAWSLambda()`
- Add the necessary [EventSourceMappings](https://godoc.org/github.com/aws/aws-sdk-go-v2/service/lambda#CreateEventSourceMappingInput) to the `LambdaAWSInfo` struct so that the lambda function is properly configured.

# Notes

- The Kinesis stream and the AWS Lambda function must be provisioned in the same region.
- The AWS docs have an excellent [Kinesis EventSource](http://docs.aws.amazon.com/lambda/latest/dg/walkthrough-kinesis-events-adminuser.html) walkthrough.
