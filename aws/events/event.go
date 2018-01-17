package events

import (
	"fmt"
	"os"
	"strings"
)

// APIGatewayIdentity is the API Gateway identity information
type APIGatewayIdentity struct {
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

// APIGatewayContext is the API-Gateway context information
type APIGatewayContext struct {
	AppID        string             `json:"appId"`
	Method       string             `json:"method"`
	RequestID    string             `json:"requestId"`
	ResourceID   string             `json:"resourceId"`
	ResourcePath string             `json:"resourcePath"`
	Stage        string             `json:"stage"`
	Identity     APIGatewayIdentity `json:"identity"`
}

// APIGatewayRequest represents the API Gateway request that
// is submitted to a Lambda function. This format matches the
// inputmapping_default.VTL templates
type APIGatewayRequest struct {
	Method      string            `json:"method"`
	Body        interface{}       `json:"body"`
	Headers     map[string]string `json:"headers"`
	QueryParams map[string]string `json:"queryParams"`
	PathParams  map[string]string `json:"pathParams"`
	Context     APIGatewayContext `json:"context"`
}

// NewAPIGatewayMockRequest creates a mock API Gateway request.
// This request format mirrors the VTL templates in
// github.com/mweagle/Sparta/resources/provision/apigateway
func NewAPIGatewayMockRequest(lambdaName string,
	httpMethod string,
	whitelistParamValues map[string]string,
	eventData interface{},
	testingURL string) (*APIGatewayRequest, error) {

	apiGatewayRequest := &APIGatewayRequest{
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
			apiGatewayRequest.Headers[keyName] = eachWhitelistValue
		case "querystring":
			apiGatewayRequest.QueryParams[keyName] = eachWhitelistValue
		case "path":
			apiGatewayRequest.PathParams[keyName] = eachWhitelistValue
		default:
			return nil, fmt.Errorf("Unsupported whitelist param type: %s", keyType)
		}
	}

	apiGatewayRequest.Context.AppID = fmt.Sprintf("spartaApp%d", os.Getpid())
	apiGatewayRequest.Context.Method = httpMethod
	apiGatewayRequest.Context.RequestID = "12341234-1234-1234-1234-123412341234"
	apiGatewayRequest.Context.ResourceID = "anon42"
	apiGatewayRequest.Context.ResourcePath = "/mock"
	apiGatewayRequest.Context.Stage = "mock"
	apiGatewayRequest.Context.Identity = APIGatewayIdentity{
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
	return apiGatewayRequest, nil
}
