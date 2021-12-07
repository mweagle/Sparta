//go:build integration
// +build integration

package cloudtest

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"testing"
	"time"

	awsv2Config "github.com/aws/aws-sdk-go-v2/config"
	awsv2STS "github.com/aws/aws-sdk-go-v2/service/sts"
)

var accountID = ""

func init() {
	awsConfig := awsv2Config.LoadDefaultConfig(context.Background())
	if awsSessionErr == nil {
		stsService := awsv2STS.NewFromConfig(awsConfig)
		callerInfo, callerInfoErr := stsService.GetCallerIdentity(&awsv2STS.GetCallerIdentityInput{})
		if callerInfoErr == nil {
			accountID = *callerInfo.Account
		}
	}
}

var helloWorldJSON = []byte(`{
    "hello" : "world"
}`)

func TestCloudMetricsTest(t *testing.T) {
	NewTest().
		Given(NewLambdaInvokeTrigger(helloWorldJSON)).
		Against(
			NewStackLambdaSelector(fmt.Sprintf("MyOCIStack-%s", accountID),
				"[Configuration][?contains(FunctionName,'Hello_World')].FunctionName | [0]")).
		Ensure(NewLambdaInvocationMetricEvaluator(DefaultLambdaFunctionMetricQueries(),
			IsSuccess),
		).
		Run(t)
}

func TestCloudLogOutputTest(t *testing.T) {
	NewTest().
		Given(NewLambdaInvokeTrigger(helloWorldJSON)).
		Against(
			NewStackLambdaSelector(fmt.Sprintf("MyOCIStack-%s", accountID),
				"[Configuration][?contains(FunctionName,'Hello_World')].FunctionName | [0]")).
		Ensure(NewLogOutputEvaluator(regexp.MustCompile("Accessing"))).
		Run(t)
}

func TestCloudLiteralLogOutputTest(t *testing.T) {
	NewTest().
		Given(NewLambdaInvokeTrigger(helloWorldJSON)).
		Against(NewLambdaLiteralSelector(fmt.Sprintf("MyOCIStack-%s_Hello_World", accountID))).
		Ensure(NewLogOutputEvaluator(regexp.MustCompile("Accessing"))).
		Run(t)
}

func TestCloudSQSLambdaHandler(t *testing.T) {
	NewTest().
		Given(NewSQSMessageTrigger(
			fmt.Sprintf("https://sqs.us-west-2.amazonaws.com/%s/SpartaTest", accountID),
			"Hello World!")).
		Against(
			NewLambdaLiteralSelector("MySampleSQSFunction")).
		Ensure(NewLambdaInvocationMetricEvaluator(
			DefaultLambdaFunctionMetricQueries(),
			IsSuccess),
		).
		Run(t)
}

func TestS3LambdaHandler(t *testing.T) {
	dataUpload := bytes.NewReader(helloWorldJSON)
	NewTest().
		Given(NewS3MessageTrigger(
			"weagle-sparta-testbucket",
			fmt.Sprintf("testKey%d", time.Now().Unix()),
			dataUpload)).
		Against(
			NewLambdaLiteralSelector("SampleS3Uploaded")).
		Ensure(NewLogOutputEvaluator(regexp.MustCompile("CONTENT TYPE"))).
		Run(t)
}

func TestFileS3LambdaHandler(t *testing.T) {
	NewTest().
		Given(NewS3FileMessageTrigger(
			"weagle-sparta-testbucket",
			fmt.Sprintf("testKey%d", time.Now().Unix()),
			"./cloudtest.go")).
		Against(NewLambdaLiteralSelector("SampleS3Uploaded")).
		Ensure(NewLogOutputEvaluator(regexp.MustCompile("CONTENT TYPE"))).
		Run(t)
}
