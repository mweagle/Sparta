// +build !lambdabinary

package sparta

import (
	"fmt"

	"github.com/Sirupsen/logrus"
)

// Explore supports interactive command line invocation of the previously
// provisioned Sparta service
func Explore(lambdaAWSInfos []*LambdaAWSInfo, port int, logger *logrus.Logger) error {
	if 0 == port {
		port = 9999
	}
	urlHost := fmt.Sprintf("http://localhost:%d", port)
	logger.Info("The following URLs are available for testing.")

	msgText := ""
	for _, eachLambdaInfo := range lambdaAWSInfos {
		functionPath := fmt.Sprintf("%s/%s", urlHost, eachLambdaInfo.lambdaFnName)
		logger.WithFields(logrus.Fields{
			"Path": functionPath,
		}).Info(eachLambdaInfo.lambdaFnName)

		if msgText == "" {
			msgText = fmt.Sprintf("\tcurl -vs -X POST -H \"Content-Type: application/json\" --data @testEvent.json %s", functionPath)
		}
	}
	logger.Info("Functions can be invoked via application/json over POST")
	logger.Info(msgText)
	logger.Info("Where @testEvent.json is a local file with top level `context` and `event` keys")
	// Start up the localhost server and publish the info
	return Execute(lambdaAWSInfos, port, 0, logger)
}
