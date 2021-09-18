package decorator

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	gof "github.com/awslabs/goformation/v5/cloudformation"
	gofcodedeploy "github.com/awslabs/goformation/v5/cloudformation/codedeploy"
	gofiam "github.com/awslabs/goformation/v5/cloudformation/iam"
	goflambda "github.com/awslabs/goformation/v5/cloudformation/lambda"
	gofpolicies "github.com/awslabs/goformation/v5/cloudformation/policies"
	sparta "github.com/mweagle/Sparta"

	"github.com/rs/zerolog"
)

// codeDeployLambdaUpdateDecorator is the per-function decorator
// that adds the necessary information for CodeDeploy
func codeDeployLambdaUpdateDecorator(updateType string,
	codeDeployApplicationName string,
	codeDeployRoleName string) sparta.TemplateDecorator {
	return func(ctx context.Context,
		serviceName string,
		lambdaResourceName string,
		lambdaResource *goflambda.Function,
		resourceMetadata map[string]interface{},
		lambdaFunctionCode *goflambda.Function_Code,
		buildID string,
		template *gof.Template,
		logger *zerolog.Logger) (context.Context, error) {

		safeDeployResourceName := func(resType string) string {
			return sparta.CloudFormationResourceName(serviceName,
				lambdaResourceName,
				resType)
		}
		// Create the AWS::Lambda::Version, with DeletionPolicy=Retain
		versionResourceName := safeDeployResourceName("version" + buildID)
		versionResource := &goflambda.Version{
			FunctionName: gof.Ref(lambdaResourceName),
		}
		versionResource.AWSCloudFormationDeletionPolicy = "Retain"
		template.Resources[versionResourceName] = versionResource

		// Create the AWS::CodeDeploy::DeploymentGroup entry that includes a reference
		// to the IAM role
		codeDeploymentGroupResourceName := safeDeployResourceName("deploymentGroup")
		codeDeploymentGroup := &gofcodedeploy.DeploymentGroup{
			ApplicationName: gof.Ref(codeDeployApplicationName),
			AutoRollbackConfiguration: &gofcodedeploy.DeploymentGroup_AutoRollbackConfiguration{
				Enabled: true,
				Events: []string{"DEPLOYMENT_FAILURE",
					"DEPLOYMENT_STOP_ON_ALARM",
					"DEPLOYMENT_STOP_ON_REQUEST"},
			},
			ServiceRoleArn:       gof.GetAtt(codeDeployRoleName, "Arn"),
			DeploymentConfigName: fmt.Sprintf("CodeDeployDefault.Lambda%s", updateType),
			DeploymentStyle: &gofcodedeploy.DeploymentGroup_DeploymentStyle{
				DeploymentType:   "BLUE_GREEN",
				DeploymentOption: "WITH_TRAFFIC_CONTROL",
			},
		}
		template.Resources[codeDeploymentGroupResourceName] = codeDeploymentGroup

		// Create the Alias entry...
		aliasResourceName := safeDeployResourceName("alias")
		aliasResource := &goflambda.Alias{
			FunctionVersion: gof.GetAtt(versionResourceName, "Version"),
			FunctionName:    gof.Ref(lambdaResourceName),
			Name:            "live",
		}
		aliasResource.AWSCloudFormationUpdatePolicy = &gofpolicies.UpdatePolicy{
			CodeDeployLambdaAliasUpdate: &gofpolicies.CodeDeployLambdaAliasUpdate{
				ApplicationName:     codeDeployApplicationName,
				DeploymentGroupName: codeDeploymentGroupResourceName,
			},
		}

		template.Resources[aliasResourceName] = aliasResource
		return ctx, nil
	}
}

// CodeDeployServiceUpdateDecorator is a service level decorator that attaches
// the CodeDeploy safe update to an upgrade operation.
// Ref: https://github.com/awslabs/serverless-application-model/blob/master/docs/safe_lambda_deployments.rst
//
func CodeDeployServiceUpdateDecorator(updateType string,
	lambdaFuncs []*sparta.LambdaAWSInfo,
	preHook *sparta.LambdaAWSInfo,
	postHook *sparta.LambdaAWSInfo) sparta.ServiceDecoratorHookFunc {
	// Define the names that are shared
	codeDeployApplicationName := sparta.CloudFormationResourceName("SafeDeploy",
		"deployment",
		"application")
	codeDeployRoleResourceName := sparta.CloudFormationResourceName("SafeDeploy",
		"deployment",
		"role")

	// Add the Execution status
	// See: https://github.com/awslabs/serverless-application-model/blob/master/docs/safe_lambda_deployments.rst#traffic-shifting-using-codedeploy
	for _, eachFunc := range []*sparta.LambdaAWSInfo{preHook, postHook} {
		if eachFunc != nil {
			eachFunc.RoleDefinition.Privileges = append(preHook.RoleDefinition.Privileges,
				sparta.IAMRolePrivilege{
					Actions: []string{"codedeploy:PutLifecycleEventHookExecutionStatus"},
					Resource: []string{"",
						"arn:aws:codedeploy:",
						gof.Ref("AWS::Region"),
						":",
						gof.Ref("AWS::AccountId"),
						":deploymentgroup:",
						codeDeployApplicationName,
						"/*"},
				},
			)
		}
	}

	// Add the decorator to each lambda
	for _, eachLambda := range lambdaFuncs {
		safeDeployDecorator := codeDeployLambdaUpdateDecorator(updateType,
			codeDeployApplicationName,
			codeDeployRoleResourceName)
		eachLambda.Decorators = append(eachLambda.Decorators,
			sparta.TemplateDecoratorHookFunc(safeDeployDecorator))
	}

	// Return the service decorator...
	return func(ctx context.Context,
		serviceName string,
		template *gof.Template,
		lambdaFunctionCode *goflambda.Function_Code,
		buildID string,
		awsSession *session.Session,
		noop bool,
		logger *zerolog.Logger) (context.Context, error) {
		// So what we really need to do is walk over all the lambda functions in the template
		// and setup all the Deployment groups...
		codeDeployApplication := &gofcodedeploy.Application{
			ComputePlatform: "Lambda",
		}
		template.Resources[codeDeployApplicationName] = codeDeployApplication
		// Create the CodeDeploy role
		// Ensure there is an IAM role for this...
		// CodeDeployServiceRole

		codeDeployRoleResource := &gofiam.Role{
			ManagedPolicyArns: []string{"arn:aws:iam::aws:policy/service-role/AWSCodeDeployRoleForLambda"},
			AssumeRolePolicyDocument: sparta.ArbitraryJSONObject{
				"Version": "2012-10-17",
				"Statement": []sparta.ArbitraryJSONObject{{
					"Action": []string{"sts:AssumeRole"},
					"Effect": "Allow",
					"Principal": sparta.ArbitraryJSONObject{
						"Service": []string{"codedeploy.amazonaws.com"},
					}},
				},
			},
		}
		template.Resources[codeDeployRoleResourceName] = codeDeployRoleResource

		// Ship it...
		return ctx, nil
	}
}
