+++
author = "Matt Weagle"
date = "2015-11-29T06:50:17"
title = "Command Line Options"
tags = ["sparta"]
type = "doc"
+++

Sparta provides a [Main](https://godoc.org/github.com/mweagle/Sparta#Main) function that transforms a set of [lambda functions](https://godoc.org/github.com/mweagle/Sparta#LambdaAWSInfo) into an application.  This function should be called from your application's `package main` as in:


{{< highlight go >}}
var lambdaFunctions []*sparta.LambdaAWSInfo
lambdaFunctions = append(lambdaFunctions, lambdaFn)
err := sparta.Main("SpartaHelloWorld",
  fmt.Sprintf("Test HelloWorld resource command"),
  lambdaFunctions,
  nil,
  nil)
{{< /highlight >}}

A compiled application provides several command line options which are available by providing the `-h/--help` option as in:

{{< highlight nohighlight >}}
$ ./SpartaHelloWorld --help

Test HelloWorld resource command

Usage:
  SpartaHelloWorld [command]

Available Commands:
  version     Sparta framework version
  provision   Provision service
  delete      Delete service
  execute     Execute
  describe    Describe service
  explore     Interactively explore service

Flags:
  -l, --level string   Log level [panic, fatal, error, warn, info, debug]' (default "info")
  -n, --noop           Dry-run behavior only (do not perform mutations)

Use "SpartaHelloWorld [command] --help" for more information about a command.
{{< /highlight >}}

It's also possible to add [custom flags](/docs/application/custom_flags) and/or [custom commands](/docs/application/custom_commands) to extend your application's behavior.

# Provision

The `provision` option is the verb most likely to be used during development.  It provisions the Sparta application to AWS Lambda.

# Delete

This simply deletes the stack (if present). Attempting to delete a non-empty stack is not treated as an error.

# Describe

The `describe` command line option produces an HTML summary (see [graph.html](/images/overview/graph.html) for an example) of your Sparta service.  

The report also includes the automatically generated CloudFormation template which can be helpful when diagnosing provisioning errors.

# Execute

The `execute` option is typically used when the compiled application is launched in the AWS Lambda environment.  It starts up an HTTP listener to which the NodeJS proxing tier forwards requests.

# Explore

The `explore` option creates a _localhost_ server to allow Sparta lambda functions to be tested locally.  

NOTE: API Gateway mapping templates are not currently supported.


# Version

The `version` option is a diagnostic command that prints the version of the Sparta framework embedded in the application.

