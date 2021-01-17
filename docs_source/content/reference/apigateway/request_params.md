---
date: 2016-03-09T19:56:50+01:00
title: Request Parameters
weight: 11
---

# Request Parameters

This example demonstrates how to accept client request params supplied as HTTP query params and return an expiring S3 URL to access content.
The source for this is the [s3ItemInfo](https://github.com/mweagle/SpartaImager/blob/master/application.go#L149)
function defined as part of the [SpartaApplication](https://github.com/mweagle/SpartaApplication).

## Lambda Definition

Our function will accept two params:

- `bucketName` : The S3 bucket name storing the asset
- `keyName` : The S3 item key

Those params will be passed as part of the URL query string. The function will fetch the item metadata, generate an expiring URL for public S3 access, and return a JSON response body with the item data.

Because [s3ItemInfo](https://github.com/mweagle/SpartaImager/blob/master/application.go#L149) is expected to be invoked by the API Gateway, we'll use the AWS Lambda Go type in the function signature:

```go
import (
  spartaAPIGateway "github.com/mweagle/Sparta/aws/apigateway"
  spartaEvents "github.com/mweagle/Sparta/aws/events"
)

func s3ItemInfo(ctx context.Context,
  apigRequest spartaEvents.APIGatewayRequest) (*spartaAPIGateway.Response, error) {
  logger, _ := ctx.Value(sparta.ContextKeyLogger).(*zerolog.Logger)
  lambdaContext, _ := awsLambdaContext.FromContext(ctx)

  logger.Info().
    Str("RequestID", lambdaContext.AwsRequestID).
    Msg("Request received")

  getObjectInput := &s3.GetObjectInput{
    Bucket: aws.String(apigRequest.QueryParams["bucketName"]),
    Key:    aws.String(apigRequest.QueryParams["keyName"]),
  }

  awsSession := spartaAWS.NewSession(logger)
  svc := s3.New(awsSession)
  result, err := svc.GetObject(getObjectInput)
  if nil != err {
    return nil, err
  }
  presignedReq, _ := svc.GetObjectRequest(getObjectInput)
  url, err := presignedReq.Presign(5 * time.Minute)
  if nil != err {
    return nil, err
  }
  return spartaAPIGateway.NewResponse(http.StatusOK,
    &itemInfoResponse{
      S3:  result,
      URL: url,
    }), nil
}
```

The [sparta.APIGatewayRequest](https://godoc.org/github.com/mweagle/Sparta/aws/events#APIGatewayRequest) fields
correspond to the Integration Response Mapping template discussed in the [previous example](/reference/apigateway/echo_event)
(see the full mapping template [here](/reference/apigateway).

Once the event is unmarshaled, we can use it to fetch the S3 item info:

```go
getObjectInput := &s3.GetObjectInput{
  Bucket: aws.String(lambdaEvent.QueryParams["bucketName"]),
  Key:    aws.String(lambdaEvent.QueryParams["keyName"]),
}
```

Assuming there are no errors (including the case where the item does not exist), the
remainder of the function fetches the data, generates a presigned URL, and returns a JSON response whose
shape matches the Sparta default [mapping templates](https://docs.aws.amazon.com/apigateway/latest/developerguide/models-mappings.html):

```go
awsSession := spartaAWS.NewSession(logger)
svc := s3.New(awsSession)
result, err := svc.GetObject(getObjectInput)
if nil != err {
  return nil, err
}
presignedReq, _ := svc.GetObjectRequest(getObjectInput)
url, err := presignedReq.Presign(5 * time.Minute)
if nil != err {
  return nil, err
}
return spartaAPIGateway.NewResponse(http.StatusOK,
  &itemInfoResponse{
    S3:  result,
    URL: url,
  }), nil
```

## API Gateway

The next step is to create a new [API](https://godoc.org/github.com/mweagle/Sparta#API) instance via `sparta.NewAPIGateway()`

```go
apiStage := sparta.NewStage("v1")
apiGateway := sparta.NewAPIGateway("SpartaImagerAPI", apiStage)
```

## Lambda Binding

Next we create an `sparta.LambdaAWSInfo` struct that references the `s3ItemInfo` function:

```go
var iamDynamicRole = sparta.IAMRoleDefinition{}
iamDynamicRole.Privileges = append(iamDynamicRole.Privileges,
  sparta.IAMRolePrivilege{
    Actions:  []string{"s3:GetObject"},
    Resource: resourceArn,
  })
s3ItemInfoLambdaFn, _ := sparta.NewAWSLambda(sparta.LambdaName(s3ItemInfo),
  s3ItemInfo,
  iamDynamicRole)
s3ItemInfoOptions.Options = &sparta.LambdaFunctionOptions{
  Description: "Get information about an item in S3 via querystring params",
  MemorySize:  128,
  Timeout:     10,
}
```

A few items to note here:

- We're providing a custom `LambdaFunctionOptions` in case the request to S3 to get item metadata exceeds the default 3 second timeout.
- We also add a custom `iamDynamicRole.Privileges` entry to the `Privileges` slice that authorizes the lambda function to _only_ access objects in a single bucket (_resourceArn_).
  - This bucket ARN is externally created and the ARN provided to this code.
  - While the API will accept any _bucketName_ value, it is only authorized to access a single bucket.

## Resources

The next step is to associate a URL path with the `sparta.LambdaAWSInfo` struct that represents the `s3ItemInfo` function. This will be the relative path component used to reference our lambda function via the API Gateway.

```go
apiGatewayResource, _ := api.NewResource("/info", s3ItemInfoLambdaFn)
method, err := apiGatewayResource.NewMethod("GET", http.StatusOK)
if err != nil {
  return nil, err
}
```

## Whitelist Input

The final step is to add the whitelisted parameters to the Method definition.

```go
// Whitelist query string params
method.Parameters["method.request.querystring.keyName"] = true
method.Parameters["method.request.querystring.bucketName"] = true
```

Note that the keynames in the `method.Parameters` map must be of the form: **method.request.{location}.{name}** where location is one of:

- `querystring`
- `path`
- `header`

See the [REST documentation](http://docs.aws.amazon.com/apigateway/api-reference/resource/method/#requestParameters) for more information.

## Provision

With everything configured, let's provision the stack:

```nohighlight
go run application.go --level debug provision --s3Bucket $S3_BUCKET
```

and check the results.

## Verify

As this Sparta application includes an API Gateway definition, the stack `Outputs` includes the API Gateway URL:

```text
INFO[0243] Stack Outputs ▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬
INFO[0243]     APIGatewayURL                             Description="API Gateway URL" Value="https://xccmsl98p1.execute-api.us-west-2.amazonaws.com/v1"
INFO[0243] Stack provisioned                             CreationTime="2018-12-11 14:56:41.051 +0000 UTC" StackId="arn:aws:cloudformation:us-west-2:123412341234:stack/SpartaImager-mweagle/f7b7d3e0-fd54-11e8-9064-0aa3372404a6" StackName=SpartaImager-mweagle
INFO[0243] ════════════════════════════════════════════════
```

Let's fetch an item we know exists:

```nohighlight
$ curl -vs "https://xccmsl98p1.execute-api.us-west-2.amazonaws.com/v1/info?keyName=twitterAvatar.jpg&bucketName=weagle-public"

*   Trying 13.32.254.241...
* TCP_NODELAY set
* Connected to xccmsl98p1.execute-api.us-west-2.amazonaws.com (13.32.254.241) port 443 (#0)
* ALPN, offering h2
* ALPN, offering http/1.1
* Cipher selection: ALL:!EXPORT:!EXPORT40:!EXPORT56:!aNULL:!LOW:!RC4:@STRENGTH
* successfully set certificate verify locations:
*   CAfile: /etc/ssl/cert.pem
  CApath: none
* TLSv1.2 (OUT), TLS handshake, Client hello (1):
* TLSv1.2 (IN), TLS handshake, Server hello (2):
* TLSv1.2 (IN), TLS handshake, Certificate (11):
* TLSv1.2 (IN), TLS handshake, Server key exchange (12):
* TLSv1.2 (IN), TLS handshake, Server finished (14):
* TLSv1.2 (OUT), TLS handshake, Client key exchange (16):
* TLSv1.2 (OUT), TLS change cipher, Client hello (1):
* TLSv1.2 (OUT), TLS handshake, Finished (20):
* TLSv1.2 (IN), TLS change cipher, Client hello (1):
* TLSv1.2 (IN), TLS handshake, Finished (20):
* SSL connection using TLSv1.2 / ECDHE-RSA-AES128-GCM-SHA256
* ALPN, server accepted to use h2
* Server certificate:
*  subject: CN=*.execute-api.us-west-2.amazonaws.com
*  start date: Oct  9 00:00:00 2018 GMT
*  expire date: Oct  9 12:00:00 2019 GMT
*  subjectAltName: host "xccmsl98p1.execute-api.us-west-2.amazonaws.com" matched cert's "*.execute-api.us-west-2.amazonaws.com"
*  issuer: C=US; O=Amazon; OU=Server CA 1B; CN=Amazon
*  SSL certificate verify ok.
* Using HTTP2, server supports multi-use
* Connection state changed (HTTP/2 confirmed)
* Copying HTTP/2 data in stream buffer to connection buffer after upgrade: len=0
* Using Stream ID: 1 (easy handle 0x7ff68b802c00)
> GET /v1/info?keyName=twitterAvatar.jpg&bucketName=weagle-public HTTP/2
> Host: xccmsl98p1.execute-api.us-west-2.amazonaws.com
> User-Agent: curl/7.54.0
> Accept: */*
>
* Connection state changed (MAX_CONCURRENT_STREAMS updated)!
< HTTP/2 200
< content-type: application/json
< content-length: 1539
< date: Tue, 11 Dec 2018 15:08:56 GMT
< x-amzn-requestid: aded8786-fd56-11e8-836c-dff86eb3938d
< access-control-allow-origin: *
< access-control-allow-headers: Content-Type,X-Amz-Date,Authorization,X-Api-Key
< x-amz-apigw-id: Rv3pRH8jPHcFTfA=
< access-control-allow-methods: *
< x-amzn-trace-id: Root=1-5c0fd308-f576dae00848eb44535a5c70;Sampled=0
< x-cache: Miss from cloudfront
< via: 1.1 8ddadd1ab84a7f1bef108d6a72eccf06.cloudfront.net (CloudFront)
< x-amz-cf-id: OO01Dua9x5dHyXr-arKJ3LKu2ahbPYv5ESqUg2lAhlzLJDQTLVyW_A==
<
{"S3":{"AcceptRanges":"bytes","Body":{},"CacheControl":null,"ContentDisposition":null,"ContentEncoding":null,"ContentLanguage":null,"ContentLength":613560,"ContentRange":null,"ContentType":"image/jpeg","DeleteMarker":null,"ETag":"\"7250a1802a5e2f94532b9ee38429a3fd\"","Expiration":null,"Expires":null,"LastModified":"2018-03-14T14:55:19Z","Metadata":{},"MissingMeta":null,"ObjectLockLegalHoldStatus":null,"ObjectLockMode":null,"ObjectLockRetainUntilDate":null,"PartsCount":null,"ReplicationStatus":null,"RequestCharged":null,"Restore":null,"SSECustomerAlgorithm":null,"SSECustomerKeyMD5":null,"SSEKMSKeyId":null,"ServerSideEncryption":null,"StorageClass":null,"TagCount":null,"VersionId":null,"WebsiteRedirectLocation":null},"URL":"https://weagle-public.s3.us-west-2.amazonaws.com/twitterAvatar.jpg?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=ASIAQMUWTUUFF65WLRLE%2F20181211%2Fus-west-2%2Fs3%2Faws4_request&X-Amz-Date=20181211T150856Z&X-Amz-Expires=300&X-Amz-Security-Token=FQoGZXIvYXdzEIH%2F%2F%2F%2F%2F%2F%2F%2F%2F%2FwEaDMMVITmbkwrrxznAHCL9AaUQwfC%2F%2F6go%2FKBZigDuI4BLLwJzqiwhquTZ9TR1oxVKOAA0h6WzWUEfjjOjZK56SFk3cIJ%2FjKIBmImKpTIGyN7fn48s6N51RFFxra2Mamrp1pDqEcP4VswnJH8C5Q7ZfmltJDiFqLbd4FCQdgoGT228Ls49Uo24EyT%2B%2BTL%2Fl0sKTVYtI1MbGSK%2B%2BKZ6rpPEsyR%2FTuIdeDvA1P%2BRlMEyvr0NhO7Wpf7ZZMs3taNcUMQDRmARyIgAp87ziwIavUTaPqbgpGNqJ6XAO%2Byf3y0g9JurYj44HrwpLWmuF5g%2B%2FtLv8VikzqD8GuWARJuo%2BPlH54KmcMrbXBpLq9sZl2Io3KO%2F4AU%3D&X-Amz-SignedHeaders=host&X-Amz-Signature=88976d33d4cdefff02265e1f40e4d18005231672f1a6e41ad12733f0ce97e91b"}
```

Pretty printing the response body:

```json
{
  "S3": {
    "AcceptRanges": "bytes",
    "Body": {},
    "CacheControl": null,
    "ContentDisposition": null,
    "ContentEncoding": null,
    "ContentLanguage": null,
    "ContentLength": 613560,
    "ContentRange": null,
    "ContentType": "image/jpeg",
    "DeleteMarker": null,
    "ETag": "\"7250a1802a5e2f94532b9ee38429a3fd\"",
    "Expiration": null,
    "Expires": null,
    "LastModified": "2018-03-14T14:55:19Z",
    "Metadata": {},
    "MissingMeta": null,
    "ObjectLockLegalHoldStatus": null,
    "ObjectLockMode": null,
    "ObjectLockRetainUntilDate": null,
    "PartsCount": null,
    "ReplicationStatus": null,
    "RequestCharged": null,
    "Restore": null,
    "SSECustomerAlgorithm": null,
    "SSECustomerKeyMD5": null,
    "SSEKMSKeyId": null,
    "ServerSideEncryption": null,
    "StorageClass": null,
    "TagCount": null,
    "VersionId": null,
    "WebsiteRedirectLocation": null
  },
  "URL": "https://weagle-public.s3.us-west-2.amazonaws.com/twitterAvatar.jpg?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=ASIAQMUWTUUFF65WLRLE%2F20181211%2Fus-west-2%2Fs3%2Faws4_request&X-Amz-Date=20181211T150856Z&X-Amz-Expires=300&X-Amz-Security-Token=FQoGZXIvYXdzEIH%2F%2F%2F%2F%2F%2F%2F%2F%2F%2FwEaDMMVITmbkwrrxznAHCL9AaUQwfC%2F%2F6go%2FKBZigDuI4BLLwJzqiwhquTZ9TR1oxVKOAA0h6WzWUEfjjOjZK56SFk3cIJ%2FjKIBmImKpTIGyN7fn48s6N51RFFxra2Mamrp1pDqEcP4VswnJH8C5Q7ZfmltJDiFqLbd4FCQdgoGT228Ls49Uo24EyT%2B%2BTL%2Fl0sKTVYtI1MbGSK%2B%2BKZ6rpPEsyR%2FTuIdeDvA1P%2BRlMEyvr0NhO7Wpf7ZZMs3taNcUMQDRmARyIgAp87ziwIavUTaPqbgpGNqJ6XAO%2Byf3y0g9JurYj44HrwpLWmuF5g%2B%2FtLv8VikzqD8GuWARJuo%2BPlH54KmcMrbXBpLq9sZl2Io3KO%2F4AU%3D&X-Amz-SignedHeaders=host&X-Amz-Signature=88976d33d4cdefff02265e1f40e4d18005231672f1a6e41ad12733f0ce97e91b"
}
```

What about an item that we know doesn't exist, but is in the bucket our lambda function has privileges to access:

```text
$ curl -vs "https://xccmsl98p1.execute-api.us-west-2.amazonaws.com/v1/info?keyName=NOT_HERE.jpg&bucketName=weagle-public"

*   Trying 13.32.254.241...
* TCP_NODELAY set
* Connected to xccmsl98p1.execute-api.us-west-2.amazonaws.com (13.32.254.241) port 443 (#0)
* ALPN, offering h2
* ALPN, offering http/1.1
* Cipher selection: ALL:!EXPORT:!EXPORT40:!EXPORT56:!aNULL:!LOW:!RC4:@STRENGTH
* successfully set certificate verify locations:
*   CAfile: /etc/ssl/cert.pem
  CApath: none
* TLSv1.2 (OUT), TLS handshake, Client hello (1):
* TLSv1.2 (IN), TLS handshake, Server hello (2):
* TLSv1.2 (IN), TLS handshake, Certificate (11):
* TLSv1.2 (IN), TLS handshake, Server key exchange (12):
* TLSv1.2 (IN), TLS handshake, Server finished (14):
* TLSv1.2 (OUT), TLS handshake, Client key exchange (16):
* TLSv1.2 (OUT), TLS change cipher, Client hello (1):
* TLSv1.2 (OUT), TLS handshake, Finished (20):
* TLSv1.2 (IN), TLS change cipher, Client hello (1):
* TLSv1.2 (IN), TLS handshake, Finished (20):
* SSL connection using TLSv1.2 / ECDHE-RSA-AES128-GCM-SHA256
* ALPN, server accepted to use h2
* Server certificate:
*  subject: CN=*.execute-api.us-west-2.amazonaws.com
*  start date: Oct  9 00:00:00 2018 GMT
*  expire date: Oct  9 12:00:00 2019 GMT
*  subjectAltName: host "xccmsl98p1.execute-api.us-west-2.amazonaws.com" matched cert's "*.execute-api.us-west-2.amazonaws.com"
*  issuer: C=US; O=Amazon; OU=Server CA 1B; CN=Amazon
*  SSL certificate verify ok.
* Using HTTP2, server supports multi-use
* Connection state changed (HTTP/2 confirmed)
* Copying HTTP/2 data in stream buffer to connection buffer after upgrade: len=0
* Using Stream ID: 1 (easy handle 0x7f9e4f00b600)
> GET /v1/info?keyName=twitterAvatarArgh.jpg&bucketName=weagle HTTP/2
> Host: xccmsl98p1.execute-api.us-west-2.amazonaws.com
> User-Agent: curl/7.54.0
> Accept: */*
>
* Connection state changed (MAX_CONCURRENT_STREAMS updated)!
< HTTP/2 404
< content-type: application/json
< content-length: 177
< date: Tue, 11 Dec 2018 15:21:18 GMT
< x-amzn-requestid: 675edef9-fd58-11e8-ae45-3fac75041f4d
< access-control-allow-origin: *
< access-control-allow-headers: Content-Type,X-Amz-Date,Authorization,X-Api-Key
< x-amz-apigw-id: Rv5dAETkvHcFvYg=
< access-control-allow-methods: *
< x-amzn-trace-id: Root=1-5c0fd5ec-1d8bba64519f71126c12b4d6;Sampled=0
< x-cache: Error from cloudfront
< via: 1.1 4c4ed81695980f3c6829b9fd229bd0f8.cloudfront.net (CloudFront)
< x-amz-cf-id: ZT5R4BUSAkZpT46s_wCjBImHsM3w6mHFlYG0lnfwONSkPCgxzOQ_lQ==
<
{"error":"AccessDenied: Access Denied\n\tstatus code: 403, request id: A10C69E17E4C9D00, host id: pAnhP+tg9rDh0yP5FJyC8bSnj1GJJjJvAFXwiluW4yHnVvt5EvkvkpKA4UzjJmCoFyI8hGST6YE="}
* Connection #0 to host xccmsl98p1.execute-api.us-west-2.amazonaws.com left intact
```

And finally, what if we try to access a bucket that our lambda function isn't authorized to access:

```text
$ curl -vs "https://xccmsl98p1.execute-api.us-west-2.amazonaws.com/v1/info?keyName=NOT_HERE.jpg&bucketName=VERY_PRIVATE_BUCKET"

*   Trying 13.32.254.241...
* TCP_NODELAY set
* Connected to xccmsl98p1.execute-api.us-west-2.amazonaws.com (13.32.254.241) port 443 (#0)
* ALPN, offering h2
* ALPN, offering http/1.1
* Cipher selection: ALL:!EXPORT:!EXPORT40:!EXPORT56:!aNULL:!LOW:!RC4:@STRENGTH
* successfully set certificate verify locations:
*   CAfile: /etc/ssl/cert.pem
  CApath: none
* TLSv1.2 (OUT), TLS handshake, Client hello (1):
* TLSv1.2 (IN), TLS handshake, Server hello (2):
* TLSv1.2 (IN), TLS handshake, Certificate (11):
* TLSv1.2 (IN), TLS handshake, Server key exchange (12):
* TLSv1.2 (IN), TLS handshake, Server finished (14):
* TLSv1.2 (OUT), TLS handshake, Client key exchange (16):
* TLSv1.2 (OUT), TLS change cipher, Client hello (1):
* TLSv1.2 (OUT), TLS handshake, Finished (20):
* TLSv1.2 (IN), TLS change cipher, Client hello (1):
* TLSv1.2 (IN), TLS handshake, Finished (20):
* SSL connection using TLSv1.2 / ECDHE-RSA-AES128-GCM-SHA256
* ALPN, server accepted to use h2
* Server certificate:
*  subject: CN=*.execute-api.us-west-2.amazonaws.com
*  start date: Oct  9 00:00:00 2018 GMT
*  expire date: Oct  9 12:00:00 2019 GMT
*  subjectAltName: host "xccmsl98p1.execute-api.us-west-2.amazonaws.com" matched cert's "*.execute-api.us-west-2.amazonaws.com"
*  issuer: C=US; O=Amazon; OU=Server CA 1B; CN=Amazon
*  SSL certificate verify ok.
* Using HTTP2, server supports multi-use
* Connection state changed (HTTP/2 confirmed)
* Copying HTTP/2 data in stream buffer to connection buffer after upgrade: len=0
* Using Stream ID: 1 (easy handle 0x7f9e4f00b600)
> GET /v1/info?keyName=twitterAvatarArgh.jpg&bucketName=weagle HTTP/2
> Host: xccmsl98p1.execute-api.us-west-2.amazonaws.com
> User-Agent: curl/7.54.0
> Accept: */*
>
* Connection state changed (MAX_CONCURRENT_STREAMS updated)!
< HTTP/2 404
< content-type: application/json
< content-length: 177
< date: Tue, 11 Dec 2018 15:21:18 GMT
< x-amzn-requestid: 675edef9-fd58-11e8-ae45-3fac75041f4d
< access-control-allow-origin: *
< access-control-allow-headers: Content-Type,X-Amz-Date,Authorization,X-Api-Key
< x-amz-apigw-id: Rv5dAETkvHcFvYg=
< access-control-allow-methods: *
< x-amzn-trace-id: Root=1-5c0fd5ec-1d8bba64519f71126c12b4d6;Sampled=0
< x-cache: Error from cloudfront
< via: 1.1 4c4ed81695980f3c6829b9fd229bd0f8.cloudfront.net (CloudFront)
< x-amz-cf-id: ZT5R4BUSAkZpT46s_wCjBImHsM3w6mHFlYG0lnfwONSkPCgxzOQ_lQ==
<
{"error":"AccessDenied: Access Denied\n\tstatus code: 403, request id: A10C69E17E4C9D00, host id: pAnhP+tg9rDh0yP5FJyC8bSnj1GJJjJvAFXwiluW4yHnVvt5EvkvkpKA4UzjJmCoFyI8hGST6YE="}
* Connection #0 to host xccmsl98p1.execute-api.us-west-2.amazonaws.com left intact
```

## Cleanup

Before moving on, remember to decommission the service via:

```nohighlight
go run application.go delete
```

## Conclusion

With this example we've walked through a simple example that whitelists user input, uses IAM Roles to
limit what S3 buckets a lambda function may access, and returns an _application/json_ response to the caller.
