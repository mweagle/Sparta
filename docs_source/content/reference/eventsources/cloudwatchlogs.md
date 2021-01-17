---
date: 2016-03-09T19:56:50+01:00
title: CloudWatch Logs
weight: 10
---

In this section we'll walkthrough how to trigger your lambda function in response to [CloudWatch Logs](https://aws.amazon.com/blogs/aws/new-cloudwatch-events-track-and-respond-to-changes-to-your-aws-resources/). This overview is based on the [SpartaApplication](https://github.com/mweagle/SpartaApplication) sample code if you'd rather jump to the end result.

# Goal

Assume that we're supposed to write a simple "HelloWorld" CloudWatch Logs function that should be triggered in response to any log message issued to a specific Log Group.

## Getting Started

Our lambda function is relatively short:

```go
import (
	awsLambdaEvents "github.com/aws/aws-lambda-go/events"
)
func echoCloudWatchLogsEvent(ctx context.Context, cwlEvent awsLambdaEvents.CloudwatchLogsEvent) (*awsLambdaEvents.CloudwatchLogsEvent, error) {
	logger, _ := ctx.Value(sparta.ContextKeyRequestLogger).(*zerolog.Logger)

  logger.Info().
    Interface("Event", cwlEvent).
		Msg("Event received")

	return &cwlEvent, nil
}
```

Our lambda function doesn't need to do much with the log message other than log and return it.

## Sparta Integration

With `echoCloudWatchLogsEvent()` implemented, the next step is to integrate the **go** function with Sparta. This is done by the `appendCloudWatchLogsLambda` in the SpartaApplication [application.go](https://github.com/mweagle/SpartaApplication/blob/master/application.go) source.

Our lambda function only needs logfile write privileges, and since these are enabled by default, we can use an empty `sparta.IAMRoleDefinition` value:

```go
func appendCloudWatchLogsHandler(api *sparta.API,
	lambdaFunctions []*sparta.LambdaAWSInfo) []*sparta.LambdaAWSInfo {
	lambdaFn, _ := sparta.NewAWSLambda(sparta.LambdaName(echoCloudWatchLogsEvent),
		echoCloudWatchLogsEvent,
		sparta.IAMRoleDefinition{})
```

The next step is to add a `CloudWatchLogsSubscriptionFilter` value that represents the [CloudWatch Lambda](http://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/Subscriptions.html#LambdaFunctionExample) subscription [filter information](http://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/CreateSubscriptionFilter.html).

```go
cloudWatchLogsPermission := sparta.CloudWatchLogsPermission{}
cloudWatchLogsPermission.Filters = make(map[string]sparta.CloudWatchLogsSubscriptionFilter, 1)
cloudWatchLogsPermission.Filters["MyFilter"] = sparta.CloudWatchLogsSubscriptionFilter{
  LogGroupName: "/aws/lambda/versions",
}
```

The `sparta.CloudWatchLogsPermission` struct provides fields for both the LogGroupName and optional Filter expression (not shown here) to use when calling [putSubscriptionFilter](http://docs.aws.amazon.com/AWSJavaScriptSDK/latest/AWS/CloudWatchLogs.html#putSubscriptionFilter-property).

## Add Permission

With the subscription information configured, the final step is to add the `sparta.CloudWatchLogsPermission` to our `sparta.LambdaAWSInfo` value:

```go
lambdaFn.Permissions = append(lambdaFn.Permissions, cloudWatchLogsPermission)
```

Our entire function is therefore:

```go
func appendCloudWatchLogsHandler(api *sparta.API,
	lambdaFunctions []*sparta.LambdaAWSInfo) []*sparta.LambdaAWSInfo {

	lambdaFn, _ := sparta.NewAWSLambda(sparta.LambdaName(echoCloudWatchLogsEvent),
		echoCloudWatchLogsEvent,
		sparta.IAMRoleDefinition{})
	cloudWatchLogsPermission := sparta.CloudWatchLogsPermission{}
	cloudWatchLogsPermission.Filters = make(map[string]sparta.CloudWatchLogsSubscriptionFilter, 1)
	cloudWatchLogsPermission.Filters["MyFilter"] = sparta.CloudWatchLogsSubscriptionFilter{
		FilterPattern: "",
		// NOTE: This LogGroupName MUST already exist in your account, otherwise
		// the `provision` step will fail. You can create a LogGroupName in the
		// AWS Console
		LogGroupName:  "/aws/lambda/versions",
	}
	lambdaFn.Permissions = append(lambdaFn.Permissions, cloudWatchLogsPermission)
	return append(lambdaFunctions, lambdaFn)
}
```

# Wrapping Up

With the `lambdaFn` fully defined, we can provide it to `sparta.Main()` and deploy our service. The workflow below is shared by all CloudWatch Logs-triggered lambda functions:

- Define the lambda function (`echoCloudWatchLogsEvent`).
- If needed, create the required [IAMRoleDefinition](https://godoc.org/github.com/mweagle/Sparta*IAMRoleDefinition) with appropriate privileges.
- Provide the lambda function & IAMRoleDefinition to `sparta.NewAWSLambda()`
- Create a [CloudWatchLogsPermission](https://godoc.org/github.com/mweagle/Sparta#CloudWatchLogsPermission) value.
- Add one or more [CloudWatchLogsSubscriptionFilter](https://godoc.org/github.com/mweagle/Sparta#CloudWatchLogsSubscriptionFilter) to the `CloudWatchLogsPermission.Filters` map that defines your lambda function's logfile subscription information.
- Append the `CloudWatchLogsPermission` value to the lambda function's `Permissions` slice.
- Include the reference in the call to `sparta.Main()`.

## Other Resources
