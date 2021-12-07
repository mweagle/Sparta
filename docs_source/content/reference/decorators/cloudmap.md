---
date: 2019-05-27 22:31:25
title: CloudMap Service Discovery
weight: 10
alwaysopen: false
---

The [CloudMapServiceDecorator](https://godoc.org/github.com/mweagle/Sparta/decorator#CloudMapServiceDecorator) allows your service to register a [service instance](https://docs.aws.amazon.com/cloud-map/latest/dg/working-with-instances.html) for your application.

For example, an application that provisions a SQS queue and an AWS Lambda function that consumes messages from that queue may need a way for the Lambda function to discover the dynamically provisioned queue.

Sparta supports an environment-based [discovery service](http://gosparta.io/reference/discovery/) but that discovery is limited to a single Service.

The `CloudMapServiceDecorator` leverages the [CloudMap](https://aws.amazon.com/cloud-map/) service to support intra- and inter-service resource discovery.

## Definition

The first step is to create an instance of the `CloudMapServiceDecorator` type that can be used to register additional resources.

```go
import (
  spartaDecorators "github.com/mweagle/Sparta/v3/decorator"
)

func main() {

  ...
  cloudMapDecorator, cloudMapDecoratorErr := spartaDecorators.NewCloudMapServiceDecorator(gocf.String("SpartaServices"),
    gocf.String("SpartaSampleCloudMapService"))
...
}
```

The first argument is the [Cloud Map Namespace ID](https://docs.aws.amazon.com/cloud-map/latest/dg/working-with-namespaces.html) value to which the service (_MyService_) will publish.

The decorator satisfies the [ServiceDecoratorHookHandler](https://godoc.org/github.com/mweagle/Sparta#ServiceDecoratorHookHandler). The instance should be provided as a `WorkflowHooks.ServiceDecorators` element to `MainEx` as in:

```go
func main() {
  // ...
  cloudMapDecorator, cloudMapDecoratorErr := spartaDecorators.NewCloudMapServiceDecorator(gocf.String("SpartaServices"),
    gocf.String("SpartaSampleCloudMapService"))

  workflowHooks := &sparta.WorkflowHooks{
    ServiceDecorators: []sparta.ServiceDecoratorHookHandler{
      cloudMapDecorator,
    },
  }

  // ...
  err := sparta.MainEx(awsName,
    "Simple Sparta application that demonstrates core functionality",
    lambdaFunctions,
    nil,
    nil,
    workflowHooks,
    false)
}
```

## Registering

The returned `CloudMapServiceDecorator` instance satisfies the [ServiceDecoratorHookHandler](https://godoc.org/github.com/mweagle/Sparta#ServiceDecoratorHookHandler) interface. When invoked, it updates the content of you CloudFormation template with the resources and permissions as described below. `CloudMapServiceDecorator` implicitly creates a new [AWS::ServiceDiscovery::Service](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-servicediscovery-service.html) resource to which your resources will be published.

### Lambda Functions

The `CloudMapServiceDecorator.PublishLambda` function publishes Lambda function information to the _(NamespaceID, ServiceID)_ pair.

```go
lambdaFn, _ := sparta.NewAWSLambda("Hello World",
    helloWorld,
    sparta.IAMRoleDefinition{})
cloudMapDecorator.PublishLambda("lambdaDiscoveryName", lambdaFn, nil)
```

The default properties published include the [Lambda Outputs](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-lambda-function.html) and Type information:

```json
{
  "Id": "CloudMapResbe2b7c536074312c-VuIPjfjuFaoc",
  "Attributes": {
    "Arn": "arn:aws:lambda:us-west-2:123412341234:function:MyHelloWorldStack-123412341234_Hello_World",
    "Name": "lambdaDiscoveryName",
    "Ref": "MyHelloWorldStack-123412341234_Hello_World",
    "Type": "AWS::Lambda::Function"
  }
}
```

### Other Resources

The `CloudMapServiceDecorator.PublishResource` function publishes arbitrary CloudFormation resource outputs information to the _(NamespaceID, ServiceID)_ pair.

For instance, to publish SQS information in the context of a standard `ServiceDecorator`

```go
func createSQSResourceDecorator(cloudMapDecorator *spartaDecorators.CloudMapServiceDecorator) sparta.ServiceDecoratorHookHandler {
  return sparta.ServiceDecoratorHookFunc(func(context map[string]interface{},
    serviceName string,
    template *gocf.Template,
    S3Bucket string,
    S3Key string,
    buildID string,
    awsSession *session.Session,
    noop bool,
    logger *zerolog.Logger) error {

    sqsResource := &gocf.SQSQueue{}
    template.AddResource("SQSResource", sqsResource)
    return cloudMapDecorator.PublishResource("SQSResource",
      "SQSResource",
      sqsResource,
      nil)
  })
}
```

The default properties published include the [SQS Outputs](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-sqs-queues.html) and Type information:

```json
{
  "Id": "CloudMapRes21cf275e8bbbe136-CqWZ27gdLHf8",
  "Attributes": {
    "Arn": "arn:aws:sqs:us-west-2:123412341234:MyHelloWorldStack-123412341234-SQSResource-S9DWKIFKP14U",
    "Name": "SQSResource",
    "QueueName": "MyHelloWorldStack-123412341234-SQSResource-S9DWKIFKP14U",
    "Ref": "https://sqs.us-west-2.amazonaws.com/123412341234/MyHelloWorldStack-123412341234-SQSResource-S9DWKIFKP14U",
    "Type": "AWS::SQS::Queue"
  }
}
```

## Enabling

Publishing instances to CloudMap only makes them available for other services to discover them. Call the `EnableDiscoverySupport` with your `*sparta.LambdaAWSInfo` instance as the only argument. This function updates your Lambda function's environment to include the provisioned _ServiceInstance_ and also the IAM role privileges to authorize:

- _servicediscovery:DiscoverInstances_
- _servicediscovery:GetNamespace_
- _servicediscovery:ListInstances_
- _servicediscovery:GetService_

For instance:

```go
func main() {
  // ...

  lambdaFn, _ := sparta.NewAWSLambda("Hello World",
      helloWorld,
      sparta.IAMRoleDefinition{})

  // ...

  cloudMapDecorator.EnableDiscoverySupport(lambdaFn)

  // ...
}

```

## Invoking

With the resources published and the lambda role properly updated, the last step is to dynamically discover the provisioned resources via CloudMap. Call `DiscoverInstancesWithContext` with the the set of key-value pairs to use for discovery as below:

```go

func helloWorld(ctx context.Context) (string, error) {
    // ...

  props := map[string]string{
    "Type": "AWS::SQS::Queue",
  }
  results, resultsErr := spartaDecorators.DiscoverInstancesWithContext(ctx,
    props,
    logger)

  logger.Info().
    Interface("Instances", results).
    Err(resultsErr).
    Msg("Discovered instances")

    // ...
}
```

Given the previous example of a single lambda function and an SQS-queue provisioning decorator, the `DiscoverInstancesWithContext` would return the matching instance with data similar to:

```json
{
  "Instances": [
    {
      "Attributes": {
        "Arn": "arn:aws:sqs:us-west-2:123412341234:MyHelloWorldStack-123412341234-SQSResource-S9DWKIFKP14U",
        "Name": "SQSResource",
        "QueueName": "MyHelloWorldStack-123412341234-SQSResource-S9DWKIFKP14U",
        "Ref": "https://sqs.us-west-2.amazonaws.com/123412341234/MyHelloWorldStack-123412341234-SQSResource-S9DWKIFKP14U",
        "Type": "AWS::SQS::Queue"
      },
      "HealthStatus": "HEALTHY",
      "InstanceId": "CloudMapResd1a507076543ccd0-Fln1ITi5cf0y",
      "NamespaceName": "SpartaServices",
      "ServiceName": "SpartaSampleCloudMapService"
    }
  ]
}
```
