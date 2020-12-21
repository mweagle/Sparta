package sparta

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS
////////////////////////////////////////////////////////////////////////////////

var (
	// SpartaTagBuildIDKey is the keyname used in the CloudFormation Output
	// that stores the user-supplied or automatically generated BuildID
	// for this run
	SpartaTagBuildIDKey = spartaTagName("buildId")

	// SpartaTagBuildTagsKey is the keyname used in the CloudFormation Output
	// that stores the optional user-supplied golang build tags
	SpartaTagBuildTagsKey = spartaTagName("buildTags")
)

const (
	// MetadataParamCloudFormationStackPath is the path to the template
	MetadataParamCloudFormationStackPath = "CloudFormationStackPath"
	// MetadataParamServiceName is the name of the stack to use
	MetadataParamServiceName = "ServiceName"
	// MetadataParamS3Bucket is the Metadata param we use for the bucket
	MetadataParamS3Bucket = "ArtifactS3Bucket"

	// Metadata params for a ZIP archive
	//

	// MetadataParamCodeArchivePath is the intemediate local path to the code
	MetadataParamCodeArchivePath = "CodeArchivePath"
	// MetadataParamS3SiteArchivePath is the intemediate local path to the S3 site contents
	MetadataParamS3SiteArchivePath = "S3SiteArtifactPath"

	// Metadata params for OCI builds
	//

	// MetadataParamECRTag is the locally tagged Docker image to push
	MetadataParamECRTag = "ECRTag"
)

const (
	// StackParamS3CodeKeyName is the Stack Parameter to the S3 key of the uploaded asset
	StackParamS3CodeKeyName = "CodeArtifactS3Key"
	// StackParamArtifactBucketName is where we uploaded the artifact to
	StackParamArtifactBucketName = MetadataParamS3Bucket
	// StackParamS3CodeVersion is the object version to use for the S3 item
	StackParamS3CodeVersion = "CodeArtifactS3ObjectVersion"
	// StackParamS3SiteArchiveKey is the param to the S3 archive for a static website.
	StackParamS3SiteArchiveKey = "SiteArtifactS3Key"
	// StackParamS3SiteArchiveVersion is the version of the S3 artifact to use
	StackParamS3SiteArchiveVersion = "SiteArtifactS3ObjectVersion"
	// StackParamCodeImageURI is the ImageURI to the uploaded image
	StackParamCodeImageURI = "CodeImageURI"
)

const (
	// StackOutputBuildTime is the Output param for when this template was built
	StackOutputBuildTime = "TemplateCreationTime"
	// StackOutputBuildID is the Output tag that holds the build id
	StackOutputBuildID = "BuildID"
)

func showOptionalAWSUsageInfo(err error, logger *zerolog.Logger) {
	if err == nil {
		return
	}
	userAWSErr, userAWSErrOk := err.(awserr.Error)
	if userAWSErrOk {
		if strings.Contains(userAWSErr.Error(), "could not find region configuration") {
			logger.Error().Msg("")
			logger.Error().Msg("Consider setting env.AWS_REGION, env.AWS_DEFAULT_REGION, or env.AWS_SDK_LOAD_CONFIG to resolve this issue.")
			logger.Error().Msg("See the documentation at https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html for more information.")
			logger.Error().Msg("")
		}
	}
}

func spartaTagName(baseKey string) string {
	return fmt.Sprintf("io:sparta:%s", baseKey)
}

// Sanitize the provided input by replacing illegal characters with underscores
func sanitizedName(input string) string {
	return reSanitize.ReplaceAllString(input, "_")
}

type pipelineBaseOp interface {
	Invoke(context.Context, *zerolog.Logger) error
	Rollback(context.Context, *zerolog.Logger) error
}

type pipelineStageBase interface {
	Run(context.Context, *zerolog.Logger) error
	Append(string, pipelineBaseOp) pipelineStageBase
	Rollback(context.Context, *zerolog.Logger) error
}

type pipelineStageOpEntry struct {
	opName string
	op     pipelineBaseOp
}
type pipelineStage struct {
	ops []*pipelineStageOpEntry
}

func (ps *pipelineStage) Append(opName string, op pipelineBaseOp) pipelineStageBase {
	ps.ops = append(ps.ops, &pipelineStageOpEntry{
		opName: opName,
		op:     op,
	})
	return ps
}

func (ps *pipelineStage) Run(ctx context.Context, logger *zerolog.Logger) error {
	var wg sync.WaitGroup
	var mapErr sync.Map

	for eachIndex, eachEntry := range ps.ops {
		wg.Add(1)
		go func(opIndex int, opEntry *pipelineStageOpEntry, goLogger *zerolog.Logger) {
			defer wg.Done()
			opErr := opEntry.op.Invoke(ctx, goLogger)
			if opErr != nil {
				mapErr.Store(opEntry.opName, opErr)
			}
		}(eachIndex, eachEntry, logger)
	}
	wg.Wait()

	// Were there any errors?
	errorText := []string{}
	mapErr.Range(func(key interface{}, value interface{}) bool {
		errorText = append(errorText, fmt.Sprintf("%s=>%v",
			key,
			value))
		return true
	})
	if len(errorText) != 0 {
		return errors.New(strings.Join(errorText, ", "))
	}
	return nil
}

func (ps *pipelineStage) Rollback(ctx context.Context, logger *zerolog.Logger) error {
	// Ok, another wg to async cleanup everything. Operations
	// need to be a bit stateful for this...
	var wgRollback sync.WaitGroup
	logger.Debug().Msgf("Rolling back %T due to errors", ps)
	for _, eachEntry := range ps.ops {
		wgRollback.Add(1)
		go func(opEntry *pipelineStageOpEntry, goLogger *zerolog.Logger) {
			defer wgRollback.Done()
			opErr := opEntry.op.Rollback(ctx, goLogger)
			if opErr != nil {
				goLogger.Warn().Msgf("Operation (%s) rollback failed: %s", opEntry.opName, opErr)
			}
		}(eachEntry, logger)
	}
	wgRollback.Wait()
	return nil
}

type pipelineStageEntry struct {
	stageName string
	stage     pipelineStageBase
	duration  time.Duration
}

type pipeline struct {
	stages    []*pipelineStageEntry
	startTime time.Time
}

func (p *pipeline) Append(stageName string, stage pipelineStageBase) *pipeline {
	p.stages = append(p.stages, &pipelineStageEntry{
		stageName: stageName,
		stage:     stage,
	})
	return p
}

func (p *pipeline) Run(ctx context.Context,
	name string,
	logger *zerolog.Logger) error {

	p.startTime = time.Now()

	// Run the stages, if there is an error, rollback
	for stageIndex, curStage := range p.stages {
		startTime := time.Now()
		stageErr := curStage.stage.Run(ctx, logger)
		if stageErr != nil {
			logger.Error().Msgf("Pipeline stage %s failed", curStage.stageName)

			for index := stageIndex; index >= 0; index-- {
				rollbackErr := p.stages[index].stage.Rollback(ctx, logger)
				if rollbackErr != nil {
					logger.Warn().Msgf("Pipeline stage %s failed to Rollback", curStage.stageName)
				}
			}
			return stageErr
		}
		curStage.duration = time.Since(startTime)
	}

	// Log the total stage execution times...
	logger.Debug().Msg(headerDivider)
	for _, eachStageEntry := range p.stages {
		logger.Debug().
			Str("Name", eachStageEntry.stageName).
			Str("Duration", eachStageEntry.duration.String()).
			Msg("Stage duration")
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// Common stages
////////////////////////////////////////////////////////////////////////////////

type userFunctionRollbackOp struct {
	serviceName   string
	awsSession    *session.Session
	noop          bool
	rollbackFuncs []RollbackHookHandler
}

func (ufro *userFunctionRollbackOp) Rollback(ctx context.Context, logger *zerolog.Logger) error {
	wg := sync.WaitGroup{}

	for _, eachRollbackHook := range ufro.rollbackFuncs {
		wg.Add(1)
		go func(ctx context.Context,
			handler RollbackHookHandler,
			serviceName string,
			awsSession *session.Session,
			noop bool,
			logger *zerolog.Logger) {
			// Decrement the counter when the goroutine completes.
			defer wg.Done()
			_, rollbackErr := handler.Rollback(ctx,
				serviceName,
				awsSession,
				noop,
				logger)
			if rollbackErr != nil {
				logger.Warn().
					Err(rollbackErr).
					Str("Function", fmt.Sprintf("%T", handler)).
					Msg("Rollback function failed")
			}
		}(ctx,
			eachRollbackHook,
			ufro.serviceName,
			ufro.awsSession,
			ufro.noop,
			logger)
	}
	wg.Wait()
	return nil
}
func (ufro *userFunctionRollbackOp) Invoke(ctx context.Context, logger *zerolog.Logger) error {

	return nil
}

func newUserRollbackEnabledPipeline(serviceName string,
	awsSession *session.Session,
	rollbackFuncs []RollbackHookHandler,
	noop bool) *pipeline {

	buildPipeline := &pipeline{}

	// Verify
	rollbackStateUserFunctions := &pipelineStage{}
	rollbackStateUserFunctions.Append("userRollbackFunctions", &userFunctionRollbackOp{
		serviceName:   serviceName,
		awsSession:    awsSession,
		noop:          noop,
		rollbackFuncs: rollbackFuncs,
	})
	buildPipeline.Append("userRollbackFunctions", rollbackStateUserFunctions)
	return buildPipeline
}
