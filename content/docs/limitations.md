---
date: 2016-03-09T19:56:50+01:00
title: Limitations
weight: 10
menu:
  main:
    parent: Documentation
    identifier: limitations
    weight: 0
---

# Sparta Limitations

  * **Go** isn't officially supported by AWS (yet)
    * But, you can [vote](https://twitter.com/awscloud/status/659795641204260864) to make _golang_ officially supported.
    * Because of this, for the default NodeJS proxying is a per-container initialization cost of:
        * Copying the embedded binary to _/tmp_
        * Changing the binary permissions
        * Launching it from the new location
        * See the [AWS Forum](https://forums.aws.amazon.com/message.jspa?messageID=583910) for more background
    * Depending on [container reuse](https://aws.amazon.com/blogs/compute/container-reuse-in-lambda/), this initialization penalty (~`700ms`) may prove burdensome.
  * Alternatively, look at leveraging Sparta's [cgo](https://medium.com/@mweagle/see-lambda-go-e39b526c1020) option to package your service into a shared library that's proxied by Python for significantly improved performance

# AWS Lambda Limitations

  * Lambda is not yet globally available. Please view the [Global Infrastructure](https://aws.amazon.com/about-aws/global-infrastructure/regional-product-services/) page for the latest deployment status.
  * There are [Lambda Limits](http://docs.aws.amazon.com/lambda/latest/dg/limits.html) that may affect your development
  * It's not possible to dynamically set HTTP response headers based on the Lambda response body:
    * https://forums.aws.amazon.com/thread.jspa?threadID=203889
    * https://forums.aws.amazon.com/thread.jspa?threadID=210826
  * Similarly, it's not possible to set proper error response bodies.
