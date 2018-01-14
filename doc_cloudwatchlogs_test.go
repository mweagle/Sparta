package sparta

import (
	"context"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/sirupsen/logrus"
)

func cloudWatchLogsProcessor(ctx context.Context,
	props map[string]interface{}) error {
	lambdaCtx, _ := lambdacontext.FromContext(ctx)
	Logger().WithFields(logrus.Fields{
		"RequestID": lambdaCtx.AwsRequestID,
	}).Info("CloudWatch log event")
	Logger().Info("CloudWatch Log event received")
	return nil
}

func ExampleCloudWatchLogsPermission() {
	var lambdaFunctions []*LambdaAWSInfo

	cloudWatchLogsLambda := HandleAWSLambda(LambdaName(cloudWatchLogsProcessor),
		cloudWatchLogsProcessor,
		IAMRoleDefinition{})

	cloudWatchLogsPermission := CloudWatchLogsPermission{}
	cloudWatchLogsPermission.Filters = make(map[string]CloudWatchLogsSubscriptionFilter, 1)
	cloudWatchLogsPermission.Filters["MyFilter"] = CloudWatchLogsSubscriptionFilter{
		LogGroupName: "/aws/lambda/*",
	}
	cloudWatchLogsLambda.Permissions = append(cloudWatchLogsLambda.Permissions, cloudWatchLogsPermission)

	lambdaFunctions = append(lambdaFunctions, cloudWatchLogsLambda)
	Main("CloudWatchLogs", "Registers for CloudWatch Logs", lambdaFunctions, nil, nil)
}
