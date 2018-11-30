---
date: 2018-11-28 20:03:46
title: DynamoDB
weight: 10
---

To create a DynamoDB reactor that subscribes via an [EventSourceMapping](https://docs.aws.amazon.com/lambda/latest/dg/with-ddb.html),
use the [NewDynamoDBReactor](https://godoc.org/github.com/mweagle/Sparta/archetype#NewDynamoDBReactor) constructor as in:

```go
import (
  awsLambdaEvents "github.com/aws/aws-lambda-go/events"
  spartaArchetype "github.com/mweagle/Sparta/archetype"
)
// DynamoDB reactor function
func reactorFunc(ctx context.Context,
  dynamoEvent awsLambdaEvents.DynamoDBEvent) (interface{}, error) {
  logger, _ := ctx.Value(sparta.ContextKeyRequestLogger).(*logrus.Entry)

  logger.WithFields(logrus.Fields{
    "Event": dynamoEvent,
  }).Info("DynamoDB Event")
  return "Hello World 👋. Welcome to AWS Lambda! 🙌🎉🍾", nil
}

func main() {
  // ...
  handler := spartaArchetype.DynamoDBReactorFunc(reactorFunc)
  lambdaFn, lambdaFnErr := spartaArchetype.NewDynamoDBReactor(handler,
    "DYNAMO_DB_ARN_OR_CLOUDFORMATION_REF_VALUE",
    "TRIM_HORIZON",
    10,
    nil)
}
```