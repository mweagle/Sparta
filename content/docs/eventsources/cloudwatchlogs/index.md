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

Assume that we're supposed to write a simple "HelloWorld" CloudWatch Logs function that should be triggered in response to any log message issued by the other lambda functions in our Sparta-based microservice.

# Getting Started

Our lambda function is relatively short:

{{< highlight go >}}
func echoCloudWatchLogs(event *json.RawMessage,
                        context *sparta.LambdaContext,
                        w http.ResponseWriter,
                        logger *logrus.Logger) {

  // Note that we're not going to log in this lambda function, as
  // we don't want to self DDOS
	fmt.Fprintf(w, "Hello World!")
}
{{< /highlight >}}   

Our lambda function doesn't need to do much with the log message other than log it.

# Sparta Integration {#spartaIntegration}  

# Add Permission


# Wrapping Up

# Other Resources
