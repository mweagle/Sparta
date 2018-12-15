package step

import (
	"math/rand"

	gocf "github.com/mweagle/go-cloudformation"
)

// BatchTaskParameters represents params for the Batch notification
// Ref: https://docs.aws.amazon.com/step-functions/latest/dg/connectors-batch.html
type BatchTaskParameters struct {
	JobDefinition      gocf.Stringable          `json:",omitempty"`
	JobName            string                   `json:",omitempty"`
	JobQueue           gocf.Stringable          `json:",omitempty"`
	ArrayProperties    map[string]interface{}   `json:",omitempty"`
	ContainerOverrides map[string]interface{}   `json:",omitempty"`
	DependsOn          []map[string]interface{} `json:",omitempty"`
	Parameters         map[string]string        `json:",omitempty"`
	RetryStrategy      map[string]interface{}   `json:",omitempty"`
	Timeout            map[string]interface{}   `json:",omitempty"`
}

// BatchTaskState represents bindings for
// https://docs.aws.amazon.com/step-functions/latest/dg/connectors-batch.html
type BatchTaskState struct {
	BaseTask
	parameters BatchTaskParameters
}

// MarshalJSON for custom marshalling, since this will be stringified and we need it
// to turn into a stringified
// Ref: https://docs.aws.amazon.com/step-functions/latest/dg/connectors-batch.html
func (bts *BatchTaskState) MarshalJSON() ([]byte, error) {
	return bts.BaseTask.marshalMergedParams("arn:aws:states:::batch:submitJob.sync",
		&bts.parameters)
}

// NewBatchTaskState returns an initialized BatchTaskState
func NewBatchTaskState(stateName string,
	parameters BatchTaskParameters) *BatchTaskState {

	sns := &BatchTaskState{
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
