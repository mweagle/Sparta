package step

import (
	"math/rand"
)

// TaskState represents bindings for
// https://states-language.net/#task-state
type TaskState struct {
	BaseTask
	resourceURI string
	parameters  map[string]interface{}
}

// MarshalJSON for custom marshalling, since this will be stringified and we need it
// to turn into a stringified
func (ts *TaskState) MarshalJSON() ([]byte, error) {
	return ts.BaseTask.marshalMergedParams(ts.resourceURI,
		&ts.parameters)
}

// NewTaskState returns an initialized TaskState. A Task State MUST include
// a "Resource" field, whose value MUST be a URI that uniquely identifies the
// specific task to execute.
// The States language does not constrain the URI scheme nor any other part
// of the URI.
func NewTaskState(stateName string,
	resourceURI string,
	parameters map[string]interface{}) *TaskState {

	sns := &TaskState{
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
