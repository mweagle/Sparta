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
	"github.com/sirupsen/logrus"
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
	api *API,
	site *S3Site,
	workflowHooks *WorkflowHooks,
	useCGO bool) error {

	// It's possible the user attached a custom command to the
	// root command. If there is no command, then just run the
	// Execute command...
	CommandLineOptions.Root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// This can only run in AWS Lambda
		formatter := &logrus.JSONFormatter{}
		mainLogLevel := "info"
		envVarLogLevel := os.Getenv(envVarLogLevel)
		if envVarLogLevel != "" {
			mainLogLevel = envVarLogLevel
		}
		logger, loggerErr := NewLoggerWithFormatter(mainLogLevel, formatter)
		if loggerErr != nil {
			return loggerErr
		}
		if logger == nil {
			return errors.Errorf("Failed to initialize logger instance")
		}

		welcomeMessage := fmt.Sprintf("Service: %s", StampedServiceName)
		logger.WithFields(logrus.Fields{
			fmt.Sprintf("%sVersion", ProperName): SpartaVersion,
			fmt.Sprintf("%sSHA", ProperName):     SpartaGitHash[0:7],
			"go Version":                         runtime.Version(),
			"BuildID":                            StampedBuildID,
			"UTC":                                (time.Now().UTC().Format(time.RFC3339)),
		}).Info(welcomeMessage)
		OptionsGlobal.ServiceName = StampedServiceName
		OptionsGlobal.Logger = logger
		return nil
	}
	CommandLineOptions.Root.RunE = func(cmd *cobra.Command, args []string) error {
		// By default run the Execute command
		return Execute(StampedServiceName,
			lambdaAWSInfos,
			OptionsGlobal.Logger)
	}
	return CommandLineOptions.Root.Execute()
}

// Delete is not available in the AWS Lambda binary
func Delete(serviceName string, logger *logrus.Logger) error {
	logger.Error("Delete() not supported in AWS Lambda binary")
	return errors.New("Delete not supported for this binary")
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
	logger *logrus.Logger) error {
	logger.Error("Provision() not supported in AWS Lambda binary")
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
	logger *logrus.Logger) error {
	logger.Error("Describe() not supported in AWS Lambda binary")
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
	logger *logrus.Logger) error {
	return errors.New("Explore not supported for this binary")
}

// Profile is the interactive command used to pull S3 assets locally into /tmp
// and run ppro against the cached profiles
func Profile(serviceName string,
	serviceDescription string,
	s3Bucket string,
	httpPort int,
	logger *logrus.Logger) error {
	return errors.New("Profile not supported for this binary")
}

// Status is the command that produces a simple status report for a given
// stack
func Status(serviceName string,
	serviceDescription string,
	redact bool,
	logger *logrus.Logger) error {
	return errors.New("Status not supported for this binary")
}

func platformLogSysInfo(lambdaFunc string, logger *logrus.Logger) {

	// Setup the files and their respective log levels
	mapFilesToLoggerCall := map[logrus.Level][]string{
		logrus.InfoLevel: []string{
			"/proc/version",
			"/etc/os-release",
		},
		logrus.DebugLevel: []string{
			"/proc/cpuinfo",
			"/proc/meminfo",
		},
	}

	for eachLevel, eachList := range mapFilesToLoggerCall {
		for _, eachEntry := range eachList {
			data, dataErr := ioutil.ReadFile(eachEntry)
			if dataErr == nil && len(data) != 0 {
				loggerFields := make(logrus.Fields, 0)
				loggerFields["filepath"] = eachEntry
				match := reExtractPlatInfo.FindAllStringSubmatch(string(data), -1)
				if match != nil {
					for _, eachMatch := range match {
						loggerFields[eachMatch[1]] = eachMatch[2]
					}
				} else {
					loggerFields["contents"] = string(data)
				}
				logger.WithFields(loggerFields).
					Log(eachLevel, "Host Info")
			} else if dataErr != nil || len(data) <= 0 {
				logger.WithFields(logrus.Fields{
					"filepath": eachEntry,
					"error":    dataErr,
					"length":   len(data),
				}).Warn("Failed to read host info")
			}
		}
	}
}

// RegisterCodePipelineEnvironment is not available during lambda execution
func RegisterCodePipelineEnvironment(environmentName string, environmentVariables map[string]string) error {
	return nil
}

// NewLoggerWithFormatter always returns a JSON formatted logger
// that is aware of the environment variable that may have been
// set and carried through to the AWS Lambda execution environment
func NewLoggerWithFormatter(level string, formatter logrus.Formatter) (*logrus.Logger, error) {

	logger := logrus.New()
	// If there is an environment override, use that
	envLogLevel := os.Getenv(envVarLogLevel)
	if envLogLevel != "" {
		level = envLogLevel
	}
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		return nil, err
	}
	logger.Level = logLevel
	// We always use JSON in AWS
	logger.Formatter = &logrus.JSONFormatter{}

	// TODO - consider writing a buffered logger that only
	// writes output following an error.
	// This was done as part of the XRay interceptor!
	logger.Out = os.Stdout
	return logger, nil
}
