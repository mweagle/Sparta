---
date: 2019-01-31 05:47:27
title: CodeCommit
weight: 10
---

The CodeCommit Lambda [event source](https://docs.aws.amazon.com/codecommit/latest/userguide/how-to-notify-lambda.html)
allows you to trigger lambda functions in response to CodeCommit repository events.

## Events

Lambda functions triggered in response to CodeCommit evetms use a combination of [events and branches](https://docs.aws.amazon.com/codecommit/latest/APIReference/API_RepositoryTrigger.html) to manage which state changes trigger your lambda function.

To create an event subscriber use a constructor as in:

```go
// CodeCommit reactor function
func reactorFunc(ctx context.Context, event awsLambdaEvents.CodeCommitEvent) (interface{}, error) {
  logger, _ := ctx.Value(sparta.ContextKeyRequestLogger).(*logrus.Entry)
  logger.WithFields(logrus.Fields{
    "Event": event,
  }).Info("Event received")
  return &event, nil
}

func main() {
  // ...
  handler := spartaArchetype.NewCodeCommitReactor(reactorFunc)
  reactor, reactorErr := spartaArchetype.NewCodeCommitReactor(handler,
    gocf.String("MyRepositoryName"),
    nil, // Defaults to 'all' branches
    nil, // Defaults to 'all' events
    nil) // Additional IAM privileges
  ...
}
```
