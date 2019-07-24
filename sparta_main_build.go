// +build !lambdabinary

package sparta

import (
	"fmt"
	"log"
	"os"
	"time"

	validator "gopkg.in/go-playground/validator.v9"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func platformLogSysInfo(lambdaFunc string, logger *logrus.Logger) {
	// NOP
}

// RegisterCodePipelineEnvironment is part of a CodePipeline deployment
// and defines the environments available for deployment. Environments
// are defined the `environmentName`. The values defined in the
// environmentVariables are made available to each service as
// environment variables. The environment key will be transformed into
// a configuration file for a CodePipeline CloudFormation action:
// TemplateConfiguration: !Sub "TemplateSource::${environmentName}".
func RegisterCodePipelineEnvironment(environmentName string,
	environmentVariables map[string]string) error {
	if _, exists := codePipelineEnvironments[environmentName]; exists {
		return errors.Errorf("Environment (%s) has already been defined", environmentName)
	}
	codePipelineEnvironments[environmentName] = environmentVariables
	return nil
}

// NewLoggerWithFormatter returns a logger with the given formatter. If formatter
// is nil, a TTY-aware formatter is used
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
	if nil != formatter {
		logger.Formatter = formatter
	}
	logger.Out = os.Stdout
	return logger, nil
}

// Main defines the primary handler for transforming an application into a Sparta package.  The
// serviceName is used to uniquely identify your service within a region and will
// be used for subsequent updates.  For provisioning, ensure that you've
// properly configured AWS credentials for the golang SDK.
// See http://docs.aws.amazon.com/sdk-for-go/api/aws/defaults.html#DefaultChainCredentials-constant
// for more information.
func Main(serviceName string, serviceDescription string, lambdaAWSInfos []*LambdaAWSInfo, api APIGateway, site *S3Site) error {
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

	//////////////////////////////////////////////////////////////////////////////
	// cmdRoot defines the root, non-executable command
	CommandLineOptions.Root.Short = fmt.Sprintf("%s - Sparta v.%s powered AWS Lambda Microservice",
		serviceName,
		SpartaVersion)
	CommandLineOptions.Root.Long = serviceDescription
	CommandLineOptions.Root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// Save the ServiceName in case a custom command wants it
		OptionsGlobal.ServiceName = serviceName
		OptionsGlobal.ServiceDescription = serviceDescription

		validateErr := validate.Struct(OptionsGlobal)
		if nil != validateErr {
			return validateErr
		}

		// Format?
		// Running in AWS?
		disableColors := OptionsGlobal.DisableColors || isRunningInAWS()
		var formatter logrus.Formatter
		switch OptionsGlobal.LogFormat {
		case "text", "txt":
			formatter = &logrus.TextFormatter{
				DisableColors: disableColors,
				FullTimestamp: OptionsGlobal.TimeStamps,
			}
		case "json":
			formatter = &logrus.JSONFormatter{}
			disableColors = true
		}
		logger, loggerErr := NewLoggerWithFormatter(OptionsGlobal.LogLevel, formatter)
		if nil != loggerErr {
			return loggerErr
		}

		// This is a NOP, but makes megacheck happy b/c it doesn't know about
		// build flags
		platformLogSysInfo("", logger)
		OptionsGlobal.Logger = logger
		welcomeMessage := fmt.Sprintf("Service: %s", serviceName)

		// Header information...
		displayPrettyHeader(headerDivider, disableColors, logger)
		// Metadata about the build...
		logger.WithFields(logrus.Fields{
			"Option":    cmd.Name(),
			"UTC":       (time.Now().UTC().Format(time.RFC3339)),
			"LinkFlags": OptionsGlobal.LinkerFlags,
		}).Info(welcomeMessage)
		logger.Info(headerDivider)

		return nil
	}

	//////////////////////////////////////////////////////////////////////////////
	// Version
	CommandLineOptions.Root.AddCommand(CommandLineOptions.Version)

	//////////////////////////////////////////////////////////////////////////////
	// Provision
	CommandLineOptions.Provision.PreRunE = func(cmd *cobra.Command, args []string) error {
		validateErr := validate.Struct(optionsProvision)

		OptionsGlobal.Logger.WithFields(logrus.Fields{
			"validateErr":      validateErr,
			"optionsProvision": optionsProvision,
		}).Debug("Provision validation results")
		return validateErr
	}

	if nil == CommandLineOptions.Provision.RunE {
		CommandLineOptions.Provision.RunE = func(cmd *cobra.Command, args []string) error {
			buildID, buildIDErr := provisionBuildID(optionsProvision.BuildID, OptionsGlobal.Logger)
			if nil != buildIDErr {
				return buildIDErr
			}
			// Save the BuildID
			StampedBuildID = buildID
			return Provision(OptionsGlobal.Noop,
				serviceName,
				serviceDescription,
				lambdaAWSInfos,
				api,
				site,
				optionsProvision.S3Bucket,
				useCGO,
				optionsProvision.InPlace,
				buildID,
				optionsProvision.PipelineTrigger,
				OptionsGlobal.BuildTags,
				OptionsGlobal.LinkerFlags,
				nil,
				workflowHooks,
				OptionsGlobal.Logger)
		}
	}
	CommandLineOptions.Root.AddCommand(CommandLineOptions.Provision)

	//////////////////////////////////////////////////////////////////////////////
	// Delete
	CommandLineOptions.Delete.RunE = func(cmd *cobra.Command, args []string) error {
		return Delete(serviceName, OptionsGlobal.Logger)
	}

	CommandLineOptions.Root.AddCommand(CommandLineOptions.Delete)

	//////////////////////////////////////////////////////////////////////////////
	// Execute
	if nil == CommandLineOptions.Execute.RunE {
		CommandLineOptions.Execute.RunE = func(cmd *cobra.Command, args []string) error {

			OptionsGlobal.Logger.Formatter = new(logrus.JSONFormatter)
			// Ensure the discovery service is initialized
			initializeDiscovery(OptionsGlobal.Logger)

			return Execute(serviceName,
				lambdaAWSInfos,
				OptionsGlobal.Logger)
		}
	}
	CommandLineOptions.Root.AddCommand(CommandLineOptions.Execute)

	//////////////////////////////////////////////////////////////////////////////
	// Describe
	if nil == CommandLineOptions.Describe.RunE {
		CommandLineOptions.Describe.RunE = func(cmd *cobra.Command, args []string) error {
			validateErr := validate.Struct(optionsDescribe)
			if nil != validateErr {
				return validateErr
			}

			fileWriter, fileWriterErr := os.Create(optionsDescribe.OutputFile)
			if fileWriterErr != nil {
				return fileWriterErr
			}
			defer fileWriter.Close()
			describeErr := Describe(serviceName,
				serviceDescription,
				lambdaAWSInfos,
				api,
				site,
				optionsDescribe.S3Bucket,
				OptionsGlobal.BuildTags,
				OptionsGlobal.LinkerFlags,
				fileWriter,
				workflowHooks,
				OptionsGlobal.Logger)

			if describeErr == nil {
				describeErr = fileWriter.Sync()
			}
			return describeErr
		}
	}
	CommandLineOptions.Root.AddCommand(CommandLineOptions.Describe)

	//////////////////////////////////////////////////////////////////////////////
	// Explore
	if nil == CommandLineOptions.Explore.RunE {
		CommandLineOptions.Explore.RunE = func(cmd *cobra.Command, args []string) error {
			validateErr := validate.Struct(optionsExplore)
			if nil != validateErr {
				return validateErr
			}

			return Explore(serviceName,
				serviceDescription,
				lambdaAWSInfos,
				api,
				site,
				optionsDescribe.S3Bucket,
				OptionsGlobal.BuildTags,
				OptionsGlobal.LinkerFlags,
				OptionsGlobal.Logger)
		}
	}
	CommandLineOptions.Root.AddCommand(CommandLineOptions.Explore)

	//////////////////////////////////////////////////////////////////////////////
	// Profile
	if nil == CommandLineOptions.Profile.RunE {
		CommandLineOptions.Profile.RunE = func(cmd *cobra.Command, args []string) error {
			validateErr := validate.Struct(optionsProfile)
			if nil != validateErr {
				return validateErr
			}
			return Profile(serviceName,
				serviceDescription,
				optionsProfile.S3Bucket,
				optionsProfile.Port,
				OptionsGlobal.Logger)
		}
	}
	CommandLineOptions.Root.AddCommand(CommandLineOptions.Profile)

	//////////////////////////////////////////////////////////////////////////////
	// Status
	if nil == CommandLineOptions.Status.RunE {
		CommandLineOptions.Status.RunE = func(cmd *cobra.Command, args []string) error {
			validateErr := validate.Struct(optionsStatus)
			if nil != validateErr {
				return validateErr
			}
			return Status(serviceName,
				serviceDescription,
				optionsStatus.Redact,
				OptionsGlobal.Logger)
		}
	}
	CommandLineOptions.Root.AddCommand(CommandLineOptions.Status)

	// Run it!
	executedCmd, executeErr := CommandLineOptions.Root.ExecuteC()
	if executeErr != nil {
		if OptionsGlobal.Logger == nil {
			newLogger, newLoggerErr := NewLogger("info")
			if newLoggerErr != nil {
				fmt.Printf("Failed to create new logger: %v", newLoggerErr)
				newLogger = logrus.New()
			}
			OptionsGlobal.Logger = newLogger
		}
		if OptionsGlobal.Logger != nil {
			validationErr, validationErrOk := executeErr.(validator.ValidationErrors)
			if validationErrOk {
				for _, eachError := range validationErr {
					OptionsGlobal.Logger.Error(eachError)
				}
				// Only show the usage if there were input validation errors
				if executedCmd != nil {
					usageErr := executedCmd.Usage()
					if usageErr != nil {
						OptionsGlobal.Logger.Error(usageErr)
					}
				}
			} else {
				OptionsGlobal.Logger.Error(executeErr)
			}
		} else {
			log.Printf("ERROR: %s", executeErr)
		}
	}

	// Cleanup, if for some reason the caller wants to re-execute later...
	CommandLineOptions.Root.PersistentPreRunE = nil
	return executeErr
}
