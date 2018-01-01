
<div align="center"><img src="https://raw.githubusercontent.com/mweagle/Sparta/master/site/SpartaLogoLarge.png" />
</div>

# Sparta <p align="center">

[![Build Status](https://travis-ci.org/mweagle/Sparta.svg?branch=master)](https://travis-ci.org/mweagle/Sparta) [![GoDoc](https://godoc.org/github.com/mweagle/Sparta?status.svg)](https://godoc.org/github.com/mweagle/Sparta) [![Sourcegraph](https://sourcegraph.com/github.com/mweagle/Sparta/-/badge.svg)](https://sourcegraph.com/github.com/mweagle/Sparta?badge)[![Go Report Card](https://goreportcard.com/badge/github.com/mweagle/Sparta)](https://goreportcard.com/report/github.com/mweagle/Sparta)

Visit [gosparta.io](http://gosparta.io) for complete documentation.

## Overview

Sparta takes a set of _golang_ functions and automatically provisions them in
[AWS Lambda](https://aws.amazon.com/lambda/) as a logical unit.

AWS lambda functions are defined as standard [HandlerFunc.ServeHTTP](https://golang.org/pkg/net/http/#HandlerFunc.ServeHTTP) functions as in

```go
type myHelloWorldFunction func(w http.ResponseWriter, r *http.Request) {
  ...
}
```

where
  * `w` : The ResponseWriter used to return the lambda response
  * `r` :  The arbitrary event data provided to the function.

The `http.Request` instance also includes [context](https://golang.org/pkg/net/http/#Request.Context) scoped values as in:
  * `*logrus.Logger` : A request scoped logger
    * `_ := r.Context().Value(sparta.ContextKeyLogger).(*logrus.Logger)`
  * [*sparta.LambdaContext](https://godoc.org/github.com/mweagle/Sparta#LambdaContext) : AWS lambda context parameters
    * `	lambdaContext, _ := r.Context().Value(sparta.ContextKeyLambdaContext).(*sparta.LambdaContext)`

Consumers define a set of lambda functions and provide them to Sparta to create a self-documentating, self-deploying AWS Lambda binary:

```go
  lambdaFn := sparta.HandleAWSLambda("Hello World",
    http.HandlerFunc(myHelloWorldFunction),
    sparta.IAMRoleDefinition{})

  var lambdaFunctions []*sparta.LambdaAWSInfo
  lambdaFunctions = append(lambdaFunctions, lambdaFn)
  err := sparta.Main("MyHelloWorldStack",
    "Simple Sparta application that demonstrates core functionality",
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

Visit [gosparta.io](http://gosparta.io) for complete documentation.

## Limitations

See the [Limitations](http://gosparta.io/docs/limitations/) page for the most up-to-date information.

## Contributors

_Thanks to all Sparta contributors (alphabetical)_

  - **Kyle Anderson**
  - [James Brook](https://github.com/jbrook)
  - [Ryan Brown](https://github.com/ryansb)
  - [sdbeard](https://github.com/sdbeard)
  - [Scott Raine](https://github.com/nylar)
  - [Paul Seiffert](https://github.com/seiffert)
  - [Thom Shutt](https://github.com/thomshutt)

