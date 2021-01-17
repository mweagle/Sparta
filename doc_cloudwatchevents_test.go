package sparta

import (
	"context"

	"github.com/aws/aws-lambda-go/lambdacontext"
)

func cloudWatchEventProcessor(ctx context.Context,
	event map[string]interface{}) (map[string]interface{}, error) {

	lambdaCtx, _ := lambdacontext.FromContext(ctx)
	Logger().Info().
		Str("RequestID", lambdaCtx.AwsRequestID).
		Msg("Request received")
	Logger().Info().Msg("CloudWatch Event received")
	return nil, nil
}

func ExampleCloudWatchEventsPermission() {
	cloudWatchEventsLambda, _ := NewAWSLambda(LambdaName(cloudWatchEventProcessor),
		cloudWatchEventProcessor,
		IAMRoleDefinition{})

	cloudWatchEventsPermission := CloudWatchEventsPermission{}
	cloudWatchEventsPermission.Rules = make(map[string]CloudWatchEventsRule)
	cloudWatchEventsPermission.Rules["Rate5Mins"] = CloudWatchEventsRule{
		ScheduleExpression: "rate(5 minutes)",
	}
	cloudWatchEventsPermission.Rules["EC2Activity"] = CloudWatchEventsRule{
		EventPattern: map[string]interface{}{
			"source":      []string{"aws.ec2"},
			"detail-type": []string{"EC2 Instance State-change Notification"},
		},
	}
	cloudWatchEventsLambda.Permissions = append(cloudWatchEventsLambda.Permissions,
		cloudWatchEventsPermission)
	var lambdaFunctions []*LambdaAWSInfo
	lambdaFunctions = append(lambdaFunctions, cloudWatchEventsLambda)
	mainErr := Main("CloudWatchLogs", "Registers for CloudWatch Logs", lambdaFunctions, nil, nil)
	if mainErr != nil {
		panic("Failed to invoke sparta.Main: %s" + mainErr.Error())
	}
}
