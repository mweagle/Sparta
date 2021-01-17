package sparta

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func TestDelete(t *testing.T) {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	serviceName := fmt.Sprintf("ServiceTesting%d", time.Now().Unix())
	deleteErr := Delete(serviceName, &logger)
	if deleteErr != nil {
		t.Fatalf("Failed to consider non-existent stack successfully deleted")
	}
}
