package accessor

import (
	"context"
	"math/rand"
	"os"
	"testing"
	"time"

	sparta "github.com/mweagle/Sparta/v3"
	"github.com/rs/zerolog"
)

type testObject struct {
	Data          string
	SomeOtherData []string
	Value         int
}

func newTestObject() *testObject {
	return &testObject{
		Data:          "Hello World  at " + time.Now().UTC().String(),
		SomeOtherData: []string{"Val1", "Val2", "Val3"},
		Value:         42,
	}
}

// Disable test in Travis.
// Ref: https://help.github.com/en/actions/automating-your-workflow-with-github-actions/using-environment-variables
//
func testDisabled() bool {
	return os.Getenv("GITHUB_WORKFLOW") != ""
}
func testObjectConstructor() interface{} {
	return &testObject{}
}

func testPut(t *testing.T, kvStore KevValueAccessor) {
	if testDisabled() {
		return
	}
	logger, _ := sparta.NewLogger(zerolog.InfoLevel.String())

	ctx := context.Background()
	ctx = context.WithValue(ctx, sparta.ContextKeyLogger, logger)

	testID := time.Now().UTC().String()
	record := newTestObject()
	err := kvStore.Put(ctx, testID, record)
	if err != nil {
		t.Fatalf("%T failed to put item: %s", kvStore, err.Error())
	}
	var emptyRecord testObject
	err = kvStore.Get(ctx, testID, &emptyRecord)
	if err != nil {
		t.Fatalf("%T failed to get item: %s", kvStore, err.Error())
	}
	err = kvStore.Delete(ctx, testID)
	if err != nil {
		t.Fatalf("%T failed to delete item: %s", kvStore, err.Error())
	}
}

func testPutAll(t *testing.T, kvStore KevValueAccessor) {
	if testDisabled() {
		return
	}
	logger, _ := sparta.NewLogger(zerolog.DebugLevel.String())
	ctx := context.Background()
	ctx = context.WithValue(ctx, sparta.ContextKeyLogger, logger)
	recordCount := int(rand.Int31n(2) + 1)

	for i := 0; i != recordCount; i++ {
		testID := time.Now().UTC().String()
		record := newTestObject()
		err := kvStore.Put(ctx, testID, record)
		if err != nil {
			t.Fatalf("%T failed to put item: %s", kvStore, err)
		}
		time.Sleep(50 * time.Millisecond)
	}
	getAll, getAllErr := kvStore.GetAll(ctx, testObjectConstructor)
	if getAllErr != nil {
		t.Fatalf("%T failed to get all items: %s", kvStore, getAllErr)
	}
	if len(getAll) != recordCount {
		t.Fatalf("%T returned item count %d doesn't match expected %d",
			kvStore,
			len(getAll),
			recordCount)
	}
	deleteAllErr := kvStore.DeleteAll(ctx)
	if deleteAllErr != nil {
		t.Fatalf("%T failed to delete all items: %s", kvStore, deleteAllErr)
	}
	getAll, getAllErr = kvStore.GetAll(ctx, testObjectConstructor)
	if getAllErr != nil {
		t.Fatalf("%T failed to confirm get all items: %s", kvStore, getAllErr)
	}
	if len(getAll) != 0 {
		t.Fatalf("%T failed to confirm all items deleted: %s", kvStore, getAll)
	}
}
