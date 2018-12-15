---
date: 2018-01-22 21:49:38
title: Supporting Packages
weight: 900
alwaysopen: false
---

The following packages are part of the Sparta ecosystem and can be used in combination
or as standalone in other applications.

## go-cloudcondensor

The [go-cloudcondensor](https://github.com/mweagle/go-cloudcondenser) package provides
utilities to express CloudFormation templates as a set of `go` functions. Templates
are evaluated and the and the resulting JSON can be integrated into existing
CLI-based workflows.

## go-cloudformation

The [go-cloudformation](https://github.com/mweagle/go-cloudformation) package provides a Go object
model for the official CloudFormation
[JSON Schema](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/cfn-resource-specification.html).

## SpartaVault

[SpartaVault](https://github.com/mweagle/SpartaVault) uses KMS to encrypt values into Go types that can be safely
committed to source control. It includes a command line utility that produces an encrypted
set of credentials that are statically compiled into your application.

## ssm-cache

The [ssm-cache](https://github.com/mweagle/ssm-cache) package provides an expiring cache for
[AWS Systems Manager](https://aws.amazon.com/systems-manager/).
SSM is the preferred service to use for sharing credentials with your service.

## Examples

There are also many Sparta [example repos](https://github.com/mweagle?utf8=%E2%9C%93&tab=repositories&q=Sparta&type=&language=) that demonstrate core concepts.