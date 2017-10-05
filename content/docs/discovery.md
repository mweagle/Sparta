---
date: 2016-03-09T19:56:50+01:00
title: Discovery Service
weight: 10
menu:
  main:
    parent: Documentation
    identifier: discovery
    weight: 100
---

# Introduction

The ability to provision [dynamic infrastructure](/docs/dynamic_infrastructure) (see also the [SES Event Source Example](/docs/eventsources/ses/#dynamic-resources:d680e8a854a7cbad6d490c445cba2eba)) as part of a Sparta application creates a need to discover those resources at lambda execution time.

Sparta exposes this functionality via [sparta.Discover](https://godoc.org/github.com/mweagle/Sparta#Discover).  This function returns information about the current stack (eg, name, region, ID) as well as metadata about the immediate dependencies of the calling **go** lambda function.

The following sections walk through provisioning a S3 bucket, declaring an explicit dependency on that resource, and then discovering the resource at lambda execution time.  It is extracted from `appendDynamicS3BucketLambda` in the  [SpartaApplication](https://github.com/mweagle/SpartaApplication/blob/master/application.go) source.

If you haven't already done so, please review the [Dynamic Infrastructure](/docs/dynamic_infrastructure) section for background on dynamic infrastructure provisioning.


# Discovery

For reference, we provision an S3 bucket and declare an explicit dependency with the code below.  Because our `gocf.S3Bucket{}` struct uses a zero-length [BucketName](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-s3-bucket.html#cfn-s3-bucket-name) property, CloudFormation will dynamically assign one.

{{< highlight go >}}

s3BucketResourceName := sparta.CloudFormationResourceName("S3DynamicBucket")

lambdaFn := sparta.HandleAWSLambda(sparta.LambdaName(echoS3DynamicBucketEvent),
  http.HandlerFunc(echoS3DynamicBucketEvent),
  sparta.IAMRoleDefinition{})

lambdaFn.Permissions = append(lambdaFn.Permissions,
  sparta.S3Permission{
    BasePermission: sparta.BasePermission{
      SourceArn: gocf.Ref(s3BucketResourceName),
    },
    Events: []string{"s3:ObjectCreated:*",
                      "s3:ObjectRemoved:*"},
  })

lambdaFn.DependsOn = append(lambdaFn.DependsOn, s3BucketResourceName)

// Add permission s.t. the lambda function could read from the S3 bucket
lambdaFn.RoleDefinition.Privileges = append(lambdaFn.RoleDefinition.Privileges,
  sparta.IAMRolePrivilege{
    Actions:  []string{"s3:GetObject",
                       "s3:HeadObject"},
    Resource: spartaCF.S3AllKeysArnForBucket(gocf.Ref(s3BucketResourceName)),
  })

// Add the S3 resource to our template
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

The key to `sparta.Discovery` is the `DependsOn` slice value.

# Template Marshaling & Decoration

Sparta only uses CloudFormation to update service state.  Each CloudFormation resource may be annotated with a [Metadata](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-attribute-metadata.html) property.  This is unstructured, application-defined data.

During template marshaling, Sparta scans for `DependsOn` relationships and propagates information (immediate-children only) across CloudFormation resource definitions.  Most importantly, this information includes `Ref` and any other [outputs](https://github.com/mweagle/Sparta/blob/master/cloudformation_resources.go#L24) of referred resources.  This information then becomes available as a [DisocveryInfo](https://godoc.org/github.com/mweagle/Sparta#DiscoveryInfo) value returned by `sparta.Discovery()`.

# Sample DiscoveryInfo

In our example, a `DiscoveryInfo` from a sample stack might be:

{{< highlight json >}}
{
  "Region": "us-west-2",
  "StackID": "arn:aws:cloudformation:us-west-2:123412341234:stack/SpartaApplication/d87bb070-cce5-11e5-b6ca-503f20f2ad1e",
  "StackName": "SpartaApplication",
  "Resources": {
    "S3DynamicBucketa3c3daf5543fe6879721e6bb0bc129042459be6b": {
      "ResourceID": "S3DynamicBucketa3c3daf5543fe6879721e6bb0bc129042459be6b",
      "Properties": {
        "DomainName": "spartaapplication-s3dynamicbucketa3c3daf5543fe687-1avr6h3lrprju.s3.amazonaws.com",
        "Ref": "spartaapplication-s3dynamicbucketa3c3daf5543fe687-1avr6h3lrprju",
        "WebsiteURL": "http://spartaapplication-s3dynamicbucketa3c3daf5543fe687-1avr6h3lrprju.s3-website-us-west-2.amazonaws.com",
        "sparta:cloudformation:restype": "AWS::S3::Bucket"
      },
      "Tags": {}
    }
  }
}
{{< /highlight >}}


With this, our lambda function can discover the dynamically assigned bucket name using non-production ready code similar to:

{{< highlight go >}}
configuration, _ := sparta.Discover()
bucketName := ""
for _, eachResource := range configuration.Resources {
  if eachResource.Properties[sparta.TagResourceType] == "AWS::S3::Bucket" {
    bucketName = eachResource.Properties["Ref"]
  }
}
{{< /highlight >}}


The `Properties` map includes the CloudFormation `Ref` and other outputs (see each resource type's [documentation](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-template-resource-type-ref.html) for the complete set) as well as the actual resource type (keyed by `sparta.TagResourceType`).

Multiple dependencies can be disambiguated by annotating the CloudFormation resource's [Tags](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-resource-tags.html) slice.  Tag key-values are published as a key-value map as part of each [DiscoveryResource](https://godoc.org/github.com/mweagle/Sparta#DiscoveryResource).  For example:

{{< highlight go >}}
cfResource := template.AddResource(s3BucketResourceName, &gocf.S3Bucket{
  AccessControl: gocf.String("PublicRead"),
  Tags: []gocf.ResourceTag{
    gocf.ResourceTag{
      Key:   gocf.String("SomeUserKey"),
      Value: gocf.String("MyMagicValue"),
    },
  },
})
{{< /highlight >}}


# Special Considerations

You may have noticed that `sparta.Discover()` doesn't accept any parameters.  Sparta distinguishes different invocations by using the enclosing Sparta-compliant **go** function name.  This name is discovered by reflection, which creates a 1-to-1 mapping with [LambdaAWSInfo](https://godoc.org/github.com/mweagle/Sparta#LambdaAWSInfo) values and enables unique discovery.

# Wrapping Up

Combined with [dynamic infrastructure](/docs/dynamic_infrastructure), `sparta.Discover()` enables a Sparta service to define its entire AWS infrastructure requirements.  Coupling application logic with infrastructure requirements moves a service towards being completely self-contained and in the direction of [immutable infrastructure](https://fugue.co/oreilly/).

# Notes
  - `sparta.Discovery()` **only** succeeds within a Sparta-compliant lambda function call block.
    - Call-site restrictions are validated in the [discovery_tests.go](https://github.com/mweagle/Sparta/blob/master/discovery_test.go) tests.
