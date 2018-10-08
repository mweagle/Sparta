package system

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"

	"github.com/sirupsen/logrus"
)

// GoVersion returns the configured go version for this system
func GoVersion(logger *logrus.Logger) (string, error) {
	runtimeVersion := runtime.Version()
	// Get the golang version from the output:
	// Matts-MBP:Sparta mweagle$ go version
	// go version go1.8.1 darwin/amd64
	golangVersionRE := regexp.MustCompile(`go(\d+\.\d+(\.\d+)?)`)
	matches := golangVersionRE.FindStringSubmatch(runtimeVersion)
	if len(matches) > 2 {
		return matches[1], nil
	}
	logger.WithFields(logrus.Fields{
		"Output": runtimeVersion,
	}).Warn("Unable to find Golang version using RegExp - using current version")
	return runtimeVersion, nil
}

// GoPath returns either $GOPATH or the new $HOME/go path
// introduced with Go 1.8
func GoPath() string {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		home := os.Getenv("HOME")
		gopath = filepath.Join(home, "go")
	}
	return gopath
}
