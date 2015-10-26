# Sparta <p align="center">

<div align="center"><img src="./SpartanShieldSmall.png" />
</div>

## Overview

Sparta takes a set of _golang_ functions and automatically provisions them in
[AWS Lambda](https://aws.amazon.com/lambda/) as a logical unit.

Functions must implement

    type LambdaFunction func(LambdaEvent, LambdaContext, http.ResponseWriter)

where

  * `LambdaEvent` :  The arbitrary JSON object data provided to the function
  * `LambdaContext` : _golang_ compatible representation of the AWS Lambda [Context](http://docs.aws.amazon.com/lambda/latest/dg/nodejs-prog-model-context.html)
  * `http.ResponseWriter` : Writer for response.  The HTTP status codes & response body is translated to pass/fail results provided to the `context.done()` handler.

Given a set of registered _golang_ functions, Sparta will:

  * Verify the [IAM roles](http://docs.aws.amazon.com/lambda/latest/dg/intro-permission-model.html)
  * Build a deployable application via `Provision()`
  * Zip the contents and associated JS proxying logic
  * Dynamically create a CloudFormation template to either create or update the service state.

Note that Lambda updates may be performed with [no interruption](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-lambda-function.html)
in service.

## Example - Provisioning

  1. Create _application.go_ :
    ```
    package main

    import (
      "fmt"
      sparta "github.com/mweagle/Sparta"
      "net/http"
    )

    // Sparta depends on this role being preconfigured.
    // To create the Lambda Execution role name, see
    // http://docs.aws.amazon.com/lambda/latest/dg/intro-permission-model.html
    const LAMBDA_EXECUTION_ROLE_NAME = "MyLambdaExecutionRole"

    func helloWorld(event sparta.LambdaEvent,
                    context sparta.LambdaContext,
                    w http.ResponseWriter) {
      fmt.Fprintf(w, "Hello World. Event data: %s", event)
    }

    func main() {
      var lambdaFunctions []*sparta.LambdaAWSInfo
      lambdaFunctions = append(lambdaFunctions,
                                sparta.NewLambda(LAMBDA_EXECUTION_ROLE_NAME,
                                helloWorld,
                                nil))
      sparta.Main("HelloWorldApp", "This is the Hello World service", lambdaFunctions)
    }
    ```
  1. `go get ./...`
  1. `go run application.go provision --s3Bucket MY_S3_BUCKET_NAME`
      - You'll need to change *MY_S3_BUCKET_NAME* to an accessible S3 bucketname
  1. Visit the AWS Lambda console and confirm your Lambda function is accessible

See also the [Sparta Application](https://github.com/mweagle/SpartaApplication) for
an example.

### Prerequisites

  1. Verify your golang SDK credentials are [properly configured](https://github.com/aws/aws-sdk-go/wiki/Getting-Started-Credentials)
  1. Verify that the lambda IAM Permissions are [properly configured](http://docs.aws.amazon.com/lambda/latest/dg/intro-permission-model.html) and that the correct IAM RoleName is provided to `sparta.NewLambda()`
      - More information on the Lambda permission model is available [here](https://aws.amazon.com/blogs/compute/easy-authorization-of-aws-lambda-functions)

## Example - Describing

It's also possible to generate a visual representation of your Lambda connections
via the `describe` command line argument.

![Description Sample Output](./describe.jpg)

## Additional documentation

Run `godoc` in the source directory.

## Caveats

  1. This is my first _golang_ project - YMMV
  1. It's a POC first release
    - Do not run your next [$1B unicorn](https://en.wikipedia.org/wiki/Unicorn_%28finance%29) on it
    - Or if you do, perhaps we should have coffee?
  1. _golang_ isn't officially supported by AWS (yet)
    - Because of this, there is a per-container initialization cost of:
        - Copying the embedded binary to _/tmp_
        - Changing the binary permissions
        - Launching it from the new location
        - See the [AWS Forum](https://forums.aws.amazon.com/message.jspa?messageID=583910) for more background
    - Depending on [container reuse](https://aws.amazon.com/blogs/compute/container-reuse-in-lambda/), this initialization penalty (~`700ms`) may prove burdensome.
  1. `dry-run` execution isn't yet implemented
  1. There's bound to be more.



