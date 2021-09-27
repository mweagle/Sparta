package sparta

import (
	"testing"

	gof "github.com/awslabs/goformation/v5/cloudformation"
)

func TestSafeMetadataInsert(t *testing.T) {
	emptyResource := &gof.CustomResource{}

	errSafe := safeMetadataInsert(emptyResource, "Woot", "Bar")
	if errSafe != nil {
		t.Fatalf("Failed to set Metadata: %s", errSafe)
	}
}
