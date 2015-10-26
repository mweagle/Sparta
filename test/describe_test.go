package sparta_test

import (
	sparta "Sparta"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"net/http"
	"os"
	"testing"
)

const LAMBDA_EXECUTE_ARN = "LambdaExecutor"

func testLogger() *logrus.Logger {
	logger := logrus.New()
	logger.Formatter = new(logrus.TextFormatter)
	return logger
}

func mockLambda1(event sparta.LambdaEvent, context sparta.LambdaContext, w http.ResponseWriter) {
	fmt.Fprintf(w, "mockLambda1!")
}

func mockLambda2(event sparta.LambdaEvent, context sparta.LambdaContext, w http.ResponseWriter) {
	fmt.Fprintf(w, "mockLambda2!")
}

func sampleData() []*sparta.LambdaAWSInfo {
	var lambdaFunctions []*sparta.LambdaAWSInfo
	lambdaFn := sparta.NewLambda(LAMBDA_EXECUTE_ARN, mockLambda1, nil)
	lambdaFn.Permissions = append(lambdaFn.Permissions, &lambda.AddPermissionInput{
		Action:      aws.String("lambda:InvokeFunction"),
		Principal:   aws.String("s3.amazonaws.com"),
		StatementId: aws.String("Woor"),
		SourceArn:   aws.String("arn:aws:s3:::myS3Bucket"),
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
		t.Errorf("Failed to describe: ", err)
	}
}
