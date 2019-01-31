package spartamage

import (
	"errors"
	"log"
	"os"
	"strconv"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Log is a mage verbose aware log function
func Log(formatSpecifier string, args ...interface{}) {
	if mg.Verbose() {
		if len(args) != 0 {
			log.Printf(formatSpecifier, args...)
		} else {
			log.Print(formatSpecifier)
		}
	}
}

// Script is a 2d array of commands to run as a script
func Script(commands [][]string) error {
	for _, eachCommand := range commands {
		var commandErr error
		if len(eachCommand) <= 1 {
			commandErr = sh.Run(eachCommand[0])
		} else {
			commandErr = sh.Run(eachCommand[0], eachCommand[1:]...)
		}
		if commandErr != nil {
			return commandErr
		}
	}
	return nil
}

// ApplyToSource is a mage compatible function that applies a
// command to your source tree
func ApplyToSource(fileExtension string,
	ignoredSubdirectories []string,
	commandParts ...string) error {
	if len(commandParts) <= 0 {
		return errors.New("applyToSource requires a command to apply to source files")
	}
	eligibleSourceFiles, eligibleSourceFilesErr := sourceFilesOfType(fileExtension, ignoredSubdirectories)
	if eligibleSourceFilesErr != nil {
		return eligibleSourceFilesErr
	}

	Log(header)
	Log("Applying `%s` to %d `*.%s` source files", commandParts[0], len(eligibleSourceFiles), fileExtension)
	Log(header)

	commandArgs := []string{}
	if len(commandParts) > 1 {
		commandArgs = append(commandArgs, commandParts[1:]...)
	}
	for _, eachFile := range eligibleSourceFiles {
		applyArgs := append(commandArgs, eachFile)
		applyErr := sh.Run(commandParts[0], applyArgs...)
		if applyErr != nil {
			return applyErr
		}
	}
	return nil
}

// SpartaCommand issues a go run command that encapsulates resolving
// global env vars that can be translated into Sparta command line options
func SpartaCommand(commandParts ...string) error {
	noopValue := ""
	parsedBool, parsedBoolErr := strconv.ParseBool(os.Getenv("NOOP"))
	if parsedBoolErr == nil && parsedBool {
		noopValue = "--noop"
	}
	curDir, curDirErr := os.Getwd()
	if curDirErr != nil {
		return errors.New("Failed to get current directory. Error: " + curDirErr.Error())
	}
	setenvErr := os.Setenv(mg.VerboseEnv, "1")
	if setenvErr != nil {
		return setenvErr
	}
	commandArgs := []string{
		"run",
		curDir,
	}
	commandArgs = append(commandArgs, commandParts...)
	if noopValue != "" {
		commandArgs = append(commandArgs, "--noop")
	}
	return sh.Run("go",
		commandArgs...)
}

// Provision deploys the given service
func Provision() error {
	// Get the bucketName
	bucketName := os.Getenv("S3_BUCKET")
	if bucketName == "" {
		return errors.New("Provision requires env.S3_BUCKET to be defined")
	}
	return SpartaCommand("provision", "--s3Bucket", bucketName)
}

// Describe deploys the given service
func Describe() error {
	// Get the bucketName
	bucketName := os.Getenv("S3_BUCKET")
	if bucketName == "" {
		return errors.New("Describe requires env.S3_BUCKET to be defined")
	}
	return SpartaCommand("describe", "--s3Bucket", bucketName, "--out", "graph.html")
}

// Delete deletes the given service
func Delete() error {
	return SpartaCommand("delete")
}

// Explore opens up the terminal GUI
func Explore() error {
	// Get the bucketName
	return SpartaCommand("explore")
}

// Status returns a report for the given status
func Status(plaintext ...bool) error {
	if len(plaintext) == 1 && plaintext[0] {
		return SpartaCommand("status")
	}
	return SpartaCommand("status", "--redact")
}

// Version returns version information about the service and embedded Sparta version
func Version() error {
	return SpartaCommand("version")
}
