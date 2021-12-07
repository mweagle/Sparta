package step

import (
	"math/rand"
)

/*
{
	"Type": "Task",
	"Resource":"arn:aws:states:::apigateway:invoke",
	"Parameters": {
			"ApiEndpoint": "example.execute-api.us-east-1.amazonaws.com",
			"Method": "GET",
			"Headers": {
					"key": ["value1", "value2"]
			},
			"Stage": "prod",
			"Path": "bills",
			"QueryParameters": {
					"billId": ["123456"]
			},
			"RequestBody": {},
			"AuthType": "NO_AUTH"
	}
}
*/

// APIGatewayTaskParameters represents params for the SNS notification
// Ref: https://docs.aws.amazon.com/sns/latest/api/API_Publish.html#API_Publish_RequestParameters
type APIGatewayTaskParameters struct {
	APIEndpoint     string                 `json:"ApiEndpoint,omitempty"`
	Method          string                 `json:",omitempty"`
	Headers         map[string]interface{} `json:",omitempty"`
	Stage           string                 `json:",omitempty"`
	Path            string                 `json:",omitempty"`
	QueryParameters map[string]interface{} `json:",omitempty"`
	RequestBody     string                 `json:",omitempty"`
	AuthType        string                 `json:",omitempty"`
}

// APIGatewayTaskState represents bindings for
// https://docs.aws.amazon.com/step-functions/latest/dg/connectors-sns.html
type APIGatewayTaskState struct {
	BaseTask
	parameters APIGatewayTaskParameters
}

// MarshalJSON for custom marshalling, since this will be stringified and we need it
// to turn into a stringified
// Ref: https://docs.aws.amazon.com/step-functions/latest/dg/connectors-sns.html
func (sts *APIGatewayTaskState) MarshalJSON() ([]byte, error) {
	return sts.BaseTask.marshalMergedParams("arn:aws:states:::apigateway:invoke",
		&sts.parameters)
}

// NewAPIGatewayTaskState returns an initialized APIGatewayTaskState
func NewAPIGatewayTaskState(stateName string,
	parameters APIGatewayTaskParameters) *APIGatewayTaskState {

	apigwTask := &APIGatewayTaskState{
		BaseTask: BaseTask{
			baseInnerState: baseInnerState{
				name: stateName,
				id:   rand.Int63(),
			},
		},
		parameters: parameters,
	}
	return apigwTask
}
