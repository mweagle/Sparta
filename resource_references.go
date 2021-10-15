//go:build !lambdabinary
// +build !lambdabinary

package sparta

import (
	"encoding/base64"
	"reflect"
	"strings"

	gof "github.com/awslabs/goformation/v5/cloudformation"
	gofintrinsics "github.com/awslabs/goformation/v5/intrinsics"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type resourceRefType int

const (
	resourceLiteral resourceRefType = iota
	resourceRefFunc
	resourceGetAttrFunc
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

// resolveResourceRef takes an interface representing an ARN (either static or explicit)
// and tries to determine the CloudFormation resource name it resolves to
func resolveResourceRef(expr string) (*resourceRef, error) {

	// Ref: https://github.com/awslabs/goformation/blob/346053f16b2c9aba3a050c6a2956e18fe3a3f56f/intrinsics/intrinsics.go#L76
	// Decode the expression, hand it off to goformation to unmarshall
	// with the nested Base64 strings inside.
	// then turn it into a Map, look at the key and determine what kind of reference it is.

	// Is there a chance it's a Base64 encoded string? That indicates
	// it's a goformation reference.

	base64Decoded, base64DecodedErr := base64.StdEncoding.DecodeString(expr)
	if base64DecodedErr != nil {
		// It's possible it's a plain old literal...test it
		// Ref: https://docs.aws.amazon.com/general/latest/gr/aws-arns-and-namespaces.html
		if strings.HasPrefix(expr, "arn:aws") {
			return &resourceRef{
				RefType:      resourceLiteral,
				ResourceName: expr,
			}, nil
		}
	}

	// Setup the intrinsic handlers so that when we unmarshal
	// the encoded functions we keep track of what we found

	var hookedResourceRef *resourceRef

	hookedOverrides := map[string]gofintrinsics.IntrinsicHandler{
		"Ref": func(name string, input interface{}, template interface{}) interface{} {
			hookedResourceRef = &resourceRef{
				RefType:      resourceRefFunc,
				ResourceName: input.(string),
			}
			return nil
		},
		"Fn::GetAtt": func(name string, input interface{}, template interface{}) interface{} {
			// The input should be an array...
			inputArr, inputArrOk := input.([]interface{})
			if !inputArrOk {
				return nil
			}
			inputStringElemZero, inputStringElemZeroOk := inputArr[0].(string)
			if !inputStringElemZeroOk {
				return nil
			}
			hookedResourceRef = &resourceRef{
				RefType:      resourceGetAttrFunc,
				ResourceName: inputStringElemZero,
			}
			return nil
		},
	}

	procOptions := &gofintrinsics.ProcessorOptions{
		IntrinsicHandlerOverrides: hookedOverrides,
	}
	_, processedJSONErr := gofintrinsics.ProcessJSON(base64Decoded, procOptions)
	if processedJSONErr != nil {
		return nil, processedJSONErr
	}
	// Whatever we have at this point is what it is...
	return hookedResourceRef, nil
}

// isResolvedResourceType is a utility function to determine if a resolved
// reference is a given type. If it is a literal, the literalTokenIndicator
// substring match is used for the predicate. If it is a resource provisioned
// by this template, the &gocf.RESOURCE_TYPE{} will be used via reflection
// Example:
// isResolvedResourceType(resourceRef, template, ":dynamodb:", &gocf.DynamoDBTable{}) {
//
func isResolvedResourceType(resource *resourceRef,
	template *gof.Template,
	literalTokenIndicator string,
	templateType interface{}) bool {

	if resource.RefType == resourceLiteral {
		return strings.Contains(resource.ResourceName, literalTokenIndicator)
	}

	// Dynamically provisioned resource included in the template definition?
	existingResource, existingResourceExists := template.Resources[resource.ResourceName]
	if existingResourceExists {
		if reflect.TypeOf(existingResource) == reflect.TypeOf(templateType) {
			return true
		}
	}
	return false

}

// visitResolvedEventSourceMapping is a utility function that visits all
// the EventSourceMapping entries for the given lambdaAWSInfo struct
func visitResolvedEventSourceMapping(visitor resolvedResourceVisitor,
	lambdaAWSInfos []*LambdaAWSInfo,
	template *gof.Template,
	logger *zerolog.Logger) error {

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
				"Visiting event source mapping: %#v",
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
