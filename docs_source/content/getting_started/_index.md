---
date: 2019-11-27 16:12:19
title: Getting Started
description: Getting started with Sparta
weight: 5
alwaysopen: false
---

To build a Sparta application, follow these steps:

1. `go get -u -v https://github.com/mweagle/Sparta/...`
2. Configure your AWS Credentials according to the [go SDK docs](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#specifying-credentials). The most reliable approach is to use environment variables as in:

   ```shell
   $ env | grep AWS
   AWS_DEFAULT_REGION=us-xxxx-x
   AWS_SECRET_ACCESS_KEY=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
   AWS_REGION=us-xxxx-x
   AWS_ACCESS_KEY_ID=xxxxxxxxxxxxxxxxxxxx
   ```

3. Create a sample _main.go_ file as in:

   ```go
   package main

   import (
       "context"
       "fmt"
       "os"

       "github.com/aws/aws-sdk-go/aws/session"
       sparta "github.com/mweagle/Sparta"
       spartaCF "github.com/mweagle/Sparta/aws/cloudformation"
       "github.com/sirupsen/logrus"
   )

   // Standard AWS λ function

   func helloWorld(ctx context.Context) (string, error) {
       logger, loggerOk := ctx.Value(sparta.ContextKeyLogger).(*logrus.Logger)
       if loggerOk {
           logger.Info("Accessing structured logger 🙌")
       }
       return "Hello World 👋. Welcome to AWS Lambda! 🙌🎉🍾", nil
   }

   ////////////////////////////////////////////////////////////////////////////////
   // Main
   func main() {
       lambdaFn, _ := sparta.NewAWSLambda("Hello World",
           helloWorld,
           sparta.IAMRoleDefinition{})

       sess := session.Must(session.NewSession())
       awsName, awsNameErr := spartaCF.UserAccountScopedStackName("MyHelloWorldStack",
           sess)
       if awsNameErr != nil {
           fmt.Print("Failed to create stack name\n")
           os.Exit(1)
       }
       var lambdaFunctions []*sparta.LambdaAWSInfo
       lambdaFunctions = append(lambdaFunctions, lambdaFn)

       err := sparta.Main(awsName,
           "Simple Sparta HelloWorld application",
           lambdaFunctions,
           nil,
           nil)
       if err != nil {
           os.Exit(1)
       }
   }
   ```

4. Build with `go run main.go provision --s3Bucket YOUR_S3_BUCKET_NAME` where `YOUR_S3_BUCKET_NAME` is an S3 bucket to which your account has write privileges.

The following [Example Service](/example_service) section provides more details regarding how Sparta transforms your application into a self-deploying service.
