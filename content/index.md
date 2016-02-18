+++
author = "Matt Weagle"
comments = true
date = "2016-02-16T05:36:20Z"
draft = false
share = false
title = "homepage"
+++

<br />
<div class="jumbotron">
<img src="images/SpartaLogoNoDomain.png" alt="Sparta shield" height="128">
<h2>A <b>Go</b> framework for <a href="https://aws.amazon.com/lambda">AWS Lambda</a> microservices</h2>

  <hr />
  <blockquote>
    <p>"No Server Is Easier To Manage Than No Server."</p>
    <footer>Werner Vogels <cite title="Source Title">AWS re:Invent 2015</cite></footer>
  </blockquote>  
  <iframe width="50%" height="200" src="https://www.youtube.com/embed/y-0Wf2Zyi5Q?start=1742" frameborder="0" allowfullscreen></iframe>
</div>

# Overview

Sparta provides a framework that enables you to deploy a set of **Go** HTTP request/response handlers to [AWS Lambda](https://aws.amazon.com/lambda/).  Sparta is more than a deployment tool though, as it offers the ability, in **Go**, to fully define and manage the **other AWS resources and security policies** that constitute your service:


  * [CloudFormation](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-template-resource-type-ref.html) resources that should be created during your service's provisioning.  
    * Sparta directly supports supplementary infrastructure provisioning by providing hooks to annotate the CloudFormation Template.
  - Discovery of those dependent resources' CloudFormation outputs (`Ref` && `Fn::Att` values) at AWS Lambda execution time
    * This enables a service to close over its AWS infrastructure requirements.  Eliminate hardcoded _Magic ARNs_ from your codebase & move towards [immutable infrastructure](http://chadfowler.com/blog/2013/06/23/immutable-deployments/).
  - Dedicated [IAM Roles](http://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies.html) under which your **Go** functions will execute.  
    * The ability to limit lambda execution privileges helps support [POLP](http://searchsecurity.techtarget.com/definition/principle-of-least-privilege-POLP) and [#SecOps](https://twitter.com/hashtag/secops).  
    * The IAM Policy entries can reference dynamically assigned AWS ARN values.
  * Optional registration of your **Go** function with push-based [AWS Event Sources](http://docs.aws.amazon.com/lambda/latest/dg/intro-core-components.html).
    * Sparta handles the remote configuration of other AWS services as part of your service's deployment.

For instance, your service can express in **Go**:

  * [AWS Lambda Event Sources](http://docs.aws.amazon.com/lambda/latest/dg/intro-core-components.html)
    * DynamoDB
    * S3
    * Kinesis
    * SNS
    * SES
    * CloudWatch Events
    * CloudWatch Logs
  * Other AWS resources
    * S3 buckets with dynamic outputs that your lambda function can [discover at runtime](http://gosparta.io/docs/eventsources/ses/)
    * SNS resources
    * Any other [CloudFormation Resource Type](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-template-resource-type-ref.html)
  * [API Gateway](http://docs.aws.amazon.com/apigateway/latest/developerguide/welcome.html) resources that trigger your lambda functions
    * Sparta automatically creates [API Gateway Mapping Templates](http://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-mapping-template-reference.html) with all request data, including user-defined whitelisted parameters, so that you can focus on your core application logic.
  * [S3 Static Websites](http://docs.aws.amazon.com/AmazonS3/latest/dev/WebsiteHosting.html)
    - Sparta can provision an S3 bucket with your static resources, including [CORS](http://docs.aws.amazon.com/AmazonS3/latest/dev/cors.html) support

Sparta exclusively relies on [CloudFormation](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/Welcome.html) to deploy and update your application.  For resources that CloudFormation does not yet support, it uses [Lambda-backed Custom Resources](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/template-custom-resources-lambda.html) so that all service updates support both update and rollback semantics.  Sparta's automatically generated CloudFormation resources use content-based logical IDs  whenever possible to preserve [service availability](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks.html) during updates.

<hr />
<a href="https://cloudcraft.co/view/8571b3bc-76ef-48c1-8401-0b6ae1d36b4e?key=d44zi4j1pxj00000" rel="Sparta Arch">![Sparta Overview](images/sparta_overview.png)]</a>


# Getting Started

To get started using Sparta, begin with the [Documentation](./docs).

# Administration
  - Problems?  Please open an [issue](https://github.com/mweagle/Sparta/issues/new) in GitHub.
  - See [Trello](https://trello.com/b/WslDce70/sparta) for the Sparta backlog.

## Other resources

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
