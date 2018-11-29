---
date: 2018-01-22 21:49:38
title: S3 Artifact Publisher
weight: 10
alwaysopen: false
---

The [S3ArtifactPublisherDecorator](https://godoc.org/github.com/mweagle/Sparta/decorator#S3ArtifactPublisherDecorator)
enables a service to publish objects to S3 locations as part of the service lifecycle.

This decorator is implemented as a [ServiceDecoratorHookHandler](https://godoc.org/github.com/mweagle/Sparta#ServiceDecoratorHookHandler) which
is supplied to [MainEx](https://godoc.org/github.com/mweagle/Sparta#MainEx). For example:

```go

hooks := &sparta.WorkflowHooks{}
payloadData := map[string]interface{}{
  "SomeValue": gocf.Ref("AWS::StackName"),
}

serviceHook := spartaDecorators.S3ArtifactPublisherDecorator(gocf.String("MY-S3-BUCKETNAME"),
  gocf.Join("",
    gocf.String("metadata/"),
    gocf.Ref("AWS::StackName"),
    gocf.String(".json")),
  payloadData)
hooks.ServiceDecorators = []sparta.ServiceDecoratorHookHandler{serviceHook}
```

