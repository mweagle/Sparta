---
date: 2018-11-28 20:03:46
title: SNS
weight: 10
---

To create a SNS reactor that subscribes via an [subscription configuration](https://docs.aws.amazon.com/lambda/latest/dg/invoking-lambda-function.html#supported-event-source-sns),
use the [NewSNSReactor](https://godoc.org/github.com/mweagle/Sparta/archetype#NewSNSReactor) constructor as in:

```go
import (
  awsLambdaEvents "github.com/aws/aws-lambda-go/events"
  spartaArchetype "github.com/mweagle/Sparta/archetype"
)
// DynamoDB reactor function
func reactorFunc(ctx context.Context,
  snsEvent awsLambdaEvents.SNSEvent) (interface{}, error) {
  logger, _ := ctx.Value(sparta.ContextKeyRequestLogger).(*zerolog.Logger)

  logger.Info().
    Interface("Event", snsEvent).
    Msg("SNS Event")

  return "Hello World 👋. Welcome to AWS Lambda! 🙌🎉🍾", nil
}

func main() {
  // ...
  handler := spartaArchetype.SNSReactorFunc(reactorFunc)
  lambdaFn, lambdaFnErr := spartaArchetype.NewDynamoDBReactor(handler,
    "SNS_ARN_OR_CLOUDFORMATION_REF_VALUE",
    nil)
}
```
