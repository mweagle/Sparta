---
date: 2018-01-22 21:49:38
title: CloudWatch Alarms
weight: 10
alwaysopen: false
---

The [CloudWatchErrorAlarmDecorator](https://godoc.org/github.com/mweagle/Sparta/decorator#CloudWatchErrorAlarmDecorator) associates a CloudWatch alarm and destination with your Lambda function.

Sample usage:

```go
lambdaFn := sparta.HandleAWSLambda("Hello World",
  helloWorld,
  sparta.IAMRoleDefinition{})

lambdaFn.Decorators = []sparta.TemplateDecoratorHandler{
  spartaDecorators.CloudWatchErrorAlarmDecorator(1,
    1,
    1,
    gocf.String("MY_SNS_ARN")),
}
```
