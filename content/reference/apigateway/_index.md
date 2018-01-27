---
date: 2016-03-09T19:56:50+01:00
title: API Gateway
weight: 10
menu:
  main:
    parent: Documentation
    identifier: apigateway
    weight: 20
---

## Examples

One of the most powerful ways to use AWS Lambda is to make function publicly available over HTTPS.  This is accomplished by connecting the AWS Lambda function with the [API Gateway](https://aws.amazon.com/api-gateway/).  In this section we'll start with a simple "echo" example and move on to a lambda function that accepts user parameters and returns an expiring S3 URL.

  * [Example 1 - Echo Event](/docs/apigateway/example1)
  * [Example 2 - User Input & JSON Response](/docs/apigateway/example2)
  * [Example 3 - Request Context](/docs/apigateway/example3)
  * [Example 4 - Slack SlashCommand](/docs/apigateway/slack)
  * [Example 5 - CORS](/docs/apigateway/cors)

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

With that overview, let's start with a simple [example](/docs/apigateway/example1).


## Custom HTTP Headers

API Gateway supports returning custom HTTP headers whose values are extracted from your response payload.

Assume your Sparta lambda function returns a JSON struct as in:

{{< highlight go >}}
// API response struct
type helloWorldResponse struct {
  Location string `json:"location"`
  Body     string `json:"body"`
}
{{< /highlight >}}

To extract the `location` field and promote it to the HTTP `Location` header, you must configure the [response data mappings](http://docs.aws.amazon.com/apigateway/latest/developerguide/request-response-data-mappings.html
):


{{< highlight go >}}
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
{{< /highlight >}}

See the related [AWS Forum thread](https://forums.aws.amazon.com/thread.jspa?threadID=199443).


## Other Resources
  * [Walkthrough: API Gateway and Lambda Functions](http://docs.aws.amazon.com/apigateway/latest/developerguide/getting-started.html)
