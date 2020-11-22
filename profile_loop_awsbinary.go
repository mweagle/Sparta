// +build lambdabinary

package sparta

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime/pprof"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	spartaAWS "github.com/mweagle/Sparta/aws"
	"github.com/rs/zerolog"
)

var currentSlot int
var stackName string
var profileBucket string

const snapshotCount = 3

func nextUploadSlot() int {
	uploadSlot := currentSlot
	currentSlot = (currentSlot + 1) % snapshotCount
	return uploadSlot
}

func init() {
	currentSlot = 0
	// These correspond to the environment variables that were published
	// into the Lambda environment by the profile decorator
	stackName = os.Getenv(envVarStackName)
	profileBucket = os.Getenv(envVarProfileBucketName)
}

func profileOutputFile(basename string) (*os.File, error) {
	fileName := fmt.Sprintf("%s.%s.profile", basename, InstanceID())
	// http://docs.aws.amazon.com/lambda/latest/dg/current-supported-versions.html
	if os.Getenv("_LAMBDA_SERVER_PORT") != "" {
		fileName = filepath.Join("/tmp", fileName)
	}
	return os.Create(fileName)
}

////////////////////////////////////////////////////////////////////////////////
// Type returned from worker pool uploading profiles to S3
type uploadResult struct {
	err      error
	uploaded bool
}

func (ur *uploadResult) Error() error {
	return ur.err
}
func (ur *uploadResult) Result() interface{} {
	return ur.uploaded
}

func uploadFileTask(uploader *s3manager.Uploader,
	profileType string,
	uploadSlot int,
	localFilePath string,
	logger *zerolog.Logger) taskFunc {
	return func() workResult {
		fileReader, fileReaderErr := os.Open(localFilePath)
		if fileReaderErr != nil {
			return &uploadResult{err: fileReaderErr}
		}
		defer fileReader.Close()
		defer os.Remove(localFilePath)

		uploadFileName := fmt.Sprintf("%d-%s", uploadSlot, path.Base(localFilePath))
		keyPath := path.Join(profileSnapshotRootKeypathForType(profileType, stackName), uploadFileName)
		uploadInput := &s3manager.UploadInput{
			Bucket: aws.String(profileBucket),
			Key:    aws.String(keyPath),
			Body:   fileReader,
		}
		uploadOutput, uploadErr := uploader.Upload(uploadInput)
		return &uploadResult{
			err:      uploadErr,
			uploaded: uploadOutput != nil,
		}
	}
}

func snapshotProfiles(s3BucketArchive interface{},
	snapshotInterval time.Duration,
	cpuProfileDuration time.Duration,
	profileTypes ...string) {

	// The session the S3 Uploader will use
	profileLogger, _ := NewLogger("")

	publishProfiles := func(cpuProfilePath string) {

		profileLogger.Info().
			Str("CPUProfilePath", cpuProfilePath).
			Interface("Types", profileTypes).
			Msg("Publishing CPU profile")

		uploadSlot := nextUploadSlot()
		sess := spartaAWS.NewSession(profileLogger)
		uploader := s3manager.NewUploader(sess)
		uploadTasks := make([]*workTask, 0)

		if cpuProfilePath != "" {
			uploadTasks = append(uploadTasks,
				newWorkTask(uploadFileTask(uploader,
					"cpu",
					uploadSlot,
					cpuProfilePath,
					profileLogger)))
		}
		for _, eachProfileType := range profileTypes {
			namedProfile := pprof.Lookup(eachProfileType)
			if namedProfile != nil {
				outputProfile, outputFileErr := profileOutputFile(eachProfileType)
				if outputFileErr != nil {

					profileLogger.Error().
						Err(outputFileErr).
						Msg("Failed to CPU profile file")
				} else {
					namedProfile.WriteTo(outputProfile, 0)
					outputProfile.Close()
					uploadTasks = append(uploadTasks,
						newWorkTask(uploadFileTask(uploader,
							eachProfileType,
							uploadSlot,
							outputProfile.Name(),
							profileLogger)))
				}
			}
		}
		workerPool := newWorkerPool(uploadTasks, 32)
		workerPool.Run()
		ScheduleProfileLoop(s3BucketArchive,
			snapshotInterval,
			cpuProfileDuration,
			profileTypes...)
	}

	if cpuProfileDuration != 0 {
		outputFile, outputFileErr := profileOutputFile("cpu")
		if outputFileErr != nil {
			profileLogger.Warn().
				Err(outputFileErr).
				Msg("Failed to create cpu profile path")
			return
		}
		startErr := pprof.StartCPUProfile(outputFile)
		if startErr != nil {
			profileLogger.Warn().
				Err(startErr).
				Msg("Failed to start CPU profile")
		}
		profileLogger.Info().Msg("Opened CPU profile")
		time.AfterFunc(cpuProfileDuration, func() {
			pprof.StopCPUProfile()
			profileLogger.Info().Msg("Opened CPU profile")
			closeErr := outputFile.Close()
			if closeErr != nil {
				profileLogger.Warn().
					Err(closeErr).
					Msg("Failed to close CPU profile output")
			} else {
				publishProfiles(outputFile.Name())
			}
		})
	} else {
		publishProfiles("")
	}
}

// ScheduleProfileLoop installs a profiling loop that pushes profile information
// to S3 for local consumption using a `profile` command that wraps
// pprof
func ScheduleProfileLoop(s3BucketArchive interface{},
	snapshotInterval time.Duration,
	cpuProfileDuration time.Duration,
	profileTypes ...string) {

	time.AfterFunc(snapshotInterval, func() {
		snapshotProfiles(s3BucketArchive, snapshotInterval, cpuProfileDuration, profileTypes...)
	})
}
