+++
author = "Matt Weagle"
date = "2015-11-29T06:50:17"
title = "API Gateway - Echo"
tags = ["sparta"]
type = "doc"
+++

To start, we'll create a HTTPS accessible lambda function that simply echoes back the contents of the Lambda event.  The source for this is the [SpartaApplication](https://github.com/mweagle/SpartaApplication).

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

The next step is to associate a URL path with the `sparta.LambdaAWSInfo` struct that represents the **Go** function:

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
  apiGateway,
  nil)
{{< /highlight >}}

Once the service is successfully provisioned, the `Outputs` key will include the API Gateway Deployed URL (sample):

{{< highlight javascript >}}
INFO[0113] Stack output   Description=API Gateway URL Key=APIGatewayURL Value=https://7ljn63rysd.execute-api.us-west-2.amazonaws.com/prod
INFO[0113] Stack output   Description=Sparta Home Key=SpartaHome Value=https://github.com/mweagle/Sparta
INFO[0113] Stack output   Description=Sparta Version Key=SpartaVersion Value=0.1.0
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
{"method":"GET","body":{},"headers":{"Accept":"*/*","CloudFront-Forwarded-Proto":"https","CloudFront-Is-Desktop-Viewer":"true","CloudFront-Is-Mobile-Viewer":"false","CloudFront-Is-SmartTV-Viewer":"false","CloudFront-Is-Tablet-Viewer":"false","CloudFront-Viewer-Country":"US","Via":"1.1 5c98e8df8806ae26f9ae3c33615610d2.cloudfront.net (CloudFront)","X-Amz-Cf-Id":"sRMCwKpH3jIPbwgIo4pPHv_YJXEo9KEojEFw8yrljFVP2krJbyewLg==","X-Forwarded-For":"50.135.43.1, 54.240.158.211","X-Forwarded-Port":"443","X-Forwarded-Proto":"https"},"queryParams":{},"pathParams":{},"context":{"apiId":"bmik0opc3l","method":"GET","requestId":"c113fd3b-a76b-11e5-b5e6-4ff04e5da412","resourceId":"mp2mrk","resourcePath":"/hello/world/test","stage":"prod","identity":{"accountId":"","apiKey":"","caller":"","cognitoAuthenticationProvider":"","cognitoAuthenticationType":"","cognitoIdentityId":"","cognitoIdentityPoolId":"","sourceIp":"50.135.43.1","user":"","userAgent":"curl/7.43.0","userArn":""}}}

{{< /highlight >}}

Pretty-printing the response body to make things more readable:

{{< highlight json >}}
{
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
    "Via": "1.1 5c98e8df8806ae26f9ae3c33615610d2.cloudfront.net (CloudFront)",
    "X-Amz-Cf-Id": "sRMCwKpH3jIPbwgIo4pPHv_YJXEo9KEojEFw8yrljFVP2krJbyewLg==",
    "X-Forwarded-For": "50.135.43.1, 54.240.158.211",
    "X-Forwarded-Port": "443",
    "X-Forwarded-Proto": "https"
  },
  "queryParams": {},
  "pathParams": {},
  "context": {
    "apiId": "bmik0opc3l",
    "method": "GET",
    "requestId": "c113fd3b-a76b-11e5-b5e6-4ff04e5da412",
    "resourceId": "mp2mrk",
    "resourcePath": "/hello/world/test",
    "stage": "prod",
    "identity": {
      "accountId": "",
      "apiKey": "",
      "caller": "",
      "cognitoAuthenticationProvider": "",
      "cognitoAuthenticationType": "",
      "cognitoIdentityId": "",
      "cognitoIdentityPoolId": "",
      "sourceIp": "50.135.43.1",
      "user": "",
      "userAgent": "curl/7.43.0",
      "userArn": ""
    }
  }
}
{{< /highlight >}}

While this demonstrates that our lambda function is publicly accessible, it's not immediately obvious where the `*event` data is being populated.

### <a href="{{< relref "#example1Mapping" >}}">Mapping Templates</a>

The event data that's actually supplied to `echoS3Event` is the complete HTTP response body.  This content is what the API Gateway sends to our lambda function, which is defined by  the integration mapping.  This event data also includes the values of any whitelisted parameters.  When the API Gateway Method is defined, it optionally includes any  whitelisted query params and header values that should be forwarded to the integration target.  For this example, we're not whitelisting any params, so those fields (`queryParams`, `pathParams`) are empty.  Then for each integration target (which can be AWS Lambda, a mock, or a HTTP Proxy), it's possible to transform the API Gateway request data and whitelisted arguments into a format that's more amenable to the target.

Sparta uses a pass-through template that passes all valid data, with minor **Body** differences based on the inbound _Content-Type_:

  * [application/json](https://github.com/mweagle/Sparta/blob/master/resources/provision/apigateway/inputmapping_json.vtl)
  * [*](https://github.com/mweagle/Sparta/blob/master/resources/provision/apigateway/inputmapping_default.vtl)

The `application/json` template is copied below:

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
  },
  "context" : {
    "apiId" : "$util.escapeJavaScript($context.apiId)",
    "method" : "$util.escapeJavaScript($context.httpMethod)",
    "requestId" : "$util.escapeJavaScript($context.requestId)",
    "resourceId" : "$util.escapeJavaScript($context.resourceId)",
    "resourcePath" : "$util.escapeJavaScript($context.resourcePath)",
    "stage" : "$util.escapeJavaScript($context.stage)",
    "identity" : {
      "accountId" : "$util.escapeJavaScript($context.identity.accountId)",
      "apiKey" : "$util.escapeJavaScript($context.identity.apiKey)",
      "caller" : "$util.escapeJavaScript($context.identity.caller)",
      "cognitoAuthenticationProvider" : "$util.escapeJavaScript($context.identity.cognitoAuthenticationProvider)",
      "cognitoAuthenticationType" : "$util.escapeJavaScript($context.identity.cognitoAuthenticationType)",
      "cognitoIdentityId" : "$util.escapeJavaScript($context.identity.cognitoIdentityId)",
      "cognitoIdentityPoolId" : "$util.escapeJavaScript($context.identity.cognitoIdentityPoolId)",
      "sourceIp" : "$util.escapeJavaScript($context.identity.sourceIp)",
      "user" : "$util.escapeJavaScript($context.identity.user)",
      "userAgent" : "$util.escapeJavaScript($context.identity.userAgent)",
      "userArn" : "$util.escapeJavaScript($context.identity.userArn)"
    }
  }
}
{{< /highlight >}}

This template forwards all whitelisted data & body to the lambda function.  You can see by switching on the `method` field would permit a single function to service multiple HTTP method names.

The next example will show how to unmarshal this data and perform request-specific actions.  

### <a href="{{< relref "#example1ProxyingEnvelope" >}}">Proxying Envelope</a>

Because the integration request returned a successful response, the API Gateway response body contains only our lambda's output.  

If there were an error, the response would include additional fields (`code`, `status`, `headers`).  Those fields are injected by the NodeJS proxying tier as part of translating the **Go** HTTP response to a Lambda compatible result.  

A primary benefit of this envelope is to provide an automatic mapping from Integration Error Response Regular Expression mappings to Method Response codes.  If you look at the **Integration Response** section of the _/hello/world/test_ resource in the Console, you'll see a list of Regular Expression matches:

![API Gateway](/images/apigateway/IntegrationMapping.png)

The regular expressions are used to translate the integration response, which is just a blob of text provided to `context.done()`, into API Gateway Method responses.  Sparta annotates your lambda functions response with **Go**'s [HTTP StatusText](https://golang.org/src/net/http/status.go) values based on the HTTP status code your lambda function produced.  Sparta also provides a corresponding Method Response entry for all valid HTTP codes:

![API Gateway](/images/apigateway/MethodResponse.png)

These mappings are defaults, and it's possible to override either one by providing a non-zero length values to either:

  * [Integration.Responses](https://godoc.org/github.com/mweagle/Sparta#Integration).  See the [DefaultIntegrationResponses](https://github.com/mweagle/Sparta/blob/master/apigateway.go#L60) for the default values.
  * [Method.Responses](https://godoc.org/github.com/mweagle/Sparta#Method).  See the [DefaultMethodResponses](https://godoc.org/github.com/mweagle/Sparta#DefaultMethodResponses) for the default method response mappings.

### <a href="{{< relref "#cleanup" >}}">Cleaning Up</a>

Before moving on, remember to decommission the service via:

{{< highlight nohighlight >}}
go run application.go delete
{{< /highlight >}}

### <a href="{{< relref "#example1WrappingUp" >}}">Wrapping Up</a>

Now that we know what data is actually being sent to our API Gateway-connected Lambda function, we'll move on to performing a more complex operation, including returning a custom HTTP response body.

## Other Resources

  * [Mapping Template Reference](http://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-mapping-template-reference.html)
