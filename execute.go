package sparta

import (
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"net/http"
	"strings"
	"syscall"
	"time"
)

// Port used for HTTP proxying communication
const default_HTTP_Port = 9999

type dispatchMap map[string]*LambdaAWSInfo

type lambdaHandler struct {
	lambdaDispatchMap dispatchMap
	logger            *logrus.Logger
}

func (handler *lambdaHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Remove the leading slash and dispatch it to the golang handler
	lambdaFunc := strings.TrimLeft(req.URL.Path, "/")
	decoder := json.NewDecoder(req.Body)
	var reqeuest lambdaRequest
	err := decoder.Decode(&reqeuest)
	if nil != err {
		http.Error(w, "Failed to decode request", http.StatusBadRequest)
		return
	}
	lambdaAWSInfo := handler.lambdaDispatchMap[lambdaFunc]
	if nil == lambdaAWSInfo {
		http.Error(w, "Unsupported path: "+lambdaFunc, http.StatusBadRequest)
		return
	}
	lambdaAWSInfo.lambdaFn(&reqeuest.Event, &reqeuest.Context, &w, handler.logger)
}

// Creates an HTTP listener to dispatch execution
func Execute(lambdaAWSInfos []*LambdaAWSInfo, port int, parentProcessPID int, logger *logrus.Logger) error {
	if port <= 0 {
		port = default_HTTP_Port
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
