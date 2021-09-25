//go:build !lambdabinary
// +build !lambdabinary

package sparta

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	validator "gopkg.in/go-playground/validator.v9"
)

func platformLogSysInfo(lambdaFunc string, logger *zerolog.Logger) {
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
		OptionsGlobal.startTime = time.Now()

		validateErr := validate.Struct(OptionsGlobal)
		if nil != validateErr {
			return validateErr
		}

		// Format?
		// Running in AWS?
		disableColors := OptionsGlobal.DisableColors ||
			isRunningInAWS() ||
			OptionsGlobal.LogFormat == "json"
		logger, loggerErr := NewLoggerForOutput(OptionsGlobal.LogLevel,
			OptionsGlobal.LogFormat,
			disableColors)
		if nil != loggerErr {
			return loggerErr
		}

		// This is a NOP, but makes megacheck happy b/c it doesn't know about
		// build flags
		platformLogSysInfo("", logger)
		OptionsGlobal.Logger = logger
		welcomeMessage := fmt.Sprintf("Service: %s", serviceName)

		// Header information...,
		displayPrettyHeader(headerDivider, disableColors, logger)

		// Metadata about the build...
		logger.Info().
			Str("Option", cmd.Name()).
			Str("LinkFlags", OptionsGlobal.LinkerFlags).
			Str("UTC", time.Now().UTC().Format(time.RFC3339)).
			Msg(welcomeMessage)
		logger.Info().Msg(headerDivider)
		return nil
	}
	CommandLineOptions.Root.PersistentPostRunE = func(cmd *cobra.Command, args []string) error {
		commandTimeDuration := time.Since(OptionsGlobal.startTime)
		OptionsGlobal.Logger.Info().Msg(headerDivider)
		curTime := time.Now()
		OptionsGlobal.Logger.Info().
			Str("Time (UTC)", curTime.UTC().Format(time.RFC3339)).
			Str("Time (Local)", curTime.Format(time.RFC822)).
			Dur(fmt.Sprintf("Duration (%s)", durationUnitLabel), commandTimeDuration).
			Msg("Complete")
		return nil
	}
	//////////////////////////////////////////////////////////////////////////////
	// Version
	CommandLineOptions.Root.AddCommand(CommandLineOptions.Version)

	//////////////////////////////////////////////////////////////////////////////
	// Build
	CommandLineOptions.Build.PreRunE = func(cmd *cobra.Command, args []string) error {
		validateErr := validate.Struct(optionsBuild)

		OptionsGlobal.Logger.Debug().
			Interface("ValidateErr", validateErr).
			Interface("OptionsProvision", optionsProvision).
			Msg("Build validation results")
		return validateErr
	}

	if nil == CommandLineOptions.Build.RunE {
		CommandLineOptions.Build.RunE = func(cmd *cobra.Command, args []string) (provisionErr error) {
			defer func() {
				showOptionalAWSUsageInfo(provisionErr, OptionsGlobal.Logger)
			}()

			buildID, buildIDErr := computeBuildID(optionsProvision.BuildID, OptionsGlobal.Logger)
			if nil != buildIDErr {
				return buildIDErr
			}

			// Save the BuildID
			StampedBuildID = buildID

			// Ok, for this we're going some way to tell the Build Command
			// where to write the output...I suppose we could just use a TeeWriter...
			templateFile, templateFileErr := templateOutputFile(optionsProvision.OutputDir,
				serviceName)
			if templateFileErr != nil {
				return templateFileErr
			}
			buildErr := Build(OptionsGlobal.Noop,
				serviceName,
				serviceDescription,
				lambdaAWSInfos,
				api,
				site,
				useCGO,
				buildID,
				optionsBuild.DockerFile,
				optionsBuild.OutputDir,
				OptionsGlobal.BuildTags,
				OptionsGlobal.LinkerFlags,
				templateFile,
				workflowHooks,
				OptionsGlobal.Logger)
			closeErr := templateFile.Close()
			if closeErr != nil {
				OptionsGlobal.Logger.Warn().
					Err(closeErr).
					Msg("Failed to close template file output")
			}
			return buildErr
		}
	}
	CommandLineOptions.Root.AddCommand(CommandLineOptions.Build)

	//////////////////////////////////////////////////////////////////////////////
	// Provision
	CommandLineOptions.Provision.PreRunE = func(cmd *cobra.Command, args []string) error {
		validateErr := validate.Struct(optionsProvision)

		OptionsGlobal.Logger.Debug().
			Interface("validateErr", validateErr).
			Interface("optionsProvision", optionsProvision).
			Msg("Provision validation results")
		return validateErr
	}

	if nil == CommandLineOptions.Provision.RunE {
		CommandLineOptions.Provision.RunE = func(cmd *cobra.Command, args []string) (provisionErr error) {
			defer func() {
				showOptionalAWSUsageInfo(provisionErr, OptionsGlobal.Logger)
			}()

			buildID, buildIDErr := computeBuildID(optionsProvision.BuildID, OptionsGlobal.Logger)
			if nil != buildIDErr {
				return buildIDErr
			}
			StampedBuildID = buildID

			templateFile, templateFileErr := templateOutputFile(optionsProvision.OutputDir,
				serviceName)
			if templateFileErr != nil {
				return templateFileErr
			}

			// TODO: Build, then Provision
			buildErr := Build(OptionsGlobal.Noop,
				serviceName,
				serviceDescription,
				lambdaAWSInfos,
				api,
				site,
				useCGO,
				buildID,
				optionsProvision.DockerFile,
				optionsProvision.OutputDir,
				OptionsGlobal.BuildTags,
				OptionsGlobal.LinkerFlags,
				templateFile,
				workflowHooks,
				OptionsGlobal.Logger)

			/* #nosec */
			defer func() {
				closeErr := templateFile.Close()
				if closeErr != nil {
					OptionsGlobal.Logger.Warn().
						Err(closeErr).
						Msg("Failed to close template file handle")
				}
			}()

			if buildErr != nil {
				return buildErr
			}
			// So for this, we need to take command
			// line params and turn them into a map...
			parseErr := optionsProvision.parseParams()
			if parseErr != nil {
				return parseErr
			}
			OptionsGlobal.Logger.Debug().
				Interface("params", optionsProvision.stackParams).
				Msg("ParseParams")

			// We don't need to walk the params because we
			// put values in the Metadata block for them all...
			return Provision(OptionsGlobal.Noop,
				templateFile.Name(),
				optionsProvision.stackParams,
				optionsProvision.stackTags,
				optionsProvision.InPlace,
				optionsProvision.PipelineTrigger,
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
				return errors.Wrapf(validateErr, "Failed to validate `describe` options")
			}
			fileWriter, fileWriterErr := os.Create(optionsDescribe.OutputFile)
			if fileWriterErr != nil {
				return fileWriterErr
			}
			/* #nosec */
			defer func() {
				closeErr := fileWriter.Close()
				if closeErr != nil {
					OptionsGlobal.Logger.Warn().
						Err(closeErr).
						Msg("Failed to close describe output writer")
				}
			}()

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

			return ExploreWithInputFilter(serviceName,
				serviceDescription,
				lambdaAWSInfos,
				api,
				site,
				optionsExplore.InputExtensions,
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
			// Use a default console logger
			newLogger, newLoggerErr := NewLoggerForOutput(zerolog.InfoLevel.String(),
				"text",
				isRunningInAWS())
			if newLoggerErr != nil {
				fmt.Printf("Failed to create new logger: %v", newLoggerErr)
				zLogger := zerolog.New(os.Stderr).With().Timestamp().Logger()
				newLogger = &zLogger
			}
			OptionsGlobal.Logger = newLogger
		}
		if OptionsGlobal.Logger != nil {
			validationErr, validationErrOk := executeErr.(validator.ValidationErrors)
			if validationErrOk {
				for _, eachError := range validationErr {
					OptionsGlobal.Logger.Error().
						Interface("Error", eachError).
						Msg("Validation error")
				}
				// Only show the usage if there were input validation errors
				if executedCmd != nil {
					usageErr := executedCmd.Usage()
					if usageErr != nil {
						OptionsGlobal.Logger.Error().Err(usageErr).Msg("Usage error")
					}
				}
			} else {
				displayPrettyHeader(headerDivider, isRunningInAWS(), OptionsGlobal.Logger)
				OptionsGlobal.Logger.Error().Err(executeErr).Msg("Failed to execute command")
			}
		} else {
			log.Printf("ERROR: %s", executeErr)
		}
	}

	// Cleanup, if for some reason the caller wants to re-execute later...
	CommandLineOptions.Root.PersistentPreRunE = nil
	return executeErr
}
