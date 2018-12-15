package accessor

// Simple dynamo accessor to get put range over items...

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	sparta "github.com/mweagle/Sparta"
	spartaAWS "github.com/mweagle/Sparta/aws"
	"github.com/sirupsen/logrus"
)

const (
	attrID    = "id"
	attrValue = "value"
)

// DynamoAccessor to make it a bit easier to work with Dynamo
// as the backing store
type DynamoAccessor struct {
	testingTableName        string
	DynamoTableResourceName string
}

func (svc *DynamoAccessor) dynamoSvc(ctx context.Context) *dynamodb.DynamoDB {
	logger, _ := ctx.Value(sparta.ContextKeyLogger).(*logrus.Logger)
	sess := spartaAWS.NewSession(logger)
	dynamoClient := dynamodb.New(sess)
	xrayInit(dynamoClient.Client)
	return dynamoClient
}

func (svc *DynamoAccessor) dynamoTableName() string {
	if svc.testingTableName != "" {
		return svc.testingTableName
	}
	discover, discoveryInfoErr := sparta.Discover()
	if discoveryInfoErr != nil {
		return ""
	}
	dynamoTableRes, dynamoTableResExists := discover.Resources[svc.DynamoTableResourceName]
	if !dynamoTableResExists {
		return ""
	}
	return dynamoTableRes.ResourceRef
}

func dynamoKeyValueAttrMap(keyPath string) map[string]*dynamodb.AttributeValue {
	return map[string]*dynamodb.AttributeValue{
		attrID: {
			S: aws.String(keyPath),
		}}
}

// Delete handles deleting the resource
func (svc *DynamoAccessor) Delete(ctx context.Context, keyPath string) error {
	deleteItemInput := &dynamodb.DeleteItemInput{
		TableName: aws.String(svc.dynamoTableName()),
		Key:       dynamoKeyValueAttrMap(keyPath),
	}
	_, deleteResultErr := svc.
		dynamoSvc(ctx).
		DeleteItemWithContext(ctx, deleteItemInput)
	return deleteResultErr
}

// DeleteAll handles deleting all the items
func (svc *DynamoAccessor) DeleteAll(ctx context.Context) error {
	var deleteErr error
	input := &dynamodb.ScanInput{
		TableName: aws.String(svc.dynamoTableName()),
	}

	scanHandler := func(output *dynamodb.ScanOutput, lastPage bool) bool {
		writeDeleteRequests := make([]*dynamodb.WriteRequest, len(output.Items))
		for index, eachItem := range output.Items {
			keyID := ""
			stringVal, stringValOk := eachItem[attrID]
			if stringValOk && stringVal.S != nil {
				keyID = *(stringVal.S)
			}
			writeDeleteRequests[index] = &dynamodb.WriteRequest{
				DeleteRequest: &dynamodb.DeleteRequest{
					Key: dynamoKeyValueAttrMap(keyID),
				},
			}
		}
		input := &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]*dynamodb.WriteRequest{
				svc.dynamoTableName(): writeDeleteRequests,
			},
		}
		_, deleteErr = svc.dynamoSvc(ctx).BatchWriteItem(input)
		return deleteErr == nil
	}

	scanErr := svc.dynamoSvc(ctx).ScanPagesWithContext(ctx, input, scanHandler)
	if scanErr != nil {
		return scanErr
	}
	return deleteErr
}

// Put handles saving the item
func (svc *DynamoAccessor) Put(ctx context.Context, keyPath string, object interface{}) error {

	// What's the type of the object?
	if object == nil {
		return errors.New("DynamoAccessor Put object must not be nil")
	}
	// Map it...
	marshal, marshalErr := dynamodbattribute.MarshalMap(object)
	if marshalErr != nil {
		return marshalErr
	}
	// TODO - consider using tags for this...
	_, idExists := marshal[attrID]
	if !idExists {
		marshal[attrID] = &dynamodb.AttributeValue{
			S: aws.String(keyPath),
		}
	}
	putItemInput := &dynamodb.PutItemInput{
		TableName: aws.String(svc.dynamoTableName()),
		Item:      marshal,
	}
	_, putItemErr := svc.dynamoSvc(ctx).PutItemWithContext(ctx, putItemInput)
	return putItemErr
}

// Get handles getting the item
func (svc *DynamoAccessor) Get(ctx context.Context,
	keyPath string,
	destObject interface{}) error {
	getItemInput := &dynamodb.GetItemInput{
		TableName: aws.String(svc.dynamoTableName()),
		Key:       dynamoKeyValueAttrMap(keyPath),
	}

	getItemResult, getItemResultErr := svc.dynamoSvc(ctx).GetItemWithContext(ctx, getItemInput)
	if getItemResultErr != nil {
		return getItemResultErr
	}
	return dynamodbattribute.UnmarshalMap(getItemResult.Item, destObject)
}

// GetAll handles returning all of the items
func (svc *DynamoAccessor) GetAll(ctx context.Context,
	ctor NewObjectConstructor) ([]interface{}, error) {
	var getAllErr error

	results := make([]interface{}, 0)
	input := &dynamodb.ScanInput{
		TableName: aws.String(svc.dynamoTableName()),
	}
	scanHandler := func(output *dynamodb.ScanOutput, lastPage bool) bool {
		for _, eachItem := range output.Items {
			unmarshalTarget := ctor()
			unmarshalErr := dynamodbattribute.UnmarshalMap(eachItem, unmarshalTarget)
			if unmarshalErr != nil {
				getAllErr = unmarshalErr
				return false
			}
			results = append(results, unmarshalTarget)
		}
		return true
	}
	scanErr := svc.dynamoSvc(ctx).ScanPagesWithContext(ctx, input, scanHandler)
	if scanErr != nil {
		return nil, scanErr
	}
	if getAllErr != nil {
		return nil, getAllErr
	}
	return results, nil
}
