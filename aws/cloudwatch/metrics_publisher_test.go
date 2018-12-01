package cloudwatch

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/mweagle/Sparta"
)

func testLambdaData(t *testing.T) []*sparta.LambdaAWSInfo {
	mockLambda := func(ctx context.Context) (string, error) {
		return "mockLambda!", nil
	}
	RegisterLambdaUtilizationMetricPublisher(map[string]string{
		"BuildId": sparta.StampedBuildID,
	})

	lambdaFn, lambdaFnErr := sparta.NewAWSLambda(sparta.LambdaName(mockLambda),
		mockLambda,
		sparta.IAMRoleDefinition{})
	if lambdaFnErr != nil {
		t.Fatal(lambdaFnErr.Error())
	}
	return []*sparta.LambdaAWSInfo{lambdaFn}
}

func TestRegisterMetricsPublisher(t *testing.T) {
	lambdas := testLambdaData(t)
	logger, _ := sparta.NewLogger("info")
	var templateWriter bytes.Buffer
	err := sparta.Provision(true,
		"SampleProvision",
		"",
		lambdas,
		nil,
		nil,
		os.Getenv("S3_BUCKET"),
		false,
		false,
		"testBuildID",
		"",
		"",
		"",
		&templateWriter,
		nil,
		logger)
	if nil != err {
		t.Fatal(err.Error())
	}
}
