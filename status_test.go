package sparta

import (
	"fmt"
	"testing"
	"time"
)

func TestStatus(t *testing.T) {
	logger, _ := NewLogger("info")
	serviceName := fmt.Sprintf("ServiceTesting%d", time.Now().Unix())
	statusErr := Status(serviceName, "Test desc", false, logger)
	if statusErr != nil {
		t.Fatalf("Failed to error for non-existent stack")
	}
}
