//go:build lambdabinary
// +build lambdabinary

package cloudwatch

import (
	"context"
	"os"
	"time"

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	awsv2CW "github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	awsv2CWTypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	sparta "github.com/mweagle/Sparta/v3"
	spartaAWS "github.com/mweagle/Sparta/v3/aws"
	"github.com/rs/zerolog"
	gopsutilCPU "github.com/shirou/gopsutil/v3/cpu"
	gopsutilDisk "github.com/shirou/gopsutil/v3/disk"
	gopsutilHost "github.com/shirou/gopsutil/v3/host"
	gopsutilLoad "github.com/shirou/gopsutil/v3/load"
	gopsutilNet "github.com/shirou/gopsutil/v3/net"
)

// publishMetrics is the actual metric publishing logic. T
func publishMetrics(customDimensionMap map[string]string) {
	currentTime := time.Now()
	publishContext := context.Background()

	// https://docs.awsv2.amazon.com/lambda/latest/dg/current-supported-versions.html
	functionName := os.Getenv("AWS_LAMBDA_FUNCTION_NAME")
	cpuMetrics, cpuMetricsErr := gopsutilCPU.Percent(0, false)
	// https://docs.awsv2.amazon.com/lambda/latest/dg/limits.html
	diskMetrics, diskMetricsErr := gopsutilDisk.Usage("/tmp")
	uptime, uptimeErr := gopsutilHost.Uptime()
	loadMetrics, loadMetricsErr := gopsutilLoad.Avg()
	netMetrics, netMetricsErr := gopsutilNet.IOCounters(false)

	// For now, just log everything...
	logger, _ := sparta.NewLogger(zerolog.InfoLevel.String())
	if logger != nil {
		logger.Info().
			Str("functionName", functionName).
			Interface("cpuMetrics", cpuMetrics).
			Interface("cpuMetricsErr", cpuMetricsErr).
			Interface("diskMetrics", diskMetrics).
			Interface("diskMetricsErr", diskMetricsErr).
			Interface("uptime", uptime).
			Interface("uptimeErr", uptimeErr).
			Interface("loadMetrics", loadMetrics).
			Interface("loadMetricsErr", loadMetricsErr).
			Interface("netMetrics", netMetrics).
			Interface("netMetricsErr", netMetricsErr).
			Msg("Metric info")
	}
	// Return the array of metricDatum for the item
	metricDatum := func(name string, value float64, unit awsv2CWTypes.StandardUnit) []awsv2CWTypes.MetricDatum {
		defaultDatum := []awsv2CWTypes.MetricDatum{{
			MetricName: awsv2.String(name),
			Dimensions: []awsv2CWTypes.Dimension{{
				Name:  awsv2.String("Name"),
				Value: awsv2.String(sparta.StampedServiceName),
			}},
			Value:     awsv2.Float64(value),
			Timestamp: &currentTime,
			Unit:      unit,
		},
		}
		if len(customDimensionMap) != 0 {
			metricDimension := []awsv2CWTypes.Dimension{{
				Name:  awsv2.String("Name"),
				Value: awsv2.String(sparta.StampedServiceName),
			}}
			for eachKey, eachValue := range customDimensionMap {
				metricDimension = append(metricDimension, awsv2CWTypes.Dimension{
					Name:  awsv2.String(eachKey),
					Value: awsv2.String(eachValue),
				})
			}
			defaultDatum = append(defaultDatum, awsv2CWTypes.MetricDatum{
				MetricName: awsv2.String(name),
				Dimensions: metricDimension,
				Value:      awsv2.Float64(value),
				Timestamp:  &currentTime,
				Unit:       unit,
			})
		}
		return defaultDatum
	}
	// Publish all the metrics...
	// https://docs.awsv2.amazon.com/AmazonCloudWatch/latest/APIReference/API_MetricDatum.html
	metricData := []awsv2CWTypes.MetricDatum{}
	// CPU?
	if len(cpuMetrics) == 1 {
		metricData = append(metricData, metricDatum("CPUPercent", cpuMetrics[0], awsv2CWTypes.StandardUnitPercent)...)
	}
	if diskMetricsErr == nil {
		metricData = append(metricData, metricDatum("DiskUsedPercent", diskMetrics.UsedPercent, awsv2CWTypes.StandardUnitPercent)...)
	}
	if uptimeErr == nil {
		metricData = append(metricData, metricDatum("Uptime", float64(uptime), awsv2CWTypes.StandardUnitMilliseconds)...)
	}
	if loadMetricsErr == nil {
		metricData = append(metricData, metricDatum("Load1", loadMetrics.Load1, awsv2CWTypes.StandardUnitNone)...)
		metricData = append(metricData, metricDatum("Load5", loadMetrics.Load5, awsv2CWTypes.StandardUnitNone)...)
		metricData = append(metricData, metricDatum("Load15", loadMetrics.Load15, awsv2CWTypes.StandardUnitNone)...)
	}
	if netMetricsErr == nil && len(netMetrics) == 1 {
		metricData = append(metricData, metricDatum("NetBytesSent", float64(netMetrics[0].BytesSent), awsv2CWTypes.StandardUnitBytes)...)
		metricData = append(metricData, metricDatum("NetBytesRecv", float64(netMetrics[0].BytesRecv), awsv2CWTypes.StandardUnitBytes)...)
		metricData = append(metricData, metricDatum("NetErrin", float64(netMetrics[0].Errin), awsv2CWTypes.StandardUnitCount)...)
		metricData = append(metricData, metricDatum("NetErrout", float64(netMetrics[0].Errout), awsv2CWTypes.StandardUnitCount)...)
	}
	putMetricInput := &awsv2CW.PutMetricDataInput{
		MetricData: metricData,
		Namespace:  awsv2.String(sparta.ProperName),
	}
	awsConfig, awsConfigErr := spartaAWS.NewConfig(publishContext, logger)
	if awsConfigErr != nil {
		return
	}

	awsCloudWatchSvc := awsv2CW.NewFromConfig(awsConfig)
	putMetricResponse, putMetricResponseErr := awsCloudWatchSvc.PutMetricData(publishContext,
		putMetricInput)
	if putMetricResponseErr != nil {
		logger.Error().Err(putMetricResponseErr).Msg("Failed to submit CloudWatch Metric data")
	} else {
		logger.Info().Interface("Response", putMetricResponse).Msg("CloudWatch Metric response")
	}
}

// RegisterLambdaUtilizationMetricPublisher installs a periodic task
// to publish the current system metrics to CloudWatch Metrics. See
// https://docs.awsv2.amazon.com/AmazonCloudWatch/latest/monitoring/cloudwatch_concepts.html
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
