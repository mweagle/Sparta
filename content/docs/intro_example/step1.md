---
date: 2017-10-03 07:15:40
title: Sample Service
weight: 10
---

Sparta is a framework for developing and deploying **Go** based AWS Lambda-backed microservices.  To help understand what that means we'll begin with a "Hello World" lambda function and eventually deploy that to AWS.  Note that we're not going to handle all error cases to keep the example code to a minimum.

{{< warning title="Pricing" >}}
   Please be aware that running Lambda functions may incur <a href="https://aws.amazon.com/lambda/pricing">costs</a>. Be sure to decommission Sparta stacks after you are finished using them (via the <code>delete</code> command line option) to avoid unwanted charges.  It's likely that you'll be well under the free tier, but secondary AWS resources provisioned during development (eg, Kinesis streams) are not pay-per-invocation.
{{< /warning >}}

# Preconditions

Sparta uses the [AWS SDK for Go](http://aws.amazon.com/sdk-for-go/) to interact with AWS APIs.  Before you get started, ensure that you've properly configured the [SDK credentials](https://github.com/aws/aws-sdk-go/wiki/configuring-sdk).

Note that you must use an AWS region that supports Lambda.  Consult the [Global Infrastructure](https://aws.amazon.com/about-aws/global-infrastructure/regional-product-services/) page for the most up to date release information.

# Lambda Definition

The first place to start is with the lambda function definition.

{{< highlight go >}}
func helloWorld(w http.ResponseWriter, r *http.Request) {
  fmt.Fprintf(w, "Hello World!")
}
{{< /highlight >}}

This is a standard [http.HandlerFunc](https://golang.org/pkg/net/http/#HandlerFunc) signature.

The `http.Request` Body is the raw event input and the `http.ResponseWriter` results are used to return a success/fail message to AWS lambda via a is published back via via the [callback](http://docs.aws.amazon.com/lambda/latest/dg/nodejs-prog-model-context.html). In the case of an HTTP error, the response body is used as the error text.

The [request context](https://golang.org/pkg/context/) object contains two additional objects that provide parity with the existing AWS lambda programming model:

  * The AWS Lambda [Context](http://docs.aws.amazon.com/lambda/latest/dg/nodejs-prog-model-context.html) object is available via:
    * `	lambdaContext, lambdaContextOk := r.Context().Value(sparta.ContextKeyLambdaContext).(*LambdaContext)`
    * This struct includes fields such as `AWSRequestID`, CloudWatch's `LogGroupName`, and the provisioned AWS lambda's ARN (`InvokedFunctionARN`).
  * A [*logrus.Logger](https://github.com/sirupsen/logrus) instance preconfigured to produce JSON output that consumed by CloudWatch. Available via:
    * `	loggerVal, loggerValOK := r.Context().Value(sparta.ContextKeyLogger).(*logrus.Logger)`

All Sparta lambda functions shouuld use this signature starting with version 0.20.0.

{{< note title="Deprecation Notice" >}}
Sparta versions prior to 0.20.0 supported a legacy function signature as in:
```
helloWorld(event *json.RawMessage,
                context *sparta.LambdaContext,
                w http.ResponseWriter,
                logger *logrus.Logger)
```
This is officially deprecated and will be removed in a subsequent release.
{{< /note >}}

# Creation

The next step is to create a Sparta-wrapped version of the `helloWorld` function.

{{< highlight go >}}
var lambdaFunctions []*sparta.LambdaAWSInfo
helloWorldFn := sparta.HandleAWSLambda("Hello World",
  http.HandlerFunc(helloWorld),
  sparta.IAMRoleDefinition{})
lambdaFunctions = append(lambdaFunctions, helloWorldFn)
{{< /highlight >}}

We first declare an empty slice `lambdaFunctions` to which all our service's lambda functions will be appended.  The next step is to register a new lambda target via `HandleAWSLambda`.  `HandleAWSLambda` accepts three parameters:

  * `string`: The function name. A sanitized version of this value is used as the [FunctionName](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-lambda-function.html#cfn-lambda-function-functionname).
  * `http.HandlerFunc`: The **Go** function to execute.
  * `string|IAMRoleDefinition` : *Either* a string literal that refers to a pre-existing IAM Role under which the lambda function will be executed, *OR* a `sparta.IAMRoleDefinition` value that will be provisioned as part of this deployment and used as the execution role for the lambda function.
    - In this example, we're defining a new `IAMRoleDefinition` as part of the stack.  This role definition will automatically include privileges for actions such as CloudWatch logging, and since our function doesn't access any additional AWS services that's all we need.


# Delegation

The final step is to define a Sparta service under your application's `main` package and provide the non-empty slice of lambda functions:

{{< highlight go >}}
sparta.Main("MyHelloWorldStack",
            "Simple Sparta application that demonstrates core functionality",
            lambdaFunctions,
            nil,
            nil)
{{< /highlight >}}

`sparta.Main` accepts five parameters:

  * `serviceName` : The string to use as the CloudFormation stackName. Note that there can be only a single stack with this name within a given AWS account, region pair.
    - The `serviceName` is used as the stable identifier to determine when updates should be applied rather than new stacks provisioned, as well as the target of a `delete` command line request.
    - Consider using [UserScopedStackName](https://godoc.org/github.com/mweagle/Sparta/aws/cloudformation#UserScopedStackName) to generate unique, stable names across a team.
  * `serviceDescription`: An optional string used to describe the stack.
  * `[]*LambdaAWSInfo` : Slice of `sparta.lambdaAWSInfo` that define a service
  * `*API` : Optional pointer to data if you would like to provision and associate an API Gateway with the set of lambda functions.
    - We'll walk through how to do that in [another section](/docs/apigateway/apigateway/), but for now our lambda function will only be accessible via the AWS SDK or Console.
  * `*S3Site` : Optional pointer to data if you would like to provision an [static website on S3](http://docs.aws.amazon.com/AmazonS3/latest/dev/WebsiteHosting.html), initialized with local resources.
    - We'll walk through how to do that in [another section](/docs/s3site), but for now our lambda function will only be accessible via the AWS SDK or Console.

Delegating `main()` to `Sparta.Main()` transforms the set of lambda functions into a standalone executable with several command line options.  Run `go run main.go --help` to see the available options.

# Putting It Together

Putting everything together, and including the necessary imports, we have:

{{< highlight go >}}
// File: main.go
package main

import (
  "fmt"
  "net/http"

  "github.com/Sirupsen/logrus"
  sparta "github.com/mweagle/Sparta"
)

// Standard AWS λ function
func helloWorld(w http.ResponseWriter, r *http.Request) {
  fmt.Fprintf(w, "Hello World!")
}

func main() {
  var lambdaFunctions []*sparta.LambdaAWSInfo
	helloWorldFn := sparta.HandleAWSLambda("Hello World",
		http.HandlerFunc(helloWorld),
		sparta.IAMRoleDefinition{})
  lambdaFunctions = append(lambdaFunctions, helloWorldFn)
  sparta.Main("MyHelloWorldStack",
    "Simple Sparta application that demonstrates core functionality",
    lambdaFunctions,
    nil,
    nil)
}
{{< /highlight >}}

# Running It

Next download the Sparta dependencies via `go get ./...` in the directory that you saved _main.go_.  Once the packages are downloaded, first get a view of what's going on by the `describe` command (replacing `$S3_BUCKET` with an S3 bucket you own):

{{< highlight nohighlight >}}
go run main.go --level info describe --out ./graph.html --s3Bucket $S3_BUCKET

INFO[0000] ========================================
INFO[0000] Welcome to MyHelloWorldStack                  GoVersion=go1.8.3 LinkFlags= Option=describe SpartaSHA=d3479d7 SpartaVersion=0.20.1 UTC="2017-10-03T13:14:34Z"
INFO[0000] ========================================
INFO[0000] Provisioning service                          BuildID=N/A CodePipelineTrigger= InPlaceUpdates=false NOOP=true Tags=
INFO[0000] Verifying IAM Lambda execution roles
INFO[0000] IAM roles verified                            Count=1
INFO[0000] Bypassing S3 upload due to -n/-noop command line argument.  Bucket=weagle VersioningEnabled=false
INFO[0000] Running `go generate`
INFO[0000] Compiling binary                              Name=Sparta.lambda.amd64
INFO[0010] Executable binary size                        KB=22144 MB=21
INFO[0010] Creating code ZIP archive for upload          TempName=./.sparta/MyHelloWorldStack-code.zip
INFO[0010] Registering Sparta JS function                FunctionName=Hello_World ScriptName=Hello_World
INFO[0010] Lambda function deployment package size       KB=22243 MB=21
INFO[0010] Bypassing S3 upload due to -n/-noop command line argument  Bucket=weagle File=MyHelloWorldStack-code.zip Key=MyHelloWorldStack/MyHelloWorldStack-code-3421c386f41a765e8d6abc8820e2b435de5fb827.zip
INFO[0010] Bypassing Stack creation due to -n/-noop command line argument  Bucket=weagle TemplateName=MyHelloWorldStack-cftemplate.json
INFO[0010] ------------------------------------------
INFO[0010] Summary (2017-10-03T06:14:45-07:00)
INFO[0010] ------------------------------------------
INFO[0010] Verifying IAM roles                           Duration (s)=0
INFO[0010] Verifying AWS preconditions                   Duration (s)=0
INFO[0010] Creating code bundle                          Duration (s)=10
INFO[0010] Uploading code                                Duration (s)=0
INFO[0010] Ensuring CloudFormation stack                 Duration (s)=0
INFO[0010] Total elapsed time                            Duration (s)=10
INFO[0010] ------------------------------------------
{{< /highlight >}}

Then open _graph.html_ in your browser (also linked [here](/images/overview/graph.html) ) to see what will be provisioned.

Since everything looks good, we'll provision the stack via `provision` and verify the lambda function.  Note that the `$S3_BUCKET` value must be an S3 bucket to which you have write access since Sparta uploads the lambda package and CloudFormation template to that bucket as part of provisioning.

{{< highlight nohighlight >}}
go run main.go provision --s3Bucket $S3_BUCKET

INFO[0000] ========================================
INFO[0000] Welcome to MyHelloWorldStack                  GoVersion=go1.8.3 LinkFlags= Option=provision SpartaSHA=d3479d7 SpartaVersion=0.20.1 UTC="2017-10-03T13:19:22Z"
INFO[0000] ========================================
INFO[0000] Provisioning service                          BuildID=0ad5fea31a524e8eb4e9fb2ecce8cf784c8a7a12 CodePipelineTrigger= InPlaceUpdates=false NOOP=false Tags=
INFO[0000] Verifying IAM Lambda execution roles
INFO[0000] IAM roles verified                            Count=1
INFO[0000] Checking S3 versioning                        Bucket=weagle VersioningEnabled=true
INFO[0000] Running `go generate`
INFO[0000] Compiling binary                              Name=Sparta.lambda.amd64
INFO[0010] Executable binary size                        KB=22144 MB=21
INFO[0010] Creating code ZIP archive for upload          TempName=./.sparta/MyHelloWorldStack-code.zip
INFO[0010] Registering Sparta JS function                FunctionName=Hello_World ScriptName=Hello_World
INFO[0010] Lambda function deployment package size       KB=22243 MB=21
INFO[0010] Uploading local file to S3                    Bucket=weagle Key=MyHelloWorldStack/MyHelloWorldStack-code.zip Path=./.sparta/MyHelloWorldStack-code.zip
INFO[0026] Uploading local file to S3                    Bucket=weagle Key=MyHelloWorldStack/MyHelloWorldStack-cftemplate.json Path=./.sparta/MyHelloWorldStack-cftemplate.json
INFO[0027] Creating stack                                StackID="arn:aws:cloudformation:us-west-2:027159405834:stack/MyHelloWorldStack/88863e20-a83d-11e7-87a1-500c32c86c29"
INFO[0039] Waiting for CloudFormation operation to complete
INFO[0058] Stack provisioned                             CreationTime="2017-10-03 13:19:49.578 +0000 UTC" StackId="arn:aws:cloudformation:us-west-2:027159405834:stack/MyHelloWorldStack/88863e20-a83d-11e7-87a1-500c32c86c29" StackName=MyHelloWorldStack
INFO[0058] ------------------------------------------
INFO[0058] Summary (2017-10-03T06:20:20-07:00)
INFO[0058] ------------------------------------------
INFO[0058] Verifying IAM roles                           Duration (s)=0
INFO[0058] Verifying AWS preconditions                   Duration (s)=0
INFO[0058] Creating code bundle                          Duration (s)=10
INFO[0058] Uploading code                                Duration (s)=16
INFO[0058] Ensuring CloudFormation stack                 Duration (s)=32
INFO[0058] Total elapsed time                            Duration (s)=58
INFO[0058] ------------------------------------------
{{< /highlight >}}

Once the stack has been provisioned (`CREATE_COMPLETE`), login to the AWS console and navigate to the Lambda section.

# Testing

Find your Lambda function in the list of AWS Lambda functions and click the hyperlink.  The display name will be prefixed by the name of your stack (_MyHelloWorldStack_ in our example):

![AWS Lambda List](/images/overview/AWS_Lambda_List.png)

On the Lambda details page, click the *Test* button:

![AWS Lambda Test](/images/overview/AWS_Lambda_Test.png)

Accept the Input Test Event sample (our Lambda function doesn't consume the event data) and click *Save and test*.  The execution result pane should display something similar to:

![AWS Lambda Execution](/images/overview/AWS_Lambda_Execution.png)

# Cleaning Up

To prevent unauthorized usage and potential charges, make sure to `delete` your stack before moving on:

{{< highlight nohighlight >}}
go run main.go delete

INFO[0000] ========================================
INFO[0000] Welcome to MyHelloWorldStack                  GoVersion=go1.8.3 LinkFlags= Option=delete SpartaSHA=d3479d7 SpartaVersion=0.20.1 UTC="2017-10-03T13:21:19Z"
INFO[0000] ========================================
INFO[0000] Stack existence check                         Exists=true Name=MyHelloWorldStack
INFO[0000] Delete request submitted                      Response="{\n\n}"
{{< /highlight >}}


# Conclusion

Congratulations! You've just deployed your first "serverless" service.  The following sections will dive deeper into what's going on under the hood as well as how to integrate your lambda function(s) into the broader AWS landscape.

# Next Steps

Walkthrough what Sparta actually does to deploy your application in the [next section](/docs/intro_example/step2/).
