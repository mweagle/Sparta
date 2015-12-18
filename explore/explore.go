package explore

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// NewRequest sends a mock request to a localhost server that
// was created by httptest.NewServer(NewLambdaHTTPHandler(lambdaFunctions, logger)).
// lambdaName is the lambdaFnName to be called, eventData is optional event-specific
// data, and the testingURL is the URL returned by httptest.NewServer().
func NewRequest(lambdaName string, eventData interface{}, testingURL string) (*http.Response, error) {
	nowTime := time.Now()

	context := map[string]interface{}{
		"AWSRequestID":       "12341234-1234-1234-1234-123412341234",
		"InvokeID":           fmt.Sprintf("%d-12341234-1234-1234-1234-123412341234", nowTime.Unix()),
		"LogGroupName":       "/aws/lambda/SpartaApplicationMockLogGroup-9ZX7FITHEAG8",
		"LogStreamName":      fmt.Sprintf("%d/%d/%d/[$LATEST]%d", nowTime.Year(), nowTime.Month(), nowTime.Day(), nowTime.Unix()),
		"FunctionName":       "SpartaFunction",
		"MemoryLimitInMB":    "128",
		"FunctionVersion":    "[LATEST]",
		"InvokedFunctionARN": fmt.Sprintf("arn:aws:lambda:us-west-2:123412341234:function:SpartaMockFunction-%d", nowTime.Unix()),
	}

	requestBody := map[string]interface{}{
		"context": context,
	}
	if nil != eventData {
		requestBody["event"] = eventData
	}
	// Generate a complete CloudFormation template
	jsonRequestBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("Failed to Marshal request body: ", err.Error())
	}
	fmt.Printf("Sending POST: %s", string(jsonRequestBody))

	// POST IT...
	var host = fmt.Sprintf("%s/%s", testingURL, lambdaName)
	req, err := http.NewRequest("POST", host, strings.NewReader(string(jsonRequestBody)))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	return resp, nil
}
