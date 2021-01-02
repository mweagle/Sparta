---
date: 2020-12-31 21:48:47
title: Testing
weight: 800
---

## Unit Tests

While developing Sparta lambda functions it may be useful to test them locally without needing to `provision` each new code change. You can test your lambda functions
using standard `go test` functionality.

To create proper event types, consider:

- [AWS Lambda Go](https://godoc.org/github.com/aws/aws-lambda-go/events) types
- Sparta types
- Use [NewAPIGatewayMockRequest](https://godoc.org/github.com/mweagle/Sparta/aws/events#NewAPIGatewayMockRequest) to generate API Gateway style requests.

## Acceptance Tests

The _cloudtest_ package provides a BDD-style interface to represent tests
that verify Lambda behavior that is asynchronously triggered. For instance, to
verify that a direct LambdaInvocation with a known payload also produces a
known CloudWatch log output:

```go
func TestCloudLiteralLogOutputTest(t *testing.T) {
  NewTest().
    Given(NewLambdaInvokeTrigger(helloWorldJSON)).
    Against(NewLambdaLiteralSelector(fmt.Sprintf("MyOCIStack-%s_Hello_World", accountID))).
    Ensure(NewLogOutputEvaluator(regexp.MustCompile("Accessing"))).
    Run(t)
}
```

Tests can also use lambda invocation metrics using [JMESPath](http://jmespath.org) selectors against the functions [GetFunctionOutput](https://docs.aws.amazon.com/sdk-for-go/api/service/lambda/#GetFunctionOutput) JSON representation:

```go
func TestCloudMetricsTest(t *testing.T) {
  NewTest().
    Given(NewLambdaInvokeTrigger(helloWorldJSON)).
    Against(
      NewStackLambdaSelector(fmt.Sprintf("MyOCIStack-%s", accountID),
        "[Configuration][?contains(FunctionName,'Hello_World')].FunctionName | [0]")).
    Ensure(NewLambdaInvocationMetricEvaluator(DefaultLambdaFunctionMetricQueries(),
      IsSuccess),
    ).
    Run(t)
}
```

The _cloudtest_ package also provides S3, SQS, and other event source triggers.
Clients can define their own `cloudtest.Trigger` compliant instance to
extend the test scenarios.
