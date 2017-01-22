package cloudformation

import "encoding/json"
import "reflect"

// If returns a new instance of IfFunc for the provided string expressions.
//
// See also: IfList
func If(condition string, valueIfTrue, valueIfFalse Stringable) IfFunc {
	return IfFunc{
		list:         false,
		Condition:    condition,
		ValueIfTrue:  *valueIfTrue.String(),
		ValueIfFalse: *valueIfFalse.String(),
	}
}

// IfList returns a new instance of IfFunc for the provided string list expressions.
//
// See also: If
func IfList(condition string, valueIfTrue, valueIfFalse StringListable) IfFunc {
	return IfFunc{
		list:         true,
		Condition:    condition,
		ValueIfTrue:  *valueIfTrue.StringList(),
		ValueIfFalse: *valueIfFalse.StringList(),
	}
}

// IfFunc represents an invocation of the Fn::If intrinsic.
//
// See http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/intrinsic-function-reference-conditions.html
type IfFunc struct {
	list         bool
	Condition    string
	ValueIfTrue  interface{} // a StringExpr if list==false, otherwise a StringListExpr
	ValueIfFalse interface{} // a StringExpr if list==false, otherwise a StringListExpr
}

// MarshalJSON returns a JSON representation of the object
func (f IfFunc) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		FnIf []interface{} `json:"Fn::If"`
	}{FnIf: []interface{}{f.Condition, f.ValueIfTrue, f.ValueIfFalse}})
}

// UnmarshalJSON sets the object from the provided JSON representation
func (f *IfFunc) UnmarshalJSON(buf []byte) error {
	v := struct {
		FnIf [3]json.RawMessage `json:"Fn::If"`
	}{}
	if err := json.Unmarshal(buf, &v); err != nil {
		return err
	}
	if err := json.Unmarshal(v.FnIf[0], &f.Condition); err != nil {
		return err
	}

	var probeValue interface{}
	if err := json.Unmarshal(v.FnIf[1], &probeValue); err != nil {
		return err
	}

	switch reflect.ValueOf(probeValue).Kind() {
	case reflect.Array:
		f.list = true
	case reflect.String:
		f.list = false
	case reflect.Map:
		expr, err := unmarshalFunc(v.FnIf[1])
		if err == nil {
			if _, ok := expr.(StringListFunc); ok {
				f.list = true
			}
		}
	}

	if f.list {
		f.ValueIfTrue = StringListExpr{}
		f.ValueIfFalse = StringListExpr{}
	} else {
		f.ValueIfTrue = StringExpr{}
		f.ValueIfFalse = StringExpr{}
	}

	if err := json.Unmarshal(v.FnIf[1], &f.ValueIfTrue); err != nil {
		return err
	}

	if err := json.Unmarshal(v.FnIf[2], &f.ValueIfFalse); err != nil {
		return err
	}
	return nil
}

func (f IfFunc) String() *StringExpr {
	if f.list {
		panic("IfFunc is a list, but being treated as a scalar")
	}
	return &StringExpr{Func: f}
}

// StringList returns a new StringListExpr representing the literal value v.
func (f IfFunc) StringList() *StringListExpr {
	if !f.list {
		panic("IfFunc is a scalar, but being treated as a list of strings.")
	}
	return &StringListExpr{Func: f}
}

var _ StringFunc = IfFunc{}     // IfFunc must implement StringFunc
var _ StringListFunc = IfFunc{} // IfFunc must implement StringListFunc
