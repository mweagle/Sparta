---
date: 2016-03-09T19:56:50+01:00
title: Event Sources
weight: 1
---

The true power of the AWS Lambda architecture is the ability to integrate Lambda execution with other AWS service state transitions.  Depending on the service type, state change events are either pushed or transparently polled and used as the input to a Lambda execution.

There are several [event sources](http://docs.aws.amazon.com/lambda/latest/dg/intro-core-components.html) available.  They are grouped into Pull and Push types.  Pull based models use `sparta.EventSourceMapping` values, as the trigger configuration is stored in the AWS Lambda service.  Push based types use service specific `sparta.*Permission` types to denote the fact that the trigger logic is configured in the remote service.


## Pull Based

  * [DynamoDB](/reference/eventsources/dynamodb)
  * [Kinesis](/reference/eventsources/kinesis)
  * [SQS](/reference/eventsources/sqs)

## Push Based

  * [CloudFormation](/reference/eventsources/cloudformation) <span class="label label-warning">NOT YET IMPLEMENTED</span>
  * [CloudWatch Events](/reference/eventsources/cloudwatchevents)
  * [CloudWatch Logs](/reference/eventsources/cloudwatchlogs)
  * [Cognito](/reference/eventsources/cognito) <span class="label label-warning">NOT YET IMPLEMENTED</span>
  * [S3](/reference/eventsources/s3)
  * [SES](/reference/eventsources/ses)
  * [SNS](/reference/eventsources/sns)
