---
date: 2016-03-09T19:56:50+01:00
title: Limitations
weight: 10
---

# AWS Lambda Limitations

  * Lambda is not yet globally available. Please view the [Global Infrastructure](https://aws.amazon.com/about-aws/global-infrastructure/regional-product-services/) page for the latest deployment status.
  * There are [Lambda Limits](http://docs.aws.amazon.com/lambda/latest/dg/limits.html) that may affect your development
  * It's not possible to dynamically set HTTP response headers based on the Lambda response body:
    * https://forums.aws.amazon.com/thread.jspa?threadID=203889
    * https://forums.aws.amazon.com/thread.jspa?threadID=210826
  * Similarly, it's not possible to set proper error response bodies.
