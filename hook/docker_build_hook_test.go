package hook

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"
	"time"

	sparta "github.com/mweagle/Sparta"
	"github.com/mweagle/Sparta/system"
	"github.com/rs/zerolog"
)

func TestBuildUPXImage(t *testing.T) {
	repo := "mweagle/upxlocal"
	version := fmt.Sprintf("%d", time.Now().Unix())
	tempFile, tempFileErr := ioutil.TempFile("", "Dockerfile")
	if tempFileErr != nil {
		t.Error(tempFileErr)
	}
	defer os.Remove(tempFile.Name()) // clean up

	// Ok, write the UPX content to the file
	_, writeErr := io.WriteString(tempFile, UPXDockerFile)
	if writeErr != nil {
		t.Error(tempFileErr)
	}
	tempFile.Close()
	// Write it out
	workflowHooks := &sparta.WorkflowHooks{
		PreBuilds: []sparta.WorkflowHookHandler{
			BuildDockerImageHook(tempFile.Name(),
				".",
				map[string]string{
					repo: version,
				}),
		},
	}

	logger, loggerErr := sparta.NewLogger(zerolog.InfoLevel.String())
	if loggerErr != nil {
		t.Fatalf("Failed to create test logger: %s", loggerErr)
	}
	var templateWriter bytes.Buffer
	err := sparta.Build(true,
		"SampleProvision",
		"",
		nil,
		nil,
		nil,
		false,
		"testBuildID",
		"",
		"",
		"",
		"",
		&templateWriter,
		workflowHooks,
		logger)
	if err != nil {
		t.Fatalf("Failed to provision test stack with workflow hook: " + err.Error())
	}
	// So if this worked, we should be able to run the Docker image...
	dockerTagName := fmt.Sprintf("%s:%s", repo, version)
	dockerRunCmd := exec.Command("docker",
		"run",
		dockerTagName,
		"-V",
		"-L")
	dockerErr := system.RunOSCommand(dockerRunCmd, logger)

	// Try and delete it either way...
	dockerDeleteCommand := exec.Command("docker", "rmi", "-f", dockerTagName)
	system.RunOSCommand(dockerDeleteCommand, logger)
	if dockerErr != nil {
		t.Fatalf("Failed to run Docker command to verify UPX image: " + dockerErr.Error())
	}
}
