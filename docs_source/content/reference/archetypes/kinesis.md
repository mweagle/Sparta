---
date: 2018-11-28 20:03:46
title: Kinesis
weight: 10
---

To create a Kinesis Stream reactor that subscribes via an [EventSourceMapping](https://docs.aws.amazon.com/lambda/latest/dg/with-kinesis.html),
use the [NewKinesisReactor](http://localhost:6060/pkg/github.com/mweagle/Sparta/archetype/#NewKinesisReactor) constructor as in:

```go
import (
  awsLambdaEvents "github.com/aws/aws-lambda-go/events"
  spartaArchetype "github.com/mweagle/Sparta/archetype"
)
// KinesisStream reactor function
func reactorFunc(ctx context.Context,
  kinesisEvent awsLambdaEvents.KinesisEvent) (interface{}, error) {
  logger, _ := ctx.Value(sparta.ContextKeyRequestLogger).(*zerolog.Logger)

  logger.Info().
    Interface("Event", kinesisEvent).
    Msg("Kinesis Event")

  return "Hello World 👋. Welcome to AWS Lambda! 🙌🎉🍾", nil
}

func main() {
  // ...
  handler := spartaArchetype.KinesisReactorFunc(reactorFunc)
  lambdaFn, lambdaFnErr := spartaArchetype.NewKinesisReactor(handler,
    "KINESIS_STREAM_ARN_OR_CLOUDFORMATION_REF_VALUE",
    "TRIM_HORIZON",
    10,
    nil)
}
```
