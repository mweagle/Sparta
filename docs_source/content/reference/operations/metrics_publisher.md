
---
date: 2018-12-01 06:02:32
title: Metrics Publisher
weight: 30
---

AWS Lambda is tightly integrated with other AWS services and provides excellent
opportunities for improving your service's observability posture. Sparta includes
a [CloudWatch Metrics](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/working_with_metrics.html)
publisher that periodically publishes metrics to CloudWatch.

This periodic task publishes environment-level metrics that have been
detected by the [gopsutil](https://github.com/shirou/gopsutil) package. Metrics include:

- CPU
  - Percent used
- Disk
  - Percent used
- Host
  - Uptime (milliseconds)
- Load
  - Load1 (no units)
  - Load5 (no units)
  - Load15 (no units)
- Network
  - NetBytesSent (bytes)
  - NetBytesRecv (bytes)
  - NetErrin (count)
  - NetErrout (count)

You can provide an optional `map[string]string` set of dimensions to which the metrics
should be published. This enables targeted [alert conditions](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/Create-alarm-on-metric-math-expression.html)
that can be used to improve system resiliency.

To register the metric publisher, call the `RegisterLambdaUtilizationMetricPublisher` at some
point in your `main()` call graph. For example:

```go
import spartaCloudWatch "github.com/mweagle/Sparta/aws/cloudwatch"
func main() {
  ...
  spartaCloudWatch.RegisterLambdaUtilizationMetricPublisher(map[string]string{
    "BuildId":    sparta.StampedBuildID,
  })
  ...
}
```
- The optional `map[string]string` parameter is the custom Name-Value pairs to use as a [CloudWatch Dimension](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/cloudwatch_concepts.html#Dimension)
```

{{% notice note %}}
TODO: Document the [RegisterLambdaUtilizationMetricPublisher](https://godoc.org/github.com/mweagle/Sparta/aws/cloudwatch#RegisterLambdaUtilizationMetricPublisher) utility function.
{{% /notice %}}