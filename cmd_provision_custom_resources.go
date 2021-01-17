package sparta

import "fmt"

func customResourceDescription(serviceName string, targetType string) string {
	return fmt.Sprintf("%s: CloudFormation CustomResource to configure %s",
		serviceName,
		targetType)
}
