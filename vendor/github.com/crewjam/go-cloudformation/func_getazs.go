package cloudformation

// GetAZs returns a new instance of GetAZsFunc.
func GetAZs(region Stringable) *StringListExpr {
	return GetAZsFunc{Region: *region.String()}.StringList()
}

// GetAZsFunc represents an invocation of the Fn::GetAZs intrinsic.
//
// The intrinsic function Fn::GetAZs returns an array that lists Availability
// Zones for a specified region. Because customers have access to different
// Availability Zones, the intrinsic function Fn::GetAZs enables template
// authors to write templates that adapt to the calling user's access. That
// way you don't have to hard-code a full list of Availability Zones for a
// specified region.
//
// See http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/intrinsic-function-reference-getavailabilityzones.html
type GetAZsFunc struct {
	Region StringExpr `json:"Fn::GetAZs"`
}

// StringList returns a new StringListExpr representing the literal value v.
func (f GetAZsFunc) StringList() *StringListExpr {
	return &StringListExpr{Func: f}
}

// Note: Fn::GetAZs does *not* implement StringFunc.
var _ StringListFunc = GetAZsFunc{} // GetAZsFunc must implement StringListFunc
