package step

import (
	"math/rand"

	gocf "github.com/mweagle/go-cloudformation"
)

// SQSTaskParameters represents params for the SQS notification
// Ref: https://docs.aws.amazon.com/sns/latest/api/API_Publish.html#API_Publish_RequestParameters
type SQSTaskParameters struct {
	MessageBody            string
	QueueURL               gocf.Stringable
	DelaySeconds           int
	MessageAttributes      map[string]interface{}
	MessageDeduplicationID string
	MessageGroupID         string
}

// SQSTaskState represents bindings for
// https://docs.aws.amazon.com/step-functions/latest/dg/connectors-sqs.html
type SQSTaskState struct {
	BaseTask
	parameters SQSTaskParameters
}

// MarshalJSON for custom marshalling, since this will be stringified and we need it
// to turn into a stringified
// Ref: https://docs.aws.amazon.com/step-functions/latest/dg/connectors-sqs.html
func (sqs *SQSTaskState) MarshalJSON() ([]byte, error) {

	additionalParams := sqs.BaseTask.additionalParams()
	additionalParams["Resource"] = "arn:aws:states:::sqs:sendMessage"
	parameterMap := map[string]interface{}{}

	if sqs.parameters.MessageBody != "" {
		parameterMap["MessageBody"] = sqs.parameters.MessageBody
	}
	if sqs.parameters.QueueURL != nil {
		parameterMap["QueueUrl"] = sqs.parameters.QueueURL
	}
	if sqs.parameters.DelaySeconds != 0 {
		parameterMap["DelaySeconds"] = sqs.parameters.DelaySeconds
	}
	if sqs.parameters.MessageAttributes != nil {
		parameterMap["MessageAttributes"] = sqs.parameters.MessageAttributes
	}
	if sqs.parameters.MessageDeduplicationID != "" {
		parameterMap["MessageDeduplicationId"] = sqs.parameters.MessageDeduplicationID
	}
	if sqs.parameters.MessageGroupID != "" {
		parameterMap["MessageGroupId"] = sqs.parameters.MessageGroupID
	}
	additionalParams["Parameters"] = parameterMap
	return sqs.marshalStateJSON("Task", additionalParams)
}

// NewSQSTaskState returns an initialized SQSTaskState
func NewSQSTaskState(stateName string,
	parameters SQSTaskParameters) *SQSTaskState {
	sns := &SQSTaskState{
		BaseTask: BaseTask{
			baseInnerState: baseInnerState{
				name: stateName,
				id:   rand.Int63(),
			},
		},
		parameters: parameters,
	}
	return sns
}
