+++
author = "Matt W."
comments = true
date = "2015-11-29T06:50:17"
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
  <iframe width="560" height="315" src="https://www.youtube.com/embed/y-0Wf2Zyi5Q?start=1742" frameborder="0" allowfullscreen></iframe>
</div>

Sparta provides a framework to build & deploy *Go* functions in [AWS Lambda](https://aws.amazon.com/lambda/). While *Go* is not _yet_ officially supported by AWS Lambda (see [poll](https://twitter.com/awscloud/status/659795641204260864)), it's possible to bundle & launch arbitrary executables in Lambda.  

Sparta provides a HTTP-based translation layer between the proper [NodeJS](http://docs.aws.amazon.com/lambda/latest/dg/programming-model.html) environment and your *Go* binary.  In addition to this translation layer, Sparta is also able to:

  * Manage S3 and SNS-based [event sources](http://docs.aws.amazon.com/lambda/latest/dg/intro-core-components.html#intro-core-components-event-sources).
      * A Lambda function may be invoked in response to S3 or SNS broadcasted events.
  * Provision an HTTPS [API Gateway](https://aws.amazon.com/api-gateway/details/) service that allows Lambda functions to be publicly invoked.
  * Produce diagrams of Lambda & event source interactions

## Getting Started

To get started using Sparta, begin with the [Documentation](./docs).

## Problems?

Please file an [issue](https://github.com/mweagle/Sparta/issues/new) in GitHub.  Thanks.

### Other resources


  * [Lambda limits](http://docs.aws.amazon.com/lambda/latest/dg/limits.html)
  * [The Twelve Days of Lambda](https://aws.amazon.com/blogs/compute/the-twelve-days-of-lambda/)
