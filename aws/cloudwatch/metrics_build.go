//go:build !lambdabinary
// +build !lambdabinary

package cloudwatch

// RegisterLambdaUtilizationMetricPublisher installs a periodic task
// to publish the current system metrics to CloudWatch Metrics.
func RegisterLambdaUtilizationMetricPublisher(customDimensionMap map[string]string) {
	// NOP
}
