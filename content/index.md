---
date: 2016-03-08T21:07:13+01:00
title:
type: index
weight: 0
---

<br />
<img src="/images/SpartaLogoNoDomain.png" alt="Sparta shield" height="192">
<br />

A <b>Go</b> framework for <a href="https://aws.amazon.com/lambda">AWS Lambda</a> microservices
<br />


<table style="width:90%">
  <!-- Row 1 -->
  <tr>
    <td style="width:50%">
      <h2>Unified Language</h2>
      <p>Use a single <b>Go</b> codebase to define your microservice's:
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
      <p>Define <a href="http://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies.html">IAM Roles</a> with limited privileges to minimize your service's attack surface.  Both string literal and ARN expressions are supported in order to reference dynamically created resources.  Sparta treats <a href="http://searchsecurity.techtarget.com/definition/principle-of-least-privilege-POLP">POLP</a> and <a href="https://twitter.com/hashtag/secops">#SecOps</a> as first-class goals.
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

Sparta exclusively relies on [CloudFormation](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/Welcome.html) to deploy and update your application.  For resources that CloudFormation does not yet support, it uses [Lambda-backed Custom Resources](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/template-custom-resources-lambda.html) so that all service updates support both update and rollback semantics.  Sparta's automatically generated CloudFormation resources use content-based logical IDs whenever possible to preserve [service availability](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks.html) during updates.

# Hello Lambda World

{{< highlight go >}}
// File: application.go
package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
	sparta "github.com/mweagle/Sparta"
)

////////////////////////////////////////////////////////////////////////////////
// Hello world event handler
//
func helloWorld(event *json.RawMessage,
              	context *sparta.LambdaContext,
              	w http.ResponseWriter,
              	logger *logrus.Logger) {
	logger.Info("Hello World: ", string(*event))
	fmt.Fprint(w, string(*event))
}

////////////////////////////////////////////////////////////////////////////////
// Main
func main() {
  var lambdaFunctions []*sparta.LambdaAWSInfo
  lambdaFn := sparta.NewLambda(sparta.IAMRoleDefinition{}, helloWorld, nil)
  lambdaFunctions = append(lambdaFunctions, lambdaFn)

  // Deploy it
  sparta.Main("SpartaHelloWorld",
		"Simple Sparta application that creates a single AWS Lambda function",
		lambdaFunctions,
                nil,
                nil)
}
{{< /highlight >}}

# Getting Started

To get started using Sparta, begin with the [Documentation](/docs/overview).

# Administration
  - Problems?  Please open an [issue](https://github.com/mweagle/Sparta/issues/new) in GitHub.
  - See [Trello](https://trello.com/b/WslDce70/sparta) for the Sparta backlog.

# Questions?

Get in touch via:

  - <i class="fa fa-twitter">&nbsp; @mweagle</i>
  - <i class="fa fa-slack">&nbsp; Gophers: <a href="https://gophers.slack.com/team/mweagle">@mweagle</a></i>
    - [Signup page](https://invite.slack.golangbridge.org/)
  - <i class="fa fa-slack">&nbsp; Serverless: <a href="https://serverless-forum.slack.com/team/mweagle">@mweagle</a></i>
    - [Signup page](https://wt-serverless-seattle.run.webtask.io/serverless-forum-signup?webtask_no_cache=1)

## Other resources

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
