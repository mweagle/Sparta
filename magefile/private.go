// +build mage

package magefile

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

var header = strings.Repeat("-", 80)

func sourceFilesOfType(extension string, ignoredSubdirectories []string) ([]string, error) {
	testExtension := strings.TrimPrefix(extension, ".")
	testExtension = fmt.Sprintf(".%s", testExtension)

	files := make([]string, 0)
	walker := func(path string, info os.FileInfo, err error) error {
		contains := false
		for _, eachComponent := range ignoredSubdirectories {
			contains = strings.Contains(path, eachComponent)
			if contains {
				break
			}
		}
		if !contains && (filepath.Ext(path) == testExtension) {
			files = append(files, path)
		}
		return nil
	}
	goSourceFilesErr := filepath.Walk(".", walker)
	return files, goSourceFilesErr
}

func spartaCommand(commandParts ...string) error {
	noopValue := ""
	parsedBool, _ := strconv.ParseBool(os.Getenv("NOOP"))
	if parsedBool {
		noopValue = "--noop"
	}
	curDir, curDirErr := os.Getwd()
	if curDirErr != nil {
		return errors.New("Failed to get current directory. Error: " + curDirErr.Error())
	}
	os.Setenv(mg.VerboseEnv, "1")
	commandArgs := []string{
		"run",
		curDir,
	}
	for _, eachCommandPart := range commandParts {
		commandArgs = append(commandArgs, eachCommandPart)
	}
	if noopValue != "" {
		commandArgs = append(commandArgs, "--noop")
	}
	return sh.Run("go",
		commandArgs...)
}
