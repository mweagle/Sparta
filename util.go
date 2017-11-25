package sparta

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Create a stable temporary filename in the current working
// directory
func temporaryFile(name string) (*os.File, error) {
	workingDir, err := os.Getwd()
	if nil != err {
		return nil, err
	}

	// Use a stable temporary name
	temporaryPath := filepath.Join(workingDir, ScratchDirectory, name)
	buildDir := filepath.Dir(temporaryPath)
	mkdirErr := os.MkdirAll(buildDir, os.ModePerm)
	if nil != mkdirErr {
		return nil, mkdirErr
	}

	tmpFile, err := os.Create(temporaryPath)
	if err != nil {
		return nil, errors.New("Failed to create temporary file: " + err.Error())
	}
	return tmpFile, nil
}

// relativePath returns the relative path of logPath if it's relative to the current
// workint directory
func relativePath(logPath string) string {
	cwd, cwdErr := os.Getwd()
	if cwdErr == nil {
		relPath := strings.TrimPrefix(logPath, cwd)
		if relPath != logPath {
			logPath = fmt.Sprintf(".%s", relPath)
		}
	}
	return logPath
}
