// Code generated by github.com/mweagle/Sparta/aws/step/generator/main.go. DO NOT EDIT.

package step

import (
	"encoding/json"
	"time"
)

/*******************************************************************************
   ___ ___  __  __ ___  _   ___ ___ ___  ___  _  _ ___
  / __/ _ \|  \/  | _ \/_\ | _ \_ _/ __|/ _ \| \| / __|
 | (_| (_) | |\/| |  _/ _ \|   /| |\__ \ (_) |    \__ \
  \___\___/|_|  |_|_|/_/ \_\_|_\___|___/\___/|_|\_|___/

/******************************************************************************/

// For path based selectors see the
// JSONPath: https://github.com/NodePrime/jsonpath
// documentation

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
// BooleanEqualsPath
////////////////////////////////////////////////////////////////////////////////

// BooleanEqualsPath comparison
type BooleanEqualsPath struct {
	Comparison
	Variable string
	Value    string
}

// MarshalJSON for custom marshalling
func (cmp *BooleanEqualsPath) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Variable          string
		BooleanEqualsPath string
	}{
		Variable:          cmp.Variable,
		BooleanEqualsPath: cmp.Value,
	})
}

////////////////////////////////////////////////////////////////////////////////
// IsBoolean
////////////////////////////////////////////////////////////////////////////////

// IsBoolean comparison
type IsBoolean struct {
	Comparison
	Variable string
	Value    string
}

// MarshalJSON for custom marshalling
func (cmp *IsBoolean) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Variable  string
		IsBoolean string
	}{
		Variable:  cmp.Variable,
		IsBoolean: cmp.Value,
	})
}

////////////////////////////////////////////////////////////////////////////////
// IsNull
////////////////////////////////////////////////////////////////////////////////

// IsNull comparison
type IsNull struct {
	Comparison
	Variable string
	Value    string
}

// MarshalJSON for custom marshalling
func (cmp *IsNull) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Variable string
		IsNull   string
	}{
		Variable: cmp.Variable,
		IsNull:   cmp.Value,
	})
}

////////////////////////////////////////////////////////////////////////////////
// IsNumeric
////////////////////////////////////////////////////////////////////////////////

// IsNumeric comparison
type IsNumeric struct {
	Comparison
	Variable string
	Value    string
}

// MarshalJSON for custom marshalling
func (cmp *IsNumeric) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Variable  string
		IsNumeric string
	}{
		Variable:  cmp.Variable,
		IsNumeric: cmp.Value,
	})
}

////////////////////////////////////////////////////////////////////////////////
// IsPresent
////////////////////////////////////////////////////////////////////////////////

// IsPresent comparison
type IsPresent struct {
	Comparison
	Variable string
	Value    string
}

// MarshalJSON for custom marshalling
func (cmp *IsPresent) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Variable  string
		IsPresent string
	}{
		Variable:  cmp.Variable,
		IsPresent: cmp.Value,
	})
}

////////////////////////////////////////////////////////////////////////////////
// IsString
////////////////////////////////////////////////////////////////////////////////

// IsString comparison
type IsString struct {
	Comparison
	Variable string
	Value    string
}

// MarshalJSON for custom marshalling
func (cmp *IsString) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Variable string
		IsString string
	}{
		Variable: cmp.Variable,
		IsString: cmp.Value,
	})
}

////////////////////////////////////////////////////////////////////////////////
// IsTimestamp
////////////////////////////////////////////////////////////////////////////////

// IsTimestamp comparison
type IsTimestamp struct {
	Comparison
	Variable string
	Value    string
}

// MarshalJSON for custom marshalling
func (cmp *IsTimestamp) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Variable    string
		IsTimestamp string
	}{
		Variable:    cmp.Variable,
		IsTimestamp: cmp.Value,
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
// NumericEqualsPath
////////////////////////////////////////////////////////////////////////////////

// NumericEqualsPath comparison
type NumericEqualsPath struct {
	Comparison
	Variable string
	Value    string
}

// MarshalJSON for custom marshalling
func (cmp *NumericEqualsPath) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Variable          string
		NumericEqualsPath string
	}{
		Variable:          cmp.Variable,
		NumericEqualsPath: cmp.Value,
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
// NumericGreaterThanEqualsPath
////////////////////////////////////////////////////////////////////////////////

// NumericGreaterThanEqualsPath comparison
type NumericGreaterThanEqualsPath struct {
	Comparison
	Variable string
	Value    string
}

// MarshalJSON for custom marshalling
func (cmp *NumericGreaterThanEqualsPath) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Variable                     string
		NumericGreaterThanEqualsPath string
	}{
		Variable:                     cmp.Variable,
		NumericGreaterThanEqualsPath: cmp.Value,
	})
}

////////////////////////////////////////////////////////////////////////////////
// NumericGreaterThanPath
////////////////////////////////////////////////////////////////////////////////

// NumericGreaterThanPath comparison
type NumericGreaterThanPath struct {
	Comparison
	Variable string
	Value    string
}

// MarshalJSON for custom marshalling
func (cmp *NumericGreaterThanPath) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Variable               string
		NumericGreaterThanPath string
	}{
		Variable:               cmp.Variable,
		NumericGreaterThanPath: cmp.Value,
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
// NumericLessThanEqualsPath
////////////////////////////////////////////////////////////////////////////////

// NumericLessThanEqualsPath comparison
type NumericLessThanEqualsPath struct {
	Comparison
	Variable string
	Value    string
}

// MarshalJSON for custom marshalling
func (cmp *NumericLessThanEqualsPath) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Variable                  string
		NumericLessThanEqualsPath string
	}{
		Variable:                  cmp.Variable,
		NumericLessThanEqualsPath: cmp.Value,
	})
}

////////////////////////////////////////////////////////////////////////////////
// NumericLessThanPath
////////////////////////////////////////////////////////////////////////////////

// NumericLessThanPath comparison
type NumericLessThanPath struct {
	Comparison
	Variable string
	Value    string
}

// MarshalJSON for custom marshalling
func (cmp *NumericLessThanPath) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Variable            string
		NumericLessThanPath string
	}{
		Variable:            cmp.Variable,
		NumericLessThanPath: cmp.Value,
	})
}

////////////////////////////////////////////////////////////////////////////////
// StringEquals
////////////////////////////////////////////////////////////////////////////////

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
// StringEqualsPath
////////////////////////////////////////////////////////////////////////////////

// StringEqualsPath comparison
type StringEqualsPath struct {
	Comparison
	Variable string
	Value    string
}

// MarshalJSON for custom marshalling
func (cmp *StringEqualsPath) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Variable         string
		StringEqualsPath string
	}{
		Variable:         cmp.Variable,
		StringEqualsPath: cmp.Value,
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
// StringGreaterThanEqualsPath
////////////////////////////////////////////////////////////////////////////////

// StringGreaterThanEqualsPath comparison
type StringGreaterThanEqualsPath struct {
	Comparison
	Variable string
	Value    string
}

// MarshalJSON for custom marshalling
func (cmp *StringGreaterThanEqualsPath) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Variable                    string
		StringGreaterThanEqualsPath string
	}{
		Variable:                    cmp.Variable,
		StringGreaterThanEqualsPath: cmp.Value,
	})
}

////////////////////////////////////////////////////////////////////////////////
// StringGreaterThanPath
////////////////////////////////////////////////////////////////////////////////

// StringGreaterThanPath comparison
type StringGreaterThanPath struct {
	Comparison
	Variable string
	Value    string
}

// MarshalJSON for custom marshalling
func (cmp *StringGreaterThanPath) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Variable              string
		StringGreaterThanPath string
	}{
		Variable:              cmp.Variable,
		StringGreaterThanPath: cmp.Value,
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
// StringLessThanEquals
////////////////////////////////////////////////////////////////////////////////

// StringLessThanEquals comparison
type StringLessThanEquals struct {
	Comparison
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
// StringLessThanEqualsPath
////////////////////////////////////////////////////////////////////////////////

// StringLessThanEqualsPath comparison
type StringLessThanEqualsPath struct {
	Comparison
	Variable string
	Value    string
}

// MarshalJSON for custom marshalling
func (cmp *StringLessThanEqualsPath) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Variable                 string
		StringLessThanEqualsPath string
	}{
		Variable:                 cmp.Variable,
		StringLessThanEqualsPath: cmp.Value,
	})
}

////////////////////////////////////////////////////////////////////////////////
// StringLessThanPath
////////////////////////////////////////////////////////////////////////////////

// StringLessThanPath comparison
type StringLessThanPath struct {
	Comparison
	Variable string
	Value    string
}

// MarshalJSON for custom marshalling
func (cmp *StringLessThanPath) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Variable           string
		StringLessThanPath string
	}{
		Variable:           cmp.Variable,
		StringLessThanPath: cmp.Value,
	})
}

////////////////////////////////////////////////////////////////////////////////
// StringMatches
////////////////////////////////////////////////////////////////////////////////

// StringMatches comparison
type StringMatches struct {
	Comparison
	Variable string
	Value    string
}

// MarshalJSON for custom marshalling
func (cmp *StringMatches) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Variable      string
		StringMatches string
	}{
		Variable:      cmp.Variable,
		StringMatches: cmp.Value,
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
		TimestampEquals time.Time
	}{
		Variable:        cmp.Variable,
		TimestampEquals: cmp.Value,
	})
}

////////////////////////////////////////////////////////////////////////////////
// TimestampEqualsPath
////////////////////////////////////////////////////////////////////////////////

// TimestampEqualsPath comparison
type TimestampEqualsPath struct {
	Comparison
	Variable string
	Value    string
}

// MarshalJSON for custom marshalling
func (cmp *TimestampEqualsPath) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Variable            string
		TimestampEqualsPath string
	}{
		Variable:            cmp.Variable,
		TimestampEqualsPath: cmp.Value,
	})
}

////////////////////////////////////////////////////////////////////////////////
// TimestampGreaterThan
////////////////////////////////////////////////////////////////////////////////

// TimestampGreaterThan comparison
type TimestampGreaterThan struct {
	Comparison
	Variable string
	Value    time.Time
}

// MarshalJSON for custom marshalling
func (cmp *TimestampGreaterThan) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Variable             string
		TimestampGreaterThan time.Time
	}{
		Variable:             cmp.Variable,
		TimestampGreaterThan: cmp.Value,
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
		TimestampGreaterThanEquals time.Time
	}{
		Variable:                   cmp.Variable,
		TimestampGreaterThanEquals: cmp.Value,
	})
}

////////////////////////////////////////////////////////////////////////////////
// TimestampGreaterThanEqualsPath
////////////////////////////////////////////////////////////////////////////////

// TimestampGreaterThanEqualsPath comparison
type TimestampGreaterThanEqualsPath struct {
	Comparison
	Variable string
	Value    string
}

// MarshalJSON for custom marshalling
func (cmp *TimestampGreaterThanEqualsPath) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Variable                       string
		TimestampGreaterThanEqualsPath string
	}{
		Variable:                       cmp.Variable,
		TimestampGreaterThanEqualsPath: cmp.Value,
	})
}

////////////////////////////////////////////////////////////////////////////////
// TimestampGreaterThanPath
////////////////////////////////////////////////////////////////////////////////

// TimestampGreaterThanPath comparison
type TimestampGreaterThanPath struct {
	Comparison
	Variable string
	Value    string
}

// MarshalJSON for custom marshalling
func (cmp *TimestampGreaterThanPath) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Variable                 string
		TimestampGreaterThanPath string
	}{
		Variable:                 cmp.Variable,
		TimestampGreaterThanPath: cmp.Value,
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
		TimestampLessThan time.Time
	}{
		Variable:          cmp.Variable,
		TimestampLessThan: cmp.Value,
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
		TimestampLessThanEquals time.Time
	}{
		Variable:                cmp.Variable,
		TimestampLessThanEquals: cmp.Value,
	})
}

////////////////////////////////////////////////////////////////////////////////
// TimestampLessThanEqualsPath
////////////////////////////////////////////////////////////////////////////////

// TimestampLessThanEqualsPath comparison
type TimestampLessThanEqualsPath struct {
	Comparison
	Variable string
	Value    string
}

// MarshalJSON for custom marshalling
func (cmp *TimestampLessThanEqualsPath) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Variable                    string
		TimestampLessThanEqualsPath string
	}{
		Variable:                    cmp.Variable,
		TimestampLessThanEqualsPath: cmp.Value,
	})
}

////////////////////////////////////////////////////////////////////////////////
// TimestampLessThanPath
////////////////////////////////////////////////////////////////////////////////

// TimestampLessThanPath comparison
type TimestampLessThanPath struct {
	Comparison
	Variable string
	Value    string
}

// MarshalJSON for custom marshalling
func (cmp *TimestampLessThanPath) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Variable              string
		TimestampLessThanPath string
	}{
		Variable:              cmp.Variable,
		TimestampLessThanPath: cmp.Value,
	})
}
