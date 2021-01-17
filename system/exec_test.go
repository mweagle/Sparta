package system

import (
	"os"
	"os/exec"
	"runtime"
	"testing"

	"github.com/rs/zerolog"
)

func TestRunCommand(t *testing.T) {
	commandName := ""
	switch runtime.GOOS {
	case "windows":
		commandName = "ipconfig"
	default:
		commandName = "date"
	}
	cmd := exec.Command(commandName)
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	runErr := RunOSCommand(cmd, &logger)
	if runErr != nil {
		t.Fatalf("Failed to run command `%s` (OS: %s). Error: %s",
			commandName,
			runtime.GOOS,
			runErr)
	}
}
