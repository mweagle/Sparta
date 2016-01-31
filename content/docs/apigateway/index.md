+++
author = "Matt Weagle"
date = "2015-11-29T06:50:17"
title = "API Gateway"
tags = ["sparta"]
type = "doc"
+++

One of the most powerful ways to use AWS Lambda is to make function publicly available over HTTPS.  This is accomplished by connecting the AWS Lambda function with the [API Gateway](https://aws.amazon.com/api-gateway/).  In this section we'll start with a simple "echo" example and move on to a lambda function that accepts user parameters and returns an expiring S3 URL.  

  * [Example 1 - Echo Event](/docs/apigateway/example1)
  * [Example 2 - User Input & JSON Response](/docs/apigateway/example2)
  * [Example 3 - Request Context](/docs/apigateway/example3)
  * [Example 4 - Slack SlashCommand](/docs/apigateway/slack)
  * [Example 5 - CORS](/docs/apigateway/cors)
    
## <a href="{{< relref "#concepts" >}}">Concepts</a>

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

## Other Resources
  * [Walkthrough: API Gateway and Lambda Functions](http://docs.aws.amazon.com/apigateway/latest/developerguide/getting-started.html)
