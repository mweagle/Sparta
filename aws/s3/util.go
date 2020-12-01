package s3

import (
	"fmt"
	"mime"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	humanize "github.com/dustin/go-humanize"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// RollbackFunction called in the event of a stack provisioning failure
type RollbackFunction func(logger *zerolog.Logger) error

// CreateS3RollbackFunc creates an S3 rollback function that attempts to delete a previously
// uploaded item. Note that s3ArtifactURL may include a `versionId` query arg
// to denote the specific version to delete.
func CreateS3RollbackFunc(awsSession *session.Session, s3ArtifactURL string) RollbackFunction {
	return func(logger *zerolog.Logger) error {
		logger.Info().
			Str("URL", s3ArtifactURL).
			Msg("Deleting S3 object")
		artifactURLParts, artifactURLPartsErr := url.Parse(s3ArtifactURL)
		if nil != artifactURLPartsErr {
			return artifactURLPartsErr
		}
		// Bucket is the first component
		s3Bucket := strings.Split(artifactURLParts.Host, ".")[0]
		s3Client := s3.New(awsSession)
		params := &s3.DeleteObjectInput{
			Bucket: aws.String(s3Bucket),
			Key:    aws.String(artifactURLParts.Path),
		}
		versionID := artifactURLParts.Query().Get("versionId")
		if versionID != "" {
			params.VersionId = aws.String(versionID)
		}
		_, err := s3Client.DeleteObject(params)
		if err != nil {
			logger.Warn().
				Err(err).
				Msg("Failed to delete S3 item during rollback cleanup")
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
	S3KeyName string,
	logger *zerolog.Logger) (string, error) {

	// Then do the actual work
	/* #nosec */
	reader, err := os.Open(localPath)
	if nil != err {
		return "", fmt.Errorf("failed to open file for S3 upload: %s", err.Error())
	}
	uploadInput := &s3manager.UploadInput{
		Bucket:      &S3Bucket,
		Key:         &S3KeyName,
		ContentType: aws.String(mime.TypeByExtension(path.Ext(localPath))),
		Body:        reader,
	}
	// Ensure we close the reader...
	defer func() {
		closeErr := reader.Close()
		if closeErr != nil {
			logger.Warn().
				Err(closeErr).
				Msg("Failed to close upload Body reader input")
		}
	}()
	// If we can get the current working directory, let's try and strip
	// it from the path just to keep the log statement a bit shorter
	logPath := localPath
	cwd, cwdErr := os.Getwd()
	if cwdErr == nil {
		logPath = strings.TrimPrefix(logPath, cwd)
		if logPath != localPath {
			logPath = fmt.Sprintf(".%s", logPath)
		}
	}
	// Binary size
	stat, err := os.Stat(localPath)
	if err != nil {
		return "", fmt.Errorf("failed to calculate upload size for file: %s", localPath)
	}
	logger.Info().
		Str("Path", logPath).
		Str("Bucket", S3Bucket).
		Str("Key", S3KeyName).
		Str("Size", humanize.Bytes(uint64(stat.Size()))).
		Msg("Uploading")

	uploader := s3manager.NewUploader(awsSession)
	result, err := uploader.Upload(uploadInput)
	if nil != err {
		return "", errors.Wrapf(err, "Failed to upload object to S3")
	}
	if result.VersionID != nil {

		logger.Debug().
			Str("URL", result.Location).
			Str("VersionID", string(*result.VersionID)).
			Msg("S3 upload complete")

	} else {
		logger.Debug().
			Str("URL", result.Location).
			Msg("S3 upload complete")
	}
	locationURL := result.Location
	if nil != result.VersionID {
		// http://docs.aws.amazon.com/AmazonS3/latest/dev/RetrievingObjectVersions.html
		locationURL = fmt.Sprintf("%s?versionId=%s", locationURL, string(*result.VersionID))
	}
	return locationURL, nil
}

// BucketVersioningEnabled determines if a given S3 bucket has object
// versioning enabled.
func BucketVersioningEnabled(awsSession *session.Session,
	S3Bucket string,
	logger *zerolog.Logger) (bool, error) {

	s3Svc := s3.New(awsSession)
	params := &s3.GetBucketVersioningInput{
		Bucket: aws.String(S3Bucket), // Required
	}
	versioningEnabled := false
	resp, err := s3Svc.GetBucketVersioning(params)
	if err == nil && resp != nil && resp.Status != nil {
		// What's the versioning policy?
		logger.Debug().
			Interface("VersionPolicy", *resp).
			Str("BucketName", S3Bucket).
			Msg("Bucket version policy")
		versioningEnabled = (strings.ToLower(*resp.Status) == "enabled")
	}
	return versioningEnabled, err
}

// BucketRegion returns the AWS region that hosts the bucket
func BucketRegion(awsSession *session.Session,
	S3Bucket string,
	logger *zerolog.Logger) (string, error) {
	regionHint := ""
	if awsSession.Config.Region != nil {
		regionHint = *awsSession.Config.Region
	}
	awsContext := aws.BackgroundContext()
	return s3manager.GetBucketRegion(awsContext,
		awsSession,
		S3Bucket,
		regionHint)
}
