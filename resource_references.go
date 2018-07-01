// +build !lambdabinary

package sparta

import (
	"encoding/json"
	"reflect"
	"strings"

	gocf "github.com/mweagle/go-cloudformation"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type resourceRefType int

const (
	resourceLiteral resourceRefType = iota
	resourceRefFunc
	resourceGetAttrFunc
	resourceStringFunc
)

type resourceRef struct {
	RefType      resourceRefType
	ResourceName string
}

// resolvedResourceVisitor represents the signature of a function that
// visits
type resolvedResourceVisitor func(lambdaAWSInfo *LambdaAWSInfo,
	eventSourceMapping *EventSourceMapping,
	mappingIndex int,
	resource *resourceRef) error

// resolveResourceRef takes an interface representing a dynamic ARN
// and tries to determine the CloudFormation resource name it resolves to
func resolveResourceRef(expr interface{}) (*resourceRef, error) {

	// Is there any chance it's just a string?
	typedString, typedStringOk := expr.(string)
	if typedStringOk {
		return &resourceRef{
			RefType:      resourceLiteral,
			ResourceName: typedString,
		}, nil
	}
	// Some type of intrinsic function?
	marshalled, marshalledErr := json.Marshal(expr)
	if marshalledErr != nil {
		return nil, errors.Errorf("Failed to unmarshal dynamic resource ref %v", expr)
	}
	var refFunc gocf.RefFunc
	if json.Unmarshal(marshalled, &refFunc) == nil &&
		len(refFunc.Name) != 0 {
		return &resourceRef{
			RefType:      resourceRefFunc,
			ResourceName: refFunc.Name,
		}, nil
	}

	var getAttFunc gocf.GetAttFunc
	if json.Unmarshal(marshalled, &getAttFunc) == nil && len(getAttFunc.Resource) != 0 {
		return &resourceRef{
			RefType:      resourceGetAttrFunc,
			ResourceName: getAttFunc.Resource,
		}, nil
	}
	// Any chance it's a string?
	var stringExprFunc gocf.StringExpr
	if json.Unmarshal(marshalled, &stringExprFunc) == nil && len(stringExprFunc.Literal) != 0 {
		return &resourceRef{
			RefType:      resourceStringFunc,
			ResourceName: stringExprFunc.Literal,
		}, nil
	}

	// Nope
	return nil, nil
}

// isResolvedResourceType is a utility function to determine if a resolved
// reference is a given type. If it is a literal, the literalTokenIndicator
// substring match is used for the predicate. If it is a resource provisioned
// by this template, the &gocf.RESOURCE_TYPE{} will be used via reflection
// Example:
// isResolvedResourceType(resourceRef, template, ":dynamodb:", &gocf.DynamoDBTable{}) {
//
func isResolvedResourceType(resource *resourceRef,
	template *gocf.Template,
	literalTokenIndicator string,
	templateType gocf.ResourceProperties) bool {
	if resource.RefType == resourceLiteral ||
		resource.RefType == resourceStringFunc {
		return strings.Contains(resource.ResourceName, literalTokenIndicator)
	}

	// Dynamically provisioned resource included in the template definition?
	existingResource, existingResourceExists := template.Resources[resource.ResourceName]
	if existingResourceExists {
		if reflect.TypeOf(existingResource.Properties) == reflect.TypeOf(templateType) {
			return true
		}
	}
	return false
}

// visitResolvedEventSourceMapping is a utility function that visits all the EventSourceMapping
// entries for the given lambdaAWSInfo struct
func visitResolvedEventSourceMapping(visitor resolvedResourceVisitor,
	lambdaAWSInfos []*LambdaAWSInfo,
	template *gocf.Template,
	logger *logrus.Logger) error {

	//
	// BEGIN
	// Inline closure to wrap the visitor function so that we can provide
	// specific error messages
	visitEventSourceMappingRef := func(lambdaAWSInfo *LambdaAWSInfo,
		eventSourceMapping *EventSourceMapping,
		mappingIndex int,
		resource *resourceRef) error {

		annotateStatementsErr := visitor(lambdaAWSInfo,
			eventSourceMapping,
			mappingIndex,
			resource)

		// Early exit?
		if annotateStatementsErr != nil {
			return errors.Wrapf(annotateStatementsErr,
				"Visiting event source mapping: %s",
				eventSourceMapping)
		}
		return nil
	}
	//
	// END

	// Iterate through every lambda function. If there is an EventSourceMapping
	// that points to a piece of infastructure provisioned by this stack,
	// find the referred resource and supply it to the visitor
	for _, eachLambda := range lambdaAWSInfos {
		for eachIndex, eachEventSource := range eachLambda.EventSourceMappings {
			resourceRef, resourceRefErr := resolveResourceRef(eachEventSource.EventSourceArn)
			if resourceRefErr != nil {
				return errors.Wrapf(resourceRefErr,
					"Failed to resolve EventSourceArn: %#v", eachEventSource)
			}

			// At this point everything is a string, so we need to unmarshall
			// and see if the Arn is supplied by either a Ref or a GetAttr
			// function. In those cases, we need to look around in the template
			// to go from: EventMapping -> Type -> Lambda -> LambdaIAMRole
			// so that we can add the permissions
			if resourceRef != nil {
				annotationErr := visitEventSourceMappingRef(eachLambda,
					eachEventSource,
					eachIndex,
					resourceRef)
				// Anything go wrong?
				if annotationErr != nil {
					return errors.Wrapf(annotationErr,
						"Failed to annotate template for EventSourceMapping: %#v",
						eachEventSource)
				}
			}
		}
	}
	return nil
}
