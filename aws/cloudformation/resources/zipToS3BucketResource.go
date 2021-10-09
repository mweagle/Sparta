package resources

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"os"
	"path"
	"strings"

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	awsv2S3 "github.com/aws/aws-sdk-go-v2/service/s3"
	awsv2S3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"

	gof "github.com/awslabs/goformation/v5/cloudformation"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// DefaultManifestName is the name of the file that will be created
// at the root of the S3 bucket with user-supplied metadata
const DefaultManifestName = "MANIFEST.json"

// ZipToS3BucketResourceRequest is the data request made to a ZipToS3BucketResource
// lambda handler
type ZipToS3BucketResourceRequest struct {
	CustomResourceRequest
	SrcBucket    string
	SrcKeyName   string
	DestBucket   string
	ManifestName string
	Manifest     map[string]interface{}
}

// ZipToS3BucketResource manages populating an S3 bucket with the contents
// of a ZIP file...
type ZipToS3BucketResource struct {
	gof.CustomResource
}

func (command ZipToS3BucketResource) unzip(awsConfig awsv2.Config,
	event *CloudFormationLambdaEvent,
	logger *zerolog.Logger) (map[string]interface{}, error) {
	request := ZipToS3BucketResourceRequest{}
	unmarshalErr := json.Unmarshal(event.ResourceProperties, &request)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	// Fetch the ZIP contents and unpack them to the S3 bucket
	svc := awsv2S3.NewFromConfig(awsConfig)
	s3Object, s3ObjectErr := svc.GetObject(context.Background(), &awsv2S3.GetObjectInput{
		Bucket: awsv2.String(request.SrcBucket),
		Key:    awsv2.String(request.SrcKeyName),
	})
	if nil != s3ObjectErr {
		return nil, s3ObjectErr
	}
	// Put all the ZIP contents to the bucket
	defer s3Object.Body.Close()
	destFile, destFileErr := ioutil.TempFile("", "s3")
	if nil != destFileErr {
		return nil, destFileErr
	}
	defer os.Remove(destFile.Name())

	_, copyErr := io.Copy(destFile, s3Object.Body)
	if nil != copyErr {
		return nil, copyErr
	}
	zipReader, zipErr := zip.OpenReader(destFile.Name())
	if nil != zipErr {
		return nil, zipErr
	}
	// Iterate through the files in the archive,
	// printing some of their contents.
	// TODO - refactor to a worker pool
	totalFiles := 0
	for _, eachFile := range zipReader.File {
		totalFiles++

		stream, streamErr := eachFile.Open()
		if nil != streamErr {
			return nil, streamErr
		}
		bodySource, bodySourceErr := ioutil.ReadAll(stream)
		if nil != bodySourceErr {
			return nil, bodySourceErr
		}
		normalizedName := strings.TrimLeft(eachFile.Name, "/")
		// Mime type?
		fileExtension := path.Ext(eachFile.Name)
		mimeType := mime.TypeByExtension(fileExtension)
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}

		if len(normalizedName) > 0 {
			s3PutObject := &awsv2S3.PutObjectInput{
				Body:        bytes.NewReader(bodySource),
				Bucket:      awsv2.String(request.DestBucket),
				Key:         awsv2.String(fmt.Sprintf("/%s", eachFile.Name)),
				ContentType: awsv2.String(mimeType),
			}
			_, err := svc.PutObject(context.Background(), s3PutObject)
			if err != nil {
				return nil, err
			}
		}
		errClose := stream.Close()
		if errClose != nil {
			return nil, errors.Wrapf(errClose, "Failed to close S3 PutObject stream")
		}
	}
	// Need to add the manifest data iff defined
	if nil != request.Manifest {
		manifestBytes, manifestErr := json.Marshal(request.Manifest)
		if nil != manifestErr {
			return nil, manifestErr
		}
		name := request.ManifestName
		if name == "" {
			name = DefaultManifestName
		}
		s3PutObject := &awsv2S3.PutObjectInput{
			Body:        bytes.NewReader(manifestBytes),
			Bucket:      awsv2.String(request.DestBucket),
			Key:         awsv2.String(name),
			ContentType: awsv2.String("application/json"),
		}
		_, err := svc.PutObject(context.Background(), s3PutObject)
		if err != nil {
			return nil, err
		}
	}
	// Log some information
	logger.Info().
		Int("TotalFileCount", totalFiles).
		Int64("ArchiveSize", s3Object.ContentLength).
		Interface("S3Bucket", request.DestBucket).
		Msg("Expanded ZIP archive")

	// All good
	return nil, nil
}

// IAMPrivileges returns the IAM privs for this custom action
func (command *ZipToS3BucketResource) IAMPrivileges() []string {
	// Empty implementation - s3Site.go handles setting up the IAM privs for this.
	return []string{}
}

// Create implements the custom resource create operation
func (command ZipToS3BucketResource) Create(ctx context.Context, awsConfig awsv2.Config,
	event *CloudFormationLambdaEvent,
	logger *zerolog.Logger) (map[string]interface{}, error) {
	return command.unzip(awsConfig, event, logger)
}

// Update implements the custom resource update operation
func (command ZipToS3BucketResource) Update(ctx context.Context, awsConfig awsv2.Config,
	event *CloudFormationLambdaEvent,
	logger *zerolog.Logger) (map[string]interface{}, error) {
	return command.unzip(awsConfig, event, logger)
}

// Delete implements the custom resource delete operation
func (command ZipToS3BucketResource) Delete(ctx context.Context, awsConfig awsv2.Config,
	event *CloudFormationLambdaEvent,
	logger *zerolog.Logger) (map[string]interface{}, error) {
	request := ZipToS3BucketResourceRequest{}
	unmarshalErr := json.Unmarshal(event.ResourceProperties, &request)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	// Remove all objects from the bucket
	totalItemsDeleted := 0
	svc := awsv2S3.NewFromConfig(awsConfig)
	deleteItemsHandler := func(objectOutputs *awsv2S3.ListObjectsOutput, lastPage bool) bool {
		params := &awsv2S3.DeleteObjectsInput{
			Bucket: awsv2.String(request.DestBucket),
			Delete: &awsv2S3Types.Delete{ // Required
				Objects: []awsv2S3Types.ObjectIdentifier{},
				Quiet:   true,
			},
		}
		for _, eachObject := range objectOutputs.Contents {
			totalItemsDeleted++
			params.Delete.Objects = append(params.Delete.Objects,
				awsv2S3Types.ObjectIdentifier{
					Key: eachObject.Key,
				})
		}
		_, deleteResultErr := svc.DeleteObjects(context.Background(), params)
		return nil == deleteResultErr
	}

	// Walk the bucket and cleanup...
	params := &awsv2S3.ListObjectsInput{
		Bucket:  awsv2.String(request.DestBucket),
		MaxKeys: 1000,
	}
	listObjResponse, listObjectResponsErr := svc.ListObjects(context.Background(), params)
	if nil != listObjectResponsErr {
		return nil, listObjectResponsErr
	}
	// TODO - pages handler
	deleteItemsHandler(listObjResponse, true)
	// Cleanup the Manifest iff defined
	var deleteErr error
	if nil != request.Manifest {
		name := request.ManifestName
		if name == "" {
			name = DefaultManifestName
		}
		manifestDeleteParams := &awsv2S3.DeleteObjectInput{
			Bucket: awsv2.String(request.DestBucket),
			Key:    awsv2.String(name),
		}
		_, deleteErr = svc.DeleteObject(context.Background(), manifestDeleteParams)
		logger.Info().
			Int("TotalDeletedCount", totalItemsDeleted).
			Interface("S3Bucket", request.DestBucket).
			Msg("Purged S3 Bucket")
	}
	return nil, deleteErr
}
