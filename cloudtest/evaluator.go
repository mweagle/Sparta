package cloudtest

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	awsv2CW "github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	awsv2CWTypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	awsv2Lambda "github.com/aws/aws-sdk-go-v2/service/lambda"
	spartaCWLogs "github.com/mweagle/Sparta/aws/cloudwatch/logs"
	"github.com/pkg/errors"
)

const metricPeriod = 15

////////////////////////////////////////////////////////////////////////////////
type timedNOPEvaluator struct {
	duration time.Duration
}

func (tne *timedNOPEvaluator) Evaluate(t CloudTest, output *awsv2Lambda.GetFunctionOutput) error {
	time.Sleep(tne.duration)
	t.Logf("NOP success after artificial delay of %s for function: %s",
		tne.duration.String(),
		*output.Configuration.FunctionName)
	return nil
}

// NewTimedNOPEvaluator returns a NOP evaluator to ensure the test always
// passes
func NewTimedNOPEvaluator(timeout time.Duration) CloudEvaluator {
	return &timedNOPEvaluator{
		duration: timeout,
	}
}

////////////////////////////////////////////////////////////////////////////////

type lambdaLogOutputEvaluator struct {
	initTime time.Time
	matcher  *regexp.Regexp
}

func (lloe *lambdaLogOutputEvaluator) Evaluate(t CloudTest, output *awsv2Lambda.GetFunctionOutput) error {
	// Start reading the logfiles as soon as we're initialized...
	lambdaParts := strings.Split(*output.Configuration.FunctionArn, ":")
	logGroupName := fmt.Sprintf("/aws/lambda/%s", lambdaParts[len(lambdaParts)-1])

	// Put this as the label in the view...
	doneChan := make(chan bool)
	messages := spartaCWLogs.TailWithContext(context.Background(),
		doneChan,
		t.Config(),
		logGroupName,
		"",
		t.ZeroLog())

	// Look for a match
	isMatched := false
	for !isMatched {
		select {
		case <-t.Context().Done():
			return errors.Errorf("Test deadline exceeded")
		case event := <-messages:
			{
				messageMatch := lloe.matcher.FindString(*event.Message)
				isMatched = (messageMatch != "")
				if isMatched {
					break
				}
			}
		}
	}
	// Need to wait for this to end...
	return nil
}

// NewLogOutputEvaluator returns a CloudEvaluator that scans CloudWatchLogs
// for a given regexp pattern
func NewLogOutputEvaluator(matcher *regexp.Regexp) CloudEvaluator {
	return &lambdaLogOutputEvaluator{
		initTime: time.Now(),
		matcher:  matcher,
	}
}

////////////////////////////////////////////////////////////////////////////////

// MetricEvaluator returns whether the evaluatoin should continue or an error
// occurred.
type MetricEvaluator func(map[MetricName][]float64) (bool, error)

// MetricName is tha alias type for the reserved Lambda invocation metrics
// defined at https://docs.awsv2.amazon.com/lambda/latest/dg/monitoring-metrics.html
type MetricName string

const (
	// MetricNameInvocations is the Lambda invocation
	MetricNameInvocations MetricName = "Invocations"
	// MetricNameErrors is the number of invocations that result in a function error
	MetricNameErrors MetricName = "Errors"
	// MetricNameDeadLetterErrors the number of times Lambda attempts to send an event to a
	// dead-letter queue but fails
	MetricNameDeadLetterErrors MetricName = "DeadLetterErrors"
	// MetricNameDestinationDeliveryFailures the number of times Lambda attempts
	// to send an event to a destination but fails
	MetricNameDestinationDeliveryFailures MetricName = "DestinationDeliveryFailures"
	// MetricNameThrottles is the number of invocation requests that are throttled
	MetricNameThrottles MetricName = "Throttles"
	// MetricNameProvisionedConcurrencyInvocations is the number of times your
	// function code is executed on provisioned concurrency.
	MetricNameProvisionedConcurrencyInvocations MetricName = "ProvisionedConcurrencyInvocations"
	// MetricNameProvisionedConcurrencySpilloverInvocations is the number of times
	// your function code is executed on standard concurrency when all provisioned
	// concurrency is in use
	MetricNameProvisionedConcurrencySpilloverInvocations MetricName = "ProvisionedConcurrencySpilloverInvocations"
)

func idValue(name MetricName) string {
	return strings.ToLower(string(name))
}

// So we need a mapping of ID to metric name...argh...
var mapIDToMetricName = map[string]MetricName{
	idValue(MetricNameInvocations):                                MetricNameInvocations,
	idValue(MetricNameErrors):                                     MetricNameErrors,
	idValue(MetricNameDeadLetterErrors):                           MetricNameDeadLetterErrors,
	idValue(MetricNameDestinationDeliveryFailures):                MetricNameDestinationDeliveryFailures,
	idValue(MetricNameThrottles):                                  MetricNameThrottles,
	idValue(MetricNameProvisionedConcurrencyInvocations):          MetricNameProvisionedConcurrencyInvocations,
	idValue(MetricNameProvisionedConcurrencySpilloverInvocations): MetricNameProvisionedConcurrencySpilloverInvocations,
}

// IsSuccess is the default evaluator to see if a set of Lambda metrics
// indicate a successful result
var IsSuccess = func(values map[MetricName][]*float64) (bool, error) {

	invocationsTotal := func(values []*float64) float64 {
		totalAgg := float64(0)
		for _, eachValue := range values {
			totalAgg += *eachValue
		}
		return totalAgg
	}

	errorChecks := []MetricName{MetricNameErrors, MetricNameDeadLetterErrors}
	for _, eachErr := range errorChecks {
		errValues := values[eachErr]
		totalAggregateValue := invocationsTotal(errValues)
		if totalAggregateValue != 0 {
			return true, errors.Errorf("At least one %s found for invocation. Count: %f",
				eachErr,
				totalAggregateValue)
		}
	}

	if len(values[MetricNameInvocations]) > 0 {
		return true, nil
	}
	// Keep going
	return false, nil
}

type lambdaInvocationMetricEvaluator struct {
	initTime        time.Time
	queries         []awsv2CWTypes.MetricDataQuery
	metricEvaluator MetricEvaluator
}

func (lime *lambdaInvocationMetricEvaluator) Evaluate(t CloudTest,
	output *awsv2Lambda.GetFunctionOutput) error {
	// Just sit there and see if the thing successfully executed...so this is a
	// cloudwatch metric query?
	for _, eachQuery := range lime.queries {
		for _, eachDimension := range eachQuery.MetricStat.Metric.Dimensions {
			eachDimension.Value = output.Configuration.FunctionName
		}
	}
	cwService := awsv2CW.NewFromConfig(t.Config())
	tickerDuration := metricPeriod * time.Second
	ticker := time.NewTicker(tickerDuration)

	// Poller duraction
	offsetDuration, _ := time.ParseDuration(fmt.Sprintf("%ds", metricPeriod))
	getMetricParams := &awsv2CW.GetMetricDataInput{
		StartTime:         awsv2.Time(lime.initTime),
		EndTime:           awsv2.Time(lime.initTime),
		MetricDataQueries: lime.queries,
	}
	breakTest := false
	for {
		if breakTest {
			break
		}
		select {
		case <-t.Context().Done():
			ticker.Stop()
			return errors.Errorf("Deadline exceeded for test")
		case <-ticker.C:
			getMetricParams.EndTime = awsv2.Time(getMetricParams.StartTime.Add(offsetDuration))
			//t.Logf("getMetricParams: #%v", getMetricParams)
			getMetricOutput, getMetricOutputErr := cwService.GetMetricData(context.Background(), getMetricParams)
			if getMetricOutputErr != nil {
				return getMetricOutputErr
			}

			// For each response, create a map of ID to Values
			metricOutput := map[MetricName][]float64{}
			for _, eachResult := range getMetricOutput.MetricDataResults {
				name, nameExists := mapIDToMetricName[*eachResult.Id]
				if !nameExists {
					return errors.Errorf("Metric ID %s in query is not recognized as valid MetricName", *eachResult.Id)
				}
				metricOutput[name] = eachResult.Values
			}
			stopTest, testError := lime.metricEvaluator(metricOutput)

			// mapOutput := map[string]interface{}{
			// 	"Metrics":         metricOutput,
			// 	"Stop":            stopTest,
			// 	"EvaluationError": testError,
			// }
			//logOutput, _ := json.Marshal(mapOutput)
			//t.Logf("Metric evaluation result: %s", string(logOutput))
			if testError != nil {
				return testError
			}
			breakTest = stopTest
		}
	}
	return nil
}

// NewLambdaFunctionMetricQuery returns a awsv2CW.MetricDataQuery
// that will be lazily completed in the Evaluation function
func NewLambdaFunctionMetricQuery(invocationMetricName MetricName) *awsv2CWTypes.MetricDataQuery {
	return &awsv2CWTypes.MetricDataQuery{
		Id: awsv2.String(strings.ToLower(string(invocationMetricName))),
		MetricStat: &awsv2CWTypes.MetricStat{
			Period: awsv2.Int32(30),
			Stat:   awsv2.String(string(awsv2CWTypes.StatisticSum)),
			Unit:   awsv2CWTypes.StandardUnitCount,
			Metric: &awsv2CWTypes.Metric{
				Namespace:  awsv2.String("AWS/Lambda"),
				MetricName: awsv2.String(string(invocationMetricName)),
				Dimensions: []awsv2CWTypes.Dimension{
					{
						Name:  awsv2.String("FunctionName"),
						Value: nil,
					},
				},
			},
		},
	}
}

// DefaultLambdaFunctionMetricQueries is the standard set of queries
// to issue to determine in a Lambda successfully executed
func DefaultLambdaFunctionMetricQueries() []*awsv2CWTypes.MetricDataQuery {
	return []*awsv2CWTypes.MetricDataQuery{
		NewLambdaFunctionMetricQuery(MetricNameInvocations),
		NewLambdaFunctionMetricQuery(MetricNameErrors),
		NewLambdaFunctionMetricQuery(MetricNameDeadLetterErrors),
	}
}

// NewLambdaInvocationMetricEvaluator needs a list of
// awsv2CW.MetricDataQuery results that
// need the FunctionName. Then the evaluation will take a map
// of metrics to values.
func NewLambdaInvocationMetricEvaluator(queries []awsv2CWTypes.MetricDataQuery,
	metricEvaluator MetricEvaluator) CloudEvaluator {
	nowTime := time.Now()

	// We won't get initialized before the trigger function is called, so ensure there's
	// enough buffer for a lower bound
	addDuration, _ := time.ParseDuration("2s")
	return &lambdaInvocationMetricEvaluator{
		initTime:        nowTime.Add(-addDuration),
		queries:         queries,
		metricEvaluator: metricEvaluator,
	}
}
