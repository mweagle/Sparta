// Copyright (c) 2015 Matt Weagle <mweagle@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.
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

func mockLambda1(event *sparta.LambdaEvent, context *sparta.LambdaContext, w *http.ResponseWriter, logger *logrus.Logger) {
	fmt.Fprintf(*w, "mockLambda1!")
}

func mockLambda2(event *sparta.LambdaEvent, context *sparta.LambdaContext, w *http.ResponseWriter, logger *logrus.Logger) {
	fmt.Fprintf(*w, "mockLambda2!")
}

func sampleData() []*sparta.LambdaAWSInfo {
	var lambdaFunctions []*sparta.LambdaAWSInfo
	lambdaFn := sparta.NewLambda(LAMBDA_EXECUTE_ARN, mockLambda1, nil)
	lambdaFn.Permissions = append(lambdaFn.Permissions, sparta.S3Permission{
		BasePermission: sparta.BasePermission{
			StatementId: "MyUniqueID",
			SourceArn:   "arn:aws:s3:::sampleBucket",
		},
		// Event Filters are defined at
		// http://docs.aws.amazon.com/AmazonS3/latest/dev/NotificationHowTo.html
		Events: []string{"s3:ObjectCreated:*", "s3:ObjectRemoved:*"},
	})

	lambdaFn.Permissions = append(lambdaFn.Permissions, sparta.SNSPermission{
		BasePermission: sparta.BasePermission{
			StatementId: "MyUniqueID",
			SourceArn:   "arn:aws:sns:us-west-2:000000000000:someTopic",
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
