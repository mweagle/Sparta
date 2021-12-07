package sparta

import (
	"bytes"
	cryptoRand "crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	validator "gopkg.in/go-playground/validator.v9"
)

func init() {
	validate = validator.New()
	codePipelineEnvironments = make(map[string]map[string]string)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	instanceID = fmt.Sprintf("i-%d", r.Int63())
}

var (
	// Have we already shown the header/usage?
	headerDisplayed = false

	// The Lambda instance ID for this execution
	instanceID string

	// Validation instance
	validate *validator.Validate

	// CodePipeline environments
	codePipelineEnvironments map[string]map[string]string
)

func isRunningInAWS() bool {
	return len(os.Getenv("AWS_LAMBDA_FUNCTION_NAME")) != 0
}

func displayPrettyHeader(headerDivider string, disableColors bool, logger *zerolog.Logger) {
	if headerDisplayed {
		return
	}
	headerDisplayed = true
	logger.Info().Msg(colorize(headerDivider, colorRed, disableColors))
	logger.Info().Msg(fmt.Sprintf(colorize(`╔═╗┌─┐┌─┐┬─┐┌┬┐┌─┐`, colorRed, disableColors)+"   Version : %s", SpartaVersion))
	logger.Info().Msg(fmt.Sprintf(colorize(`╚═╗├─┘├─┤├┬┘ │ ├─┤`, colorRed, disableColors)+"   SHA     : %s", SpartaGitHash[0:7]))
	logger.Info().Msg(fmt.Sprintf(colorize(`╚═╝┴  ┴ ┴┴└─ ┴ ┴ ┴`, colorRed, disableColors)+"   Go      : %s", runtime.Version()))
	logger.Info().Msg(colorize(headerDivider, colorRed, disableColors))
}

func templateOutputFile(outputDir string, serviceName string) (*os.File, error) {
	// Ok, for this we're going some way to tell the Build Command
	// where to write the output...I suppose we could just use a TeeWriter...
	sanitizedServiceName := sanitizedName(serviceName)
	templateName := fmt.Sprintf("%s-cftemplate.json", sanitizedServiceName)
	templateFilePath := filepath.Join(outputDir, templateName)
	mkdirErr := os.MkdirAll(outputDir, os.ModePerm)
	if nil != mkdirErr {
		return nil, errors.Wrapf(mkdirErr, "Attempting to create output directory: %s", outputDir)
	}
	return os.Create(templateFilePath)
}

// Logger returns the sparta Logger instance for this process
func Logger() *zerolog.Logger {
	return OptionsGlobal.Logger
}

// InstanceID returns the uniquely assigned instanceID for this lambda
// container. The InstanceID is created at the time the this Lambda function
// initializes
func InstanceID() string {
	return instanceID
}

// CommandLineOptions defines the commands available via the Sparta command
// line interface.  Embedding applications can extend existing commands
// and add their own to the `Root` command.  See https://github.com/spf13/cobra
// for more information.
var CommandLineOptions = struct {
	Root      *cobra.Command
	Version   *cobra.Command
	Build     *cobra.Command
	Provision *cobra.Command
	Delete    *cobra.Command
	Execute   *cobra.Command
	Describe  *cobra.Command
	Explore   *cobra.Command
	Profile   *cobra.Command
	Status    *cobra.Command
}{}

/*============================================================================*/
// Build options
// Ref: http://docs.aws.amazon.com/AmazonS3/latest/dev/BucketRestrictions.html
type optionsBuildStruct struct {
	BuildID    string `validate:"-"` // non-whitespace
	OutputDir  string `validate:"-"` // non-whitespace
	DockerFile string `validate:"-"` // non-whitespace
}

func computeBuildID(userSuppliedValue string, logger *zerolog.Logger) (string, error) {
	buildID := userSuppliedValue
	if buildID == "" {
		// That's cool, let's see if we can find a git SHA
		cmd := exec.Command("git",
			"rev-parse",
			"HEAD")
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		cmdErr := cmd.Run()
		if cmdErr == nil {
			// Great, let's use the SHA
			buildID = strings.TrimSpace(string(stdout.String()))
			if buildID != "" {
				logger.Info().
					Str("SHA", buildID).
					Str("Command", "git rev-parse HEAD").
					Msg("Using `git` SHA for StampedBuildID")
			}
		}
		// Ignore any errors and make up a random one
		if buildID == "" {
			// No problem, let's use an arbitrary SHA
			hash := sha1.New()
			randomBytes := make([]byte, 256)
			_, err := cryptoRand.Read(randomBytes)
			if err != nil {
				return "", err
			}
			_, err = hash.Write(randomBytes)
			if err != nil {
				return "", err
			}
			buildID = hex.EncodeToString(hash.Sum(nil))
		}
	}
	return buildID, nil
}

var optionsBuild optionsBuildStruct

/*============================================================================*/
// Provision options
// Ref: http://docs.aws.amazon.com/AmazonS3/latest/dev/BucketRestrictions.html
type optionsProvisionStruct struct {
	optionsBuildStruct
	StackParams     []string
	StackTags       []string
	S3Bucket        string `validate:"required"`
	PipelineTrigger string `validate:"-"`
	InPlace         bool   `validate:"-"`
	stackParams     map[string]string
	stackTags       map[string]string
}

func (ops *optionsProvisionStruct) parseParams() error {

	splitter := func(eachVal string) []string {
		parts := strings.SplitN(eachVal, "=", 2)
		keyName := parts[0]
		paramVal := ""
		if len(parts) > 1 {
			unquoteVal, unquoteValErr := strconv.Unquote(parts[1])
			if unquoteValErr == nil {
				paramVal = unquoteVal
			} else {
				paramVal = parts[1]
			}
		}
		return []string{keyName, paramVal}
	}

	ops.stackParams = make(map[string]string)
	for _, eachPair := range ops.StackParams {
		pairVals := splitter(eachPair)
		ops.stackParams[pairVals[0]] = pairVals[1]
	}
	// Special affordance for S3 bucket
	ops.stackParams[StackParamArtifactBucketName] = ops.S3Bucket

	// Tags, including user defined
	ops.stackTags = map[string]string{
		SpartaTagBuildIDKey:       StampedBuildID,
		SpartaTagSpartaVersionKey: SpartaVersion,
	}
	for _, eachPair := range ops.StackTags {
		pairVals := splitter(eachPair)
		ops.stackTags[pairVals[0]] = pairVals[1]
	}
	return nil
}

var optionsProvision optionsProvisionStruct

/*============================================================================*/
// Describe options
type optionsDescribeStruct struct {
	OutputFile string `validate:"required"`
	S3Bucket   string `validate:"required"`
}

var optionsDescribe optionsDescribeStruct

/*============================================================================*/
// Explore options?
type optionsExploreStruct struct {
	InputExtensions []string `validate:"-"`
}

var optionsExplore optionsExploreStruct

/*============================================================================*/
// Profile options
type optionsProfileStruct struct {
	S3Bucket string `validate:"required"`
	Port     int    `validate:"-"`
}

var optionsProfile optionsProfileStruct

/*============================================================================*/
// Status options
type optionsStatusStruct struct {
	Redact bool `validate:"-"`
}

var optionsStatus optionsStatusStruct

/*============================================================================*/
// Initialization
// Initialize all the Cobra commands and their associated flags
/*============================================================================*/
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
	CommandLineOptions.Root.PersistentFlags().BoolVarP(&OptionsGlobal.TimeStamps,
		"timestamps",
		"z",
		false,
		"Include UTC timestamp log line prefix")
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

	// Support disabling log colors for CLI friendliness
	CommandLineOptions.Root.PersistentFlags().BoolVarP(&OptionsGlobal.DisableColors,
		"nocolor",
		"",
		false,
		"Boolean flag to suppress colorized TTY output")

	// Version
	CommandLineOptions.Version = &cobra.Command{
		Use:          "version",
		Short:        "Display version information",
		Long:         `Displays the Sparta framework version `,
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {

		},
	}
	// Build
	CommandLineOptions.Build = &cobra.Command{
		Use:          "build",
		Short:        "Build service",
		Long:         `Builds the binary and associated CloudFormation parameterized template`,
		SilenceUsage: true,
	}
	CommandLineOptions.Build.Flags().StringVarP(&optionsBuild.BuildID,
		"buildID",
		"i",
		"",
		"Optional BuildID to use")
	CommandLineOptions.Build.Flags().StringVarP(&optionsBuild.OutputDir,
		"outputDir",
		"o",
		ScratchDirectory,
		"Optional output directory for artifacts")
	CommandLineOptions.Build.Flags().StringVarP(&optionsBuild.DockerFile,
		"dockerFile",
		"d",
		"",
		"Optional Dockerfile path to use OCI image rather than ZIP")

	// Provision
	CommandLineOptions.Provision = &cobra.Command{
		Use:          "provision",
		Short:        "Provision service",
		Long:         `Provision the service (either create or update) via CloudFormation`,
		SilenceUsage: true,
	}
	CommandLineOptions.Provision.Flags().StringArrayVarP(&optionsProvision.StackParams,
		"param",
		"m",
		[]string{},
		"List of params in A=B format")
	CommandLineOptions.Provision.Flags().StringArrayVarP(&optionsProvision.StackTags,
		"tag",
		"g",
		[]string{},
		"List of Stack Tags in A=B format")
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
		"Name of CodePipeline package that includes cloudformation.json Template and ZIP config files")
	CommandLineOptions.Provision.Flags().BoolVarP(&optionsProvision.InPlace,
		"inplace",
		"c",
		false,
		"If the provision operation results in *only* function updates, bypass CloudFormation")
	CommandLineOptions.Provision.Flags().StringVarP(&optionsProvision.OutputDir,
		"outputDir",
		"o",
		ScratchDirectory,
		"Optional output directory for artifacts")
	CommandLineOptions.Provision.Flags().StringVarP(&optionsProvision.DockerFile,
		"dockerFile",
		"d",
		"",
		"Optional Dockerfile path")

	// Delete
	CommandLineOptions.Delete = &cobra.Command{
		Use:          "delete",
		Short:        "Delete service",
		Long:         `Ensure service is successfully deleted`,
		SilenceUsage: true,
	}

	// Execute
	CommandLineOptions.Execute = &cobra.Command{
		Use:          "execute",
		Short:        "Start the application and begin handling events",
		Long:         `Start the application and begin handling events`,
		SilenceUsage: true,
	}

	// Describe
	CommandLineOptions.Describe = &cobra.Command{
		Use:          "describe",
		Short:        "Describe service",
		Long:         `Produce an HTML report of the service`,
		SilenceUsage: true,
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
		Use:          "explore",
		Short:        "Interactively explore a provisioned service",
		Long:         `Startup a local CLI GUI to explore and trigger your AWS service`,
		SilenceUsage: true,
	}
	CommandLineOptions.Explore.Flags().StringArrayVarP(&optionsExplore.InputExtensions,
		"inputExtension",
		"e",
		[]string{"json"},
		"One or more file extensions to include as sample inputs")

	// Profile
	CommandLineOptions.Profile = &cobra.Command{
		Use:          "profile",
		Short:        "Interactively examine service pprof output",
		Long:         `Startup a local pprof webserver to interrogate profiles snapshots on S3`,
		SilenceUsage: true,
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

	// Status
	CommandLineOptions.Status = &cobra.Command{
		Use:          "status",
		Short:        "Produce a report for a provisioned service",
		Long:         `Produce a report for a provisioned service`,
		SilenceUsage: true,
	}
	CommandLineOptions.Status.Flags().BoolVarP(&optionsStatus.Redact, "redact",
		"r",
		false,
		"Redact AWS Account ID from report")
}

// CommandLineOptionsHook allows embedding applications the ability
// to validate caller-defined command line arguments.  Return an error
// if the command line fails.
type CommandLineOptionsHook func(command *cobra.Command) error

// ParseOptions parses the command line options
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
			logger, loggerErr := NewLoggerForOutput(OptionsGlobal.LogLevel,
				OptionsGlobal.LogFormat,
				OptionsGlobal.DisableColors)
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
		CommandLineOptions.Build,
		CommandLineOptions.Provision,
		CommandLineOptions.Delete,
		CommandLineOptions.Execute,
		CommandLineOptions.Describe,
		CommandLineOptions.Explore,
		CommandLineOptions.Profile,
		CommandLineOptions.Status,
	}
	for _, eachCommand := range spartaCommands {
		eachCommand.PreRunE = func(cmd *cobra.Command, args []string) error {
			switch eachCommand {
			case CommandLineOptions.Build:
				StampedBuildID = optionsBuild.BuildID
			case CommandLineOptions.Provision:
				StampedBuildID = optionsProvision.BuildID
			default:
				// NOP
			}

			if handler != nil {
				return handler(eachCommand)
			}
			return nil
		}
		parseCmdRoot.AddCommand(CommandLineOptions.Version)
	}

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
		executeErr = parseCmdRoot.Root().Help()
	}
	return executeErr
}
