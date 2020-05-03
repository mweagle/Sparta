package spartamage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var header = strings.Repeat("-", 80)

func sourceFilesOfType(extension string, ignoreGlobs []string) ([]string, error) {
	testExtension := strings.TrimPrefix(extension, ".")
	testExtension = fmt.Sprintf(".%s", testExtension)

	files := make([]string, 0)
	walker := func(path string, info os.FileInfo, err error) error {
		rejectFile := false
		for _, eachComponent := range ignoreGlobs {
			matched, matchErr := filepath.Match(eachComponent, path)
			if matchErr != nil {
				return nil
			}
			if matched {
				rejectFile = true
				break
			}
		}
		if !rejectFile && (filepath.Ext(path) == testExtension) {
			files = append(files, path)
		}
		return nil
	}
	goSourceFilesErr := filepath.Walk(".", walker)
	return files, goSourceFilesErr
}
