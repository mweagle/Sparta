package system

import (
	"io"
	"os/exec"

	"github.com/rs/zerolog"
)

type zerologWriter struct {
	targetLog *zerolog.Logger
}

func (zw *zerologWriter) Write(p []byte) (n int, err error) {
	zw.targetLog.Info().Msg(string(p))
	return len(p), nil
}

// RunOSCommand properly executes a system command
// and writes the output to the provided logger
func RunOSCommand(cmd *exec.Cmd, logger *zerolog.Logger) error {
	logger.Debug().
		Interface("Arguments", cmd.Args).
		Str("Dir", cmd.Dir).
		Str("Path", cmd.Path).
		Interface("Env", cmd.Env).
		Msg("Running Command")

	outputWriter := &zerologWriter{targetLog: logger}
	cmdErr := RunAndCaptureOSCommand(cmd,
		outputWriter,
		outputWriter,
		logger)
	return cmdErr
}

// RunAndCaptureOSCommand runs the given command and
// captures the stdout and stderr
func RunAndCaptureOSCommand(cmd *exec.Cmd,
	stdoutWriter io.Writer,
	stderrWriter io.Writer,
	logger *zerolog.Logger) error {
	logger.Debug().
		Interface("Arguments", cmd.Args).
		Str("Dir", cmd.Dir).
		Str("Path", cmd.Path).
		Interface("Env", cmd.Env).
		Msg("Running Command")
	cmd.Stdout = stdoutWriter
	cmd.Stderr = stderrWriter
	return cmd.Run()

}
