package sparta

import (
	"fmt"
	"os"
	"strings"
)

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
