package resources

import (
	"bytes"
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	gocf "github.com/mweagle/go-cloudformation"
	"github.com/sirupsen/logrus"
)

// S3ArtifactPublisherResourceRequest is what the UserProperties
// should be set to in the CustomResource invocation
type S3ArtifactPublisherResourceRequest struct {
	Bucket *gocf.StringExpr
	Key    *gocf.StringExpr
	Body   *gocf.StringExpr
}

// S3ArtifactPublisherResource is a simple POC showing how to create custom resources
type S3ArtifactPublisherResource struct {
	gocf.CloudFormationCustomResource
	S3ArtifactPublisherResourceRequest
}

// IAMPrivileges returns the IAM privs for this custom action
func (command *S3ArtifactPublisherResource) IAMPrivileges() []string {
	return []string{"s3:PutObject",
		"s3:DeleteObject"}
}

// Create implements the S3 create operation
func (command S3ArtifactPublisherResource) Create(awsSession *session.Session,
	event *CloudFormationLambdaEvent,
	logger *logrus.Logger) (map[string]interface{}, error) {

	unmarshalErr := json.Unmarshal(event.ResourceProperties, &command)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}
	itemInput := bytes.NewReader([]byte(command.Body.Literal))
	s3PutObjectParams := &s3.PutObjectInput{
		Body:   itemInput,
		Bucket: aws.String(command.Bucket.Literal),
		Key:    aws.String(command.Key.Literal),
	}
	s3Svc := s3.New(awsSession)
	s3Response, s3ResponseErr := s3Svc.PutObject(s3PutObjectParams)
	if s3ResponseErr != nil {
		return nil, s3ResponseErr
	}
	return map[string]interface{}{
		"ObjectVersion": s3Response.VersionId,
	}, nil
}

// Update implements the S3 update operation
func (command S3ArtifactPublisherResource) Update(awsSession *session.Session,
	event *CloudFormationLambdaEvent,
	logger *logrus.Logger) (map[string]interface{}, error) {
	return command.Create(awsSession, event, logger)
}

// Delete implements the S3 delete operation
func (command S3ArtifactPublisherResource) Delete(awsSession *session.Session,
	event *CloudFormationLambdaEvent,
	logger *logrus.Logger) (map[string]interface{}, error) {

	unmarshalErr := json.Unmarshal(event.ResourceProperties, &command)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}
	s3DeleteObjectParams := &s3.DeleteObjectInput{
		Bucket: aws.String(command.Bucket.Literal),
		Key:    aws.String(command.Key.Literal),
	}
	s3Svc := s3.New(awsSession)
	_, s3ResponseErr := s3Svc.DeleteObject(s3DeleteObjectParams)
	if s3ResponseErr != nil {
		return nil, s3ResponseErr
	}
	logger.WithFields(logrus.Fields{
		"Bucket": command.Bucket.Literal,
		"Key":    command.Key.Literal,
	}).Info("Object deleted")
	return nil, nil
}
