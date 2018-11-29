---
date: 2016-03-09T19:56:50+01:00
title: Discovery Service
weight: 150
---

The ability to provision [dynamic infrastructure](/reference/dynamic_infrastructure) (see also the [SES Event Source Example](/reference/eventsources/ses/#dynamic-resources:d680e8a854a7cbad6d490c445cba2eba)) as part of a Sparta application creates a need to discover those resources at lambda execution time.

Sparta exposes this functionality via [sparta.Discover](https://godoc.org/github.com/mweagle/Sparta#Discover).  This function returns information about the current stack (eg, name, region, ID) as well as metadata about the immediate dependencies of the calling **go** lambda function.

The following sections walk through provisioning a S3 bucket, declaring an explicit dependency on that resource, and then discovering the resource at lambda execution time.  It is extracted from `appendDynamicS3BucketLambda` in the  [SpartaApplication](https://github.com/mweagle/SpartaApplication/blob/master/application.go) source.

If you haven't already done so, please review the [Dynamic Infrastructure](/reference/dynamic_infrastructure) section for background on dynamic infrastructure provisioning.


# Discovery

For reference, we provision an S3 bucket and declare an explicit dependency with the code below.  Because our `gocf.S3Bucket{}` struct uses a zero-length [BucketName](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-s3-bucket.html#cfn-s3-bucket-name) property, CloudFormation will dynamically assign one.

```go

s3BucketResourceName := sparta.CloudFormationResourceName("S3DynamicBucket")
lambdaFn := sparta.HandleAWSLambda(sparta.LambdaName(echoS3DynamicBucketEvent),
  echoS3DynamicBucketEvent,
  sparta.IAMRoleDefinition{})
lambdaFn.Permissions = append(lambdaFn.Permissions, sparta.S3Permission{
  BasePermission: sparta.BasePermission{
    SourceArn: gocf.Ref(s3BucketResourceName),
  },
  Events: []string{"s3:ObjectCreated:*", "s3:ObjectRemoved:*"},
})
lambdaFn.DependsOn = append(lambdaFn.DependsOn, s3BucketResourceName)

// Add permission s.t. the lambda function could read from the S3 bucket
lambdaFn.RoleDefinition.Privileges = append(lambdaFn.RoleDefinition.Privileges,
  sparta.IAMRolePrivilege{
    Actions:  []string{"s3:GetObject", "s3:HeadObject"},
    Resource: spartaCF.S3AllKeysArnForBucket(gocf.Ref(s3BucketResourceName)),
  })

lambdaFn.Decorator = func(serviceName string,
  lambdaResourceName string,
  lambdaResource gocf.LambdaFunction,
  resourceMetadata map[string]interface{},
  S3Bucket string,
  S3Key string,
  buildID string,
  template *gocf.Template,
  context map[string]interface{},
  logger *logrus.Logger) error {
  cfResource := template.AddResource(s3BucketResourceName, &gocf.S3Bucket{
    AccessControl: gocf.String("PublicRead"),
    Tags: &gocf.TagList{gocf.Tag{
      Key:   gocf.String("SpecialKey"),
      Value: gocf.String("SpecialValue"),
    },
    },
  })
  cfResource.DeletionPolicy = "Delete"
  return nil
}
```

The key to `sparta.Discovery` is the `DependsOn` slice value.

# Template Marshaling & Decoration

By default, Sparta uses CloudFormation to update service state.  During template marshaling, Sparta scans for `DependsOn` relationships and propagates information (immediate-children only) across CloudFormation resource definitions.  Most importantly, this information includes `Ref` and any other [outputs](https://github.com/mweagle/Sparta/blob/master/cloudformation_resources.go#L16) of referred resources.  This information then becomes available as a [DisocveryInfo](https://godoc.org/github.com/mweagle/Sparta#DiscoveryInfo) value returned by `sparta.Discovery()`. Behind the scenes, Sparta

# Sample DiscoveryInfo

In our example, a `DiscoveryInfo` from a sample stack might be:

```json
{
    "ResourceID": "mainechoS3DynamicBucketEventLambda41ca034273726cf36154137cbf8d7e5bd45f863a",
    "Region": "us-west-2",
    "StackID": "arn:aws:cloudformation:us-west-2:123412341234:stack/SpartaApplication-mweagle/d4e07d80-03eb-11e8-b6fd-50d5ca789e4a",
    "StackName": "SpartaApplication-mweagle",
    "Resources": {
        "S3DynamicBucket62b0e7a664dc29c1c4fbe231fbcef30f8463aaa3": {
            "ResourceID": "S3DynamicBucket62b0e7a664dc29c1c4fbe231fbcef30f8463aaa3",
            "ResourceRef": "spartaapplication-mweagl-s3dynamicbucket62b0e7a66-194zzvtfk757a",
            "ResourceType": "AWS::S3::Bucket",
            "Properties": {
                "DomainName": "spartaapplication-mweagl-s3dynamicbucket62b0e7a66-194zzvtfk757a.s3.amazonaws.com",
                "WebsiteURL": "http://spartaapplication-mweagl-s3dynamicbucket62b0e7a66-194zzvtfk757a.s3-website-us-west-2.amazonaws.com"
            }
        }
    }
}
```


This JSON data is Base64 encoded and published into the Lambda function's _Environment_ using the `SPARTA_DISCOVERY_INFO` key. The `sparta.Discover()` function is responsible for
accessing the encoded discovery information:

```go
configuration, _ := sparta.Discover()
bucketName := ""
for _, eachResource := range configuration.Resources {
  if eachResource.ResourceType == "AWS::S3::Bucket" {
    bucketName = eachResource.ResourceRef
  }
}
```


The `Properties` object includes resource-specific [Fn::GetAtt](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/intrinsic-function-reference-getatt.html) outputs (see each resource type's [documentation](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-template-resource-type-ref.html) for the complete set)

# Wrapping Up

Combined with [dynamic infrastructure](/reference/dynamic_infrastructure), `sparta.Discover()` enables a Sparta service to define its entire AWS infrastructure requirements.  Coupling application logic with infrastructure requirements moves a service towards being completely self-contained and in the direction of [immutable infrastructure](https://fugue.co/oreilly/).

# Notes
  - `sparta.Discovery()` **only** succeeds within a Sparta-compliant lambda function call block.
    - Call-site restrictions are validated in the [discovery_tests.go](https://github.com/mweagle/Sparta/blob/master/discovery_test.go) tests.
