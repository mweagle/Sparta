+++
author = "Matt Weagle"
date = "2015-11-29T06:50:17"
title = "Event Source - Kinesis"
tags = ["sparta", "event_source"]
type = "doc"
+++

In this section we'll walkthrough how to trigger your lambda function in response to [Amazon Kinesis](https://aws.amazon.com/kinesis/) streams.  This overview is based on the [SpartaApplication](https://github.com/mweagle/SpartaApplication/blob/master/application.go#L130) sample code if you'd rather jump to the end result.  

## <a href="{{< relref "#goal" >}}">Goal</a>

The goal of this example is to provision a Sparta lambda function that logs Amazon Kinesis events to CloudWatch logs.

## <a href="{{< relref "#gettingStarted" >}}">Getting Started</a>

We'll start with an empty lambda function and build up the needed functionality.

{{< highlight go >}}
func echoKinesisEvent(event *json.RawMessage,
                      context *sparta.LambdaContext,
                      w http.ResponseWriter,
                      logger *logrus.Logger)
{
  logger.WithFields(logrus.Fields{
		"RequestID": context.AWSRequestID,
		"Event":     string(*event),
	}).Info("Request received")

{{< /highlight >}}   

For this sample all we're going to do is unmarshal the Kinesis [event](http://docs.aws.amazon.com/lambda/latest/dg/walkthrough-kinesis-events-adminuser-create-test-function.html#wt-kinesis-invoke-manually) to a Sparta [kinesis event](https://godoc.org/github.com/mweagle/Sparta/aws/kinesis#Event) and log the id to CloudWatch Logs:

{{< highlight go >}}

  var lambdaEvent spartaKinesis.Event
  err := json.Unmarshal([]byte(*event), &lambdaEvent)
  if err != nil {
    logger.Error("Failed to unmarshal event data: ", err.Error())
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
  for _, eachRecord := range lambdaEvent.Records {
    logger.WithFields(logrus.Fields{
      "EventID": eachRecord.EventID,
    }).Info("Kinesis Event")
  }
}
{{< /highlight >}}   

With the function defined let's register it with Sparta.

## <a href="{{< relref "#spartaIntegration" >}}">Sparta Integration</a>

First we wrap the **Go** function in a [LambdaAWSInfo](https://godoc.org/github.com/mweagle/Sparta#LambdaAWSInfo) struct:

{{< highlight go >}}
lambdaFn := sparta.NewLambda(sparta.IAMRoleDefinition{}, echoKinesisEvent, nil)
{{< /highlight >}}   

Since our lambda function doesn't access any other AWS Services, we can use an empty IAMRoleDefinition (`sparta.IAMRoleDefinition{}`).

## <a href="{{< relref "#eventSourceRegistration" >}}">Event Source Registration</a>

Then last step is to configure our AWS Lambda function with Kinesis as the [EventSource](http://docs.aws.amazon.com/lambda/latest/dg/intro-core-components.html)

{{< highlight go >}}
lambdaFn.EventSourceMappings = append(lambdaFn.EventSourceMappings, &lambda.CreateEventSourceMappingInput{
  EventSourceArn:   aws.String(kinesisTestStream),
  StartingPosition: aws.String("TRIM_HORIZON"),
  BatchSize:        aws.Int64(100),
  Enabled:          aws.Bool(true),
})
{{< /highlight >}}   

The `kinesisTestStream` parameter is the Kinesis stream ARN (eg: _arn:aws:kinesis:us-west-2:123412341234:stream/kinesisTestStream_) whose events will trigger lambda execution.

## <a href="{{< relref "#wrappingUp" >}}">Wrapping Up</a>

With the `lambdaFn` fully defined, we can provide it to `sparta.Main()` and deploy our service.  The workflow below is shared by all Kinesis-triggered lambda functions:

  * Define the lambda function (`echoKinesisEvent`).
  * If needed, create the required [IAMRoleDefinition](https://godoc.org/github.com/mweagle/Sparta*IAMRoleDefinition) with appropriate privileges if the lambda function accesses other AWS services.
  * Provide the lambda function & IAMRoleDefinition to `sparta.NewLambda()`
  * Add the necessary [EventSourceMappings](https://godoc.org/github.com/aws/aws-sdk-go/service/lambda#CreateEventSourceMappingInput) to the `LambdaAWSInfo` struct so that the lambda function is properly configured.

## <a href="{{< relref "#otherResources" >}}">Notes</a>

  * The Kinesis stream and the AWS Lambda function must be provisioned in the same region.
  * The AWS docs have an excellent [Kinesis EventSource](http://docs.aws.amazon.com/lambda/latest/dg/walkthrough-kinesis-events-adminuser.html) walkthrough.
