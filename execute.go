package sparta

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
)

// Port used for HTTP proxying communication
const defaultHTTPPort = 9999

// Execute creates an HTTP listener to dispatch execution. Typically
// called via Main() via command line arguments.
func Execute(lambdaAWSInfos []*LambdaAWSInfo, port int, parentProcessPID int, logger *logrus.Logger) error {
	if port <= 0 {
		port = defaultHTTPPort
	}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      NewLambdaHTTPHandler(lambdaAWSInfos, logger),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	if 0 != parentProcessPID {
		platformKill(parentProcessPID)
	}
	logger.WithFields(logrus.Fields{
		"URL": fmt.Sprintf("http://localhost:%d", port),
	}).Info("Starting Sparta server")

	err := server.ListenAndServe()
	if err != nil {
		logger.WithFields(logrus.Fields{
			"Error": err.Error(),
		}).Error("Failed to launch server")
		return err
	}

	return nil
}
