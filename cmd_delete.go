//go:build !lambdabinary
// +build !lambdabinary

package sparta

import (
	"context"

	spartaAWS "github.com/mweagle/Sparta/aws"
	spartaCF "github.com/mweagle/Sparta/aws/cloudformation"
	"github.com/rs/zerolog"

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	awsv2CF "github.com/aws/aws-sdk-go-v2/service/cloudformation"
)

// Delete ensures that the provided serviceName is deleted.
// Failing to delete a non-existent service is considered a success.
func Delete(ctx context.Context, serviceName string, logger *zerolog.Logger) error {
	awsConfig := spartaAWS.NewConfig(logger)
	awsCloudFormation := awsv2CF.NewFromConfig(awsConfig)

	exists, err := spartaCF.StackExists(ctx, serviceName, awsConfig, logger)
	if nil != err {
		return err
	}
	logger.Info().
		Bool("Exists", exists).
		Str("Name", serviceName).
		Msg("Stack existence check")

	if exists {

		params := &awsv2CF.DeleteStackInput{
			StackName: awsv2.String(serviceName),
		}
		resp, err := awsCloudFormation.DeleteStack(ctx, params)
		if nil != resp {
			logger.Info().
				Interface("Response", resp).
				Msg("Delete request submitted")
		}
		return err
	}
	logger.Info().Msg("Stack does not exist")
	return nil
}
