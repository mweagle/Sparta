+++
author = "Matt Weagle"
date = "2015-11-29T06:50:17"
title = "Overview"
tags = ["sparta"]
type = "doc"
+++

In this section we'll walkthrough how to trigger your lambda function in response to SNS events.  This overview is based on the [SpartaApplication](https://github.com/mweagle/SpartaApplication/blob/master/application.go#L79) sample code if you'd rather jump to the end result.

## <a href="{{< relref "#goal" >}}">Goal</a>

Assume that we have an SNS topic that broadcasts notifications.  We've been asked to write a lambda function that logs the _Subject_ and _Message_ text to CloudWatch logs for later processing.

## <a href="{{< relref "#gettingStarted" >}}">Getting Started</a>

We'll start with an empty lambda function and build up the needed functionality.

{{< highlight go >}}

func echoSNSEvent(event *json.RawMessage,
                  context *sparta.LambdaContext,
                  w http.ResponseWriter,
                  logger *logrus.Logger)
{
  logger.WithFields(logrus.Fields{
    "RequestID": context.AWSRequestID,
  }).Info("Request received")
}
{{< /highlight >}}

## <a href="{{< relref "#unmarshalSNSEvent" >}}">Unmarshalling the SNS Event</a>


Since the `echoSNSEvent` is expected to be triggered by SNS notifications, we will unmarshal the `*json.RawMessage` data into an SNS-specific event provided by Sparta via:

{{< highlight go >}}

var lambdaEvent spartaSNS.Event
err := json.Unmarshal([]byte(*event), &lambdaEvent)
if err != nil {
  logger.Error("Failed to unmarshal event data: ", err.Error())
  http.Error(w, err.Error(), http.StatusInternalServerError)
}
{{< /highlight >}}   

SNS events are delivered in batches, via lists of [EventRecords](https://godoc.org/github.com/mweagle/Sparta/aws/sns#EventRecord
), so we'll need to process each record.

{{< highlight go >}}
for _, eachRecord := range lambdaEvent.Records {
  logger.WithFields(logrus.Fields{
    "Subject": eachRecord.Sns.Subject,
    "Message": eachRecord.Sns.Message,
  }).Info("SNS Event")
}
{{< /highlight >}}   

That's enough to get the data into CloudWatch Logs.

## <a href="{{< relref "#spartaIntegration" >}}">Sparta Integration</a>

With the core of the `echoSNSEvent` complete, the next step is to integrate the *Go* function with Sparta.  This is performed by the [appendSNSLambda](https://github.com/mweagle/SpartaApplication/blob/master/application.go#L79) function.  Since the `echoSNSEvent` function doesn't access any additional services (Sparta enables CloudWatch Logs privileges by default), the integration is pretty straightforward:

{{< highlight go >}}
lambdaFn = sparta.NewLambda(sparta.IAMRoleDefinition{}, echoSNSEvent, nil)
{{< /highlight >}}   

## <a href="{{< relref "#eventSourceRegistration" >}}">Event Source Registration</a>

If we were to deploy this Sparta application, the `echoSNSEvent` function would have the ability to log SNS events, but would not be invoked in response to messages published to that topic.  To register for notifications, we need to configure the lambda's [Permissions](http://docs.aws.amazon.com/lambda/latest/dg/intro-permission-model.html):

{{< highlight go >}}
lambdaFn.Permissions = append(lambdaFn.Permissions, sparta.SNSPermission{
  BasePermission: sparta.BasePermission{
    SourceArn: snsTopic,
  },
})
lambdaFunctions = append(lambdaFunctions, lambdaFn)
{{< /highlight >}}  

The `snsTopic` param is the ARN of the SNS topic that will notify your lambda function (eg: _arn:aws:sns:us-west-2:000000000000:myTopicName).  

See the [S3 docs](http://gosparta.io/docs/eventsources/s3/#eventSourceRegistration) for more information on how the _Permissions_ data is processed.

## <a href="{{< relref "#wrappingUp" >}}">Wrapping Up</a>

With the `lambdaFn` fully defined, we can provide it to `sparta.Main()` and deploy our service.  The workflow below is shared by all SNS-triggered lambda:

  * Define the lambda function (`echoSNSEvent`).
  * If needed, create the required [IAMRoleDefinition](https://godoc.org/github.com/mweagle/Sparta*IAMRoleDefinition) with appropriate privileges if the lambda function accesses other AWS services.
  * Provide the lambda function & IAMRoleDefinition to `sparta.NewLambda()`
  * Add the necessary [Permissions](https://godoc.org/github.com/mweagle/Sparta#LambdaAWSInfo) to the `LambdaAWSInfo` struct so that the lambda function is triggered.

## <a href="{{< relref "#otherResources" >}}">Other Resources</a>

  * TBD
