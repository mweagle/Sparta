---
date: 2018-01-22 21:49:38
title: CLI Options
weight: 15
alwaysopen: true
---

Sparta applications delegate `func main()` responsibilities to one of Sparta's Main entrypoints ([Main](https://godoc.org/github.com/mweagle/Sparta#Main), [MainEx](https://godoc.org/github.com/mweagle/Sparta#MainEx)). This provides each application with some standard command line options as shown below:

{{< highlight bash >}}
$ go run main.go --help
Simple Sparta application that demonstrates core functionality

Usage:
  main [command]

Available Commands:
  delete      Delete service
  describe    Describe service
  execute     Execute
  explore     Interactively explore service
  help        Help about any command
  profile     Interactively examine service pprof output
  provision   Provision service
  version     Display version information

Flags:
  -f, --format string    Log format [text, json] (default "text")
  -h, --help             help for main
      --ldflags string   Go linker string definition flags (https://golang.org/cmd/link/)
  -l, --level string     Log level [panic, fatal, error, warn, info, debug] (default "info")
  -n, --noop             Dry-run behavior only (do not perform mutations)
  -t, --tags string      Optional build tags for conditional compilation
  -z, --timestamps       Include UTC timestamp log line prefix

{{< /highlight >}}

It's also possible to add [custom flags](/reference/application/custom_flags) and/or [custom commands](/reference/application/custom_commands) to extend your application's behavior.

These command line options are briefly described in the following sections. For the most up to date information, use the `--help` subcommand option.


# Standard Commands

## Delete

This simply deletes the stack (if present). Attempting to delete a non-empty stack is not treated as an error.

## Describe

The `describe` command line option produces an HTML summary (see [graph.html](/images/overview/graph.html) for an example) of your Sparta service.

The report also includes the automatically generated CloudFormation template which can be helpful when diagnosing provisioning errors.

## Execute

This command is used when the cross compiled binary is provisioned in AWS lambda. It is not (typically) applicable to the local development workflow.

## Explore

The `explore` option creates a terminal GUI that supports interactive exploration of lambda functions deployed to AWS. This ui recursively searches for all _*.json_ files in the source tree to populate the set of eligible events that can be submitted.

![Explore](/images/explore.jpg "Explore")

## Profile

The `profile` command line option enters an interactive session where a previously profiled application can be locally visualized using snapshots posted to S3 and provided to a local [pprof ui](https://rakyll.org/pprof-ui/).

## Provision

The `provision` option is the subcommand most likely to be used during development.  It provisions the Sparta application to AWS Lambda.


## Version

The `version` option is a diagnostic command that prints the version of the Sparta framework embedded in the application.

{{< highlight bash >}}
$ go run main.go version
INFO[0000] ════════════════════════════════════════════════
INFO[0000] ╔═╗┌─┐┌─┐┬─┐┌┬┐┌─┐   Version : 1.2.1
INFO[0000] ╚═╗├─┘├─┤├┬┘ │ ├─┤   SHA     : b76b71d
INFO[0000] ╚═╝┴  ┴ ┴┴└─ ┴ ┴ ┴   Go      : go1.10
INFO[0000] ════════════════════════════════════════════════
INFO[0000] Service: MyHelloWorldStack-mweagle            LinkFlags= Option=version UTC="2018-07-20T04:44:17Z"
INFO[0000] ════════════════════════════════════════════════
{{< /highlight  >}}
