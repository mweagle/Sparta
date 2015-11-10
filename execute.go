package sparta

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"syscall"
	"time"

	"github.com/Sirupsen/logrus"
)

// Port used for HTTP proxying communication
const defaultHTTPPort = 9999

type dispatchMap map[string]*LambdaAWSInfo

type lambdaHandler struct {
	lambdaDispatchMap dispatchMap
	logger            *logrus.Logger
}

func (handler *lambdaHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Remove the leading slash and dispatch it to the golang handler
	lambdaFunc := strings.TrimLeft(req.URL.Path, "/")
	decoder := json.NewDecoder(req.Body)
	var request lambdaRequest
	err := decoder.Decode(&request)
	if nil != err {
		http.Error(w, "Failed to decode request", http.StatusBadRequest)
		return
	}
	handler.logger.WithFields(logrus.Fields{
		"Request": request,
	}).Debug("Dispatching")

	lambdaAWSInfo := handler.lambdaDispatchMap[lambdaFunc]
	if nil == lambdaAWSInfo {
		http.Error(w, "Unsupported path: "+lambdaFunc, http.StatusBadRequest)
		return
	}
	lambdaAWSInfo.lambdaFn(&request.Event, &request.Context, &w, handler.logger)
}

// Execute creates an HTTP listener to dispatch execution. Typically
// called via Main() via command line arguments.
func Execute(lambdaAWSInfos []*LambdaAWSInfo, port int, parentProcessPID int, logger *logrus.Logger) error {
	if port <= 0 {
		port = defaultHTTPPort
	}
	logger.Info("Execute!")

	lookupMap := make(dispatchMap, 0)
	for _, eachLambdaInfo := range lambdaAWSInfos {
		lookupMap[eachLambdaInfo.lambdaFnName] = eachLambdaInfo
	}
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      &lambdaHandler{lookupMap, logger},
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	if 0 != parentProcessPID {
		logger.Debug("Sending SIGUSR2 to parent process: ", parentProcessPID)
		syscall.Kill(parentProcessPID, syscall.SIGUSR2)
	}
	logger.Debug("Binding to port: ", port)
	err := server.ListenAndServe()
	if err != nil {
		logger.Error("FAILURE: " + err.Error())
		return err
	}
	logger.Debug("Server available at: ", port)
	return nil
}
