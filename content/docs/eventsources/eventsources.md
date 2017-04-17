---
date: 2016-03-09T19:56:50+01:00
title: Event Sources
weight: 10
menu:
  main:
    parent: Documentation
    identifier: eventsources-overview
    weight: 0
---

The true power of the AWS Lambda architecture is the ability to integrate Lambda execution with other AWS service state transitions.  Depending on the service type, state change events are either pushed or transparently polled and used as the input to a Lambda execution.

There are several [event sources](http://docs.aws.amazon.com/lambda/latest/dg/intro-core-components.html) available.  They are grouped into Pull and Push types.  Pull based models use `sparta.EventSourceMapping` values, as the trigger configuration is stored in the AWS Lambda service.  Push based types use service specific `sparta.*Permission` types to denote the fact that the trigger logic is configured in the remote service.

  * Pull Based
    * [DynamoDB](/docs/eventsources/dynamodb)
    * [Kinesis](/docs/eventsources/kinesis)
  * Pushed Based
    * [CloudFormation](/docs/eventsources/cloudformation) <span class="label label-warning">NOT YET IMPLEMENTED</span>
    * [CloudWatch Events](/docs/eventsources/cloudwatchevents)
    * [CloudWatch Logs](/docs/eventsources/cloudwatchlogs)
    * [Cognito](/docs/eventsources/cognito) <span class="label label-warning">NOT YET IMPLEMENTED</span>
    * [S3](/docs/eventsources/s3)
    * [SES](/docs/eventsources/ses)
    * [SNS](/docs/eventsources/sns)
