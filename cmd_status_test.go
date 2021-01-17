package sparta

import (
	"fmt"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func TestStatus(t *testing.T) {
	logger, _ := NewLogger(zerolog.InfoLevel.String())
	serviceName := fmt.Sprintf("ServiceTesting%d", time.Now().Unix())
	statusErr := Status(serviceName, "Test desc", false, logger)
	if statusErr != nil {
		t.Fatalf("Failed to error for non-existent stack")
	}
}
