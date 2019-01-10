package system

import (
	"io"
	"os/exec"

	"github.com/sirupsen/logrus"
)

// RunOSCommand properly executes a system command
// and writes the output to the provided logger
func RunOSCommand(cmd *exec.Cmd, logger *logrus.Logger) error {
	logger.WithFields(logrus.Fields{
		"Arguments": cmd.Args,
		"Dir":       cmd.Dir,
		"Path":      cmd.Path,
		"Env":       cmd.Env,
	}).Debug("Running Command")
	outputWriter := logger.Writer()
	cmdErr := RunAndCaptureOSCommand(cmd,
		outputWriter,
		outputWriter,
		logger)
	closeErr := outputWriter.Close()
	if closeErr != nil {
		logger.WithField("closeError", closeErr).Warn("Failed to close OS command writer")
	}
	return cmdErr
}

// RunAndCaptureOSCommand runs the given command and
// captures the stdout and stderr
func RunAndCaptureOSCommand(cmd *exec.Cmd,
	stdoutWriter io.Writer,
	stderrWriter io.Writer,
	logger *logrus.Logger) error {
	logger.WithFields(logrus.Fields{
		"Arguments": cmd.Args,
		"Dir":       cmd.Dir,
		"Path":      cmd.Path,
		"Env":       cmd.Env,
	}).Debug("Running Command")
	cmd.Stdout = stdoutWriter
	cmd.Stderr = stderrWriter
	return cmd.Run()

}
