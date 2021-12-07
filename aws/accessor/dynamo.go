package accessor

// Simple dynamo accessor to get put range over items...

import (
	"context"
	"errors"

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	awsv2DynamoAttributeValue "github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	awsv2Dynamo "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	awsv2DynamoTypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	sparta "github.com/mweagle/Sparta/v3"
	spartaAWS "github.com/mweagle/Sparta/v3/aws"
	"github.com/rs/zerolog"
)

const (
	attrID = "id"
)

// DynamoAccessor to make it a bit easier to work with Dynamo
// as the backing store
type DynamoAccessor struct {
	testingTableName        string
	DynamoTableResourceName string
}

func (svc *DynamoAccessor) dynamoSvc(ctx context.Context) *awsv2Dynamo.Client {
	logger, _ := ctx.Value(sparta.ContextKeyLogger).(*zerolog.Logger)
	awsConfig, awsConfigErr := spartaAWS.NewConfig(ctx, logger)
	if awsConfigErr != nil {
		return nil
	}
	xrayInit(&awsConfig)
	dynamoClient := awsv2Dynamo.NewFromConfig(awsConfig)
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

func dynamoKeyValueAttrMap(keyPath string) map[string]awsv2DynamoTypes.AttributeValue {
	return map[string]awsv2DynamoTypes.AttributeValue{
		attrID: &awsv2DynamoTypes.AttributeValueMemberS{
			Value: keyPath,
		},
	}
}

// Delete handles deleting the resource
func (svc *DynamoAccessor) Delete(ctx context.Context, keyPath string) error {
	deleteItemInput := &awsv2Dynamo.DeleteItemInput{
		TableName: awsv2.String(svc.dynamoTableName()),
		Key:       dynamoKeyValueAttrMap(keyPath),
	}
	_, deleteResultErr := svc.
		dynamoSvc(ctx).DeleteItem(ctx, deleteItemInput)
	return deleteResultErr
}

// DeleteAll handles deleting all the items
func (svc *DynamoAccessor) DeleteAll(ctx context.Context) error {
	var deleteErr error
	input := &awsv2Dynamo.ScanInput{
		TableName: awsv2.String(svc.dynamoTableName()),
	}

	scanHandler := func(output *awsv2Dynamo.ScanOutput, lastPage bool) bool {
		writeDeleteRequests := make([]awsv2DynamoTypes.WriteRequest, len(output.Items))
		for index, eachItem := range output.Items {
			keyID := ""
			stringVal, stringValOk := eachItem[attrID]
			if stringValOk {
				switch typedVal := stringVal.(type) {
				case *awsv2DynamoTypes.AttributeValueMemberS:
					keyID = typedVal.Value
				}
			}
			writeDeleteRequests[index] = awsv2DynamoTypes.WriteRequest{
				DeleteRequest: &awsv2DynamoTypes.DeleteRequest{
					Key: dynamoKeyValueAttrMap(keyID),
				},
			}
		}
		input := &awsv2Dynamo.BatchWriteItemInput{
			RequestItems: map[string][]awsv2DynamoTypes.WriteRequest{
				svc.dynamoTableName(): writeDeleteRequests,
			},
		}
		_, deleteErr = svc.dynamoSvc(ctx).BatchWriteItem(ctx, input)
		return deleteErr == nil
	}

	scanResponse, scanErr := svc.dynamoSvc(ctx).Scan(ctx, input)
	if scanErr != nil {
		return scanErr
	}
	scanHandler(scanResponse, true)
	return deleteErr
}

// Put handles saving the item
func (svc *DynamoAccessor) Put(ctx context.Context, keyPath string, object interface{}) error {

	// What's the type of the object?
	if object == nil {
		return errors.New("DynamoAccessor Put object must not be nil")
	}
	// Map it...
	marshal, marshalErr := awsv2DynamoAttributeValue.MarshalMap(object)
	if marshalErr != nil {
		return marshalErr
	}
	// // TODO - consider using tags for this...

	_, idExists := marshal[attrID]
	if !idExists {
		marshal[attrID] = &awsv2DynamoTypes.AttributeValueMemberS{
			Value: keyPath,
		}
	}
	putItemInput := &awsv2Dynamo.PutItemInput{
		TableName: awsv2.String(svc.dynamoTableName()),
		Item:      marshal,
	}
	_, putItemErr := svc.dynamoSvc(ctx).PutItem(ctx, putItemInput)
	return putItemErr
}

// Get handles getting the item
func (svc *DynamoAccessor) Get(ctx context.Context,
	keyPath string,
	destObject interface{}) error {
	getItemInput := &awsv2Dynamo.GetItemInput{
		TableName: awsv2.String(svc.dynamoTableName()),
		Key:       dynamoKeyValueAttrMap(keyPath),
	}

	getItemResult, getItemResultErr := svc.dynamoSvc(ctx).GetItem(ctx, getItemInput)
	if getItemResultErr != nil {
		return getItemResultErr
	}
	return awsv2DynamoAttributeValue.UnmarshalMap(getItemResult.Item, destObject)
}

// GetAll handles returning all of the items
func (svc *DynamoAccessor) GetAll(ctx context.Context,
	ctor NewObjectConstructor) ([]interface{}, error) {
	var getAllErr error

	results := make([]interface{}, 0)
	input := &awsv2Dynamo.ScanInput{
		TableName: awsv2.String(svc.dynamoTableName()),
	}
	scanHandler := func(output *awsv2Dynamo.ScanOutput, lastPage bool) bool {
		for _, eachItem := range output.Items {
			unmarshalTarget := ctor()
			unmarshalErr := awsv2DynamoAttributeValue.UnmarshalMap(eachItem, unmarshalTarget)
			if unmarshalErr != nil {
				getAllErr = unmarshalErr
				return false
			}
			results = append(results, unmarshalTarget)
		}
		return true
	}
	scanResult, scanErr := svc.dynamoSvc(ctx).Scan(ctx, input)
	if scanErr != nil {
		return nil, scanErr
	}
	scanHandler(scanResult, false)
	if getAllErr != nil {
		return nil, getAllErr
	}
	return results, nil
}
