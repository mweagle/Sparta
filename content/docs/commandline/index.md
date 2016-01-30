+++
author = "Matt Weagle"
date = "2015-11-29T06:50:17"
title = "Command Line Options"
tags = ["sparta"]
type = "doc"
+++

Sparta provides a [Main](https://godoc.org/github.com/mweagle/Sparta#Main) function that transforms a set of [lambda functions](https://godoc.org/github.com/mweagle/Sparta#LambdaAWSInfo) into an application.  This function should be called from your application's `package main` as in:


{{< highlight go >}}
sparta.Main("MyStack",
  "Simple Sparta application",
  myLambdaFunctions,
  nil,
  nil)
{{< /highlight >}}


The application provides several command line options which are available by providing the `-h/--help` option as in:

{{< highlight nohighlight >}}
go run application.go --help
Usage: application [global options] <verb> [verb options]

Global options:
        -n, --noop     Dry-run behavior only (do not provision stack)
        -l, --level    Log level [panic, fatal, error, warn, info, debug] (default: info)
        -h, --help     Show this help

Verbs:
    delete:
    describe:
        -o, --out      Output file for HTML description (*)
    execute:
        -p, --port     Alternative port for HTTP binding (default=9999)
        -s, --signal   Process ID to signal with SIGUSR2 once ready
    explore:
    provision:
        -b, --s3Bucket S3 Bucket to use for Lambda source (*)
{{< /highlight >}}


### <a href="{{< relref "#delete" >}}">Delete</a>

This simply deletes the stack (if present). Attempting to delete a non-empty stack is not treated as an error.

### <a href="{{< relref "#describe" >}}">Describe</a>

The `describe` command line option produces an HTML summary (see [graph.html](/images/overview/graph.html) for an example) of your Sparta service.  

The report also includes the automatically generated CloudFormation template which can be helpful when diagnosing provisioning errors.

### <a href="{{< relref "#execute" >}}">Execute</a>

The `execute` option is typically used when the compiled application is launched in the AWS Lambda environment.  It starts up an HTTP listener to which the NodeJS proxing tier forwards requests.

### <a href="{{< relref "#explore" >}}">Explore</a>

The `explore` option creates a _localhost_ server to allow Sparta lambda functions to be tested locally.  

NOTE: API Gateway mapping templates are not currently supported.

### <a href="{{< relref "#provision" >}}">Provision</a>

The `provision` option is the verb most likely to be used during development.  It provisions the Sparta application to AWS Lambda.
