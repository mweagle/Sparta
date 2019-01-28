package docker

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"

	sparta "github.com/mweagle/Sparta"
	"github.com/mweagle/Sparta/system"
)

func TestLogin(t *testing.T) {
	logger, _ := sparta.NewLogger("info")

	// If docker -v doesn't work, then login definitely won't
	dockerVersionCmd := exec.Command("docker", "-v")
	dockerVersionCmdErr := system.RunOSCommand(dockerVersionCmd, logger)
	if dockerVersionCmdErr != nil {
		t.Logf("WARNING: failed to execute `docker -v` as prerequisite for testing STDIN password")
		return
	}
	// Don't supply password on command line

	dockerLoginCmd := exec.Command("docker",
		"login",
		"-u",
		"mweagle",
		"--password-stdin")
	dockerLoginCmd.Stdout = os.Stdout
	dockerLoginCmd.Stdin = bytes.NewReader([]byte("0AA421A3-931B-4985-8E99-9F5432A2BB58\n"))
	dockerLoginCmd.Stderr = os.Stderr
	dockerLoginCmdErr := system.RunOSCommand(dockerLoginCmd, logger)
	if dockerLoginCmdErr != nil {
		if strings.Contains(dockerLoginCmdErr.Error(), "Cannot perform an interactive login") {
			// The stdin write failed...
			t.Fatalf("Failed to write password to STDIN")
		} else if strings.Contains(dockerLoginCmdErr.Error(), "unauthorized: incorrect username or password") {
			t.Logf("Expected authorization rejection detected")
		}
	} else {
		// This should never happen
		t.Fatalf("Docker login with invalid credentials")
	}
}
