package cloudformation

import (
	"encoding/json"
	"fmt"
)

// StringListable is an interface that describes structures that are convertable
// to a *StringListExpr.
type StringListable interface {
	StringList() *StringListExpr
}

// StringListExpr is a string expression. If the value is computed then
// Func will be non-nil. If it is a literal string then Literal gives
// the value. Typically instances of this function are created by
// StringList() or one of the function constructors. Ex:
//
//   type LocalBalancer struct {
//     Name *StringListExpr
//   }
//
//   lb := LocalBalancer{Name: StringList("hello")}
//   lb2 := LocalBalancer{Name: Ref("LoadBalancerNane").StringList()}
//
type StringListExpr struct {
	Func    StringListFunc
	Literal []*StringExpr
}

// StringList implements StringListable
func (x StringListExpr) StringList() *StringListExpr {
	return &x
}

// MarshalJSON returns a JSON representation of the object
func (x StringListExpr) MarshalJSON() ([]byte, error) {
	if x.Func != nil {
		return json.Marshal(x.Func)
	}
	return json.Marshal(x.Literal)
}

// UnmarshalJSON sets the object from the provided JSON representation
func (x *StringListExpr) UnmarshalJSON(data []byte) error {
	var v []*StringExpr
	err := json.Unmarshal(data, &v)
	if err == nil {
		x.Func = nil
		x.Literal = v
		return nil
	}

	// Perhaps we have a serialized function call (like `{"Ref": "Foo"}`)
	// so we'll try to unmarshal it with UnmarshalFunc. Not all Funcs also
	// implement StringListFunc, so we have to make sure that the referenced
	// function actually works in the boolean context
	funcCall, err2 := unmarshalFunc(data)
	if err2 == nil {
		stringFunc, ok := funcCall.(StringListFunc)
		if ok {
			x.Func = stringFunc
			return nil
		}
		return fmt.Errorf("%#v is not a StringListFunc", funcCall)
	}

	// Perhaps we have a single item, like "foo" which
	// occurs occasionally.
	var v2 StringExpr
	err3 := json.Unmarshal(data, &v2)
	if err3 == nil {
		x.Func = nil
		x.Literal = []*StringExpr{&v2}
		return nil
	}

	// Return the original error trying to unmarshal the literal expression,
	// which will be the most expressive.
	return err
}

// StringList returns a new StringListExpr representing the literal value v.
func StringList(v ...Stringable) *StringListExpr {
	rv := &StringListExpr{Literal: []*StringExpr{}}
	for _, item := range v {
		rv.Literal = append(rv.Literal, item.String())
	}
	return rv
}
