---
date: 2016-03-09T19:56:50+01:00
title: Custom Commands
weight: 10
---

In addition to [custom flags](/reference/application/custom_flags), an application may register completely new commands. For example, to support [alternative topologies](/reference/application/custom_flags) or integrated automated acceptance tests as part of a CI/CD pipeline.

To register a custom command, define a new [cobra.Command](https://github.com/spf13/cobra) and add it to the `sparta.CommandLineOptions.Root` command value.  Ensure you use the `xxxxE` Cobra functions so that errors can be properly propagated.

```go
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
```

Registering a user-defined command makes that command's usage information seamlessly integrate with the standard commands:

```bash
$ go run main.go --help
Provision AWS Lambda and EC2 instance with same code

Usage:
  main [command]

Available Commands:
  delete      Delete service
  describe    Describe service
  execute     Start the application and begin handling events
  explore     Interactively explore a provisioned service
  help        Help about any command
  httpServer  Sample HelloWorld HTTP server
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
```

And you can query for user-command specific usage as in:

```bash

$ ./SpartaOmega httpServer --help
Custom command

Usage:
  SpartaOmega httpServer [flags]

Global Flags:
  -l, --level string   Log level [panic, fatal, error, warn, info, debug] (default "info")
  -n, --noop           Dry-run behavior only (do not perform mutations)

```
