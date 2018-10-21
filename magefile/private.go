// +build mage

package spartamage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
