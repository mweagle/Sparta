package apigateway

import (
	"encoding/json"
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

// NewError returns an Error satisfying
// object that is compatible with the API Gateway integration
// mappings
func NewError(statusCode int, messages ...string) *Error {
	err := &Error{
		Code:    statusCode,
		Err:     http.StatusText(statusCode),
		Message: strings.Join(messages, " "),
		Context: make(map[string]interface{}),
	}
	if len(err.Err) <= 0 {
		err.Code = http.StatusInternalServerError
		err.Err = http.StatusText(http.StatusInternalServerError)
	}
	return err
}
