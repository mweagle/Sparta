---
date: 2018-01-22 21:49:38
title: Concepts
description: Core Sparta Concepts
weight: 10
alwaysopen: false
---

This is a brief overview of Sparta's core concepts.  Additional information regarding specific features is available from the menu.

# Terms and Concepts

At a high level, Sparta transforms a **go** binary's registered lambda functions into a set of independently addressable AWS Lambda functions .  Additionally, Sparta provides microservice authors an opportunity to satisfy other requirements such as defining the IAM Roles under which their function will execute in AWS, additional infrastructure requirements, and telemetry and alerting information (via CloudWatch).

## Service Name

Sparta applications are deployed as a single unit, using the **ServiceName** as a stable logical identifier.  The **ServiceName** is used as your application's [CloudFormation StackName](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/pseudo-parameter-reference.html)

```go
stackName := "MyUniqueServiceName"
sparta.Main(stackName,
  "Simple Sparta application",
  myLambdaFunctions,
  nil,
  nil)
```

## Lambda Function

A Sparta-compatible lambda is a standard [AWS Lambda Go](https://docs.aws.amazon.com/lambda/latest/dg/go-programming-model-handler-types.html) function. The following function signatures are supported:

* `func ()`
* `func () error`
* `func (TIn), error`
* `func () (TOut, error)`
* `func (context.Context) error`
* `func (context.Context, TIn) error`
* `func (context.Context) (TOut, error)`
* `func (context.Context, TIn) (TOut, error)`

where the `TIn` and `TOut` parameters represent [encoding/json](https://golang.org/pkg/encoding/json) un/marshallable types.  Supplying an invalid signature will produce a run time error as in:

{{< highlight text >}}
ERRO[0000] Lambda function (Hello World) has invalid returns: handler
returns a single value, but it does not implement error exit status 1
{{< /highlight >}}

## Privileges

To support accessing other AWS resources in your **go** function, Sparta allows you to define and link [IAM Roles](http://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles.html) with narrowly defined [sparta.IAMRolePrivilege](https://godoc.org/github.com/mweagle/Sparta#IAMRolePrivilege) values. This allows you to define the _minimal_ set of privileges under which your **go** function will execute.  The `Privilege.Resource` field value may also be a [StringExpression](https://godoc.org/github.com/mweagle/go-cloudformation#StringExpr) referencing a dynamically provisioned CloudFormation resource.

```go
lambdaFn.RoleDefinition.Privileges = append(lambdaFn.RoleDefinition.Privileges,
  sparta.IAMRolePrivilege{
    Actions:  []string{"s3:GetObject", "s3:HeadObject"},
    Resource: "arn:aws:s3:::MyS3Bucket",
})
```

## Permissions

To configure AWS Lambda [Event Sources](http://docs.aws.amazon.com/lambda/latest/dg/intro-core-components.html), Sparta provides both [sparta.LambdaPermission](https://godoc.org/github.com/mweagle/Sparta#LambdaPermission) and service-specific _Permission_ types like [sparta.CloudWatchEventsPermission](https://godoc.org/github.com/mweagle/Sparta#CloudWatchEventsPermission). The service-specific _Permission_ types automatically register your lambda function with the remote AWS service, using each service's specific API.

```go
cloudWatchEventsPermission := sparta.CloudWatchEventsPermission{}
cloudWatchEventsPermission.Rules = make(map[string]sparta.CloudWatchEventsRule, 0)
cloudWatchEventsPermission.Rules["Rate5Mins"] = sparta.CloudWatchEventsRule{
  ScheduleExpression: "rate(5 minutes)",
}
lambdaFn.Permissions = append(lambdaFn.Permissions, cloudWatchEventsPermission)
```

## Decorators

Decorators are associated with either [Lambda functions](https://godoc.org/github.com/mweagle/Sparta#TemplateDecoratorHandler) or
the larger service workflow via [WorkflowHooks](https://godoc.org/github.com/mweagle/Sparta#WorkflowHooks). They are user-defined
functions that provide an opportunity for your service to perform secondary actions such as automatically generating a
[CloudFormation Dashboard](https://godoc.org/github.com/mweagle/Sparta/decorator#DashboardDecorator) or automatically publish
an [S3 Artifact](https://godoc.org/github.com/mweagle/Sparta/decorator#S3ArtifactPublisherDecorator) from your service.

Decorators are applied at `provision` time.

## Interceptors

Interceptors are the runtime analog to Decorators. They are user-defined functions that are executed in the
context of handling an event. They provide an opportunity for you to support cross-cutting concerns such as automatically
registering [XRayTraces](https://godoc.org/github.com/mweagle/Sparta/interceptor#RegisterXRayInterceptor) that can capture
service performance and log messages in the event of an error.

{{< interceptorflow >}}

## Dynamic Resources

Sparta applications can specify other [AWS Resources](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-template-resource-type-ref.html) (eg, _SNS Topics_) as part of their application. The dynamic resource outputs can be referenced by Sparta lambda functions via [gocf.Ref](https://godoc.org/github.com/mweagle/go-cloudformation#Ref) and [gocf.GetAtt](https://godoc.org/github.com/mweagle/go-cloudformation#GetAtt) functions.

```go
snsTopicName := sparta.CloudFormationResourceName("SNSDynamicTopic")
snsTopic := &gocf.SNSTopic{
  DisplayName: gocf.String("Sparta Application SNS topic"),
})
lambdaFn, _ := sparta.NewAWSLambda(sparta.LambdaName(echoDynamicSNSEvent),
  echoDynamicSNSEvent,
  sparta.IAMRoleDefinition{})

lambdaFn.Permissions = append(lambdaFn.Permissions, sparta.SNSPermission{
  BasePermission: sparta.BasePermission{
    SourceArn: gocf.Ref(snsTopicName),
  },
})
```

## Discovery

To support Sparta lambda functions discovering dynamically assigned AWS resource values, Sparta provides [sparta.Discover](https://godoc.org/github.com/mweagle/Sparta#Discover). This function returns information about resources that a given
entity specifies a [DependsOn](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-attribute-dependson.html) relationship.

```go
func echoS3DynamicBucketEvent(ctx context.Context,
  s3Event awsLambdaEvents.S3Event) (*awsLambdaEvents.S3Event, error) {

  discoveryInfo, discoveryInfoErr := sparta.Discover()
  logger.WithFields(logrus.Fields{
    "Event":        s3Event,
    "Discovery":    discoveryInfo,
    "DiscoveryErr": discoveryInfoErr,
  }).Info("Event received")

  // Use discoveryInfo to determine the bucket name to which RawMessage should be stored
  ...
}
```

# Summary

Given a set of registered Sparta lambda function, a typical `provision` build to create a new service follows this workflow. Items with dashed borders are opt-in user behaviors.

{{< spartaflow >}}

During provisioning, Sparta uses [AWS Lambda-backed Custom Resources](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/template-custom-resources-lambda.html) to support operations for which CloudFormation doesn't yet support (eg, [API Gateway](https://aws.amazon.com/api-gateway/) creation).

# Next Steps

Walk through a starting [Sparta Application](/sample_service/).
