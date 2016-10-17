package s3

import (
	"fmt"
	"os"

	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"strings"
)

// RollbackFunction called in the event of a stack provisioning failure
type RollbackFunction func(logger *logrus.Logger) error

// CreateS3RollbackFunc creates an S3 rollback function that attempts to delete a previously
// uploaded item.
func CreateS3RollbackFunc(awsSession *session.Session, s3Bucket string, s3Key string) RollbackFunction {
	return func(logger *logrus.Logger) error {
		logger.Info("Attempting to cleanup S3 item: ", s3Key)
		s3Client := s3.New(awsSession)
		params := &s3.DeleteObjectInput{
			Bucket: aws.String(s3Bucket),
			Key:    aws.String(s3Key),
		}
		_, err := s3Client.DeleteObject(params)
		if err != nil {
			logger.WithFields(logrus.Fields{
				"Error": err,
			}).Warn("Failed to delete S3 item during rollback cleanup")
		} else {
			logger.WithFields(logrus.Fields{
				"Bucket": s3Bucket,
				"Key":    s3Key,
			}).Debug("Item deleted during rollback cleanup")
		}
		return err
	}
}

// UploadLocalFileToS3 takes a local path and uploads the content at localPath
// to the given S3Bucket and KeyPrefix.  The final S3 keyname is the S3KeyPrefix+
// the basename of the localPath.
func UploadLocalFileToS3(localPath string,
	awsSession *session.Session,
	S3Bucket string,
	S3KeyPrefix string,
	logger *logrus.Logger) (string, error) {

	// Then do the actual work
	reader, err := os.Open(localPath)
	if nil != err {
		return "", fmt.Errorf("Failed to open local archive for S3 upload: %s", err.Error())
	}

	// Make sure the key prefix ends with a trailing slash
	canonicalKeyPrefix := S3KeyPrefix
	if !strings.HasSuffix(canonicalKeyPrefix, "/") {
		canonicalKeyPrefix += "/"
	}

	// Cache it in case there was an error & we need to cleanup
	keyName := fmt.Sprintf("%s%s", canonicalKeyPrefix, filepath.Base(localPath))

	uploadInput := &s3manager.UploadInput{
		Bucket:      &S3Bucket,
		Key:         &keyName,
		ContentType: aws.String("application/zip"),
		Body:        reader,
	}
	logger.WithFields(logrus.Fields{
		"Source": localPath,
	}).Info("Uploading local file to S3")
	uploader := s3manager.NewUploader(awsSession)
	result, err := uploader.Upload(uploadInput)
	if nil != err {
		return "", err
	}
	logger.WithFields(logrus.Fields{
		"URL": result.Location,
	}).Info("Upload complete")

	return keyName, nil
}
