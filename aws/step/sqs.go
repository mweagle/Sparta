package step

import (
	"math/rand"

	gocf "github.com/mweagle/go-cloudformation"
)

// SQSTaskParameters represents params for the SQS notification
// Ref: https://docs.aws.amazon.com/sns/latest/api/API_Publish.html#API_Publish_RequestParameters
type SQSTaskParameters struct {
	MessageBody            string                 `json:",omitempty"`
	QueueURL               gocf.Stringable        `json:",omitempty"`
	DelaySeconds           int                    `json:",omitempty"`
	MessageAttributes      map[string]interface{} `json:",omitempty"`
	MessageDeduplicationID string                 `json:"MessageDeduplicationId,omitempty"`
	MessageGroupID         string                 `json:"MessageGroupId,omitempty"`
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
	return sqs.BaseTask.marshalMergedParams("arn:aws:states:::sqs:sendMessage",
		&sqs.parameters)
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
