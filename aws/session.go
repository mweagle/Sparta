package aws

import (
	"context"
	"fmt"

	awsv2Config "github.com/aws/aws-sdk-go-v2/config"

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"

	"github.com/rs/zerolog"

	smithyLogging "github.com/aws/smithy-go/logging"
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
func NewConfigWithConfig(awsConfig awsv2.Config, logger *zerolog.Logger) awsv2.Config {
	return NewConfigWithConfigLevel(awsConfig, 0, logger)
}

// NewConfig that attaches a debug level handler to all AWS requests from services
// sharing the session value.
func NewConfig(logger *zerolog.Logger) awsv2.Config {
	return NewConfigWithLevel(awsv2.ClientLogMode(0), logger)
}

// NewConfigWithLevel returns an AWS Session (https://github.com/aws/aws-sdk-go-v2/blob/main/config/doc)
// object that attaches a debug level handler to all AWS requests from services
// sharing the session value.
func NewConfigWithLevel(level awsv2.ClientLogMode, logger *zerolog.Logger) awsv2.Config {
	awsConfig := awsv2.Config{}
	return NewConfigWithConfigLevel(awsConfig, level, logger)
}

// NewConfigWithConfigLevel returns an AWS Session (https://github.com/aws/aws-sdk-go-v2/blob/main/config/doc)
// object that attaches a debug level handler to all AWS requests from services
// sharing the session value.
func NewConfigWithConfigLevel(awsConfig awsv2.Config,
	level awsv2.ClientLogMode,
	logger *zerolog.Logger) awsv2.Config {
	awsConfig.ClientLogMode = level

	awsConfig, awsConfigErr := awsv2Config.LoadDefaultConfig(context.Background())
	if awsConfigErr != nil {
		panic("WAT")
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

	return awsConfig
	/*
		sess, sessErr := session.NewConfig(awsConfig)
		if sessErr != nil {
			logger.Warn().
				Interface("Error", sessErr).
				Msg("Failed to create AWS Session")
		} else {
			sess.Handlers.Send.PushFront(func(r *request.Request) {
				logger.Debug().
					Str("Service", r.ClientInfo.ServiceName).
					Str("Operation", r.Operation.Name).
					Str("Method", r.Operation.HTTPMethod).
					Str("Path", r.Operation.HTTPPath).
					Interface("Payload", r.Params).
					Msg("AWS Request")
			})
		}
		return sess
	*/
}
