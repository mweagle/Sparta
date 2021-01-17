---
date: 2018-12-01 06:18:10
title:
pre: "<b>Build-Time Decorators</b>"
chapter: false
weight: 120
---

# Build-Time Decorators

Sparta uses build-time [decorators](https://godoc.org/github.com/mweagle/Sparta/decorator) to annotate the CloudFormation template with
additional functionality.

While Sparta tries to provide workflows common across service lifecycles, it may be the case that an application requires additional functionality or runtime resources.

To support this, Sparta allows you to customize the build pipeline via [WorkflowHooks](https://godoc.org/github.com/mweagle/Sparta#WorkflowHooks) structure. These hooks are called at specific points in the _provision_ lifecycle and support augmenting the standard pipeline:

{{< spartaflow >}}

The following sections describe the types of WorkflowHooks available. All hooks accept a `context map[string]interface{}` as their first parameter. Sparta treats this as an opaque property bag that enables hooks to communicate state.

## WorkflowHook Types

### Builder Hooks

BuilderHooks share the [WorkflowHook](https://godoc.org/github.com/mweagle/Sparta#WorkflowHook) signature:

```go
type WorkflowHook func(ctx context.Context,
  serviceName string,
  S3Bucket gocf.Stringable,
  buildID string,
  awsSession *session.Session,
  noop bool,
  logger *zerolog.Logger) (context.Context, error)
```

These functions include:

- PreBuild
- PostBuild
- PreMarshall
- PostMarshall

### Archive Hook

The `ArchiveHook` allows a service to add custom resources to the ZIP archive and have the signature:

```go
type ArchiveHook func(ctx context.Context,
  serviceName string,
  zipWriter *zip.Writer,
  awsSession *session.Session,
  noop bool,
  logger *zerolog.Logger) (context.Context, error)
```

This function is called _after_ Sparta has written the standard resources to the `*zip.Writer` stream.

### Rollback Hook

The `RollbackHook` is called _iff_ the _provision_ operation fails and has the signature:

```go

type RollbackHook func(context map[string]interface{},
  serviceName string,
  awsSession *session.Session,
  noop bool,
  logger *zerolog.Logger)
```

## Using WorkflowHooks

To use the Workflow Hooks feature, initialize a [WorkflowHooks](https://godoc.org/github.com/mweagle/Sparta#WorkflowHooks) structure with 1 or more hook functions and call [sparta.MainEx](https://godoc.org/github.com/mweagle/Sparta#MainEx).

## Available Decorators

{{% children description="true"   %}}

### Notes

- Workflow hooks can be used to support [Dockerizing](https://github.com/mweagle/SpartaDocker) your application
- You may need to add [custom CLI commands](/reference/application/custom_commands) to fully support Docker
- Enable `--level debug` for detailed workflow hook debugging information
