+++
author = "Matt Weagle"
date = "2015-11-29T06:50:17"
title = "Overview"
type = "doc"
+++

This is a brief overview of the fundamental concepts behind Sparta.  Additional information regarding specific features is available from the menu.

At a high level, Sparta transforms a single **Go** binary's registered lambda functions into a set of AWS Lambda functions.  A _registered lambda function_ is simply an HTTP-style request/response function with a specific signature:

{{< highlight go >}}

func mySpartaLambdaFunction(event *json.RawMessage,
                      context *sparta.LambdaContext,
                      w http.ResponseWriter,
                      logger *logrus.Logger) {
  //
}

{{< /highlight >}}

These functions are grouped into a **ServiceName**, which is the logical, unique application identifier.  For example, `"MyEmailHandlerService-Dev"`. Only a single **ServiceInstance** (aka, deployment) may exist within an `(awsAccountId, awsRegion)` pair at any time.  

The **ServiceName** name has a 1:1 relationship to a [CloudFormation Stack](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/Welcome.html) name.  Sparta only uses CloudFormation to deploy and update service state.  

When creating or updating a service, Sparta follows this workflow:


{{< flowchart summary>}}
st=>start: Start
opIAM=>operation: Verify static IAM Roles
opCompile=>operation: Cross-compile for AWS Linux AMI
opZipApp=>operation: Zip application
opUploadApp=>operation: Upload archive to S3
condS3Site=>condition: S3 Site Included?
opZipSite=>operation: Zip S3 static site assets
opUploadSite=>operation: Upload S3 static assets

opMarshal=>operation: Marshal to CloudFormation
opDecorate=>operation:  Call TemplateDecorators:>https://godoc.org/github.com/mweagle/Sparta#TemplateDecorator[blank]
opUploadTemplate=>operation: Upload template to S3:>http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/cloudformation-limits.html[blank]
opProvision=>operation: Create/Update stack state
opWait=>operation: Wait for Complete/Failure result
e=>end: End

st->opIAM->opCompile->opZipApp->condS3Site
condS3Site(yes)->opZipSite->opUploadSite->opMarshal
condS3Site(no)->opMarshal
opMarshal->opDecorate->opUploadTemplate->opProvision->opWait->e
{{< /flowchart >}}

During provisioning, Sparta uses [AWS Lambda-backed Custom Resources](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/template-custom-resources-lambda.html) to support operations for which CloudFormation doesn't yet support (eg, [API Gateway](https://aws.amazon.com/api-gateway/) creation).

At runtime, Sparta uses [NodeJS](http://docs.aws.amazon.com/lambda/latest/dg/programming-model.html) shims to proxy the request to your **Go** handler.

Next up: writing a simple [Sparta Application](/docs/intro_example).
