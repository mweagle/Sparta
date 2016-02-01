+++
author = "Matt Weagle"
date = "2015-11-29T06:50:17"
title = "Event Sources - Overview"
tags = ["sparta"]
type = "doc"
+++

The true power of the AWS Lambda architecture is the ability to integrate Lambda execution with other AWS service state transitions.  Depending on the service type, state change events are either pushed or transparently polled and used as the input to a Lambda execution.  

There are several [event sources](http://docs.aws.amazon.com/lambda/latest/dg/intro-core-components.html) available:

  * [CloudFormation](/docs/eventsources/cloudformation) <span class="label label-warning">NOT YET IMPLEMENTED</span>
  * [CloudWatch Events](/docs/eventsources/cloudwatchevents)
  * [CloudWatch Logs](/docs/eventsources/cloudwatchlogs) <span class="label label-warning">NOT YET IMPLEMENTED</span>
  * [Cognito](/docs/eventsources/cognito) <span class="label label-warning">NOT YET IMPLEMENTED</span>
  * [DynamoDB](/docs/eventsources/dynamodb)
  * [Kinesis](/docs/eventsources/kinesis)
  * [S3](/docs/eventsources/s3)
  * [SES](/docs/eventsources/ses)
  * [SNS](/docs/eventsources/sns)
