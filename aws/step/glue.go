package step

import (
	"math/rand"
)

// GlueParameters represents params for Glue step
// Ref: https://docs.aws.amazon.com/step-functions/latest/dg/connectors-glue.html
type GlueParameters struct {
	JobName               string                 `json:",omitempty"`
	JobRunID              string                 `json:"JobRunId,omitempty"`
	Arguments             map[string]interface{} `json:",omitempty"`
	AllocatedCapacity     int                    `json:",omitempty"`
	Timeout               int                    `json:",omitempty"`
	SecurityConfiguration string                 `json:",omitempty"`
	NotificationProperty  interface{}            `json:",omitempty"`
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
	return gs.BaseTask.marshalMergedParams("arn:aws:states:::glue:startJobRun.sync",
		&gs.parameters)
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
