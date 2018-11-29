---
date: 2016-03-09T19:56:50+01:00
title: Request Context
weight: 12
---


This example demonstrates how to use the `Context` struct provided as part of the [APIGatewayLambdaJSONEvent](https://godoc.org/github.com/mweagle/Sparta#APIGatewayLambdaJSONEvent) event.  The [SpartaGeoIP](https://github.com/mweagle/SpartaGeoIP) service will return Geo information based on the inbound request's IP address.

# Define the Lambda Function

Our function will examine the inbound request, lookup the user's IP address in the [GeoLite2 Database](http://dev.maxmind.com/geoip/geoip2/geolite2/) and return any information to the client.

As this function is only expected to be invoked from the API Gateway, we'll unmarshall the inbound event:


```go
import (
	spartaAWSEvents "github.com/mweagle/Sparta/aws/events"
)
func ipGeoLambda(ctx context.Context,
  apiRequest spartaAWSEvents.APIGatewayRequest) (map[string]interface{}, error) {
	parsedIP := net.ParseIP(apiRequest.Context.Identity.SourceIP)
	record, err := dbHandle.City(parsedIP)

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
	return requestResponse, nil
```

# Sparta Integration

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
  lambdaFn := sparta.HandleAWSLambda(
    sparta.LambdaName(ipGeoLambda),
    ipGeoLambda,
    sparta.IAMRoleDefinition{})
	apiGatewayResource, _ := apiGateway.NewResource("/info", lambdaFn)
	apiGatewayResource.NewMethod("GET", http.StatusOK)
	lambdaFunctions = append(lambdaFunctions, lambdaFn)

	sparta.Main(stackName,
		"Sparta app supporting ip->geo mapping",
		lambdaFunctions,
		apiGateway,
    nil)
}
```

# Provision

The next step is to provision the stack:

```nohighlight
S3_BUCKET=<MY-S3-BUCKETNAME> make provision
```

Assuming all goes well, the log output will include the API Gateway URL as in:

```nohighlight
INFO[0113] Stack output   Description=API Gateway URL Key=APIGatewayURL Value=https://qyslujefsf.execute-api.us-west-2.amazonaws.com/ipgeo
INFO[0113] Stack output   Description=Sparta Home Key=SpartaHome Value=https://github.com/mweagle/Sparta
INFO[0113] Stack output   Description=Sparta Version Key=SpartaVersion Value=0.1.0
```

# Query

With the API Gateway provisioned, let's check the response:

```bash
curl -vs https://qyslujefsf.execute-api.us-west-2.amazonaws.com/ipgeo/info

*   Trying 54.192.70.206...
* Connected to qyslujefsf.execute-api.us-west-2.amazonaws.com (54.192.70.206) port 443 (#0)
* TLS 1.2 connection using TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
* Server certificate: *.execute-api.us-west-2.amazonaws.com
* Server certificate: Symantec Class 3 Secure Server CA - G4
* Server certificate: VeriSign Class 3 Public Primary Certification Authority - G5
> GET /ipgeo/info HTTP/1.1
> Host: qyslujefsf.execute-api.us-west-2.amazonaws.com
> User-Agent: curl/7.43.0
> Accept: */*
>
< HTTP/1.1 200 OK
< Content-Type: application/json
< Content-Length: 1129
< Connection: keep-alive
< Date: Sun, 06 Dec 2015 21:50:18 GMT
< x-amzn-RequestId: 572adc18-9c63-11e5-b827-81d99c02192f
< X-Cache: Miss from cloudfront
< Via: 1.1 29bfa9b96f4ea66dc02526ee845ca6b0.cloudfront.net (CloudFront)
< X-Amz-Cf-Id: 5mXHuOlbDyk5CejDouAy7nUS3YUn4eXJdQWzU_1VqX9Yh5PE_BdlAw==
<
* Connection #0 to host qyslujefsf.execute-api.us-west-2.amazonaws.com left intact
{"code":200,"status":"OK","headers":{"content-type":"application/json","date":"Sun, 06 Dec 2015 21:50:18 GMT","content-length":"984"},"results":{"info":{"City":{"GeoNameID":0,"Names":null},"Continent":{"Code":"NA","GeoNameID":6255149,"Names":{"de":"Nordamerika","en":"North America","es":"Norteamérica","fr":"Amérique du Nord","ja":"北アメリカ","pt-BR":"América do Norte","ru":"Северная Америка","zh-CN":"北美洲"}},"Country":{"GeoNameID":6252001,"IsoCode":"US","Names":{"de":"USA","en":"United States","es":"Estados Unidos","fr":"États-Unis","ja":"アメリカ合衆国","pt-BR":"Estados Unidos","ru":"США","zh-CN":"美国"}},"Location":{"Latitude":0,"Longitude":0,"MetroCode":0,"TimeZone":""},"Postal":{"Code":""},"RegisteredCountry":{"GeoNameID":6252001,"IsoCode":"US","Names":{"de":"USA","en":"United States","es":"Estados Unidos","fr":"États-Unis","ja":"アメリカ合衆国","pt-BR":"Estados Unidos","ru":"США","zh-CN":"美国"}},"RepresentedCountry":{"GeoNameID":0,"IsoCode":"","Names":null,"Type":""},"Subdivisions":null,"Traits":{"IsAnonymousProxy":false,"IsSatelliteProvider":false}}}}

```

Pretty-printing the response body:


```json

{
  "code": 200,
  "status": "OK",
  "headers": {
    "content-type": "application/json",
    "date": "Sun, 06 Dec 2015 17:50:15 GMT",
    "content-length": "984"
  },
  "results": {
    "info": {
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
}
```


Please see the [first example](/reference/apigateway/echo_event/) for more information on the `code`, `status`, and `headers` keys.

# Cleaning Up

Before moving on, remember to decommission the service via:

```nohighlight
go run main.go delete
```

# Notes

  * The _GeoLite2-Country.mmdb_ content is embedded in the go binary via [esc](https://github.com/mjibson/esc) as part of the [go generate](https://github.com/mweagle/SpartaGeoIP/blob/master/main.go#L27) phase.
  * This is a port of Tom Maiaroto's https://github.com/tmaiaroto/go-lambda-geoip implementation.
