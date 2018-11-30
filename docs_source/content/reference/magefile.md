---
date: 2018-11-28 16:17:43
title: Magefiles
weight: 150
---

{{% notice info %}}
TODO: Document [magefile](https://github.com/mweagle/SpartaHelloWorld/blob/master/magefile.go) and associated
[tasks](https://godoc.org/github.com/mweagle/Sparta/magefile)
{{% /notice %}}

## Default Sparta magefile

This _magefile.go_ can be used, unchanged, for most standard Sparta projects.

```go
// +build mage

package main

import (
  spartaMage "github.com/mweagle/Sparta/magefile"
)

// Provision the service
func Provision() error {
  return spartaMage.Provision()
}

// Describe the stack by producing an HTML representation of the CloudFormation
// template
func Describe() error {
  return spartaMage.Describe()
}

// Delete the service, iff it exists
func Delete() error {
  return spartaMage.Delete()
}

// Status report if the stack has been provisioned
func Status() error {
  return spartaMage.Status()
}

// Version information
func Version() error {
  return spartaMage.Version()
}
```