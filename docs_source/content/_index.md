---
date: 2016-03-08T21:07:13+01:00
title:
type: index
weight: 0
---

<img src="/images/SpartaLogoNoDomain.png" width="33%" height="33%">

## Serverless _go_ microservices for AWS

<table style="width:90%">
  <!-- Row 1 -->
  <tr>
    <td style="width:33%" valign="top">
      <h4>
      Sparta is a framework that transforms a <i>go</i> application into a self-deploying AWS Lambda powered service.
      </h4>
      All configuration and infrastructure requirements are expressed as <i>go</i> types for GitOps, repeatable, typesafe deployments.
      </h4>
    </td>
    <td style="width:66%">
      <img src="/images/sparta_ecosystem.png" alt="Sparta Ecosystem" />
    </td>
  </tr>
</table>

<hr />

## Features

<table style="width:90%">
  <!-- Row 1 -->
  <tr>
    <td style="width:50%">
      <h2>Unified</h2>
      <p>Use a <i>go</i> monorepo to define and your microservice's:
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
          <li>CloudMap</li>
          <li>CloudWatch Events</li>
          <li>CloudWatch Logs</li>
          <li>Step Functions</li>
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
      <h2>API Gateways</h2>
      <p>Make your service HTTPS accessible by binding it to an <a href="http://docs.aws.amazon.com/apigateway/latest/developerguide/welcome.html">API Gateway</a> REST API during provisioning.  Alternatively, expose a WebSocket [APIV2Gateway](https://aws.amazon.com/blogs/compute/announcing-websocket-apis-in-amazon-api-gateway/) API for an even more interactive experience.</p>
    </td>
    <td style="width:50%">
      <h2>Static Sites</h2>
      <p>Include a <a href="http://docs.aws.amazon.com/AmazonS3/latest/dev/cors.html">CORS-enabled</a> <a href="http://docs.aws.amazon.com/AmazonS3/latest/dev/WebsiteHosting.html">S3-backed site</a> with your service.  S3-backed sites include API Gateway discovery information for turnkey deployment.</p>
    </td>
  </tr>
</table>
<hr />
<a href="https://cloudcraft.co/view/8571b3bc-76ef-48c1-8401-0b6ae1d36b4e?key=d44zi4j1pxj00000" rel="Sparta Arch">
<img src="/images/sparta_overview.png" width="50%" height="50%">
</a>

Sparta relies on [CloudFormation](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/Welcome.html) to deploy and update your application. For resources that CloudFormation does not yet support, it uses [Lambda-backed Custom Resources](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/template-custom-resources-lambda.html) so that all service updates support both update and rollback semantics. Sparta's automatically generated CloudFormation resources use content-based logical IDs whenever possible to preserve [service availability](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks.html) and minimize resource churn during updates.

## Getting Started

To get started using Sparta, begin with the [Overview](/example_service/step1/).

## Administration

- Problems? Please open an [issue](https://github.com/mweagle/Sparta/issues/new) in GitHub.

<br />
<img src="/images/GopherInclusion.png" alt="Eveyone Welcome" height="256">
<center>
<h6>Courtesy of <a href="https://github.com/ashleymcnamara/gophers">gophers</a>
</h6>
</center>
<br />

## Questions?

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
  - Lightweight cache for [Systems Manager Parameter Store](https://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-paramstore.html) values

## Other resources

- [AWS SAM](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/what-is-sam.html)
- [Build an S3 website with API Gateway and AWS Lambda for Go using Sparta](https://medium.com/@mweagle/go-aws-lambda-building-an-html-website-with-api-gateway-and-lambda-for-go-using-sparta-5e6fe79f63ef)
- [AWS blog post announcing Go support](https://aws.amazon.com/blogs/compute/announcing-go-support-for-aws-lambda/)
- [Sparta - A Go framework for AWS Lambda](https://medium.com/@mweagle/a-go-framework-for-aws-lambda-ab14f0c42cb#.6gtlwe5vg)
- Other libraries & frameworks:
  - [Serverless](https://github.com/serverless/serverless)
  - [PAWS](https://github.com/braahyan/PAWS)
  - [Apex](http://apex.run)
  - [lambda_proc](https://github.com/jasonmoo/lambda_proc)
  - [go-lambda](https://github.com/xlab/go-lambda)
  - [go-lambda (GRPC)](https://github.com/pilwon/go-lambda)
- Supported AWS Lambda [programming models](http://docs.aws.amazon.com/lambda/latest/dg/programming-model-v2.html)
- [Serverless Code Blog](https://serverlesscode.com)
- [AWS Serverless Multi-Tier Architectures Whitepaper](https://d0.awsstatic.com/whitepapers/AWS_Serverless_Multi-Tier_Architectures.pdf)
- [Lambda limits](http://docs.aws.amazon.com/lambda/latest/dg/limits.html)
- [The Twelve Days of Lambda](https://aws.amazon.com/blogs/compute/the-twelve-days-of-lambda/)
- [CloudCraft](http://cloudcraft.co) is a great tool for AWS architecture diagrams
