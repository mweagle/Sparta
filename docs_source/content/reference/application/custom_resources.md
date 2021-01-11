---
date: 2016-03-09T19:56:50+01:00
title: Custom Resources
weight: 150
---

In some circumstances your service may need to provision or access resources that fall outside the standard workflow. In this case you can use [CloudFormation Lambda-backed CustomResources](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/template-custom-resources-lambda.html) to create or access resources during your CloudFormation stack's lifecycle.

Sparta provides unchecked access to the CloudFormation resource lifecycle via the [RequireCustomResource](https://godoc.org/github.com/mweagle/Sparta#LambdaAWSInfo.RequireCustomResource) function. This function registers an AWS Lambda Function as an CloudFormation custom resource [lifecycle](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/crpg-ref-requesttypes.html).

In this section we'll walk through a sample user-defined custom resource and discuss how a custom resource's outputs can be propagated to an application-level Sparta lambda function.

## Components

Defining a custom resource is a two stage process, depending on whether your application-level lambda function requires access to the custom resource outputs:

1. The user-defined AWS Lambda Function


    - This function defines your resource's logic.  The multiple return value is `map[string]interface{}, error` which signify resource results and operation error, respectively.

1. The `LambdaAWSInfo` struct which declares a dependency on your custom resource via the [RequireCustomResource](https://godoc.org/github.com/mweagle/Sparta#LambdaAWSInfo.RequireCustomResource) member function.
1. _Optional_ - A call to _github.com/mweagle/Sparta/aws/cloudformation/resources.SendCloudFormationResponse_ to signal CloudFormation creation status.
1. _Optional_ - The template decorator that binds your CustomResource's [data results](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/crpg-ref-responses.html) to the owning `LambdaAWSInfo` caller.
1. _Optional_ - A call from your standard Lambda's function body to discover the CustomResource outputs via `sparta.Discover()`.

### Custom Resource Functioon

A Custom Resource Function is a standard AWS Lambda Go function type that
accepts a `CloudFormationLambdaEvent` input type. This type holds all information
for the requested [operation](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/template-custom-resources-lambda.html).

The multiple return values denote success with non-empty results, or an error.

As an example, we'll use the following custom resource function:

```go
import (
	awsLambdaCtx "github.com/aws/aws-lambda-go/lambdacontext"
	spartaCFResources "github.com/mweagle/Sparta/aws/cloudformation/resources"
)

// User defined λ-backed CloudFormation CustomResource
func userDefinedCustomResource(ctx context.Context,
	event spartaCFResources.CloudFormationLambdaEvent) (map[string]interface{}, error) {

	logger, _ := ctx.Value(ContextKeyLogger).(*zerolog.Logger)
	lambdaCtx, _ := awsLambdaCtx.FromContext(ctx)

	var opResults = map[string]interface{}{
		"CustomResourceResult": "Victory!",
	}

	opErr := spartaCFResources.SendCloudFormationResponse(lambdaCtx,
		&event,
		opResults,
		nil,
		logger)
	return opResults, opErr
}
```

This function always succeeds and publishes a non-empty map consisting of a single key (`CustomResourceResult`)
to CloudFormation. This value can be accessed by other CloudFormation resources.

### RequireCustomResource

The next step is to associate this custom resource function with a previously created Sparta `LambdaAWSInfo` instance via [RequireCustomResource](https://godoc.org/github.com/mweagle/Sparta#LambdaAWSInfo.RequireCustomResource). This function accepts:

- `roleNameOrIAMRoleDefinition`: The IAM role name or definition under which the custom resource function should be executed. Equivalent to the same argument in [NewAWSLambda](https://godoc.org/github.com/mweagle/Sparta#NewAWSLambda).
- `userFunc`: Custom resource function handler
- `lambdaOptions`: Lambda execution options. Equivalent to the same argument in [NewAWSLambda](https://godoc.org/github.com/mweagle/Sparta#NewAWSLambda).
- `resourceProps`: Arbitrary, optional properties that will be provided to the `userFunc` during execution.

The multiple return values denote the logical, stable CloudFormation resource ID of the new custom resource, or an error if one occurred.

For example, our custom resource function above can be associated via:

```go
// Standard AWS λ function
func helloWorld(ctx context.Context) (string, error) {
  return "Hello World", nil
}

func ExampleLambdaAWSInfo_RequireCustomResource() {
  lambdaFn, _ := sparta.NewAWSLambda(sparta.LambdaName(helloWorld),
    helloWorld,
    sparta.IAMRoleDefinition{})

  cfResName, _ := lambdaFn.RequireCustomResource(IAMRoleDefinition{},
    userDefinedCustomResource,
    nil,
    nil)
}
```

Since our custom resource function doesn't require any additional AWS resources, we provide an empty [IAMRoleDefinition](https://godoc.org/github.com/mweagle/Sparta#IAMRoleDefinition).

These two steps are sufficient to include your custom resource function in the CloudFormation stack lifecycle.

It's possible to share state from the custom resource to a standard Sparta lambda function by annotating your Sparta lambda function's metadata and then discovering it at execution time.

### Optional - Template Decorator

To link these resources together, the first step is to include a [TemplateDecorator](https://godoc.org/github.com/mweagle/Sparta#TemplateDecorator) that annotates your Sparta lambda function's CloudFormation resource metadata. This function specifies which user defined output keys (`CustomResourceResult` in this example) you wish to make available during your lambda function's execution.

```go
lambdaFn.Decorator = func(serviceName string,
	lambdaResourceName string,
	lambdaResource gocf.LambdaFunction,
	resourceMetadata map[string]interface{},
	S3Bucket string,
	S3Key string,
	buildID string,
	cfTemplate *gocf.Template,
	context map[string]interface{},
	logger *zerolog.Logger)  error {

  // Pass CustomResource outputs to the λ function
  resourceMetadata["CustomResource"] = gocf.GetAtt(cfResName, "CustomResourceResult")
  return nil
}
```

The `cfResName` value is the CloudFormation resource name returned by `RequireCustomResource`. The template decorator specifies which of your [CustomResourceFunction](https://godoc.org/github.com/mweagle/Sparta#CustomResourceFunction) outputs should be discoverable during the paren't lambda functions execution time through a [go-cloudformation](https://godoc.org/github.com/crewjam/go-cloudformation#GetAtt) version of [Fn::GetAtt](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/intrinsic-function-reference-getatt.html).

### Optional - Discovery

Discovery is handled by [sparta.Discover()](https://godoc.org/github.com/mweagle/Sparta#Discover) which returns a [DiscoveryInfo](https://godoc.org/github.com/mweagle/Sparta#DiscoveryInfo) instance pointer containing the linked Custom Resource outputs. The calling Sparta lambda function can discover its own [DiscoveryResource](https://godoc.org/github.com/mweagle/Sparta#DiscoveryResource) keyname via the top-level `ResourceID` field. Once found, the calling function then looks up the linked custom resource output via the `Properties` field using the keyname (`CustomResource`) provided in the previous template decorator.

In this example, the unmarshalled _DiscoveryInfo_ struct looks like:

```json
{
  "Discovery": {
    "ResourceID": "mainhelloWorldLambda837e49c53be175a0f75018a148ab6cd22841cbfb",
    "Region": "us-west-2",
    "StackID": "arn:aws:cloudformation:us-west-2:123412341234:stack/SpartaHelloWorld/70b28170-13f9-11e6-b0c7-50d5ca11b8d2",
    "StackName": "SpartaHelloWorld",
    "Resources": {
      "mainhelloWorldLambda837e49c53be175a0f75018a148ab6cd22841cbfb": {
        "ResourceID": "mainhelloWorldLambda837e49c53be175a0f75018a148ab6cd22841cbfb",
        "Properties": {
          "CustomResource": "Victory!"
        },
        "Tags": {}
      }
    }
  },
  "level": "info",
  "msg": "Custom resource request",
  "time": "2016-05-07T14:13:37Z"
}
```

To lookup the output, the calling function might do something like:

```go
configuration, _ := sparta.Discover()
customResult := configuration.Resources[configuration.ResourceID].Properties["CustomResourceResult"]
```

## Wrapping Up

CloudFormation Custom Resources are a powerful tool that can help pre-existing applications migrate to a Sparta application.

# Notes

- Sparta uses [Lambda-backed CustomResource](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/template-custom-resources-lambda.html) functions, so they are subject to the same [Lambda limits](http://docs.aws.amazon.com/lambda/latest/dg/limits.html) as application-level Sparta lambda functions.
- Returning an error from the CustomResourceFunction will result in a _FAILED_ reason being returned in the CloudFormation [response object](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/crpg-ref-responses.html).
