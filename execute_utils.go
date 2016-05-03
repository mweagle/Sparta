package sparta

import (
	"encoding/json"
	"fmt"
	"github.com/mweagle/cloudformationresources"
	"net/http"

	"strings"

	"github.com/Sirupsen/logrus"
)

type dispatchMap map[string]*LambdaAWSInfo

// LambdaHTTPHandler is an HTTP compliant handler that implements
// ServeHTTP
type LambdaHTTPHandler struct {
	lambdaDispatchMap dispatchMap
	logger            *logrus.Logger
}

func (handler *LambdaHTTPHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Remove the leading slash and dispatch it to the golang handler
	lambdaFunc := strings.TrimLeft(req.URL.Path, "/")
	decoder := json.NewDecoder(req.Body)
	var request lambdaRequest
	defer func() {
		if r := recover(); r != nil {
			err, ok := r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
			errorString := fmt.Sprintf("Lambda handler panic: %#v", err)
			http.Error(w, errorString, http.StatusBadRequest)
		}
	}()

	err := decoder.Decode(&request)
	if nil != err {
		errorString := fmt.Sprintf("Failed to decode proxy request: %s", err.Error())
		http.Error(w, errorString, http.StatusBadRequest)
		return
	}
	handler.logger.WithFields(logrus.Fields{
		"Request":    request,
		"LookupName": lambdaFunc,
	}).Debug("Dispatching")

	lambdaAWSInfo := handler.lambdaDispatchMap[lambdaFunc]
	var lambdaFn LambdaFunction
	if nil != lambdaAWSInfo {
		lambdaFn = lambdaAWSInfo.lambdaFn
	} else if strings.Contains(lambdaFunc, "::") {
		// Not the most exhaustive guard, but the CloudFormation custom resources
		// all have "::" delimiters in their type field.  Even if there is a false
		// positive, the customResourceForwarder will simply error out.
		lambdaFn = customResourceForwarder
	}

	if nil == lambdaFn {
		http.Error(w, "Unsupported path: "+lambdaFunc, http.StatusBadRequest)
		return
	}
	lambdaFn(&request.Event, &request.Context, w, handler.logger)
}

// NewLambdaHTTPHandler returns an initialized LambdaHTTPHandler instance.  The returned value
// can be provided to https://golang.org/pkg/net/http/httptest/#NewServer to perform
// localhost testing.
func NewLambdaHTTPHandler(lambdaAWSInfos []*LambdaAWSInfo, logger *logrus.Logger) *LambdaHTTPHandler {
	lookupMap := make(dispatchMap, 0)
	for _, eachLambdaInfo := range lambdaAWSInfos {
		logger.WithFields(logrus.Fields{
			"Path": eachLambdaInfo.lambdaFnName,
		}).Debug("Registering lambda URL")

		lookupMap[eachLambdaInfo.lambdaFnName] = eachLambdaInfo
	}

	return &LambdaHTTPHandler{
		lambdaDispatchMap: lookupMap,
		logger:            logger,
	}
}
