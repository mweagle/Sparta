---
date: 2020-02-22T17:36:59
title: Event Bridge
weight: 10
---

The EventBridge lambda [event source](https://aws.amazon.com/eventbridge/)
allows you to trigger lambda functions in response to either cron schedules or account events. There
are two different archetype functions available.

## Scheduled

Scheduled Lambdas execute either at fixed times or periodically depending on the [schedule expression](https://docs.aws.amazon.com/eventbridge/latest/userguide/scheduled-events.html).

To create a scheduled function use a constructor as in:

```go
import (
  spartaArchetype "github.com/mweagle/Sparta/archetype"
)

// EventBridge reactor function
func echoEventBridgeEvent(ctx context.Context, msg json.RawMessage) (interface{}, error) {
  logger, _ := ctx.Value(sparta.ContextKeyLogger).(*logrus.Logger)
  var eventData map[string]interface{}
  err := json.Unmarshal(msg, &eventData)
  logger.WithFields(logrus.Fields{
    "error":   err,
    "message": eventData,
  }).Info("EventBridge event")
  return nil, err
}

func main() {
  // ...
  eventBridgeReactorFunc := spartaArchetype.EventBridgeReactorFunc(echoEventBridgeEvent)
  lambdaFn, lambdaFnErr := spartaArchetype.NewEventBridgeScheduledReactor(eventBridgeReactorFunc,
    "rate(1 minute)",
    nil)
  // ...
}
```

When the scheduled event is triggered, the log statement outputs the full payload:

```json
{
  "error": null,
  "level": "info",
  "message": {
    "account": "123412341234",
    "detail": {},
    "detail-type": "Scheduled Event",
    "id": "f453bd1e-ccea-9df4-4e40-938097e82869",
    "region": "us-west-2",
    "resources": [
      "arn:aws:events:us-west-2:123412341234:rule/SpartaEventBridge-0271594-EventBridgexmainechoEven-2WMCXA1LWGZY"
    ],
    "source": "aws.events",
    "time": "2020-02-23T00:47:31Z",
    "version": "0"
  },
  "msg": "EventBridge event",
  "time": "2020-02-23T00:48:08Z"
}
```

See the [scheduled event payload](https://docs.aws.amazon.com/eventbridge/latest/userguide/aws-events.html) documentation.

## Events

Lambda functions can also be triggered via EventBridge by providing an [event patterns](https://docs.aws.amazon.com/eventbridge/latest/userguide/eventbridge-and-event-patterns.html) to
select which events should trigger your function's execution.

To create an event subscriber use a constructor as in:

```go
func echoEventBridgeEvent(ctx context.Context, msg json.RawMessage) (interface{}, error) {
  logger, _ := ctx.Value(sparta.ContextKeyLogger).(*logrus.Logger)
  var eventData map[string]interface{}
  err := json.Unmarshal(msg, &eventData)
  logger.WithFields(logrus.Fields{
    "error":   err,
    "message": eventData,
  }).Info("EventBridge event")
  return nil, err
}

func main() {
  // ...
  eventBridgeReactorFunc := spartaArchetype.EventBridgeReactorFunc(echoEventBridgeEvent)
  lambdaFn, lambdaFnErr := spartaArchetype.NewEventBridgeEventReactor(eventBridgeReactorFunc,
    map[string]interface{}{
      "source": []string{"aws.ec2"},
    },
    nil)
  // ...
}
```

The event payload data depends on what event is being subscribed to. See the [event patterns](https://docs.aws.amazon.com/eventbridge/latest/userguide/filtering-examples-structure.html) documentation for more information.
