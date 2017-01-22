package cloudformation

import "encoding/json"

type selectArg interface{}

// Select returns a new instance of SelectFunc chooses among items via selector. If you
func Select(selector string, items ...interface{}) *StringExpr {
	if len(items) == 1 {
		if itemList, ok := items[0].(StringListable); ok {
			return SelectFunc{Selector: selector, Items: *itemList.StringList()}.String()
		}
	}
	stringableItems := make([]Stringable, len(items))
	for i, item := range items {
		stringableItems[i] = item.(Stringable)
	}
	return SelectFunc{Selector: selector, Items: *StringList(stringableItems...)}.String()
}

// SelectFunc represents an invocation of the Fn::Select intrinsic.
//
// The intrinsic function Fn::Select returns a single object from a
// list of objects by index.
//
// See http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/intrinsic-function-reference-select.html
type SelectFunc struct {
	Selector string // XXX int?
	Items    StringListExpr
}

// MarshalJSON returns a JSON representation of the object
func (f SelectFunc) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		FnSelect []interface{} `json:"Fn::Select"`
	}{FnSelect: []interface{}{f.Selector, f.Items}})
}

// UnmarshalJSON sets the object from the provided JSON representation
func (f *SelectFunc) UnmarshalJSON(data []byte) error {
	v := struct {
		FnSelect []json.RawMessage `json:"Fn::Select"`
	}{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	if len(v.FnSelect) != 2 {
		return &json.UnsupportedValueError{Str: string(data)}
	}
	if err := json.Unmarshal(v.FnSelect[0], &f.Selector); err != nil {
		return err
	}
	if err := json.Unmarshal(v.FnSelect[1], &f.Items); err != nil {
		return err
	}

	return nil
}

func (f SelectFunc) String() *StringExpr {
	return &StringExpr{Func: f}
}

var _ StringFunc = SelectFunc{} // SelectFunc must implement StringFunc
