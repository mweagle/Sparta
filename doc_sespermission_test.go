package sparta

import (
	"encoding/json"
	"net/http"

	"github.com/Sirupsen/logrus"
)

func sesLambdaProcessor(event *json.RawMessage, context *LambdaContext, w http.ResponseWriter, logger *logrus.Logger) {
	logger.WithFields(logrus.Fields{
		"RequestID": context.AWSRequestID,
	}).Info("SES Event")

	logger.Info("Event data: ", string(*event))
}

func ExampleSESPermission() {
	var lambdaFunctions []*LambdaAWSInfo
	// Define the IAM role
	roleDefinition := IAMRoleDefinition{}

	// Create the Lambda
	sesLambda := NewLambda(roleDefinition, sesLambdaProcessor, nil)

	// Add a Permission s.t. the Lambda function automatically registers for S3 events
	sesLambda.Permissions = append(sesLambda.Permissions, SESPermission{
		BasePermission: BasePermission{
			SourceArn: "*",
		},
		InvocationType: "Event",
	})

	lambdaFunctions = append(lambdaFunctions, sesLambda)
	Main("SESLambdaApp", "Registers for SES events", lambdaFunctions, nil, nil)
}
