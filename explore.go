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

// +build !lambdabinary

package sparta

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"strconv"
)

type provisionedResources []*cloudformation.StackResourceSummary

func stackLambdaResources(serviceName string, cf *cloudformation.CloudFormation, logger *logrus.Logger) (provisionedResources, error) {

	resources := make(provisionedResources, 0)
	nextToken := ""
	for {
		params := &cloudformation.ListStackResourcesInput{
			StackName: aws.String(serviceName),
		}
		if "" != nextToken {
			params.NextToken = aws.String(nextToken)
		}
		resp, err := cf.ListStackResources(params)

		if err != nil {
			logger.Error(err.Error())
			return nil, err
		}
		for _, eachSummary := range resp.StackResourceSummaries {
			if *eachSummary.ResourceType == "AWS::Lambda::Function" {
				resources = append(resources, eachSummary)
			}
		}
		if nil != resp.NextToken {
			nextToken = *resp.NextToken
		} else {
			break
		}
	}
	return resources, nil
}

func promptForSelection(lambdaFunctions provisionedResources) *cloudformation.StackResourceSummary {
	fmt.Printf("Please choose the lambda function to test:\n")
	for index, eachSummary := range lambdaFunctions {
		fmt.Printf("  (%d) %s\n", index+1, *eachSummary.PhysicalResourceId)
	}
	fmt.Printf("Selection: ")
	var selection string
	fmt.Scanln(&selection)
	selectedIndex, err := strconv.Atoi(selection)
	if nil != err {
		return nil
	} else if selectedIndex > 0 && selectedIndex <= len(lambdaFunctions) {
		return lambdaFunctions[selectedIndex-1]
	} else {
		return nil
	}
}

// Interactively invoke the previously provisioned Lambda functions
func Explore(serviceName string, logger *logrus.Logger) error {
	session := awsSession(logger)
	awsCloudFormation := cloudformation.New(session)

	exists, err := stackExists(serviceName, awsCloudFormation, logger)
	if nil != err {
		return err
	} else if !exists {
		logger.Info("Stack does not exist: ", serviceName)
		return nil
	} else {
		resources, err := stackLambdaResources(serviceName, awsCloudFormation, logger)
		if nil != err {
			return nil
		}
		selected := promptForSelection(resources)
		if nil != selected {
			logger.Info("TODO: Invoke", selected)
		}
	}
	return nil
}
