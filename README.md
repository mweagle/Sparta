[![Build Status](https://travis-ci.org/mweagle/Sparta.svg?branch=master)](https://travis-ci.org/mweagle/Sparta)

# Sparta <p align="center">

<div align="center"><img src="https://raw.githubusercontent.com/mweagle/Sparta/master/site/SpartanShieldSmall.png" />
</div>

## Overview

Sparta takes a set of _golang_ functions and automatically provisions them in
[AWS Lambda](https://aws.amazon.com/lambda/) as a logical unit.

Functions must implement

    type LambdaFunction func(*json.RawMessage,
                              *LambdaContext,
                              http.ResponseWriter,
                              *logrus.Logger)

where

  * `json.RawMessage` :  The arbitrary `json.RawMessage` event data provided to the function.
  * `LambdaContext` : _golang_ compatible representation of the AWS Lambda [Context](http://docs.aws.amazon.com/lambda/latest/dg/nodejs-prog-model-context.html)
  * `http.ResponseWriter` : Writer for response. The HTTP status code & response body is translated to a pass/fail result provided to the `context.done()` handler.
  * `logrus.Logger` : [logrus](https://github.com/Sirupsen/logrus) logger with JSON output. See an [example](https://github.com/Sirupsen/logrus#example) for including JSON fields.

Given a set of registered _golang_ functions, Sparta will:

  * Either verify or provision the defined [IAM roles](http://docs.aws.amazon.com/lambda/latest/dg/intro-permission-model.html)
  * Build a deployable application via `Provision()`
  * Zip the contents and associated JS proxying logic
  * Dynamically create a CloudFormation template to either create or update the service state.
  * Optionally:
    * Register with S3 and SNS for push source configuration
    * Provision an [API Gateway](https://aws.amazon.com/api-gateway/) service to make your functions publicly available

Note that Lambda updates may be performed with [no interruption](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-lambda-function.html)
in service.

## Sample Lambda Application

  1. Create _application.go_ :

    ```go
    package main

    import (
      "encoding/json"
      "fmt"
      "github.com/Sirupsen/logrus"
      sparta "github.com/mweagle/Sparta"
      "net/http"
    )

    func echoEvent(event *sparta.LambdaEvent,
                   context *sparta.LambdaContext,
                   w http.ResponseWriter,
                   logger *logrus.Logger) {

      logger.WithFields(logrus.Fields{
        "RequestID": context.AWSRequestID,
      }).Info("Request received")

      eventData, err := json.Marshal(*event)
      if err != nil {
        logger.Error("Failed to marshal event data: ", err.Error())
        http.Error(w, err.Error(), http.StatusInternalServerError)
      }
      logger.Info("Event data: ", string(eventData))
    }

    func main() {
      var lambdaFunctions []*sparta.LambdaAWSInfo

      lambdaEcho := sparta.NewLambda(sparta.IAMRoleDefinition{},
                                      echoEvent,
                                      nil)
      lambdaFunctions = append(lambdaFunctions, lambdaEcho)
      sparta.Main("SpartaEcho",
                   "This is a sample Sparta application",
                   lambdaFunctions)
    }
    ```

  1. `go get ./...`
  1. `go run application.go provision --s3Bucket MY_S3_BUCKET_NAME`
      - You'll need to change *MY_S3_BUCKET_NAME* to an accessible S3 bucketname
  1. Visit the AWS Lambda console and confirm your Lambda function is accessible

See also the [Sparta Application](https://github.com/mweagle/SpartaApplication) for
an example.

## Examples - Advanced

The `[]sparta.LambdaAWSInfo.Permissions` slice allows Lambda functions to automatically manage remote [event source](http://docs.aws.amazon.com/lambda/latest/dg/intro-core-components.html#intro-core-components-event-sources) subscriptions. Push-based event sources are updated via CustomResources that are injected into the CloudFormation template if appropriate.

Examples:

  * [S3 Subscriber](https://github.com/mweagle/Sparta/blob/master/doc_s3permission_test.go)
  * [SNS Subscriber](https://github.com/mweagle/Sparta/blob/master/doc_snspermission_test.go)

The per-service API logic is inline NodeJS [ZipFile](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-lambda-function-code.html#cfn-lambda-function-code-zipfile) code. See the [provision](https://github.com/mweagle/Sparta/tree/master/resources/provision)
directory for more.

See also the [Sparta Application](https://github.com/mweagle/SpartaApplication) for a standalone example.

## Examples - API Gateway (Preliminary)

It's possible to expose to your _golang_ functions over HTTPS by associating an API Gateway.  To enable API Gateway support, you must:
  * Define a  [Stage](http://docs.aws.amazon.com/apigateway/latest/developerguide/stages.html)
  * Define an [API Gateway](http://docs.aws.amazon.com/apigateway/latest/developerguide/how-to-create-api.html)  name and provide the previously defined stage.
  * Create one or more `(resource, httpMethod)` pairs that are bound to your lambda function.  

Example:

```go
package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
	sparta "github.com/mweagle/Sparta"
)

////////////////////////////////////////////////////////////////////////////////
// Echo handler
//
func echoEvent(event *json.RawMessage,
                context *sparta.LambdaContext,
                w http.ResponseWriter,
                logger *logrus.Logger) {
	logger.WithFields(logrus.Fields{
		"RequestID": context.AWSRequestID,
		"Event":     string(*event),
	}).Info("Request received")

	fmt.Fprintf(w, "Hello World!")
}

func main() {
	stage := sparta.NewStage("test")
	apiGateway := sparta.NewAPIGateway("MySpartaAPI", stage)

	var lambdaFunctions []*sparta.LambdaAWSInfo
	lambdaFn := sparta.NewLambda(sparta.IAMRoleDefinition{}, echoEvent, nil)
	apiGatewayResource, _ := apiGateway.NewResource("/hello/echo", lambdaFn)
	apiGatewayResource.NewMethod("GET")

	lambdaFunctions = append(lambdaFunctions, lambdaFn)
	sparta.Main("SampleApplication",
		"Sample application with API Gateway support",
		lambdaFunctions,
		apiGateway)
}

```

This API can be deployed via:

```go run api.go --level debug provision --s3Bucket $MY_S3_BUCKET_NAME```

The provisioning log will output the AWS-assigned API Gateway URL as in:

```
...
Outputs: [{
    Description: "API Gateway URL",
    OutputKey: "URL",
    OutputValue: "https://vpv0e9nv83.execute-api.us-west-2.amazonaws.com/test"
  }],
...
```

You can then access a specific resource by appending the path component to the _OutputValue_ of the _URL_ value as in:

```bash
$ curl -vs https://jhn4bubx7h.execute-api.us-west-2.amazonaws.com/test/hello/echo

*   Trying 54.230.147.237...
* Connected to jhn4bubx7h.execute-api.us-west-2.amazonaws.com (54.230.147.237) port 443 (#0)
* TLS 1.2 connection using TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
* Server certificate: *.execute-api.us-west-2.amazonaws.com
* Server certificate: Symantec Class 3 Secure Server CA - G4
* Server certificate: VeriSign Class 3 Public Primary Certification Authority - G5
> GET /test/hello/echo HTTP/1.1
> Host: jhn4bubx7h.execute-api.us-west-2.amazonaws.com
> User-Agent: curl/7.43.0
> Accept: */*
>
< HTTP/1.1 200 OK
< Content-Type: application/json
< Content-Length: 14
< Connection: keep-alive
< Date: Mon, 23 Nov 2015 17:02:01 GMT
< x-amzn-RequestId: e968e0f1-9203-11e5-9134-61048221e24a
< X-Cache: Miss from cloudfront
< Via: 1.1 5687015cb50d88319b87aae0ee898267.cloudfront.net (CloudFront)
< X-Amz-Cf-Id: lxP1CTKaV5ArYkfZCNJeWnUaF-63AwCM3-puXo315d_HwSvdhz7ibQ==
<
* Connection #0 to host jhn4bubx7h.execute-api.us-west-2.amazonaws.com left intact
"Hello World!"
```

**NOTE:** Providing `nil` as the Stage argument to `sparta.NewAPIGateway()` will provision an API instance, but will not [deploy](http://docs.aws.amazon.com/apigateway/latest/developerguide/how-to-deploy-api.html) it.  

### Prerequisites

  1. Verify your golang SDK credentials are [properly configured](https://github.com/aws/aws-sdk-go/wiki/Getting-Started-Credentials)
  1. If referring to pre-existing IAM Roles, verify that the Lambda IAM Permissions are [properly configured](http://docs.aws.amazon.com/lambda/latest/dg/intro-permission-model.html) and that the correct IAM RoleName is provided to `sparta.NewLambda()`
      - More information on the Lambda permission model is available [here](https://aws.amazon.com/blogs/compute/easy-authorization-of-aws-lambda-functions)

## Lambda Flow Graph

It's also possible to generate a visual representation of your Lambda connections
via the `describe` command line argument.

```
go run application.go describe --out ./graph.html && open ./graph.html
```

![Description Sample Output](https://raw.githubusercontent.com/mweagle/Sparta/master/site/describe.jpg)

## Additional documentation

View the latest versions at [GoDoc](https://godoc.org/github.com/mweagle/Sparta) or run `make docs` in the source directory & visit http://localhost:8090.

## Caveats

  1. _golang_ isn't officially supported by AWS (yet)
    - But, you can [vote](https://twitter.com/awscloud/status/659795641204260864) to make _golang_ officially supported.
    - Because of this, there is a per-container initialization cost of:
        - Copying the embedded binary to _/tmp_
        - Changing the binary permissions
        - Launching it from the new location
        - See the [AWS Forum](https://forums.aws.amazon.com/message.jspa?messageID=583910) for more background
    - Depending on [container reuse](https://aws.amazon.com/blogs/compute/container-reuse-in-lambda/), this initialization penalty (~`700ms`) may prove burdensome.
    - See the [JAWS](https://github.com/jaws-framework/JAWS) project for a pure NodeJS alternative.
    - See the [PAWS](https://github.com/braahyan/PAWS) project for a pure Python alternative.
  1. There are [Lambda Limits](http://docs.aws.amazon.com/lambda/latest/dg/limits.html) that may affect your development

## Outstanding
  - Eliminate NodeJS CustomResources
  - Support API Gateway updates
    - Currently API reprovisioning is done by `delete` => `create`
  - Optimize _CONSTANTS.go_ for deployed binary
  - Implement APIGateway graph
  - Support APIGateway inline Model definition
  - Support custom domains
