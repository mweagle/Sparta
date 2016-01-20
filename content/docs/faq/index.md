+++
author = "Matt Weagle"
date = "2016-01-20T21:12:27Z"
title = "FAQ"
tags = ["sparta"]
type = "doc"
+++

* Development
  * [How can I test locally?]({{< relref "#devfaq1" >}})
* Event Sources
  * SES
      * [Where does the _SpartaRuleSet_ come from?]({{< relref "#sesfaq1" >}})  
      * [Why does `provision` not enable the _SpartaRuleSet_?]({{< relref "#sesfaq2" >}})  
* Operations
  * [Where can I view my function's `*logger` output?]({{< relref "#opsfaq1" >}})  
  * [Where can I view Sparta's golang spawn metrics?]({{< relref "#opsfaq2" >}})  
  * [How can I include additional AWS resources as part of my Sparta application?]({{< relref "#opsfaq3" >}})
  * [How can I determine the outputs available in sparta.Discover() for dynamic AWS resources?]({{< relref "#opsfaq4" >}})

## Development
<hr />

### How can I test locally? {#faq1}

Local testing is available via the [explore](/docs/local_testing/) command.

## Event Sources - SES
<hr />

### Where does the _SpartaRuleSet_ come from?  {sesfaq1}  

SES only permits a single [active receipt rule](http://docs.aws.amazon.com/ses/latest/APIReference/API_SetActiveReceiptRuleSet.html).  Additionally, it's possible that multiple Sparta-based services are handing different SES recipients.  

All Sparta-based services share the _SpartaRuleSet_ SES ruleset, and uniquely identify their Rules by including the current servicename as part of the SES [ReceiptRule](http://docs.aws.amazon.com/ses/latest/APIReference/API_CreateReceiptRule.html).

### Why does `provision` not always enable the _SpartaRuleSet_?  {#sesfaq2}  

Initial _SpartaRuleSet_ will make it the active ruleset, but Sparta assumes that manual updates made outside of the context of the framework were done with good reason and doesn't attempt to override the user setting.

## Operations
<hr />

### Where can I view my function's `*logger` output?  {#opsfaq1}  

Each lambda function includes privileges to write to [CloudWatch Logs](https://console.aws.amazon.com/cloudwatch/home).  The `*logrus.logger` output is written (with a brief delay) to a lambda-specific log group.  

The CloudWatch log group name includes a sanitized version of your **Go** function name & owning service name.

### Where can I view Sparta's golang spawn metrics?  {#opsfaq2}  

Visit the [CloudWatch Metrics](https://aws.amazon.com/cloudwatch/) AWS console page and select the `Sparta/{SERVICE_NAME}` namespace:

![CloudWatch](/images/faq/CloudWatch_Management_Console.jpg)

Sparta publishes two counters:

  * `ProcessSpawned`: A new **Go** process was spawned to handle requests
  * `ProcessReused`: An existing **Go** process was used to handle requests.  See also the discussion on AWS Lambda [container reuse](https://aws.amazon.com/blogs/compute/container-reuse-in-lambda/).

### How can I include additional AWS resources as part of my Sparta application?  {#opsfaq4}  

Define a [TemplateDecorator](https://godoc.org/github.com/mweagle/Sparta#TemplateDecorator) function and annotate the `*gocf.Template` with additional AWS resources.

### How can I determine the outputs available in sparta.Discover() for dynamic AWS resources?  {#opsfaq4}  

The list of registered output provider types is defined by `cloudformationTypeMapDiscoveryOutputs` in [cloudformation_resources.go](https://github.com/mweagle/Sparta/blob/master/cloudformation_resources.go).  See the [CloudFormation Resource Types Reference](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-template-resource-type-ref.html) for information on interpreting the values.
