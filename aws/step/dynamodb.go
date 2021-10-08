package step

import (
	"math/rand"

	awsv2DynamoTypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

////////////////////////////////////////////////////////////////////////////////
/*
   ___     _     ___ _
  / __|___| |_  |_ _| |_ ___ _ __
 | (_ / -_)  _|  | ||  _/ -_) '  \
  \___\___|\__| |___|\__\___|_|_|_|
*/
////////////////////////////////////////////////////////////////////////////////

// DynamoDBGetItemParameters represents params for the DynamoDBGetItem
// parameters. Ref: https://docs.aws.amazon.com/step-functions/latest/dg/connectors-ddb.html
type DynamoDBGetItemParameters struct {
	Key                      map[string]awsv2DynamoTypes.AttributeValue `json:",omitempty"`
	TableName                string                                     `json:",omitempty"`
	AttributesToGet          []string                                   `json:",omitempty"`
	ConsistentRead           bool                                       `json:",omitempty"`
	ExpressionAttributeNames map[string]string                          `json:",omitempty"`
	ProjectionExpression     string                                     `json:",omitempty"`
	ReturnConsumedCapacity   string                                     `json:",omitempty"`
}

// DynamoDBGetItemState represents bindings for
// https://docs.aws.amazon.com/step-functions/latest/dg/connectors-ddb.html
type DynamoDBGetItemState struct {
	BaseTask
	parameters DynamoDBGetItemParameters
}

// MarshalJSON for custom marshalling, since this will be stringified and we need it
// to turn into a stringified
// Ref: https://docs.aws.amazon.com/step-functions/latest/dg/connectors-sns.html
func (dgis *DynamoDBGetItemState) MarshalJSON() ([]byte, error) {
	return dgis.BaseTask.marshalMergedParams("arn:aws:states:::dynamodb:getItem",
		&dgis.parameters)
}

// NewDynamoDBGetItemState returns an initialized DynamoDB GetItem state
func NewDynamoDBGetItemState(stateName string,
	parameters DynamoDBGetItemParameters) *DynamoDBGetItemState {

	dgis := &DynamoDBGetItemState{
		BaseTask: BaseTask{
			baseInnerState: baseInnerState{
				name: stateName,
				id:   rand.Int63(),
			},
		},
		parameters: parameters,
	}
	return dgis
}

////////////////////////////////////////////////////////////////////////////////
/*
  ___      _     ___ _
 | _ \_  _| |_  |_ _| |_ ___ _ __
 |  _/ || |  _|  | ||  _/ -_) '  \
 |_|  \_,_|\__| |___|\__\___|_|_|_|

*/
////////////////////////////////////////////////////////////////////////////////

// DynamoDBPutItemParameters represents params for the SNS notification
// Ref: https://docs.aws.amazon.com/sns/latest/api/API_Publish.html#API_Publish_RequestParameters
type DynamoDBPutItemParameters struct {
	Item                        map[string]*dynamodb.AttributeValue
	TableName                   string
	ConditionExpression         string
	ConsistentRead              bool
	ExpressionAttributeNames    map[string]string
	ExpressionAttributeValues   map[string]*dynamodb.AttributeValue
	ReturnConsumedCapacity      string // INDEXES | TOTAL | NONE
	ReturnItemCollectionMetrics string // SIZE | NONE
	ReturnValues                string // NONE | ALL_OLD | UPDATED_OLD | ALL_NEW | UPDATED_NEW
}

// DynamoDBPutItemState represents bindings for
// https://docs.aws.amazon.com/step-functions/latest/dg/connectors-ddb.html
type DynamoDBPutItemState struct {
	BaseTask
	parameters DynamoDBPutItemParameters
}

// MarshalJSON for custom marshalling, since this will be stringified and we need it
// to turn into a stringified
// Ref: https://docs.aws.amazon.com/step-functions/latest/dg/connectors-sns.html
func (dgis *DynamoDBPutItemState) MarshalJSON() ([]byte, error) {
	return dgis.BaseTask.marshalMergedParams("arn:aws:states:::dynamodb:putItem",
		&dgis.parameters)
}

// NewDynamoDBPutItemState returns an initialized DynamoDB PutItem state
func NewDynamoDBPutItemState(stateName string,
	parameters DynamoDBPutItemParameters) *DynamoDBPutItemState {

	dpis := &DynamoDBPutItemState{
		BaseTask: BaseTask{
			baseInnerState: baseInnerState{
				name: stateName,
				id:   rand.Int63(),
			},
		},
		parameters: parameters,
	}
	return dpis
}
