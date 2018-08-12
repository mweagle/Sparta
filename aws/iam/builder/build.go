package iambuilder

import (
	sparta "github.com/mweagle/Sparta"
	spartaIAM "github.com/mweagle/Sparta/aws/iam"
	gocf "github.com/mweagle/go-cloudformation"
)

// IAMPrivilegeBuilder encapsulates the IAM builder
type IAMPrivilegeBuilder struct {
	resource      *IAMResourceBuilder
	resourceParts []gocf.Stringable
}

// Ref inserts a go-cloudformation Ref entry
func (iamRes *IAMPrivilegeBuilder) Ref(resName string, delimiter ...string) *IAMPrivilegeBuilder {
	iamRes.resourceParts = append(iamRes.resourceParts,
		gocf.Ref(resName))
	for _, eachDelimiter := range delimiter {
		iamRes.resourceParts = append(iamRes.resourceParts,
			gocf.String(eachDelimiter))
	}
	return iamRes
}

// Attr inserts a go-cloudformation GetAtt entry
func (iamRes *IAMPrivilegeBuilder) Attr(resName string, propName string, delimiter ...string) *IAMPrivilegeBuilder {
	iamRes.resourceParts = append(iamRes.resourceParts,
		gocf.GetAtt(resName, propName))
	for _, eachDelimiter := range delimiter {
		iamRes.resourceParts = append(iamRes.resourceParts,
			gocf.String(eachDelimiter))
	}
	return iamRes
}

// Region inserts the AWS::Region pseudo param into the privilege
func (iamRes *IAMPrivilegeBuilder) Region(delimiter ...string) *IAMPrivilegeBuilder {
	iamRes.resourceParts = append(iamRes.resourceParts,
		gocf.Ref("AWS::Region"))
	for _, eachDelimiter := range delimiter {
		iamRes.resourceParts = append(iamRes.resourceParts,
			gocf.String(eachDelimiter))
	}
	return iamRes
}

// AccountID inserts the AWS::AccountId pseudo param into the privilege
func (iamRes *IAMPrivilegeBuilder) AccountID(delimiter ...string) *IAMPrivilegeBuilder {
	iamRes.resourceParts = append(iamRes.resourceParts,
		gocf.Ref("AWS::AccountId"))
	for _, eachDelimiter := range delimiter {
		iamRes.resourceParts = append(iamRes.resourceParts,
			gocf.String(eachDelimiter))
	}
	return iamRes
}

// NotificationARNS inserts the AWS::NotificationARNs pseudo param into the privilege
func (iamRes *IAMPrivilegeBuilder) NotificationARNS(delimiter ...string) *IAMPrivilegeBuilder {
	iamRes.resourceParts = append(iamRes.resourceParts,
		gocf.Ref("AWS::NotificationARNs"))
	for _, eachDelimiter := range delimiter {
		iamRes.resourceParts = append(iamRes.resourceParts,
			gocf.String(eachDelimiter))
	}
	return iamRes
}

// Partition inserts the AWS::Partition pseudo param into the privilege
func (iamRes *IAMPrivilegeBuilder) Partition(delimiter ...string) *IAMPrivilegeBuilder {
	iamRes.resourceParts = append(iamRes.resourceParts,
		gocf.Ref("AWS::Partition"))
	for _, eachDelimiter := range delimiter {
		iamRes.resourceParts = append(iamRes.resourceParts,
			gocf.String(eachDelimiter))
	}
	return iamRes
}

// StackID inserts the AWS::StackID pseudo param into the privilege
func (iamRes *IAMPrivilegeBuilder) StackID(delimiter ...string) *IAMPrivilegeBuilder {
	iamRes.resourceParts = append(iamRes.resourceParts,
		gocf.Ref("AWS::StackId"))
	for _, eachDelimiter := range delimiter {
		iamRes.resourceParts = append(iamRes.resourceParts,
			gocf.String(eachDelimiter))
	}
	return iamRes
}

// StackName inserts the AWS::StackName pseudo param into the privilege
func (iamRes *IAMPrivilegeBuilder) StackName(delimiter ...string) *IAMPrivilegeBuilder {
	iamRes.resourceParts = append(iamRes.resourceParts,
		gocf.Ref("AWS::StackName"))
	for _, eachDelimiter := range delimiter {
		iamRes.resourceParts = append(iamRes.resourceParts,
			gocf.String(eachDelimiter))
	}
	return iamRes
}

// URLSuffix inserts the AWS::URLSuffix pseudo param into the privilege
func (iamRes *IAMPrivilegeBuilder) URLSuffix(delimiter ...string) *IAMPrivilegeBuilder {
	iamRes.resourceParts = append(iamRes.resourceParts,
		gocf.Ref("AWS::URLSuffix"))
	for _, eachDelimiter := range delimiter {
		iamRes.resourceParts = append(iamRes.resourceParts,
			gocf.String(eachDelimiter))
	}
	return iamRes
}

// Literal inserts a string literal into the ARN being constructed
func (iamRes *IAMPrivilegeBuilder) Literal(arnPart string) *IAMPrivilegeBuilder {
	iamRes.resourceParts = append(iamRes.resourceParts,
		gocf.String(arnPart))
	return iamRes
}

// ToPolicyStatement finalizes the builder and returns a spartaIAM.PolicyStatements
func (iamRes *IAMPrivilegeBuilder) ToPolicyStatement() spartaIAM.PolicyStatement {
	return spartaIAM.PolicyStatement{
		Action:   iamRes.resource.apiCalls,
		Effect:   "Allow",
		Resource: gocf.Join("", iamRes.resourceParts...),
	}
}

// ToPrivilege returns a legacy sparta.IAMRolePrivilege type for this
// entry
func (iamRes *IAMPrivilegeBuilder) ToPrivilege() sparta.IAMRolePrivilege {
	return sparta.IAMRolePrivilege{
		Actions:  iamRes.resource.apiCalls,
		Resource: gocf.Join("", iamRes.resourceParts...),
	}
}

// IAMResourceBuilder is the intermediate type that
// creates the Resource to which the privilege applies
type IAMResourceBuilder struct {
	apiCalls []string
}

// ForResource returns the IAMPrivilegeBuilder instance
// which can be finalized into an IAMRolePrivilege
func (iamRes *IAMResourceBuilder) ForResource() *IAMPrivilegeBuilder {
	return &IAMPrivilegeBuilder{
		resource:      iamRes,
		resourceParts: make([]gocf.Stringable, 0),
	}
}

// Allow creates a IAMPrivilegeBuilder instance for the supplied API calls
func Allow(apiCalls ...string) *IAMResourceBuilder {
	resource := IAMResourceBuilder{
		apiCalls: apiCalls,
	}
	return &resource
}
