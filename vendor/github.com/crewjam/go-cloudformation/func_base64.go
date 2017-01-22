package cloudformation

// Base64 represents the Fn::Base64 function called over value.
func Base64(value Stringable) *StringExpr {
	return Base64Func{Value: *value.String()}.String()
}

// Base64Func represents an invocation of Fn::Base64.
//
// The intrinsic function Fn::Base64 returns the Base64 representation of the
// input string. This function is typically used to pass encoded data to
// Amazon EC2 instances by way of the UserData property.
//
// See http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/intrinsic-function-reference-base64.html
type Base64Func struct {
	Value StringExpr `json:"Fn::Base64"`
}

func (f Base64Func) String() *StringExpr {
	return &StringExpr{Func: f}
}

var _ Stringable = Base64Func{} // Base64Func must implement Stringable
var _ StringFunc = Base64Func{} // Base64Func must implement StringFunc
