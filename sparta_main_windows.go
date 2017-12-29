// +build windows

package sparta

import (
	"fmt"
	"runtime"

	"github.com/Sirupsen/logrus"
)

func displayPrettyHeader(headerDivider string, logger *logrus.Logger) {
	logger.Info(headerDivider)
	logger.Info(fmt.Sprintf(`╔═╗┌─┐┌─┐┬─┐┌┬┐┌─┐   Version : %s`, SpartaVersion))
	logger.Info(fmt.Sprintf(`╚═╗├─┘├─┤├┬┘ │ ├─┤   SHA     : %s`, SpartaGitHash[0:7]))
	logger.Info(fmt.Sprintf(`╚═╝┴  ┴ ┴┴└─ ┴ ┴ ┴   Go      : %s`, runtime.Version()))
	logger.Info(headerDivider)
}
