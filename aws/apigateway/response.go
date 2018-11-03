package apigateway

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// Error represents an error to return in response to an
// API Gateway request
type Error struct {
	Code    int                    `json:"code"`
	Err     string                 `json:"err"`
	Message string                 `json:"message"`
	Context map[string]interface{} `json:"context,omitempty"`
}

// Error returns the JSONified version of this error which will
// trigger the appropriate integration mapping.
func (apigError *Error) Error() string {
	bytes, bytesErr := json.Marshal(apigError)
	if bytesErr != nil {
		bytes = []byte(http.StatusText(http.StatusInternalServerError))
	}
	return string(bytes)
}

// NewErrorResponse returns a response that satisfies
// the regular expression used to determine integration mappings
// via the API Gateway. messages is a stringable type. Error interface
// instances will be properly typecast.
func NewErrorResponse(statusCode int, messages ...interface{}) *Error {

	additionalMessages := make([]string, len(messages), len(messages))
	for eachIndex, eachMessage := range messages {
		switch typedValue := eachMessage.(type) {
		case error:
			additionalMessages[eachIndex] = typedValue.Error()
		default:
			additionalMessages[eachIndex] = fmt.Sprintf("%v", typedValue)
		}
	}

	err := &Error{
		Code:    statusCode,
		Err:     http.StatusText(statusCode),
		Message: strings.Join(additionalMessages, " "),
		Context: make(map[string]interface{}),
	}
	if len(err.Err) <= 0 {
		err.Code = http.StatusInternalServerError
		err.Err = http.StatusText(http.StatusInternalServerError)
	}
	return err
}

// Response is the type returned by an API Gateway function
type Response struct {
	Code    int               `json:"code,omitempty"`
	Body    interface{}       `json:"body,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

// canonicalResponse is the type for the canonicalized response with the
// headers lowercased to match any API-Gateway case sensitive whitelist
// matching
type canonicalResponse struct {
	Code    int               `json:"code,omitempty"`
	Body    interface{}       `json:"body,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

// MarshalJSON is a custom marshaller to ensure that the marshalled
// headers are always lowercase
func (resp *Response) MarshalJSON() ([]byte, error) {
	canonicalResponse := canonicalResponse{
		Code: resp.Code,
		Body: resp.Body,
	}
	if len(resp.Headers) != 0 {
		canonicalResponse.Headers = make(map[string]string)
		for eachKey, eachValue := range resp.Headers {
			canonicalResponse.Headers[strings.ToLower(eachKey)] = eachValue
		}
	}
	return json.Marshal(&canonicalResponse)
}

// NewResponse returns an API Gateway response object
func NewResponse(code int, body interface{}, headers ...map[string]string) *Response {
	response := &Response{
		Code: code,
		Body: body,
	}
	if len(headers) != 0 {
		response.Headers = make(map[string]string)
		for _, eachHeaderMap := range headers {
			for eachKey, eachValue := range eachHeaderMap {
				response.Headers[eachKey] = eachValue
			}
		}
	}
	return response
}
