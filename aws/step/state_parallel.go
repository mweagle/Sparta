package step

import (
	"math/rand"
)

////////////////////////////////////////////////////////////////////////////////
// ParallelState
////////////////////////////////////////////////////////////////////////////////

// ParallelState is a synthetic state that executes a lot of independent
// branches in parallel
type ParallelState struct {
	baseInnerState
	Branches   []*StateMachine
	Parameters map[string]interface{}
	ResultPath string
	Retriers   []*TaskRetry
	Catchers   []*TaskCatch
}

// WithResultPath is the fluent builder for the result path
func (ps *ParallelState) WithResultPath(resultPath string) *ParallelState {
	ps.ResultPath = resultPath
	return ps
}

// WithRetriers is the fluent builder for TaskState
func (ps *ParallelState) WithRetriers(retries ...*TaskRetry) *ParallelState {
	if ps.Retriers == nil {
		ps.Retriers = make([]*TaskRetry, 0)
	}
	ps.Retriers = append(ps.Retriers, retries...)
	return ps
}

// WithCatchers is the fluent builder for TaskState
func (ps *ParallelState) WithCatchers(catch ...*TaskCatch) *ParallelState {
	if ps.Catchers == nil {
		ps.Catchers = make([]*TaskCatch, 0)
	}
	ps.Catchers = append(ps.Catchers, catch...)
	return ps
}

// Next returns the next state
func (ps *ParallelState) Next(nextState MachineState) MachineState {
	ps.next = nextState
	return nextState
}

// AdjacentStates returns nodes reachable from this node
func (ps *ParallelState) AdjacentStates() []MachineState {
	if ps.next == nil {
		return nil
	}
	return []MachineState{ps.next}
}

// Name returns the name of this Task state
func (ps *ParallelState) Name() string {
	return ps.name
}

// WithComment returns the TaskState comment
func (ps *ParallelState) WithComment(comment string) TransitionState {
	ps.comment = comment
	return ps
}

// WithInputPath returns the TaskState input data selector
func (ps *ParallelState) WithInputPath(inputPath string) TransitionState {
	ps.inputPath = inputPath
	return ps
}

// WithOutputPath returns the TaskState output data selector
func (ps *ParallelState) WithOutputPath(outputPath string) TransitionState {
	ps.outputPath = outputPath
	return ps
}

// MarshalJSON for custom marshalling
func (ps *ParallelState) MarshalJSON() ([]byte, error) {
	/*
		A state in a Parallel state branch “States” field MUST NOT have a “Next” field that targets a field outside of that “States” field. A state MUST NOT have a “Next” field which matches a state name inside a Parallel state branch’s “States” field unless it is also inside the same “States” field.

		Put another way, states in a branch’s “States” field can transition only to each other, and no state outside of that “States” field can transition into it.
	*/
	additionalParams := map[string]interface{}{
		"Branches": ps.Branches,
	}
	if ps.ResultPath != "" {
		additionalParams["ResultPath"] = ps.ResultPath
	}
	if len(ps.Retriers) != 0 {
		additionalParams["Retry"] = ps.Retriers
	}
	if ps.Catchers != nil {
		additionalParams["Catch"] = ps.Catchers
	}
	if ps.Parameters != nil {
		additionalParams["Parameters"] = ps.Parameters
	}
	return ps.marshalStateJSON("Parallel", additionalParams)
}

// NewParallelState returns a "ParallelState" with the supplied
// information
func NewParallelState(parallelStateName string, branches ...*StateMachine) *ParallelState {
	return &ParallelState{
		baseInnerState: baseInnerState{
			name: parallelStateName,
			id:   rand.Int63(),
		},
		Branches: branches,
	}
}
