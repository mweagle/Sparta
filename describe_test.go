package sparta_test

import (
	sparta "Sparta"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
)

const LAMBDA_EXECUTE_ARN = "LambdaExecutor"

func mockLambda1(event *json.RawMessage, context *sparta.LambdaContext, w http.ResponseWriter, logger *logrus.Logger) {
	fmt.Fprintf(w, "mockLambda1!")
}

func mockLambda2(event *json.RawMessage, context *sparta.LambdaContext, w http.ResponseWriter, logger *logrus.Logger) {
	fmt.Fprintf(w, "mockLambda2!")
}

func sampleData() []*sparta.LambdaAWSInfo {
	var lambdaFunctions []*sparta.LambdaAWSInfo
	lambdaFn := sparta.NewLambda(LAMBDA_EXECUTE_ARN, mockLambda1, nil)
	lambdaFn.Permissions = append(lambdaFn.Permissions, sparta.S3Permission{
		BasePermission: sparta.BasePermission{
			SourceArn: "arn:aws:s3:::sampleBucket",
		},
		// Event Filters are defined at
		// http://docs.aws.amazon.com/AmazonS3/latest/dev/NotificationHowTo.html
		Events: []string{"s3:ObjectCreated:*", "s3:ObjectRemoved:*"},
	})

	lambdaFn.Permissions = append(lambdaFn.Permissions, sparta.SNSPermission{
		BasePermission: sparta.BasePermission{
			SourceArn: "arn:aws:sns:us-west-2:000000000000:someTopic",
		},
	})

	lambdaFn.EventSourceMappings = append(lambdaFn.EventSourceMappings, &lambda.CreateEventSourceMappingInput{
		EventSourceArn:   aws.String("arn:aws:dynamodb:us-west-2:000000000000:table/sampleTable"),
		StartingPosition: aws.String("TRIM_HORIZON"),
		BatchSize:        aws.Int64(10),
	})

	lambdaFunctions = append(lambdaFunctions, lambdaFn)
	lambdaFunctions = append(lambdaFunctions, sparta.NewLambda(LAMBDA_EXECUTE_ARN, mockLambda2, nil))
	return lambdaFunctions
}

func TestDescribe(t *testing.T) {
	logger, err := sparta.NewLogger("info")
	//err := sparta.Main("SampleService", "SampleService Description", sampleData())
	output, err := os.Create("./graph.html")
	if nil != err {
		t.Fatalf(err.Error())
		return
	}
	defer output.Close()
	err = sparta.Describe("SampleService", "SampleService Description", sampleData(), output, logger)
	if nil != err {
		t.Errorf("Failed to describe: %s", err)
	}
}
