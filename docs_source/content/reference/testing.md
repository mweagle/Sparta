---
date: 2016-03-09T19:56:50+01:00
title: Testing
weight: 800
---

While developing Sparta lambda functions it may be useful to test them locally without needing to `provision` each new code change.  You can test your lambda functions
using standard `go test` functionality.

To create proper event types, consider:

* [AWS Lambda Go](https://godoc.org/github.com/aws/aws-lambda-go/events) types
* Sparta types
* Use [NewAPIGatewayMockRequest](https://godoc.org/github.com/mweagle/Sparta/aws/events#NewAPIGatewayMockRequest) to generate API Gateway style requests.
