package step

import (
	"math/rand"

	gocf "github.com/mweagle/go-cloudformation"
)

// BatchArrayProperties is how long it takes to
type BatchArrayProperties struct {
	Size int
}

// BatchContainerOverrides stores AWS Batch override info
type BatchContainerOverrides struct {
	Command      []string
	Environment  map[string]string
	InstanceType string
}

// BatchDependsOn is an entry for Depends
type BatchDependsOn struct {
	JobID string `json:"JobId"`
	Type  string
}

// BatchRetryStrategy is the retry strategy
type BatchRetryStrategy struct {
	Attempts int
}

// BatchTimeout is how long it takes to
type BatchTimeout struct {
	AttemptDurationSeconds int
}

// BatchTaskParameters represents params for the Batch notification
// Ref: https://docs.aws.amazon.com/step-functions/latest/dg/connectors-batch.html
type BatchTaskParameters struct {
	JobDefinition      gocf.Stringable
	JobName            string
	JobQueue           gocf.Stringable
	ArrayProperties    *BatchArrayProperties
	ContainerOverrides *BatchContainerOverrides
	DependsOn          []*BatchDependsOn
	Parameters         map[string]string
	RetryStrategy      *BatchRetryStrategy
	Timeout            *BatchTimeout
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
	additionalParams := bts.BaseTask.additionalParams()

	additionalParams["Resource"] = "arn:aws:states:::batch:submitJob.sync"
	parameterMap := map[string]interface{}{}
	if bts.parameters.JobDefinition != nil {
		parameterMap["JobDefinition"] = bts.parameters.JobDefinition
	}
	if bts.parameters.JobName != "" {
		parameterMap["JobName"] = bts.parameters.JobName
	}
	if bts.parameters.JobQueue != nil {
		parameterMap["JobQueue"] = bts.parameters.JobQueue
	}
	if bts.parameters.ArrayProperties != nil {
		parameterMap["ArrayProperties"] = bts.parameters.ArrayProperties
	}
	if bts.parameters.ContainerOverrides != nil {
		parameterMap["ContainerOverrides"] = bts.parameters.ContainerOverrides
	}
	if bts.parameters.DependsOn != nil {
		parameterMap["DependsOn"] = bts.parameters.DependsOn
	}
	if bts.parameters.Parameters != nil {
		parameterMap["Parameters"] = bts.parameters.Parameters
	}
	if bts.parameters.RetryStrategy != nil {
		parameterMap["RetryStrategy"] = bts.parameters.RetryStrategy
	}
	if bts.parameters.Timeout != nil {
		parameterMap["Timeout"] = bts.parameters.Timeout
	}
	additionalParams["Parameters"] = parameterMap
	return bts.marshalStateJSON("Task", additionalParams)
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
