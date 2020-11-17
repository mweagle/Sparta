package sparta

import (
	"fmt"
	"testing"
	"time"
)

func TestDelete(t *testing.T) {
	logger, _ := NewLogger("info")
	serviceName := fmt.Sprintf("ServiceTesting%d", time.Now().Unix())
	deleteErr := Delete(serviceName, logger)
	if deleteErr != nil {
		t.Fatalf("Failed to consider non-existent stack successfully deleted")
	}
}
