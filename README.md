
<div align="center"><img src="https://raw.githubusercontent.com/mweagle/Sparta/master/site/SpartaLogoLarge.png" />
</div>

# Sparta <p align="center">

[![Build Status](https://travis-ci.org/mweagle/Sparta.svg?branch=master)](https://travis-ci.org/mweagle/Sparta)

[![GoDoc](https://godoc.org/github.com/mweagle/Sparta?status.svg)](https://godoc.org/github.com/mweagle/Sparta)

[![Go Report Card](https://goreportcard.com/badge/github.com/mweagle/Sparta)](https://goreportcard.com/report/github.com/mweagle/Sparta)

Visit [gosparta.io](https://gosparta.io) for complete documentation.

## Overview

Sparta takes a set of _golang_ functions and automatically provisions them in
[AWS Lambda](https://aws.amazon.com/lambda/) as a logical unit.

AWS Lambda functions are defined using the standard [AWS Lambda signatures](https://aws.amazon.com/blogs/compute/announcing-go-support-for-aws-lambda/):

 * `func()`
 * `func() error`
 * `func(TIn) error`
 * `func() (TOut, error)`
 * `func(context.Context) error`
 * `func(context.Context, TIn) error`
 * `func(context.Context) (TOut, error)`
 * `func(context.Context, TIn) (TOut, error)`

 The TIn and TOut parameters represent encoding/json un/marshallable types.

For instance:

```go
// Standard AWS λ function
func helloWorld(ctx context.Context) (string, error) {
  ...
}
```

where
  * `ctx` : The request context that includes Sparta both the [AWS Context](https://github.com/aws/aws-lambda-go/blob/master/lambdacontext/context.go) as well as Sparta specific [values](https://godoc.org/github.com/mweagle/Sparta#pkg-constants.)


Consumers define a set of lambda functions and provide them to Sparta to create a self-documenting, self-deploying AWS Lambda binary:

```go
	lambdaFn := sparta.HandleAWSLambda("Hello World",
		helloWorld,
		sparta.IAMRoleDefinition{})

	var lambdaFunctions []*sparta.LambdaAWSInfo
	lambdaFunctions = append(lambdaFunctions, lambdaFn)

	err := sparta.Main("HelloWorldStack",
		"My Hello World stack",
		lambdaFunctions,
		nil,
		nil)
```

Given a set of registered _golang_ functions, Sparta will:

  * Either verify or provision the defined [IAM roles](http://docs.aws.amazon.com/lambda/latest/dg/intro-permission-model.html)
  * Build a deployable application via `Provision()`
  * Zip the contents and associated proxying logic
  * Dynamically create a CloudFormation template to either create or update the service state.
  * Optionally:
    * Register with S3 and SNS for push source configuration
    * Provision an [API Gateway](https://aws.amazon.com/api-gateway/) service to make your functions publicly available
    * Provision an [S3 static website](http://docs.aws.amazon.com/AmazonS3/latest/dev/WebsiteHosting.html)

Visit [gosparta.io](https://gosparta.io) for complete documentation.

## Contributors

_Thanks to all Sparta contributors (alphabetical)_

  - **Kyle Anderson**
  - [James Brook](https://github.com/jbrook)
  - [Ryan Brown](https://github.com/ryansb)
  - [sdbeard](https://github.com/sdbeard)
  - [Scott Raine](https://github.com/nylar)
  - [Paul Seiffert](https://github.com/seiffert)
  - [Thom Shutt](https://github.com/thomshutt)

