package cloudformation

import (
	"encoding/json"
	"strconv"
)

// BoolExpr represents a boolean expression. If the value is computed then
// Func will be non-nil. If it is a literal `true` or `false` then
// the Literal gives the value. Typically instances of this function
// are created by Bool() or one of the function constructors. Ex:
//
//   type LocalBalancer struct {
//     CrossZone *BoolExpr
//   }
//
//   lb := LocalBalancer{CrossZone: Bool(true)}
//   lb2 := LocalBalancer{CrossZone: Ref("LoadBalancerCrossZone").Bool()}
//
type BoolExpr struct {
	Func    BoolFunc
	Literal bool
}

// MarshalJSON returns a JSON representation of the object
func (x BoolExpr) MarshalJSON() ([]byte, error) {
	if x.Func != nil {
		return json.Marshal(x.Func)
	}
	return json.Marshal(x.Literal)
}

// UnmarshalJSON sets the object from the provided JSON representation
func (x *BoolExpr) UnmarshalJSON(data []byte) error {
	var v bool
	err := json.Unmarshal(data, &v)
	if err == nil {
		x.Func = nil
		x.Literal = v
		return nil
	}

	// Cloudformation allows bool values to be represented as the
	// strings "true" and "false"
	var strValue string
	if err := json.Unmarshal(data, &strValue); err == nil {
		if v, err := strconv.ParseBool(strValue); err == nil {
			x.Func = nil
			x.Literal = v
			return nil
		}
	}

	// Perhaps we have a serialized function call (like `{"Ref": "Foo"}`)
	// so we'll try to unmarshal it with UnmarshalFunc. Not all Funcs also
	// implement BoolFunc, so we have to make sure that the referenced
	// function actually works in the boolean context
	funcCall, err2 := unmarshalFunc(data)
	if err2 == nil {
		boolFunc, ok := funcCall.(BoolFunc)
		if ok {
			x.Func = boolFunc
			return nil
		}
	} else if unknownFunctionErr, ok := err2.(UnknownFunctionError); ok {
		return unknownFunctionErr
	}

	// Return the original error trying to unmarshal the literal expression,
	// which will be the most expressive.
	return err
}

// Bool returns a new BoolExpr representing the literal value v.
func Bool(v bool) *BoolExpr {
	return &BoolExpr{Literal: v}
}
