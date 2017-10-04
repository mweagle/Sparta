---
date: 2016-03-09T19:56:50+01:00
title: Python/cgo Support
weight: 10
menu:
  main:
    parent: Documentation
    identifier: cgo
    weight: 100
---

## Introduction

The initial Sparta release supported running a normal Go binary [alongside](https://aws.amazon.com/blogs/compute/running-executables-in-aws-lambda/) a NodeJS HTTP-based proxy.

With Sparta [v0.11.0](https://github.com/mweagle/Sparta/blob/master/CHANGES.md#v0110) it's possible to transform your Sparta application into a CGO-library proxied by a Python 3.6 ctypes interface. This provides significant cold-start & hot exeecution performance improvements.

## Requirements

- Docker - Tested on OSX:

  ```
  $ docker -v
  Docker version 17.03.1-ce, build c6d412)
  ```
- Ability to build [CGO](https://blog.golang.org/c-go-cgo) libraries
- You *MUST* call `cgo.Main(...)` from your application's `main` package. The CGO functionality depends on being able to rewrite your application code into a CGO-compabible [source](https://golang.org/cmd/cgo/).  This enables Sparta to transform your application's `main()` into a library-equivalent `init()` function to initialize the internal function registry.

## Usage

To enable CGO packaging , choose either `Sparta/cgo.Main` or `Sparta/cgo.MainEx` functions. These functions are signature compatible with the existing `Main*` functions that produce in a NodeJS package.

## Example

### Before

{{< highlight go >}}
// File main.go
package main
import (
  // ================================================== //
	sparta "github.com/mweagle/Sparta"
  // ================================================== //

	"github.com/mweagle/SpartaPython"
)
// HelloWorld is a standard Sparta AWS λ function
func HelloWorld(event *json.RawMessage,
	context *sparta.LambdaContext,
	w http.ResponseWriter,
	logger *logrus.Logger) {
...
}

////////////////////////////////////////////////////////////////////////////////
// Main
func main() {
	lambdaFn, _ := sparta.Lambda(sparta.IAMRoleDefinition{},
		HelloWorld,
		nil)

	var lambdaFunctions []*sparta.LambdaAWSInfo
	lambdaFunctions = append(lambdaFunctions, lambdaFn)
  // ================================================== //
	sparta.Main("SpartaHelloNodeJSProcess",
		fmt.Sprintf("Test HelloWorld resource command"),
		lambdaFunctions,
		nil,
		nil)
  // ================================================== //

}
{{< /highlight >}}

### After

{{< highlight go >}}
// File main.go
package main
import (
  // ================================================== //
	spartaCGO "github.com/mweagle/Sparta/cgo"
  // ================================================== //
	"github.com/mweagle/SpartaPython"
)
// HelloWorld is a standard Sparta AWS λ function
func HelloWorld(event *json.RawMessage,
	context *sparta.LambdaContext,
	w http.ResponseWriter,
	logger *logrus.Logger) {
...
}

////////////////////////////////////////////////////////////////////////////////
// Main
func main() {
	lambdaFn, _ := sparta.Lambda(sparta.IAMRoleDefinition{},
		HelloWorld,
		nil)

	var lambdaFunctions []*sparta.LambdaAWSInfo
	lambdaFunctions = append(lambdaFunctions, lambdaFn)
  // ================================================== //
	spartaCGO.Main("SpartaHelloPythonCGO",
		fmt.Sprintf("Test HelloWorld resource command"),
		lambdaFunctions,
		nil,
		nil)
  // ================================================== //

}
{{< /highlight >}}


You should see log output that includes a statement similar to:

{{< highlight bash >}}
INFO[0000] Building `cgo` library in Docker              Args=[run --rm -v /Users/mweagle/Documents/gopath:/usr/src/gopath -w /usr/src/gopath/src/github.com/mweagle/SpartaPython/cgo -e GOPATH=/usr/src/gopath -e GOOS=linux -e GOARCH=amd64 golang:1.8.1 go build -o SpartaHelloPythonCGOUSEast.lambda.so -tags lambdabinary linux  -buildmode=c-shared -tags lambdabinary ] Name=SpartaHelloPythonCGOUSEast.lambda.so
{{< /highlight >}}

## Questions

### How do I initialize other AWS services?

The `cgo` variant of your Sparta application is proxied by a Python 3.6 handler. This handler provides access to the lambda credentials via:

```python
from botocore.credentials import get_credentials
```

Sparta makes these credentials available via `cgo.NewSession()` which returns a [*session.Session](http://docs.aws.amazon.com/sdk-for-go/api/aws/session/) instance. This value can be supplied to AWS service `New` functions as in:

{{< highlight go >}}
// Create a APIGateway client from just a session.
svc := apigateway.New(cgo.NewSession())
{{< /highlight >}}

### How does Sparta determine the Docker image for the CGO build?

Sparta parses the output from your host machine to determine the container tag.

```shell
$ go version
go version go1.8.1 darwin/amd64
```

### How can I pass environment variables to the Docker build

All host environment variables with a `SPARTA_` prefix will be passed via the `-e` flag to the Docker run command.

### How can I see what Sparta built?

Add a `--noop` command line argument to your `provision` command and examine the artifacts in the _/.sparta_ workspace directory. You can also enable debug logging via `--level debug` for more verbose runtime logging.

### How much of a performance increase should I expect?

Significant, particuarly at cold start times. In very limited testing I've seen times drop from 1500-2000ms to ~500ms.  See also https://twitter.com/mweagle/status/854178789814882304 which was a comparison through an API-GW exposed function exercised by a load testing service.

### What transformations are applied to my source?

See [cgo_main_run.go](https://github.com/mweagle/Sparta/blob/master/cgo/cgo_main_run.go#L44) for the set of changes applied to the input. If there is an error compiling the source, Sparta leaves the input file (with a `sparta-cgo` suffix) in the working directory for debugging.

