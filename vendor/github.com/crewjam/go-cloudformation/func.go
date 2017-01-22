package cloudformation

import (
	"encoding/json"
	"fmt"
)

// Func is an interface provided by objects that represent Cloudformation
// function calls.
type Func interface {
}

// BoolFunc is an interface provided by objects that represent Cloudformation
// function that can return a boolean value.
type BoolFunc interface {
	Func
	Bool() *BoolExpr
}

// IntegerFunc is an interface provided by objects that represent Cloudformation
// function that can return an integer value.
type IntegerFunc interface {
	Func
	Integer() *IntegerExpr
}

// StringFunc is an interface provided by objects that represent Cloudformation
// function that can return a string value.
type StringFunc interface {
	Func
	String() *StringExpr
}

// StringListFunc is an interface provided by objects that represent Cloudformation
// function that can return a list of strings.
type StringListFunc interface {
	Func
	StringList() *StringListExpr
}

// UnknownFunctionError is returned by various UnmarshalJSON
// functions when they encounter a function that is not
// implemented.
type UnknownFunctionError struct {
	Name string
}

func (ufe UnknownFunctionError) Error() string {
	return fmt.Sprintf("unknown function %s", ufe.Name)
}

// unmarshalFunc unmarshals data into a Func, or returns an error
// if the function call is invalid.
func unmarshalFunc(data []byte) (Func, error) {
	rawDecode := map[string]json.RawMessage{}
	err := json.Unmarshal(data, &rawDecode)
	if err != nil {
		return nil, err
	}
	for key := range rawDecode {
		switch key {
		case "Ref":
			f := RefFunc{}
			if err := json.Unmarshal(data, &f); err == nil {
				return f, nil
			}
		case "Fn::Join":
			f := JoinFunc{}
			if err := json.Unmarshal(data, &f); err == nil {
				return f, nil
			}
		case "Fn::Select":
			f := SelectFunc{}
			if err := json.Unmarshal(data, &f); err == nil {
				return f, nil
			}
		case "Fn::GetAtt":
			f := GetAttFunc{}
			if err := json.Unmarshal(data, &f); err == nil {
				return f, nil
			}
		case "Fn::FindInMap":
			f := FindInMapFunc{}
			if err := json.Unmarshal(data, &f); err == nil {
				return f, nil
			}
		case "Fn::Base64":
			f := Base64Func{}
			if err := json.Unmarshal(data, &f); err == nil {
				return f, nil
			}
		case "Fn::GetAZs":
			f := GetAZsFunc{}
			if err := json.Unmarshal(data, &f); err == nil {
				return f, nil
			}
		case "Fn::If":
			f := IfFunc{}
			if err := json.Unmarshal(data, &f); err == nil {
				return f, nil
			}
		case "Fn::ImportValue":
			f := ImportValueFunc{}
			if err := json.Unmarshal(data, &f); err == nil {
				return f, nil
			}
		default:
			return nil, UnknownFunctionError{Name: key}
		}
	}
	return nil, fmt.Errorf("cannot decode function")
}
