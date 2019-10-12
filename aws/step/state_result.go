package step

import (
	"math/rand"
)

////////////////////////////////////////////////////////////////////////////////
// SuccessState
////////////////////////////////////////////////////////////////////////////////

// SuccessState represents the end of the state machine
type SuccessState struct {
	baseInnerState
}

// Name returns the WaitDelay name
func (ss *SuccessState) Name() string {
	return ss.name
}

// Next sets the step after the wait delay
func (ss *SuccessState) Next(nextState MachineState) MachineState {
	ss.next = nextState
	return ss
}

// AdjacentStates returns nodes reachable from this node
func (ss *SuccessState) AdjacentStates() []MachineState {
	if ss.next == nil {
		return nil
	}
	return []MachineState{ss.next}
}

// WithComment returns the WaitDelay comment
func (ss *SuccessState) WithComment(comment string) TransitionState {
	ss.comment = comment
	return ss
}

// WithInputPath returns the TaskState input data selector
func (ss *SuccessState) WithInputPath(inputPath string) TransitionState {
	ss.inputPath = inputPath
	return ss
}

// WithOutputPath returns the TaskState output data selector
func (ss *SuccessState) WithOutputPath(outputPath string) TransitionState {
	ss.outputPath = outputPath
	return ss
}

// MarshalJSON for custom marshalling
func (ss *SuccessState) MarshalJSON() ([]byte, error) {
	return ss.marshalStateJSON("Succeed", nil)
}

// NewSuccessState returns a "SuccessState" with the supplied
// name
func NewSuccessState(name string) *SuccessState {
	return &SuccessState{
		baseInnerState: baseInnerState{
			name:              name,
			id:                rand.Int63(),
			isEndStateInvalid: true,
		},
	}
}

////////////////////////////////////////////////////////////////////////////////

// FailState represents the end of state machine
type FailState struct {
	baseInnerState
	ErrorName string
	Cause     error
}

// Name returns the WaitDelay name
func (fs *FailState) Name() string {
	return fs.name
}

// Next sets the step after the wait delay
func (fs *FailState) Next(nextState MachineState) MachineState {
	return fs
}

// AdjacentStates returns nodes reachable from this node
func (fs *FailState) AdjacentStates() []MachineState {
	return nil
}

// WithComment returns the WaitDelay comment
func (fs *FailState) WithComment(comment string) TransitionState {
	fs.comment = comment
	return fs
}

// WithInputPath returns the TaskState input data selector
func (fs *FailState) WithInputPath(inputPath string) TransitionState {
	return fs
}

// WithOutputPath returns the TaskState output data selector
func (fs *FailState) WithOutputPath(outputPath string) TransitionState {
	return fs
}

// MarshalJSON for custom marshaling
func (fs *FailState) MarshalJSON() ([]byte, error) {
	additionalParams := make(map[string]interface{})
	additionalParams["Error"] = fs.ErrorName
	if fs.Cause != nil {
		additionalParams["Cause"] = fs.Cause.Error()
	}
	return fs.marshalStateJSON("Fail", additionalParams)
}

// NewFailState returns a "FailState" with the supplied
// information
func NewFailState(failStateName string, errorName string, cause error) *FailState {
	return &FailState{
		baseInnerState: baseInnerState{
			name:              failStateName,
			id:                rand.Int63(),
			isEndStateInvalid: true,
		},
		ErrorName: errorName,
		Cause:     cause,
	}
}
