---
date: 2020-12-31 21:48:47
title: Testing
weight: 800
---

## Unit Tests

While developing Sparta lambda functions it may be useful to test them locally without needing to `provision` each new code change. You can test your lambda functions
using standard `go test` functionality.

To create proper event types, consider:

- [AWS Lambda Go](https://godoc.org/github.com/aws/aws-lambda-go/events) types
- Sparta types
- Use [NewAPIGatewayMockRequest](https://godoc.org/github.com/mweagle/Sparta/aws/events#NewAPIGatewayMockRequest) to generate API Gateway style requests.

## Acceptance Tests

Use cloudtest package
