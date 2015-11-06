package sparta

const SNS_TOPIC = "arn:aws:sns:us-west-2:123412341234:mySNSTopic"

func snsProcessor(event *sparta.LambdaEvent, context *sparta.LambdaContext, w *http.ResponseWriter, logger *logrus.Logger) {
	logger.WithFields(logrus.Fields{
		"RequestID": context.AWSRequestId,
	}).Info("SNSEvent")

	eventData, err := json.Marshal(*event)
	if err != nil {
		logger.Error("Failed to marshal event data: ", err.Error())
		http.Error(*w, err.Error(), http.StatusInternalServerError)
		return
	}
	logger.Info("Event data: ", string(eventData))
}

func ExampleSNSPermission() {
	var lambdaFunctions []*sparta.LambdaAWSInfo

	snsLambda := sparta.NewLambda(sparta.IAMRoleDefinition{}, snsProcessor, nil)
	lambdaFn.Permissions = append(lambdaFn.Permissions, sparta.SNSPermission{
		BasePermission: sparta.BasePermission{
			SourceArn: SNS_TOPIC,
		},
	})
	lambdaFunctions = append(lambdaFunctions, snsLambda)
	sparta.Main("SNSLambdaApp", "Registers for SNS events", lambdaFunctions)
}
