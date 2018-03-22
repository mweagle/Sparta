package sparta

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/go-playground/validator.v9"
)

var exampleValidator *validator.Validate

// NOTE: your application MUST use `package main` and define a `main()` function.  The
// example text is to make the documentation compatible with godoc.
// Should be main() in your application

// Additional command line options used for both the provision
// and CLI commands
type optionsStruct struct {
	Username   string `validate:"required"`
	Password   string `validate:"required"`
	SSHKeyName string `validate:"-"`
}

var options optionsStruct

// Common function to register shared command line flags
// across multiple Sparta commands
func registerSpartaCommandLineFlags(command *cobra.Command) {
	command.Flags().StringVarP(&options.Username,
		"username",
		"u",
		"",
		"HTTP Basic Auth username")
	command.Flags().StringVarP(&options.Password,
		"password",
		"p",
		"",
		"HTTP Basic Auth password")
}

func ExampleParseOptions() {
	//////////////////////////////////////////////////////////////////////////////
	// Add the custom command to run the sync loop
	syncCommand := &cobra.Command{
		Use:   "sync",
		Short: "Periodically perform a task",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Sync command!\n")
			return nil
		},
	}
	// Include the basic auth flags for the sync command
	registerSpartaCommandLineFlags(syncCommand)
	CommandLineOptions.Root.AddCommand(syncCommand)

	//////////////////////////////////////////////////////////////////////////////
	// Register custom flags for pre-existing Sparta commands
	registerSpartaCommandLineFlags(CommandLineOptions.Provision)
	CommandLineOptions.Provision.Flags().StringVarP(&options.SSHKeyName,
		"key",
		"k",
		"",
		"SSH Key Name to use for EC2 instances")

	//////////////////////////////////////////////////////////////////////////////
	// Define a validation hook s.t. we can validate the CLI user input
	validationHook := func(command *cobra.Command) error {
		if command.Name() == "provision" && len(options.SSHKeyName) <= 0 {
			return errors.Errorf("SSHKeyName option is required")
		}
		fmt.Printf("Command: %s\n", command.Name())
		switch command.Name() {
		case "provision",
			"sync":
			validationErr := exampleValidator.Struct(options)
			return errors.Wrapf(validationErr, "Validating input")
		default:
			return nil
		}
	}
	// If the validation hooks failed, exit the application
	parseErr := ParseOptions(validationHook)
	if nil != parseErr {
		os.Exit(3)
	}
	//////////////////////////////////////////////////////////////////////////////
	//
	// Standard Sparta application
	// ...
}
