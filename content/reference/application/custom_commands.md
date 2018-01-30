---
date: 2016-03-09T19:56:50+01:00
title: Custom Application Commands
weight: 10
---

# Introduction

In addition to [custom flags](/reference/application/custom_flags), an application may register completely new commands. For example, to support [alternative topologies](/reference/application/custom_flags) or integrated automated acceptance tests as part of a CI/CD pipeline.

To register a custom command, define a new [cobra.Command](https://github.com/spf13/cobra) and add it to the `sparta.CommandLineOptions.Root` command value.  Ensure you use the `xxxxE` Cobra functions so that errors can be properly propagated.

{{< highlight go >}}
httpServerCommand := &cobra.Command{
  Use:   "httpServer",
  Short: "Sample HelloWorld HTTP server",
  Long:  `Sample HelloWorld HTTP server that binds to port: ` + HTTPServerPort,
  RunE: func(cmd *cobra.Command, args []string) error {
    http.HandleFunc("/", helloWorldResource)
    return http.ListenAndServe(fmt.Sprintf(":%d", HTTPServerPort), nil)
  },
}
sparta.CommandLineOptions.Root.AddCommand(httpServerCommand)
{{< /highlight >}}

Registering a user-defined command makes that command's usage information seamlessly integrate with the standard commands:

{{< highlight bash >}}
$ ./SpartaOmega --help
Provision AWS Lambda and EC2 instance with same code

Usage:
  SpartaOmega [command]

Available Commands:
  httpServer  Sample HelloWorld HTTP server
  version     Sparta framework version
  provision   Provision service
  delete      Delete service
  execute     Execute
  describe    Describe service
  explore     Interactively explore service

Flags:
  -l, --level string   Log level [panic, fatal, error, warn, info, debug] (default "info")
  -n, --noop           Dry-run behavior only (do not perform mutations)

Use "SpartaOmega [command] --help" for more information about a command.

{{< /highlight >}}

And you can query for user-command specific usage as in:

{{< highlight bash >}}

$ ./SpartaOmega httpServer --help
Custom command

Usage:
  SpartaOmega httpServer [flags]

Global Flags:
  -l, --level string   Log level [panic, fatal, error, warn, info, debug] (default "info")
  -n, --noop           Dry-run behavior only (do not perform mutations)

{{< /highlight >}}
