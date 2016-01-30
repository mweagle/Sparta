+++
author = "Matt Weagle"
date = "2015-11-29T06:50:17"
title = "Limitations"
tags = ["sparta"]
type = "doc"
+++

  * **Go** isn't officially supported by AWS (yet)
    * But, you can [vote](https://twitter.com/awscloud/status/659795641204260864) to make _golang_ officially supported.
    * Because of this, there is a per-container initialization cost of:
        * Copying the embedded binary to _/tmp_
        * Changing the binary permissions
        * Launching it from the new location
        * See the [AWS Forum](https://forums.aws.amazon.com/message.jspa?messageID=583910) for more background
    * Depending on [container reuse](https://aws.amazon.com/blogs/compute/container-reuse-in-lambda/), this initialization penalty (~`700ms`) may prove burdensome.
    * Once **Go** is officially supported, Sparta will eliminate the NodeJS proxying tier to improve performance & lower execution costs.

## AWS Lambda Limitations

  * Lambda is not yet globally available. Please view the [Global Infrastructure](https://aws.amazon.com/about-aws/global-infrastructure/regional-product-services/) page for the latest deployment status.
  * There are [Lambda Limits](http://docs.aws.amazon.com/lambda/latest/dg/limits.html) that may affect your development
  * It's not possible to dynamically set HTTP response headers based on the Lambda response body:
    * https://forums.aws.amazon.com/thread.jspa?threadID=203889
    * https://forums.aws.amazon.com/thread.jspa?threadID=210826
  * Similarly, it's not possible to set proper error response bodies.
