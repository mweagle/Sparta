---
date: 2016-03-09T19:56:50+01:00
title: CORS
weight: 20
---

[Cross Origin Resource Sharing](https://en.wikipedia.org/wiki/Cross-origin_resource_sharing) defines a protocol by which resources on different domains may establish whether cross site operations are permissible.

Sparta makes CORS support a single `CORSEnabled` field of the [API](https://godoc.org/github.com/mweagle/Sparta#API) struct:

```go

// Register the function with the API Gateway
apiStage := sparta.NewStage("v1")
apiGateway := sparta.NewAPIGateway("SpartaHTML", apiStage)
// Enable CORS s.t. the S3 site can access the resources
apiGateway.CORSEnabled = true

```

Setting the boolean to `true` will add the necessary `OPTIONS` and mock responses to _all_ resources exposed by your API.  See the [SpartaHTML](/reference/s3site) sample for a complete example.

# Customization

Sparta provides two ways to customize the CORS headers available:

  * Via the [apigateway.CORSOptions](https://godoc.org/github.com/mweagle/Sparta#CORSOptions) field.
  * Customization may use the [S3Site.CloudformationS3ResourceName](https://godoc.org/github.com/mweagle/Sparta#S3Site) to get the _WebsiteURL_ value
  so that the CORS origin options can be minimally scoped.

# References
  * [API Gateway Docs](http://docs.aws.amazon.com/apigateway/latest/developerguide/how-to-cors.html)
