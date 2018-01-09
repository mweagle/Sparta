package decorator

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/mweagle/Sparta"
	gocf "github.com/mweagle/go-cloudformation"
)

// codeDeployLambdaUpdateDecorator is the per-function decorator
// that adds the necessary
func codeDeployLambdaUpdateDecorator(updateType string,
	codeDeployApplicationName string,
	codeDeployRoleName string) sparta.TemplateDecorator {
	return func(serviceName string,
		lambdaResourceName string,
		lambdaResource gocf.LambdaFunction,
		resourceMetadata map[string]interface{},
		S3Bucket string,
		S3Key string,
		buildID string,
		template *gocf.Template,
		context map[string]interface{},
		logger *logrus.Logger) error {

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
		return nil
	}
}

// CodeDeployServiceUpdateDecorator is a service level decorator that attaches
// the CodeDeploy safe update to an upgrade operation
func CodeDeployServiceUpdateDecorator(updateType string,
	lambdaFuncs []*sparta.LambdaAWSInfo,
	preHook *sparta.LambdaAWSInfo,
	postHook *sparta.LambdaAWSInfo) sparta.ServiceDecoratorHook {

	// Define the names that are shared
	codeDeployApplicationName := sparta.CloudFormationResourceName("SafeDeploy",
		"deployment",
		"application")
	codeDeployRoleResourceName := sparta.CloudFormationResourceName("SafeDeploy",
		"deployment",
		"role")

	// Add the decorator to each lambda
	for _, eachLambda := range lambdaFuncs {
		safeDeployDecorator := codeDeployLambdaUpdateDecorator(updateType,
			codeDeployApplicationName,
			codeDeployRoleResourceName)
		eachLambda.Decorators = append(eachLambda.Decorators,
			sparta.TemplateDecoratorHookFunc(safeDeployDecorator))
	}

	// Return the service decorator...
	return func(context map[string]interface{},
		serviceName string,
		template *gocf.Template,
		S3Bucket string,
		buildID string,
		awsSession *session.Session,
		noop bool,
		logger *logrus.Logger) error {
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
				"Statement": []sparta.ArbitraryJSONObject{
					sparta.ArbitraryJSONObject{
						"Action": []string{"sts:AssumeRole"},
						"Effect": "Allow",
						"Principal": sparta.ArbitraryJSONObject{
							"Service": []string{"codedeploy.amazonaws.com"},
						},
					},
				},
			},
		}
		template.AddResource(codeDeployRoleResourceName, codeDeployRoleResource)

		// Ship it...
		return nil
	}
}
