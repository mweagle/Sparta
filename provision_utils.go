package sparta

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/Sirupsen/logrus"
)

const salt = "213EA743-A98F-499D-8FEF-B87015FE13E7"

// PushSourceConfigurationActions map stores common IAM Policy Actions for Lambda
// push-source configuration management.
// The configuration is handled by CustomResources inserted into the generated
// CloudFormation template.
var PushSourceConfigurationActions = map[string][]string{
	"s3.amazonaws.com": {"s3:GetBucketLocation",
		"s3:GetBucketNotification",
		"s3:PutBucketNotification",
		"s3:GetBucketNotificationConfiguration",
		"s3:PutBucketNotificationConfiguration"},
	"sns.amazonaws.com": {"sns:ConfirmSubscription",
		"sns:GetTopicAttributes",
		"sns:Subscribe",
		"sns:Unsubscribe"},
	"apigateway.amazonaws.com": {"apigateway:*",
		"lambda:AddPermission",
		"lambda:RemovePermission",
		"lambda:GetPolicy"},
}

func awsPrincipalToService(awsPrincipalName string) string {
	return strings.ToUpper(strings.SplitN(awsPrincipalName, ".", 2)[0])
}

func ensureConfiguratorLambdaResource(awsPrincipalName string, sourceArn string, resources ArbitraryJSONObject, S3Bucket string, S3Key string, logger *logrus.Logger) (string, error) {
	// AWS service basename
	awsServiceName := awsPrincipalToService(awsPrincipalName)
	configuratorExportName := strings.ToLower(awsServiceName)

	//////////////////////////////////////////////////////////////////////////////
	// IAM Role definition
	// TODO - Check sourceArn for equivalence
	iamResourceName, err := ensureIAMRoleResource(awsPrincipalName, sourceArn, resources, logger)
	if nil != err {
		return "", err
	}

	iamRoleRef := ArbitraryJSONObject{
		"Fn::GetAtt": []string{iamResourceName, "Arn"},
	}
	// Custom handler resource for this service type
	subscriberHandlerName := fmt.Sprintf("%sSubscriber", awsServiceName)
	_, exists := resources[subscriberHandlerName]
	if !exists {
		logger.Info("Creating Subscription Lambda Resource for AWS service: ", awsServiceName)

		//////////////////////////////////////////////////////////////////////////////
		// Custom Resource Lambda Handler
		// NOTE: This brittle function name has an analog in ./resources/index.js b/c the
		// AWS Lamba execution treats the entire ZIP file as a module.  So all module exports
		// need to be forwarded through the module's index.js file.
		handlerName := fmt.Sprintf("index.%sConfiguration", configuratorExportName)
		logger.Debug("Lambda Configuration handler: ", handlerName)

		customResourceHandlerDef := ArbitraryJSONObject{
			"Type": "AWS::Lambda::Function",
			"Properties": ArbitraryJSONObject{
				"Code": ArbitraryJSONObject{
					"S3Bucket": S3Bucket,
					"S3Key":    S3Key,
				},
				"Role":    iamRoleRef,
				"Handler": handlerName,
				"Runtime": "nodejs",
				"Timeout": "30",
			},
		}
		resources[subscriberHandlerName] = customResourceHandlerDef
	}
	return subscriberHandlerName, nil
}

func ensureIAMRoleResource(awsPrincipalName string, sourceArn string, resources ArbitraryJSONObject, logger *logrus.Logger) (string, error) {
	principalActions, exists := PushSourceConfigurationActions[awsPrincipalName]
	if !exists {
		return "", fmt.Errorf("Unsupported principal for IAM role creation: %s", awsPrincipalName)
	}

	hash := sha1.New()
	hash.Write([]byte(fmt.Sprintf("%s%s", awsPrincipalName, salt)))
	roleName := fmt.Sprintf("ConfigIAMRole%s", hex.EncodeToString(hash.Sum(nil)))

	logger.WithFields(logrus.Fields{
		"PrincipalActions": principalActions,
		"Principal":        awsPrincipalName,
	}).Debug("Ensuring IAM Role results")

	existingResource, exists := resources[roleName]

	// If it exists, make sure these permissions are enabled on it...
	if exists {
		statementExists := false
		properties := existingResource.(ArbitraryJSONObject)["Properties"]
		policies := properties.(ArbitraryJSONObject)["Policies"]
		for _, eachPolicy := range policies.([]ArbitraryJSONObject) {
			statements := eachPolicy["PolicyDocument"].(ArbitraryJSONObject)["Statement"]
			for _, eachStatement := range statements.([]ArbitraryJSONObject) {
				if eachStatement["Resource"] == sourceArn {
					statementExists = true
				}
			}
		}
		if !statementExists {
			properties := existingResource.(ArbitraryJSONObject)["Properties"]
			policies := properties.(ArbitraryJSONObject)["Policies"]
			rootPolicy := policies.([]ArbitraryJSONObject)[0]
			policyDocument := rootPolicy["PolicyDocument"].(ArbitraryJSONObject)
			statements := policyDocument["Statements"].([]ArbitraryJSONObject)
			policyDocument["Statements"] = append(statements, ArbitraryJSONObject{
				"Effect":   "Allow",
				"Action":   principalActions,
				"Resource": sourceArn,
			})
		}
		logger.Debug("Using prexisting IAM Role: " + roleName)
		return roleName, nil
	}

	// Create a new IAM Role resource
	logger.WithFields(logrus.Fields{
		"RoleName": roleName,
		"Actions":  principalActions,
	}).Debug("Inserting IAM Role")

	// Provision a new one and add it...
	statements := CommonIAMStatements["core"]
	statements = append(statements, ArbitraryJSONObject{
		"Effect":   "Allow",
		"Action":   principalActions,
		"Resource": sourceArn,
	})

	iamPolicy := ArbitraryJSONObject{"Type": "AWS::IAM::Role",
		"Properties": ArbitraryJSONObject{
			"AssumeRolePolicyDocument": AssumePolicyDocument,
			"Policies": []ArbitraryJSONObject{
				{
					"PolicyName": fmt.Sprintf("Configurator%s", CloudFormationResourceName(awsPrincipalName)),
					"PolicyDocument": ArbitraryJSONObject{
						"Version":   "2012-10-17",
						"Statement": statements,
					},
				},
			},
		},
	}
	resources[roleName] = iamPolicy
	return roleName, nil
}
