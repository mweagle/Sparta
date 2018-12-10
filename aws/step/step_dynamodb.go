package step

import (
	"math/rand"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	gocf "github.com/mweagle/go-cloudformation"
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
	Key                      map[string]*dynamodb.AttributeValue
	TableName                gocf.Stringable
	AttributesToGet          []string
	ConsistentRead           bool
	ExpressionAttributeNames map[string]string
	ProjectionExpression     string
	ReturnConsumedCapacity   string
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

	additionalParams := dgis.BaseTask.additionalParams()
	additionalParams["Resource"] = "arn:aws:states:::dynamodb:getItem"
	parameterMap := map[string]interface{}{}

	if dgis.parameters.Key != nil {
		parameterMap["Key"] = dgis.parameters.Key
	}
	if dgis.parameters.TableName != nil {
		parameterMap["TableName"] = dgis.parameters.TableName
	}
	if dgis.parameters.AttributesToGet != nil {
		parameterMap["AttributesToGet"] = dgis.parameters.AttributesToGet
	}
	if dgis.parameters.ConsistentRead {
		parameterMap["ConsistentRead"] = dgis.parameters.ConsistentRead
	}
	if dgis.parameters.ExpressionAttributeNames != nil {
		parameterMap["ExpressionAttributeNames"] = dgis.parameters.ExpressionAttributeNames
	}
	if dgis.parameters.ProjectionExpression != "" {
		parameterMap["ProjectionExpression"] = dgis.parameters.ProjectionExpression
	}
	if dgis.parameters.ReturnConsumedCapacity != "" {
		parameterMap["ReturnConsumedCapacity"] = dgis.parameters.ReturnConsumedCapacity
	}
	additionalParams["Parameters"] = parameterMap
	return dgis.marshalStateJSON("Task", additionalParams)
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
	TableName                   gocf.Stringable
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
	additionalParams := dgis.BaseTask.additionalParams()

	additionalParams["Resource"] = "arn:aws:states:::dynamodb:putItem"
	parameterMap := map[string]interface{}{}
	if dgis.parameters.Item != nil {
		parameterMap["Item"] = dgis.parameters.Item
	}
	if dgis.parameters.TableName != nil {
		parameterMap["TableName"] = dgis.parameters.TableName
	}
	if dgis.parameters.ConditionExpression != "" {
		parameterMap["ConditionExpression"] = dgis.parameters.ConditionExpression
	}
	if dgis.parameters.ConsistentRead {
		parameterMap["ConsistentRead"] = dgis.parameters.ConsistentRead
	}
	if dgis.parameters.ExpressionAttributeNames != nil {
		parameterMap["ExpressionAttributeNames"] = dgis.parameters.ExpressionAttributeNames
	}
	if dgis.parameters.ExpressionAttributeValues != nil {
		parameterMap["ExpressionAttributeValues"] = dgis.parameters.ExpressionAttributeValues
	}
	if dgis.parameters.ReturnConsumedCapacity != "" {
		parameterMap["ReturnConsumedCapacity"] = dgis.parameters.ReturnConsumedCapacity
	}
	if dgis.parameters.ReturnItemCollectionMetrics != "" {
		parameterMap["ReturnItemCollectionMetrics"] = dgis.parameters.ReturnItemCollectionMetrics
	}
	if dgis.parameters.ReturnValues != "" {
		parameterMap["ReturnValues"] = dgis.parameters.ReturnValues
	}
	additionalParams["Parameters"] = parameterMap
	return dgis.marshalStateJSON("Task", additionalParams)
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
