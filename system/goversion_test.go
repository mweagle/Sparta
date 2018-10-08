package system

import (
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
