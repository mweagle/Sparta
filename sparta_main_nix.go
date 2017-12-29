// +build !windows

package sparta

import (
	"fmt"
	"runtime"

	"github.com/Sirupsen/logrus"
)

func displayPrettyHeader(headerDivider string, logger *logrus.Logger) {
	logger.Info(headerDivider)
	logger.Info(fmt.Sprintf(`   _______  ___   ___  _________ `))
	logger.Info(fmt.Sprintf(`  / __/ _ \/ _ | / _ \/_  __/ _ |     Version : %s`, SpartaVersion))
	logger.Info(fmt.Sprintf(` _\ \/ ___/ __ |/ , _/ / / / __ |     SHA     : %s`, SpartaGitHash[0:7]))
	logger.Info(fmt.Sprintf(`/___/_/  /_/ |_/_/|_| /_/ /_/ |_|     Go      : %s`, runtime.Version()))
	logger.Info(headerDivider)
}
