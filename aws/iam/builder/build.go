package iambuilder

import (
	gof "github.com/awslabs/goformation/v5/cloudformation"
	sparta "github.com/mweagle/Sparta"
	spartaIAM "github.com/mweagle/Sparta/aws/iam"
	iamtypes "github.com/mweagle/Sparta/aws/iam/builder/types"
)

////////////////////////////////////////////////////////////////////////////////
/*
  ___ ___ ___  ___  _   _ ___  ___ ___
 | _ \ __/ __|/ _ \| | | | _ \/ __| __|
 |   / _|\__ \ (_) | |_| |   / (__| _|
 |_|_\___|___/\___/ \___/|_|_\\___|___|
*/
////////////////////////////////////////////////////////////////////////////////

// IAMResourceBuilder encapsulates the IAM builder for a resource
type IAMResourceBuilder struct {
	builder       *IAMBuilder
	resourceParts []string
}

// Ref inserts a go-cloudformation Ref entry
func (iamRes *IAMResourceBuilder) Ref(resName string, delimiter ...string) *IAMResourceBuilder {
	iamRes.resourceParts = append(iamRes.resourceParts,
		gof.Ref(resName))
	for _, eachDelimiter := range delimiter {
		iamRes.resourceParts = append(iamRes.resourceParts,
			eachDelimiter)
	}
	return iamRes
}

// Attr inserts a go-cloudformation GetAtt entry
func (iamRes *IAMResourceBuilder) Attr(resName string, propName string, delimiter ...string) *IAMResourceBuilder {
	iamRes.resourceParts = append(iamRes.resourceParts,
		gof.GetAtt(resName, propName))
	for _, eachDelimiter := range delimiter {
		iamRes.resourceParts = append(iamRes.resourceParts,
			eachDelimiter)
	}
	return iamRes
}

// Region inserts the AWS::Region pseudo param into the privilege
func (iamRes *IAMResourceBuilder) Region(delimiter ...string) *IAMResourceBuilder {
	iamRes.resourceParts = append(iamRes.resourceParts,
		gof.Ref("AWS::Region"))
	for _, eachDelimiter := range delimiter {
		iamRes.resourceParts = append(iamRes.resourceParts,
			eachDelimiter)
	}
	return iamRes
}

// AccountID inserts the AWS::AccountId pseudo param into the privilege
func (iamRes *IAMResourceBuilder) AccountID(delimiter ...string) *IAMResourceBuilder {
	iamRes.resourceParts = append(iamRes.resourceParts,
		gof.Ref("AWS::AccountId"))
	for _, eachDelimiter := range delimiter {
		iamRes.resourceParts = append(iamRes.resourceParts,
			eachDelimiter)
	}
	return iamRes
}

// NotificationARNS inserts the AWS::NotificationARNs pseudo param into the privilege
func (iamRes *IAMResourceBuilder) NotificationARNS(delimiter ...string) *IAMResourceBuilder {
	iamRes.resourceParts = append(iamRes.resourceParts,
		gof.Ref("AWS::NotificationARNs"))
	for _, eachDelimiter := range delimiter {
		iamRes.resourceParts = append(iamRes.resourceParts,
			eachDelimiter)
	}
	return iamRes
}

// Partition inserts the AWS::Partition pseudo param into the privilege
func (iamRes *IAMResourceBuilder) Partition(delimiter ...string) *IAMResourceBuilder {
	iamRes.resourceParts = append(iamRes.resourceParts,
		gof.Ref("AWS::Partition"))
	for _, eachDelimiter := range delimiter {
		iamRes.resourceParts = append(iamRes.resourceParts,
			eachDelimiter)
	}
	return iamRes
}

// StackID inserts the AWS::StackID pseudo param into the privilege
func (iamRes *IAMResourceBuilder) StackID(delimiter ...string) *IAMResourceBuilder {
	iamRes.resourceParts = append(iamRes.resourceParts,
		gof.Ref("AWS::StackId"))
	for _, eachDelimiter := range delimiter {
		iamRes.resourceParts = append(iamRes.resourceParts,
			eachDelimiter)
	}
	return iamRes
}

// StackName inserts the AWS::StackName pseudo param into the privilege
func (iamRes *IAMResourceBuilder) StackName(delimiter ...string) *IAMResourceBuilder {
	iamRes.resourceParts = append(iamRes.resourceParts,
		gof.Ref("AWS::StackName"))
	for _, eachDelimiter := range delimiter {
		iamRes.resourceParts = append(iamRes.resourceParts,
			eachDelimiter)
	}
	return iamRes
}

// URLSuffix inserts the AWS::URLSuffix pseudo param into the privilege
func (iamRes *IAMResourceBuilder) URLSuffix(delimiter ...string) *IAMResourceBuilder {
	iamRes.resourceParts = append(iamRes.resourceParts,
		gof.Ref("AWS::URLSuffix"))
	for _, eachDelimiter := range delimiter {
		iamRes.resourceParts = append(iamRes.resourceParts,
			eachDelimiter)
	}
	return iamRes
}

// Literal inserts a string literal into the ARN being constructed
func (iamRes *IAMResourceBuilder) Literal(arnPart string) *IAMResourceBuilder {
	iamRes.resourceParts = append(iamRes.resourceParts,
		arnPart)
	return iamRes
}

// ToPolicyStatement finalizes the builder and returns a spartaIAM.PolicyStatements
func (iamRes *IAMResourceBuilder) ToPolicyStatement() spartaIAM.PolicyStatement {
	return spartaIAM.PolicyStatement{
		Action:   iamRes.builder.apiCalls,
		Effect:   iamRes.builder.effect,
		Resource: gof.Join("", iamRes.resourceParts),
	}
}

// ToPrivilege returns a legacy sparta.IAMRolePrivilege type for this
// entry
func (iamRes *IAMResourceBuilder) ToPrivilege() sparta.IAMRolePrivilege {
	return sparta.IAMRolePrivilege{
		Actions:  iamRes.builder.apiCalls,
		Resource: gof.Join("", iamRes.resourceParts),
	}
}

// IAMBuilder is the intermediate type that
// creates the Resource to which the privilege applies
type IAMBuilder struct {
	apiCalls  []string
	effect    string
	condition interface{}
}

// ForResource returns the IAMPrivilegeBuilder instance
// which can be finalized into an IAMRolePrivilege
func (iamRes *IAMBuilder) ForResource() *IAMResourceBuilder {
	return &IAMResourceBuilder{
		builder:       iamRes,
		resourceParts: []string{},
	}
}

// WithCondition applies the given condition to the policy
func (iamRes *IAMBuilder) WithCondition(conditionExpression interface{}) *IAMBuilder {
	iamRes.condition = conditionExpression
	return iamRes
}

////////////////////////////////////////////////////////////////////////////////
/*
  ___ ___ ___ _  _  ___ ___ ___  _   _
 | _ \ _ \_ _| \| |/ __|_ _| _ \/_\ | |
 |  _/   /| || .` | (__ | ||  _/ _ \| |__
 |_| |_|_\___|_|\_|\___|___|_|/_/ \_\____|

*/
////////////////////////////////////////////////////////////////////////////////

// IAMPrincipalBuilder is the builder for a Principal allowance
type IAMPrincipalBuilder struct {
	builder   *IAMBuilder
	principal *iamtypes.IAMPrincipal
}

// ForPrincipals returns the IAMPrincipalBuilder instance
// which can be finalized into an IAMRolePrivilege
func (iamRes *IAMBuilder) ForPrincipals(principals ...string) *IAMPrincipalBuilder {
	stringablePrincipals := make([]string, len(principals))
	for index, eachPrincipal := range principals {
		stringablePrincipals[index] = eachPrincipal
	}

	return &IAMPrincipalBuilder{
		builder: iamRes,
		principal: &iamtypes.IAMPrincipal{
			Service: stringablePrincipals,
		},
	}
}

// ForFederatedPrincipals returns the IAMPrincipalBuilder instance
// which can be finalized into an IAMRolePrivilege
func (iamRes *IAMBuilder) ForFederatedPrincipals(principals ...string) *IAMPrincipalBuilder {
	stringablePrincipals := make([]string, len(principals))
	for index, eachPrincipal := range principals {
		stringablePrincipals[index] = eachPrincipal
	}

	return &IAMPrincipalBuilder{
		builder: iamRes,
		principal: &iamtypes.IAMPrincipal{
			Federated: stringablePrincipals,
		},
	}
}

// ToPolicyStatement finalizes the builder and returns a spartaIAM.PolicyStatements
func (iampb *IAMPrincipalBuilder) ToPolicyStatement() spartaIAM.PolicyStatement {
	statement := spartaIAM.PolicyStatement{
		Action:    iampb.builder.apiCalls,
		Effect:    iampb.builder.effect,
		Principal: iampb.principal,
	}
	if iampb.builder.condition != nil {
		statement.Condition = iampb.builder.condition
	}
	return statement
}

// ToPrivilege returns a legacy sparta.IAMRolePrivilege type for this
// IAMPrincipalBuilder entry
func (iampb *IAMPrincipalBuilder) ToPrivilege() sparta.IAMRolePrivilege {
	privilege := sparta.IAMRolePrivilege{
		Actions:   iampb.builder.apiCalls,
		Principal: iampb.principal,
	}
	if iampb.builder.condition != nil {
		privilege.Condition = iampb.builder.condition
	}
	return privilege
}

////////////////////////////////////////////////////////////////////////////////
/*
   ___ _____ ___  ___
  / __|_   _/ _ \| _ \
 | (__  | || (_) |   /
  \___| |_| \___/|_|_\
*/
////////////////////////////////////////////////////////////////////////////////

// Allow creates a IAMPrivilegeBuilder instance Allowing the supplied API calls
func Allow(apiCalls ...string) *IAMBuilder {
	builder := IAMBuilder{
		apiCalls: apiCalls,
		effect:   "Allow",
	}
	return &builder
}

// Deny creates a IAMPrivilegeBuilder instance Denying the supplied API calls
func Deny(apiCalls ...string) *IAMBuilder {
	builder := IAMBuilder{
		apiCalls: apiCalls,
		effect:   "Deny",
	}
	return &builder
}
