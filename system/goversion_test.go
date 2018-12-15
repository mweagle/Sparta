package system

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestGoVersion(t *testing.T) {
	logger := logrus.New()
	goVersion, goVersionError := GoVersion(logger)
	if goVersionError != nil {
		t.Fatalf("Failed to get go version: %s", goVersionError.Error())
	}
	t.Logf("Go version: %s", goVersion)
}
func TestGoPath(t *testing.T) {
	goPath := GoPath()
	// There should be a `go` binary in there
	goBinPath := filepath.Join(goPath, "bin")
	_, statErr := os.Stat(goBinPath)
	if statErr != nil && os.IsNotExist(statErr) {
		t.Fatalf("Failed to find `GOPATH` at: %s. Error: %s", goBinPath, statErr)
	}
}
