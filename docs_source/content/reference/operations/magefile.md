---
date: 2018-11-28 16:17:43
title: Magefiles
weight: 5
---

To support cross platform development and usage, Sparta uses [magefiles](https://magefile.org) rather
than _Makefiles_.  Most projects can start with the _magefile.go_ sample below. The Magefiles
provide a discoverable CLI, but are entirely optional. `go run main.go XXXX` style invocation remains
available as well.

## Default Sparta magefile.go

This _magefile.go_ can be used, unchanged, for most standard Sparta projects.

```go
// +build mage

// File: magefile.go

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

```shell
$ mage -l
Targets:
  delete       the service, iff it exists
  describe     the stack by producing an HTML representation of the CloudFormation template
  provision    the service
  status       report if the stack has been provisioned
  version      information
```

## Sparta Magefile Helpers

There are several [magefile helpers](https://godoc.org/github.com/mweagle/Sparta/magefile) available
in the Sparta package. These are in addition to and often delegate to, the core
[mage libraries](https://magefile.org/libraries/).