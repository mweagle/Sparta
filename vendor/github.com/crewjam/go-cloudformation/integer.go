package cloudformation

import (
	"encoding/json"
	"strconv"
)

// IntegerExpr is a integer expression. If the value is computed then
// Func will be non-nill. If it is a literal constant integer then
// the Literal gives the value. Typically instances of this function
// are created by Integer() Ex:
//
//   type LocalBalancer struct {
//     Timeout *IntegerExpr
//   }
//
//   lb := LocalBalancer{Timeout: Integer(300)}
//
type IntegerExpr struct {
	Func    IntegerFunc
	Literal int64
}

// MarshalJSON returns a JSON representation of the object
func (x IntegerExpr) MarshalJSON() ([]byte, error) {
	if x.Func != nil {
		return json.Marshal(x.Func)
	}
	return json.Marshal(x.Literal)
}

// UnmarshalJSON sets the object from the provided JSON representation
func (x *IntegerExpr) UnmarshalJSON(data []byte) error {
	var v int64
	err := json.Unmarshal(data, &v)
	if err == nil {
		x.Func = nil
		x.Literal = v
		return nil
	}

	// Cloudformation allows int values to be represented as strings
	var strValue string
	if err := json.Unmarshal(data, &strValue); err == nil {
		if v, err := strconv.ParseInt(strValue, 10, 64); err == nil {
			x.Func = nil
			x.Literal = v
			return nil
		}
	}

	// Perhaps we have a serialized function call (like `{"Ref": "Foo"}`)
	// so we'll try to unmarshal it with UnmarshalFunc. Not all Funcs also
	// implement IntegerFunc, so we have to make sure that the referenced
	// function actually works in the intean context
	funcCall, err2 := unmarshalFunc(data)
	if err2 == nil {
		intFunc, ok := funcCall.(IntegerFunc)
		if ok {
			x.Func = intFunc
			return nil
		}
	} else if unknownFunctionErr, ok := err2.(UnknownFunctionError); ok {
		return unknownFunctionErr
	}

	// Return the original error trying to unmarshal the literal expression,
	// which will be the most expressive.
	return err
}

// Integer returns a new IntegerExpr representing the literal value v.
func Integer(v int64) *IntegerExpr {
	return &IntegerExpr{Literal: v}
}
