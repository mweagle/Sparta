package aws

import (
	"context"
	"fmt"

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	awsv2Config "github.com/aws/aws-sdk-go-v2/config"
	smithyLogging "github.com/aws/smithy-go/logging"
	"github.com/rs/zerolog"
)

type zerologProxy struct {
	logger *zerolog.Logger
}

// Log is a utility function to comply with the AWS signature
func (proxy *zerologProxy) Logf(classification smithyLogging.Classification,
	format string,
	args ...interface{}) {
	proxy.logger.Debug().Msg(fmt.Sprintf(format, args...))
}

// NewConfigWithConfig returns an awsSession that includes the user supplied
// configuration information
func NewConfigWithConfig(ctx context.Context,
	awsConfig awsv2.Config,
	logger *zerolog.Logger) (awsv2.Config, error) {
	return NewConfigWithConfigLevel(ctx, awsConfig, 0, logger)
}

// NewConfig that attaches a debug level handler to all AWS requests from services
// sharing the session value.
func NewConfig(ctx context.Context, logger *zerolog.Logger) (awsv2.Config, error) {
	return NewConfigWithLevel(ctx, awsv2.ClientLogMode(0), logger)
}

// NewConfigWithLevel returns an AWS Session (https://github.com/aws/aws-sdk-go-v2/blob/main/config/doc)
// object that attaches a debug level handler to all AWS requests from services
// sharing the session value.
func NewConfigWithLevel(ctx context.Context,
	level awsv2.ClientLogMode,
	logger *zerolog.Logger) (awsv2.Config, error) {
	awsConfig := awsv2.Config{}
	return NewConfigWithConfigLevel(ctx, awsConfig, level, logger)
}

// NewConfigWithConfigLevel returns an AWS Session (https://github.com/aws/aws-sdk-go-v2/blob/main/config/doc)
// object that attaches a debug level handler to all AWS requests from services
// sharing the session value.
func NewConfigWithConfigLevel(ctx context.Context,
	awsConfig awsv2.Config,
	level awsv2.ClientLogMode,
	logger *zerolog.Logger) (awsv2.Config, error) {
	awsConfig.ClientLogMode = level

	awsConfig, awsConfigErr := awsv2Config.LoadDefaultConfig(ctx)
	if awsConfigErr != nil {
		return awsv2.Config{}, awsConfigErr
	}
	// Log AWS calls if needed
	switch logger.GetLevel() {
	case zerolog.DebugLevel:
		awsConfig.ClientLogMode = awsv2.LogRequest | awsv2.LogResponse | awsv2.LogRetries
	}
	awsConfig.Logger = &zerologProxy{logger}

	logger.Debug().
		Str("Name", awsv2.SDKName).
		Str("Version", awsv2.SDKVersion).
		Msg("AWS SDK Info.")

	return awsConfig, nil
}
