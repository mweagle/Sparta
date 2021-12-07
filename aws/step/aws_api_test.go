package step

import (
	"testing"

	gof "github.com/awslabs/goformation/v5/cloudformation"
	sparta "github.com/mweagle/Sparta/v3"
)

func TestAWSAPIStepFunction(t *testing.T) {
	awsGetMetadataState := NewAWSSDKState("getMetadata",
		"s3",
		"headObject",
		"",
		map[string]interface{}{
			"Bucket": "weagle",
			"Key.$":  "$.SampleDataInputKey",
		})
	startMachine := NewStateMachine("AWSAPIMachine", awsGetMetadataState).WithRoleArn(gof.GetAtt("StepMachineRole", "Arn"))

	testStepProvision(t,
		[]*sparta.LambdaAWSInfo{},
		startMachine)
}
