package sparta

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mweagle/cloudformationresources"

	"github.com/Sirupsen/logrus"
)

func javascriptExportNameForCustomResourceType(customResourceTypeName string) string {
	return sanitizedName(customResourceTypeName)
}
func lambdaExportNameForCustomResourceType(customResourceTypeName string) string {
	return fmt.Sprintf("index.%s", javascriptExportNameForCustomResourceType(customResourceTypeName))
}

func customResourceDescription(serviceName string, targetType string) string {
	return fmt.Sprintf("%s: CloudFormation CustomResource to configure %s",
		serviceName,
		targetType)
}

// Extract the fields and forward the event to the resource
func customResourceForwarder(event *json.RawMessage,
	context *LambdaContext,
	w http.ResponseWriter,
	logger *logrus.Logger) {

	var rawProps map[string]interface{}
	json.Unmarshal([]byte(*event), &rawProps)

	var lambdaEvent cloudformationresources.CloudFormationLambdaEvent
	jsonErr := json.Unmarshal([]byte(*event), &lambdaEvent)
	if jsonErr != nil {
		logger.WithFields(logrus.Fields{
			"RawEvent":       rawProps,
			"UnmarshalError": jsonErr,
		}).Warn("Raw event data")
		http.Error(w, jsonErr.Error(), http.StatusInternalServerError)
	}

	logger.WithFields(logrus.Fields{
		"LambdaEvent": lambdaEvent,
	}).Debug("CloudFormation Lambda event")

	// Setup the request and send it off
	customResourceRequest := cloudformationresources.CustomResourceRequest{
		RequestType:        lambdaEvent.RequestType,
		ResponseURL:        lambdaEvent.ResponseURL,
		StackID:            lambdaEvent.StackID,
		RequestID:          lambdaEvent.RequestID,
		LogicalResourceID:  lambdaEvent.LogicalResourceID,
		PhysicalResourceID: lambdaEvent.PhysicalResourceID,
		LogGroupName:       context.LogGroupName,
		LogStreamName:      context.LogStreamName,
		ResourceProperties: lambdaEvent.ResourceProperties,
	}
	if "" == customResourceRequest.PhysicalResourceID {
		customResourceRequest.PhysicalResourceID = fmt.Sprintf("LogStreamName: %s", context.LogStreamName)
	}

	requestErr := cloudformationresources.Handle(&customResourceRequest, logger)
	if requestErr != nil {
		http.Error(w, requestErr.Error(), http.StatusInternalServerError)
	} else {
		fmt.Fprint(w, "CustomResource handled: "+lambdaEvent.LogicalResourceID)
	}
}
