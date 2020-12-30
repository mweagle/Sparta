package cloudtest

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sqs"
)

////////////////////////////////////////////////////////////////////////////////
//

type cloudNOPTrigger struct {
}

func (cnp *cloudNOPTrigger) Send(t CloudTest, output *lambda.GetFunctionOutput) (interface{}, error) {
	t.Logf("NOP Trigger")
	return nil, nil
}

func (cnp *cloudNOPTrigger) Cleanup(t CloudTest, output *lambda.GetFunctionOutput) {

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

func (lim *lambdaInvokeTrigger) Send(t CloudTest, output *lambda.GetFunctionOutput) (interface{}, error) {
	lambdaSvc := lambda.New(t.Session())

	invokeInput := &lambda.InvokeInput{
		FunctionName: output.Configuration.FunctionArn,
		LogType:      aws.String(lambda.LogTypeTail),
		Payload:      lim.eventBody,
	}
	t.Logf("Submitting LambdaInvoke: %#v\n", invokeInput)
	return lambdaSvc.Invoke(invokeInput)
}

func (lim *lambdaInvokeTrigger) Cleanup(t CloudTest, output *lambda.GetFunctionOutput) {

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
	output *lambda.GetFunctionOutput) (interface{}, error) {

	sqsSvc := sqs.New(t.Session())
	messageInput := &sqs.SendMessageInput{
		MessageBody: aws.String(ssub.messageBody),
		QueueUrl:    aws.String(ssub.sqsURL),
	}
	t.Logf("Submitting SQS Message: %#v\n", messageInput)
	return sqsSvc.SendMessage(messageInput)
}

func (ssub *sqsMessageTrigger) Cleanup(t CloudTest, output *lambda.GetFunctionOutput) {

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
	putObjectParams *s3.PutObjectInput
}

func (spt *s3PayloadTrigger) Send(t CloudTest,
	output *lambda.GetFunctionOutput) (interface{}, error) {
	s3Svc := s3.New(t.Session())
	t.Logf("Submitting S3 object: %#v\n", spt.putObjectParams)
	return s3Svc.PutObject(spt.putObjectParams)
}

func (spt *s3PayloadTrigger) Cleanup(t CloudTest, output *lambda.GetFunctionOutput) {
	delObjectInput := &s3.DeleteObjectInput{
		Bucket: spt.putObjectParams.Bucket,
		Key:    spt.putObjectParams.Key,
	}
	s3Svc := s3.New(t.Session())
	_, delErr := s3Svc.DeleteObject(delObjectInput)
	if delErr != nil {
		t.Logf("Failed to delete object: %s", delErr.Error())
	}
}

// NewS3MessageTrigger returns an S3 Trigger that posts a payload
// to the given bucket, key
func NewS3MessageTrigger(s3Bucket string, s3Key string, body io.ReadSeeker) Trigger {
	return &s3PayloadTrigger{
		putObjectParams: &s3.PutObjectInput{
			Bucket: aws.String(s3Bucket),
			Key:    aws.String(s3Key),
			Body:   body,
		},
	}
}

// NewS3FileMessageTrigger returns an S3 Trigger that posts a payload
// to the given bucket, key
func NewS3FileMessageTrigger(s3Bucket string, s3Key string, localFilePath string) Trigger {
	trigger := &s3PayloadTrigger{
		putObjectParams: &s3.PutObjectInput{
			Bucket: aws.String(s3Bucket),
			Key:    aws.String(s3Key),
			Body:   nil,
		},
	}
	// Read the body...
	/* #nosec G304 */
	allData, allDataErr := ioutil.ReadFile(localFilePath)
	if allDataErr != nil {
		trigger.putObjectParams.Body = bytes.NewReader(allData)
		contentType := http.DetectContentType(allData)
		trigger.putObjectParams.ContentType = aws.String(contentType)

	}
	return trigger
}
