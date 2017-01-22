package cloudformation

// Ref returns a new instance of RefFunc that refers to name.
func Ref(name string) RefFunc {
	return RefFunc{Name: name}
}

// RefFunc represents an invocation of the Ref intrinsic.
//
// The intrinsic function Ref returns the value of the specified parameter or resource.
//
// - When you specify a parameter's logical name, it returns the value of the
//   parameter.
//
// - When you specify a resource's logical name, it returns a value that you
//   can typically use to refer to that resource.
//
// When you are declaring a resource in a template and you need to specify
// another template resource by name, you can use the Ref to refer to that
// other resource. In general, Ref returns the name of the resource. For
// example, a reference to an AWS::AutoScaling::AutoScalingGroup returns the
// name of that Auto Scaling group resource.
//
// For some resources, an identifier is returned that has another significant
// meaning in the context of the resource. An AWS::EC2::EIP resource, for
// instance, returns the IP address, and an AWS::EC2::Instance returns the
// instance ID.
//
// See http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/intrinsic-function-reference-ref.html
type RefFunc struct {
	Name string `json:"Ref"`
}

// Bool returns this reference as a BoolExpr
func (r RefFunc) Bool() *BoolExpr {
	return &BoolExpr{Func: r}
}

// Integer returns this reference as a IntegerExpr
func (r RefFunc) Integer() *IntegerExpr {
	return &IntegerExpr{Func: r}
}

// String returns this reference as a StringExpr
func (r RefFunc) String() *StringExpr {
	return &StringExpr{Func: r}
}

// StringList returns this reference as a StringListExpr
func (r RefFunc) StringList() *StringListExpr {
	return &StringListExpr{Func: r}
}

var _ Func = RefFunc{}
var _ BoolFunc = RefFunc{}
var _ IntegerFunc = RefFunc{}
var _ StringFunc = RefFunc{}
var _ StringListFunc = RefFunc{}
