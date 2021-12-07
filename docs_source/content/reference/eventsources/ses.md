---
date: 2016-03-09T19:56:50+01:00
title: SES
weight: 10
---

In this section we'll walkthrough how to trigger your lambda function in response to inbound email. This overview is based on the [SpartaApplication](https://github.com/mweagle/SpartaApplication/blob/master/application.go) sample code if you'd rather jump to the end result.

# Goal

Assume that we have already [verified our email domain](http://docs.aws.amazon.com/ses/latest/DeveloperGuide/verify-domains.html) with AWS. This allows our domain's email to be handled by SES.

We've been asked to write a lambda function that logs inbound messages, including the metadata associated with the message body itself.

There is also an additional requirement to support [immutable infrastructure](http://radar.oreilly.com/2015/06/an-introduction-to-immutable-infrastructure.html), so our service needs to manage the S3 bucket to which message bodies should be stored. Our service cannot rely on a pre-existing S3 bucket. The infrastructure (and associated security policies) together with the application logic is coupled.

## Getting Started

We'll start with an empty lambda function and build up the needed functionality.

```go
import (
  spartaSES "github.com/mweagle/Sparta/v3/aws/ses"
)
func echoSESEvent(ctx context.Context, sesEvent spartaSES.Event) (*spartaSES.Event, error) {
  logger, _ := ctx.Value(sparta.ContextKeyRequestLogger).(*zerolog.Logger)
  configuration, configErr := sparta.Discover()

  logger.Info().
    Err(configErr).
    Interface("Configuration", configuration).
    Msg("Discovery results")
}
```

## Unmarshalling the SES Event

At this point we would normally continue processing the SES event, using Sparta types until the
official [events](https://godoc.org/github.com/aws/aws-lambda-go/events) are available.

However, before moving on to the event processing, we need to take a detour into [dynamic infrastructure](/reference/dynamic_infrastructure/)
because of the immutable infrastructure requirement.

This requirement implies that our service must be self-contained: we can't assume that
the S3 bucket already exists. How can our locally compiled code access AWS-created resources?

## Dynamic Resources

The immutable infrastructure requirement makes this lambda function a bit more complex. Our service needs to:

- Provision a new S3 bucket for email message body storage
  - SES will not provide the message body in the event data. It will only store
    the email body in an S3 bucket, from which your lambda function can later consume it.
- Wait for the S3 bucket to be provisioned
  - As we need a new S3 bucket, we're relying on AWS to generate a [unique name](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-s3-bucket.html#cfn-s3-bucket-name). But this means that our lambda function doesn't know the S3 bucket name during provisioning.
  - In fact, we shouldn't even create an AWS Lambda function if the S3 bucket can't be created.
- Include an IAMPrivilege so that our **go** function can access the dynamically created bucket
- Discover the S3 Bucket at lambda execution time

### Provision Message Body Storage Resource

Let's first take a look at how the SES lambda handler provisions a new S3 bucket via the [MessageBodyStorage](https://godoc.org/github.com/mweagle/Sparta#MessageBodyStorage) type:

```go

func appendSESLambda(api *sparta.API,
  lambdaFunctions []*sparta.LambdaAWSInfo)
  []*sparta.LambdaAWSInfo {

  // Our lambda function will need to be able to read from the bucket, which
  // will be handled by the S3MessageBodyBucketDecorator below
  lambdaFn, _ := sparta.NewAWSLambda(sparta.LambdaName(echoSESEvent),
    echoSESEvent,
    sparta.IAMRoleDefinition{})
  // Setup options s.t. the lambda function has time to consume the message body
  lambdaFn.Options = &sparta.LambdaFunctionOptions{
    Description: "",
    MemorySize:  128,
    Timeout:     10,
  }

  // Add a Permission s.t. the Lambda function automatically manages SES registration
  sesPermission := sparta.SESPermission{
    BasePermission: sparta.BasePermission{
      // SES only supports wildcard ARNs
      SourceArn: "*",
    },
    InvocationType: "Event",
  }
  // Store the message body
  bodyStorage, _ := sesPermission.NewMessageBodyStorageResource("Special")
  sesPermission.MessageBodyStorage = bodyStorage

```

The `MessageBodyStorage` type (and the related [MessageBodyStorageOptions](https://godoc.org/github.com/mweagle/Sparta#MessageBodyStorageOptions) type)
cause our SESPermission handler to add an [S3 ReceiptRule](http://docs.aws.amazon.com/ses/latest/DeveloperGuide/receiving-email-action-s3.html)
at the head of the rules list. This rule instructs SES to store the message body in the supplied bucket before invoking our lambda function.

The single parameter `"Special"` is an application-unique literal value that is used to create a stable CloudFormation
resource identifier so that new buckets are not created in response to stack update requests.

Our SES handler then adds two [ReceiptRules](http://docs.aws.amazon.com/ses/latest/APIReference/API_ReceiptRule.html):

```go

sesPermission.ReceiptRules = make([]sparta.ReceiptRule, 0)
sesPermission.ReceiptRules = append(sesPermission.ReceiptRules,
  sparta.ReceiptRule{
    Name:       "Special",
    Recipients: []string{"sombody_special@gosparta.io"},
    TLSPolicy:  "Optional",
  })
sesPermission.ReceiptRules = append(sesPermission.ReceiptRules,
  sparta.ReceiptRule{
    Name:       "Default",
    Recipients: []string{},
    TLSPolicy:  "Optional",
  })
```

### Dynamic IAMPrivilege Arn

Our lambda function is required to access the message body in the dynamically created `MessageBodyStorage` resource, but
the S3 resource Arn is only defined _after_ the service is provisioned. The solution to this is to reference the
dynamically generated `BucketArnAllKeys()` value in the `sparta.IAMRolePrivilege` initializer:

```go

// Then add the privilege to the Lambda function s.t. we can actually get at the data
lambdaFn.RoleDefinition.Privileges = append(lambdaFn.RoleDefinition.Privileges,
  sparta.IAMRolePrivilege{
    Actions:  []string{"s3:GetObject", "s3:HeadObject"},
    Resource: sesPermission.MessageBodyStorage.BucketArnAllKeys(),
})
```

The last step is to register the `SESPermission` with the lambda info:

```go
// Finally add the SES permission to the lambda function
lambdaFn.Permissions = append(lambdaFn.Permissions, sesPermission)
```

At this point we've implicitly created an S3 bucket via the `MessageBodyStorage` value. Our lambda function now needs to dynamically determine the AWS-assigned bucket name.

### Dynamic Message Body Storage Discovery

Our `echoSESEvent` function needs to determine, at execution time, the `MessageBodyStorage` S3 bucket name. This is done via `sparta.Discover()`:

```go
configuration, configErr := sparta.Discover()

logger.Info().
  Err(configErr).
  Interface("Configuration", configuration).
  Msg("Discovery results")

// The message bucket is an explicit `DependsOn` relationship, so it'll be in the
// resources map.  We'll find it by looking for the dependent resource with the "AWS::S3::Bucket" type
bucketName := ""
for _, eachResourceInfo := range configuration.Resources {
  if eachResourceInfo.ResourceType == "AWS::S3::Bucket" {
    bucketName = eachResourceInfo.Properties["Ref"]
  }
}
if "" == bucketName {
  return nil, errors.Errorf("Failed to discover SES bucket from sparta.Discovery: %#v", configuration)
}
```

The `sparta.Discover()` function returns a [DiscoveryInfo](https://godoc.org/github.com/mweagle/Sparta#DiscoveryInfo) structure. This
data is published into the Lambda's environment variables to enable it to discover other
resources published in the same Stack.

The structure includes the stack's [Pseudo Parameters](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/pseudo-parameter-reference.html)
as well information about any _immediate_ resource dependencies.
Eg, those that were explicitly marked as `DependsOn`. See the [discovery documentation](/reference/discovery/) for more details.

As we only have a single dependency, our discovery filter is:

```go
// The message bucket is an explicit `DependsOn` relationship, so it'll be in the
// resources map.  We'll find it by looking for the dependent resource with the "AWS::S3::Bucket" type
bucketName := ""
for _, eachResourceInfo := range configuration.Resources {
  if eachResourceInfo.ResourceType == "AWS::S3::Bucket" {
    bucketName = eachResourceInfo.Properties["Ref"]
  }
}
if "" == bucketName {
  return nil, errors.Errorf("Failed to discover SES bucket from sparta.Discovery: %#v", configuration)
}
```

# #Sparta Integration

The rest of `echoSESEvent` satisfies the other requirements, with a bit of help from the SES [event types](https://godoc.org/github.com/mweagle/Sparta/aws/ses):

```go
// Get the metdata about the item...
svc := s3.New(session.New())
for _, eachRecord := range sesEvent.Records {

  logger.Info().
    Str("Source", eachRecord.SES.Mail.Source).
    Str("MessageID", eachRecord.SES.Mail.MessageID).
    Str("BucketName", bucketName).
    Msg("SES Event")

  if "" != bucketName {
    params := &s3.HeadObjectInput{
      Bucket: aws.String(bucketName),
      Key:    aws.String(eachRecord.SES.Mail.MessageID),
    }
    resp, err := svc.HeadObject(params)

    logger.Info().
      Interface("Metadata", resp).
      Err(err).
      Msg("SES MessageBody")
  }
}
return &sesEvent, nil
```

# Wrapping Up

With the `lambdaFn` fully defined, we can provide it to `sparta.Main()` and deploy our service. The workflow below is shared by all SES-triggered lambda function:

- Define the lambda function (`echoSESEvent`).
- If needed, create the required [IAMRoleDefinition](https://godoc.org/github.com/mweagle/Sparta*IAMRoleDefinition) with appropriate privileges if the lambda function accesses other AWS services.
- Provide the lambda function & IAMRoleDefinition to `sparta.NewAWSLambda()`
- Add the necessary [Permissions](https://godoc.org/github.com/mweagle/Sparta#LambdaAWSInfo) to the `LambdaAWSInfo` struct so that the lambda function is triggered.

Additionally, if the SES handler needs to access the raw email message body:

- Create a new `sesPermission.NewMessageBodyStorageResource("Special")` value to store the message body
- Assign the value to the `sesPermission.MessageBodyStorage` field
- If your lambda function needs to consume the message body, add an entry to `sesPermission.[]IAMPrivilege` that includes the `sesPermission.MessageBodyStorage.BucketArnAllKeys()` Arn
- In your **go** lambda function definition, discover the S3 bucketname via `sparta.Discover()`

# Notes

- The SES message (including headers) is stored in the [raw format](http://stackoverflow.com/questions/33549327/what-is-the-format-of-the-aws-ses-body-stored-in-s3)
- More on Immutable Infrastructure:
  - [Subbu - Automate Everything](https://www.subbu.org/blog/2014/10/automate-everything-but-dont-ignore-drift)
  - [Chad Fowler - Immutable Deployments](http://chadfowler.com/2013/06/23/immutable-deployments.html)
  - [The Cloudcast - What is Immutable Infrastructure](http://www.thecloudcast.net/2015/09/the-cloudcast-213-what-is-immutable.html)
  - [The New Stack](http://thenewstack.io/a-brief-look-at-immutable-infrastructure-and-why-it-is-such-a-quest/)
