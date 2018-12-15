package cloudwatch

import (
	"context"
	"testing"

	sparta "github.com/mweagle/Sparta"
	spartaTesting "github.com/mweagle/Sparta/testing"
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
	spartaTesting.Provision(t, testLambdaData(t), nil)
}
