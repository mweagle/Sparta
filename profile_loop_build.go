// +build !lambdabinary

package sparta

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	survey "github.com/AlecAivazis/survey/v2"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/google/pprof/driver"
	"github.com/google/pprof/profile"
	spartaAWS "github.com/mweagle/Sparta/aws"
	spartaCF "github.com/mweagle/Sparta/aws/cloudformation"
	gocf "github.com/mweagle/go-cloudformation"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type userAnswers struct {
	StackName            string `survey:"stackName"`
	StackInstance        string
	ProfileType          string `survey:"profileType"`
	DownloadNewSnapshots string `survey:"downloadNewSnapshots"`
	ProfileOptions       []string
	RefreshSnapshots     bool
}

func cachedProfileNames() []string {
	globPattern := filepath.Join(ScratchDirectory, "*.profile")
	matchingFiles, matchingFilesErr := filepath.Glob(globPattern)
	if matchingFilesErr != nil {
		return []string{}
	}
	// Just get the base name of the profile...
	cachedNames := []string{}
	for _, eachMatch := range matchingFiles {
		baseName := path.Base(eachMatch)
		filenameParts := strings.Split(baseName, ".")
		cachedNames = append(cachedNames, filenameParts[0])
	}
	return cachedNames
}

func askQuestions(userStackName string, stackNameToIDMap map[string]string) (*userAnswers, error) {
	stackNames := []string{}
	for eachKey := range stackNameToIDMap {
		stackNames = append(stackNames, eachKey)
	}
	sort.Strings(stackNames)
	cachedProfiles := cachedProfileNames()
	sort.Strings(cachedProfiles)

	var qs = []*survey.Question{
		{
			Name: "stackName",
			Prompt: &survey.Select{
				Message: "Which stack would you like to profile:",
				Options: stackNames,
				Default: userStackName,
			},
		},
		{
			Name: "profileType",
			Prompt: &survey.Select{
				Message: "What type of profile would you like to view?",
				Options: profileTypes,
				Default: profileTypes[0],
			},
		},
	}

	// Ask the known questions, figure out if they want to download a new
	// version of the snapshots...
	var responses userAnswers
	responseError := survey.Ask(qs, &responses)
	if responseError != nil {
		return nil, responseError
	}
	responses.StackInstance = stackNameToIDMap[responses.StackName]

	// Based on the first set, ask whether then want to download a new snapshot
	cachedProfileExists := strings.Contains(strings.Join(cachedProfiles, " "), responses.ProfileType)

	refreshCacheOptions := []string{}
	if cachedProfileExists {
		refreshCacheOptions = append(refreshCacheOptions, "Use cached snapshot")
	}
	refreshCacheOptions = append(refreshCacheOptions, "Download new snapshots from S3")
	var questionsRefresh = []*survey.Question{
		{
			Name: "downloadNewSnapshots",
			Prompt: &survey.Select{
				Message: "What profile snapshot(s) would you like to view?",
				Options: refreshCacheOptions,
				Default: refreshCacheOptions[0],
			},
		},
	}
	var refreshAnswers userAnswers
	refreshQuestionError := survey.Ask(questionsRefresh, &refreshAnswers)
	if refreshQuestionError != nil {
		return nil, refreshQuestionError
	}
	responses.RefreshSnapshots = (refreshAnswers.DownloadNewSnapshots == "Download new snapshots from S3")

	// Final set of questions regarding heap information
	// If this is a memory profile, what kind?
	if responses.ProfileType == "heap" {
		// the answers will be written to this struct
		heapAnswers := struct {
			Type string `survey:"type"`
		}{}
		// the questions to ask
		var heapQuestions = []*survey.Question{
			{
				Name: "type",
				Prompt: &survey.Select{
					Message: "Please select a heap profile type:",
					Options: []string{"inuse_space", "inuse_objects", "alloc_space", "alloc_objects"},
					Default: "inuse_space",
				},
			},
		}
		// perform the questions
		heapErr := survey.Ask(heapQuestions, &heapAnswers)
		if heapErr != nil {
			return nil, heapErr
		}
		responses.ProfileOptions = []string{fmt.Sprintf("-%s", heapAnswers.Type)}
	}
	return &responses, nil
}

func objectKeysForProfileType(profileType string,
	stackName string,
	s3BucketName string,
	maxCount int64,
	awsSession *session.Session,
	logger *zerolog.Logger) ([]string, error) {
	// http://weagle.s3.amazonaws.com/gosparta.io/pprof/SpartaPPropStack/profiles/cpu/cpu.42.profile

	// gosparta.io/pprof/SpartaPPropStack/profiles/cpu/cpu.42.profile
	// List all these...
	rootPath := profileSnapshotRootKeypathForType(profileType, stackName)
	listObjectInput := &s3.ListObjectsInput{
		Bucket: aws.String(s3BucketName),
		//	Delimiter: aws.String("/"),
		Prefix:  aws.String(rootPath),
		MaxKeys: aws.Int64(maxCount),
	}
	allItems := []string{}
	s3Svc := s3.New(awsSession)
	for {
		listItemResults, listItemResultsErr := s3Svc.ListObjects(listObjectInput)
		if listItemResultsErr != nil {
			return nil, errors.Wrapf(listItemResultsErr, "Attempting to list bucket: %s", s3BucketName)
		}
		for _, eachEntry := range listItemResults.Contents {
			logger.Debug().
				Str("FoundItem", *eachEntry.Key).
				Int64("Size", *eachEntry.Size).
				Msg("Profile file")
		}

		for _, eachItem := range listItemResults.Contents {
			if *eachItem.Size > 0 {
				allItems = append(allItems, *eachItem.Key)
			}
		}
		if int64(len(allItems)) >= maxCount || listItemResults.NextMarker == nil {
			return allItems, nil
		}
		listObjectInput.Marker = listItemResults.NextMarker
	}
}

////////////////////////////////////////////////////////////////////////////////
// Type returned from worker pool pulling down S3 snapshots
type downloadResult struct {
	err           error
	localFilePath string
}

func (dr *downloadResult) Error() error {
	return dr.err
}
func (dr *downloadResult) Result() interface{} {
	return dr.localFilePath
}

var _ workResult = (*downloadResult)(nil)

func downloaderTask(profileType string,
	stackName string,
	bucketName string,
	cacheRootPath string,
	downloadKey string,
	s3Service *s3.S3,
	downloader *s3manager.Downloader,
	logger *zerolog.Logger) taskFunc {

	return func() workResult {
		downloadInput := &s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(downloadKey),
		}
		cachedFilename := filepath.Join(cacheRootPath, filepath.Base(downloadKey))
		outputFile, outputFileErr := os.Create(cachedFilename)
		if outputFileErr != nil {
			return &downloadResult{
				err: outputFileErr,
			}
		}
		defer func() {
			closeErr := outputFile.Close()
			if closeErr != nil {
				logger.Warn().
					Err(closeErr).
					Msg("Failed to close output file writer")
			}
		}()

		_, downloadErr := downloader.Download(outputFile, downloadInput)
		// If we're all good, delete the one on s3...
		if downloadErr == nil {
			deleteObjectInput := &s3.DeleteObjectInput{
				Bucket: aws.String(bucketName),
				Key:    aws.String(downloadKey),
			}
			_, deleteErr := s3Service.DeleteObject(deleteObjectInput)
			if deleteErr != nil {
				logger.Warn().
					Err(deleteErr).
					Msg("Failed to delete S3 profile snapshot")
			} else {
				logger.Debug().
					Str("Bucket", bucketName).
					Str("Key", downloadKey).
					Msg("Deleted S3 profile")
			}
		}
		return &downloadResult{
			err:           downloadErr,
			localFilePath: outputFile.Name(),
		}
	}
}

func syncStackProfileSnapshots(profileType string,
	refreshSnapshots bool,
	stackName string,
	stackInstance string,
	s3BucketName string,
	awsSession *session.Session,
	logger *zerolog.Logger) ([]string, error) {
	s3KeyRoot := profileSnapshotRootKeypathForType(profileType, stackName)

	if !refreshSnapshots {
		cachedProfilePath := cachedAggregatedProfilePath(profileType)
		// Just used the cached ones...
		logger.Info().
			Str("CachedProfile", cachedProfilePath).
			Msg("Using cached profiles")

		// Make sure they exist...
		_, cachedInfoErr := os.Stat(cachedProfilePath)
		if os.IsNotExist(cachedInfoErr) {
			return nil, fmt.Errorf("no cache files found for profile type: %s. Please run again and fetch S3 artifacts", profileType)
		}
		return []string{cachedProfilePath}, nil
	}
	// Rebuild the cache...
	cacheRoot := cacheDirectoryForProfileType(profileType, stackName)
	logger.Info().
		Str("StackName", stackName).
		Str("S3Bucket", s3BucketName).
		Str("ProfileRootKey", s3KeyRoot).
		Str("Type", profileType).
		Str("CacheRoot", cacheRoot).
		Msg("Refreshing cached profiles")

	removeErr := os.RemoveAll(cacheRoot)
	if removeErr != nil {
		return nil, errors.Wrapf(removeErr, "Attempting delete local directory: %s", cacheRoot)
	}
	mkdirErr := os.MkdirAll(cacheRoot, os.ModePerm)
	if nil != mkdirErr {
		return nil, errors.Wrapf(mkdirErr, "Attempting to create local directory: %s", cacheRoot)
	}

	// Ok, let's get some user information
	s3Svc := s3.New(awsSession)
	downloader := s3manager.NewDownloader(awsSession)
	downloadKeys, downloadKeysErr := objectKeysForProfileType(profileType,
		stackName,
		s3BucketName,
		1024,
		awsSession,
		logger)

	if downloadKeys != nil {
		return nil, errors.Wrapf(downloadKeysErr,
			"Failed to determine pprof download keys")
	}
	downloadTasks := make([]*workTask, len(downloadKeys))
	for index, eachKey := range downloadKeys {
		taskFunc := downloaderTask(profileType,
			stackName,
			s3BucketName,
			cacheRoot,
			eachKey,
			s3Svc,
			downloader,
			logger)
		downloadTasks[index] = newWorkTask(taskFunc)
	}
	p := newWorkerPool(downloadTasks, 8)
	results, runErrors := p.Run()
	if len(runErrors) > 0 {
		return nil, fmt.Errorf("errors reported: %#v", runErrors)
	}

	// Read them all and merge them into a single profile...
	var accumulatedProfiles []*profile.Profile
	for _, eachResult := range results {
		profileFile := eachResult.(string)
		/* #nosec */
		profileInput, profileInputErr := os.Open(profileFile)
		if profileInputErr != nil {
			return nil, profileInputErr
		}
		parsedProfile, parsedProfileErr := profile.Parse(profileInput)
		// Ignore broken profiles
		if parsedProfileErr != nil {
			logger.Warn().
				Interface("Path", eachResult).
				Interface("Error", parsedProfileErr).
				Msg("Invalid cached profile")
		} else {
			logger.Info().
				Str("Input", profileFile).
				Msg("Aggregating profile")
			accumulatedProfiles = append(accumulatedProfiles, parsedProfile)
			profileInputCloseErr := profileInput.Close()
			if profileInputCloseErr != nil {
				logger.Warn().
					Err(profileInputCloseErr).
					Msg("Failed to close profile file writer")
			}
		}
	}
	logger.Info().
		Int("ProfileCount", len(accumulatedProfiles)).
		Msg("Consolidating profiles")

	if len(accumulatedProfiles) <= 0 {
		return nil, fmt.Errorf("unable to find %s snapshots in s3://%s for profile type: %s",
			stackName,
			s3BucketName,
			profileType)
	}

	// Great, merge them all
	consolidatedProfile, consolidatedProfileErr := profile.Merge(accumulatedProfiles)
	if consolidatedProfileErr != nil {
		return nil, fmt.Errorf("failed to merge profiles: %s", consolidatedProfileErr.Error())
	}
	// Write it out as the "canonical" path...
	consolidatedPath := cachedAggregatedProfilePath(profileType)
	logger.Info().
		Interface("ConsolidatedProfile", consolidatedPath).
		Msg("Creating consolidated profile")

	outputFile, outputFileErr := os.Create(consolidatedPath)
	if outputFileErr != nil {
		return nil, errors.Wrapf(outputFileErr,
			"failed to create consolidated file: %s", consolidatedPath)
	}
	writeErr := consolidatedProfile.Write(outputFile)
	if writeErr != nil {
		return nil, errors.Wrapf(writeErr,
			"failed to write profile: %s", consolidatedPath)
	}

	// Delete all the other ones, just return the consolidated one...
	for _, eachResult := range results {
		unlinkErr := os.Remove(eachResult.(string))
		if unlinkErr != nil {
			logger.Info().
				Str("File", consolidatedPath).
				Interface("Error", unlinkErr).
				Msg("Failed to delete file")
		}
		outputFileErr := outputFile.Close()
		if outputFileErr != nil {
			logger.Warn().
				Err(outputFileErr).
				Msg("Failed to close output file")
		}
	}
	return []string{consolidatedPath}, nil
}

// Profile is the interactive command used to pull S3 assets locally into /tmp
// and run ppro against the cached profiles
func Profile(serviceName string,
	serviceDescription string,
	s3BucketName string,
	httpPort int,
	logger *zerolog.Logger) error {

	awsSession := spartaAWS.NewSession(logger)

	// Get the currently active stacks...
	// Ref: http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-describing-stacks.html#w2ab2c15c15c17c11
	stackSummaries, stackSummariesErr := spartaCF.ListStacks(awsSession, 1024, "CREATE_COMPLETE",
		"UPDATE_COMPLETE",
		"UPDATE_ROLLBACK_COMPLETE")

	if stackSummariesErr != nil {
		return stackSummariesErr
	}
	// Get the stack names
	stackNameToIDMap := make(map[string]string)
	for _, eachSummary := range stackSummaries {
		stackNameToIDMap[*eachSummary.StackName] = *eachSummary.StackId
	}
	responses, responsesErr := askQuestions(serviceName, stackNameToIDMap)
	if responsesErr != nil {
		return responsesErr
	}

	// What does the user want to view?
	tempFilePaths, tempFilePathsErr := syncStackProfileSnapshots(responses.ProfileType,
		responses.RefreshSnapshots,
		responses.StackName,
		responses.StackInstance,
		s3BucketName,
		awsSession,
		logger)
	if tempFilePathsErr != nil {
		return tempFilePathsErr
	}
	// We can't hook the PProf webserver, so put some friendly output
	logger.Info().
		Msgf("Starting pprof webserver on http://localhost:%d. Enter Ctrl+C to exit.",
			httpPort)

	// Startup a server we manage s.t we can gracefully exit..
	newArgs := []string{os.Args[0]}
	newArgs = append(newArgs, responses.ProfileOptions...)
	newArgs = append(newArgs, "-http", fmt.Sprintf(":%d", httpPort), os.Args[0])
	newArgs = append(newArgs, tempFilePaths...)
	os.Args = newArgs
	return driver.PProf(&driver.Options{})
}

// ScheduleProfileLoop installs a profiling loop that pushes profile information
// to S3 for local consumption using a `profile` command that wraps
// pprof
func ScheduleProfileLoop(s3BucketArchive interface{},
	snapshotInterval time.Duration,
	cpuProfileDuration time.Duration,
	profileNames ...string) {

	// When we're building, we want a template decorator that will be called
	// by `provision`. This decorator will be responsible for:
	// ensuring each function has IAM creds (if the role isn't a string)
	// to write to the profile location and also pushing the
	// Stack name info as reseved environment variables into the function
	// execution context so that the AWS lambda version of this function
	// can quickly lookup the StackName and instance information ...
	profileDecorator = func(stackName string, info *LambdaAWSInfo, S3Bucket string, logger *zerolog.Logger) error {
		// If we have a role definition, ensure the function has rights to upload
		// to that bucket, with the limited ARN key
		logger.Info().
			Str("Function", info.lambdaFunctionName()).
			Msg("Instrumenting function for profiling")

		// The bucket is either a literal or a gocf.StringExpr - which one?
		var bucketValue gocf.Stringable
		if s3BucketArchive != nil {
			bucketValue = spartaCF.DynamicValueToStringExpr(s3BucketArchive)
		} else {
			bucketValue = gocf.String(S3Bucket)
		}

		// 1. Add the env vars to the map
		if info.Options.Environment == nil {
			info.Options.Environment = make(map[string]*gocf.StringExpr)
		}
		info.Options.Environment[envVarStackName] = gocf.Ref("AWS::StackName").String()
		info.Options.Environment[envVarStackInstanceID] = gocf.Ref("AWS::StackId").String()
		info.Options.Environment[envVarProfileBucketName] = bucketValue.String()

		// Update the IAM role...
		if info.RoleDefinition != nil {
			arn := gocf.Join("",
				gocf.String("arn:aws:s3:::"),
				bucketValue,
				gocf.String("/"),
				gocf.String(profileSnapshotRootKeypath(stackName)),
				gocf.String("/*"))

			info.RoleDefinition.Privileges = append(info.RoleDefinition.Privileges, IAMRolePrivilege{
				Actions:  []string{"s3:PutObject"},
				Resource: arn.String(),
			})
		}
		return nil
	}
}
