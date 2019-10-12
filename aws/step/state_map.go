package step

import (
	"math/rand"
)

////////////////////////////////////////////////////////////////////////////////
// MapState
////////////////////////////////////////////////////////////////////////////////

/*
"Validate-All": {
  "Type": "Map",
  "InputPath": "$.detail",
  "ItemsPath": "$.shipped",
  "MaxConcurrency": 0,
  "Iterator": {
    "StartAt": "Validate",
    "States": {
      "Validate": {
        "Type": "Task",
        "Resource": "arn:aws:lambda:us-east-1:123456789012:function:ship-val",
        "End": true
      }
    }
  },
  "ResultPath": "$.detail.shipped",
  "End": true
}
*/

// MapState is a synthetic state that executes a dynamically determined set
// of nodes in parallel
type MapState struct {
	baseInnerState
	States         *StateMachine
	Parameters     map[string]interface{}
	ResultPath     string
	ItemsPath      string // optional
	MaxConcurrency int    //optional
	Retriers       []*TaskRetry
	Catchers       []*TaskCatch
}

// WithResultPath is the fluent builder for the result path
func (ms *MapState) WithResultPath(resultPath string) *MapState {
	ms.ResultPath = resultPath
	return ms
}

// WithRetriers is the fluent builder for TaskState
func (ms *MapState) WithRetriers(retries ...*TaskRetry) *MapState {
	if ms.Retriers == nil {
		ms.Retriers = make([]*TaskRetry, 0)
	}
	ms.Retriers = append(ms.Retriers, retries...)
	return ms
}

// WithCatchers is the fluent builder for TaskState
func (ms *MapState) WithCatchers(catch ...*TaskCatch) *MapState {
	if ms.Catchers == nil {
		ms.Catchers = make([]*TaskCatch, 0)
	}
	ms.Catchers = append(ms.Catchers, catch...)
	return ms
}

// Next returns the next state
func (ms *MapState) Next(nextState MachineState) MachineState {
	ms.next = nextState
	return nextState
}

// AdjacentStates returns nodes reachable from this node
func (ms *MapState) AdjacentStates() []MachineState {
	if ms.next == nil {
		return nil
	}
	return []MachineState{ms.next}
}

// Name returns the name of this Task state
func (ms *MapState) Name() string {
	return ms.name
}

// WithComment returns the TaskState comment
func (ms *MapState) WithComment(comment string) TransitionState {
	ms.comment = comment
	return ms
}

// WithInputPath returns the TaskState input data selector
func (ms *MapState) WithInputPath(inputPath string) TransitionState {
	ms.inputPath = inputPath
	return ms
}

// WithOutputPath returns the TaskState output data selector
func (ms *MapState) WithOutputPath(outputPath string) TransitionState {
	ms.outputPath = outputPath
	return ms
}

// MarshalJSON for custom marshalling
func (ms *MapState) MarshalJSON() ([]byte, error) {
	/*
		A Map State MUST contain an object field named “Iterator” which MUST
		contain fields named “States” and “StartAt”, whose meanings are exactly
		like those in the top level of a State Machine.

		A state in the “States” field of an “Iterator” field MUST NOT have a
		“Next” field that targets a field outside of that “States” field. A state
		MUST NOT have a “Next” field which matches a state name inside an
		“Iterator” field’s “States” field unless it is also inside the same
		“States” field.

		Put another way, states in an Iterator’s “States” field can transition
		only to each other, and no state outside of that “States” field can
		transition into it.
	*/
	// Don't marshal the "End" flag
	ms.States.disableEndState = true
	additionalParams := make(map[string]interface{})
	additionalParams["Iterator"] = ms.States
	if ms.ItemsPath != "" {
		additionalParams["ItemsPath"] = ms.ItemsPath
	}
	if ms.ResultPath != "" {
		additionalParams["ResultPath"] = ms.ResultPath
	}
	if len(ms.Retriers) != 0 {
		additionalParams["Retry"] = ms.Retriers
	}
	if ms.Catchers != nil {
		additionalParams["Catch"] = ms.Catchers
	}
	if ms.Parameters != nil {
		additionalParams["Parameters"] = ms.Parameters
	}

	return ms.marshalStateJSON("Map", additionalParams)
}

// NewMapState returns a "MapState" with the supplied
// information
func NewMapState(parallelStateName string, states *StateMachine) *MapState {
	return &MapState{
		baseInnerState: baseInnerState{
			name: parallelStateName,
			id:   rand.Int63(),
		},
		States: states,
	}
}
