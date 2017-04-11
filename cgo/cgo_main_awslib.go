// +build lambdabinary

package cgo

// #include <stdlib.h>
// #include <string.h>
import "C"

import (
	"bytes"
	"fmt"
	"github.com/mweagle/Sparta"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type spartaMockHTTPResponse struct {
	// Private vars
	statusCode int
	headers    http.Header
	bytes      bytes.Buffer
}

func (spartaResp *spartaMockHTTPResponse) Header() http.Header {
	return spartaResp.headers
}

func (spartaResp *spartaMockHTTPResponse) Write(data []byte) (int, error) {
	return spartaResp.bytes.Write(data)
}

func (spartaResp *spartaMockHTTPResponse) WriteHeader(statusCode int) {
	spartaResp.statusCode = statusCode
}

func newspartaMockHTTPResponse() *spartaMockHTTPResponse {
	resp := &spartaMockHTTPResponse{
		statusCode: 200,
		headers:    make(map[string][]string, 0),
	}
	return resp
}

var LambdaHTTPHandlerInstance *sparta.LambdaHTTPHandler

func cgoMain(callerFile string,
	serviceName string,
	serviceDescription string,
	lambdaAWSInfos []*sparta.LambdaAWSInfo,
	api *sparta.API,
	site *sparta.S3Site,
	workflowHooks *sparta.WorkflowHooks) error {

	logger, loggerErr := sparta.NewLogger("info")
	if nil != loggerErr {
		panic("Failed to initialize logger")
	}
	LambdaHTTPHandlerInstance = sparta.NewLambdaHTTPHandler(lambdaAWSInfos, logger)
	return nil
}

func LambdaHandler(functionName string,
	eventJSON string) ([]byte, http.Header, error) {
	readableBody := ioutil.NopCloser(strings.NewReader(eventJSON))
	return makeRequest(functionName, readableBody, int64(len(eventJSON)))
}

func makeRequest(functionName string,
	eventBody io.ReadCloser,
	eventBodySize int64) ([]byte, http.Header, error) {
	spartaResp := newspartaMockHTTPResponse()

	// Create an http.Request object with this data...
	spartaReq := &http.Request{
		Method: "POST",
		URL: &url.URL{
			Scheme: "http",
			Path:   fmt.Sprintf("/%s", functionName),
		},
		Proto:            "HTTP/1.1",
		ProtoMajor:       1,
		ProtoMinor:       1,
		Body:             eventBody,
		ContentLength:    eventBodySize,
		TransferEncoding: make([]string, 0),
		Host:             "localhost",
	}
	LambdaHTTPHandlerInstance.ServeHTTP(spartaResp, spartaReq)
	return spartaResp.bytes.Bytes(), spartaResp.headers, nil
}
