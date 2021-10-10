//go:build !lambdabinary
// +build !lambdabinary

package sparta

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsv2CF "github.com/aws/aws-sdk-go-v2/service/cloudformation"
	awsv2STS "github.com/aws/aws-sdk-go-v2/service/sts"
	spartaAWS "github.com/mweagle/Sparta/aws"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

func logSectionHeader(text string,
	dividerWidth int,
	logger *zerolog.Logger) {
	// Add a nice divider if there are Stack specific output
	outputHeader := fmt.Sprintf("%s ", text)
	suffix := strings.Repeat("▬", dividerWidth-len(outputHeader))
	logger.Info().Msgf("%s%s", outputHeader, suffix)
}

// Status produces a status report for the given stack
func Status(ctx context.Context,
	serviceName string,
	serviceDescription string,
	redact bool,
	logger *zerolog.Logger) error {

	awsConfig, awsConfigErr := spartaAWS.NewConfig(ctx, logger)
	if awsConfigErr != nil {
		return awsConfigErr
	}
	cfSvc := awsv2CF.NewFromConfig(awsConfig)

	params := &awsv2CF.DescribeStacksInput{
		StackName: aws.String(serviceName),
	}
	describeStacksResponse, describeStacksResponseErr := cfSvc.DescribeStacks(ctx, params)

	if describeStacksResponseErr != nil {
		if strings.Contains(describeStacksResponseErr.Error(), "does not exist") {
			logger.Info().Str("Region", awsConfig.Region).Msg("Stack does not exist")
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
		input := &awsv2STS.GetCallerIdentityInput{}

		stsSvc := awsv2STS.NewFromConfig(awsConfig)
		identityResponse, identityResponseErr := stsSvc.GetCallerIdentity(ctx, input)
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
	logSectionHeader("Stack Summary", dividerLength, logger)
	stackInfo := describeStacksResponse.Stacks[0]
	logger.Info().Str("Id", redactor(*stackInfo.StackId)).Msg("StackId")
	logger.Info().Str("Description", redactor(*stackInfo.Description)).Msg("Description")
	logger.Info().Str("State", string(stackInfo.StackStatus)).Msg("Status")
	if stackInfo.StackStatusReason != nil {
		logger.Info().Str("Reason", *stackInfo.StackStatusReason).Msg("Reason")
	}
	logger.Info().Str("Time", stackInfo.CreationTime.UTC().String()).Msg("Created")
	if stackInfo.LastUpdatedTime != nil {
		logger.Info().Str("Time", stackInfo.LastUpdatedTime.UTC().String()).Msg("Last Update")
	}
	if stackInfo.DeletionTime != nil {
		logger.Info().Str("Time", stackInfo.DeletionTime.UTC().String()).Msg("Deleted")
	}

	logger.Info()
	if len(stackInfo.Parameters) != 0 {
		logSectionHeader("Parameters", dividerLength, logger)
		for _, eachParam := range stackInfo.Parameters {
			logger.Info().Str("Value",
				redactor(*eachParam.ParameterValue)).Msg(*eachParam.ParameterKey)
		}
		logger.Info().Msg("")
	}
	if len(stackInfo.Tags) != 0 {
		logSectionHeader("Tags", dividerLength, logger)
		for _, eachTag := range stackInfo.Tags {
			logger.Info().Str("Value",
				redactor(*eachTag.Value)).Msg(*eachTag.Key)
		}
		logger.Info().Msg("")
	}
	if len(stackInfo.Outputs) != 0 {
		logSectionHeader("Outputs", dividerLength, logger)
		for _, eachOutput := range stackInfo.Outputs {
			statement := logger.Info().Str("Value",
				redactor(*eachOutput.OutputValue))
			if eachOutput.ExportName != nil {
				statement.Str("ExportName", *eachOutput.ExportName)
			}
			statement.Msg(*eachOutput.OutputKey)
		}
		logger.Info()
	}
	return nil
}
