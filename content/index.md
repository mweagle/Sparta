+++
author = "Matt W."
comments = true
date = "2016-01-19T14:58:42Z"
draft = false
share = false
title = "homepage"
+++

<div class="jumbotron">
  <h1>Sparta <img src="images/spartanshieldsmall.png" alt="Sparta shield" height="80" width="80"></h1>
  Build & deploy <b>Go</b> functions in AWS Lambda
  <hr />
  <blockquote>
    <p>"No Server Is Easier To Manage Than No Server."</p>
    <footer>Werner Vogels <cite title="Source Title">AWS re:Invent 2015</cite></footer>
  </blockquote>  
  <iframe width="50%" height="200" src="https://www.youtube.com/embed/y-0Wf2Zyi5Q?start=1742" frameborder="0" allowfullscreen></iframe>
</div>

Sparta defines a framework that deploys a set of *Go* HTTP request/response handlers to [AWS Lambda](https://aws.amazon.com/lambda/).

What differentiates Sparta from similar approaches is that it enables you to create and manage **the other AWS resources** associated with your application.   It also exposes the ability to generate, as part of your deployment, individual [IAM Policies](http://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies.html) under which your *Go* functions will execute.  The ability to limit lambda execution privileges helps support [POLP](http://searchsecurity.techtarget.com/definition/principle-of-least-privilege-POLP) and [#SecOps](https://twitter.com/hashtag/secops).

Sparta allows your application to create or reference, in *Go*, additional AWS resource relations including:   

  - [AWS Lambda Event Sources](http://docs.aws.amazon.com/lambda/latest/dg/intro-core-components.html)
    - DynamoDB
    - S3
    - Kinesis
    - SNS
    - SES
  - Other AWS resources
    - S3 buckets with dynamic names
    - SNS resources
    - Any other [CloudFormation Resource Type](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-template-resource-type-ref.html)
  - [API Gateway](http://docs.aws.amazon.com/apigateway/latest/developerguide/welcome.html) resources
  - [S3 Static Websites](http://docs.aws.amazon.com/AmazonS3/latest/dev/WebsiteHosting.html)
    - Sparta can provision an S3 bucket with your static resources, including [CORS](http://docs.aws.amazon.com/AmazonS3/latest/dev/cors.html) support

![Sparta Overview](images/sparta_overview.png)

Sparta leverages [CloudFormation](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/Welcome.html) to deploy and update your application.  For resources that CloudFormation does not yet support, it uses [Lambda-backed Custom Resources](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/template-custom-resources-lambda.html) so that all application updates support both update and rollback semantics.  CloudFormation resources use stable identifiers whenever possible to preserve [service availability](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks.html) during updates.

## Getting Started

To get started using Sparta, begin with the [Documentation](./docs).

## Problems?

Please file an [issue](https://github.com/mweagle/Sparta/issues/new) in GitHub.  Thanks.

### Other resources

  * [AWS Serverless Multi-Tier Architectures Whitepaper](https://d0.awsstatic.com/whitepapers/AWS_Serverless_Multi-Tier_Architectures.pdf)
  * [Lambda limits](http://docs.aws.amazon.com/lambda/latest/dg/limits.html)
  * [The Twelve Days of Lambda](https://aws.amazon.com/blogs/compute/the-twelve-days-of-lambda/)
