// Copyright (c) 2015 Matt Weagle <mweagle@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

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

const DEFAULT_HTTP_PORT = 9999

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
	lambdaAWSInfo.lambdaFn(reqeuest.Event, reqeuest.Context, w)
}

// Creates an HTTP listener to dispatch execution
func Execute(lambdaAWSInfos []*LambdaAWSInfo, port int, parentProcessPID int, logger *logrus.Logger) error {
	if port <= 0 {
		port = DEFAULT_HTTP_PORT
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
