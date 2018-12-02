package sparta

import (
	"context"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/sirupsen/logrus"
)

func cloudWatchEventProcessor(ctx context.Context,
	event map[string]interface{}) (map[string]interface{}, error) {

	lambdaCtx, _ := lambdacontext.FromContext(ctx)
	Logger().WithFields(logrus.Fields{
		"RequestID": lambdaCtx.AwsRequestID,
	}).Info("Request received")
	Logger().Info("CloudWatch Event received")
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
	Main("CloudWatchLogs", "Registers for CloudWatch Logs", lambdaFunctions, nil, nil)
}
