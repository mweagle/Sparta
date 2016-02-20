+++
author = "Matt Weagle"
date = "2016-02-16T06:40:36Z"
title = "Event Source - CloudWatch Logs"
tags = ["sparta", "event_source"]
type = "doc"
+++

TODO: Finish CloudWatch Logs docs


In this section we'll walkthrough how to trigger your lambda function in response to  [CloudWatch Logs](https://aws.amazon.com/blogs/aws/new-cloudwatch-events-track-and-respond-to-changes-to-your-aws-resources/).  This overview is based on the [SpartaApplication](https://github.com/mweagle/SpartaApplication) sample code if you'd rather jump to the end result.

# Goal

Assume that we're supposed to write a simple "HelloWorld" CloudWatch Logs function that should be triggered in response to any log message issued to a specific Log Group.

# Getting Started

Our lambda function is relatively short:

{{< highlight go >}}
func echoCloudWatchLogsEvent(event *json.RawMessage,
                        context *sparta.LambdaContext,
                        w http.ResponseWriter,
                        logger *logrus.Logger) {

  // Note that we're not going to log in this lambda function, as
  // we don't want to self DDOS
  fmt.Fprintf(w, "Hello World!")
}
{{< /highlight >}}   

Our lambda function doesn't need to do much with the log message other than log it.

# Sparta Integration

With `echoCloudWatchLogsEvent()` implemented, the next step is to integrate the **Go** function with Sparta.  This is done by the `appendCloudWatchLogsLambda` in the SpartaApplication [application.go](https://github.com/mweagle/SpartaApplication/blob/master/application.go) source.

Our lambda function only needs logfile write privileges, and since these are enabled by default, we can use an empty `sparta.IAMRoleDefinition` value:

{{< highlight go >}}
func appendCloudWatchLogsLambda(api *sparta.API,
	lambdaFunctions []*sparta.LambdaAWSInfo) []*sparta.LambdaAWSInfo {
	lambdaFn := sparta.NewLambda(sparta.IAMRoleDefinition{}, echoCloudWatchLogsEvent, nil)

{{< /highlight >}}   

The next step is to add a `CloudWatchLogsSubscriptionFilter` value that represents the [CloudWatch Lambda](http://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/Subscriptions.html#LambdaFunctionExample) subscription [filter information](http://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/CreateSubscriptionFilter.html).

{{< highlight go >}}
cloudWatchLogsPermission := sparta.CloudWatchLogsPermission{}
cloudWatchLogsPermission.Filters = make(map[string]sparta.CloudWatchLogsSubscriptionFilter, 1)
cloudWatchLogsPermission.Filters["MyFilter"] = sparta.CloudWatchLogsSubscriptionFilter{
  LogGroupName: "/aws/lambda/versions",
}
{{< /highlight >}}   

The `sparta.CloudWatchLogsPermission` struct provides fields for both the LogGroupName and optional Filter expression (not shown here) to use when calling [putSubscriptionFilter](http://docs.aws.amazon.com/AWSJavaScriptSDK/latest/AWS/CloudWatchLogs.html#putSubscriptionFilter-property).

  # Add Permission

  With the subscription information configured, the final step is to add the `sparta.CloudWatchLogsPermission` to our `sparta.LambdaAWSInfo` value:

{{< highlight go >}}
lambdaFn.Permissions = append(lambdaFn.Permissions, cloudWatchLogsPermission)
{{< /highlight >}}  

Our entire function is therefore:

{{< highlight go >}}
func appendCloudWatchLogsLambda(api *sparta.API,
	lambdaFunctions []*sparta.LambdaAWSInfo) []*sparta.LambdaAWSInfo {
    
	lambdaFn := sparta.NewLambda(sparta.IAMRoleDefinition{}, echoCloudWatchLogsEvent, nil)

	cloudWatchLogsPermission := sparta.CloudWatchLogsPermission{}
	cloudWatchLogsPermission.Filters = make(map[string]sparta.CloudWatchLogsSubscriptionFilter, 1)
	cloudWatchLogsPermission.Filters["MyFilter"] = sparta.CloudWatchLogsSubscriptionFilter{
		LogGroupName: "/aws/lambda/versions",
	}
	lambdaFn.Permissions = append(lambdaFn.Permissions, cloudWatchLogsPermission)
	return append(lambdaFunctions, lambdaFn)
}
{{< /highlight >}}  


# Wrapping Up

With the `lambdaFn` fully defined, we can provide it to `sparta.Main()` and deploy our service.  The workflow below is shared by all CloudWatch Logs-triggered lambda functions:

  * Define the lambda function (`echoCloudWatchLogsEvent`).
  * If needed, create the required [IAMRoleDefinition](https://godoc.org/github.com/mweagle/Sparta*IAMRoleDefinition) with appropriate privileges.
  * Provide the lambda function & IAMRoleDefinition to `sparta.NewLambda()`
  * Create a [CloudWatchLogsPermission](https://godoc.org/github.com/mweagle/Sparta#CloudWatchLogsPermission) value.
  * Add one or more [CloudWatchLogsSubscriptionFilter](https://godoc.org/github.com/mweagle/Sparta#CloudWatchLogsSubscriptionFilter) to the `CloudWatchLogsPermission.Filters` map that defines your lambda function's logfile subscription information.
  * Append the `CloudWatchLogsPermission` value to the lambda function's `Permissions` slice.
  * Include the reference in the call to `sparta.Main()`.

# Other Resources
