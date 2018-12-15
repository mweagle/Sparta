---
date: 2016-03-09T19:56:50+01:00
pre: "<b>API Gateway</b>"
alwaysopen: false
weight: 100
---

# API Gateway

One of the most powerful ways to use AWS Lambda is to make function publicly available over HTTPS.  This is accomplished by connecting the AWS Lambda function with the [API Gateway](https://aws.amazon.com/api-gateway/).  In this section we'll start with a simple "echo" example and move on to a lambda function that accepts user parameters and returns an expiring S3 URL.

{{% children %}}

## Concepts

Before moving on to the examples, it's suggested you familiarize yourself with the API Gateway concepts.

* [Gettting Started with Amazon API Gateway](http://docs.aws.amazon.com/apigateway/latest/developerguide/getting-started-intro.html)

The API Gateway presents a powerful and complex domain model.  In brief, to integrate with the API Gateway, a service must:

  1. Define one or more AWS Lambda functions
  1. Create an API Gateway REST API instance
  1. Create one or more resources associated with the REST API
  1. Create one or more methods for each resource
  1. For each method:
      1. Define the method request params
      1. Define the integration request mapping
      1. Define the integration response mapping
      1. Define the method response mapping
  1. Create a stage for a REST API
  1. Deploy the given stage

See a the [echo example](/reference/apigateway/echo_event) for a complete version.

## Request Types

AWS Lambda supports multiple [function signatures](https://docs.aws.amazon.com/lambda/latest/dg/go-programming-model-handler-types.html). Some supported signatures include structured types, which are JSON un/marshalable structs that are automatically managed.

To simplify handling API Gateway requests, Sparta exposes the [APIGatewayEnvelope](https://godoc.org/github.com/mweagle/Sparta/aws/events#APIGatewayEnvelope) type. This type provides an embeddable struct type whose fields and JSON serialization match up with the [Velocity Template](https://github.com/mweagle/Sparta/blob/master/resources/provision/apigateway/inputmapping_json.vtl) that's applied to the incoming API Gateway request.

Embed the [APIGatewayEnvelope](https://godoc.org/github.com/mweagle/Sparta/aws/events#APIGatewayEnvelope) type in your own lambda's request type as in:

```go
type FeedbackBody struct {
  Language string `json:"lang"`
  Comment  string `json:"comment"`
}

type FeedbackRequest struct {
  spartaEvents.APIGatewayEnvelope
  Body FeedbackBody `json:"body"`
}
```

Then accept your custom type in your lambda function as in:

```go
func myLambdaFunction(ctx context.Context, apiGatewayRequest FeedbackRequest) (map[string]string, error) {
  language := apiGatewayRequest.Body.Language
  ...
}
```

## Response Types

The API Gateway [response mappings](https://docs.aws.amazon.com/apigateway/latest/developerguide/mappings.html) must make
assumptions about the shape of the Lambda response. The default _application/json_ mapping template is:

{{% import file="./static/source/resources/provision/apigateway/outputmapping_json.vtl" language="nohighlight" %}}

This template assumes that your response type has the following JSON shape:

```json
{
  "code" : int,
  "body" : {...},
  "headers": {...}
}
```

The [apigateway.NewResponse](https://godoc.org/github.com/mweagle/Sparta/aws/apigateway#NewResponse) constructor
is a utility function to produce a canonical version of this response shape. Note that `header` keys must be lower-cased.

To return a different structure change the content-specific mapping templates defined by the
[IntegrationResponse](https://godoc.org/github.com/mweagle/Sparta#IntegrationResponse). See the
[mapping template reference](https://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-mapping-template-reference.html) for more information.

## Custom HTTP Headers

API Gateway supports returning custom HTTP headers whose values are extracted from your response. To return custom HTTP
headers using the default VTL mappings, provide them as the optional third `map[string]string` argument to
[NewResponse](https://godoc.org/github.com/mweagle/Sparta/aws/apigateway#NewResponse) as in:

```go
func helloWorld(ctx context.Context,
  gatewayEvent spartaAWSEvents.APIGatewayRequest) (*spartaAPIGateway.Response, error) {

  logger, loggerOk := ctx.Value(sparta.ContextKeyLogger).(*logrus.Logger)
  if loggerOk {
    logger.Info("Hello world structured log message")
  }

  // Return a message, together with the incoming input...
  return spartaAPIGateway.NewResponse(http.StatusOK, &helloWorldResponse{
    Message: fmt.Sprintf("Hello world 🌏"),
    Request: gatewayEvent,
  },
    map[string]string{
      "X-Response": "Some-value",
    }), nil
}
```

## Other Resources

* [Walkthrough: API Gateway and Lambda Functions](http://docs.aws.amazon.com/apigateway/latest/developerguide/getting-started.html)
* [Use a Mapping Template to Override an API's Request and Response Parameters and Status Codes](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-override-request-response-parameters.html)
