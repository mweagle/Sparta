package testing

import (
	"bytes"
	"context"
	"testing"

	sparta "github.com/mweagle/Sparta/v3"
)

// ProvisionEvaluator is the function that is called following a
// provision to determine if the result was successful
type ProvisionEvaluator func(t *testing.T, didError error) error

// AssertSuccess is a default handler for the ProvisionRunner. If no
// evaluator is supplied, defaults to expecting no didError
func AssertSuccess(t *testing.T, didError error) error {
	if didError != nil {
		t.Fatal("Provision failed: " + didError.Error())
	}
	return nil
}

// AssertError returns a test evaluator that enforces that didError is not nil
func AssertError(message string) ProvisionEvaluator {
	return func(t *testing.T, didError error) error {
		t.Logf("Checking provisioning error: %s", didError)
		if didError == nil {
			t.Fatal("Failed to reject error due to: " + message)
		}
		return nil
	}
}

// Provision is a convenience function for ProvisionEx
func Provision(t *testing.T,
	lambdaAWSInfos []*sparta.LambdaAWSInfo,
	evaluator ProvisionEvaluator) {

	ProvisionEx(t, lambdaAWSInfos, nil, nil, nil, false, evaluator)
}

// ProvisionEx handles mock provisioning a service and then
// supplying the result to the evaluator function. If no evaluator
// is performed it's assumed that the provision operation should succeed without
// error.
func ProvisionEx(t *testing.T,
	lambdaAWSInfos []*sparta.LambdaAWSInfo,
	api *sparta.API,
	site *sparta.S3Site,
	workflowHooks *sparta.WorkflowHooks,
	useCGO bool,
	evaluator ProvisionEvaluator) {

	if evaluator == nil {
		evaluator = AssertSuccess
	}

	logger, loggerErr := sparta.NewLogger("info")
	if loggerErr != nil {
		t.Fatalf("Failed to create test logger: %s", loggerErr)
	}
	var templateWriter bytes.Buffer
	err := sparta.Build(context.Background(),
		true,
		"SampleProvision",
		"",
		lambdaAWSInfos,
		nil,
		nil,
		false,
		"testBuildID",
		"",
		"",
		"",
		"",
		&templateWriter,
		workflowHooks,
		logger)
	if evaluator != nil {
		err = evaluator(t, err)
	}
	if err != nil {
		t.Fatalf("Failed to apply evaluator: " + err.Error())
	}
}
