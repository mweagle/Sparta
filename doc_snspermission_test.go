package sparta

import (
	"encoding/json"
	"github.com/Sirupsen/logrus"
	"net/http"
)

const SNS_TOPIC = "arn:aws:sns:us-west-2:123412341234:mySNSTopic"

func snsProcessor(event *json.RawMessage, context *LambdaContext, w *http.ResponseWriter, logger *logrus.Logger) {
	logger.WithFields(logrus.Fields{
		"RequestID": context.AWSRequestId,
	}).Info("SNSEvent")
	logger.Info("Event data: ", string(*event))
}

func ExampleSNSPermission() {
	var lambdaFunctions []*LambdaAWSInfo

	snsLambda := NewLambda(IAMRoleDefinition{}, snsProcessor, nil)
	snsLambda.Permissions = append(snsLambda.Permissions, SNSPermission{
		BasePermission: BasePermission{
			SourceArn: SNS_TOPIC,
		},
	})
	lambdaFunctions = append(lambdaFunctions, snsLambda)
	Main("SNSLambdaApp", "Registers for SNS events", lambdaFunctions)
}
