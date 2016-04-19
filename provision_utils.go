package sparta

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mweagle/cloudformationresources"

	gocf "github.com/crewjam/go-cloudformation"

	"github.com/Sirupsen/logrus"
)

const salt = "213EA743-A98F-499D-8FEF-B87015FE13E7"

// PushSourceConfigurationActions map stores common IAM Policy Actions for Lambda
// push-source configuration management.
// The configuration is handled by CustomResources inserted into the generated
// CloudFormation template.
var PushSourceConfigurationActions = map[string][]string{}

func init() {

	PushSourceConfigurationActions[cloudformationresources.S3LambdaEventSource] = []string{"s3:GetBucketLocation",
		"s3:GetBucketNotification",
		"s3:PutBucketNotification",
		"s3:GetBucketNotificationConfiguration",
		"s3:PutBucketNotificationConfiguration"}

	PushSourceConfigurationActions[APIGatewayPrincipal] = []string{"apigateway:*",
		"lambda:AddPermission",
		"lambda:RemovePermission",
		"lambda:GetPolicy"}

	cloudWatchLogsRegions := []string{
		"us-east-1",
		"us-west-2",
		"us-west-1",
		"eu-west-1",
		"eu-central-1",
		"ap-southeast-1",
		"ap-northeast-1",
		"ap-southeast-2",
		"ap-northeast-2",
	}
	for _, eachCloudWatchLogsRegion := range cloudWatchLogsRegions {
		regionalPrincipal := fmt.Sprintf("logs.%s.amazonaws.com", eachCloudWatchLogsRegion)
		PushSourceConfigurationActions[regionalPrincipal] = []string{"logs:DescribeSubscriptionFilters",
			"logs:DeleteSubscriptionFilter",
			"logs:PutSubscriptionFilter"}
	}

	PushSourceConfigurationActions[SESPrincipal] = []string{"ses:CreateReceiptRuleSet",
		"ses:CreateReceiptRule",
		"ses:DeleteReceiptRule",
		"ses:DeleteReceiptRuleSet",
		"ses:DescribeReceiptRuleSet"}

	PushSourceConfigurationActions[SNSPrincipal] = []string{"sns:ConfirmSubscription",
		"sns:GetTopicAttributes",
		"sns:Subscribe",
		"sns:Unsubscribe"}
}

func nodeJSHandlerName(jsBaseFilename string) string {
	return fmt.Sprintf("index.%sConfiguration", jsBaseFilename)
}

func awsPrincipalToService(awsPrincipalName string) string {
	return strings.ToUpper(strings.SplitN(awsPrincipalName, ".", 2)[0])
}

func ensureCustomResourceHandler(customResourceTypeName string,
	sourceArn *gocf.StringExpr,
	dependsOn []string,
	template *gocf.Template,
	S3Bucket string,
	S3Key string,
	logger *logrus.Logger) (string, error) {

	// AWS service basename
	awsServiceName := awsPrincipalToService(customResourceTypeName)

	// Use a stable resource CloudFormation resource name to represent
	// the single CustomResource that can configure the different
	// PushSource's for the given principal.
	keyName, err := json.Marshal(ArbitraryJSONObject{
		"Principal":   customResourceTypeName,
		"ServiceName": awsServiceName,
	})
	if err != nil {
		logger.Error("Failed to create configurator resource name: ", err.Error())
		return "", err
	}
	subscriberHandlerName := CloudFormationResourceName(fmt.Sprintf("%sCustomResource", awsServiceName),
		string(keyName))

	//////////////////////////////////////////////////////////////////////////////
	// IAM Role definition
	iamResourceName, err := ensureIAMRoleForCustomResource(customResourceTypeName, sourceArn, template, logger)
	if nil != err {
		return "", err
	}
	iamRoleRef := gocf.GetAtt(iamResourceName, "Arn")
	_, exists := template.Resources[subscriberHandlerName]
	if !exists {
		logger.WithFields(logrus.Fields{
			"Service": awsServiceName,
		}).Debug("Including Lambda CustomResource for AWS Service")

		configuratorDescription := fmt.Sprintf("Sparta created Lambda CustomResource to configure %s service",
			awsServiceName)

		//////////////////////////////////////////////////////////////////////////////
		// Custom Resource Lambda Handler
		// The export name MUST correspond to the createForwarder entry that is dynamically
		// written into the index.js file during compile in createNewSpartaCustomResourceEntry

		handlerName := fmt.Sprintf("index.%s", javascriptExportNameForResourceType(customResourceTypeName))

		logger.WithFields(logrus.Fields{
			"CustomResourceType": customResourceTypeName,
			"NodeJSExport":       handlerName,
		}).Debug("Sparta CloudFormation custom resource handler info")

		customResourceHandlerDef := gocf.LambdaFunction{
			Code: &gocf.LambdaFunctionCode{
				S3Bucket: gocf.String(S3Bucket),
				S3Key:    gocf.String(S3Key),
			},
			Description: gocf.String(configuratorDescription),
			Handler:     gocf.String(handlerName),
			Role:        iamRoleRef,
			Runtime:     gocf.String(NodeJSVersion),
			Timeout:     gocf.Integer(30),
		}

		cfResource := template.AddResource(subscriberHandlerName, customResourceHandlerDef)
		if nil != dependsOn && (len(dependsOn) > 0) {
			cfResource.DependsOn = append(cfResource.DependsOn, dependsOn...)
		}
	}
	return subscriberHandlerName, nil
}

func ensureConfiguratorLambdaResource(awsPrincipalName string,
	sourceArn *gocf.StringExpr,
	dependsOn []string,
	template *gocf.Template,
	S3Bucket string,
	S3Key string,
	logger *logrus.Logger) (string, error) {

	// AWS service basename
	awsServiceName := awsPrincipalToService(awsPrincipalName)
	configuratorExportName := strings.ToLower(awsServiceName)

	logger.WithFields(logrus.Fields{
		"ServiceName":      awsServiceName,
		"NodeJSExportName": configuratorExportName,
	}).Debug("Ensuring AWS push service configurator CustomResource")

	// Use a stable resource CloudFormation resource name to represent
	// the single CustomResource that can configure the different
	// PushSource's for the given principal.
	keyName, err := json.Marshal(ArbitraryJSONObject{
		"Principal":   awsPrincipalName,
		"ServiceName": awsServiceName,
	})
	if err != nil {
		logger.Error("Failed to create configurator resource name: ", err.Error())
		return "", err
	}
	subscriberHandlerName := CloudFormationResourceName(fmt.Sprintf("%sCustomResource", awsServiceName),
		string(keyName))

	//////////////////////////////////////////////////////////////////////////////
	// IAM Role definition
	iamResourceName, err := ensureIAMRoleForCustomResource(awsPrincipalName, sourceArn, template, logger)
	if nil != err {
		return "", err
	}
	iamRoleRef := gocf.GetAtt(iamResourceName, "Arn")
	_, exists := template.Resources[subscriberHandlerName]
	if !exists {
		logger.WithFields(logrus.Fields{
			"Service": awsServiceName,
		}).Debug("Including Lambda CustomResource for AWS Service")

		configuratorDescription := fmt.Sprintf("Sparta created Lambda CustomResource to configure %s service",
			awsServiceName)

		//////////////////////////////////////////////////////////////////////////////
		// Custom Resource Lambda Handler
		// NOTE: This brittle function name has an analog in ./resources/index.js b/c the
		// AWS Lamba execution treats the entire ZIP file as a module.  So all module exports
		// need to be forwarded through the module's index.js file.
		handlerName := nodeJSHandlerName(configuratorExportName)
		logger.Debug("Lambda Configuration handler: ", handlerName)

		customResourceHandlerDef := gocf.LambdaFunction{
			Code: &gocf.LambdaFunctionCode{
				S3Bucket: gocf.String(S3Bucket),
				S3Key:    gocf.String(S3Key),
			},
			Description: gocf.String(configuratorDescription),
			Handler:     gocf.String(handlerName),
			Role:        iamRoleRef,
			Runtime:     gocf.String("nodejs"),
			Timeout:     gocf.Integer(30),
		}

		cfResource := template.AddResource(subscriberHandlerName, customResourceHandlerDef)
		if nil != dependsOn && (len(dependsOn) > 0) {
			cfResource.DependsOn = append(cfResource.DependsOn, dependsOn...)
		}
	}
	return subscriberHandlerName, nil
}

// ensureIAMRoleForCustomResource ensures that the single IAM::Role for a single
// AWS principal (eg, s3.*.*) exists, and includes statements for the given
// sourceArn.  Sparta uses a single IAM::Role for the CustomResource configuration
// lambda, which is the union of all Arns in the application.
func ensureIAMRoleForCustomResource(awsPrincipalName string,
	sourceArn *gocf.StringExpr,
	template *gocf.Template,
	logger *logrus.Logger) (string, error) {

	principalActions, exists := PushSourceConfigurationActions[awsPrincipalName]
	if !exists {
		return "", fmt.Errorf("Unsupported principal for IAM role creation: %s", awsPrincipalName)
	}

	// What's the stable IAMRoleName?
	resourceBaseName := fmt.Sprintf("CustomResource%sIAMRole", awsPrincipalToService(awsPrincipalName))
	stableRoleName := CloudFormationResourceName(resourceBaseName, awsPrincipalName)

	// Ensure it exists, then check to see if this Source ARN is already specified...
	// Checking equality with Stringable?

	// Create a new Role
	var existingIAMRole *gocf.IAMRole
	existingResource, exists := template.Resources[stableRoleName]
	logger.WithFields(logrus.Fields{
		"PrincipalActions": principalActions,
		"SourceArn":        sourceArn,
	}).Debug("Ensuring IAM Role results")

	if !exists {
		// Insert the IAM role here.  We'll walk the policies data in the next section
		// to make sure that the sourceARN we have is in the list
		statements := CommonIAMStatements["core"]
		existingIAMRole = &gocf.IAMRole{
			AssumeRolePolicyDocument: AssumePolicyDocument,
			Policies: &gocf.IAMPoliciesList{
				gocf.IAMPolicies{
					PolicyDocument: ArbitraryJSONObject{
						"Version":   "2012-10-17",
						"Statement": statements,
					},
					PolicyName: gocf.String(fmt.Sprintf("%sPolicy", stableRoleName)),
				},
			},
		}
		template.AddResource(stableRoleName, existingIAMRole)

		// Create a new IAM Role resource
		logger.WithFields(logrus.Fields{
			"RoleName": stableRoleName,
		}).Debug("Inserting IAM Role")
	} else {
		existingIAMRole = existingResource.Properties.(*gocf.IAMRole)
	}
	// Walk the existing statements
	if nil != existingIAMRole.Policies {
		for _, eachPolicy := range *existingIAMRole.Policies {
			policyDoc := eachPolicy.PolicyDocument.(ArbitraryJSONObject)
			statements := policyDoc["Statement"]
			for _, eachStatement := range statements.([]iamPolicyStatement) {
				if sourceArn.String() == eachStatement.Resource.String() {

					logger.WithFields(logrus.Fields{
						"RoleName":  stableRoleName,
						"SourceArn": sourceArn.String(),
					}).Debug("SourceArn already exists for IAM Policy")
					return stableRoleName, nil
				}
			}
		}

		logger.WithFields(logrus.Fields{
			"RoleName": stableRoleName,
			"Action":   principalActions,
			"Resource": sourceArn,
		}).Debug("Inserting Actions for configuration ARN")

		// Add this statement to the first policy, iff the actions are non-empty
		if len(principalActions) > 0 {
			rootPolicy := (*existingIAMRole.Policies)[0]
			rootPolicyDoc := rootPolicy.PolicyDocument.(ArbitraryJSONObject)
			rootPolicyStatements := rootPolicyDoc["Statement"].([]iamPolicyStatement)
			rootPolicyDoc["Statement"] = append(rootPolicyStatements, iamPolicyStatement{
				Effect:   "Allow",
				Action:   principalActions,
				Resource: sourceArn,
			})
		}

		return stableRoleName, nil
	}

	return "", fmt.Errorf("Unable to find Policies entry for IAM role: %s", stableRoleName)
}
