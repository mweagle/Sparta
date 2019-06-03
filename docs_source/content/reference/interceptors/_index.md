---
date: 2016-03-09T19:56:50+01:00
title:
pre: "<b>Runtime Interceptors</b>"
weight: 130
alwaysopen: false
---

# Runtime Interceptors

Sparta uses runtime [interceptors](https://godoc.org/github.com/mweagle/Sparta#LambdaEventInterceptors) to hook into
the event handling workflow. Interceptors provide an opportunity to handle concerns (logging, metrics, etc) independent
of core event handling workflow.

{{< interceptorflow >}}

## Available Interceptors

{{% children description="true"   %}}
