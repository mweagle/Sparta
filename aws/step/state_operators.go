package step

import (
	"encoding/json"
)

// Comparison is the generic comparison operator interface
type Comparison interface {
	json.Marshaler
}

// ChoiceBranch represents a type for a ChoiceState "Choices" entry
type ChoiceBranch interface {
	nextState() MachineState
}

/*******************************************************************************
   ___  ___ ___ ___    _ _____ ___  ___  ___
  / _ \| _ \ __| _ \  /_\_   _/ _ \| _ \/ __|
 | (_) |  _/ _||   / / _ \| || (_) |   /\__ \
  \___/|_| |___|_|_\/_/ \_\_| \___/|_|_\|___/
/******************************************************************************/

////////////////////////////////////////////////////////////////////////////////
// And
////////////////////////////////////////////////////////////////////////////////

// And operator
type And struct {
	ChoiceBranch
	Comparison []Comparison
	Next       MachineState
}

func (andOperation *And) nextState() MachineState {
	return andOperation.Next
}

// MarshalJSON for custom marshalling
func (andOperation *And) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Comparison []Comparison `json:"And,omitempty"`
		Next       string       `json:",omitempty"`
	}{
		Comparison: andOperation.Comparison,
		Next:       andOperation.Next.Name(),
	})
}

////////////////////////////////////////////////////////////////////////////////
// Or
////////////////////////////////////////////////////////////////////////////////

// Or operator
type Or struct {
	ChoiceBranch
	Comparison []Comparison
	Next       MachineState
}

func (orOperation *Or) nextState() MachineState {
	return orOperation.Next
}

// MarshalJSON for custom marshalling
func (orOperation *Or) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Comparison []Comparison `json:"Or,omitempty"`
		Next       string       `json:",omitempty"`
	}{
		Comparison: orOperation.Comparison,
		Next:       orOperation.Next.Name(),
	})
}

////////////////////////////////////////////////////////////////////////////////
// Not
////////////////////////////////////////////////////////////////////////////////

// Not operator
type Not struct {
	ChoiceBranch
	Comparison Comparison
	Next       MachineState
}

func (notOperation *Not) nextState() MachineState {
	return notOperation.Next
}

// MarshalJSON for custom marshalling
func (notOperation *Not) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Not  Comparison
		Next string
	}{
		Not:  notOperation.Comparison,
		Next: notOperation.Next.Name(),
	})
}
