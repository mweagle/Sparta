---
date: 2016-03-09T19:56:50+01:00
title: CGO
weight: 10
menu:
  main:
    parent: Documentation
    identifier: cgo
    weight: 0
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
- You *MUST* call `cgo.Main()` from your application's `main` package. The CGO step temporarily rewrites your application code into a CGO-compliant source file, and proper Sparta initialization depends on being able to transform your application's `main()` into a library-equivalent `init()` function.

## Usage

To enable CGO packaging , choose either `Sparta/cgo.Main` or `Sparta/cgo.MainEx` functions. These functions are signature compatible with the existing `Main*` functions that produce in a NodeJS package.

## Example

### Before

```golang
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
```

### After

```golang
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
```

