+++
author = "Matt W."
comments = true
date = "2016-01-19T14:58:42Z"
draft = false
share = false
title = "homepage"
+++

<div class="jumbotron">
<img src="images/spartalogoSmall.png" alt="Sparta shield" height="128">
<h2>Use <b>Go</b> to write and manage <a href="https://aws.amazon.com/lambda">AWS Lambda</a> services</h2>

  <hr />
  <blockquote>
    <p>"No Server Is Easier To Manage Than No Server."</p>
    <footer>Werner Vogels <cite title="Source Title">AWS re:Invent 2015</cite></footer>
  </blockquote>  
  <iframe width="50%" height="200" src="https://www.youtube.com/embed/y-0Wf2Zyi5Q?start=1742" frameborder="0" allowfullscreen></iframe>
</div>

Sparta defines a framework that deploys a set of **Go** HTTP request/response handlers to [AWS Lambda](https://aws.amazon.com/lambda/).

What differentiates Sparta from similar solutions (see below), is that is also helps create & discover **the other AWS resources** a service typically requires:

  -  [CloudFormation](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-template-resource-type-ref.html) resources that should be created during your service's provisioning.  
  - Discovery of those dependent resources' CloudFormation outputs (`Ref` && `Fn::Att` values) at Lambda execution time
    - This enables a service to close over its AWS infrastructure requirements.  Eliminate hardcoded _Magic ARNs_ from your codebase.
  - Individual [IAM Policies](http://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies.html) under which your **Go** functions will execute.  The ability to limit lambda execution privileges helps support [POLP](http://searchsecurity.techtarget.com/definition/principle-of-least-privilege-POLP) and [#SecOps](https://twitter.com/hashtag/secops).
  - Registration of your **Go** function with push-based [AWS Event Sources](http://docs.aws.amazon.com/lambda/latest/dg/intro-core-components.html).

For instance, your service can express in **Go**:

  - [AWS Lambda Event Sources](http://docs.aws.amazon.com/lambda/latest/dg/intro-core-components.html)
    - DynamoDB
    - S3
    - Kinesis
    - SNS
    - SES
  - Other AWS resources
    - S3 buckets with dynamic outputs that your lambda function can [discover at runtime](http://gosparta.io/docs/eventsources/ses/)
    - SNS resources
    - Any other [CloudFormation Resource Type](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-template-resource-type-ref.html)
  - [API Gateway](http://docs.aws.amazon.com/apigateway/latest/developerguide/welcome.html) resources
  - [S3 Static Websites](http://docs.aws.amazon.com/AmazonS3/latest/dev/WebsiteHosting.html)
    - Sparta can provision an S3 bucket with your static resources, including [CORS](http://docs.aws.amazon.com/AmazonS3/latest/dev/cors.html) support

<a href="https://cloudcraft.co/view/8571b3bc-76ef-48c1-8401-0b6ae1d36b4e?key=d44zi4j1pxj00000" rel="Sparta Arch">![Sparta Overview](images/sparta_overview.png)]</a>

Sparta leverages [CloudFormation](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/Welcome.html) to deploy and update your application.  For resources that CloudFormation does not yet support, it uses [Lambda-backed Custom Resources](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/template-custom-resources-lambda.html) so that all application updates support both update and rollback semantics.  CloudFormation resources use stable identifiers whenever possible to preserve [service availability](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks.html) during updates.

## Getting Started

To get started using Sparta, begin with the [Documentation](./docs).

## Problems?

Please file an [issue](https://github.com/mweagle/Sparta/issues/new) in GitHub.  Thanks.

### Other resources

  * Other libraries & frameworks:
    * [Serverless](https://github.com/serverless/serverless)
    * [PAWS](https://github.com/braahyan/PAWS)
    * [Apex](https://github.com/apex/apex)
    * [lambda_proc](https://github.com/jasonmoo/lambda_proc)
    * [go-lambda](https://github.com/xlab/go-lambda)
    * [go-lambda (GRPC)](https://github.com/pilwon/go-lambda)
  * [Serverless Code Blog](https://serverlesscode.com)
  * [AWS Serverless Multi-Tier Architectures Whitepaper](https://d0.awsstatic.com/whitepapers/AWS_Serverless_Multi-Tier_Architectures.pdf)
  * [Lambda limits](http://docs.aws.amazon.com/lambda/latest/dg/limits.html)
  * [The Twelve Days of Lambda](https://aws.amazon.com/blogs/compute/the-twelve-days-of-lambda/)
  * [CloudCraft](http://cloudcraft.co) is a great tool for AWS architecture diagrams
