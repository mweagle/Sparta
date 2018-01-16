package explore

import (
	"fmt"
	"os"
	"strings"
)

type mockAPIGatewayIdentity struct {
	AccountID                     string `json:"accountId"`
	APIKey                        string `json:"apiKey"`
	Caller                        string `json:"caller"`
	CognitoAuthenticationProvider string `json:"cognitoAuthenticationProvider"`
	CognitoAuthenticationType     string `json:"cognitoAuthenticationType"`
	CognitoIdentityID             string `json:"cognitoIdentityId"`
	CognitoIdentityPoolID         string `json:"cognitoIdentityPoolId"`
	SourceIP                      string `json:"sourceIp"`
	User                          string `json:"user"`
	UserAgent                     string `json:"userAgent"`
	UserArn                       string `json:"userArn"`
}

type mockAPIGatewayContext struct {
	AppID        string                 `json:"appId"`
	Method       string                 `json:"method"`
	RequestID    string                 `json:"requestId"`
	ResourceID   string                 `json:"resourceId"`
	ResourcePath string                 `json:"resourcePath"`
	Stage        string                 `json:"stage"`
	Identity     mockAPIGatewayIdentity `json:"identity"`
}

// SpartaAPIGatewayRequest represents the API Gateway request that
// is submitted to a Lambda function. This format matches the
// inputmapping_default.VTL templates
type SpartaAPIGatewayRequest struct {
	Method      string                `json:"method"`
	Body        interface{}           `json:"body"`
	Headers     map[string]string     `json:"headers"`
	QueryParams map[string]string     `json:"queryParams"`
	PathParams  map[string]string     `json:"pathParams"`
	Context     mockAPIGatewayContext `json:"context"`
}

// NewAPIGatewayRequest sends a mock request to a localhost server that
// was created by httptest.NewServer(NewLambdaHTTPHandler(lambdaFunctions, logger)).
// lambdaName is the lambdaFnName to be called, eventData is optional event-specific
// data, and the testingURL is the URL returned by httptest.NewServer().  The optional event data is
// embedded in the Sparta input mapping templates.
func NewAPIGatewayRequest(lambdaName string,
	httpMethod string,
	whitelistParamValues map[string]string,
	eventData interface{},
	testingURL string) (*SpartaAPIGatewayRequest, error) {

	mockAPIGatewayRequest := &SpartaAPIGatewayRequest{
		Method:      httpMethod,
		Body:        eventData,
		Headers:     make(map[string]string, 0),
		QueryParams: make(map[string]string, 0),
		PathParams:  make(map[string]string, 0),
	}
	for eachWhitelistKey, eachWhitelistValue := range whitelistParamValues {
		// Whitelisted params include their
		// namespace as part of the whitelist expression:
		// method.request.querystring.keyName
		parts := strings.Split(eachWhitelistKey, ".")

		// The string should have 4 parts...
		if len(parts) != 4 {
			return nil, fmt.Errorf("Invalid whitelist param name: %s (MUST be: method.request.KEY_TYPE.KEY_NAME, ex: method.request.querystring.myQueryParam", eachWhitelistKey)
		}
		keyType := parts[2]
		keyName := parts[3]
		switch keyType {
		case "header":
			mockAPIGatewayRequest.Headers[keyName] = eachWhitelistValue
		case "querystring":
			mockAPIGatewayRequest.QueryParams[keyName] = eachWhitelistValue
		case "path":
			mockAPIGatewayRequest.PathParams[keyName] = eachWhitelistValue
		default:
			return nil, fmt.Errorf("Unsupported whitelist param type: %s", keyType)
		}
	}

	mockAPIGatewayRequest.Context.AppID = fmt.Sprintf("spartaApp%d", os.Getpid())
	mockAPIGatewayRequest.Context.Method = httpMethod
	mockAPIGatewayRequest.Context.RequestID = "12341234-1234-1234-1234-123412341234"
	mockAPIGatewayRequest.Context.ResourceID = "anon42"
	mockAPIGatewayRequest.Context.ResourcePath = "/mock"
	mockAPIGatewayRequest.Context.Stage = "mock"
	mockAPIGatewayRequest.Context.Identity = mockAPIGatewayIdentity{
		AccountID: "123412341234",
		APIKey:    "",
		Caller:    "",
		CognitoAuthenticationProvider: "",
		CognitoAuthenticationType:     "",
		CognitoIdentityID:             "",
		CognitoIdentityPoolID:         "",
		SourceIP:                      "127.0.0.1",
		User:                          "Unknown",
		UserAgent:                     "Mozilla/Gecko",
		UserArn:                       "",
	}
	return mockAPIGatewayRequest, nil
}
