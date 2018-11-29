---
date: 2018-01-22 21:49:38
title: CLI Options
weight: 15
alwaysopen: false
---

Sparta applications delegate `func main()` responsibilities to one of Sparta's Main entrypoints ([Main](https://godoc.org/github.com/mweagle/Sparta#Main), [MainEx](https://godoc.org/github.com/mweagle/Sparta#MainEx)). This provides each application with some standard command line options as shown below:

```bash
$ go run main.go --help
Simple Sparta application that demonstrates core functionality

Usage:
  main [command]

Available Commands:
  delete      Delete service
  describe    Describe service
  execute     Start the application and begin handling events
  explore     Interactively explore a provisioned service
  help        Help about any command
  profile     Interactively examine service pprof output
  provision   Provision service
  status      Produce a report for a provisioned service
  version     Display version information

Flags:
  -f, --format string    Log format [text, json] (default "text")
  -h, --help             help for main
      --ldflags string   Go linker string definition flags (https://golang.org/cmd/link/)
  -l, --level string     Log level [panic, fatal, error, warn, info, debug] (default "info")
      --nocolor          Boolean flag to suppress colorized TTY output
  -n, --noop             Dry-run behavior only (do not perform mutations)
  -t, --tags string      Optional build tags for conditional compilation
  -z, --timestamps       Include UTC timestamp log line prefix

Use "main [command] --help" for more information about a command.

```

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

## Status

The `status` option queries AWS for the current stack status
and produces an optionally account-id redacted report. Stack
outputs, tags, and other metadata are included in the status report:

```bash
$ go run main.go status --redact
INFO[0000] ════════════════════════════════════════════════
INFO[0000] ╔═╗╔═╗╔═╗╦═╗╔╦╗╔═╗   Version : 1.5.0
INFO[0000] ╚═╗╠═╝╠═╣╠╦╝ ║ ╠═╣   SHA     : 8f199e1
INFO[0000] ╚═╝╩  ╩ ╩╩╚═ ╩ ╩ ╩   Go      : go1.11.1
INFO[0000] ════════════════════════════════════════════════
INFO[0000] Service: MyHelloWorldStack-mweagle            LinkFlags= Option=status UTC="2018-10-14T12:28:18Z"
INFO[0000] ════════════════════════════════════════════════
INFO[0001] StackId                                       Id="arn:aws:cloudformation:us-west-2:************:stack/MyHelloWorldStack-mweagle/5817dff0-c5f1-11e8-b43a-503ac9841a99"
INFO[0001] Stack status                                  State=UPDATE_COMPLETE
INFO[0001] Created                                       Time="2018-10-02 03:14:59.127 +0000 UTC"
INFO[0001] Last Update                                   Time="2018-10-06 14:20:40.267 +0000 UTC"
INFO[0001] Tag                                           io:gosparta:buildId=7ee3e1bc52f15c4a636e05061eaec7b748db22a9
```

## Version

The `version` option is a diagnostic command that prints the version of the Sparta framework embedded in the application.

```bash
$ go run main.go version
INFO[0000] ════════════════════════════════════════════════
INFO[0000] ╔═╗╔═╗╔═╗╦═╗╔╦╗╔═╗   Version : 1.5.0
INFO[0000] ╚═╗╠═╝╠═╣╠╦╝ ║ ╠═╣   SHA     : 8f199e1
INFO[0000] ╚═╝╩  ╩ ╩╩╚═ ╩ ╩ ╩   Go      : go1.11.1
INFO[0000] ════════════════════════════════════════════════
INFO[0000] Service: MyHelloWorldStack-mweagle            LinkFlags= Option=version UTC="2018-10-14T12:27:36Z"
INFO[0000] ════════════════════════════════════════════════
```
