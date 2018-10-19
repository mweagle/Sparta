// +build mage

package magefile

import (
	"errors"
	"log"
	"os"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Log is a mage verbose aware log function
func Log(formatSpecifier string, args ...interface{}) {
	if mg.Verbose() {
		if len(args) != 0 {
			log.Printf(formatSpecifier, args...)
		} else {
			log.Printf(formatSpecifier)
		}
	}
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
		for _, eachPart := range commandParts[1:] {
			commandArgs = append(commandArgs, eachPart)
		}
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

// Provision deploys the given service
func Provision() error {
	// Get the bucketName
	bucketName := os.Getenv("S3_BUCKET")
	if bucketName == "" {
		return errors.New("Provision requires env.S3_BUCKET to be defined")
	}
	return spartaCommand("provision", "--s3Bucket", bucketName)
}

// Describe deploys the given service
func Describe() error {
	// Get the bucketName
	bucketName := os.Getenv("S3_BUCKET")
	if bucketName == "" {
		return errors.New("Describe requires env.S3_BUCKET to be defined")
	}
	return spartaCommand("describe", "--s3Bucket", bucketName, "--out", "graph.html")
}

// Delete deletes the given service
func Delete() error {
	return spartaCommand("delete")
}

// Status returns a report for the given status
func Status() error {
	return spartaCommand("status", "--redact")
}

// Version returns version information about the service and embedded Sparta version
func Version() error {
	return spartaCommand("version")
}
