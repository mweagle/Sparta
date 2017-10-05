---
date: 2016-03-09T19:56:50+01:00
title: Dynamic Infrastructure
weight: 10
menu:
  main:
    parent: Documentation
    identifier: dynamic-infrastructure
    weight: 50
---

# Introduction

In addition to provisioning AWS Lambda functions, Sparta supports the creation of other [CloudFormation Resources](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-template-resource-type-ref.html).  This enables a service to move towards [immutable infrastructure](https://fugue.co/oreilly/), where the service and its infrastructure requirements are treated as a logical unit.

For instance, consider the case where two developers are working in the same AWS account.

  * Developer 1 is working on analyzing text documents.
    - Their lambda code is triggered in response to uploading sample text documents to S3.
  * Developer 2 is working on image recognition.
    - Their lambda code is triggered in response to uploading sample images to S3.

{{< mermaid >}}
graph LR
  sharedBucket[S3 Bucket]

  dev1Lambda[Dev1 LambdaCode]
  dev2Lambda[Dev2 LambdaCode]

  sharedBucket --> dev1Lambda
  sharedBucket --> dev2Lambda
{{< /mermaid >}}

Using a shared, externally provisioned S3 bucket has several impacts:

  * Adding conditionals in each lambda codebase to scope valid processing targets.
  * Ambiguity regarding which codebase handled an event.
  * Infrastructure ownership/lifespan management.  When a service is decommissioned, its infrastructure requirements may be automatically decommissioned as well.
    - Eg, "Is this S3 bucket in use by any service?".
  * Overly permissive IAM roles due to static Arns.
    - Eg, "Arn hugging".
  * Contention updating the shared bucket's [notification configuration](http://docs.aws.amazon.com/AWSJavaScriptSDK/latest/AWS/S3.html#putBucketNotificationConfiguration-property).

Alternatively, each developer could provision and manage disjoint topologies:

{{< mermaid >}}
graph LR
  dev1S3Bucket[Dev1 S3 Bucket]
  dev1Lambda[Dev1 LambdaCode]

  dev2S3Bucket[Dev2 S3 Bucket]
  dev2Lambda[Dev2 LambdaCode]

  dev1S3Bucket --> dev1Lambda
  dev2S3Bucket --> dev2Lambda
{{< /mermaid >}}

Enabling each developer to create other AWS resources also means more complex topologies can be expressed.  These topologies can benefit from CloudWatch monitoring (eg, [per-Lambda Metrics](http://docs.aws.amazon.com/lambda/latest/dg/monitoring-functions-metrics.html) ) without the need to add [custom metrics](http://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/publishingMetrics.html).

{{< mermaid >}}
graph LR
  dev1S3Bucket[Dev1 S3 Bucket]
  dev1Lambda[Dev1 LambdaCode]

  dev2S3Bucket[Dev2 S3 Images Bucket]
  dev2PNGLambda[Dev2 PNG LambdaCode]
  dev2JPGLambda[Dev2 JPEG LambdaCode]
  dev2TIFFLambda[Dev2 TIFF LambdaCode]
  dev2S3VideoBucket[Dev2 VideoBucket]
  dev2VideoLambda[Dev2 Video LambdaCode]

  dev1S3Bucket --> dev1Lambda
  dev2S3Bucket -->|SuffixFilter=*.PNG|dev2PNGLambda
  dev2S3Bucket -->|SuffixFilter=*.JPEG,*.JPG|dev2JPGLambda
  dev2S3Bucket -->|SuffixFilter=*.TIFF|dev2TIFFLambda
  dev2S3VideoBucket -->dev2VideoLambda
{{< /mermaid >}}

Sparta supports Dynamic Resources via [TemplateDecorator](https://godoc.org/github.com/mweagle/Sparta#TemplateDecorator) functions.

# Template Decorators

A template decorator is a **go** function with the following signature

{{< highlight go >}}

type TemplateDecorator func(serviceName string,
	lambdaResourceName string,
	lambdaResource gocf.LambdaFunction,
	resourceMetadata map[string]interface{},
	S3Bucket string,
	S3Key string,
	buildID string,
	cfTemplate *gocf.Template,
	context map[string]interface{},
	logger *logrus.Logger)  error {

}
{{< /highlight >}}

Clients use [go-cloudformation](https://godoc.org/github.com/crewjam/go-cloudformation) types for CloudFormation resources and  `template.AddResource` to add them to the `*template` parameter.  After a decorator is invoked, Sparta verifies that the user-supplied function has not produced entities that collide with the internally-generated ones.

## Unique Resource Names

CloudFormation uses [Logical IDs](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/resources-section-structure.html) as resource key names.

To minimize collision likelihood, Sparta publishes [CloudFormationResourceName(prefix, ...parts)](https://godoc.org/github.com/mweagle/Sparta#CloudFormationResourceName) to generate compliant identifiers.  To produce content-based hash values, callers can provide a non-empty set of values as the `...parts` variadic argument.  This produces stable identifiers across Sparta execution (which may affect availability during updates).

When called with only a single value (eg: `CloudFormationResourceName("myResource")`), Sparta will return a random resource name that is **NOT** stable across executions.

# Example - S3 Bucket

Let's work through an example to make things a bit more concrete.  We have the following requirements:

  * Our lambda function needs a immutable-infrastructure compliant S3 bucket
  * Our lambda function should be notified when items are created or deleted from the bucket
  * Our lambda function must be able to access the contents in the bucket (not shown below)

## Lambda Function

To start with, we'll need a Sparta lambda function to expose:

{{< highlight go >}}
func echoS3DynamicBucketEvent(w http.ResponseWriter, r *http.Request) {
	logger, _ := r.Context().Value(sparta.ContextKeyLogger).(*logrus.Logger)
	lambdaContext, _ := r.Context().Value(sparta.ContextKeyLambdaContext).(*sparta.LambdaContext)
	eventData, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	config, _ := sparta.Discover()
	logger.WithFields(logrus.Fields{
		"RequestID":     lambdaContext.AWSRequestID,
		"Event":         string(eventData),
		"Configuration": config,
	}).Info("Request received")

	w.Header().Set("Content-Type", "application/json")
	w.Write(eventData)
}
{{< /highlight >}}

For brevity our demo function doesn't access the S3 bucket objects.  To support that we'll need to discuss the `sparta.Discover` function in another [section](/docs/discovery/).

## S3 Resource Name

The next thing we need is a _Logical ID_ for our bucket:

{{< highlight go >}}

s3BucketResourceName := sparta.CloudFormationResourceName("S3DynamicBucket", "myServiceBucket")
{{< /highlight >}}

## Sparta Integration

With these two values we're ready to get started building up the lambda function:

{{< highlight go >}}
lambdaFn := sparta.HandleAWSLambda(sparta.LambdaName(echoS3DynamicBucketEvent),
  http.HandlerFunc(echoS3DynamicBucketEvent),
  sparta.IAMRoleDefinition{})
{{< /highlight >}}

The open issue is how to publish the CloudFormation-defined S3 Arn to the `compile`-time application.  Our lambda function needs to provide both:

  * [IAMRolePrivilege](https://godoc.org/github.com/mweagle/Sparta#IAMRolePrivilege) values that reference the (as yet) undefined Arn.
  * [S3Permission](https://godoc.org/github.com/mweagle/Sparta#S3Permission) values to configure our lambda's event triggers on the (as yet) undefined Arn.

The missing piece is [gocf.Ref()](https://godoc.org/github.com/crewjam/go-cloudformation#Ref), whose single argument is the _Logical ID_ of the S3 resource we'll be inserting in the decorator call.

### Dynamic IAM Role Privilege

The `IAMRolePrivilege` struct references the dynamically assigned S3 Arn as follows:

{{< highlight go >}}

lambdaFn.Permissions = append(lambdaFn.Permissions, sparta.S3Permission{
  BasePermission: sparta.BasePermission{
    SourceArn: gocf.Ref(s3BucketResourceName),
  },
  Events: []string{"s3:ObjectCreated:*", "s3:ObjectRemoved:*"},
})
lambdaFn.DependsOn = append(lambdaFn.DependsOn, s3BucketResourceName)
{{< /highlight >}}

### Dynamic S3 Permissions

The `S3Permission` struct also requires the dynamic Arn, to which it will append `"/*"` to enable object read access.

{{< highlight go >}}

lambdaFn.RoleDefinition.Privileges = append(lambdaFn.RoleDefinition.Privileges,
  sparta.IAMRolePrivilege{
    Actions:  []string{"s3:GetObject", "s3:HeadObject"},
    Resource: spartaCF.S3AllKeysArnForBucket(gocf.Ref(s3BucketResourceName)),
  })
{{< /highlight >}}

The `spartaCF.S3AllKeysArnForBucket` call is a convenience wrapper around [gocf.Join](https://godoc.org/github.com/crewjam/go-cloudformation#Join) to generate the concatenated, dynamic Arn expression.

## S3 Resource Insertion

All that's left to do is actually insert the S3 resource in our decorator:

{{< highlight go >}}

lambdaFn.Decorator = func(lambdaResourceName string,
                          lambdaResource gocf.LambdaFunction,
                          template *gocf.Template,
                          logger *logrus.Logger) error {

  cfResource := template.AddResource(s3BucketResourceName, &gocf.S3Bucket{
    AccessControl: gocf.String("PublicRead"),
  })
  cfResource.DeletionPolicy = "Delete"
  return nil
}
{{< /highlight >}}

### Dependencies

In reality, we shouldn't even attempt to create the AWS Lambda function if the S3 bucket creation fails.  As application developers, we can help CloudFormation sequence infrastructure operations by stating this hard dependency on the S3 bucket via the [DependsOn](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-attribute-dependson.html) attribute:

{{< highlight go >}}

lambdaFn.DependsOn = append(lambdaFn.DependsOn, s3BucketResourceName)
{{< /highlight >}}

## Code Listing

Putting everything together, our Sparta lambda function with dynamic infrastructure is listed below.

{{< highlight go >}}

s3BucketResourceName := sparta.CloudFormationResourceName("S3DynamicBucket")
lambdaFn := sparta.HandleAWSLambda(sparta.LambdaName(echoS3DynamicBucketEvent),
  http.HandlerFunc(echoS3DynamicBucketEvent),
  sparta.IAMRoleDefinition{})

// Our lambda function requires the S3 bucket
lambdaFn.DependsOn = append(lambdaFn.DependsOn, s3BucketResourceName)

// Add a permission s.t. the lambda function could read from the S3 bucket
lambdaFn.RoleDefinition.Privileges = append(lambdaFn.RoleDefinition.Privileges,
  sparta.IAMRolePrivilege{
    Actions:  []string{"s3:GetObject",
                       "s3:HeadObject"},
    Resource: spartaCF.S3AllKeysArnForBucket(gocf.Ref(s3BucketResourceName)),
  })

// Configure the S3 event source
lambdaFn.Permissions = append(lambdaFn.Permissions, sparta.S3Permission{
  BasePermission: sparta.BasePermission{
    SourceArn: gocf.Ref(s3BucketResourceName),
  },
  Events: []string{"s3:ObjectCreated:*",
                   "s3:ObjectRemoved:*"},
})

// Actually add the resource
lambdaFn.Decorator = func(lambdaResourceName string,
                          lambdaResource gocf.LambdaFunction,
                          template *gocf.Template,
                          logger *logrus.Logger) error {
  cfResource := template.AddResource(s3BucketResourceName, &gocf.S3Bucket{
    AccessControl: gocf.String("PublicRead"),
  })
  cfResource.DeletionPolicy = "Delete"
  return nil
}
{{< /highlight >}}

## Wrapping Up

Sparta provides an opportunity to bring infrastructure management into the application programming model.  It's still possible to use literal Arn strings, but the ability to include other infrastructure requirements brings a service closer to being self-contained and more operationally sustainable.

# Notes
  * The `echoS3DynamicBucketEvent` function can also access the bucket Arn via [sparta.Discover](/docs/discovery).
  * See the [DeletionPolicy](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-attribute-deletionpolicy.html) documentation regarding S3 management.
  * CloudFormation resources also publish [other outputs](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/intrinsic-function-reference-getatt.html) that can be retrieved via [gocf.GetAtt](https://godoc.org/github.com/crewjam/go-cloudformation#GetAtt).
  * `go-cloudformation` exposes [gocf.Join](https://godoc.org/github.com/crewjam/go-cloudformation#Join) to create compound, dynamic expressions.
    - See the CloudWatch docs on [Fn::Join](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/intrinsic-function-reference-join.html) for more information.
