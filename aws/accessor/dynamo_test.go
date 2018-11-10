package accessor

import (
	"testing"
)

func TestDynamoPutObject(t *testing.T) {

	dynamoAccessor := &DynamoAccessor{
		testingTableName: "JSONDocuments",
	}
	testPut(t, dynamoAccessor)
}

func TestDynamoPutAllObject(t *testing.T) {
	dynamoAccessor := &DynamoAccessor{
		testingTableName: "JSONDocuments",
	}
	testPutAll(t, dynamoAccessor)
}
