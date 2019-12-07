---
date: 2017-10-03 07:15:40
title: Overview
weight: 10
---

Sparta is a framework for developing and deploying **go** based AWS Lambda-backed microservices. To help understand what that means we'll begin with a "Hello World" lambda function and eventually deploy that to AWS. Note that we're not going to handle all error cases to keep the example code to a minimum.

{{% notice warning %}}
Please be aware that running Lambda functions may incur [costs](https://aws.amazon.com/lambda/pricing"). Be sure to decommission Sparta stacks after you are finished using them (via the `delete` command line option) to avoid unwanted charges. It's likely that you'll be well under the free tier, but secondary AWS resources provisioned during development (eg, Kinesis streams) are not pay-per-invocation.
{{% /notice %}}

# Preconditions

Sparta uses the [AWS SDK for Go](http://aws.amazon.com/sdk-for-go/) to interact with AWS APIs. Before you get started, ensure that you've properly configured the [SDK credentials](https://github.com/aws/aws-sdk-go/wiki/configuring-sdk).

Note that you must use an AWS region that supports Lambda. Consult the [Global Infrastructure](https://aws.amazon.com/about-aws/global-infrastructure/regional-product-services/) page for the most up to date release information.

# Lambda Definition

The first place to start is with the lambda function definition.

```go
// Standard AWS λ function
func helloWorld(ctx context.Context) (string, error) {
  return "Hello World!", nil
}
```

The `ctx` parameter includes the following entries:

- The [AWS LambdaContext](https://godoc.org/github.com/aws/aws-lambda-go/lambdacontext#FromContext)
- A [\*logrus.Logger](https://github.com/sirupsen/logrus) instance (`sparta.ContextKeyLogger`)
- A per-request annotated [\*logrus.Entry](https://godoc.org/github.com/sirupsen/logrus#Entry) instance (`sparta.ContextKeyRequestLogger`)

# Creation

The next step is to create a Sparta-wrapped version of the `helloWorld` function.

```go
var lambdaFunctions []*sparta.LambdaAWSInfo
helloWorldFn, _ := sparta.NewAWSLambda("Hello World",
  helloWorld,
  sparta.IAMRoleDefinition{})
```

We first declare an empty slice `lambdaFunctions` to which all our service's lambda functions will be appended. The next step is to register a new lambda target via [NewAWSLambda](https://godoc.org/github.com/mweagle/Sparta#NewAWSLambda). `NewAWSLambda` accepts three parameters:

- `string`: The function name. A sanitized version of this value is used as the [FunctionName](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-lambda-function.html#cfn-lambda-function-functionname).
- `func(...)`: The **go** function to execute.
- `string|IAMRoleDefinition` : _Either_ a string literal that refers to a pre-existing IAM Role under which the lambda function will be executed, _OR_ a `sparta.IAMRoleDefinition` value that will be provisioned as part of this deployment and used as the execution role for the lambda function.
  - In this example, we're defining a new `IAMRoleDefinition` as part of the stack. This role definition will automatically include privileges for actions such as CloudWatch logging, and since our function doesn't access any additional AWS services that's all we need.

# Delegation

The final step is to define a Sparta service under your application's `main` package and provide the non-empty slice of lambda functions:

```go
sparta.Main("MyHelloWorldStack",
  "Simple Sparta application that demonstrates core functionality",
  lambdaFunctions,
  nil,
  nil)
```

`sparta.Main` accepts five parameters:

- `serviceName` : The string to use as the CloudFormation stackName. Note that there can be only a single stack with this name within a given AWS account, region pair.
  - The `serviceName` is used as the stable identifier to determine when updates should be applied rather than new stacks provisioned, as well as the target of a `delete` command line request.
  - Consider using [UserScopedStackName](https://godoc.org/github.com/mweagle/Sparta/aws/cloudformation#UserScopedStackName) to generate unique, stable names across a team.
- `serviceDescription`: An optional string used to describe the stack.
- `[]*LambdaAWSInfo` : Slice of `sparta.lambdaAWSInfo` that define a service
- `*API` : Optional pointer to data if you would like to provision and associate an API Gateway with the set of lambda functions.
  - We'll walk through how to do that in [another section](/reference/apigateway/apigateway/), but for now our lambda function will only be accessible via the AWS SDK or Console.
- `*S3Site` : Optional pointer to data if you would like to provision an [static website on S3](http://docs.aws.amazon.com/AmazonS3/latest/dev/WebsiteHosting.html), initialized with local resources.
  - We'll walk through how to do that in [another section](/reference/s3site), but for now our lambda function will only be accessible via the AWS SDK or Console.

Delegating `main()` to `Sparta.Main()` transforms the set of lambda functions into a standalone executable with several command line options. Run `go run main.go --help` to see the available options.

# Putting It Together

Putting everything together, and including the necessary imports, we have:

```go
// File: main.go
package main

import (
  "context"

  sparta "github.com/mweagle/Sparta"
)

// Standard AWS λ function
func helloWorld(ctx context.Context) (string, error) {
  return "Hello World!", nil
}

func main() {
  var lambdaFunctions []*sparta.LambdaAWSInfo
  helloWorldFn, _ := sparta.NewAWSLambda("Hello World",
    helloWorld,
    sparta.IAMRoleDefinition{})
  lambdaFunctions = append(lambdaFunctions, helloWorldFn)
  sparta.Main("MyHelloWorldStack",
    "Simple Sparta application that demonstrates core functionality",
    lambdaFunctions,
    nil,
    nil)
}
```

# Running It

Next download the Sparta dependencies via:

- `go get ./...`

in the directory that you saved _main.go_. Once the packages are downloaded, first get a view of what's going on by the `describe` command (replacing `$S3_BUCKET` with an S3 bucket you own):

```nohighlight
$ go run main.go --level info describe --out ./graph.html --s3Bucket $S3_BUCKET
INFO[0000] ════════════════════════════════════════════════
INFO[0000] ╔═╗╔═╗╔═╗╦═╗╔╦╗╔═╗   Version : 1.13.0
INFO[0000] ╚═╗╠═╝╠═╣╠╦╝ ║ ╠═╣   SHA     : 03cdb90
INFO[0000] ╚═╝╩  ╩ ╩╩╚═ ╩ ╩ ╩   Go      : go1.13.3
INFO[0000] ════════════════════════════════════════════════
INFO[0000] Service: MyHelloWorldStack-123412341234       LinkFlags= Option=describe UTC="2019-12-07T20:01:48Z"
INFO[0000] ════════════════════════════════════════════════
INFO[0000] Provisioning service                          BuildID=none CodePipelineTrigger= InPlaceUpdates=false NOOP=true Tags=
INFO[0000] Verifying IAM Lambda execution roles
INFO[0000] IAM roles verified                            Count=1
INFO[0000] Skipping S3 preconditions check due to -n/-noop flag  Bucket=weagle Region=us-west-2 VersioningEnabled=false
INFO[0000] Running `go generate`
INFO[0000] Compiling binary                              Name=Sparta.lambda.amd64
INFO[0001] Creating code ZIP archive for upload          TempName=./.sparta/MyHelloWorldStack_123412341234-code.zip
INFO[0001] Lambda code archive size                      Size="24 MB"
INFO[0001] Skipping S3 upload due to -n/-noop flag       Bucket=weagle File=MyHelloWorldStack_123412341234-code.zip Key=MyHelloWorldStack-123412341234/MyHelloWorldStack_123412341234-code-ec0d6f8bae7b6a7abaa77db394c96265e213d20d.zip Size="24 MB"
INFO[0001] Skipping Stack creation due to -n/-noop flag  Bucket=weagle TemplateName=MyHelloWorldStack_123412341234-cftemplate.json
INFO[0001] ════════════════════════════════════════════════
INFO[0001] MyHelloWorldStack-123412341234 Summary
INFO[0001] ════════════════════════════════════════════════
INFO[0001] Verifying IAM roles                           Duration (s)=0
INFO[0001] Verifying AWS preconditions                   Duration (s)=0
INFO[0001] Creating code bundle                          Duration (s)=1
INFO[0001] Uploading code                                Duration (s)=0
INFO[0001] Ensuring CloudFormation stack                 Duration (s)=0
INFO[0001] Total elapsed time                            Duration (s)=1
```

Then open _graph.html_ in your browser (also linked [here](/images/overview/graph.html) ) to see what will be provisioned.

Since everything looks good, we'll provision the stack via `provision` and verify the lambda function. Note that the `$S3_BUCKET` value must be an S3 bucket to which you have write access since Sparta uploads the lambda package and CloudFormation template to that bucket as part of provisioning.

```nohighlight
INFO[0000] ════════════════════════════════════════════════
INFO[0000] ╔═╗╔═╗╔═╗╦═╗╔╦╗╔═╗   Version : 1.13.0
INFO[0000] ╚═╗╠═╝╠═╣╠╦╝ ║ ╠═╣   SHA     : 03cdb90
INFO[0000] ╚═╝╩  ╩ ╩╩╚═ ╩ ╩ ╩   Go      : go1.13.3
INFO[0000] ════════════════════════════════════════════════
INFO[0000] Service: MyHelloWorldStack-123412341234       LinkFlags= Option=provision UTC="2019-12-07T19:53:24Z"
INFO[0000] ════════════════════════════════════════════════
INFO[0000] Using `git` SHA for StampedBuildID            Command="git rev-parse HEAD" SHA=b114e329ed37b532e1f7d2e727aa8211d9d5889c
INFO[0000] Provisioning service                          BuildID=b114e329ed37b532e1f7d2e727aa8211d9d5889c CodePipelineTrigger= InPlaceUpdates=false NOOP=false Tags=
INFO[0000] Verifying IAM Lambda execution roles
INFO[0000] IAM roles verified                            Count=1
INFO[0000] Checking S3 versioning                        Bucket=weagle VersioningEnabled=true
INFO[0000] Checking S3 region                            Bucket=weagle Region=us-west-2
INFO[0000] Running `go generate`
INFO[0001] Compiling binary                              Name=Sparta.lambda.amd64
INFO[0002] Creating code ZIP archive for upload          TempName=./.sparta/MyHelloWorldStack_123412341234-code.zip
INFO[0002] Lambda code archive size                      Size="24 MB"
INFO[0002] Uploading local file to S3                    Bucket=weagle Key=MyHelloWorldStack-123412341234/MyHelloWorldStack_123412341234-code.zip Path=./.sparta/MyHelloWorldStack_123412341234-code.zip Size="24 MB"
INFO[0011] Uploading local file to S3                    Bucket=weagle Key=MyHelloWorldStack-123412341234/MyHelloWorldStack_123412341234-cftemplate.json Path=./.sparta/MyHelloWorldStack_123412341234-cftemplate.json Size="2.2 kB"
INFO[0011] Issued CreateChangeSet request                StackName=MyHelloWorldStack-123412341234
INFO[0016] Issued ExecuteChangeSet request               StackName=MyHelloWorldStack-123412341234
INFO[0033] CloudFormation Metrics ▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬
INFO[0033]     Operation duration                        Duration=8.26s Resource=MyHelloWorldStack-123412341234 Type="AWS::CloudFormation::Stack"
INFO[0033]     Operation duration                        Duration=1.35s Resource=HelloWorldLambda80576f7b21690b0cb485a6b69c927aac972cd693 Type="AWS::Lambda::Function"
INFO[0033] Stack provisioned                             CreationTime="2019-11-28 00:05:04.508 +0000 UTC" StackId="arn:aws:cloudformation:us-west-2:123412341234:stack/MyHelloWorldStack-123412341234/bab01fb0-1172-11ea-84a9-0ab88639bbc6" StackName=MyHelloWorldStack-123412341234
INFO[0033] ════════════════════════════════════════════════
INFO[0033] MyHelloWorldStack-123412341234 Summary
INFO[0033] ════════════════════════════════════════════════
INFO[0033] Verifying IAM roles                           Duration (s)=0
INFO[0033] Verifying AWS preconditions                   Duration (s)=0
INFO[0033] Creating code bundle                          Duration (s)=1
INFO[0033] Uploading code                                Duration (s)=9
INFO[0033] Ensuring CloudFormation stack                 Duration (s)=22
INFO[0033] Total elapsed time                            Duration (s)=33
```

Once the stack has been provisioned (`CREATE_COMPLETE`), login to the AWS console and navigate to the Lambda section.

# Testing

Find your Lambda function in the list of AWS Lambda functions and click the hyperlink. The display name will be prefixed by the name of your stack (_MyHelloWorldStack_ in our example):

![AWS Lambda List](/images/overview/AWS_Lambda_List.png)

On the Lambda details page, click the _Test_ button:

![AWS Lambda Test](/images/overview/AWS_Lambda_Test.png)

Accept the and name the _Hello World_ event template sample (our Lambda function doesn't consume the event data) and click _Save and test_. The execution result pane should display something similar to:

![AWS Lambda Execution](/images/overview/AWS_Lambda_Execution.png)

# Cleaning Up

To prevent unauthorized usage and potential charges, make sure to `delete` your stack before moving on:

```nohighlight
$ go run main.go delete

INFO[0000] ════════════════════════════════════════════════
INFO[0000] ╔═╗┌─┐┌─┐┬─┐┌┬┐┌─┐   Version : 1.0.2
INFO[0000] ╚═╗├─┘├─┤├┬┘ │ ├─┤   SHA     : b37b93e
INFO[0000] ╚═╝┴  ┴ ┴┴└─ ┴ ┴ ┴   Go      : go1.9.2
INFO[0000] ════════════════════════════════════════════════
INFO[0000] Service: MyHelloWorldStack                    LinkFlags= Option=delete UTC="2018-01-27T22:01:59Z"
INFO[0000] ════════════════════════════════════════════════
INFO[0000] Stack existence check                         Exists=true Name=MyHelloWorldStack
INFO[0000] Delete request submitted                      Response="{\n\n}"
```

# Conclusion

Congratulations! You've just deployed your first "serverless" service. The following sections will dive
deeper into what's going on under the hood as well as how to integrate your lambda function(s) into the broader AWS landscape.

# Next Steps

Walkthrough what Sparta actually does to deploy your application in the [next section](/reference/intro_example/step2/).
