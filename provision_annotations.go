// +build !lambdabinary

package sparta

import (
	"encoding/json"
	"reflect"
	"runtime"
	"strings"

	spartaCF "github.com/mweagle/Sparta/aws/cloudformation"
	spartaIAM "github.com/mweagle/Sparta/aws/iam"
	gocf "github.com/mweagle/go-cloudformation"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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

// resolveResourceRef takes an interface representing a dynamic ARN
// and tries to determine the CloudFormation resource name it resolves to
func resolveResourceRef(expr interface{}) (*resourceRef, error) {

	// Is ther any chance it's just a string?
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
	if json.Unmarshal(marshalled, &getAttFunc) == nil &&
		len(getAttFunc.Resource) != 0 {
		return &resourceRef{
			RefType:      resourceGetAttrFunc,
			ResourceName: getAttFunc.Resource,
		}, nil
	}

	// Nope
	return nil, nil
}

func eventSourceMappingPoliciesForResource(resource *resourceRef,
	template *gocf.Template,
	logger *logrus.Logger) ([]spartaIAM.PolicyStatement, error) {
	// String literal?
	policyStatements := []spartaIAM.PolicyStatement{}
	if resource.RefType == resourceLiteral {
		if strings.Contains(resource.ResourceName, ":dynamodb:") {
			policyStatements = append(policyStatements, CommonIAMStatements.DynamoDB...)
		} else if strings.Contains(resource.ResourceName, ":kinesis:") {
			policyStatements = append(policyStatements, CommonIAMStatements.Kinesis...)
		} else {
			logger.WithFields(logrus.Fields{
				"ARN": resource.ResourceName,
			}).Debug("No additional permissions found for static resource type")
		}
	} else {
		existingResource, existingResourceExists := template.Resources[resource.ResourceName]

		if !existingResourceExists {
			return policyStatements, errors.Errorf("Failed to find resource %s in template",
				resource.ResourceName)
		}
		// What permissions do we need to add?
		switch existingResource.Properties.(type) {
		case gocf.DynamoDBTable:
			policyStatements = append(policyStatements, CommonIAMStatements.DynamoDB...)
		case gocf.KinesisStream:
			policyStatements = append(policyStatements, CommonIAMStatements.Kinesis...)
		default:
			logger.WithFields(logrus.Fields{
				"ResourceType": existingResource.Properties.CfnResourceType(),
			}).Debug("No additional permissions found for dynamic resource reference type")
		}
	}
	return policyStatements, nil
}

// annotationFunc represents an internal annotation function
// called to stich the template together
type annotationFunc func(lambdaAWSInfos []*LambdaAWSInfo,
	template *gocf.Template,
	logger *logrus.Logger) error

func annotateBuildInformation(lambdaAWSInfo *LambdaAWSInfo,
	template *gocf.Template,
	buildID string,
	logger *logrus.Logger) (*gocf.Template, error) {

	// Add the build id s.t. the logger can get stamped...
	if lambdaAWSInfo.Options == nil {
		lambdaAWSInfo.Options = &LambdaFunctionOptions{}
	}
	lambdaEnvironment := lambdaAWSInfo.Options.Environment
	if lambdaEnvironment == nil {
		lambdaAWSInfo.Options.Environment = make(map[string]*gocf.StringExpr)
	}
	return template, nil
}

func annotateDiscoveryInfo(lambdaAWSInfo *LambdaAWSInfo,
	template *gocf.Template,
	logger *logrus.Logger) (*gocf.Template, error) {
	depMap := make(map[string]string)

	// Update the metdata with a reference to the output of each
	// depended on item...
	for _, eachDependsKey := range lambdaAWSInfo.DependsOn {
		dependencyText, dependencyTextErr := discoveryResourceInfoForDependency(template, eachDependsKey, logger)
		if dependencyTextErr != nil {
			return nil, errors.Wrapf(dependencyTextErr, "Failed to determine discovery info for resource")
		}
		depMap[eachDependsKey] = string(dependencyText)
	}
	if lambdaAWSInfo.Options == nil {
		lambdaAWSInfo.Options = &LambdaFunctionOptions{}
	}
	lambdaEnvironment := lambdaAWSInfo.Options.Environment
	if lambdaEnvironment == nil {
		lambdaAWSInfo.Options.Environment = make(map[string]*gocf.StringExpr)
	}

	discoveryInfo, discoveryInfoErr := discoveryInfoForResource(lambdaAWSInfo.LogicalResourceName(),
		depMap)
	if discoveryInfoErr != nil {
		return nil, errors.Wrap(discoveryInfoErr, "Failed to create resource discovery info")
	}

	// Update the env map
	lambdaAWSInfo.Options.Environment[envVarDiscoveryInformation] = discoveryInfo
	return template, nil
}

func annotateCodePipelineEnvironments(lambdaAWSInfo *LambdaAWSInfo, logger *logrus.Logger) {
	if nil != codePipelineEnvironments {
		if nil == lambdaAWSInfo.Options {
			lambdaAWSInfo.Options = defaultLambdaFunctionOptions()
		}
		if nil == lambdaAWSInfo.Options.Environment {
			lambdaAWSInfo.Options.Environment = make(map[string]*gocf.StringExpr)
		}
		for _, eachEnvironment := range codePipelineEnvironments {

			logger.WithFields(logrus.Fields{
				"Environment":    eachEnvironment,
				"LambdaFunction": lambdaAWSInfo.lambdaFunctionName(),
			}).Debug("Annotating Lambda environment for CodePipeline")

			for eachKey := range eachEnvironment {
				lambdaAWSInfo.Options.Environment[eachKey] = gocf.Ref(eachKey).String()
			}
		}
	}
}

func annotateEventSourceMappings(lambdaAWSInfos []*LambdaAWSInfo,
	template *gocf.Template,
	logger *logrus.Logger) error {

	//
	// BEGIN
	// Inline closure to handle the update of a lambda function that includes
	// an eventSourceMapping entry.
	annotatePermissions := func(lambdaAWSInfo *LambdaAWSInfo,
		eventSourceMapping *EventSourceMapping,
		resource *resourceRef) error {

		annotateStatements, annotateStatementsErr := eventSourceMappingPoliciesForResource(resource,
			template,
			logger)

		// Early exit?
		if annotateStatementsErr != nil {
			return annotateStatementsErr
		} else if len(annotateStatements) <= 0 {
			return nil
		}
		// If we have statements, let's go ahead and ensure they
		// include a reference to our ARN
		populatedStatements := []spartaIAM.PolicyStatement{}
		for _, eachStatement := range annotateStatements {
			populatedStatements = append(populatedStatements,
				spartaIAM.PolicyStatement{
					Action:   eachStatement.Action,
					Effect:   "Allow",
					Resource: spartaCF.DynamicValueToStringExpr(eventSourceMapping.EventSourceArn).String(),
				})
		}

		// Something to push onto the resource. The resource
		// is hopefully defined in this template. It technically
		// could be a string literal, in which case we're not going
		// to have a lot of luck with that...
		cfResource, cfResourceOk := template.Resources[lambdaAWSInfo.LogicalResourceName()]
		if !cfResourceOk {
			return errors.Errorf("Unable to locate lambda function for annotation")
		}
		lambdaResource, lambdaResourceOk := cfResource.Properties.(gocf.LambdaFunction)
		if !lambdaResourceOk {
			return errors.Errorf("CloudFormation resource exists, but is incorrect type: %s (%v)",
				cfResource.Properties.CfnResourceType(),
				cfResource.Properties)
		}
		// Ok, go get the IAM Role
		resourceRef, resourceRefErr := resolveResourceRef(lambdaResource.Role)
		if resourceRefErr != nil {
			return errors.Wrapf(resourceRefErr, "Failed to resolve IAM Role for event source mappings: %#v",
				lambdaResource.Role)
		}
		// If it's not nil and also not a literal, go ahead and try and update it
		if resourceRef != nil &&
			resourceRef.RefType != resourceLiteral {
			// Excellent, go ahead and find the role in the template
			// and stitch things together
			iamRole, iamRoleExists := template.Resources[resourceRef.ResourceName]
			if !iamRoleExists {
				return errors.Errorf("IAM role not found: %s", resourceRef.ResourceName)
			}
			// Coerce to the IAMRole and update the statements
			typedIAMRole, typedIAMRoleOk := iamRole.Properties.(gocf.IAMRole)
			if !typedIAMRoleOk {
				return errors.Errorf("Failed to type convert iamRole to proper IAMRole resource")
			}
			policyList := typedIAMRole.Policies
			if policyList == nil {
				policyList = &gocf.IAMRolePolicyList{}
			}
			*policyList = append(*policyList,
				gocf.IAMRolePolicy{
					PolicyDocument: ArbitraryJSONObject{
						"Version":   "2012-10-17",
						"Statement": populatedStatements,
					},
					PolicyName: gocf.String("LambdaEventSourceMappingPolicy"),
				})
			typedIAMRole.Policies = policyList
		}
		return nil
	}
	//
	// END

	// Iterate through every lambda function. If there is an EventSourceMapping
	// that points to a piece of infastructure provisioned by this stack,
	// figure out the resourcename used by that infrastructure and ensure
	// that the IAMRole the lambda function is using includes permissions to
	// perform the necessary pull-based operations against the source.
	for _, eachLambda := range lambdaAWSInfos {
		for _, eachEventSource := range eachLambda.EventSourceMappings {
			resourceRef, resourceRefErr := resolveResourceRef(eachEventSource.EventSourceArn)
			if resourceRefErr != nil {
				return errors.Wrapf(resourceRefErr, "Failed to resolve EventSourceArn: %#v",
					eachEventSource)
			}
			// At this point everything is a string, so we need to unmarshall
			// and see if the Arn is supplied by either a Ref or a GetAttr
			// function. In those cases, we need to look around in the template
			// to go from: EventMapping -> Type -> Lambda -> LambdaIAMRole
			// so that we can add the permissions
			if resourceRef != nil {
				annotationErr := annotatePermissions(eachLambda,
					eachEventSource,
					resourceRef)
				// Anything go wrong?
				if annotationErr != nil {
					return errors.Wrapf(annotationErr,
						"Failed to annotate template for EventSourceMapping: %#v", eachEventSource)
				}
			}
		}
	}
	return nil
}

func annotateMaterializedTemplate(
	lambdaAWSInfos []*LambdaAWSInfo,
	template *gocf.Template,
	logger *logrus.Logger) (*gocf.Template, error) {
	// Setup the annotation functions
	annotationFuncs := []annotationFunc{
		annotateEventSourceMappings,
	}
	for _, eachAnnotationFunc := range annotationFuncs {
		funcName := runtime.FuncForPC(reflect.ValueOf(eachAnnotationFunc).Pointer()).Name()
		logger.WithFields(logrus.Fields{
			"Annotator": funcName,
		}).Debug("Evaluating annotator")

		annotationErr := eachAnnotationFunc(lambdaAWSInfos,
			template,
			logger)
		if annotationErr != nil {
			return nil, errors.Wrapf(annotationErr,
				"Function %s failed to annotate template",
				funcName)
		}
	}
	return template, nil
}
