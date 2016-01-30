+++
author = "Matt Weagle"
date = "2015-11-29T06:50:17"
title = "API Gateway - CORS "
tags = ["sparta"]
type = "doc"
+++

[Cross Origin Resource Sharing](https://en.wikipedia.org/wiki/Cross-origin_resource_sharing) defines a protocol by which resources on different domains may establish whether cross site operations are permissible.  

Sparta makes CORS support a single `CORSEnabled` field of the [API](https://godoc.org/github.com/mweagle/Sparta#API) struct:

{{< highlight go >}}
// Register the function with the API Gateway
apiStage := sparta.NewStage("v1")
apiGateway := sparta.NewAPIGateway("SpartaHTML", apiStage)
// Enable CORS s.t. the S3 site can access the resources
apiGateway.CORSEnabled = true
{{< /highlight >}}

Setting the boolean to `true` will add the necessary `OPTIONS` and mock responses to _all_ resources exposed by your API.  See the [SpartaHTML](/docs/s3site) sample for a complete example.

## References
  * [API Gateway Docs](http://docs.aws.amazon.com/apigateway/latest/developerguide/how-to-cors.html)
