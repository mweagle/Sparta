---
date: 2016-03-09T19:56:50+01:00
title: SQS
weight: 10
---

In this section we'll walkthrough how to trigger your lambda function in response to AWS Simple Queue Service (SQS) events. This overview is based on the [SpartaSQS](https://github.com/mweagle/SpartaSQS) sample code if you'd rather jump to the end result.

# Goal

The goal here is to create a self-contained service that provisions a SQS queue, an AWS Lambda function that processes messages posted to the queue

## Getting Started

We'll start with an empty lambda function and build up the needed functionality.

```go
import (
	"context"

	awsLambdaGo "github.com/aws/aws-lambda-go/events"
	sparta "github.com/mweagle/Sparta"
	spartaCF "github.com/mweagle/Sparta/aws/cloudformation"
	gocf "github.com/mweagle/go-cloudformation"
	"github.com/rs/zerolog"
)

func sqsHandler(ctx context.Context, sqsRequest awsLambdaGo.SQSEvent) error {
	logger, _ := ctx.Value(sparta.ContextKeyLogger).(*zerolog.Logger)
	logger.WithField("Event", sqsRequest).Info("SQS Event Received")
	return nil
}
```

Since the `sqsHandler` function subscribes to SQS messages, it can use the AWS provided [SQSEvent](https://godoc.org/github.com/aws/aws-lambda-go/events#SQSEvent) to automatically unmarshal the incoming event.

Typically the lambda function would process each record in the event, but for this example we'll just log the entire batch and then return.

## Sparta Integration

The next step is to integrate the lambda function with Sparta:

```go
// 1. Create the Sparta Lambda function
lambdaFn, _ := sparta.NewAWSLambda(sparta.LambdaName(sqsHandler),
  sqsHandler,
  sparta.IAMRoleDefinition{})
```

Once the lambda function is integrated with Sparta, we can use a [TemplateDecoratorHandler](https://godoc.org/github.com/mweagle/Sparta#TemplateDecoratorHandler) to include the SQS provisioning request as part of the overall service creation.

## SQS Topic Definition

Decorators enable a Sparta service to provision other types of infrastructure together with the core lambda functions. In this example, our `sqsHandler` function should also provision an SQS queue from which it will receive events. This is done as in the following:

```go
sqsResourceName := "LambdaSQSFTW"
sqsDecorator := func(serviceName string,
  lambdaResourceName string,
  lambdaResource gocf.LambdaFunction,
  resourceMetadata map[string]interface{},
  S3Bucket string,
  S3Key string,
  buildID string,
  template *gocf.Template,
  context map[string]interface{},
  logger *zerolog.Logger) error {

  // Include the SQS resource in the application
  sqsResource := &gocf.SQSQueue{}
  template.AddResource(sqsResourceName, sqsResource)
  return nil
}
lambdaFn.Decorators = []sparta.TemplateDecoratorHandler{sparta.TemplateDecoratorHookFunc(sqsDecorator)}
```

This function-level decorator includes an AWS CloudFormation [SQS::Queue](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-sqs-queues.html) definition that will be included with the stack definition.

## Connecting SQS to AWS Lambda

The final step is to make the `sqsHandler` the Lambda's [EventSourceMapping](https://godoc.org/github.com/mweagle/Sparta#EventSourceMapping) target for the dynamically provisioned Queue's _ARN_:

```go
lambdaFn.EventSourceMappings = append(lambdaFn.EventSourceMappings,
  &sparta.EventSourceMapping{
    EventSourceArn: gocf.GetAtt(sqsResourceName, "Arn"),
    BatchSize:      2,
  })
```

# Wrapping Up

With the `lambdaFn` fully defined, we can provide it to `sparta.Main()` and deploy our service. It's also possible to use a pre-existing SQS resource by providing a string literal as the `EventSourceArn` value.

## Other Resources

- The AWS docs have an excellent [SQS event source](https://aws.amazon.com/blogs/aws/aws-lambda-adds-amazon-simple-queue-service-to-supported-event-sources/) walkthrough.
