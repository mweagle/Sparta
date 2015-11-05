// Copyright (c) 2015 Matt Weagle <mweagle@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package sparta

import (
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	"reflect"
	"strings"
)

var PushSourceConfigurationPermissions = map[string][]string{
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
	principalActions, exists := PushSourceConfigurationPermissions[awsPrincipalName]
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
		newIAMRoleResourceName := cloudFormationResourceName("IAMRole")
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
