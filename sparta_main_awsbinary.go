// +build lambdabinary

package sparta

// Provides NOP implementations for functions that do not need to execute
// in the Lambda context

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"runtime"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var reExtractPlatInfo = regexp.MustCompile(`(\w+)=\"(.*)\"`)

// Main defines the primary handler for transforming an application into a Sparta package.  The
// serviceName is used to uniquely identify your service within a region and will
// be used for subsequent updates.  For provisioning, ensure that you've
// properly configured AWS credentials for the golang SDK.
// See http://docs.aws.amazon.com/sdk-for-go/api/aws/defaults.html#DefaultChainCredentials-constant
// for more information.
func Main(serviceName string,
	serviceDescription string,
	lambdaAWSInfos []*LambdaAWSInfo,
	api *API,
	site *S3Site) error {
	return MainEx(serviceName,
		serviceDescription,
		lambdaAWSInfos,
		api,
		site,
		nil,
		false)
}

// MainEx provides an "extended" Main that supports customizing the standard Sparta
// workflow via the `workflowHooks` parameter.
func MainEx(serviceName string,
	serviceDescription string,
	lambdaAWSInfos []*LambdaAWSInfo,
	api APIGateway,
	site *S3Site,
	workflowHooks *WorkflowHooks,
	useCGO bool) error {

	// It's possible the user attached a custom command to the
	// root command. If there is no command, then just run the
	// Execute command...
	CommandLineOptions.Root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// This can only run in AWS Lambda
		mainLogLevel := "info"
		envVarLogLevel := os.Getenv(envVarLogLevel)
		if envVarLogLevel != "" {
			mainLogLevel = envVarLogLevel
		}
		// We never want colors in AWS because the console can't show them...
		logger, loggerErr := NewLoggerForOutput(mainLogLevel, "json", true)
		if loggerErr != nil {
			return loggerErr
		}
		if logger == nil {
			return errors.Errorf("Failed to initialize logger instance")
		}

		welcomeMessage := fmt.Sprintf("Service: %s", StampedServiceName)
		logger.Info().
			Str(fmt.Sprintf("%sVersion", ProperName), SpartaVersion).
			Str(fmt.Sprintf("%sSHA", ProperName), SpartaGitHash[0:7]).
			Str("go Version", runtime.Version()).
			Str("BuildID", StampedBuildID).
			Str("UTC", time.Now().UTC().Format(time.RFC3339)).
			Msg(welcomeMessage)
		OptionsGlobal.ServiceName = StampedServiceName
		OptionsGlobal.Logger = logger
		return nil
	}
	CommandLineOptions.Root.RunE = func(cmd *cobra.Command, args []string) (err error) {
		defer func() {
			if r := recover(); r != nil {
				OptionsGlobal.Logger.Error().Msgf("Panic recovered: %v", r)
				err = errors.Errorf(fmt.Sprintf("%v", r))
			}
		}()

		// By default run the Execute command
		err = Execute(StampedServiceName,
			lambdaAWSInfos,
			OptionsGlobal.Logger)
		return err
	}
	return CommandLineOptions.Root.Execute()
}

// Delete is not available in the AWS Lambda binary
func Delete(serviceName string, logger *zerolog.Logger) error {
	logger.Error().Msg("Delete() not supported in AWS Lambda binary")
	return errors.New("Delete not supported for this binary")
}

// Build is not available in the AWS Lambda binary
func Build(noop bool,
	serviceName string,
	serviceDescription string,
	lambdaAWSInfos []*LambdaAWSInfo,
	api APIGateway,
	site *S3Site,
	useCGO bool,
	buildID string,
	outputDirectory string,
	buildTags string,
	linkerFlags string,
	templateWriter io.Writer,
	workflowHooks *WorkflowHooks,
	logger *zerolog.Logger) error {
	logger.Error().Msg("Build() not supported in AWS Lambda binary")
	return errors.New("Build not supported for this binary")
}

// Provision is not available in the AWS Lambda binary
func Provision(noop bool,
	serviceName string,
	serviceDescription string,
	lambdaAWSInfos []*LambdaAWSInfo,
	api *API,
	site *S3Site,
	s3Bucket string,
	useCGO bool,
	inplace bool,
	buildID string,
	codePipelineTrigger string,
	buildTags string,
	linkerFlags string,
	writer io.Writer,
	workflowHooks *WorkflowHooks,
	logger *zerolog.Logger) error {
	logger.Error().Msg("Provision() not supported in AWS Lambda binary")
	return errors.New("Provision not supported for this binary")
}

// Describe is not available in the AWS Lambda binary
func Describe(serviceName string,
	serviceDescription string,
	lambdaAWSInfos []*LambdaAWSInfo,
	api *API,
	site *S3Site,
	s3BucketName string,
	buildTags string,
	linkerFlags string,
	outputWriter io.Writer,
	workflowHooks *WorkflowHooks,
	logger *zerolog.Logger) error {
	logger.Error().Msg("Describe() not supported in AWS Lambda binary")
	return errors.New("Describe not supported for this binary")
}

// Explore is an interactive command that brings up a GUI to test
// lambda functions previously deployed into AWS lambda. It's not supported in the
// AWS binary build
func Explore(serviceName string,
	serviceDescription string,
	lambdaAWSInfos []*LambdaAWSInfo,
	api *API,
	site *S3Site,
	s3BucketName string,
	buildTags string,
	linkerFlags string,
	logger *zerolog.Logger) error {
	return errors.New("Explore not supported for this binary")
}

// Profile is the interactive command used to pull S3 assets locally into /tmp
// and run ppro against the cached profiles
func Profile(serviceName string,
	serviceDescription string,
	s3Bucket string,
	httpPort int,
	logger *zerolog.Logger) error {
	return errors.New("Profile not supported for this binary")
}

// Status is the command that produces a simple status report for a given
// stack
func Status(serviceName string,
	serviceDescription string,
	redact bool,
	logger *zerolog.Logger) error {
	return errors.New("Status not supported for this binary")
}

func platformLogSysInfo(lambdaFunc string, logger *zerolog.Logger) {

	// Setup the files and their respective log levels
	mapFilesToLoggerCall := map[zerolog.Level][]string{
		zerolog.InfoLevel: {
			"/proc/version",
			"/etc/os-release",
		},
		zerolog.DebugLevel: {
			"/proc/cpuinfo",
			"/proc/meminfo",
		},
	}

	for eachLevel, eachList := range mapFilesToLoggerCall {
		for _, eachEntry := range eachList {
			data, dataErr := ioutil.ReadFile(eachEntry)
			if dataErr == nil && len(data) != 0 {

				entry := logger.WithLevel(eachLevel).Str("filepath", eachEntry)
				match := reExtractPlatInfo.FindAllStringSubmatch(string(data), -1)
				if match != nil {
					for _, eachMatch := range match {
						entry = entry.Str(eachMatch[1], eachMatch[2])
					}
				} else {
					entry = entry.Str("contents", string(data))
				}
				entry.Msg("Host Info")
			} else if dataErr != nil || len(data) <= 0 {
				logger.Warn().
					Str("filepath", eachEntry).
					Interface("error", dataErr).
					Int("length", len(data)).
					Msg("Failed to read host info")
			}
		}
	}
}

// RegisterCodePipelineEnvironment is not available during lambda execution
func RegisterCodePipelineEnvironment(environmentName string, environmentVariables map[string]string) error {
	return nil
}
