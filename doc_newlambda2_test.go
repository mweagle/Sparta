package sparta

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/sirupsen/logrus"
)

func lambdaHelloWorld2(ctx context.Context,
	props map[string]interface{}) error {
	lambdaCtx, _ := lambdacontext.FromContext(ctx)
	Logger().WithFields(logrus.Fields{
		"RequestID": lambdaCtx.AwsRequestID,
	}).Info("Lambda event")
	Logger().Info("Event received")
	return nil
}
func ExampleNewAWSLambda_iAMRoleDefinition() {
	roleDefinition := IAMRoleDefinition{}
	roleDefinition.Privileges = append(roleDefinition.Privileges, IAMRolePrivilege{
		Actions: []string{"s3:GetObject",
			"s3:PutObject"},
		Resource: "arn:aws:s3:::*",
	})
	helloWorldLambda, _ := NewAWSLambda(LambdaName(lambdaHelloWorld2),
		lambdaHelloWorld2,
		IAMRoleDefinition{})
	if nil != helloWorldLambda {
		fmt.Printf("Failed to create new Lambda function")
	}
}
