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

The following flowchart illustrates Sparta's flow during a provisioning operation:

{{< mermaid >}}
    graph TD
      iam[Verify Static IAM Roles]
      compile[Cross Compile App for AWS Linux AMI]
      package[ZIP archive]
      upload[Upload Archive to S3]
      packageAssets[Conditionally ZIP S3 Site Assets]
      uploadAssets[Upload S3 Assets]
      generate[Marshal to CloudFormation]
      decorate[Call User Template Decorators - Dynamic AWS Resources]
      uploadTemplate[Upload Template to S3]
      converge[Create/Update Stack]
      wait[Wait for Complete/Failure Result]

      iam-->compile
      compile-->package
      compile-->packageAssets
      package-->upload
      packageAssets-->uploadAssets
      uploadAssets-->generate
      upload-->generate
      generate-->decorate
      decorate-->uploadTemplate
      uploadTemplate-->converge
      converge-->wait
{{< /mermaid >}}

During provisioning, Sparta uses [AWS Lambda-backed Custom Resources](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/template-custom-resources-lambda.html) to support operations for which CloudFormation doesn't yet support (eg, [API Gateway](https://aws.amazon.com/api-gateway/) creation).

At runtime, Sparta uses [NodeJS](http://docs.aws.amazon.com/lambda/latest/dg/programming-model.html) shims to proxy the request to your **Go** handler.

Next up: writing a simple [Sparta Application](/docs/intro_example).
