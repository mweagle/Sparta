package private

import "errors"

// InsertSpartaOutput adds a CloudFormation-compatible "Output" entry to the `outputs`
// object. The key will be namespaced and checked for collisions.
func InsertSpartaOutput(key string, value interface{}, description string, outputs map[string]interface{}) error {
	if _, exists := outputs[key]; exists {
		return errors.New("Output key already exists: " + key)
	}

	outputs[key] = map[string]interface{}{
		"Description": description,
		"Value":       value,
	}
	return nil
}
