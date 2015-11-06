package sparta

const S3_BUCKET = "arn:aws:sns:us-west-2:123412341234:myBucket"

func s3LambdaProcessor(event *sparta.LambdaEvent, context *sparta.LambdaContext, w *http.ResponseWriter, logger *logrus.Logger) {
	logger.WithFields(logrus.Fields{
		"RequestID": context.AWSRequestId,
	}).Info("S3Event")

	eventData, err := json.Marshal(*event)
	if err != nil {
		logger.Error("Failed to marshal event data: ", err.Error())
	}
	logger.Info("Event data: ", string(eventData))
}

func ExampleS3Permission() {
	var lambdaFunctions []*LambdaAWSInfo
	// Define the IAM role
	roleDefinition := sparta.IAMRoleDefinition{}
	roleDefinition.Privileges = append(roleDefinition.Privileges, sparta.IAMRolePrivilege{
		Actions: []string{"s3:GetObject",
			"s3:PutObject"},
		Resource: S3_BUCKET,
	})
	// Create the Lambda
	s3Lambda := NewLambda(sparta.IAMRoleDefinition{}, s3LambdaProcessor, nil)

	// Add a Permission s.t. the Lambda function automatically registers for S3 events
	s3Lambda.Permissions = append(s3Lambda.Permissions, sparta.S3Permission{
		BasePermission: sparta.BasePermission{
			SourceArn: S3_BUCKET,
		},
		Events: []string{"s3:ObjectCreated:*", "s3:ObjectRemoved:*"},
	})

	lambdaFunctions = append(lambdaFunctions, s3Lambda)
	sparta.Main("S3LambdaApp", "Registers for S3 events", lambdaFunctions)
}
