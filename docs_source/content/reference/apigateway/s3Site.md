---
date: 2016-03-09T19:56:50+01:00
title: S3 Sites with CORS
weight: 150
---

Sparta supports provisioning an S3-backed [static website](http://docs.aws.amazon.com/AmazonS3/latest/dev/WebsiteHosting.html) as part of provisioning. We'll walk through provisioning a minimal [Bootstrap](http://getbootstrap.com) website that accesses API Gateway lambda functions provisioned by a single service in this example.

The source for this is the [SpartaHTML](https://github.com/mweagle/SpartaHTML) example application.

## Lambda Definition

We'll start by creating a very simple lambda function:

```go
import (
  spartaAPIGateway "github.com/mweagle/Sparta/v3/aws/apigateway"
  spartaAWSEvents "github.com/mweagle/Sparta/v3/aws/events"
)
type helloWorldResponse struct {
  Message string
  Request spartaAWSEvents.APIGatewayRequest
}

////////////////////////////////////////////////////////////////////////////////
// Hello world event handler
func helloWorld(ctx context.Context,
  gatewayEvent spartaAWSEvents.APIGatewayRequest) (*spartaAPIGateway.Response, error) {
  logger, loggerOk := ctx.Value(sparta.ContextKeyLogger).(*zerolog.Logger)
  if loggerOk {
    logger.Info("Hello world structured log message")
  }
  // Return a message, together with the incoming input...
  return spartaAPIGateway.NewResponse(http.StatusOK, &helloWorldResponse{
    Message: fmt.Sprintf("Hello world 🌏"),
    Request: gatewayEvent,
  }), nil
}
```

This lambda function returns a reply that consists of the inbound
request plus a sample message. See the API Gateway [examples](/reference/apigateway)
for more information.

## API Gateway

The next step is to create an API Gateway instance and Stage, so that the API will be publicly available.

```go
apiStage := sparta.NewStage("v1")
apiGateway := sparta.NewAPIGateway("SpartaHTML", apiStage)
```

Since we want to be able to access this API from another domain (the one provisioned by the S3 bucket), we'll need to [enable CORS](http://docs.aws.amazon.com/apigateway/latest/developerguide/how-to-cors.html) as well:

```go
// Enable CORS s.t. the S3 site can access the resources
apiGateway.CORSEnabled = true
```

Finally, we register the `helloWorld` lambda function with an API Gateway resource:

```go

func spartaLambdaFunctions(api *sparta.API) []*sparta.LambdaAWSInfo {
  var lambdaFunctions []*sparta.LambdaAWSInfo
  lambdaFn, _ := sparta.NewAWSLambda(sparta.LambdaName(helloWorld),
    helloWorld,
    sparta.IAMRoleDefinition{})

  if nil != api {
    apiGatewayResource, _ := api.NewResource("/hello", lambdaFn)
    _, err := apiGatewayResource.NewMethod("GET", http.StatusOK)
    if nil != err {
      panic("Failed to create /hello resource")
    }
  }
  return append(lambdaFunctions, lambdaFn)
}
```

## S3 Site

The next part is to define the S3 site resources via `sparta.NewS3Site(localFilePath)`. The _localFilePath_ parameter
typically points to a directory, which will be:

1. Recursively ZIP'd
1. Posted to S3 alongside the Lambda code archive and CloudFormation Templates
1. Dynamically unpacked by a CloudFormation CustomResource during `provision` to a new S3 bucket.

## Provision

Putting it all together, our `main()` function looks like:

```go

////////////////////////////////////////////////////////////////////////////////
// Main
func main() {
  // Register the function with the API Gateway
  apiStage := sparta.NewStage("v1")
  apiGateway := sparta.NewAPIGateway("SpartaHTML", apiStage)
  // Enable CORS s.t. the S3 site can access the resources
  apiGateway.CORSEnabled = true

  // Provision a new S3 bucket with the resources in the supplied subdirectory
  s3Site, _ := sparta.NewS3Site("./resources")

  // Deploy it
  sparta.Main("SpartaHTML",
    fmt.Sprintf("Sparta app that provisions a CORS-enabled API Gateway together with an S3 site"),
    spartaLambdaFunctions(apiGateway),
    apiGateway,
    s3Site)
}
```

which can be provisioned using the standard [command line](/reference/commandline) option.

The _Outputs_ section of the `provision` command includes the hostname of our new S3 site:

```nohighlight
INFO[0092] ────────────────────────────────────────────────
INFO[0092] Stack Outputs
INFO[0092] ────────────────────────────────────────────────
INFO[0092] S3SiteURL                                     Description="S3 Website URL" Value="http://spartahtml-mweagle-s3site89c05c24a06599753eb3ae4e-9kil6qlqk0yt.s3-website-us-west-2.amazonaws.com"
INFO[0092] APIGatewayURL                                 Description="API Gateway URL" Value="https://ksuo0qlc3m.execute-api.us-west-2.amazonaws.com/v1"
INFO[0092] ────────────────────────────────────────────────
```

Open your browser to the `S3SiteURL` value (eg: _http://spartahtml-mweagle-s3site89c05c24a06599753eb3ae4e-9kil6qlqk0yt.s3-website-us-west-2.amazonaws.com_) and view the deployed site.

## Discover

An open issue is how to communicate the dynamically assigned API Gateway hostname to the dynamically provisioned S3 site.

As part of expanding the ZIP archive to a target S3 bucket, Sparta also creates a _MANIFEST.json_ discovery file with discovery information. If your application has provisioned an APIGateway this JSON file will include that dynamically assigned URL as in:

1. **MANIFEST.json**

```json
{
  "APIGatewayURL": {
    "Description": "API Gateway URL",
    "Value": "https://ksuo0qlc3m.execute-api.us-west-2.amazonaws.com/v1"
  }
}
```

### Notes

- See the [Medium](https://read.acloud.guru/go-aws-lambda-building-an-html-website-with-api-gateway-and-lambda-for-go-using-sparta-5e6fe79f63ef) post for an additional walk through this sample.
