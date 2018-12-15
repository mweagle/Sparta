package step

import (
	"math/rand"

	gocf "github.com/mweagle/go-cloudformation"
)

// SNSTaskParameters represents params for the SNS notification
// Ref: https://docs.aws.amazon.com/sns/latest/api/API_Publish.html#API_Publish_RequestParameters
type SNSTaskParameters struct {
	Message           string                 `json:",omitempty"`
	Subject           string                 `json:",omitempty"`
	MessageAttributes map[string]interface{} `json:",omitempty"`
	MessageStructure  string                 `json:",omitempty"`
	PhoneNumber       string                 `json:",omitempty"`
	TargetArn         gocf.Stringable        `json:",omitempty"`
	TopicArn          gocf.Stringable        `json:",omitempty"`
}

// SNSTaskState represents bindings for
// https://docs.aws.amazon.com/step-functions/latest/dg/connectors-sns.html
type SNSTaskState struct {
	BaseTask
	parameters SNSTaskParameters
}

// MarshalJSON for custom marshalling, since this will be stringified and we need it
// to turn into a stringified
// Ref: https://docs.aws.amazon.com/step-functions/latest/dg/connectors-sns.html
func (sts *SNSTaskState) MarshalJSON() ([]byte, error) {
	return sts.BaseTask.marshalMergedParams("arn:aws:states:::sns:publish",
		&sts.parameters)
}

// NewSNSTaskState returns an initialized SNSTaskState
func NewSNSTaskState(stateName string,
	parameters SNSTaskParameters) *SNSTaskState {

	sns := &SNSTaskState{
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
