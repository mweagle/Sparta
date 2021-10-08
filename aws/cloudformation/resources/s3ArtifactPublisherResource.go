package resources

import (
	"bytes"
	"context"
	"encoding/json"

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	awsv2S3 "github.com/aws/aws-sdk-go-v2/service/s3"

	gof "github.com/awslabs/goformation/v5/cloudformation"
	"github.com/rs/zerolog"
)

// S3ArtifactPublisherResourceRequest is what the UserProperties
// should be set to in the CustomResource invocation
type S3ArtifactPublisherResourceRequest struct {
	CustomResourceRequest
	Bucket string
	Key    string
	Body   map[string]interface{}
}

// S3ArtifactPublisherResource is a simple POC showing how to create custom resources
type S3ArtifactPublisherResource struct {
	gof.CustomResource
}

// IAMPrivileges returns the IAM privs for this custom action
func (command *S3ArtifactPublisherResource) IAMPrivileges() []string {
	return []string{"s3:PutObject",
		"s3:DeleteObject"}
}

// Create implements the S3 create operation
func (command S3ArtifactPublisherResource) Create(awsConfig awsv2.Config,
	event *CloudFormationLambdaEvent,
	logger *zerolog.Logger) (map[string]interface{}, error) {

	s3ArtifactPublisherRequest := S3ArtifactPublisherResourceRequest{}
	unmarshalErr := json.Unmarshal(event.ResourceProperties, &s3ArtifactPublisherRequest)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}
	mapData, mapDataErr := json.Marshal(s3ArtifactPublisherRequest.Body)
	if mapDataErr != nil {
		return nil, mapDataErr
	}
	itemInput := bytes.NewReader(mapData)
	s3PutObjectParams := &awsv2S3.PutObjectInput{
		Body:   itemInput,
		Bucket: awsv2.String(s3ArtifactPublisherRequest.Bucket),
		Key:    awsv2.String(s3ArtifactPublisherRequest.Key),
	}
	s3Svc := awsv2S3.NewFromConfig(awsConfig)
	s3Response, s3ResponseErr := s3Svc.PutObject(context.Background(), s3PutObjectParams)
	if s3ResponseErr != nil {
		return nil, s3ResponseErr
	}
	return map[string]interface{}{
		"ObjectVersion": s3Response.VersionId,
	}, nil
}

// Update implements the S3 update operation
func (command S3ArtifactPublisherResource) Update(awsConfig awsv2.Config,
	event *CloudFormationLambdaEvent,
	logger *zerolog.Logger) (map[string]interface{}, error) {
	return command.Create(awsConfig, event, logger)
}

// Delete implements the S3 delete operation
func (command S3ArtifactPublisherResource) Delete(awsConfig awsv2.Config,
	event *CloudFormationLambdaEvent,
	logger *zerolog.Logger) (map[string]interface{}, error) {

	s3ArtifactPublisherRequest := S3ArtifactPublisherResourceRequest{}
	unmarshalErr := json.Unmarshal(event.ResourceProperties, &s3ArtifactPublisherRequest)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}
	s3DeleteObjectParams := &awsv2S3.DeleteObjectInput{
		Bucket: awsv2.String(s3ArtifactPublisherRequest.Bucket),
		Key:    awsv2.String(s3ArtifactPublisherRequest.Key),
	}
	s3Svc := awsv2S3.NewFromConfig(awsConfig)
	_, s3ResponseErr := s3Svc.DeleteObject(context.Background(), s3DeleteObjectParams)
	if s3ResponseErr != nil {
		return nil, s3ResponseErr
	}
	logger.Info().
		Str("Bucket", s3ArtifactPublisherRequest.Bucket).
		Str("Key", s3ArtifactPublisherRequest.Key).
		Msg("Object deleted")
	return nil, nil
}
