package system

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"os/exec"
	"strings"

	"github.com/rs/zerolog"
)

// RunOSCommand properly executes a system command
// and writes the output to the provided logger
func RunOSCommand(cmd *exec.Cmd, logger *zerolog.Logger) error {
	logger.Debug().
		Interface("Arguments", cmd.Args).
		Str("Dir", cmd.Dir).
		Str("Path", cmd.Path).
		Interface("Env", cmd.Env).
		Msg("Running Command")

	// NOP write
	cmdErr := RunAndCaptureOSCommand(cmd,
		ioutil.Discard,
		ioutil.Discard,
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

	teeStdOut := io.MultiWriter(&commandStdOutput, stdoutWriter)
	teeStdErr := io.MultiWriter(&commandStdErr, stderrWriter)

	cmd.Stdout = teeStdOut
	cmd.Stderr = teeStdErr
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
