package resources

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"os"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	gocf "github.com/mweagle/go-cloudformation"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// DefaultManifestName is the name of the file that will be created
// at the root of the S3 bucket with user-supplied metadata
const DefaultManifestName = "MANIFEST.json"

// ZipToS3BucketResourceRequest is the data request made to a ZipToS3BucketResource
// lambda handler
type ZipToS3BucketResourceRequest struct {
	SrcBucket    *gocf.StringExpr
	SrcKeyName   *gocf.StringExpr
	DestBucket   *gocf.StringExpr
	ManifestName string
	Manifest     map[string]interface{}
}

// ZipToS3BucketResource manages populating an S3 bucket with the contents
// of a ZIP file...
type ZipToS3BucketResource struct {
	gocf.CloudFormationCustomResource
	ZipToS3BucketResourceRequest
}

func (command ZipToS3BucketResource) unzip(session *session.Session,
	event *CloudFormationLambdaEvent,
	logger *logrus.Logger) (map[string]interface{}, error) {

	unmarshalErr := json.Unmarshal(event.ResourceProperties, &command)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	// Fetch the ZIP contents and unpack them to the S3 bucket
	svc := s3.New(session)
	s3Object, s3ObjectErr := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(command.SrcBucket.Literal),
		Key:    aws.String(command.SrcKeyName.Literal),
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
			s3PutObject := &s3.PutObjectInput{
				Body:        bytes.NewReader(bodySource),
				Bucket:      aws.String(command.DestBucket.Literal),
				Key:         aws.String(fmt.Sprintf("/%s", eachFile.Name)),
				ContentType: aws.String(mimeType),
			}
			_, err := svc.PutObject(s3PutObject)
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
	if nil != command.Manifest {
		manifestBytes, manifestErr := json.Marshal(command.Manifest)
		if nil != manifestErr {
			return nil, manifestErr
		}
		name := command.ManifestName
		if name == "" {
			name = DefaultManifestName
		}
		s3PutObject := &s3.PutObjectInput{
			Body:        bytes.NewReader(manifestBytes),
			Bucket:      aws.String(command.DestBucket.Literal),
			Key:         aws.String(name),
			ContentType: aws.String("application/json"),
		}
		_, err := svc.PutObject(s3PutObject)
		if err != nil {
			return nil, err
		}
	}
	// Log some information
	logger.WithFields(logrus.Fields{
		"TotalFileCount": totalFiles,
		"ArchiveSize":    *s3Object.ContentLength,
		"S3Bucket":       command.DestBucket,
	}).Info("Expanded ZIP archive")

	// All good
	return nil, nil
}

// IAMPrivileges returns the IAM privs for this custom action
func (command *ZipToS3BucketResource) IAMPrivileges() []string {
	// Empty implementation - s3Site.go handles setting up the IAM privs for this.
	return []string{}
}

// Create implements the custom resource create operation
func (command ZipToS3BucketResource) Create(awsSession *session.Session,
	event *CloudFormationLambdaEvent,
	logger *logrus.Logger) (map[string]interface{}, error) {
	return command.unzip(awsSession, event, logger)
}

// Update implements the custom resource update operation
func (command ZipToS3BucketResource) Update(awsSession *session.Session,
	event *CloudFormationLambdaEvent,
	logger *logrus.Logger) (map[string]interface{}, error) {
	return command.unzip(awsSession, event, logger)
}

// Delete implements the custom resource delete operation
func (command ZipToS3BucketResource) Delete(awsSession *session.Session,
	event *CloudFormationLambdaEvent,
	logger *logrus.Logger) (map[string]interface{}, error) {

	unmarshalErr := json.Unmarshal(event.ResourceProperties, &command)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	// Remove all objects from the bucket
	totalItemsDeleted := 0
	svc := s3.New(awsSession)
	deleteItemsHandler := func(objectOutputs *s3.ListObjectsOutput, lastPage bool) bool {
		params := &s3.DeleteObjectsInput{
			Bucket: aws.String(command.DestBucket.Literal),
			Delete: &s3.Delete{ // Required
				Objects: []*s3.ObjectIdentifier{},
				Quiet:   aws.Bool(true),
			},
		}
		for _, eachObject := range objectOutputs.Contents {
			totalItemsDeleted++
			params.Delete.Objects = append(params.Delete.Objects, &s3.ObjectIdentifier{
				Key: eachObject.Key,
			})
		}
		_, deleteResultErr := svc.DeleteObjects(params)
		return nil == deleteResultErr
	}

	// Walk the bucket and cleanup...
	params := &s3.ListObjectsInput{
		Bucket:  aws.String(command.DestBucket.Literal),
		MaxKeys: aws.Int64(1000),
	}
	err := svc.ListObjectsPages(params, deleteItemsHandler)
	if nil != err {
		return nil, err
	}

	// Cleanup the Manifest iff defined
	var deleteErr error
	if nil != command.Manifest {
		name := command.ManifestName
		if name == "" {
			name = DefaultManifestName
		}
		manifestDeleteParams := &s3.DeleteObjectInput{
			Bucket: aws.String(command.DestBucket.Literal),
			Key:    aws.String(name),
		}
		_, deleteErr = svc.DeleteObject(manifestDeleteParams)
		logger.WithFields(logrus.Fields{
			"TotalDeletedCount": totalItemsDeleted,
			"S3Bucket":          command.DestBucket,
		}).Info("Purged S3 Bucket")
	}
	return nil, deleteErr
}
