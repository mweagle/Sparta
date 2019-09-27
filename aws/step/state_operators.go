package step

import (
	"encoding/json"
	"time"
)

/*******************************************************************************
   ___ ___  __  __ ___  _   ___ ___ ___  ___  _  _ ___
  / __/ _ \|  \/  | _ \/_\ | _ \_ _/ __|/ _ \| \| / __|
 | (_| (_) | |\/| |  _/ _ \|   /| |\__ \ (_) | .` \__ \
  \___\___/|_|  |_|_|/_/ \_\_|_\___|___/\___/|_|\_|___/

/******************************************************************************/

// Comparison is the generic comparison operator interface
type Comparison interface {
	json.Marshaler
}

// ChoiceBranch represents a type for a ChoiceState "Choices" entry
type ChoiceBranch interface {
	nextState() MachineState
}

////////////////////////////////////////////////////////////////////////////////
// StringEquals
////////////////////////////////////////////////////////////////////////////////

/**

Validations
	- JSONPath: https://github.com/NodePrime/jsonpath
	- Choices lead to existing states
	- Choice statenames are scoped to same depth
*/

// StringEquals comparison
type StringEquals struct {
	Comparison
	Variable string
	Value    string
}

// MarshalJSON for custom marshalling
func (cmp *StringEquals) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Variable     string
		StringEquals string
	}{
		Variable:     cmp.Variable,
		StringEquals: cmp.Value,
	})
}

////////////////////////////////////////////////////////////////////////////////
// StringLessThan
////////////////////////////////////////////////////////////////////////////////

// StringLessThan comparison
type StringLessThan struct {
	Comparison
	Variable string
	Value    string
}

// MarshalJSON for custom marshalling
func (cmp *StringLessThan) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Variable       string
		StringLessThan string
	}{
		Variable:       cmp.Variable,
		StringLessThan: cmp.Value,
	})
}

////////////////////////////////////////////////////////////////////////////////
// StringGreaterThan
////////////////////////////////////////////////////////////////////////////////

// StringGreaterThan comparison
type StringGreaterThan struct {
	Comparison
	Variable string
	Value    string
}

// MarshalJSON for custom marshalling
func (cmp *StringGreaterThan) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Variable          string
		StringGreaterThan string
	}{
		Variable:          cmp.Variable,
		StringGreaterThan: cmp.Value,
	})
}

////////////////////////////////////////////////////////////////////////////////
// StringLessThanEquals
////////////////////////////////////////////////////////////////////////////////

// StringLessThanEquals comparison
type StringLessThanEquals struct {
	Variable string
	Value    string
}

// MarshalJSON for custom marshalling
func (cmp *StringLessThanEquals) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Variable             string
		StringLessThanEquals string
	}{
		Variable:             cmp.Variable,
		StringLessThanEquals: cmp.Value,
	})
}

////////////////////////////////////////////////////////////////////////////////
// StringGreaterThanEquals
////////////////////////////////////////////////////////////////////////////////

// StringGreaterThanEquals comparison
type StringGreaterThanEquals struct {
	Comparison
	Variable string
	Value    string
}

// MarshalJSON for custom marshalling
func (cmp *StringGreaterThanEquals) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Variable                string
		StringGreaterThanEquals string
	}{
		Variable:                cmp.Variable,
		StringGreaterThanEquals: cmp.Value,
	})
}

////////////////////////////////////////////////////////////////////////////////
// NumericEquals
////////////////////////////////////////////////////////////////////////////////

// NumericEquals comparison
type NumericEquals struct {
	Comparison
	Variable string
	Value    int64
}

// MarshalJSON for custom marshalling
func (cmp *NumericEquals) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Variable      string
		NumericEquals int64
	}{
		Variable:      cmp.Variable,
		NumericEquals: cmp.Value,
	})
}

////////////////////////////////////////////////////////////////////////////////
// NumericLessThan
////////////////////////////////////////////////////////////////////////////////

// NumericLessThan comparison
type NumericLessThan struct {
	Comparison
	Variable string
	Value    int64
}

// MarshalJSON for custom marshalling
func (cmp *NumericLessThan) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Variable        string
		NumericLessThan int64
	}{
		Variable:        cmp.Variable,
		NumericLessThan: cmp.Value,
	})
}

////////////////////////////////////////////////////////////////////////////////
// NumericGreaterThan
////////////////////////////////////////////////////////////////////////////////

// NumericGreaterThan comparison
type NumericGreaterThan struct {
	Comparison
	Variable string
	Value    int64
}

// MarshalJSON for custom marshalling
func (cmp *NumericGreaterThan) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Variable           string
		NumericGreaterThan int64
	}{
		Variable:           cmp.Variable,
		NumericGreaterThan: cmp.Value,
	})
}

////////////////////////////////////////////////////////////////////////////////
// NumericLessThanEquals
////////////////////////////////////////////////////////////////////////////////

// NumericLessThanEquals comparison
type NumericLessThanEquals struct {
	Comparison
	Variable string
	Value    int64
}

// MarshalJSON for custom marshalling
func (cmp *NumericLessThanEquals) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Variable              string
		NumericLessThanEquals int64
	}{
		Variable:              cmp.Variable,
		NumericLessThanEquals: cmp.Value,
	})
}

////////////////////////////////////////////////////////////////////////////////
// NumericGreaterThanEquals
////////////////////////////////////////////////////////////////////////////////

// NumericGreaterThanEquals comparison
type NumericGreaterThanEquals struct {
	Comparison
	Variable string
	Value    int64
}

// MarshalJSON for custom marshalling
func (cmp *NumericGreaterThanEquals) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Variable                 string
		NumericGreaterThanEquals int64
	}{
		Variable:                 cmp.Variable,
		NumericGreaterThanEquals: cmp.Value,
	})
}

////////////////////////////////////////////////////////////////////////////////
// BooleanEquals
////////////////////////////////////////////////////////////////////////////////

// BooleanEquals comparison
type BooleanEquals struct {
	Comparison
	Variable string
	Value    interface{}
}

// MarshalJSON for custom marshalling
func (cmp *BooleanEquals) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Variable      string
		BooleanEquals interface{}
	}{
		Variable:      cmp.Variable,
		BooleanEquals: cmp.Value,
	})
}

////////////////////////////////////////////////////////////////////////////////
// TimestampEquals
////////////////////////////////////////////////////////////////////////////////

// TimestampEquals comparison
type TimestampEquals struct {
	Comparison
	Variable string
	Value    time.Time
}

// MarshalJSON for custom marshalling
func (cmp *TimestampEquals) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Variable        string
		TimestampEquals string
	}{
		Variable:        cmp.Variable,
		TimestampEquals: cmp.Value.Format(time.RFC3339),
	})
}

////////////////////////////////////////////////////////////////////////////////
// TimestampLessThan
////////////////////////////////////////////////////////////////////////////////

// TimestampLessThan comparison
type TimestampLessThan struct {
	Comparison
	Variable string
	Value    time.Time
}

// MarshalJSON for custom marshalling
func (cmp *TimestampLessThan) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Variable          string
		TimestampLessThan string
	}{
		Variable:          cmp.Variable,
		TimestampLessThan: cmp.Value.Format(time.RFC3339),
	})
}

////////////////////////////////////////////////////////////////////////////////
// TimestampGreaterThan
////////////////////////////////////////////////////////////////////////////////

// TimestampGreaterThan comparison
type TimestampGreaterThan struct {
	Variable string
	Value    time.Time
}

// MarshalJSON for custom marshalling
func (cmp *TimestampGreaterThan) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Variable             string
		TimestampGreaterThan string
	}{
		Variable:             cmp.Variable,
		TimestampGreaterThan: cmp.Value.Format(time.RFC3339),
	})
}

////////////////////////////////////////////////////////////////////////////////
// TimestampLessThanEquals
////////////////////////////////////////////////////////////////////////////////

// TimestampLessThanEquals comparison
type TimestampLessThanEquals struct {
	Comparison
	Variable string
	Value    time.Time
}

// MarshalJSON for custom marshalling
func (cmp *TimestampLessThanEquals) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Variable                string
		TimestampLessThanEquals string
	}{
		Variable:                cmp.Variable,
		TimestampLessThanEquals: cmp.Value.Format(time.RFC3339),
	})
}

////////////////////////////////////////////////////////////////////////////////
// TimestampGreaterThanEquals
////////////////////////////////////////////////////////////////////////////////

// TimestampGreaterThanEquals comparison
type TimestampGreaterThanEquals struct {
	Comparison
	Variable string
	Value    time.Time
}

// MarshalJSON for custom marshalling
func (cmp *TimestampGreaterThanEquals) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Variable                   string
		TimestampGreaterThanEquals string
	}{
		Variable:                   cmp.Variable,
		TimestampGreaterThanEquals: cmp.Value.Format(time.RFC3339),
	})
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
