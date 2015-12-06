+++
author = "Matt Weagle"
date = "2015-11-29T06:50:17"
title = "Overview"
tags = ["sparta"]
type = "doc"
+++

## <a href="{{< relref "#introduction" >}}">Introduction</a>

One of the most powerful ways to use AWS Lambda is to make function publicly available over HTTPS.  This is accomplished by connecting the AWS Lambda function with the [API Gateway](https://aws.amazon.com/api-gateway/).  In this section we'll start with a simple "echo" example and move on to a lambda function that accepts user parameters and returns an expiring S3 URL.  

## <a href="{{< relref "#concepts" >}}">Concepts</a>

Before moving on to the example, it's suggested you familiarize yourself with the API Gateway concepts.  

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

With that overview, let's start with a simple example.

## <a href="{{< relref "#exampleEcho" >}}">Example 1 - Echo</a>

To start, we'll create a HTTPS accessible lambda function that simply echoes back the contents of the Lambda event.  The source for this is the [SpartaApplication](https://github.com/mweagle/SpartaApplication/blob/master/application.go#L43).

For reference, the `echoS3Event` function is below.

{{< highlight go >}}
func echoS3Event(event *json.RawMessage, context *sparta.LambdaContext, w http.ResponseWriter, logger *logrus.Logger) {
	logger.WithFields(logrus.Fields{
		"RequestID": context.AWSRequestID,
		"Event":     string(*event),
	}).Info("Request received")

	fmt.Fprintf(w, string(*event))
}
{{< /highlight >}}


### <a href="{{< relref "#example1API" >}}">Create the API Gateway</a>

The first requirement is to create a new [API](https://godoc.org/github.com/mweagle/Sparta#API) instance via `sparta.NewAPIGateway()`

{{< highlight go >}}
stage := sparta.NewStage("prod")
apiGateway := sparta.NewAPIGateway("MySpartaAPI", stage)
{{< /highlight >}}

In the example above, we're also including a [Stage](https://godoc.org/github.com/mweagle/Sparta#Stage) value.  A non-`nil` Stage value will cause the registered API to be deployed.  If the Stage value is `nil`, a REST API will be created, but it will not be [deployed](http://docs.aws.amazon.com/apigateway/latest/developerguide/how-to-deploy-api.html) (and therefore not publicly accessible).

### <a href="{{< relref "#example1API" >}}">Create a Resource</a>

The next step is to associate a URL path with the `sparta.LambdaAWSInfo` struct that represents the *Go* function:

{{< highlight go >}}
apiGatewayResource, _ := api.NewResource("/hello/world/test", lambdaFn)
apiGatewayResource.NewMethod("GET")
{{< /highlight >}}

Our [echoS3Event](https://github.com/mweagle/SpartaApplication/blob/master/application.go#L34) only supports `GET`.  We'll see how a single lambda function can support multiple HTTP methods shortly.

### <a href="{{< relref "#example1API" >}}">Provision</a>

The final step is to to provide the API instance to `Sparta.Main()`

{{< highlight go >}}
stage := sparta.NewStage("prod")
apiGateway := sparta.NewAPIGateway("MySpartaAPI", stage)
stackName := "SpartaApplication"
sparta.Main(stackName,
  "Simple Sparta application",
  spartaLambdaData(apiGateway),
  apiGateway)
{{< /highlight >}}

Once the service is successfully provisioned, the `Outputs` key will include the API Gateway Deployed URL (sample):

{{< highlight javascript >}}
[{
    Description: "Sparta Home",
    OutputKey: "SpartaHome",
    OutputValue: "https://github.com/mweagle/Sparta"
  },{
    Description: "Sparta Version",
    OutputKey: "SpartaVersion",
    OutputValue: "0.0.7"
  },{
    Description: "API Gateway URL",
    OutputKey: "URL",
    OutputValue: "https://7ljn63rysd.execute-api.us-west-2.amazonaws.com/prod"
}]
{{< /highlight >}}

Combining the _API Gateway URL_ `OutputValue` with our resource path (_/hello/world/test_), we get the absolute URL to our lambda function: _https://7ljn63rysd.execute-api.us-west-2.amazonaws.com/prod/hello/world/test_

### <a href="{{< relref "#example1Querying" >}}">Querying</a>

Let's query the lambda function and see what the `event` data is at execution time:

{{< highlight nohighlight >}}

curl -vs https://7ljn63rysd.execute-api.us-west-2.amazonaws.com/prod/hello/world/test
*   Trying 54.240.188.223...
* Connected to 7ljn63rysd.execute-api.us-west-2.amazonaws.com (54.240.188.223) port 443 (#0)
* TLS 1.2 connection using TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
* Server certificate: *.execute-api.us-west-2.amazonaws.com
* Server certificate: Symantec Class 3 Secure Server CA - G4
* Server certificate: VeriSign Class 3 Public Primary Certification Authority - G5
> GET /prod/hello/world/test HTTP/1.1
> Host: 7ljn63rysd.execute-api.us-west-2.amazonaws.com
> User-Agent: curl/7.43.0
> Accept: */*
>
< HTTP/1.1 200 OK
< Content-Type: application/json
< Content-Length: 708
< Connection: keep-alive
< Date: Sat, 05 Dec 2015 21:24:44 GMT
< x-amzn-RequestId: 99dfd15d-9b96-11e5-9705-fdd3a4d9c8bf
< X-Cache: Miss from cloudfront
< Via: 1.1 7a0918c01bce16cc9b165fd895f7dc87.cloudfront.net (CloudFront)
< X-Amz-Cf-Id: rx1cVURKTlc3sla3v59Ekz1YMfVdcUWG1QwFKCFPjjLzHzmL_d6r_w==
<
* Connection #0 to host 7ljn63rysd.execute-api.us-west-2.amazonaws.com left intact
{"code":200,"status":"OK","headers":{"date":"Sat, 05 Dec 2015 21:24:44 GMT","content-length":"554","content-type":"text/plain; charset=utf-8"},"results":{"method":"GET","body":{},"headers":{"Accept":"*/*","CloudFront-Forwarded-Proto":"https","CloudFront-Is-Desktop-Viewer":"true","CloudFront-Is-Mobile-Viewer":"false","CloudFront-Is-SmartTV-Viewer":"false","CloudFront-Is-Tablet-Viewer":"false","CloudFront-Viewer-Country":"US","Via":"1.1 7a0918c01bce16cc9b165fd895f7dc87.cloudfront.net (CloudFront)","X-Amz-Cf-Id":"xIGP8KHDpP6AEkAt3jdOcUDVojJ_86vFDRhsgfLKgdAsBHHzI7iNVQ==","X-Forwarded-For":"50.135.43.1, 54.239.137.4","X-Forwarded-Port":"443","X-Forwarded-Proto":"https"},"queryParams":{},"pathParams":{}}}
{{< /highlight >}}

Pretty-printing the response body to make things more readable:

{{< highlight json >}}
{
  "code": 200,
  "status": "OK",
  "headers": {
    "date": "Sat, 05 Dec 2015 21:24:44 GMT",
    "content-length": "554",
    "content-type": "text/plain; charset=utf-8"
  },
  "results": {
    "method": "GET",
    "body": {},
    "headers": {
      "Accept": "*/*",
      "CloudFront-Forwarded-Proto": "https",
      "CloudFront-Is-Desktop-Viewer": "true",
      "CloudFront-Is-Mobile-Viewer": "false",
      "CloudFront-Is-SmartTV-Viewer": "false",
      "CloudFront-Is-Tablet-Viewer": "false",
      "CloudFront-Viewer-Country": "US",
      "Via": "1.1 7a0918c01bce16cc9b165fd895f7dc87.cloudfront.net (CloudFront)",
      "X-Amz-Cf-Id": "xIGP8KHDpP6AEkAt3jdOcUDVojJ_86vFDRhsgfLKgdAsBHHzI7iNVQ==",
      "X-Forwarded-For": "50.135.43.1, 54.239.137.4",
      "X-Forwarded-Port": "443",
      "X-Forwarded-Proto": "https"
    },
    "queryParams": {},
    "pathParams": {}
  }
}
{{< /highlight >}}

While this demonstrates that our lambda function is publicly accessible, it's not immediately obvious where the `*event` data is being populated.

### <a href="{{< relref "#example1Mapping" >}}">Mapping Templates</a>

The event data that's actually supplied to `echoS3Event` function is returned in the responses `results` results.  This content is what the API Gateway forwards sends as part of the integration mapping.  The sibling `code`, `status`, and `headers` keys will be explained shortly.

When the API Gateway Method is defined, it specifies the whitelisted query params and header values that should be forwarded to the integration target.  For this example, we're not whitelisting any params, so those fields are empty.  Then for each integration target (which can be AWS Lambda, a mock, or a HTTP Proxy), it's possible to transform the API Gateway request data and whitelisted arguments into a format that's more amenable to the target.

Sparta uses a pass-through template that passes all valid data.  The [Apache Velocity](http://velocity.apache.org) template that [Sparta uses](https://raw.githubusercontent.com/mweagle/Sparta/master/resources/gateway/inputmapping_json.vtl) is:

{{< highlight nohighlight >}}
#*
Provide an automatic pass through template that transforms all inputs
into the JSON payload sent to a golang function

See
  https://forums.aws.amazon.com/thread.jspa?threadID=220274&tstart=0
  http://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-mapping-template-reference.html
*#
{
  "method": "$context.httpMethod",
  "body" : $input.json('$'),
  "headers": {
    #foreach($param in $input.params().header.keySet())
    "$param": "$util.escapeJavaScript($input.params().header.get($param))" #if($foreach.hasNext),#end

    #end
  },
  "queryParams": {
    #foreach($param in $input.params().querystring.keySet())
    "$param": "$util.escapeJavaScript($input.params().querystring.get($param))" #if($foreach.hasNext),#end

    #end
  },
  "pathParams": {
    #foreach($param in $input.params().path.keySet())
    "$param": "$util.escapeJavaScript($input.params().path.get($param))" #if($foreach.hasNext),#end

    #end
  }
}
{{< /highlight >}}

This template forwards all whitelisted data & body to the lambda function.  You can see by switching on the `method` field would permit a single function to service multiple HTTP method names.

The next example will show how to unmarshal this data and perform request-specific actions.  

### <a href="{{< relref "#example1ProxyingEnvelope" >}}">Proxying Envelope</a>

The mapping template explains the content of the `results` property, but not the other fields (`code`, `status`, `headers`).  Those fields are injected by the NodeJS proxying tier as part of translating the *Go* HTTP response to a Lambda compatible result.  

A primary benefit of this envelope is to provide an automatic mapping from Integration Response Regular Expression mappings to Method Response codes.  If you look at the **Integration Response** section of the _/hello/world/test_ resource in the Console, you'll see a list of Regular Expression matches:

![API Gateway](/images/apigateway/IntegrationMapping.png)

The regular expressions are used to translate the integration response, which is just a blob of text provided to `context.done()`, into API Gateway Method responses.  Sparta annotates your lambda functions response with *Go*'s [HTTP StatusText](https://golang.org/src/net/http/status.go) values based on the HTTP status code your lambda function produced.  Sparta also provides a corresponding Method Response entry for all valid HTTP codes:

![API Gateway](/images/apigateway/MethodResponse.png)

These mappings are defaults, and it's possible to override either one by providing a non-zero length values to either:

  * [Integration.Responses](https://godoc.org/github.com/mweagle/Sparta#Integration).  See the [DefaultIntegrationResponses](https://github.com/mweagle/Sparta/blob/master/apigateway.go#L60) for the default values.
  * [Method.Responses](https://godoc.org/github.com/mweagle/Sparta#Method).  See the [DefaultMethodResponses](https://godoc.org/github.com/mweagle/Sparta#DefaultMethodResponses) for the default method response mappings.

### <a href="{{< relref "#example1WrappingUp" >}}">Wrapping Up</a>

Now that we know what data is actually being sent to our API Gateway-connected Lambda function, we'll move on to performing a more complex operation, including returning a custom HTTP response body.

## <a href="{{< relref "#exampleS3" >}}">Example 2 - Accepting Input</a>

TODO: Update with docs for [s3ItemInfo](https://github.com/mweagle/SpartaImager/blob/master/application.go#L149).

## Other Resources
  * [Walkthrough: API Gateway and Lambda Functions](http://docs.aws.amazon.com/apigateway/latest/developerguide/getting-started.html)
