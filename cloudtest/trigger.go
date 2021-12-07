package cloudtest

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	awsv2Lambda "github.com/aws/aws-sdk-go-v2/service/lambda"
	awsv2LambdaTypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	awsv2S3 "github.com/aws/aws-sdk-go-v2/service/s3"
	awsv2SQS "github.com/aws/aws-sdk-go-v2/service/sqs"
)

////////////////////////////////////////////////////////////////////////////////
//

type cloudNOPTrigger struct {
}

func (cnp *cloudNOPTrigger) Send(t CloudTest, output *awsv2Lambda.GetFunctionOutput) (interface{}, error) {
	t.Logf("NOP Trigger")
	return nil, nil
}

func (cnp *cloudNOPTrigger) Cleanup(t CloudTest, output *awsv2Lambda.GetFunctionOutput) {

}

// NewNOPTrigger is an empty trigger that does nothing
func NewNOPTrigger() Trigger {
	return &cloudNOPTrigger{}
}

////////////////////////////////////////////////////////////////////////////////
//

type lambdaInvokeTrigger struct {
	eventBody []byte
}

func (lim *lambdaInvokeTrigger) Send(t CloudTest, output *awsv2Lambda.GetFunctionOutput) (interface{}, error) {
	lambdaSvc := awsv2Lambda.NewFromConfig(t.Config())

	invokeInput := &awsv2Lambda.InvokeInput{
		FunctionName: output.Configuration.FunctionArn,
		LogType:      awsv2LambdaTypes.LogTypeTail,
		Payload:      lim.eventBody,
	}
	t.Logf("Submitting LambdaInvoke: %#v\n", invokeInput)
	return lambdaSvc.Invoke(context.Background(), invokeInput)
}

func (lim *lambdaInvokeTrigger) Cleanup(t CloudTest, output *awsv2Lambda.GetFunctionOutput) {

}

// NewLambdaInvokeTrigger returns a Trigger instance that directly invokes
// a Lambda function
func NewLambdaInvokeTrigger(event []byte) Trigger {
	return &lambdaInvokeTrigger{
		eventBody: event,
	}
}

////////////////////////////////////////////////////////////////////////////////
//

type sqsMessageTrigger struct {
	sqsURL      string
	messageBody string
}

func (ssub *sqsMessageTrigger) Send(t CloudTest,
	output *awsv2Lambda.GetFunctionOutput) (interface{}, error) {

	sqsSvc := awsv2SQS.NewFromConfig(t.Config())
	messageInput := &awsv2SQS.SendMessageInput{
		MessageBody: awsv2.String(ssub.messageBody),
		QueueUrl:    awsv2.String(ssub.sqsURL),
	}
	t.Logf("Submitting SQS Message: %#v\n", messageInput)
	return sqsSvc.SendMessage(context.Background(), messageInput)
}

func (ssub *sqsMessageTrigger) Cleanup(t CloudTest, output *awsv2Lambda.GetFunctionOutput) {

}

// NewSQSMessageTrigger is a trigger that submits a message to an SQS URL
func NewSQSMessageTrigger(sqsURL string, messageBody string) Trigger {
	return &sqsMessageTrigger{
		sqsURL:      sqsURL,
		messageBody: messageBody,
	}
}

////////////////////////////////////////////////////////////////////////////////
//

type s3PayloadTrigger struct {
	putObjectParams *awsv2S3.PutObjectInput
}

func (spt *s3PayloadTrigger) Send(t CloudTest,
	output *awsv2Lambda.GetFunctionOutput) (interface{}, error) {
	s3Svc := awsv2S3.NewFromConfig(t.Config())
	t.Logf("Submitting S3 object: %#v\n", spt.putObjectParams)
	return s3Svc.PutObject(context.Background(), spt.putObjectParams)
}

func (spt *s3PayloadTrigger) Cleanup(t CloudTest, output *awsv2Lambda.GetFunctionOutput) {
	delObjectInput := &awsv2S3.DeleteObjectInput{
		Bucket: spt.putObjectParams.Bucket,
		Key:    spt.putObjectParams.Key,
	}
	s3Svc := awsv2S3.NewFromConfig(t.Config())
	_, delErr := s3Svc.DeleteObject(context.Background(), delObjectInput)
	if delErr != nil {
		t.Logf("Failed to delete object: %s", delErr.Error())
	}
}

// NewS3MessageTrigger returns an S3 Trigger that posts a payload
// to the given bucket, key
func NewS3MessageTrigger(s3Bucket string, s3Key string, body io.ReadSeeker) Trigger {
	return &s3PayloadTrigger{
		putObjectParams: &awsv2S3.PutObjectInput{
			Bucket: awsv2.String(s3Bucket),
			Key:    awsv2.String(s3Key),
			Body:   body,
		},
	}
}

// NewS3FileMessageTrigger returns an S3 Trigger that posts a payload
// to the given bucket, key
func NewS3FileMessageTrigger(s3Bucket string, s3Key string, localFilePath string) Trigger {
	trigger := &s3PayloadTrigger{
		putObjectParams: &awsv2S3.PutObjectInput{
			Bucket: awsv2.String(s3Bucket),
			Key:    awsv2.String(s3Key),
			Body:   nil,
		},
	}
	// Read the body...
	/* #nosec G304 */
	allData, allDataErr := ioutil.ReadFile(localFilePath)
	if allDataErr != nil {
		trigger.putObjectParams.Body = bytes.NewReader(allData)
		contentType := http.DetectContentType(allData)
		trigger.putObjectParams.ContentType = awsv2.String(contentType)

	}
	return trigger
}
