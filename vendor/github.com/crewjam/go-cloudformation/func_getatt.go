package cloudformation

import "encoding/json"

// GetAtt returns a new instance of GetAttFunc.
func GetAtt(resource, name string) *StringExpr {
	return GetAttFunc{Resource: resource, Name: name}.String()
}

// GetAttFunc represents an invocation of the Fn::GetAtt intrinsic.
//
// The intrinsic function Fn::GetAtt returns the value of an attribute from a
// resource in the template.
//
// See http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/intrinsic-function-reference-getatt.html
type GetAttFunc struct {
	Resource string
	Name     string
}

// MarshalJSON returns a JSON representation of the object
func (f GetAttFunc) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		FnGetAtt []string `json:"Fn::GetAtt"`
	}{FnGetAtt: []string{f.Resource, f.Name}})
}

// UnmarshalJSON sets the object from the provided JSON representation
func (f *GetAttFunc) UnmarshalJSON(data []byte) error {
	v := struct {
		FnGetAtt []string `json:"Fn::GetAtt"`
	}{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	if len(v.FnGetAtt) != 2 {
		return &json.UnsupportedValueError{Str: string(data)}
	}
	f.Resource = v.FnGetAtt[0]
	f.Name = v.FnGetAtt[1]
	return nil
}

func (f GetAttFunc) String() *StringExpr {
	return &StringExpr{Func: f}
}

var _ StringFunc = GetAttFunc{} // GetAttFunc must implement StringFunc
