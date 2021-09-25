//go:build !lambdabinary
// +build !lambdabinary

package sparta

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"

	gofiam "github.com/awslabs/goformation/v5/cloudformation/iam"

	gof "github.com/awslabs/goformation/v5/cloudformation"
	gofamazonmq "github.com/awslabs/goformation/v5/cloudformation/amazonmq"
	gofdynamodb "github.com/awslabs/goformation/v5/cloudformation/dynamodb"
	gofkinesis "github.com/awslabs/goformation/v5/cloudformation/kinesis"
	goflambda "github.com/awslabs/goformation/v5/cloudformation/lambda"
	gofmsk "github.com/awslabs/goformation/v5/cloudformation/msk"
	gofsqs "github.com/awslabs/goformation/v5/cloudformation/sqs"
	spartaCF "github.com/mweagle/Sparta/aws/cloudformation"
	spartaIAM "github.com/mweagle/Sparta/aws/iam"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// eventSourceMappingPoliciesForResource returns the IAM specific privileges for each
// type of supported AWS Lambda EventSourceMapping
func eventSourceMappingPoliciesForResource(resource *resourceRef,
	template *gof.Template,
	logger *zerolog.Logger) ([]spartaIAM.PolicyStatement, error) {

	policyStatements := []spartaIAM.PolicyStatement{}

	// Map the types to their common statements
	resourceToStatementsMap := map[gof.Resource][]spartaIAM.PolicyStatement{
		&gofdynamodb.Table{}:  CommonIAMStatements.DynamoDB,
		&gofkinesis.Stream{}:  CommonIAMStatements.Kinesis,
		&gofsqs.Queue{}:       CommonIAMStatements.SQS,
		&gofmsk.Cluster{}:     CommonIAMStatements.MSKCluster,
		&gofamazonmq.Broker{}: CommonIAMStatements.AmazonMQBroker,
	}
	preLengthStatements := len(policyStatements)
	for eachResource, eachPolicyStatements := range resourceToStatementsMap {
		// Split the type
		splitTypes := strings.Split(eachResource.AWSCloudFormationType(), "::")
		if len(splitTypes) == 3 {
			typeHint := fmt.Sprintf(":%s:", strings.ToLower(splitTypes[1]))
			if isResolvedResourceType(resource, template, typeHint, eachResource) {
				policyStatements = append(policyStatements, eachPolicyStatements...)
			}
		}
	}
	if preLengthStatements == len(policyStatements) {
		logger.Info().
			Interface("Resource", resource).
			Msg("No additional EventSource IAM permissions found for event type")
	}
	return policyStatements, nil
}

// annotationFunc represents an internal annotation function
// called to stich the template together
type annotationFunc func(lambdaAWSInfos []*LambdaAWSInfo,
	template *gof.Template,
	logger *zerolog.Logger) error

func annotateBuildInformation(lambdaAWSInfo *LambdaAWSInfo,
	template *gof.Template,
	buildID string,
	logger *zerolog.Logger) (*gof.Template, error) {

	// Add the build id s.t. the logger can get stamped...
	if lambdaAWSInfo.Options == nil {
		lambdaAWSInfo.Options = &LambdaFunctionOptions{}
	}
	lambdaEnvironment := lambdaAWSInfo.Options.Environment
	if lambdaEnvironment == nil {
		lambdaAWSInfo.Options.Environment = make(map[string]string)
	}
	return template, nil
}

func annotateDiscoveryInfo(lambdaAWSInfo *LambdaAWSInfo,
	template *gof.Template,
	logger *zerolog.Logger) (*gof.Template, error) {
	depMap := make(map[string]string)

	// Update the metdata with a reference to the output of each
	// depended on item...
	for _, eachDependsKey := range lambdaAWSInfo.DependsOn {
		dependencyText, dependencyTextErr := discoveryResourceInfoForDependency(template,
			eachDependsKey,
			logger)
		if dependencyTextErr != nil {
			return nil, errors.Wrapf(dependencyTextErr,
				"Failed to determine discovery info for resource")
		}
		depMap[eachDependsKey] = string(dependencyText)
	}
	if lambdaAWSInfo.Options == nil {
		lambdaAWSInfo.Options = &LambdaFunctionOptions{}
	}
	lambdaEnvironment := lambdaAWSInfo.Options.Environment
	if lambdaEnvironment == nil {
		lambdaAWSInfo.Options.Environment = make(map[string]string)
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
			lambdaAWSInfo.Options.Environment = make(map[string]string)
		}
		for _, eachEnvironment := range codePipelineEnvironments {

			logger.Debug().
				Interface("Environment", eachEnvironment).
				Interface("LambdaFunction", lambdaAWSInfo.lambdaFunctionName()).
				Msg("Annotating Lambda environment for CodePipeline")

			for eachKey := range eachEnvironment {
				lambdaAWSInfo.Options.Environment[eachKey] = gof.Ref(eachKey)
			}
		}
	}
}

func annotateEventSourceMappings(lambdaAWSInfos []*LambdaAWSInfo,
	template *gof.Template,
	logger *zerolog.Logger) error {

	// TODO - this is brittle

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
					Resource: spartaCF.DynamicValueToStringExpr(eventSourceMapping.EventSourceArn),
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
		lambdaResource, lambdaResourceOk := cfResource.(*goflambda.Function)
		if !lambdaResourceOk {
			return errors.Errorf("CloudFormation resource exists, but is incorrect type: %s (%v)",
				cfResource.AWSCloudFormationType(),
				lambdaAWSInfo.LogicalResourceName())
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
			typedIAMRole, typedIAMRoleOk := iamRole.(*gofiam.Role)
			if !typedIAMRoleOk {
				return errors.Errorf("Failed to type convert iamRole to proper IAMRole resource")
			}
			policyList := typedIAMRole.Policies
			if policyList == nil {
				policyList = []gofiam.Role_Policy{}
			}
			policyList = append(policyList,
				gofiam.Role_Policy{
					PolicyDocument: ArbitraryJSONObject{
						"Version":   "2012-10-17",
						"Statement": populatedStatements,
					},
					PolicyName: "LambdaEventSourceMappingPolicy",
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
	template *gof.Template,
	logger *zerolog.Logger) (*gof.Template, error) {
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
