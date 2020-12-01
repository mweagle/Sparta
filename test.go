package sparta

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/rs/zerolog"
)

const lambdaTestExecuteARN = "LambdaExecutor"
const s3BucketSourceArn = "arn:aws:s3:::sampleBucket"
const snsTopicSourceArn = "arn:aws:sns:us-west-2:000000000000:someTopic"
const dynamoDBTableArn = "arn:aws:dynamodb:us-west-2:000000000000:table/sampleTable"

func mockLambda1(ctx context.Context) (string, error) {
	return "mockLambda1!", nil
}

func mockLambda2(ctx context.Context) (string, error) {
	return "mockLambda2!", nil
}

func mockLambda3(ctx context.Context) (string, error) {
	return "mockLambda3!", nil
}

func testLambdaData() []*LambdaAWSInfo {
	var lambdaFunctions []*LambdaAWSInfo

	//////////////////////////////////////////////////////////////////////////////
	// Lambda function 1
	lambdaFn1, lambdaFn1Err := NewAWSLambda(LambdaName(mockLambda1),
		mockLambda1,
		lambdaTestExecuteARN)
	if lambdaFn1Err != nil {
		panic("Failed to create lambda1")
	}
	lambdaFn1.Permissions = append(lambdaFn1.Permissions, S3Permission{
		BasePermission: BasePermission{
			SourceArn: s3BucketSourceArn,
		},
		// Event Filters are defined at
		// http://docs.aws.amazon.com/AmazonS3/latest/dev/NotificationHowTo.html
		Events: []string{"s3:ObjectCreated:*", "s3:ObjectRemoved:*"},
	})

	lambdaFn1.Permissions = append(lambdaFn1.Permissions, SNSPermission{
		BasePermission: BasePermission{
			SourceArn: snsTopicSourceArn,
		},
	})

	lambdaFn1.EventSourceMappings = append(lambdaFn1.EventSourceMappings, &EventSourceMapping{
		StartingPosition: "TRIM_HORIZON",
		EventSourceArn:   dynamoDBTableArn,
		BatchSize:        10,
	})

	lambdaFunctions = append(lambdaFunctions, lambdaFn1)

	//////////////////////////////////////////////////////////////////////////////
	// Lambda function 2
	lambdaFn2, lambdaFn2Err := NewAWSLambda(LambdaName(mockLambda2),
		mockLambda2,
		lambdaTestExecuteARN)
	if lambdaFn2Err != nil {
		panic("Failed to create lambda2")
	}
	lambdaFunctions = append(lambdaFunctions, lambdaFn2)

	//////////////////////////////////////////////////////////////////////////////
	// Lambda function 3
	// https://github.com/mweagle/Sparta/pull/1
	lambdaFn3, lambdaFn3Err := NewAWSLambda(LambdaName(mockLambda3),
		mockLambda3,
		lambdaTestExecuteARN)
	if lambdaFn3Err != nil {
		panic("Failed to create lambda3")
	}
	lambdaFn3.Permissions = append(lambdaFn3.Permissions, SNSPermission{
		BasePermission: BasePermission{
			SourceArn: snsTopicSourceArn,
		},
	})
	lambdaFunctions = append(lambdaFunctions, lambdaFn3)
	return lambdaFunctions
}

// testProvisionEvaluator is the function that is called following a
// provision to determine if the result was successful
type testProvisionEvaluator func(t *testing.T, didError error) error

// assertSuccess is a default handler for the ProvisionRunner. If no
// evaluator is supplied, defaults to expecting no didError
func assertSuccess(t *testing.T, didError error) error {
	if didError != nil {
		t.Fatal("Provision failed: " + didError.Error())
	}
	return nil
}

// assertError returns a test evaluator that enforces that didError is not nil
func assertError(message string) testProvisionEvaluator {
	return func(t *testing.T, didError error) error {
		t.Logf("Checking provisioning error: %s", didError)
		if didError == nil {
			t.Fatal("Failed to reject error due to: " + message)
		}
		return nil
	}
}

// testProvision is a convenience function for testProvisionEx
func testProvision(t *testing.T,
	lambdaAWSInfos []*LambdaAWSInfo,
	evaluator testProvisionEvaluator) {

	testProvisionEx(t, lambdaAWSInfos, nil, nil, nil, false, evaluator)
}

// testProvisionEx handles mock provisioning a service and then
// supplying the result to the evaluator function
func testProvisionEx(t *testing.T,
	lambdaAWSInfos []*LambdaAWSInfo,
	api APIGateway,
	site *S3Site,
	workflowHooks *WorkflowHooks,
	useCGO bool,
	evaluator testProvisionEvaluator) {

	if evaluator == nil {
		evaluator = assertSuccess
	}

	logger, loggerErr := NewLogger(zerolog.InfoLevel.String())
	if loggerErr != nil {
		t.Fatalf("Failed to create test logger: %s", loggerErr)
	}
	var templateWriter bytes.Buffer

	workingDir, workingDirErr := os.Getwd()
	if workingDirErr != nil {
		t.Error(workingDirErr)
	}
	fullPath, fullPathErr := filepath.Abs(workingDir)
	if fullPathErr != nil {
		t.Error(fullPathErr)
	}
	err := Build(true,
		"SampleProvision",
		"",
		lambdaAWSInfos,
		nil,
		nil,
		false,
		"testBuildID",
		fullPath,
		"",
		"",
		&templateWriter,
		nil,
		logger)
	if evaluator != nil {
		err = evaluator(t, err)
	}
	if err != nil {
		t.Fatalf("Failed to apply evaluator: " + err.Error())
	}
}
