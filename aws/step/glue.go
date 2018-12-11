package step

import (
	"math/rand"

	gocf "github.com/mweagle/go-cloudformation"
)

// GlueParameters represents params for Glue step
// Ref: https://docs.aws.amazon.com/step-functions/latest/dg/connectors-glue.html
type GlueParameters struct {
	JobName               gocf.Stringable
	JobRunID              string `json:"JobRunId"`
	Arguments             map[string]interface{}
	AllocatedCapacity     *gocf.IntegerExpr
	Timeout               *gocf.IntegerExpr
	SecurityConfiguration gocf.Stringable
	NotificationProperty  interface{}
}

// GlueState represents bindings for
// https://docs.aws.amazon.com/step-functions/latest/dg/connectors-sns.html
type GlueState struct {
	BaseTask
	parameters GlueParameters
}

// MarshalJSON for custom marshalling, since this will be stringified and we need it
// to turn into a stringified
// Ref: https://docs.aws.amazon.com/step-functions/latest/dg/connectors-sns.html
func (gs *GlueState) MarshalJSON() ([]byte, error) {
	additionalParams := gs.BaseTask.additionalParams()
	additionalParams["Resource"] = "arn:aws:states:::glue:startJobRun.sync"

	parameterMap := map[string]interface{}{}
	if gs.parameters.JobName != nil {
		parameterMap["JobName"] = gs.parameters.JobName
	}
	if gs.parameters.JobRunID != "" {
		parameterMap["JobRunId"] = gs.parameters.JobRunID
	}
	if gs.parameters.Arguments != nil {
		parameterMap["Arguments"] = gs.parameters.Arguments
	}
	if gs.parameters.AllocatedCapacity != nil {
		parameterMap["AllocatedCapacity"] = gs.parameters.AllocatedCapacity
	}
	if gs.parameters.Timeout != nil {
		parameterMap["Timeout"] = gs.parameters.Timeout
	}
	if gs.parameters.SecurityConfiguration != nil {
		parameterMap["SecurityConfiguration"] = gs.parameters.SecurityConfiguration
	}
	if gs.parameters.NotificationProperty != nil {
		parameterMap["NotificationProperty"] = gs.parameters.NotificationProperty
	}
	additionalParams["Parameters"] = parameterMap
	return gs.marshalStateJSON("Task", additionalParams)
}

// NewGlueState returns an initialized GlueState
func NewGlueState(stateName string,
	parameters GlueParameters) *GlueState {

	sns := &GlueState{
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
