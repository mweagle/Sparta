package decorator

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	sparta "github.com/mweagle/Sparta"
	gocf "github.com/mweagle/go-cloudformation"
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
		lambdaResource gocf.LambdaFunction,
		resourceMetadata map[string]interface{},
		lambdaFunctionCode *gocf.LambdaFunctionCode,
		buildID string,
		template *gocf.Template,
		logger *zerolog.Logger) (context.Context, error) {

		safeDeployResourceName := func(resType string) string {
			return sparta.CloudFormationResourceName(serviceName,
				lambdaResourceName,
				resType)
		}
		// Create the AWS::Lambda::Version, with DeletionPolicy=Retain
		versionResourceName := safeDeployResourceName("version" + buildID)
		versionResource := &gocf.LambdaVersion{
			FunctionName: gocf.Ref(lambdaResourceName).String(),
		}
		entry := template.AddResource(versionResourceName, versionResource)
		entry.DeletionPolicy = "Retain"

		// Create the AWS::CodeDeploy::DeploymentGroup entry that includes a reference
		// to the IAM role
		codeDeploymentGroupResourceName := safeDeployResourceName("deploymentGroup")
		codeDeploymentGroup := &gocf.CodeDeployDeploymentGroup{
			ApplicationName: gocf.Ref(codeDeployApplicationName).String(),
			AutoRollbackConfiguration: &gocf.CodeDeployDeploymentGroupAutoRollbackConfiguration{
				Enabled: gocf.Bool(true),
				Events: gocf.StringList(gocf.String("DEPLOYMENT_FAILURE"),
					gocf.String("DEPLOYMENT_STOP_ON_ALARM"),
					gocf.String("DEPLOYMENT_STOP_ON_REQUEST")),
			},
			ServiceRoleArn:       gocf.GetAtt(codeDeployRoleName, "Arn"),
			DeploymentConfigName: gocf.String(fmt.Sprintf("CodeDeployDefault.Lambda%s", updateType)),
			DeploymentStyle: &gocf.CodeDeployDeploymentGroupDeploymentStyle{
				DeploymentType:   gocf.String("BLUE_GREEN"),
				DeploymentOption: gocf.String("WITH_TRAFFIC_CONTROL"),
			},
		}
		template.AddResource(codeDeploymentGroupResourceName, codeDeploymentGroup)
		// Create the Alias entry...
		aliasResourceName := safeDeployResourceName("alias")
		aliasResource := &gocf.LambdaAlias{
			FunctionVersion: gocf.GetAtt(versionResourceName, "Version").String(),
			FunctionName:    gocf.Ref(lambdaResourceName).String(),
			Name:            gocf.String("live"),
		}
		aliasEntry := template.AddResource(aliasResourceName, aliasResource)
		aliasEntry.UpdatePolicy = &gocf.UpdatePolicy{
			CodeDeployLambdaAliasUpdate: &gocf.UpdatePolicyCodeDeployLambdaAliasUpdate{
				ApplicationName:     gocf.Ref(codeDeployApplicationName).String(),
				DeploymentGroupName: gocf.Ref(codeDeploymentGroupResourceName).String(),
			},
		}
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
					Resource: gocf.Join("",
						gocf.String("arn:aws:codedeploy:"),
						gocf.Ref("AWS::Region"),
						gocf.String(":"),
						gocf.Ref("AWS::AccountId"),
						gocf.String(":deploymentgroup:"),
						gocf.String(codeDeployApplicationName),
						gocf.String("/*"),
					)},
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
		template *gocf.Template,
		lambdaFunctionCode *gocf.LambdaFunctionCode,
		buildID string,
		awsSession *session.Session,
		noop bool,
		logger *zerolog.Logger) (context.Context, error) {
		// So what we really need to do is walk over all the lambda functions in the template
		// and setup all the Deployment groups...
		codeDeployApplication := &gocf.CodeDeployApplication{
			ComputePlatform: gocf.String("Lambda"),
		}
		template.AddResource(codeDeployApplicationName, codeDeployApplication)
		// Create the CodeDeploy role
		// Ensure there is an IAM role for this...
		// CodeDeployServiceRole

		codeDeployRoleResource := gocf.IAMRole{
			ManagedPolicyArns: gocf.StringList(gocf.String("arn:aws:iam::aws:policy/service-role/AWSCodeDeployRoleForLambda")),
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
		template.AddResource(codeDeployRoleResourceName, codeDeployRoleResource)

		// Ship it...
		return ctx, nil
	}
}
