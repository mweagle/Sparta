// +build !lambdabinary

package sparta

import (
	"reflect"
	"runtime"

	spartaCF "github.com/mweagle/Sparta/aws/cloudformation"
	spartaIAM "github.com/mweagle/Sparta/aws/iam"
	gocf "github.com/mweagle/go-cloudformation"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// eventSourceMappingPoliciesForResource returns the IAM specific privileges for each
// type of supported AWS Lambda EventSourceMapping
func eventSourceMappingPoliciesForResource(resource *resourceRef,
	template *gocf.Template,
	logger *zerolog.Logger) ([]spartaIAM.PolicyStatement, error) {

	policyStatements := []spartaIAM.PolicyStatement{}

	if isResolvedResourceType(resource, template, ":dynamodb:", &gocf.DynamoDBTable{}) {
		policyStatements = append(policyStatements, CommonIAMStatements.DynamoDB...)
	} else if isResolvedResourceType(resource, template, ":kinesis:", &gocf.KinesisStream{}) {
		policyStatements = append(policyStatements, CommonIAMStatements.Kinesis...)
	} else if isResolvedResourceType(resource, template, ":sqs:", &gocf.SQSQueue{}) {
		policyStatements = append(policyStatements, CommonIAMStatements.SQS...)
	} else {
		logger.Debug().
			Interface("Resource", resource).
			Msg("No additional EventSource IAM permissions found for event type")
	}
	return policyStatements, nil
}

// annotationFunc represents an internal annotation function
// called to stich the template together
type annotationFunc func(lambdaAWSInfos []*LambdaAWSInfo,
	template *gocf.Template,
	logger *zerolog.Logger) error

func annotateBuildInformation(lambdaAWSInfo *LambdaAWSInfo,
	template *gocf.Template,
	buildID string,
	logger *zerolog.Logger) (*gocf.Template, error) {

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
	logger *zerolog.Logger) (*gocf.Template, error) {
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

func annotateCodePipelineEnvironments(lambdaAWSInfo *LambdaAWSInfo, logger *zerolog.Logger) {
	if nil != codePipelineEnvironments {
		if nil == lambdaAWSInfo.Options {
			lambdaAWSInfo.Options = defaultLambdaFunctionOptions()
		}
		if nil == lambdaAWSInfo.Options.Environment {
			lambdaAWSInfo.Options.Environment = make(map[string]*gocf.StringExpr)
		}
		for _, eachEnvironment := range codePipelineEnvironments {

			logger.Debug().
				Interface("Environment", eachEnvironment).
				Interface("LambdaFunction", lambdaAWSInfo.lambdaFunctionName()).
				Msg("Annotating Lambda environment for CodePipeline")

			for eachKey := range eachEnvironment {
				lambdaAWSInfo.Options.Environment[eachKey] = gocf.Ref(eachKey).String()
			}
		}
	}
}

func annotateEventSourceMappings(lambdaAWSInfos []*LambdaAWSInfo,
	template *gocf.Template,
	logger *zerolog.Logger) error {

	//
	// BEGIN
	// Inline closure to handle the update of a lambda function that includes
	// an eventSourceMapping entry.
	annotatePermissions := func(lambdaAWSInfo *LambdaAWSInfo,
		eventSourceMapping *EventSourceMapping,
		mappingIndex int,
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
			resourceRef.RefType != resourceLiteral &&
			resourceRef.RefType != resourceStringFunc {
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

	annotationErr := visitResolvedEventSourceMapping(annotatePermissions,
		lambdaAWSInfos,
		template,
		logger)

	if annotationErr != nil {
		return errors.Wrapf(annotationErr,
			"Failed to annotate template for EventSourceMappings")
	}
	return nil
}

func annotateMaterializedTemplate(
	lambdaAWSInfos []*LambdaAWSInfo,
	template *gocf.Template,
	logger *zerolog.Logger) (*gocf.Template, error) {
	// Setup the annotation functions
	annotationFuncs := []annotationFunc{
		annotateEventSourceMappings,
	}
	for _, eachAnnotationFunc := range annotationFuncs {
		funcName := runtime.FuncForPC(reflect.ValueOf(eachAnnotationFunc).Pointer()).Name()
		logger.Debug().
			Str("Annotator", funcName).
			Msg("Evaluating annotator")

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
