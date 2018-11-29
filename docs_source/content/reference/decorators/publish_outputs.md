---
date: 2018-01-22 21:49:38
title: Publishing Outputs
weight: 10
alwaysopen: false
---

CloudFormation stack [outputs](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/outputs-section-structure.html) can be used to advertise information about a service.

Sparta provides different publishing output decorators depending on the type of CloudFormation [resource output](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/walkthrough-crossstackref.html):

- `Ref`: [PublishRefOutputDecorator](https://godoc.org/github.com/mweagle/Sparta/decorator#PublishRefOutputDecorator)
- `Fn::Att`: [PublishAttOutputDecorator](https://godoc.org/github.com/mweagle/Sparta/decorator#PublishAttOutputDecorator)

## Publishing Resource Ref Values

For example, to publish the dynamically lambda resource name for a given AWS Lambda function, use
[PublishRefOutputDecorator](https://godoc.org/github.com/mweagle/Sparta/decorator#PublishRefOutputDecorator)
such as:

```go

lambdaFunctionName := "Hello World"
lambdaFn := sparta.HandleAWSLambda(lambdaFunctionName,
  helloWorld,
  sparta.IAMRoleDefinition{})

lambdaFn.Decorators = append(lambdaFn.Decorators,
  spartaDecorators.PublishRefOutputDecorator(fmt.Sprintf("%s FunctionName", lambdaFunctionName),
    fmt.Sprintf("%s Lambda ARN", lambdaFunctionName)))
}
```


## Publishing Resource Att Values

For example, to publish the dynamically determined ARN for a given AWS Lambda function, use
[PublishAttOutputDecorator](https://godoc.org/github.com/mweagle/Sparta/decorator#PublishAttOutputDecorator)
such as:

```go
lambdaFunctionName := "Hello World"
lambdaFn := sparta.HandleAWSLambda(lambdaFunctionName,
  helloWorld,
  sparta.IAMRoleDefinition{})

lambdaFn.Decorators = append(lambdaFn.Decorators,
  spartaDecorators.PublishAttOutputDecorator(fmt.Sprintf("%s FunctionARN", lambdaFunctionName),
    fmt.Sprintf("%s Lambda ARN", lambdaFunctionName), "Arn"))
}
```