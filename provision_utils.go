package sparta

import (
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	"reflect"
	"strings"
)

// Common IAM Policy Actions for Lambda push-source configuration management.
// The configuration is handled by CustomResources inserted into the generated
// CloudFormation template.
var PushSourceConfigurationActions = map[string][]string{
	"s3.amazonaws.com": {"s3:GetBucketNotificationConfiguration",
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

func ensureIAMRoleResource(awsPrincipalName string, sourceArn string, resources ArbitraryJSONObject, logger *logrus.Logger) (string, error) {
	principalActions, exists := PushSourceConfigurationActions[awsPrincipalName]
	if !exists {
		return "", errors.New("Unsupported principal for IAM role creation: " + awsPrincipalName)
	}

	// First determine if there is one provisioned...
	var iamRoleResourceNames []string
	for eachName, eachResource := range resources {
		logger.Debug("Checking IAM Policy equality: ", eachName)
		if eachResource.(ArbitraryJSONObject)["Type"] == "AWS::IAM::Role" {
			properties := eachResource.(ArbitraryJSONObject)["Properties"]
			policies := properties.(ArbitraryJSONObject)["Policies"]
			for _, eachPolicyEntry := range policies.([]ArbitraryJSONObject) {
				policyDocument := eachPolicyEntry["PolicyDocument"]
				statements := policyDocument.(ArbitraryJSONObject)["Statement"]
				for _, eachStatement := range statements.([]ArbitraryJSONObject) {
					if eachStatement["Resource"] == sourceArn &&
						reflect.DeepEqual(eachStatement["Action"], principalActions) {
						iamRoleResourceNames = append(iamRoleResourceNames, eachName)
					}
				}
			}
		}
	}
	logger.WithFields(logrus.Fields{
		"MatchingIAMRoleNames": iamRoleResourceNames,
		"PrincipalActions":     principalActions,
		"Principal":            awsPrincipalName,
	}).Debug("Ensuring IAM Role results")

	if len(iamRoleResourceNames) > 1 {
		return "", errors.New("More than 1 IAM Role found for entry: " + awsPrincipalName)
	} else if len(iamRoleResourceNames) == 1 {
		logger.Debug("Using prexisting IAM Role: " + iamRoleResourceNames[0])
		return iamRoleResourceNames[0], nil
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
						"PolicyName": "configurator",
						"PolicyDocument": ArbitraryJSONObject{
							"Version":   "2012-10-17",
							"Statement": statements,
						},
					},
				},
			},
		}
		resources[newIAMRoleResourceName] = iamPolicy
		return newIAMRoleResourceName, nil
	}
}
