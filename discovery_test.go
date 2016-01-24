package sparta

import "testing"

func TestDiscoveryInitialized(t *testing.T) {
	// Ensure that sparta.Discover() can only be called from a lambda function
	logger, _ := NewLogger("warning")
	initializeDiscovery("DiscoveryTest", testLambdaData(), logger)

	configuration, err := Discover()
	t.Logf("Configuration: %#v", configuration)
	t.Logf("Error: %#v", err)
	if err == nil {
		t.Errorf("sparta.Discover() failed to reject invalid call site")
	}
	t.Logf("Properly rejected invalid callsite: %s", err.Error())
}

func TestDiscoveryNotInitialized(t *testing.T) {
	// Ensure that sparta.Discover() can only be called from a lambda function
	configuration, err := Discover()
	t.Logf("Configuration: %#v", configuration)
	t.Logf("Error: %#v", err)
	if err == nil {
		t.Errorf("sparta.Discover() failed to error when not initialized")
	}
	t.Logf("Properly rejected invalid callsite: %s", err.Error())
}
