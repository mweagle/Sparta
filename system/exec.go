package system

import (
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
	defer outputWriter.Close()
	cmd.Stdout = outputWriter
	cmd.Stderr = outputWriter
	return cmd.Run()
}
