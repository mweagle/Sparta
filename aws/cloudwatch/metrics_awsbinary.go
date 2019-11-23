// +build lambdabinary

package cloudwatch

import (
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	awsCloudWatch "github.com/aws/aws-sdk-go/service/cloudwatch"
	sparta "github.com/mweagle/Sparta"
	spartaAWS "github.com/mweagle/Sparta/aws"
	gopsutilCPU "github.com/shirou/gopsutil/cpu"
	gopsutilDisk "github.com/shirou/gopsutil/disk"
	gopsutilHost "github.com/shirou/gopsutil/host"
	gopsutilLoad "github.com/shirou/gopsutil/load"
	gopsutilNet "github.com/shirou/gopsutil/net"
	"github.com/sirupsen/logrus"
)

// publishMetrics is the actual metric publishing logic. T
func publishMetrics(customDimensionMap map[string]string) {
	currentTime := time.Now()

	// https://docs.aws.amazon.com/lambda/latest/dg/current-supported-versions.html
	functionName := os.Getenv("AWS_LAMBDA_FUNCTION_NAME")
	cpuMetrics, cpuMetricsErr := gopsutilCPU.Percent(0, false)
	// https://docs.aws.amazon.com/lambda/latest/dg/limits.html
	diskMetrics, diskMetricsErr := gopsutilDisk.Usage("/tmp")
	uptime, uptimeErr := gopsutilHost.Uptime()
	loadMetrics, loadMetricsErr := gopsutilLoad.Avg()
	netMetrics, netMetricsErr := gopsutilNet.IOCounters(false)

	// For now, just log everything...
	logger, _ := sparta.NewLogger("info")
	logger.WithFields(logrus.Fields{
		"functionName":   functionName,
		"cpuMetrics":     cpuMetrics,
		"cpuMetricsErr":  cpuMetricsErr,
		"diskMetrics":    diskMetrics,
		"diskMetricsErr": diskMetricsErr,
		"uptime":         uptime,
		"uptimeErr":      uptimeErr,
		"loadMetrics":    loadMetrics,
		"loadMetricsErr": loadMetricsErr,
		"netMetrics":     netMetrics,
		"netMetricsErr":  netMetricsErr,
	}).Info("Metric info")

	// Return the array of metricDatum for the item
	metricDatum := func(name string, value float64, unit MetricUnit) []*awsCloudWatch.MetricDatum {
		defaultDatum := []*awsCloudWatch.MetricDatum{{
			MetricName: aws.String(name),
			Dimensions: []*awsCloudWatch.Dimension{{
				Name:  aws.String("Name"),
				Value: aws.String(sparta.StampedServiceName),
			}},
			Value:     aws.Float64(value),
			Timestamp: &currentTime,
			Unit:      aws.String(string(unit)),
		},
		}
		if len(customDimensionMap) != 0 {
			metricDimension := []*awsCloudWatch.Dimension{{
				Name:  aws.String("Name"),
				Value: aws.String(sparta.StampedServiceName),
			}}
			for eachKey, eachValue := range customDimensionMap {
				metricDimension = append(metricDimension, &awsCloudWatch.Dimension{
					Name:  aws.String(eachKey),
					Value: aws.String(eachValue),
				})
			}
			defaultDatum = append(defaultDatum, &awsCloudWatch.MetricDatum{
				MetricName: aws.String(name),
				Dimensions: metricDimension,
				Value:      aws.Float64(value),
				Timestamp:  &currentTime,
				Unit:       aws.String(string(unit)),
			})
		}
		return defaultDatum
	}
	// Publish all the metrics...
	// https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_MetricDatum.html
	metricData := []*awsCloudWatch.MetricDatum{}
	// CPU?
	if len(cpuMetrics) == 1 {
		metricData = append(metricData, metricDatum("CPUPercent", cpuMetrics[0], UnitPercent)...)
	}
	if diskMetricsErr == nil {
		metricData = append(metricData, metricDatum("DiskUsedPercent", diskMetrics.UsedPercent, UnitPercent)...)
	}
	if uptimeErr == nil {
		metricData = append(metricData, metricDatum("Uptime", float64(uptime), UnitMilliseconds)...)
	}
	if loadMetricsErr == nil {
		metricData = append(metricData, metricDatum("Load1", loadMetrics.Load1, UnitNone)...)
		metricData = append(metricData, metricDatum("Load5", loadMetrics.Load5, UnitNone)...)
		metricData = append(metricData, metricDatum("Load15", loadMetrics.Load15, UnitNone)...)
	}
	if netMetricsErr == nil && len(netMetrics) == 1 {
		metricData = append(metricData, metricDatum("NetBytesSent", float64(netMetrics[0].BytesSent), UnitBytes)...)
		metricData = append(metricData, metricDatum("NetBytesRecv", float64(netMetrics[0].BytesRecv), UnitBytes)...)
		metricData = append(metricData, metricDatum("NetErrin", float64(netMetrics[0].Errin), UnitCount)...)
		metricData = append(metricData, metricDatum("NetErrout", float64(netMetrics[0].Errout), UnitCount)...)
	}
	putMetricInput := &awsCloudWatch.PutMetricDataInput{
		MetricData: metricData,
		Namespace:  aws.String(sparta.ProperName),
	}
	session := spartaAWS.NewSession(logger)
	awsCloudWatchSvc := awsCloudWatch.New(session)
	putMetricResponse, putMetricResponseErr := awsCloudWatchSvc.PutMetricData(putMetricInput)
	if putMetricResponseErr != nil {
		logger.WithField("Error", putMetricResponseErr).Error("Failed to submit CloudWatch Metric data")
	} else {
		logger.WithField("Response", putMetricResponse).Info("CloudWatch Metric response")
	}
}

// RegisterLambdaUtilizationMetricPublisher installs a periodic task
// to publish the current system metrics to CloudWatch Metrics. See
// https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/cloudwatch_concepts.html
// for more information.
func RegisterLambdaUtilizationMetricPublisher(customDimensionMap map[string]string) {

	// Publish when we start
	publishMetrics(customDimensionMap)

	ticker := time.NewTicker(1 * time.Minute)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				publishMetrics(customDimensionMap)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}
