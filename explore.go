// +build !lambdabinary

package sparta

import (
	"fmt"
	"strconv"

	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
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

// Explore supports interactive command line invocation of the previously
// provisioned Sparta service
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
