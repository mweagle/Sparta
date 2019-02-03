---
date: 2019-01-31 05:44:32
title: CodeCommit
weight: 10
---

In this section we'll walkthrough how to trigger your lambda function in response to [CodeCommit Events](https://docs.aws.amazon.com/codecommit/latest/userguide/how-to-notify-lambda.html/).

# Goal

Assume that we're supposed to write a Lambda function that is triggered in response to any event emitted by a CodeCommit repository.

## Getting Started

Our lambda function is relatively short:

```go
import (
  awsLambdaEvents "github.com/aws/aws-lambda-go/events"
)

func echoCodeCommit(ctx context.Context, event awsLambdaEvents.CodeCommitEvent) (interface{}, error) {
  logger, _ := ctx.Value(sparta.ContextKeyRequestLogger).(*logrus.Entry)
  logger.WithFields(logrus.Fields{
    "Event": event,
  }).Info("Event received")
  return &event, nil
}
```

Our lambda function doesn't need to do much with the repository message other than log and return it.

## Sparta Integration

With `echoCodeCommit()` defined, the next step is to integrate the **go** function with your application.

Our lambda function only needs logfile write privileges, and since these are enabled by default, we can use an empty `sparta.IAMRoleDefinition` value:

```go
func appendCloudWatchLogsHandler(api *sparta.API,
  lambdaFunctions []*sparta.LambdaAWSInfo) []*sparta.LambdaAWSInfo {
  lambdaFn, _ := sparta.NewAWSLambda(sparta.LambdaName(echoCodeCommit),
    echoCodeCommit,
    sparta.IAMRoleDefinition{})
```

The next step is to add a `CodeCommitPermission` value that represents the
[notification settings](https://docs.aws.amazon.com/codecommit/latest/userguide/how-to-notify.html).

```go
repositoryName := gocf.String("MyTestRepository")
codeCommitPermission := sparta.CodeCommitPermission{
  BasePermission: sparta.BasePermission{
    SourceArn: repositoryName,
  },
  RepositoryName: repositoryName.String(),
  Branches:       branches, // may be nil
  Events:         events,   // may be nil
}
```

The `sparta.CodeCommitPermission` struct provides fields that proxy the
[RepositoryTrigger](https://docs.aws.amazon.com/codecommit/latest/APIReference/API_RepositoryTrigger.html)
values.

## Add Permission

With the subscription information configured, the final step is to
add the `sparta.CodeCommitPermission` to our `sparta.LambdaAWSInfo` value:

```go
lambdaFn.Permissions = append(lambdaFn.Permissions, codeCommitPermission)
```

The entire function is therefore:

```go
func appendCodeCommitHandler(api *sparta.API,
  lambdaFunctions []*sparta.LambdaAWSInfo) []*sparta.LambdaAWSInfo {

  lambdaFn, _ := sparta.NewAWSLambda(sparta.LambdaName(echoCodeCommit),
    echoCodeCommit,
    sparta.IAMRoleDefinition{})

    repositoryName := gocf.String("MyTestRepository")
    codeCommitPermission := sparta.CodeCommitPermission{
      BasePermission: sparta.BasePermission{
        SourceArn: repositoryName,
      },
      RepositoryName: repositoryName.String(),
    }
  lambdaFn.Permissions = append(lambdaFn.Permissions, codeCommitPermission)
  return append(lambdaFunctions, lambdaFn)
}
```

# Wrapping Up

With the `lambdaFn` fully defined, we can provide it to `sparta.Main()` and
deploy our service.  The workflow below is shared by all
CodeCmmit-triggered lambda functions:

* Define the lambda function (`echoCodeCommit`).
* If needed, create the required [IAMRoleDefinition](https://godoc.org/github.com/mweagle/Sparta*IAMRoleDefinition) with appropriate privileges.
* Provide the lambda function & IAMRoleDefinition to `sparta.NewAWSLambda()`
* Create a [CodeCommitPermission](https://godoc.org/github.com/mweagle/Sparta#CodeCommitPermission) value.
* Define the necessary permission fields.
* Append the `CodeCommitPermission` value to the lambda function's `Permissions` slice.
* Include the reference in the call to `sparta.Main()`.

## Other Resources

* Consider the [archectype](https://gosparta.io/reference/archetypes/codecommit/) package to encapsulate these steps.
* Use the AWS CLI to inspect the configured triggers:

  ```bash
  $ aws codecommit get-repository-triggers --repository-name=TestCodeCommitRepo

  {
      "configurationId": "7dd7933a-a26c-4514-9ab8-ad8cc133f874",
      "triggers": [
          {
              "name": "MyHelloWorldStack-mweagle_main_echoCodeCommit",
              "destinationArn": "arn:aws:lambda:us-west-2:123412341234:function:MyHelloWorldStack-mweagle_main_echoCodeCommit",
              "branches": [],
              "events": [
                  "all"
              ]
          }
      ]
  }
  ```

  * Use the AWS CLI to test the configured trigger:

  ```bash
  $ aws codecommit test-repository-triggers --repository-name TestCodeCommitRepo --triggers name=MyHelloWorldStack-mweagle-MyHelloWorldStack-mweagle_main_echoCodeCommit,destinationArn=arn:aws:lambda:us-west-2:123412341234:function:MyHelloWorldStack-mweagle_main_echoCodeCommit,branches=mainline,preprod,events=all

  {
    "successfulExecutions": [
        "MyHelloWorldStack-mweagle-MyHelloWorldStack-mweagle_main_echoCodeCommit"
    ],
    "failedExecutions": []
  }

  ```

