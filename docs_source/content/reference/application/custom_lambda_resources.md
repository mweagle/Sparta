---
date: 2016-03-09T19:56:50+01:00
title: CloudFormation Resources
weight: 150
---

In addition to per-lambda [custom resources](/reference/custom_resources/), a service may benefit from the ability to
include a service-scoped [Lambda backed CustomResource](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/template-custom-resources-lambda.html).

Including a custom service scoped resource is a multi-step process. The code excerpts below are from the [SpartaCustomResource](https://github.com/mweagle/SpartaCustomResource) sample application.

## 1. Resource Type

The first step is to define a custom [CloudFormation Resource Type](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/template-custom-resources.html)

```go

////////////////////////////////////////////////////////////////////////////////
// 1 - Define the custom type
const spartaHelloWorldResourceType = "Custom::sparta::HelloWorldResource"
```

## 2. Request Parameters

The next step is to define the parameters that are supplied to the custom resource
invocation. This is done via a `struct` that will be later embedded into the
[CustomResourceCommand](https://godoc.org/github.com/mweagle/Sparta/aws/cloudformation/resources#CustomResourceCommand).

```go
// SpartaCustomResourceRequest is what the UserProperties
// should be set to in the CustomResource invocation
type SpartaCustomResourceRequest struct {
  Message *gocf.StringExpr
}
```

## 3. Command Handler

With the parameters defined, define the
[CustomResourceCommand](https://godoc.org/github.com/mweagle/Sparta/aws/cloudformation/resources#CustomResourceCommand)
that is responsible for performing the external operations based on the specified request
parameters.

```go
// SpartaHelloWorldResource is a simple POC showing how to create custom resources
type SpartaHelloWorldResource struct {
  gocf.CloudFormationCustomResource
  SpartaCustomResourceRequest
}

// Create implements resource create
func (command SpartaHelloWorldResource) Create(awsSession *session.Session,
  event *spartaAWSResource.CloudFormationLambdaEvent,
  logger *zerolog.Logger) (map[string]interface{}, error) {

  requestPropsErr := json.Unmarshal(event.ResourceProperties, &command)
  if requestPropsErr != nil {
  return nil, requestPropsErr
  }
  logger.Info("create: ", command.Message.Literal)
  return map[string]interface{}{
  "Resource": "Created message: " + command.Message.Literal,
  }, nil
}

// Update implements resource update
func (command SpartaHelloWorldResource) Update(awsSession *session.Session,
  event *spartaAWSResource.CloudFormationLambdaEvent,
  logger *zerolog.Logger) (map[string]interface{}, error) {
  return "", nil
}

// Delete implements resource delete
func (command SpartaHelloWorldResource) Delete(awsSession *session.Session,
  event *spartaAWSResource.CloudFormationLambdaEvent,
  logger *zerolog.Logger) (map[string]interface{}, error) {
  return "", nil
}
```

## 4. Register Type Provider

To make the new type available to Sparta's internal CloudFormation template
marshalling, register the new type via [go-cloudformation.RegisterCustomResourceProvider](https://godoc.org/github.com/mweagle/go-cloudformation#RegisterCustomResourceProvider):

```go
func init() {
  customResourceFactory := func(resourceType string) gocf.ResourceProperties {
    switch resourceType {
    case spartaHelloWorldResourceType:
      return &SpartaHelloWorldResource{}
    }
    return nil
  }
  gocf.RegisterCustomResourceProvider(customResourceFactory)
}
```

## 5. Annotate Template

The final step is to ensure the custom resource command is included in the Sparta
binary that defines your service and then create an invocation of that command. The
annotation is expressed as a [ServiceDecoratorHookHandler](https://godoc.org/github.com/mweagle/Sparta#ServiceDecoratorHookHandler)
that performs both operations as part of the general service build
lifecycle...

```go
func customResourceHooks() *sparta.WorkflowHooks {
  // Add the custom resource decorator
  customResourceDecorator := func(context map[string]interface{},
    serviceName string,
    template *gocf.Template,
    S3Bucket string,
    S3Key string,
    buildID string,
    awsSession *session.Session,
    noop bool,
    logger *zerolog.Logger) error {

    // 1. Ensure the Lambda Function is registered
    customResourceName, customResourceNameErr := sparta.EnsureCustomResourceHandler(serviceName,
      spartaHelloWorldResourceType,
      nil, // This custom action doesn't need to access other AWS resources
      []string{},
      template,
      S3Bucket,
      S3Key,
      logger)

    if customResourceNameErr != nil {
      return customResourceNameErr
    }

    // 2. Create the request for the invocation of the lambda resource with
    // parameters
    spartaCustomResource := &SpartaHelloWorldResource{}
    spartaCustomResource.ServiceToken = gocf.GetAtt(customResourceName, "Arn")
    spartaCustomResource.Message = gocf.String("Custom resource activated!")

    resourceInvokerName := sparta.CloudFormationResourceName("SpartaCustomResource",
      fmt.Sprintf("%v", S3Bucket),
      fmt.Sprintf("%v", S3Key))

    // Add it
    template.AddResource(resourceInvokerName, spartaCustomResource)
    return nil
  }
  // Add the decorator to the template
  hooks := &sparta.WorkflowHooks{}
  hooks.ServiceDecorators = []sparta.ServiceDecoratorHookHandler{
    sparta.ServiceDecoratorHookFunc(customResourceDecorator),
  }
  return hooks
}
```

Provide the hooks structure to [MainEx](https://godoc.org/github.com/mweagle/Sparta#MainEx) to
include this custom resource with your service's provisioning lifecycle.
