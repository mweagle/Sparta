// +build lambdabinary

package cgo

// #include <stdlib.h>
// #include <string.h>
import "C"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/mweagle/Sparta"
	spartaAWS "github.com/mweagle/Sparta/aws"
)

// Lock to update CGO related config
var muCredentials sync.Mutex
var pythonCredentialsValue credentials.Value

////////////////////////////////////////////////////////////////////////////////
// spartaMockHTTPResponse is the buffered response to handle the HTTP
// response provided by the underlying function
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

////////////////////////////////////////////////////////////////////////////////
// lambdaFunctionErrResponse is the struct used to return a CGO error response
type lambdaFunctionErrResponse struct {
	Code    int         `json:"code"`
	Status  string      `json:"status"`
	Headers http.Header `json:"headers"`
	Error   string      `json:"error"`
}

////////////////////////////////////////////////////////////////////////////////
// cgoLambdaHTTPAdapterStruct is the binding between the various params
// supplied to the LambdaHandler
type cgoLambdaHTTPAdapterStruct struct {
	serviceName               string
	lambdaHTTPHandlerInstance *sparta.LambdaHTTPHandler
	logger                    *logrus.Logger
}

var cgoLambdaHTTPAdapter cgoLambdaHTTPAdapterStruct

////////////////////////////////////////////////////////////////////////////////
// cgoMain is the primary entrypoint for the library version
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
	cgoLambdaHTTPAdapter = cgoLambdaHTTPAdapterStruct{
		serviceName:               serviceName,
		lambdaHTTPHandlerInstance: sparta.NewLambdaHTTPHandler(lambdaAWSInfos, logger),
		logger: logger,
	}
	return nil
}

// LambdaHandler is the public handler that's called by the transformed
// CGO compliant userinput. Users should not need to call this function
// directly
func LambdaHandler(functionName string,
	eventJSON string,
	awsCredentials *credentials.Credentials) ([]byte, http.Header, error) {
	startTime := time.Now()
	readableBody := ioutil.NopCloser(strings.NewReader(eventJSON))

	// Update the credentials
	muCredentials.Lock()
	value, valueErr := awsCredentials.Get()
	if nil != valueErr {
		muCredentials.Unlock()
		return nil, nil, valueErr
	}
	pythonCredentialsValue.AccessKeyID = value.AccessKeyID
	pythonCredentialsValue.SecretAccessKey = value.SecretAccessKey
	pythonCredentialsValue.SessionToken = value.SessionToken
	pythonCredentialsValue.ProviderName = "PythonCGO"
	muCredentials.Unlock()

	// Update the credentials in the HTTP handler
	// in case we're ultimately forwarding to a custom
	// resource provider
	cgoLambdaHTTPAdapter.lambdaHTTPHandlerInstance.Credentials(pythonCredentialsValue)

	// We have to get these credentials into the HTTP server s.t. we can
	// update the session used for any CustomResource calls...

	// Make the request...
	response, header, err := makeRequest(functionName, readableBody, int64(len(eventJSON)))

	// TODO: Consider go routine
	postMetrics(awsCredentials, functionName, len(response), time.Since(startTime))
	return response, header, err
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
	cgoLambdaHTTPAdapter.lambdaHTTPHandlerInstance.ServeHTTP(spartaResp, spartaReq)

	// If there was an HTTP error, transform that into a stable
	// error payload and continue. This is the same format that's
	// used by the NodeJS proxying tier at /resources/index.js
	if spartaResp.statusCode >= 400 {
		errResponseBody := lambdaFunctionErrResponse{
			Code:    spartaResp.statusCode,
			Status:  http.StatusText(spartaResp.statusCode),
			Headers: spartaResp.Header(),
			Error:   spartaResp.bytes.String(),
		}

		// Replace the response with a new one
		jsonBytes, jsonBytesErr := json.Marshal(errResponseBody)
		if nil != jsonBytesErr {
			return nil, nil, jsonBytesErr
		} else {
			errResponse := newspartaMockHTTPResponse()
			errResponse.Write(jsonBytes)
			errResponse.Header().Set("content-length", strconv.Itoa(len(jsonBytes)))
			errResponse.Header().Set("content-type", "application/json")
			spartaResp = errResponse
		}
	}
	return spartaResp.bytes.Bytes(), spartaResp.headers, nil
}

func postMetrics(awsCredentials *credentials.Credentials,
	path string,
	responseBodyLength int,
	duration time.Duration) {

	awsCloudWatchService := cloudwatch.New(NewSession())
	metricNamespace := fmt.Sprintf("Sparta/%s", cgoLambdaHTTPAdapter.serviceName)
	lambdaFunctionName := os.Getenv("AWS_LAMBDA_FUNCTION_NAME")

	metricData := make([]*cloudwatch.MetricDatum, 0)
	sharedDimensions := make([]*cloudwatch.Dimension, 0)
	sharedDimensions = append(sharedDimensions,
		&cloudwatch.Dimension{
			Name:  aws.String("Path"),
			Value: aws.String(path),
		},
		&cloudwatch.Dimension{
			Name:  aws.String("Name"),
			Value: aws.String(lambdaFunctionName),
		})

	var sysinfo syscall.Sysinfo_t
	sysinfoErr := syscall.Sysinfo(&sysinfo)
	if nil == sysinfoErr {
		metricData = append(metricData, &cloudwatch.MetricDatum{
			MetricName: aws.String("Uptime"),
			Dimensions: sharedDimensions,
			Unit:       aws.String("Seconds"),
			Value:      aws.Float64(float64(sysinfo.Uptime)),
		})
	}
	metricData = append(metricData, &cloudwatch.MetricDatum{
		MetricName: aws.String("LambdaResponseLength"),
		Dimensions: sharedDimensions,
		Unit:       aws.String("Bytes"),
		Value:      aws.Float64(float64(responseBodyLength)),
	})
	params := &cloudwatch.PutMetricDataInput{
		MetricData: metricData,
		Namespace:  aws.String(metricNamespace),
	}
	awsCloudWatchService.PutMetricData(params)
}

// NewSession returns a CGO-aware AWS session that uses the Python
// credentials provided by the CGO interface
func NewSession() *session.Session {
	muCredentials.Lock()
	defer muCredentials.Unlock()

	awsConfig := aws.
		NewConfig().
		WithCredentials(credentials.NewStaticCredentialsFromCreds(pythonCredentialsValue))
	return spartaAWS.NewSessionWithConfig(awsConfig, cgoLambdaHTTPAdapter.logger)
}
