//go:build integration
// +build integration

package cloudtest

import (
	"bytes"
	"fmt"
	awsv2Config "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"regexp"
	"testing"
	"time"
)

var accountID = ""

func init() {
	awsConfig := awsv2Config.LoadDefaultConfig(context.Background())
	if awsSessionErr == nil {
		stsService := sts.NewFromConfig(awsConfig)
		callerInfo, callerInfoErr := stsService.GetCallerIdentity(&sts.GetCallerIdentityInput{})
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
