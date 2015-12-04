+++
author = "Matt Weagle"
date = "2015-11-29T06:50:17"
title = "Overview"
tags = ["sparta"]
type = "doc"
+++

Sparta is a framework for developing and deploying *Go* based AWS Lambda functions.  To help understand what that means we'll begin with a "Hello World" lambda function and eventually deploy that to AWS.  Note that we're not going to handle all error cases to keep the example code to a minimum.


## <a href="{{< relref "#preconditions" >}}">Preconditions</a>

Sparta uses the [AWS SDK for Go](http://aws.amazon.com/sdk-for-go/) to interact with AWS APIs.  Before you get started, ensure that you've properly configured the [SDK credentials](https://github.com/aws/aws-sdk-go/wiki/configuring-sdk).  

Note that you must use an AWS region that supports Lambda.  Consult the [Global Infrastructure](https://aws.amazon.com/about-aws/global-infrastructure/regional-product-services/) page for the most up to date release information.

## <a href="{{< relref "#lambdaDefinition" >}}">Lambda Definition</a>

The first place to start is with the lambda function definition.

{{< highlight go >}}

func helloWorld(event *json.RawMessage,
                context *sparta.LambdaContext,
                w http.ResponseWriter,
                logger *logrus.Logger) {
	fmt.Fprintf(w, "Hello World!")
}

{{< /highlight >}}      

All Sparta lambda functions have the same function signature that is composed of:

  * `json.RawMessage` :  The arbitrary `json.RawMessage` event data provided to the function. Implementations may further unmarshal this data into event specific representations for events such as S3 item changes, API Gateway requests, etc.
  * `LambdaContext` : *Go* compatible representation of the AWS Lambda [Context](http://docs.aws.amazon.com/lambda/latest/dg/nodejs-prog-model-context.html). This struct includes fields such as `AWSRequestID`, CloudWatch's `LogGroupName`, and the provisioned AWS lambda's ARN (`InvokedFunctionARN`).
  * `http.ResponseWriter` : The writer for any response data. Sparta uses the HTTP status code to determine the functions success or failure status, and any data written to the `responseWriter` is published back via [context.done()](http://docs.aws.amazon.com/lambda/latest/dg/nodejs-prog-model-context.html).
  * `logrus.Logger` : A [logrus](https://github.com/Sirupsen/logrus) logger preconfigured to produce JSON output.  Content written to this logger will be available in CloudWatch logs.

## <a href="{{< relref "#creation" >}}">Creation</a>

The next step is to create a Sparta-wrapped version of the `helloWorld` function.  

{{< highlight go >}}

var lambdaFunctions []*sparta.LambdaAWSInfo

helloWorldFn := sparta.NewLambda(sparta.IAMRoleDefinition{},
                                helloWorld,
                                nil)
lambdaFunctions = append(lambdaFunctions, helloWorldFn)
{{< /highlight >}}    

We first declare an empty slice `lambdaFunctions` to which all our service's lambda functions will be appended.  The next step is to create a new lambda function via `NewLambda`.  `NewLambda` accepts three parameters:

  * `string|IAMRoleDefinition` : Either a string literal that refers to a pre-existing IAM role under which the lambda function will be executed, *OR* a `sparta.IAMRoleDefinition` that will be provisioned as part of this deployment and used as the execution role for the lambda function.
    - In this example, we're defining a new `IAMRoleDefinition` as part of the stack.  This role definition will automatically include privileges for actions such as CloudWatch logging, and since our function doesn't access any additional AWS services that's all we need.
  * `LambdaFunction`: The *Go* function to execute.
  * `*LambdaFunctionOptions`: A pointer to any additional execution settings (eg, timeout, memory settings, etc).

## <a href="{{< relref "#delegation" >}}">Delegation</a>

The final step is to define a Sparta service under your applications `main` package and provide the non-empty slice of lambda functions:

{{< highlight go >}}
sparta.Main("MyHelloWorldStack",
            "Simple Sparta application that demonstrates core functionality",
            lambdaFunctions,
            nil)
{{< /highlight >}}    

`sparta.Main` accepts four parameters:

  * `serviceName` : The string to use as the CloudFormation stackName. Note that there can be only a single stack with this name within a given AWS account, region pair.
    - The `serviceName` is used as the stable identifier to determine when updates should be applied vs new stacks provisioned.
  * `serviceDescription`: An optional string used to describe the stack.
  * `[]*LambdaAWSInfo` : Slice of `sparta.lambdaAWSInfo` to provision
  * `*API` : Optional pointer to data if you would like to provision and associate an API Gateway with the set of lambda functions.
    - We'll walk through how to do that in a later example, but for now our lambda function will only be accessible via the AWS SDK or Console.

Delegating `main()` to `Sparta.Main()` transforms the set of lambda functions into a standalone executable with several command line options.  Run `go run main.go --help` to see the available options.

## <a href="{{< relref "#puttingItTogether" >}}">Putting It Together</a>

Putting everything together, and including the necessary imports, we have:

{{< highlight go >}}
// File: main.go
package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
	sparta "github.com/mweagle/Sparta"
)

func helloWorld(event *json.RawMessage, context *sparta.LambdaContext, w http.ResponseWriter, logger *logrus.Logger) {
	fmt.Fprintf(w, "Hello World!")
}

func main() {
	var lambdaFunctions []*sparta.LambdaAWSInfo

	helloWorldFn := sparta.NewLambda(sparta.IAMRoleDefinition{},
		helloWorld,
		nil)
	lambdaFunctions = append(lambdaFunctions, helloWorldFn)
	sparta.Main("MyHelloWorldStack",
		"Simple Sparta application that demonstrates core functionality",
		lambdaFunctions,
		nil)
}
{{< /highlight >}}      

## <a href="{{< relref "#runningIt" >}}">Running It</a>

Next download the Sparta dependencies via `go get ./...` in the directory that you saved _main.go_.  Once the packages are downloaded, first get a view of what's going on by the `describe` command:

{{< highlight bash >}}

go run main.go describe --out ./graph.html

INFO[0000] Welcome to Sparta                             Option=describe Version=0.0.7
INFO[0000] Verifying IAM Lambda execution roles
INFO[0000] IAM roles verified. Count: 1
INFO[0000] Compiling binary: MyHelloWorldStack.lambda.amd64
INFO[0007] Executable binary size (MB): 10
INFO[0007] Creating ZIP archive for upload: /Users/mweagle/Documents/golang/workspace/src/HelloWorld/MyHelloWorldStack737464669
INFO[0008] Creating NodeJS proxy entry: main_helloWorld
INFO[0008] Embedding CustomResource script: cfn-response.js
INFO[0008] Embedding CustomResource script: underscore-min.js
INFO[0008] Embedding CustomResource script: async.min.js
INFO[0008] Embedding CustomResource script: apigateway.js
INFO[0008] Embedding CustomResource script: s3.js
INFO[0008] Embedding CustomResource script: sns.js
INFO[0008] Embedding CustomResource script: golang-constants.json
INFO[0008] Bypassing S3 ZIP upload due to -n/-noop command line argument  Bucket=S3Bucket Key=MyHelloWorldStack737464669
INFO[0008] Bypassing template upload & creation due to -n/-noop command line argument  Bucket=S3Bucket Key=MyHelloWorldStack-edaad4631616d70ff87806dfd1399b0bc2f7994a-cf.json
{{< /highlight >}}


Then open _graph.html_ in your browser (also linked [here](/images/overview/graph.html) ) to see what will be provisioned.

Since everything looks good, we'll provision the stack via `provision` and verify the lambda function.  Note that the `$S3_BUCKET` value must be an S3 bucket to which you have write access since Sparta uploads the lambda package and CloudFormation template to that bucket as part of provisioning.

{{< highlight bash >}}
go run main.go provision --s3Bucket $S3_BUCKET

INFO[0000] Welcome to Sparta                             Option=provision Version=0.0.7
INFO[0000] Verifying IAM Lambda execution roles
INFO[0000] IAM roles verified. Count: 1
INFO[0000] Compiling binary: MyHelloWorldStack.lambda.amd64
INFO[0007] Executable binary size (MB): 10
INFO[0007] Creating ZIP archive for upload: /Users/mweagle/Documents/golang/workspace/src/HelloWorld/MyHelloWorldStack650982716
INFO[0008] Creating NodeJS proxy entry: main_helloWorld
INFO[0008] Embedding CustomResource script: cfn-response.js
INFO[0008] Embedding CustomResource script: underscore-min.js
INFO[0008] Embedding CustomResource script: async.min.js
INFO[0008] Embedding CustomResource script: apigateway.js
INFO[0008] Embedding CustomResource script: s3.js
INFO[0008] Embedding CustomResource script: sns.js
INFO[0008] Embedding CustomResource script: golang-constants.json
INFO[0008] Uploading ZIP archive to S3
INFO[0012] ZIP archive uploaded: https://weagle.s3-us-west-2.amazonaws.com/MyHelloWorldStack650982716
INFO[0012] Uploading CloudFormation template
INFO[0012] CloudFormation template uploaded: https://weagle.s3-us-west-2.amazonaws.com/MyHelloWorldStack-8ae100efb4eddbed9debb45915a288a179b6592e-cf.json
INFO[0012] DescribeStackOutputError: ValidationError: Stack with id MyHelloWorldStack does not exist
	status code: 400, request id: defc0375-9a05-11e5-8ef6-d5823d3e35af
INFO[0012] Creating stack: arn:aws:cloudformation:us-west-2:123412341234:stack/MyHelloWorldStack/df0d6860-9a05-11e5-884c-5001230106a6
INFO[0012] Waiting for stack to complete
INFO[0022] Current state: CREATE_IN_PROGRESS
INFO[0052] Current state: CREATE_COMPLETE
INFO[0052] Stack Outputs:
INFO[0052] 	Output                                       Description=Sparta Home Key=SpartaHome Value=https://github.com/mweagle/Sparta
INFO[0052] 	Output                                       Description=Sparta Version Key=SpartaVersion Value=0.0.7
INFO[0052] Stack provisioned: {
  Capabilities: ["CAPABILITY_IAM"],
  CreationTime: 2015-12-03 21:36:11.325 +0000 UTC,
  Description: "Simple Sparta application that demonstrates core functionality",
  DisableRollback: false,
  Outputs: [{
      Description: "Sparta Home",
      OutputKey: "SpartaHome",
      OutputValue: "https://github.com/mweagle/Sparta"
    },{
      Description: "Sparta Version",
      OutputKey: "SpartaVersion",
      OutputValue: "0.0.7"
    }],
  StackId: "arn:aws:cloudformation:us-west-2:123412341234:stack/MyHelloWorldStack/df0d6860-9a05-11e5-884c-5001230106a6",
  StackName: "MyHelloWorldStack",
  StackStatus: "CREATE_COMPLETE",
  TimeoutInMinutes: 5
}
{{< /highlight >}}

Once the stack has been provisioned (`CREATE_COMPLETE`), login to the AWS console and navigate to the Lambda section.

## Testing

Find your Lambda function in the list of AWS Lambda functions and click the hyperlink.  The display name will be prefixed by the name of your stack (_MyHelloWorldStack_ in our example):

![AWS Lambda List](/images/overview/AWS_Lambda_List.png)

On the Lambda details page, click the *Test* button:

![AWS Lambda Test](/images/overview/AWS_Lambda_Test.png)

Accept the Input Test Event sample (our Lambda function doesn't consume the event data) and click *Save and test*.  The execution result pane should display something similar to:

![AWS Lambda Execution](/images/overview/AWS_Lambda_Execution.png)

## <a href="{{< relref "#cleaningUp" >}}">Cleaning Up</a>

To prevent unauthorized usage and potential charges, make sure to `delete` your stack before moving on:

{{< highlight bash >}}
go run main.go delete

INFO[0000] Welcome to Sparta                             Option=delete Version=0.0.7
INFO[0000] Stack exists: MyHelloWorldStack
INFO[0000] Stack delete issued: {

}
{{< /highlight >}}


## <a href="{{< relref "#conclusion" >}}">Conclusion</a>

Congratulations! You've just deployed your first "serverless" service.  The following sections will dive deeper into what's going on under the hood as well as how to integrate your lambda function(s) into the broader AWS landscape.      

Next: [Walkthrough](/docs/walkthrough)
