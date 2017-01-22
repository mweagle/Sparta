package cloudformation

import "encoding/json"

// Stringable is an interface that describes structures that are convertable
// to a *StringExpr.
type Stringable interface {
	String() *StringExpr
}

// StringExpr is a string expression. If the value is computed then
// Func will be non-nil. If it is a literal string then Literal gives
// the value. Typically instances of this function are created by
// String() or one of the function constructors. Ex:
//
//   type LocalBalancer struct {
//     Name *StringExpr
//   }
//
//   lb := LocalBalancer{Name: String("hello")}
//   lb2 := LocalBalancer{Name: Ref("LoadBalancerNane").String()}
//
type StringExpr struct {
	Func    StringFunc
	Literal string
}

// String implements Stringable
func (x StringExpr) String() *StringExpr {
	return &x
}

// MarshalJSON returns a JSON representation of the object
func (x StringExpr) MarshalJSON() ([]byte, error) {
	if x.Func != nil {
		return json.Marshal(x.Func)
	}
	return json.Marshal(x.Literal)
}

// UnmarshalJSON sets the object from the provided JSON representation
func (x *StringExpr) UnmarshalJSON(data []byte) error {
	var v string
	err := json.Unmarshal(data, &v)
	if err == nil {
		x.Func = nil
		x.Literal = v
		return nil
	}

	// Perhaps we have a serialized function call (like `{"Ref": "Foo"}`)
	// so we'll try to unmarshal it with UnmarshalFunc. Not all Funcs also
	// implement StringFunc, so we have to make sure that the referenced
	// function actually works in the boolean context
	funcCall, err2 := unmarshalFunc(data)
	if err2 == nil {
		stringFunc, ok := funcCall.(Stringable)
		if ok {
			x.Func = stringFunc
			return nil
		}
	} else if unknownFunctionErr, ok := err2.(UnknownFunctionError); ok {
		return unknownFunctionErr
	}

	// Return the original error trying to unmarshal the literal expression,
	// which will be the most expressive.
	return err
}

// String returns a new StringExpr representing the literal value v.
func String(v string) *StringExpr {
	return &StringExpr{Literal: v}
}
