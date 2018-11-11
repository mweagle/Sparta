package accessor

import (
	"testing"
)

func dynamoAccessor() KevValueAccessor {
	return &DynamoAccessor{
		testingTableName: "JSONDocuments",
	}
}

func TestDynamoPutObject(t *testing.T) {
	testPut(t, dynamoAccessor())
}

func TestDynamoPutAllObject(t *testing.T) {
	testPutAll(t, dynamoAccessor())
}
