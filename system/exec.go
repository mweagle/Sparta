package system

import (
	"bufio"
	"bytes"
	"io"
	"os/exec"
	"strings"

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

	// Write the command to a buffer, split the lines, log them...
	var commandStdOutput bytes.Buffer
	var commandStdErr bytes.Buffer

	cmd.Stdout = &commandStdOutput
	cmd.Stderr = &commandStdErr
	cmdErr := cmd.Run()

	// Output each one...
	scannerStdout := bufio.NewScanner(&commandStdOutput)
	stdoutLogger := logger.With().
		Str("io", "stdout").
		Logger()
	for scannerStdout.Scan() {
		text := strings.TrimSpace(scannerStdout.Text())
		if len(text) != 0 {
			stdoutLogger.Info().Msg(text)
		}
	}
	scannerStderr := bufio.NewScanner(&commandStdErr)
	stderrLogger := logger.With().
		Str("io", "stderr").
		Logger()
	for scannerStderr.Scan() {
		text := strings.TrimSpace(scannerStderr.Text())
		if len(text) != 0 {
			stderrLogger.Info().Msg(text)
		}
	}
	return cmdErr
}
