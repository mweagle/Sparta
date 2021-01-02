package sparta

import (
	"context"

	"github.com/aws/aws-lambda-go/lambdacontext"
)

func cloudWatchLogsProcessor(ctx context.Context,
	props map[string]interface{}) error {
	lambdaCtx, _ := lambdacontext.FromContext(ctx)
	Logger().Info().
		Str("RequestID", lambdaCtx.AwsRequestID).
		Msg("CloudWatch log event")
	Logger().Info().Msg("CloudWatch Log event received")
	return nil
}

func ExampleCloudWatchLogsPermission() {
	var lambdaFunctions []*LambdaAWSInfo

	cloudWatchLogsLambda, _ := NewAWSLambda(LambdaName(cloudWatchLogsProcessor),
		cloudWatchLogsProcessor,
		IAMRoleDefinition{})

	cloudWatchLogsPermission := CloudWatchLogsPermission{}
	cloudWatchLogsPermission.Filters = make(map[string]CloudWatchLogsSubscriptionFilter, 1)
	cloudWatchLogsPermission.Filters["MyFilter"] = CloudWatchLogsSubscriptionFilter{
		LogGroupName: "/aws/lambda/*",
	}
	cloudWatchLogsLambda.Permissions = append(cloudWatchLogsLambda.Permissions, cloudWatchLogsPermission)

	lambdaFunctions = append(lambdaFunctions, cloudWatchLogsLambda)
	mainErr := Main("CloudWatchLogs", "Registers for CloudWatch Logs", lambdaFunctions, nil, nil)
	if mainErr != nil {
		panic("Failed to invoke sparta.Main: %s" + mainErr.Error())
	}
}
