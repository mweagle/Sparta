package s3

import (
	"context"
	"fmt"
	"mime"
	"net/url"
	"os"
	"path"
	"strings"

	awsv2S3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"

	awsv2S3Manager "github.com/aws/aws-sdk-go-v2/feature/s3/manager"

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	awsv2S3 "github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// RollbackFunction called in the event of a stack provisioning failure
type RollbackFunction func(logger *zerolog.Logger) error

// CreateS3RollbackFunc creates an S3 rollback function that attempts to delete a previously
// uploaded item. Note that s3ArtifactURL may include a `versionId` query arg
// to denote the specific version to delete.
func CreateS3RollbackFunc(awsConfig awsv2.Config, s3ArtifactURL string) RollbackFunction {
	return func(logger *zerolog.Logger) error {
		rollbackContext := context.Background()

		logger.Info().
			Str("URL", s3ArtifactURL).
			Msg("Deleting S3 object")
		artifactURLParts, artifactURLPartsErr := url.Parse(s3ArtifactURL)
		if nil != artifactURLPartsErr {
			return artifactURLPartsErr
		}
		// Bucket is the first component
		s3Bucket := strings.Split(artifactURLParts.Host, ".")[0]
		s3Client := awsv2S3.NewFromConfig(awsConfig)
		params := &awsv2S3.DeleteObjectInput{
			Bucket: awsv2.String(s3Bucket),
			Key:    awsv2.String(artifactURLParts.Path),
		}
		versionID := artifactURLParts.Query().Get("versionId")
		if versionID != "" {
			params.VersionId = awsv2.String(versionID)
		}
		_, err := s3Client.DeleteObject(rollbackContext, params)
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
func UploadLocalFileToS3(ctx context.Context,
	localPath string,
	awsConfig awsv2.Config,
	S3Bucket string,
	S3KeyName string,
	logger *zerolog.Logger) (string, error) {

	// Then do the actual work
	/* #nosec */
	reader, err := os.Open(localPath)
	if nil != err {
		return "", fmt.Errorf("failed to open file for S3 upload: %s", err.Error())
	}
	uploadInput := &awsv2S3.PutObjectInput{
		Bucket:      &S3Bucket,
		Key:         &S3KeyName,
		ContentType: awsv2.String(mime.TypeByExtension(path.Ext(localPath))),
		Body:        reader,
	}
	// Ensure we close the reader...
	/* #nosec */
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
		Int64("Size", stat.Size()).
		Msg("Uploading")

	s3Svc := awsv2S3.NewFromConfig(awsConfig)
	uploader := awsv2S3Manager.NewUploader(s3Svc)

	result, err := uploader.Upload(ctx, uploadInput)
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
func BucketVersioningEnabled(ctx context.Context,
	awsConfig awsv2.Config,
	S3Bucket string,
	logger *zerolog.Logger) (bool, error) {

	s3Svc := awsv2S3.NewFromConfig(awsConfig)
	params := &awsv2S3.GetBucketVersioningInput{
		Bucket: awsv2.String(S3Bucket), // Required
	}
	versioningEnabled := false
	resp, err := s3Svc.GetBucketVersioning(ctx, params)
	if err == nil && resp != nil && resp.Status != "" {
		// What's the versioning policy?
		logger.Debug().
			Interface("VersionPolicy", *resp).
			Str("BucketName", S3Bucket).
			Msg("Bucket version policy")
		versioningEnabled = (resp.Status == awsv2S3Types.BucketVersioningStatusEnabled)
	}
	return versioningEnabled, err
}

// BucketRegion returns the AWS region that hosts the bucket
func BucketRegion(ctx context.Context,
	awsConfig awsv2.Config,
	S3Bucket string,
	logger *zerolog.Logger) (string, error) {
	s3Svc := awsv2S3.NewFromConfig(awsConfig)
	return awsv2S3Manager.GetBucketRegion(ctx,
		s3Svc,
		S3Bucket)
}
