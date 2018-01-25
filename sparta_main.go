package sparta

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"path"
	"runtime"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/go-playground/validator.v9"
)

// Constant for Sparta color aware stdout logging
const (
	redCode = 31
)

// Validation instance
var validate *validator.Validate

func isRunningInAWS() bool {
	return len(os.Getenv("AWS_LAMBDA_FUNCTION_NAME")) != 0
}

func displayPrettyHeader(headerDivider string, enableColors bool, logger *logrus.Logger) {
	logger.Info(headerDivider)
	if enableColors {
		red := func(inputText string) string {
			return fmt.Sprintf("\x1b[%dm%s\x1b[0m", redCode, inputText)
		}
		logger.Info(fmt.Sprintf(red("╔═╗┌─┐┌─┐┬─┐┌┬┐┌─┐")+"   Version : %s", SpartaVersion))
		logger.Info(fmt.Sprintf(red("╚═╗├─┘├─┤├┬┘ │ ├─┤")+"   SHA     : %s", SpartaGitHash[0:7]))
		logger.Info(fmt.Sprintf(red("╚═╝┴  ┴ ┴┴└─ ┴ ┴ ┴")+"   Go      : %s", runtime.Version()))
	} else {
		logger.Info(fmt.Sprintf(`╔═╗┌─┐┌─┐┬─┐┌┬┐┌─┐   Version : %s`, SpartaVersion))
		logger.Info(fmt.Sprintf(`╚═╗├─┘├─┤├┬┘ │ ├─┤   SHA     : %s`, SpartaGitHash[0:7]))
		logger.Info(fmt.Sprintf(`╚═╝┴  ┴ ┴┴└─ ┴ ┴ ┴   Go      : %s`, runtime.Version()))
	}
	logger.Info(headerDivider)
}

var codePipelineEnvironments map[string]map[string]string

func init() {
	validate = validator.New()
	codePipelineEnvironments = make(map[string]map[string]string)
}

// Logger returns the sparta Logger instance for this process
func Logger() *logrus.Logger {
	return OptionsGlobal.Logger
}

// CommandLineOptions defines the commands available via the Sparta command
// line interface.  Embedding applications can extend existing commands
// and add their own to the `Root` command.  See https://github.com/spf13/cobra
// for more information.
var CommandLineOptions = struct {
	Root      *cobra.Command
	Version   *cobra.Command
	Provision *cobra.Command
	Delete    *cobra.Command
	Execute   *cobra.Command
	Describe  *cobra.Command
	Explore   *cobra.Command
	Profile   *cobra.Command
}{}

/******************************************************************************/
// Provision options
// Ref: http://docs.aws.amazon.com/AmazonS3/latest/dev/BucketRestrictions.html
type optionsProvisionStruct struct {
	S3Bucket        string `validate:"required"`
	BuildID         string `validate:"-"` // non-whitespace
	PipelineTrigger string `validate:"-"`
	InPlace         bool   `validate:"-"`
}

var optionsProvision optionsProvisionStruct

func provisionBuildID(userSuppliedValue string) (string, error) {
	buildID := userSuppliedValue
	if "" == buildID {
		hash := sha1.New()
		randomBytes := make([]byte, 256)
		_, err := rand.Read(randomBytes)
		if err != nil {
			return "", err
		}
		hash.Write(randomBytes)
		buildID = hex.EncodeToString(hash.Sum(nil))
	}
	return buildID, nil
}

/******************************************************************************/
// Describe options
type optionsDescribeStruct struct {
	OutputFile string `validate:"required"`
	S3Bucket   string `validate:"required"`
}

var optionsDescribe optionsDescribeStruct

/******************************************************************************/
// Profile options
type optionsProfileStruct struct {
	S3Bucket string `validate:"required"`
	Port     int    `validate:"-"`
}

var optionsProfile optionsProfileStruct

/******************************************************************************/
// Initialization
// Initialize all the Cobra commands and their associated flags
/******************************************************************************/
func init() {
	// Root
	CommandLineOptions.Root = &cobra.Command{
		Use:           path.Base(os.Args[0]),
		Short:         "Sparta-powered AWS Lambda microservice",
		SilenceErrors: true,
	}
	CommandLineOptions.Root.PersistentFlags().BoolVarP(&OptionsGlobal.Noop, "noop",
		"n",
		false,
		"Dry-run behavior only (do not perform mutations)")
	CommandLineOptions.Root.PersistentFlags().StringVarP(&OptionsGlobal.LogLevel,
		"level",
		"l",
		"info",
		"Log level [panic, fatal, error, warn, info, debug]")
	CommandLineOptions.Root.PersistentFlags().StringVarP(&OptionsGlobal.LogFormat,
		"format",
		"f",
		"text",
		"Log format [text, json]")
	CommandLineOptions.Root.PersistentFlags().StringVarP(&OptionsGlobal.BuildTags,
		"tags",
		"t",
		"",
		"Optional build tags for conditional compilation")
	// Make sure there's a place to put any linker flags
	CommandLineOptions.Root.PersistentFlags().StringVar(&OptionsGlobal.LinkerFlags,
		"ldflags",
		"",
		"Go linker string definition flags (https://golang.org/cmd/link/)")

	// Version
	CommandLineOptions.Version = &cobra.Command{
		Use:   "version",
		Short: "Display version information",
		Long:  `Displays the Sparta framework version `,
		Run: func(cmd *cobra.Command, args []string) {

		},
	}
	// Provision
	CommandLineOptions.Provision = &cobra.Command{
		Use:   "provision",
		Short: "Provision service",
		Long:  `Provision the service (either create or update) via CloudFormation`,
	}
	CommandLineOptions.Provision.Flags().StringVarP(&optionsProvision.S3Bucket,
		"s3Bucket",
		"s",
		"",
		"S3 Bucket to use for Lambda source")
	CommandLineOptions.Provision.Flags().StringVarP(&optionsProvision.BuildID,
		"buildID",
		"i",
		"",
		"Optional BuildID to use")
	CommandLineOptions.Provision.Flags().StringVarP(&optionsProvision.PipelineTrigger,
		"codePipelinePackage",
		"p",
		"",
		"Name of CodePipeline package that includes cloduformation.json Template and ZIP config files")
	CommandLineOptions.Provision.Flags().BoolVarP(&optionsProvision.InPlace,
		"inplace",
		"c",
		false,
		"If the provision operation results in *only* function updates, bypass CloudFormation")

	// Delete
	CommandLineOptions.Delete = &cobra.Command{
		Use:   "delete",
		Short: "Delete service",
		Long:  `Ensure service is successfully deleted`,
	}

	// Execute
	CommandLineOptions.Execute = &cobra.Command{
		Use:   "execute",
		Short: "Execute",
		Long:  `Startup the localhost HTTP server to handle requests`,
	}

	// Describe
	CommandLineOptions.Describe = &cobra.Command{
		Use:   "describe",
		Short: "Describe service",
		Long:  `Produce an HTML report of the service`,
	}
	CommandLineOptions.Describe.Flags().StringVarP(&optionsDescribe.OutputFile,
		"out",
		"o",
		"",
		"Output file for HTML description")
	CommandLineOptions.Describe.Flags().StringVarP(&optionsDescribe.S3Bucket,
		"s3Bucket",
		"s",
		"",
		"S3 Bucket to use for Lambda source")

	// Explore
	CommandLineOptions.Explore = &cobra.Command{
		Use:   "explore",
		Short: "Interactively explore service",
		Long:  `Startup a localhost HTTP server to explore the exported Go functions`,
	}

	// Profile
	CommandLineOptions.Profile = &cobra.Command{
		Use:   "profile",
		Short: "Interactively examine service pprof output",
		Long:  `Startup a local pprof webserver to interrogate profiles snapshots on S3`,
	}
	CommandLineOptions.Profile.Flags().StringVarP(&optionsProfile.S3Bucket,
		"s3Bucket",
		"s",
		"",
		"S3 Bucket that stores lambda profile snapshots")
	CommandLineOptions.Profile.Flags().IntVarP(&optionsProfile.Port,
		"port",
		"p",
		8080,
		"Alternative port for `pprof` web UI (default=8080)")
}

// CommandLineOptionsHook allows embedding applications the ability
// to validate caller-defined command line arguments.  Return an error
// if the command line fails.
type CommandLineOptionsHook func(command *cobra.Command) error

// ParseOptions the command line options
func ParseOptions(handler CommandLineOptionsHook) error {
	// First up, create a dummy Root command for the parse...
	var parseCmdRoot = &cobra.Command{
		Use:           CommandLineOptions.Root.Use,
		Short:         CommandLineOptions.Root.Short,
		SilenceUsage:  true,
		SilenceErrors: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	parseCmdRoot.PersistentFlags().BoolVarP(&OptionsGlobal.Noop, "noop",
		"n",
		false,
		"Dry-run behavior only (do not perform mutations)")
	parseCmdRoot.PersistentFlags().StringVarP(&OptionsGlobal.LogLevel,
		"level",
		"l",
		"info",
		"Log level [panic, fatal, error, warn, info, debug]")
	parseCmdRoot.PersistentFlags().StringVarP(&OptionsGlobal.LogFormat,
		"format",
		"f",
		"text",
		"Log format [text, json]")
	parseCmdRoot.PersistentFlags().StringVarP(&OptionsGlobal.BuildTags,
		"tags",
		"t",
		"",
		"Optional build tags for conditional compilation")

	// Now, for any user-attached commands, add them to the temporary Parse
	// root command.
	for _, eachUserCommand := range CommandLineOptions.Root.Commands() {
		userProxyCmd := &cobra.Command{
			Use:   eachUserCommand.Use,
			Short: eachUserCommand.Short,
		}
		userProxyCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
			validateErr := validate.Struct(OptionsGlobal)
			if nil != validateErr {
				return validateErr
			}
			// Format?
			var formatter logrus.Formatter
			switch OptionsGlobal.LogFormat {
			case "text", "txt":
				formatter = &logrus.TextFormatter{}
			case "json":
				formatter = &logrus.JSONFormatter{}
			}
			logger, loggerErr := NewLoggerWithFormatter(OptionsGlobal.LogLevel, formatter)
			if nil != loggerErr {
				return loggerErr
			}
			OptionsGlobal.Logger = logger

			if handler != nil {
				return handler(userProxyCmd)
			}
			return nil
		}
		userProxyCmd.Flags().AddFlagSet(eachUserCommand.Flags())
		parseCmdRoot.AddCommand(userProxyCmd)
	}

	//////////////////////////////////////////////////////////////////////////////
	// Then add the standard Sparta ones...
	spartaCommands := []*cobra.Command{
		CommandLineOptions.Version,
		CommandLineOptions.Provision,
		CommandLineOptions.Delete,
		CommandLineOptions.Execute,
		CommandLineOptions.Describe,
		CommandLineOptions.Explore,
		CommandLineOptions.Profile,
	}
	CommandLineOptions.Version.PreRunE = func(cmd *cobra.Command, args []string) error {
		if handler != nil {
			return handler(CommandLineOptions.Version)
		}
		return nil
	}
	parseCmdRoot.AddCommand(CommandLineOptions.Version)

	CommandLineOptions.Provision.PreRunE = func(cmd *cobra.Command, args []string) error {
		if handler != nil {
			return handler(CommandLineOptions.Provision)
		}
		return nil
	}
	parseCmdRoot.AddCommand(CommandLineOptions.Provision)

	CommandLineOptions.Delete.PreRunE = func(cmd *cobra.Command, args []string) error {
		if handler != nil {
			return handler(CommandLineOptions.Delete)
		}
		return nil
	}
	parseCmdRoot.AddCommand(CommandLineOptions.Delete)

	CommandLineOptions.Execute.PreRunE = func(cmd *cobra.Command, args []string) error {
		if handler != nil {
			return handler(CommandLineOptions.Execute)
		}
		return nil
	}
	parseCmdRoot.AddCommand(CommandLineOptions.Execute)

	CommandLineOptions.Describe.PreRunE = func(cmd *cobra.Command, args []string) error {
		if handler != nil {
			return handler(CommandLineOptions.Describe)
		}
		return nil
	}
	parseCmdRoot.AddCommand(CommandLineOptions.Describe)

	CommandLineOptions.Explore.PreRunE = func(cmd *cobra.Command, args []string) error {
		if handler != nil {
			return handler(CommandLineOptions.Explore)
		}
		return nil
	}
	parseCmdRoot.AddCommand(CommandLineOptions.Explore)

	CommandLineOptions.Profile.PreRunE = func(cmd *cobra.Command, args []string) error {
		if handler != nil {
			return handler(CommandLineOptions.Profile)
		}
		return nil
	}
	parseCmdRoot.AddCommand(CommandLineOptions.Profile)

	// Assign each command an empty RunE func s.t.
	// Cobra doesn't print out the command info
	for _, eachCommand := range parseCmdRoot.Commands() {
		eachCommand.RunE = func(cmd *cobra.Command, args []string) error {
			return nil
		}
	}
	// Intercept the usage command - we'll end up showing this later
	// in Main...If there is an error, we will show help there...
	parseCmdRoot.SetHelpFunc(func(*cobra.Command, []string) {
		// Swallow help here
	})

	// Run it...
	executeErr := parseCmdRoot.Execute()

	// Cleanup the Sparta specific ones
	for _, eachCmd := range spartaCommands {
		eachCmd.RunE = nil
		eachCmd.PreRunE = nil
	}

	if nil != executeErr {
		parseCmdRoot.SetHelpFunc(nil)
		parseCmdRoot.Root().Help()
	}
	return executeErr
}

// NewLogger returns a new logrus.Logger instance. It is the caller's responsibility
// to set the formatter if needed.
func NewLogger(level string) (*logrus.Logger, error) {
	return NewLoggerWithFormatter(level, nil)
}
