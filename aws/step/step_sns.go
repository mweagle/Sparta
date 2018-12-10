package step

import (
	"math/rand"

	gocf "github.com/mweagle/go-cloudformation"
)

// SNSTaskParameters represents params for the SNS notification
// Ref: https://docs.aws.amazon.com/sns/latest/api/API_Publish.html#API_Publish_RequestParameters
type SNSTaskParameters struct {
	Message           string
	Subject           string
	MessageAttributes map[string]interface{}
	MessageStructure  string
	PhoneNumber       string
	TargetArn         gocf.Stringable
	TopicArn          gocf.Stringable
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

	additionalParams := sts.BaseTask.additionalParams()
	additionalParams["Resource"] = "arn:aws:states:::sns:publish"
	parameterMap := map[string]interface{}{}

	if sts.parameters.TopicArn != nil {
		parameterMap["TopicArn"] = sts.parameters.TopicArn
	}
	if sts.parameters.Message != "" {
		parameterMap["Message"] = sts.parameters.Message
	}
	if sts.parameters.MessageAttributes != nil {
		parameterMap["MessageAttributes"] = sts.parameters.MessageAttributes
	}
	if sts.parameters.MessageStructure != "" {
		parameterMap["MessageStructure"] = sts.parameters.MessageStructure
	}
	if sts.parameters.PhoneNumber != "" {
		parameterMap["PhoneNumber"] = sts.parameters.PhoneNumber
	}
	if sts.parameters.Subject != "" {
		parameterMap["Subject"] = sts.parameters.Subject
	}
	if sts.parameters.TargetArn != nil {
		parameterMap["TargetArn"] = sts.parameters.TargetArn
	}
	additionalParams["Parameters"] = parameterMap
	return sts.marshalStateJSON("Task", additionalParams)

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
