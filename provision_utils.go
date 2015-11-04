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
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
}

var commonIAMStatements = []ArbitraryJSONObject{
	{
		"Action":   []string{"logs:CreateLogGroup", "logs:CreateLogStream", "logs:PutLogEvents"},
		"Effect":   "Allow",
		"Resource": "arn:aws:logs:*:*:*",
	},
}

var PushSourceConfigurationPermissions = map[string][]string{
	"s3.amazonaws.com": {"s3:GetBucketNotificationConfiguration",
		"s3:PutBucketNotificationConfiguration"},
	"sns.amazonaws.com": {"sns:ConfirmSubscription",
		"sns:GetTopicAttributes",
		"sns:Subscribe",
		"sns:Unsubscribe"},
}

func cloudFormationResourceName(prefix string) string {
	randValue := rand.Int63()
	hash := sha1.New()
	hash.Write([]byte(prefix))
	hash.Write([]byte(strconv.FormatInt(randValue, 10)))
	return fmt.Sprintf("%s%s", prefix, hex.EncodeToString(hash.Sum(nil)))
}

func ensureConfiguratorLambdaResource(awsPrincipalName string, sourceArn string, resources ArbitraryJSONObject, logger *logrus.Logger) (string, error) {
	// AWS service basename
	awsServiceName := strings.ToUpper(strings.SplitN(awsPrincipalName, ".", 2)[0])

	//////////////////////////////////////////////////////////////////////////////
	// IAM Role definition
	iamResourceName := fmt.Sprintf("Custom%sIAMRole", awsServiceName)
	_, exists := resources[iamResourceName]
	if !exists {
		logger.Info("Creating IAM Role Resource for AWS service: ", awsServiceName)

		iamRoleDef, err := createIAMRoleResource(awsPrincipalName, sourceArn, logger)
		if nil != err {
			return "", err
		}
		resources[iamResourceName] = iamRoleDef
	}
	iamRoleRef := ArbitraryJSONObject{
		"Fn::GetAtt": []string{iamResourceName, "Arn"},
	}
	// Custom handler resource for this service type
	lambdaHandlerName := fmt.Sprintf("%sNotificationHandler", awsServiceName)
	_, exists = resources[lambdaHandlerName]
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
		resources[lambdaHandlerName] = customResourceHandlerDef
	}
	return lambdaHandlerName, nil
}

func createIAMRoleResource(awsPrincipalName string, sourceArn string, logger *logrus.Logger) (interface{}, error) {

	statements := commonIAMStatements

	principalActions, exists := PushSourceConfigurationPermissions[awsPrincipalName]
	if !exists {
		return nil, errors.New("Unsupported principal for IAM role creation: " + awsPrincipalName)
	}
	logger.Info("IAMRole Actions: ", principalActions)

	statements = append(statements, ArbitraryJSONObject{
		"Effect":   "Allow",
		"Action":   principalActions,
		"Resource": sourceArn,
	})

	iamPolicy := ArbitraryJSONObject{"Type": "AWS::IAM::Role",
		"Properties": ArbitraryJSONObject{
			"AssumeRolePolicyDocument": ArbitraryJSONObject{
				"Version": "2012-10-17",
				"Statement": []ArbitraryJSONObject{
					{
						"Effect": "Allow",
						"Principal": ArbitraryJSONObject{
							"Service": []string{"lambda.amazonaws.com"},
						},
						"Action": []string{"sts:AssumeRole"},
					},
					{
						"Effect": "Allow",
						"Principal": ArbitraryJSONObject{
							"Service": []string{"ec2.amazonaws.com"},
						},
						"Action": []string{"sts:AssumeRole"},
					},
				},
			},
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
	return iamPolicy, nil
}
