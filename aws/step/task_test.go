package step

import (
	"testing"
)

func TestTaskState(t *testing.T) {
	// Make a sample video transcription task per the
	// docs at https://aws.amazon.com/blogs/aws/now-aws-step-functions-supports-200-aws-services-to-enable-easier-workflow-automation/

	taskParams := map[string]interface{}{
		"Bucket.$":     "$.S3BucketName",
		"Key.$":        "$.SampleDataInputKey",
		"CopySource.$": "States.Format('{}/{}',$.SampleDataBucketName,$.SampleDataInputKey)",
	}
	taskStep := NewTaskState("copyObject",
		"arn:aws:states:::aws-sdk:s3:copyObject", taskParams)
	successState := NewSuccessState("success")
	// Hook them up..
	taskStep.Next(successState)

	// Startup the machine.
	startMachine := NewStateMachine("SampleStepFunction", taskStep)

	testStepProvision(t,
		nil,
		startMachine)
}
