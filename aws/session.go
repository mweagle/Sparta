package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/rs/zerolog"
)

type zerologProxy struct {
	logger *zerolog.Logger
}

// Log is a utility function to comply with the AWS signature
func (proxy *zerologProxy) Log(args ...interface{}) {
	proxy.logger.Info().Msg(fmt.Sprintf("%v", args))
}

// NewSessionWithConfig returns an awsSession that includes the user supplied
// configuration information
func NewSessionWithConfig(awsConfig *aws.Config, logger *zerolog.Logger) *session.Session {
	return NewSessionWithConfigLevel(awsConfig, aws.LogDebugWithRequestErrors, logger)
}

// NewSession that attaches a debug level handler to all AWS requests from services
// sharing the session value.
func NewSession(logger *zerolog.Logger) *session.Session {
	return NewSessionWithLevel(aws.LogDebugWithRequestErrors, logger)
}

// NewSessionWithLevel returns an AWS Session (https://github.com/aws/aws-sdk-go/wiki/Getting-Started-Configuration)
// object that attaches a debug level handler to all AWS requests from services
// sharing the session value.
func NewSessionWithLevel(level aws.LogLevelType, logger *zerolog.Logger) *session.Session {
	awsConfig := &aws.Config{
		CredentialsChainVerboseErrors: aws.Bool(true),
	}
	return NewSessionWithConfigLevel(awsConfig, level, logger)
}

// NewSessionWithConfigLevel returns an AWS Session (https://github.com/aws/aws-sdk-go/wiki/Getting-Started-Configuration)
// object that attaches a debug level handler to all AWS requests from services
// sharing the session value.
func NewSessionWithConfigLevel(awsConfig *aws.Config,
	level aws.LogLevelType,
	logger *zerolog.Logger) *session.Session {
	if nil == awsConfig {
		awsConfig = &aws.Config{
			CredentialsChainVerboseErrors: aws.Bool(true),
		}
	}

	// Log AWS calls if needed
	switch logger.GetLevel() {
	case zerolog.DebugLevel:
		awsConfig.LogLevel = aws.LogLevel(level)
	}
	awsConfig.Logger = &zerologProxy{logger}
	sess, sessErr := session.NewSession(awsConfig)
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
	logger.Debug().
		Str("Name", aws.SDKName).
		Str("Version", aws.SDKVersion).
		Msg("AWS SDK Info")
	return sess
}
