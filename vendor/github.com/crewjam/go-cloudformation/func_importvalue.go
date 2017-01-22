package cloudformation

// ImportValue returns a new instance of ImportValue that imports valueToImport.
func ImportValue(valueToImport Stringable) ImportValueFunc {
	return ImportValueFunc{ValueToImport: *valueToImport.String()}
}

// ImportValueFunc represents an invocation of the Fn::ImportValue intrinsic.
// The intrinsic function Fn::ImportValue returns the value of an output exported by
// another stack. You typically use this function to create cross-stack references.
// In the following example template snippets, Stack A exports VPC security group
// values and Stack B imports them.
//
// Note
// The following restrictions apply to cross-stack references:
//    For each AWS account, Export names must be unique within a region.
//    You can't create cross-stack references across different regions. You can
//      use the intrinsic function Fn::ImportValue only to import values that
//      have been exported within the same region.
//    For outputs, the value of the Name property of an Export can't use
//      functions (Ref or GetAtt) that depend on a resource.
//    Similarly, the ImportValue function can't include functions (Ref or GetAtt)
//      that depend on a resource.
//    You can't delete a stack if another stack references one of its outputs.
//    You can't modify or remove the output value as long as it's referenced by another stack.
//
// See http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/intrinsic-function-reference-importvalue.html
type ImportValueFunc struct {
	ValueToImport StringExpr `json:"Fn::ImportValue"`
}

// String returns this reference as a StringExpr
func (r ImportValueFunc) String() *StringExpr {
	return &StringExpr{Func: r}
}

// StringList returns this reference as a StringListExpr
func (r ImportValueFunc) StringList() *StringListExpr {
	return &StringListExpr{Func: r}
}

var _ StringFunc = ImportValueFunc{}
var _ StringListFunc = ImportValueFunc{}
