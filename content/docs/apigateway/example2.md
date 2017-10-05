---
date: 2016-03-09T19:56:50+01:00
title: API Gateway - User Input
weight: 10
---

This example demonstrates how to accept user input (delivered as HTTP query params) and return an expiring S3 URL to fetch content.  The source for this is the [s3ItemInfo](https://github.com/mweagle/SpartaImager/blob/master/application.go#L149) function defined as part of the  [SpartaApplication](https://github.com/mweagle/SpartaApplication).


# Define the Lambda Function

Our function will accept two params:

  * `bucketName` : The S3 bucket name storing the asset
  * `keyName` : The S3 item key

Those params will be passed as part of the URL query string.  The function will fetch the item metadata, generate an expiring URL for public S3 access, and return a JSON response body with the item data.

Because [s3ItemInfo](https://github.com/mweagle/SpartaImager/blob/master/application.go#L149) is expected to be invoked by the API Gateway, we'll start by unmarshalling the event data:

{{< highlight go >}}
decoder := json.NewDecoder(r.Body)
defer r.Body.Close()
var lambdaEvent sparta.APIGatewayLambdaJSONEvent
err := decoder.Decode(&lambdaEvent)
if err != nil {
  logger.Error("Failed to unmarshal event data: ", err.Error())
  http.Error(w, err.Error(), http.StatusInternalServerError)
  return
}
{{< /highlight >}}

The [sparta.APIGatewayLambdaJSONEvent](https://godoc.org/github.com/mweagle/Sparta#APIGatewayLambdaJSONEvent) fields correspond to the Integration Response Mapping template discussed in the [previous example](/docs/apigateway/example1) (see the full mapping template [here](https://raw.githubusercontent.com/mweagle/Sparta/master/resources/gateway/inputmapping_json.vtl)).

Once the event is unmarshaled, we can use it to fetch the S3 item info:

{{< highlight go >}}
getObjectInput := &s3.GetObjectInput{
  Bucket: aws.String(lambdaEvent.QueryParams["bucketName"]),
  Key:    aws.String(lambdaEvent.QueryParams["keyName"]),
}
{{< /highlight >}}

Assuming there are no errors (including the case where the item does not exist), the remainder of the function fetches the data, generates a presigned URL, and returns a JSON response:


{{< highlight go >}}
awsSession := awsSession(logger)
svc := s3.New(awsSession)
result, err := svc.GetObject(getObjectInput)
if nil != err {
  logger.Error("Failed to process event: ", err.Error())
  http.Error(w, err.Error(), http.StatusInternalServerError)
  return
}
presignedReq, _ := svc.GetObjectRequest(getObjectInput)
url, err := presignedReq.Presign(5 * time.Minute)
if nil != err {
  logger.Error("Failed to process event: ", err.Error())
  http.Error(w, err.Error(), http.StatusInternalServerError)
  return
}
httpResponse := map[string]interface{}{
  "S3":  result,
  "URL": url,
}

responseBody, err := json.Marshal(httpResponse)
if err != nil {
  http.Error(w, err.Error(), http.StatusInternalServerError)
} else {
  w.Header().Set("Content-Type", "application/json")
  fmt.Fprint(w, string(responseBody))
}
{{< /highlight >}}

# Create the API Gateway

The next step is to create a new [API](https://godoc.org/github.com/mweagle/Sparta#API) instance via `sparta.NewAPIGateway()`

{{< highlight go >}}
apiStage := sparta.NewStage("v1")
apiGateway := sparta.NewAPIGateway("SpartaImagerAPI", apiStage)
{{< /highlight >}}

# Create Lambda Binding

Next we create an `sparta.LambdaAWSInfo` struct that references the `s3ItemInfo` function:

{{< highlight go >}}
var iamDynamicRole = sparta.IAMRoleDefinition{}
iamDynamicRole.Privileges = append(iamDynamicRole.Privileges,
  sparta.IAMRolePrivilege{
    Actions:  []string{"s3:GetObject"},
    Resource: resourceArn,
  })
s3ItemInfoLambdaFn := sparta.HandleAWSLambda(sparta.LambdaName(s3ItemInfo),
  http.HandlerFunc(s3ItemInfo),
  iamDynamicRole)
s3ItemInfoOptions.Options = &sparta.LambdaFunctionOptions{
  Description: "Get information about an item in S3 via querystring params",
  MemorySize:  128,
  Timeout:     10,
}
{{< /highlight >}}

A few items to note here:

  * We're providing a custom `LambdaFunctionOptions` in case the request to S3 to get item metadata exceeds the default 3 second timeout.
  * We also add a custom `iamDynamicRole.Privileges` entry to the `Privileges` slice that authorizes the lambda function to _only_ access objects in a single bucket (_resourceArn_).
    * This bucket ARN is externally created and the ARN provided to this code.
    * While the API will accept any _bucketName_ value, it is only authorized to access a single bucket.

# Create Resources

The next step is to associate a URL path with the `sparta.LambdaAWSInfo` struct that represents the `s3ItemInfo` function. This will be the relative path component used to reference our lambda function via the API Gateway.

{{< highlight go >}}
apiGatewayResource, _ := api.NewResource("/info", s3ItemInfoLambdaFn)
method, err := apiGatewayResource.NewMethod("GET", http.StatusOK)
if err != nil {
  return nil, err
}
{{< /highlight >}}

# Whitelist Input

The final step is to add the whitelisted parameters to the Method definition.

{{< highlight go >}}
// Whitelist query string params
method.Parameters["method.request.querystring.keyName"] = true
method.Parameters["method.request.querystring.bucketName"] = true
{{< /highlight >}}

Note that the keynames in the `method.Parameters` map must be of the form: **method.request.{location}.{name}** where location is one of:

  * `querystring`
  * `path`
  * `header`

See the [REST documentation](http://docs.aws.amazon.com/apigateway/api-reference/resource/method/#requestParameters) for more information.

# Provision

With everything configured, let's provision the stack:

{{< highlight nohighlight >}}
go run application.go --level debug provision --s3Bucket $S3_BUCKET
{{< /highlight >}}

and check the results.

# Querying

As this Sparta application includes an API Gateway definition, the stack `Outputs` includes the API Gateway URL:

{{< highlight nohighlight >}}
INFO[0113] Stack output   Description=API Gateway URL Key=APIGatewayURL Value=https://0ux556ho77.execute-api.us-west-2.amazonaws.com/v1
INFO[0113] Stack output   Description=Sparta Home Key=SpartaHome Value=https://github.com/mweagle/Sparta
INFO[0113] Stack output   Description=Sparta Version Key=SpartaVersion Value=0.1.0
{{< /highlight >}}

Let's fetch an item we know exists:

{{< highlight nohighlight >}}
curl -vs "https://0ux556ho77.execute-api.us-west-2.amazonaws.com/v1/info?keyName=gopher.png&bucketName=somebucket-log"

*   Trying 54.192.70.158...
* Connected to 0ux556ho77.execute-api.us-west-2.amazonaws.com (54.192.70.158) port 443 (#0)
* TLS 1.2 connection using TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
* Server certificate: *.execute-api.us-west-2.amazonaws.com
* Server certificate: Symantec Class 3 Secure Server CA - G4
* Server certificate: VeriSign Class 3 Public Primary Certification Authority - G5
> GET /v1/info?keyName=gopher.png&bucketName=somebucket-log HTTP/1.1
> Host: 0ux556ho77.execute-api.us-west-2.amazonaws.com
> User-Agent: curl/7.43.0
> Accept: */*
>
< HTTP/1.1 200 OK
< Content-Type: application/json
< Content-Length: 1584
< Connection: keep-alive
< Date: Sun, 06 Dec 2015 02:35:03 GMT
< x-amzn-RequestId: f333f4bb-9bc1-11e5-afde-61a428c89049
< X-Cache: Miss from cloudfront
< Via: 1.1 2f31d4850470c56c3b326946dc542a6b.cloudfront.net (CloudFront)
< X-Amz-Cf-Id: 6rBpqjmi7DPax7XOHTbxDx8-FfFfvI04m2_K-PxLWfYFor7WtIcdxA==
<
* Connection #0 to host 0ux556ho77.execute-api.us-west-2.amazonaws.com left intact
{"code":200,"status":"OK","headers":{"content-type":"application/json","date":"Sun, 06 Dec 2015 02:35:03 GMT","content-length":"1468"},"results":{"S3":{"AcceptRanges":"bytes","Body":{},"CacheControl":null,"ContentDisposition":null,"ContentEncoding":null,"ContentLanguage":null,"ContentLength":70372,"ContentRange":null,"ContentType":"image/png","DeleteMarker":null,"ETag":"\"ca1f746d6f232f87fca4e4d94ef6f3ab\"","Expiration":null,"Expires":null,"LastModified":"2015-11-09T15:38:01Z","Metadata":{},"MissingMeta":null,"ReplicationStatus":null,"RequestCharged":null,"Restore":null,"SSECustomerAlgorithm":null,"SSECustomerKeyMD5":null,"SSEKMSKeyId":null,"ServerSideEncryption":null,"StorageClass":null,"VersionId":null,"WebsiteRedirectLocation":null},"URL":"https://somebucket-log.s3-us-west-2.amazonaws.com/gopher.png?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=ASIAJ5KB2P6SQ4E7IMMQ%2F20151206%2Fus-west-2%2Fs3%2Faws4_request&X-Amz-Date=20151206T023503Z&X-Amz-Expires=300&X-Amz-Security-Token=AQoDYXdzEFQawAK7vrGb%2BH9lw%2FhEHpR9Yg1KwPmmFcvyMzF7ewFBmxpOkfEM7gLZirMcFFexcxpWv%2F5CVAxpqjRf5FznOYJZHHoBqgmUcKPQZOpYKSbQG768zH5gMNdOANWin1COZU8DyuABrkJYL1bdFpwV7oHgrDmRz2G6oZqqOnfesRHW8WcehSXMV%2BcQFaAcO7IaIMAkRINMIDfxQaa%2FP8i8dbrcOfsEy6UABeaLKL3YgdZIouxcUUKzXQ6Pr4Cgrf0TAyRDAO1t6bVXzv6UFa6j00%2Fm0PYElni7xs5844UFAav%2B1weO2kX65ETzwUxBacAAnuzt%2BmTVPWeikhzgRnjBFn8mQjkZLCJklJJb6QHBO8dph2CSQsh47yw7%2BnexGjAu1y106AA2%2Bfa0WFYC552Q%2FrVVhKU7dejy%2B3jz%2F4LyWdnva9IvmCDVvY6zBQ%3D%3D&X-Amz-SignedHeaders=host&X-Amz-Signature=7d0e6663e043317b5611ddf4ae9f7514aff8c484a31deba524906ba50cbc6a2f"}}
{{< /highlight >}}

Pretty printing the response body:

{{< highlight json >}}
{
  "code": 200,
  "status": "OK",
  "headers": {
    "content-type": "application/json",
    "date": "Sun, 06 Dec 2015 02:35:03 GMT",
    "content-length": "1468"
  },
  "results": {
    "S3": {
      "AcceptRanges": "bytes",
      "Body": {},
      "CacheControl": null,
      "ContentDisposition": null,
      "ContentEncoding": null,
      "ContentLanguage": null,
      "ContentLength": 70372,
      "ContentRange": null,
      "ContentType": "image/png",
      "DeleteMarker": null,
      "ETag": "\"ca1f746d6f232f87fca4e4d94ef6f3ab\"",
      "Expiration": null,
      "Expires": null,
      "LastModified": "2015-11-09T15:38:01Z",
      "Metadata": {},
      "MissingMeta": null,
      "ReplicationStatus": null,
      "RequestCharged": null,
      "Restore": null,
      "SSECustomerAlgorithm": null,
      "SSECustomerKeyMD5": null,
      "SSEKMSKeyId": null,
      "ServerSideEncryption": null,
      "StorageClass": null,
      "VersionId": null,
      "WebsiteRedirectLocation": null
    },
    "URL": "https://somebucket-log.s3-us-west-2.amazonaws.com/gopher.png?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=ASIAJ5KB2P6SQ4E7IMMQ%2F20151206%2Fus-west-2%2Fs3%2Faws4_request&X-Amz-Date=20151206T023503Z&X-Amz-Expires=300&X-Amz-Security-Token=AQoDYXdzEFQawAK7vrGb%2BH9lw%2FhEHpR9Yg1KwPmmFcvyMzF7ewFBmxpOkfEM7gLZirMcFFexcxpWv%2F5CVAxpqjRf5FznOYJZHHoBqgmUcKPQZOpYKSbQG768zH5gMNdOANWin1COZU8DyuABrkJYL1bdFpwV7oHgrDmRz2G6oZqqOnfesRHW8WcehSXMV%2BcQFaAcO7IaIMAkRINMIDfxQaa%2FP8i8dbrcOfsEy6UABeaLKL3YgdZIouxcUUKzXQ6Pr4Cgrf0TAyRDAO1t6bVXzv6UFa6j00%2Fm0PYElni7xs5844UFAav%2B1weO2kX65ETzwUxBacAAnuzt%2BmTVPWeikhzgRnjBFn8mQjkZLCJklJJb6QHBO8dph2CSQsh47yw7%2BnexGjAu1y106AA2%2Bfa0WFYC552Q%2FrVVhKU7dejy%2B3jz%2F4LyWdnva9IvmCDVvY6zBQ%3D%3D&X-Amz-SignedHeaders=host&X-Amz-Signature=7d0e6663e043317b5611ddf4ae9f7514aff8c484a31deba524906ba50cbc6a2f"
  }
}
{{< /highlight >}}

Please see the [first example](/docs/apigateway/example1) for more information on the `code`, `status`, and `headers` keys.

What about an item that we know doesn't exist, but is in the bucket our lambda function has privileges to access:

{{< highlight nohighlight >}}

curl -vs "https://0ux556ho77.execute-api.us-west-2.amazonaws.com/v1/info?keyName=gopher42.png&bucketName=somebucket-log"

*   Trying 54.230.71.213...
* Connected to 0ux556ho77.execute-api.us-west-2.amazonaws.com (54.230.71.213) port 443 (#0)
* TLS 1.2 connection using TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
* Server certificate: *.execute-api.us-west-2.amazonaws.com
* Server certificate: Symantec Class 3 Secure Server CA - G4
* Server certificate: VeriSign Class 3 Public Primary Certification Authority - G5
> GET /v1/info?keyName=gopher42.png&bucketName=somebucket-log HTTP/1.1
> Host: 0ux556ho77.execute-api.us-west-2.amazonaws.com
> User-Agent: curl/7.43.0
> Accept: */*
>
< HTTP/1.1 500 Internal Server Error
< Content-Type: application/json
< Content-Length: 524
< Connection: keep-alive
< Date: Sun, 06 Dec 2015 02:40:14 GMT
< x-amzn-RequestId: ad5d94eb-9bc2-11e5-8fad-476a6cacabce
< X-Cache: Error from cloudfront
< Via: 1.1 29bfa9b96f4ea66dc02526ee845ca6b0.cloudfront.net (CloudFront)
< X-Amz-Cf-Id: XoVLBjm1dgozZsNAEGk8Vy_a5PXMYNWRD6eKJJBcVTXrtMgMhiLNyQ==
<
* Connection #0 to host 0ux556ho77.execute-api.us-west-2.amazonaws.com left intact
{"errorMessage":"{\"code\":500,\"status\":\"Internal Server Error\",\"headers\":{\"content-type\":\"text/plain; charset=utf-8\",\"x-content-type-options\":\"nosniff\",\"date\":\"Sun, 06 Dec 2015 02:40:14 GMT\",\"content-length\":\"60\"},\"error\":\"AccessDenied: Access Denied\\n\\tstatus code: 403, request id: \\n\"}","errorType":"Error","stackTrace":["IncomingMessage.<anonymous> (/var/task/index.js:68:53)","IncomingMessage.emit (events.js:117:20)","_stream_readable.js:944:16","process._tickCallback (node.js:442:13)"]}
{{< /highlight >}}

And finally, what if we try to access a bucket that our lambda function isn't authorized to access:

{{< highlight nohighlight >}}

curl -vs "https://0ux556ho77.execute-api.us-west-2.amazonaws.com/v1/info?keyName=gopher.png&bucketName=weagle"

*   Trying 54.192.70.129...
* Connected to 0ux556ho77.execute-api.us-west-2.amazonaws.com (54.192.70.129) port 443 (#0)
* TLS 1.2 connection using TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
* Server certificate: *.execute-api.us-west-2.amazonaws.com
* Server certificate: Symantec Class 3 Secure Server CA - G4
* Server certificate: VeriSign Class 3 Public Primary Certification Authority - G5
> GET /v1/info?keyName=gopher.png&bucketName=weagle HTTP/1.1
> Host: 0ux556ho77.execute-api.us-west-2.amazonaws.com
> User-Agent: curl/7.43.0
> Accept: */*
>
< HTTP/1.1 500 Internal Server Error
< Content-Type: application/json
< Content-Length: 524
< Connection: keep-alive
< Date: Sun, 06 Dec 2015 02:42:52 GMT
< x-amzn-RequestId: 0be0fc4f-9bc3-11e5-b827-81d99c02192f
< X-Cache: Error from cloudfront
< Via: 1.1 400bdbea4e851ce61e7df8252da93d3f.cloudfront.net (CloudFront)
< X-Amz-Cf-Id: M_7pB1UsW63xzh_9g37-CqNYDXfXlec0B6DV4bdkq3tbCANCOrTY6Q==
<
* Connection #0 to host 0ux556ho77.execute-api.us-west-2.amazonaws.com left intact
{"errorMessage":"{\"code\":500,\"status\":\"Internal Server Error\",\"headers\":{\"content-type\":\"text/plain; charset=utf-8\",\"x-content-type-options\":\"nosniff\",\"date\":\"Sun, 06 Dec 2015 02:42:52 GMT\",\"content-length\":\"60\"},\"error\":\"AccessDenied: Access Denied\\n\\tstatus code: 403, request id: \\n\"}","errorType":"Error","stackTrace":["IncomingMessage.<anonymous> (/var/task/index.js:68:53)","IncomingMessage.emit (events.js:117:20)","_stream_readable.js:944:16","process._tickCallback (node.js:442:13)"]}
{{< /highlight >}}

# Cleaning Up

Before moving on, remember to decommission the service via:

{{< highlight nohighlight >}}
go run application.go delete
{{< /highlight >}}

# Wrapping Up

With this example we've walked through a simple example that whitelists user input, uses IAM Roles to limit what S3 buckets a lambda function may access, and returns JSON data to the caller.
