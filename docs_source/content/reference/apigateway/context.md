---
date: 2016-03-09T19:56:50+01:00
title: Request Context
weight: 12
---

This example demonstrates how to use the `Context` struct provided as part of the [APIGatewayRequest](https://godoc.org/github.com/mweagle/Sparta/aws/events#APIGatewayRequest).  The [SpartaGeoIP](https://github.com/mweagle/SpartaGeoIP) service will return Geo information based on the inbound request's IP address.

## Lambda Definition

Our function will examine the inbound request, lookup the user's IP address in the [GeoLite2 Database](http://dev.maxmind.com/geoip/geoip2/geolite2/) and return any information to the client.

As this function is only expected to be invoked from the API Gateway, we'll unmarshall the inbound event:

```go
import (
  spartaAWSEvents "github.com/mweagle/Sparta/aws/events"
  spartaAPIGateway "github.com/mweagle/Sparta/aws/apigateway"
)
func ipGeoLambda(ctx context.Context,
  apiRequest spartaAWSEvents.APIGatewayRequest) (*spartaAPIGateway.Response, error) {
  parsedIP := net.ParseIP(apiRequest.Context.Identity.SourceIP)
  record, err := dbHandle.City(parsedIP)
  if err != nil {
    return nil, err
  }

```

We'll then parse the inbound IP address from the [Context](https://godoc.org/github.com/mweagle/Sparta#APIGatewayContext) and perform a lookup against the database handle opened in the [init](https://github.com/mweagle/SpartaGeoIP/blob/master/main.go#L19) block:

```go
  parsedIP := net.ParseIP(lambdaEvent.Context.Identity.SourceIP)
  record, err := dbHandle.City(parsedIP)
  if err != nil {
    return nil, err
  }
```

Finally, marshal the data or error result and we're done:

```go
  requestResponse := map[string]interface{}{
    "ip":     parsedIP,
    "record": record,
  }
  return spartaAPIGateway.NewResponse(http.StatusOK, requestResponse), nil
```

## Sparta Integration

The next steps are to:

  1. Create the [LambdaAWSInfo](https://godoc.org/github.com/mweagle/Sparta#LambdaAWSInfo) value
  1. Create an associated API Gateway
  1. Create an API Gateway resource that invokes our lambda function
  1. Add a Method name to the resource.

These four steps are managed in the service's `main()` function:

```go
////////////////////////////////////////////////////////////////////////////////
// Main
func main() {
  stage := sparta.NewStage("ipgeo")
  apiGateway := sparta.NewAPIGateway("SpartaGeoIPService", stage)
  stackName := "SpartaGeoIP"

  var lambdaFunctions []*sparta.LambdaAWSInfo
  lambdaFn, _ := sparta.NewAWSLambda(sparta.LambdaName(ipGeoLambda),
    ipGeoLambda,
    sparta.IAMRoleDefinition{})
  apiGatewayResource, _ := apiGateway.NewResource("/info", lambdaFn)
  apiMethod, _ := apiGatewayResource.NewMethod("GET", http.StatusOK, http.StatusOK)
  apiMethod.SupportedRequestContentTypes = []string{"application/json"}

  lambdaFunctions = append(lambdaFunctions, lambdaFn)

  sparta.Main(stackName,
    "Sparta app supporting ip->geo mapping",
    lambdaFunctions,
    apiGateway,
    nil)
}
```

## Provision

The next step is to provision the stack:

```nohighlight
S3_BUCKET=<MY-S3-BUCKETNAME> mage provision
```

Assuming all goes well, the log output will include the API Gateway URL as in:

```text
INFO[0077] Stack Outputs ▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬
INFO[0077]     APIGatewayURL                             Description="API Gateway URL" Value="https://52a5qqqwo4.execute-api.us-west-2.amazonaws.com/ipgeo"
INFO[0077] Stack provisioned                             CreationTime="2018-12-11 14:30:01.822 +0000 UTC" StackId="arn:aws:cloudformation:us-west-2:123412341234:stack/SpartaGeoIP-mweagle/3e803cd0-fd51-11e8-9c7e-06972e890616" Stack
```

## Verify

With the API Gateway provisioned, let's check the response:

```bash
$ curl -vs https://52a5qqqwo4.execute-api.us-west-2.amazonaws.com/ipgeo/info
*   Trying 13.32.254.81...
* TCP_NODELAY set
* Connected to 52a5qqqwo4.execute-api.us-west-2.amazonaws.com (13.32.254.81) port 443 (#0)
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
*  subjectAltName: host "52a5qqqwo4.execute-api.us-west-2.amazonaws.com" matched cert's "*.execute-api.us-west-2.amazonaws.com"
*  issuer: C=US; O=Amazon; OU=Server CA 1B; CN=Amazon
*  SSL certificate verify ok.
* Using HTTP2, server supports multi-use
* Connection state changed (HTTP/2 confirmed)
* Copying HTTP/2 data in stream buffer to connection buffer after upgrade: len=0
* Using Stream ID: 1 (easy handle 0x7f8522804200)
> GET /ipgeo/info HTTP/2
> Host: 52a5qqqwo4.execute-api.us-west-2.amazonaws.com
> User-Agent: curl/7.54.0
> Accept: */*
>
* Connection state changed (MAX_CONCURRENT_STREAMS updated)!
< HTTP/2 200
< content-type: application/json
< content-length: 1103
< date: Tue, 11 Dec 2018 14:32:00 GMT
< x-amzn-requestid: 851627ca-fd51-11e8-ba5d-9f30493b4ce1
< x-amz-apigw-id: RvyPBHuuPHcFx4w=
< x-amzn-trace-id: Root=1-5c0fca60-2eecbee8bad756981052608c;Sampled=0
< x-cache: Miss from cloudfront
< via: 1.1 400e19a7f70282e0817451f6606ca8f9.cloudfront.net (CloudFront)
< x-amz-cf-id: l4gOpUjDylhS0yHwBWpMneD4BqLBv3zkWcjv6I0j2vBoQu6qD4gKyw==
<
{"ip":"127.0.0.1","record":{"City":{"GeoNameID":0,"Names":null},"Continent":{"Code":"NA","GeoNameID":6255149,"Names":{"de":"Nordamerika","en":"North America","es":"Norteamérica","fr":"Amérique du Nord","ja":"北アメリカ","pt-BR":"América do Norte","ru":"Северная Америка","zh-CN":"北美洲"}},"Country":{"GeoNameID":6252001,"IsInEuropeanUnion":false,"IsoCode":"US","Names":{"de":"USA","en":"United States","es":"Estados Unidos","fr":"États-Unis","ja":"アメリカ合衆国","pt-BR":"Estados Unidos","ru":"США","zh-CN":"美国"}},"Location":{"AccuracyRadius":0,"Latitude":0,"Longitude":0,"MetroCode":0,"TimeZone":""},"Postal":{"Code":""},"RegisteredCountry":{"GeoNameID":6252001,"IsInEuropeanUnion":false,"IsoCode":"US","Names":{"de":"USA","en":"United States","es":"Estados Unidos","fr":"États-Unis","ja":"アメリカ合衆国","pt-BR":"Estados Unidos","ru":"США","zh-CN":"美国"}},"RepresentedCountry":{"GeoNameID":0,"IsInEuropeanUnion":false,"IsoCode":"","Names":null,"Type":""},"Subdivisions":null,"Traits":{"IsAnonymousProxy":false,"IsSatelliteProvider":false}}}

```

Pretty-printing the response body:

```json
{
    "ip": "127.0.0.1",
    "record": {
        "City": {
            "GeoNameID": 0,
            "Names": null
        },
        "Continent": {
            "Code": "NA",
            "GeoNameID": 6255149,
            "Names": {
                "de": "Nordamerika",
                "en": "North America",
                "es": "Norteamérica",
                "fr": "Amérique du Nord",
                "ja": "北アメリカ",
                "pt-BR": "América do Norte",
                "ru": "Северная Америка",
                "zh-CN": "北美洲"
            }
        },
        "Country": {
            "GeoNameID": 6252001,
            "IsInEuropeanUnion": false,
            "IsoCode": "US",
            "Names": {
                "de": "USA",
                "en": "United States",
                "es": "Estados Unidos",
                "fr": "États-Unis",
                "ja": "アメリカ合衆国",
                "pt-BR": "Estados Unidos",
                "ru": "США",
                "zh-CN": "美国"
            }
        },
        "Location": {
            "AccuracyRadius": 0,
            "Latitude": 0,
            "Longitude": 0,
            "MetroCode": 0,
            "TimeZone": ""
        },
        "Postal": {
            "Code": ""
        },
        "RegisteredCountry": {
            "GeoNameID": 6252001,
            "IsInEuropeanUnion": false,
            "IsoCode": "US",
            "Names": {
                "de": "USA",
                "en": "United States",
                "es": "Estados Unidos",
                "fr": "États-Unis",
                "ja": "アメリカ合衆国",
                "pt-BR": "Estados Unidos",
                "ru": "США",
                "zh-CN": "美国"
            }
        },
        "RepresentedCountry": {
            "GeoNameID": 0,
            "IsInEuropeanUnion": false,
            "IsoCode": "",
            "Names": null,
            "Type": ""
        },
        "Subdivisions": null,
        "Traits": {
            "IsAnonymousProxy": false,
            "IsSatelliteProvider": false
        }
    }
}
```

## Clean Up

Before moving on, remember to decommission the service via `go run main.go delete` or `mage delete`.

## Notes

* The _GeoLite2-Country.mmdb_ content is embedded in the go binary via [esc](https://github.com/mjibson/esc) as part of the [go generate](https://github.com/mweagle/SpartaGeoIP/blob/master/main.go#L27) phase.
  * This is a port of Tom Maiaroto's [go-lambda-geoip](https://github.com/tmaiaroto/go-lambda-geoip) implementation.