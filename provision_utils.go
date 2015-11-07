package sparta

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	"strings"
)

const salt = "213EA743-A98F-499D-8FEF-B87015FE13E7"

// Common IAM Policy Actions for Lambda push-source configuration management.
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
}

func ensureConfiguratorLambdaResource(awsPrincipalName string, sourceArn string, resources ArbitraryJSONObject, logger *logrus.Logger) (string, error) {
	// AWS service basename
	awsServiceName := strings.ToUpper(strings.SplitN(awsPrincipalName, ".", 2)[0])

	//////////////////////////////////////////////////////////////////////////////
	// IAM Role definition
	// TODO - Check sourceArn for equivalence
	iamResourceName, err := ensureIAMRoleResource(awsServiceName, awsPrincipalName, sourceArn, resources, logger)
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
		// NOTE: This path depends on `go generate` already having processed the provision
		// directory with the https://github.com/tdewolff/minify/tree/master/cmd/minify contents
		scriptHandlerPath := fmt.Sprintf("/resources/provision/%s.min.js", strings.ToLower(awsServiceName))
		logger.Debug("Lambda Source: ", scriptHandlerPath)

		customResourceHandlerDef := ArbitraryJSONObject{
			"Type": "AWS::Lambda::Function",
			"Properties": ArbitraryJSONObject{
				"Code": ArbitraryJSONObject{
					"ZipFile": FSMustString(false, scriptHandlerPath),
				},
				"Role":    iamRoleRef,
				"Handler": "index.handler",
				"Runtime": "nodejs",
				"Timeout": "30",
			},
		}
		resources[subscriberHandlerName] = customResourceHandlerDef
	}
	return subscriberHandlerName, nil
}

func ensureIAMRoleResource(awsServiceName string, awsPrincipalName string, sourceArn string, resources ArbitraryJSONObject, logger *logrus.Logger) (string, error) {
	principalActions, exists := PushSourceConfigurationActions[awsPrincipalName]
	if !exists {
		return "", errors.New("Unsupported principal for IAM role creation: " + awsPrincipalName)
	}

	hash := sha1.New()
	hash.Write([]byte(fmt.Sprintf("%s%s", awsPrincipalName, salt)))
	roleName := fmt.Sprintf("%sConfigIAMRole%s", awsServiceName, hex.EncodeToString(hash.Sum(nil)))

	logger.WithFields(logrus.Fields{
		"PrincipalActions": principalActions,
		"Principal":        awsPrincipalName,
	}).Debug("Ensuring IAM Role results")

	_, exists = resources[roleName]

	// If it exists, make sure these permissions are enabled on it...
	if exists {
		logger.Debug("Using prexisting IAM Role: " + roleName)
		return roleName, nil
	} else {
		// Provision a new one and add it...
		newIAMRoleResourceName := CloudFormationResourceName("IAMRole")
		logger.Debug("Inserting new IAM Role: ", newIAMRoleResourceName)

		statements := CommonIAMStatements
		logger.Info("IAMRole Actions: ", principalActions)

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
						"PolicyName": fmt.Sprintf("%sConfigurator%s", awsServiceName, CloudFormationResourceName("")),
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
}
