// +build !lambdabinary

package sparta

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/sts"
	spartaAWS "github.com/mweagle/Sparta/aws"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Status produces a status report for the given stack
func Status(serviceName string,
	serviceDescription string,
	redact bool,
	logger *logrus.Logger) error {

	awsSession := spartaAWS.NewSession(logger)
	cfSvc := cloudformation.New(awsSession)

	params := &cloudformation.DescribeStacksInput{
		StackName: aws.String(serviceName),
	}
	describeStacksResponse, describeStacksResponseErr := cfSvc.DescribeStacks(params)

	if describeStacksResponseErr != nil {
		if strings.Contains(describeStacksResponseErr.Error(), "does not exist") {
			logger.WithField("Region", *awsSession.Config.Region).Info("Stack does not exist")
			return nil
		}
		return describeStacksResponseErr
	}
	if len(describeStacksResponse.Stacks) > 1 {
		return errors.Errorf("More than 1 stack returned for %s. Count: %d",
			serviceName,
			len(describeStacksResponse.Stacks))
	}

	// What's the current accountID?
	redactor := func(stringValue string) string {
		return stringValue
	}
	if redact {
		input := &sts.GetCallerIdentityInput{}
		stsSvc := sts.New(awsSession)
		identityResponse, identityResponseErr := stsSvc.GetCallerIdentity(input)
		if identityResponseErr != nil {
			return identityResponseErr
		}
		redactedValue := strings.Repeat("*", len(*identityResponse.Account))
		redactor = func(stringValue string) string {
			return strings.Replace(stringValue,
				*identityResponse.Account,
				redactedValue,
				-1)
		}
	}

	// Report on what's up with the stack...
	stackInfo := describeStacksResponse.Stacks[0]
	logger.WithField("Id", redactor(*stackInfo.StackId)).Info("StackId")
	logger.WithField("State", *stackInfo.StackStatus).Info("Stack status")
	logger.WithField("Time", stackInfo.CreationTime.UTC().String()).Info("Created")
	if stackInfo.LastUpdatedTime != nil {
		logger.WithField("Time", stackInfo.LastUpdatedTime.UTC().String()).Info("Last Update")
	}
	if stackInfo.DeletionTime != nil {
		logger.WithField("Time", stackInfo.DeletionTime.UTC().String()).Info("Deleted")
	}
	for _, eachParam := range stackInfo.Parameters {
		logger.WithField(*eachParam.ParameterKey, redactor(*eachParam.ParameterValue)).Info("Parameter")
	}
	for _, eachTag := range stackInfo.Tags {
		logger.WithField(*eachTag.Key, redactor(*eachTag.Value)).Info("Tag")
	}
	for _, eachOutput := range stackInfo.Outputs {
		statement := logger.WithField(*eachOutput.OutputKey, redactor(*eachOutput.OutputValue))
		if eachOutput.ExportName != nil {
			statement.WithField("Output", *eachOutput.ExportName)
		}
		statement.Info("Output")
	}
	return nil
}
