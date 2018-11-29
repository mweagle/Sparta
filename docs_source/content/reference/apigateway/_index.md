---
date: 2016-03-09T19:56:50+01:00
title:
pre: "<b>API Gateway</b>&nbsp;<i class='fas fa-fw fa-globe'></i>"
alwaysopen: false
weight: 190
---

## Examples

One of the most powerful ways to use AWS Lambda is to make function publicly available over HTTPS.  This is accomplished by connecting the AWS Lambda function with the [API Gateway](https://aws.amazon.com/api-gateway/).  In this section we'll start with a simple "echo" example and move on to a lambda function that accepts user parameters and returns an expiring S3 URL.

* [Example 1 - Echo Event](/reference/apigateway/echo_event)
* [Example 2 - User Input & JSON Response](/reference/apigateway/user_input)
* [Example 3 - Request Context](/reference/apigateway/context)
* [Example 4 - Slack SlashCommand](/reference/apigateway/slack)
* [Example 5 - CORS](/reference/apigateway/cors)

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

## API Gateway Request Types

AWS Lambda supports multiple [function signatures](https://docs.aws.amazon.com/lambda/latest/dg/go-programming-model-handler-types.html). Some supported signatures include structured types, which are JSON un/marshalable structs that are automatically managed.

To simplify handling API Gateway requests, Sparta exposes the [APIGatewayEnvelope](https://godoc.org/github.com/mweagle/Sparta/aws/events#APIGatewayEnvelope) type. This type provides an embeddable struct type whose fields and JSON serialization match up with the [Velocity Template](https://github.com/mweagle/Sparta/blob/master/resources/provision/apigateway/inputmapping_json.vtl) that's applied to the incoming API Gateway request.

To use the `APIGatewayEnvelope` type with your own custom request body, create a set of types as in:

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

Then reference your custom type in your lambda function as in:

```go
func myLambdaFunction(ctx context.Context, apiGatewayRequest FeedbackRequest) (map[string]string, error) {
  language := apiGatewayRequest.Body.Language
  ...
}
```

## Custom HTTP Headers

API Gateway supports returning custom HTTP headers whose values are extracted from your response payload.

Assume your Sparta lambda function returns a JSON struct as in:

```go
// API response struct
type helloWorldResponse struct {
  Location    string `json:"location"`
  Message     string `json:"message"`
}
```

To extract the `location` field and promote it to the HTTP `Location` header, you must configure the [response data mappings](http://docs.aws.amazon.com/apigateway/latest/developerguide/request-response-data-mappings.html
):



```go
//
// Promote the location key value to an HTTP header
//
  lambdaFn := sparta.HandleAWSLambda(
    sparta.LambdaName(helloWorldResponseFunc),
    helloWorldResponseFunc,
    sparta.IAMRoleDefinition{})
	apiGatewayResource, _ := api.NewResource("/hello", lambdaFn)

apiGWMethod, _ := apiGatewayResource.NewMethod("GET", http.StatusOK)
apiGWMethod.Responses[http.StatusOK].Parameters = map[string]bool{
  "method.response.header.Location": true,
}
apiGWMethod.Integration.Responses[http.StatusOK].Parameters["method.response.header.Location"] =
  "integration.response.body.location"
```

Note that as the `helloWorldResponse` structured type is serialized to the _body_ property of the response, we include that path selector in the _integration.response.body.location_ value.

See the related [AWS Forum thread](https://forums.aws.amazon.com/thread.jspa?threadID=199443).

## Other Resources
  * [Walkthrough: API Gateway and Lambda Functions](http://docs.aws.amazon.com/apigateway/latest/developerguide/getting-started.html)
