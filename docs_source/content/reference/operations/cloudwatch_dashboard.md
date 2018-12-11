---
date: 2018-01-22 21:49:38
title: CloudWatch Dashboard
weight: 10
alwaysopen: false
---

The [DashboardDecorator](https://godoc.org/github.com/mweagle/Sparta/decorator#DashboardDecorator) creates a CloudWatch Dashboard that produces a single [CloudWatch Dashboard](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/CloudWatch_Dashboards.html) to summarize your stack's behavior.

Sample usage:

```go
func workflowHooks(connections *service.Connections,
  lambdaFunctions []*sparta.LambdaAWSInfo,
  websiteURL *gocf.StringExpr) *sparta.WorkflowHooks {
  // Setup the DashboardDecorator lambda hook
  workflowHooks := &sparta.WorkflowHooks{
    ServiceDecorators: []sparta.ServiceDecoratorHookHandler{
      spartaDecorators.DashboardDecorator(lambdaFunctions, 60),
      serviceResourceDecorator(connections, websiteURL),
    },
  }
  return workflowHooks
}
```

A sample dashboard for the [SpartaGeekwire](https://github.com/mweagle/SpartaGeekwire) project is:

![Sparta](/images/dashboard/CloudWatch_Management_Console.jpg "CloudWatch Dashboard")

Related to this, see the recently announced [AWS Lambda Application Dashboard](https://aws.amazon.com/about-aws/whats-new/2018/08/aws-lambda-console-enables-managing-and-monitoring/).