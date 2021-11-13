---
date: 2018-11-28 20:03:46
title: CloudWatch
weight: 10
---

The CloudWatch Logs Lambda [event source](https://docs.aws.amazon.com/lambda/latest/dg/invoking-lambda-function.html#supported-event-source-cloudwatch-logs)
allows you to trigger lambda functions in response to either cron schedules or account events. There
are three different archetype functions available.

## Scheduled

Scheduled Lambdas execute either at fixed times or periodically depending on the [schedule expression](https://docs.aws.amazon.com/lambda/latest/dg/tutorial-scheduled-events-schedule-expressions.html).

To create a scheduled function use a constructor as in:

```go
import (
  awsLambdaEvents "github.com/aws/aws-lambda-go/events"
  spartaArchetype "github.com/mweagle/Sparta/v3/archetype"
)
// CloudWatch reactor function
func reactorFunc(ctx context.Context,
  cwLogs awsLambdaEvents.CloudwatchLogsEvent) (interface{}, error) {
  logger, _ := ctx.Value(sparta.ContextKeyRequestLogger).(*zerolog.Logger)

  logger.Info().
    Interface("Event", cwLogs).
    Msg("Cron triggered")
  return "Hello World 👋. Welcome to AWS Lambda! 🙌🎉🍾", nil
}

func main() {
  // ...
  handler := spartaArchetype.CloudWatchLogsReactorFunc(reactorFunc)
  subscriptions := map[string]string{
    "every5Mins": "rate(5 minutes)",
  }
  lambdaFn, lambdaFnErr := spartaArchetype.NewCloudWatchScheduledReactor(handler, subscriptions, nil)
}
```

## Events

Lambda functions triggered in response to CloudWatch Events use [event patterns](https://docs.aws.amazon.com/AmazonCloudWatch/latest/events/CloudWatchEventsandEventPatterns.html) to
select which events should trigger your function's execution.

To create an event subscriber use a constructor as in:

```go
// CloudWatch reactor function
func reactorFunc(ctx context.Context,
  cwLogs awsLambdaEvents.CloudwatchLogsEvent) (interface{}, error) {
  logger, _ := ctx.Value(sparta.ContextKeyRequestLogger).(*zerolog.Logger)

  logger.Info().
    Interface("Event", cwLogs).
    Msg("Cron triggered")

  return "Hello World 👋. Welcome to AWS Lambda! 🙌🎉🍾", nil
}

func main() {
  // ...
  handler := spartaArchetype.CloudWatchLogsReactorFunc(reactorFunc)
  subscriptions := map[string]string{
    "ec2StateChanges": map[string]interface{}{
      "source":      []string{"aws.ec2"},
      "detail-type": []string{"EC2 Instance state change"},
    },
  }
  lambdaFn, lambdaFnErr := spartaArchetype.NewCloudWatchEventedReactor(handler, subscriptions, nil)
}
```

## Generic

Both `NewCloudWatchScheduledReactor` and `NewCloudWatchEventedReactor` are convenience functions
for the generic `NewCloudWatchReactor` constructor. For example, it's possible to create a
scheduled lambda execution using the generic constructor as in:

```go
func main() {
  // ...
  subscriptions := map[string]sparta.CloudWatchEventsRule{
    "every5Mins": sparta.CloudWatchEventsRule{
      ScheduleExpression: "rate(5 minutes)",
    },
  }
  lambdaFn, lambdaFnErr := spartaArchetype.NewCloudWatchReactor(handler, subscriptions, nil)
}
```
