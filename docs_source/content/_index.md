---
date: 2016-03-08T21:07:13+01:00
title:
type: index
weight: 0
---

<br />

![Sparta](/images/SpartaLogoNoDomain.png "Sparta")

# Self-deploying *go* microservices for [AWS Lambda](https://aws.amazon.com/lambda)


Sparta is a framework that transforms a standard *go* application into a self-deploying AWS Lambda powered service. All configuration and infrastructure requirements are expressed as go types - no JSON or YAML needed!

# Sample Application

## 1. Definition

```go
// File: application.go
package main

import (
  sparta "github.com/mweagle/Sparta"
)

////////////////////////////////////////////////////////////////////////////////
// Hello world event handler
//
func helloWorld() (string, error) {
  return "Hello World 🌏", nil
}

////////////////////////////////////////////////////////////////////////////////
// Main
func main() {

  var lambdaFunctions []*sparta.LambdaAWSInfo
  lambdaFn := sparta.HandleAWSLambda("Hello world test",
    helloWorld,
    sparta.IAMRoleDefinition{})
  lambdaFunctions = append(lambdaFunctions, lambdaFn)

  // Delegate to Sparta
  sparta.Main("SpartaHelloWorld",
    "Simple Sparta application that creates a single AWS Lambda function",
    lambdaFunctions,
                nil,
                nil)
}
```

## 2. Deployment

```shell
$ go run main.go provision --s3Bucket $MY_S3_BUCKET
[0000]  INFO ════════════════════════════════════════════════
[0000]  INFO ╔═╗┌─┐┌─┐┬─┐┌┬┐┌─┐   Version : 1.1.1
[0000]  INFO ╚═╗├─┘├─┤├┬┘ │ ├─┤   SHA     : beb5700
[0000]  INFO ╚═╝┴  ┴ ┴┴└─ ┴ ┴ ┴   Go      : go1.10
[0000]  INFO ════════════════════════════════════════════════
[0000]  INFO Service: SpartaHelloWorld                          LinkFlags= Option=provision UTC=2018-05-25T04:36:50Z
[0000]  INFO ════════════════════════════════════════════════
[0000]  INFO Using `git` SHA for StampedBuildID                 Command=git rev-parse HEAD SHA=7ee3e1bc52f15c4a636e05061eaec7b748db22a9
[0000]  INFO Provisioning service                               BuildID=7ee3e1bc52f15c4a636e05061eaec7b748db22a9 CodePipelineTrigger= InPlaceUpdates=false NOOP=false Tags=
[0000]  INFO Verifying IAM Lambda execution roles
[0000]  INFO IAM roles verified                                 Count=1
[0000]  INFO Checking S3 versioning                             Bucket=MY_S3_BUCKET VersioningEnabled=true
[0000]  INFO Checking S3 region                                 Bucket=MY_S3_BUCKET Region=us-west-2
[0000]  INFO Running `go generate`
[0000]  INFO Compiling binary                                   Name=Sparta.lambda.amd64
[0001]  INFO Creating code ZIP archive for upload               TempName=./.sparta/SpartaHelloWorld-code.zip
[0001]  INFO Lambda code archive size                           Size=13 MB
[0001]  INFO Uploading local file to S3                         Bucket=MY_S3_BUCKET Key=SpartaHelloWorld/SpartaHelloWorld-code.zip Path=./.sparta/SpartaHelloWorld-code.zip Size=13 MB
[0011]  INFO Uploading local file to S3                         Bucket=MY_S3_BUCKET Key=SpartaHelloWorld/SpartaHelloWorld-cftemplate.json Path=./.sparta/SpartaHelloWorld-cftemplate.json Size=2.2 kB
[0011]  INFO Creating stack                                     StackID=arn:aws:cloudformation:us-west-2:123412341234:stack/SpartaHelloWorld/44b426d0-5fd5-11e8-90cd-503f20f2ad82
[0039]  INFO CloudFormation Metrics ▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬
[0039]  INFO     Operation duration                             Duration=24.49s Resource=SpartaHelloWorld Type=AWS::CloudFormation::Stack
[0039]  INFO     Operation duration                             Duration=15.02s Resource=IAMRolebc3b888fab1664f42799785a6d9a3bb76ea63798 Type=AWS::IAM::Role
[0039]  INFO     Operation duration                             Duration=2.90s Resource=HelloworldtestLambdaa4a9e95c84255c65c284ce95dd7d3c45776c1a68 Type=AWS::Lambda::Function
[0039]  INFO Stack provisioned                                  CreationTime=2018-05-25 04:37:02.099 +0000 UTC StackId=arn:aws:cloudformation:us-west-2:123412341234:stack/SpartaHelloWorld/44b426d0-5fd5-11e8-90cd-503f20f2ad82 StackName=SpartaHelloWorld
[0039]  INFO ════════════════════════════════════════════════
[0039]  INFO SpartaHelloWorld Summary
[0039]  INFO ════════════════════════════════════════════════
[0039]  INFO Verifying IAM roles                                Duration (s)=0
[0039]  INFO Verifying AWS preconditions                        Duration (s)=0
[0039]  INFO Creating code bundle                               Duration (s)=1
[0039]  INFO Uploading code                                     Duration (s)=10
[0039]  INFO Ensuring CloudFormation stack                      Duration (s)=28
[0039]  INFO Total elapsed time                                 Duration (s)=39
```

## 3. Invoke

![Console GUI](/images/invoke.jpg "Invoke")

<hr />

# Features

<table style="width:90%">
  <!-- Row 1 -->
  <tr>
    <td style="width:50%">
      <h2>Unified</h2>
      <p>Use a <b>go</b> monorepo to define and your microservice's:
      <ul>
        <li>Application logic</li>
        <li>AWS infrastructure</li>
        <li>Operational metrics</li>
        <li>Alert conditions</li>
        <li>Security policies</li>
      </ul>
    </td>
    <td style="width:50%">
      <h2>Complete AWS Ecosystem</h2>
      <p>Sparta enables your lambda-based service to seamlessly integrate with the entire set of AWS lambda <a href="http://docs.aws.amazon.com/lambda/latest/dg/intro-core-components.html">event sources</a> such as:
        <ul>
          <li>DynamoDB</li>
          <li>S3</li>
          <li>Kinesis</li>
          <li>SNS</li>
          <li>SES</li>
          <li>CloudWatch Events</li>
          <li>CloudWatch Logs</li>
        </ul>
        Additionally, your service may provision any other <a href="http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-template-resource-type-ref.html">CloudFormation</a> supported resource and even your own <a href="http://gosparta.io/docs/custom_resources">CustomResources</a>.
        </p>
    </td>
  </tr>
  <!-- Row 2 -->
  <tr>
    <td style="width:50%">
      <h2>Security</h2>
      <p>Define <a href="http://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies.html">IAM Roles</a> with limited privileges to minimize your service's attack surface.  Both string literal and ARN expressions are supported in order to reference dynamically created resources.  Sparta treats <a href="http://searchsecurity.techtarget.com/definition/principle-of-least-privilege-POLP">POLA</a> and <a href="https://twitter.com/hashtag/secops">#SecOps</a> as first-class goals.
      </p>
    </td>
    <td style="width:50%">
      <h2>Discovery</h2>
      <p>A service may provision dynamic AWS infrastructure, and <a href="http://gosparta.io/docs/eventsources">discover</a>, at lambda execution time, the dependent resources' AWS-assigned outputs (<code>Ref</code> &amp; <code>Fn::Att</code>).  Eliminate hardcoded <i>Magic ARNs</i> from your codebase and move towards <a href="http://chadfowler.com/2013/06/23/immutable-deployments.html">immutable infrastructure</a></p>
    </td>
  </tr>
  <!-- Row 3 -->
  <tr>
    <td style="width:50%">
      <h2>API Gateway</h2>
      <p>Make your service HTTPS accessible by binding it to an <a href="http://docs.aws.amazon.com/apigateway/latest/developerguide/welcome.html">API Gateway</a> REST API during provisioning.  As part of API Gateway creation, Sparta includes <a href="http://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-mapping-template-reference.html">API Gateway Mapping Templates</a> with all request data, including user-defined whitelisted parameters, so that you can focus on your core application logic.</p>
    </td>
    <td style="width:50%">
      <h2>Static Sites</h2>
      <p>Include a <a href="http://docs.aws.amazon.com/AmazonS3/latest/dev/cors.html">CORS-enabled</a> <a href="http://docs.aws.amazon.com/AmazonS3/latest/dev/WebsiteHosting.html">S3-backed site</a> with your service.  S3-backed sites include API Gateway discovery information for turnkey deployment.</p>
    </td>
  </tr>
</table>


<hr />
<a href="https://cloudcraft.co/view/8571b3bc-76ef-48c1-8401-0b6ae1d36b4e?key=d44zi4j1pxj00000" rel="Sparta Arch">![Sparta Overview](/images/sparta_overview.png)</a>

Sparta relies on [CloudFormation](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/Welcome.html) to deploy and update your application.  For resources that CloudFormation does not yet support, it uses [Lambda-backed Custom Resources](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/template-custom-resources-lambda.html) so that all service updates support both update and rollback semantics.  Sparta's automatically generated CloudFormation resources use content-based logical IDs whenever possible to preserve [service availability](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks.html) and minimize resource churn during updates.

# Getting Started

To get started using Sparta, begin with the [Overview](/overview).

# Administration
  - Problems?  Please open an [issue](https://github.com/mweagle/Sparta/issues/new) in GitHub.

<br />
<img src="/images/GopherInclusion.png" alt="Eveyone Welcome" height="256">
<center>
<h6>Courtesy of <a href="https://github.com/ashleymcnamara/gophers">gophers</a>
</h6>
</center>
<br />


# Questions?

Get in touch via:

  - <i class="fas fas-twitter">&nbsp; @mweagle</i>
  - <i class="fas fas-slack">&nbsp; Gophers: <a href="https://gophers.slack.com/team/mweagle">@mweagle</a></i>
    - [Signup page](https://invite.slack.golangbridge.org/)
  - <i class="fas fas-slack">&nbsp; Serverless: <a href="https://serverless-forum.slack.com/team/mweagle">@mweagle</a></i>
    - [Signup page](https://wt-serverless-seattle.run.webtask.io/serverless-forum-signup?webtask_no_cache=1)


## Related Projects
  - [go-cloudcondensor](https://github.com/mweagle/go-cloudcondenser)
    - Define AWS CloudFormation templates in `go`
  - [go-cloudformation](https://github.com/mweagle/go-cloudformation)
    - `go` types for CloudFormation resources
  - [ssm-cache](https://github.com/mweagle/ssm-cache)
    - Lightweight cache for [Systems Manager Paramter Store](https://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-paramstore.html) values

## Other resources
  * [Build an S3 website with API Gateway and AWS Lambda for Go using Sparta](https://medium.com/@mweagle/go-aws-lambda-building-an-html-website-with-api-gateway-and-lambda-for-go-using-sparta-5e6fe79f63ef)
  * [AWS blog post announcing Go support](https://aws.amazon.com/blogs/compute/announcing-go-support-for-aws-lambda/)
  * [Sparta - A Go framework for AWS Lambda](https://medium.com/@mweagle/a-go-framework-for-aws-lambda-ab14f0c42cb#.6gtlwe5vg)
  * Other libraries & frameworks:
    * [Serverless](https://github.com/serverless/serverless)
    * [PAWS](https://github.com/braahyan/PAWS)
    * [Apex](http://apex.run)
    * [lambda_proc](https://github.com/jasonmoo/lambda_proc)
    * [go-lambda](https://github.com/xlab/go-lambda)
    * [go-lambda (GRPC)](https://github.com/pilwon/go-lambda)
  * Supported AWS Lambda [programming models](http://docs.aws.amazon.com/lambda/latest/dg/programming-model-v2.html)
  * [Serverless Code Blog](https://serverlesscode.com)
  * [AWS Serverless Multi-Tier Architectures Whitepaper](https://d0.awsstatic.com/whitepapers/AWS_Serverless_Multi-Tier_Architectures.pdf)
  * [Lambda limits](http://docs.aws.amazon.com/lambda/latest/dg/limits.html)
  * [The Twelve Days of Lambda](https://aws.amazon.com/blogs/compute/the-twelve-days-of-lambda/)
  * [CloudCraft](http://cloudcraft.co) is a great tool for AWS architecture diagrams
