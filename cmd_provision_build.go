//go:build !lambdabinary
// +build !lambdabinary

package sparta

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/lambda"
	gofroot "github.com/awslabs/goformation"
	gof "github.com/awslabs/goformation/v5/cloudformation"
	spartaAWS "github.com/mweagle/Sparta/aws"
	spartaCF "github.com/mweagle/Sparta/aws/cloudformation"
	spartaS3 "github.com/mweagle/Sparta/aws/s3"
	spartaDocker "github.com/mweagle/Sparta/docker"
	"github.com/mweagle/Sparta/system"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

const (
	s3UploadCloudFormationStackKey = "cloudformationStackS3URL"
)

////////////////////////////////////////////////////////////////////////////////
// Type that encapsulates an S3 URL with accessors to return either the
// full URL or just the valid S3 Keyname
type s3UploadURL struct {
	location string
	path     string
	version  string
}

// func (s3URL *s3UploadURL) keyName() string {
// 	return s3URL.path
// }

func newS3UploadURL(s3URL string) *s3UploadURL {
	urlParts, urlPartsErr := url.Parse(s3URL)
	if nil != urlPartsErr {
		return nil
	}
	queryParams, queryParamsErr := url.ParseQuery(urlParts.RawQuery)
	if nil != queryParamsErr {
		return nil
	}
	versionIDValues := queryParams["versionId"]
	version := ""
	if len(versionIDValues) == 1 {
		version = versionIDValues[0]
	}
	return &s3UploadURL{location: s3URL,
		path:    strings.TrimPrefix(urlParts.Path, "/"),
		version: version}
}

////////////////////////////////////////////////////////////////////////////////
// TemplateMetadataReader encapsulates a reader of Metadata from the
// template
type templateMetadataReader struct {
	template *gof.Template
}

func (tmr *templateMetadataReader) AsString(keyName string) (string, error) {
	val, valExists := tmr.template.Metadata[keyName]
	if !valExists {
		return "", nil
	}
	typedVal, typedValOk := val.(string)
	if !typedValOk {
		return "", errors.Errorf("Failed to convert %#v to string", val)
	}
	return typedVal, nil
}

////////////////////////////////////////////////////////////////////////////////
// Represents data associated with provisioning the S3 Site iff defined
type s3SiteContext struct {
	s3Site *S3Site
}

// provisionContext is data that is mutated during the provisioning workflow
type provisionContext struct {
	serviceName string
	// AWS Session to be used for all API calls made in the process of provisioning
	// this service.
	awsSession *session.Session
	// Path to cfTemplate
	cfTemplatePath string
	// CloudFormation Template
	cfTemplate *gof.Template
	// Is the S3 bucket version enabled?
	isVersionAwareBucket bool
	// Is this a NOOP operation?
	noop bool
	// s3URLS that have been uploaded...
	s3Uploads map[string]*s3UploadURL
	// stack that we mutated
	stack *cloudformation.Stack
	// the code pipeline trigger
	codePipelineTrigger string
	// stack parameters supplied to the template. These will be upserted
	// to get either the user supplied, metadata tunneled value.
	stackParameterValues map[string]string
	// additional stack tags for the provisioned stack
	stackTags map[string]string
	// Is this inplace udpates?
	inPlaceUpdates bool
}

////////////////////////////////////////////////////////////////////////////////
// Private - START
//

// maximumStackOperationTimeout returns the timeout
// value to use for a stack operation based on the type
// of resources that it provisions. In general the timeout
// is short with an exception made for CloudFront
// distributions
func maximumStackOperationTimeout(template *gof.Template, logger *zerolog.Logger) time.Duration {
	stackOperationTimeout := 20 * time.Minute
	// If there is a CloudFront distributon in there then
	// let's give that a bit more time to settle down...In general
	// the initial CloudFront distribution takes ~30 minutes
	for _, eachResource := range template.Resources {
		if eachResource.AWSCloudFormationType() == "AWS::CloudFront::Distribution" {
			stackOperationTimeout = 60 * time.Minute
			break
		}
	}
	logger.Debug().
		Dur("OperationTimeout", stackOperationTimeout).
		Msg("Computed operation timeout value")
	return stackOperationTimeout
}

// versionAwareS3KeyName returns a keyname that provides the correct cache
// invalidation semantics based on whether the target bucket
// has versioning enabled
func versionAwareS3KeyName(s3DefaultKey string,
	s3VersioningEnabled bool,
	logger *zerolog.Logger) (string, error) {
	versionKeyName := s3DefaultKey
	if !s3VersioningEnabled {
		var extension = path.Ext(s3DefaultKey)
		var prefixString = strings.TrimSuffix(s3DefaultKey, extension)

		hash := sha1.New()
		salt := fmt.Sprintf("%s-%d", s3DefaultKey, time.Now().UnixNano())
		_, writeErr := hash.Write([]byte(salt))
		if writeErr != nil {
			return "", errors.Wrapf(writeErr, "Failed to update hash digest")
		}
		versionKeyName = fmt.Sprintf("%s-%s%s",
			prefixString,
			hex.EncodeToString(hash.Sum(nil)),
			extension)

		logger.Debug().
			Str("Default", s3DefaultKey).
			Str("Extension", extension).
			Str("PrefixString", prefixString).
			Str("Unique", versionKeyName).
			Msg("Created unique S3 keyname")
	}
	return versionKeyName, nil
}

// Upload a local file to S3.  Returns the full S3 URL to the file that was
// uploaded. If the target bucket does not have versioning enabled,
// this function will automatically make a new key to ensure uniqueness
func uploadLocalFileToS3(awsSession *session.Session,
	localPath string,
	serviceName string,
	s3ObjectKey string,
	s3ObjectBucket string,
	isVersionAwareBucket bool,
	noop bool,
	logger *zerolog.Logger) (string, error) {

	// If versioning is enabled, use a stable name, otherwise use a name
	// that's dynamically created. By default assume that the bucket is
	// enabled for versioning
	if s3ObjectKey == "" {
		defaultS3KeyName := fmt.Sprintf("%s/%s", serviceName, filepath.Base(localPath))
		s3KeyName, s3KeyNameErr := versionAwareS3KeyName(defaultS3KeyName, isVersionAwareBucket, logger)
		if nil != s3KeyNameErr {
			return "", errors.Wrapf(s3KeyNameErr, "Failed to create version aware S3 keyname")
		}
		s3ObjectKey = s3KeyName
	}

	s3URL := ""
	if noop {
		// Binary size
		filesize := int64(0)
		stat, statErr := os.Stat(localPath)
		if statErr == nil {
			filesize = stat.Size()
		}
		logger.Info().
			Str("Bucket", s3ObjectBucket).
			Str("Key", s3ObjectKey).
			Str("File", filepath.Base(localPath)).
			Int64("Size", filesize).
			Msg(noopMessage("S3 upload"))

		s3URL = fmt.Sprintf("https://%s-s3.amazonaws.com/%s",
			s3ObjectBucket,
			s3ObjectKey)
	} else {
		// Then upload it
		uploadLocation, uploadURLErr := spartaS3.UploadLocalFileToS3(localPath,
			awsSession,
			s3ObjectBucket,
			s3ObjectKey,
			logger)
		if nil != uploadURLErr {
			return "", errors.Wrapf(uploadURLErr, "Failed to upload file to S3")
		}
		s3URL = uploadLocation
	}
	return s3URL, nil
}

// Private - END
////////////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////////////
// Workflow steps
////////////////////////////////////////////////////////////////////////////////

type provisionWorkflowOp struct {
	provisionContext *provisionContext
}

func (pwo *provisionWorkflowOp) MetadataString(keyName string) (string, error) {
	reader := templateMetadataReader{
		template: pwo.provisionContext.cfTemplate,
	}
	return reader.AsString(keyName)
}

func (pwo *provisionWorkflowOp) s3Bucket() string {
	s3ParamBucketName, s3ParamBucketNameExists := pwo.provisionContext.stackParameterValues[StackParamArtifactBucketName]
	if !s3ParamBucketNameExists {
		s3BucketName, s3BucketNameErr := pwo.MetadataString(MetadataParamS3Bucket)
		if s3BucketNameErr == nil {
			s3ParamBucketName = s3BucketName
		}
	}
	return s3ParamBucketName
}

func (pwo *provisionWorkflowOp) stackParameters() map[string]string {
	stackParameterValues := make(map[string]string)

	for eachKey, eachParam := range pwo.provisionContext.cfTemplate.Parameters {
		userVal, userValExists := pwo.provisionContext.stackParameterValues[eachKey]
		if !userValExists {
			noUserVal, noUserValErr := pwo.MetadataString(eachKey)
			if noUserValErr == nil && len(noUserVal) > 0 {
				userVal = fmt.Sprintf("%v", eachParam.Default)
			}
		}
		stackParameterValues[eachKey] = userVal
	}
	return stackParameterValues
}

////////////////////////////////////////////////////////////////////////////////
// precondition checks for the operation to get some metadata bout the
//
type ensureProvisionPreconditionsOp struct {
	provisionWorkflowOp
}

func (eppo *ensureProvisionPreconditionsOp) Rollback(ctx context.Context, logger *zerolog.Logger) error {
	return nil
}

func (eppo *ensureProvisionPreconditionsOp) Invoke(ctx context.Context, logger *zerolog.Logger) error {
	// So the first thing we need to do is turn all the stack parameters
	// into variables. If there is a parameter value we'll use that. If not, we
	// need to use the default template value. Based on that we can do the
	// other work...

	// Update the servicename
	serviceName, serviceNameErr := eppo.MetadataString(MetadataParamServiceName)
	logger.Debug().
		Str("serviceName", serviceName).
		Interface("serviceNameErr", serviceNameErr).
		Msg("ServiceName")
	if serviceNameErr == nil {
		eppo.provisionContext.serviceName = serviceName
	}
	if eppo.provisionContext.serviceName == "" {
		return errors.Errorf("Failed to find serviceName in MetadataString: %s", MetadataParamServiceName)
	}

	// S3 Bucket? Try stack params first, then metadata...
	s3BucketName := eppo.s3Bucket()

	// If this a NOOP, assume that versioning is not enabled
	if eppo.provisionContext.noop {
		logger.Info().
			Bool("VersioningEnabled", false).
			Str("Bucket", s3BucketName).
			Str("Region", *eppo.provisionContext.awsSession.Config.Region).
			Msg(noopMessage("S3 preconditions check"))
	} else if len(s3BucketName) != 0 {
		// Bucket region should match target
		/*
			The name of the Amazon S3 bucket where the .zip file that contains your deployment package is stored. This bucket must reside in the same AWS Region that you're creating the Lambda function in. You can specify a bucket from another AWS account as long as the Lambda function and the bucket are in the same region.
		*/
		bucketRegion, bucketRegionErr := spartaS3.BucketRegion(eppo.provisionContext.awsSession,
			s3BucketName,
			logger)

		if bucketRegionErr != nil {
			return errors.Wrap(bucketRegionErr,
				fmt.Sprintf("Checking S3 bucket <%s>", s3BucketName))
		}
		logger.Info().
			Str("Bucket", s3BucketName).
			Str("Region", bucketRegion).
			Str("CredentialsRegion", *eppo.provisionContext.awsSession.Config.Region).
			Msg("Checking S3 region")
		if bucketRegion != *eppo.provisionContext.awsSession.Config.Region {
			return fmt.Errorf("region (%s) does not match bucket region (%s)",
				*eppo.provisionContext.awsSession.Config.Region,
				bucketRegion)
		}
		// Check versioning
		// Get the S3 bucket and see if it has versioning enabled
		isEnabled, versioningPolicyErr := spartaS3.BucketVersioningEnabled(eppo.provisionContext.awsSession,
			s3BucketName,
			logger)
		// If this is an error and suggests missing region, output some helpful error text
		if nil != versioningPolicyErr {
			return versioningPolicyErr
		}
		logger.Info().
			Bool("VersioningEnabled", isEnabled).
			Str("Bucket", s3BucketName).
			Str("Region", *eppo.provisionContext.awsSession.Config.Region).
			Msg("Checking S3 versioning policy")
		eppo.provisionContext.isVersionAwareBucket = isEnabled

		// Nothing else to do...
		logger.Debug().
			Str("Region", bucketRegion).
			Msg("Confirmed S3 region match")
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// uploadPackageOp
// uplaod the ZIP packages
type uploadPackageOp struct {
	provisionWorkflowOp
}

func (upo *uploadPackageOp) Invoke(ctx context.Context, logger *zerolog.Logger) error {
	// All upload tasks will be pushed into this
	var uploadTasks []*workTask

	//////////////////////////////////////////////////////////////////////////////
	// ZIP Package Format

	// Map of keys to local paths that will be uploaded. The
	// updated S3 key will be pushed into the map
	s3UploadMap := map[string]string{
		s3UploadCloudFormationStackKey: upo.provisionContext.cfTemplatePath,
	}

	// These are optional keys and depend on whether we have a ZIP archive
	// or a Site archive. If they are nonempty in the Metadata block,
	// then we'll upload them...
	metadataKeys := []string{MetadataParamCodeArchivePath,
		MetadataParamS3SiteArchivePath}

	for _, eachKey := range metadataKeys {
		s3Path, s3PathErr := upo.MetadataString(eachKey)
		if s3PathErr == nil && len(s3Path) > 0 {
			s3UploadMap[eachKey] = s3Path
		}
	}
	s3BucketName := upo.s3Bucket()

	// For each non-empty S3 local file, upload it in here...
	uploadLocalFileTask := func(keyName string, localPath string) *workTask {
		uploadTask := func() workResult {
			// Keyname is the name of the zip file
			archiveBaseName := filepath.Base(localPath)
			// Put it in the service bucket
			uploadKeyPath := fmt.Sprintf("%s/%s", upo.provisionContext.serviceName,
				archiveBaseName)
			// Create the S3 key...
			zipS3URL, zipS3URLErr := uploadLocalFileToS3(upo.provisionContext.awsSession,
				localPath,
				upo.provisionContext.serviceName,
				uploadKeyPath,
				s3BucketName,
				upo.provisionContext.isVersionAwareBucket,
				upo.provisionContext.noop,
				logger)
			if nil != zipS3URLErr {
				return newTaskResult(nil, zipS3URLErr)
			}
			// All good, save it...
			upo.provisionContext.s3Uploads[keyName] = newS3UploadURL(zipS3URL)
			return newTaskResult(upo.provisionContext.s3Uploads[keyName], nil)
		}
		return newWorkTask(uploadTask)
	}
	// For each nonEmpty S3 upload, push it.
	for eachKey, eachLocalPath := range s3UploadMap {
		uploadTasks = append(uploadTasks, uploadLocalFileTask(eachKey, eachLocalPath))
	}

	//////////////////////////////////////////////////////////////////////////////
	// OCI Package Format

	// If this is an ECR, we need to push the image....
	ecrImageTag, ecrImageTagErr := upo.MetadataString(MetadataParamECRTag)
	logger.Debug().
		Str("ECRImageTag", ecrImageTag).
		Str("ECRImageTagErr", fmt.Sprintf("%v", ecrImageTagErr)).
		Msg("ECR Image Information")

	if ecrImageTag != "" {
		// Push the image, store the resulting URI in the image URI...
		// We have another push task here..., where the result goes into the
		// IMAGEURI param...
		pushTask := func() workResult {
			if upo.provisionContext.noop {
				logger.Info().
					Str("ECRTag", ecrImageTag).
					Msg("Bypassing ECR push due to -n/--noop flag")
				return newTaskResult("ECR Push bypassed", nil)
			}
			logger.Info().
				Str("Tag", ecrImageTag).
				Msg("Pushing local image to ECR")

			pushErr := spartaDocker.PushECRTaggedImage(ecrImageTag,
				upo.provisionContext.awsSession,
				logger)
			return newTaskResult(ecrImageTag, pushErr)
		}
		uploadTasks = append(uploadTasks, newWorkTask(pushTask))
	}

	//////////////////////////////////////////////////////////////////////////////
	// Everything is merged into a single workerpool to do the work...
	p := newWorkerPool(uploadTasks, len(uploadTasks))
	_, uploadErrors := p.Run()

	if len(uploadErrors) > 0 {
		errorText := ""
		for eachIndex, eachError := range uploadErrors {
			errorText += fmt.Sprintf("(%d) %v, ", eachIndex, eachError)
		}
		errorText = strings.TrimSuffix(errorText, ", ")
		return errors.Errorf("Encountered errors during upload: %s", errorText)
	}
	//////////////////////////////////////////////////////////////////////////////
	// Save the stack params, based on what we uploaded
	//////////////////////////////////////////////////////////////////////////////
	// TODO: This could be a bit cleaner...
	if len(s3UploadMap[MetadataParamCodeArchivePath]) != 0 {
		upo.provisionContext.stackParameterValues[StackParamS3CodeKeyName] =
			upo.provisionContext.s3Uploads[MetadataParamCodeArchivePath].path
		upo.provisionContext.stackParameterValues[StackParamS3CodeVersion] =
			upo.provisionContext.s3Uploads[MetadataParamCodeArchivePath].version
	}
	if len(s3UploadMap[MetadataParamS3SiteArchivePath]) != 0 {
		upo.provisionContext.stackParameterValues[StackParamS3SiteArchiveKey] =
			upo.provisionContext.s3Uploads[MetadataParamS3SiteArchivePath].path
		upo.provisionContext.stackParameterValues[StackParamS3SiteArchiveVersion] =
			upo.provisionContext.s3Uploads[MetadataParamS3SiteArchivePath].version
	}
	if len(ecrImageTag) != 0 {
		upo.provisionContext.stackParameterValues[StackParamCodeImageURI] = ecrImageTag
	}
	return nil
}

func (upo *uploadPackageOp) Rollback(ctx context.Context, logger *zerolog.Logger) error {
	if !upo.provisionContext.noop {
		wg := sync.WaitGroup{}
		for _, eachUploaded := range upo.provisionContext.s3Uploads {
			rollbackFunc := spartaS3.CreateS3RollbackFunc(upo.provisionContext.awsSession, eachUploaded.location)
			wg.Add(1)
			go func(rollFunc spartaS3.RollbackFunction, logger *zerolog.Logger) {
				defer wg.Done()
				errResult := rollFunc(logger)
				if errResult != nil {
					logger.Warn().
						Err(errResult).
						Msg("S3 upload rollback function failed")
				}
			}(rollbackFunc, logger)
		}
		// IF we have an ECRTag, log that we won't be deleting it
		ecrTag := upo.provisionContext.stackParameterValues[StackParamCodeImageURI]
		if ecrTag != "" {
			logger.Info().
				Str("Tag", ecrTag).
				Msg("Image will not be deleted from repository")
		}
		wg.Wait()
	} else {
		logger.Info().
			Msg("Nothing to rollback from S3 in NOOP mode")
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// codePipelineTriggerOp
// create the pipeline trigger op
type codePipelineTriggerOp struct {
	provisionWorkflowOp
}

func (cpto *codePipelineTriggerOp) Rollback(ctx context.Context, logger *zerolog.Logger) error {
	return nil
}

func (cpto *codePipelineTriggerOp) Invoke(ctx context.Context, logger *zerolog.Logger) error {
	tmpFile, err := system.TemporaryFile(ScratchDirectory, cpto.provisionContext.codePipelineTrigger)
	if err != nil {
		return errors.Wrapf(err, "Failed to create temporary file for CodePipeline")
	}

	logger.Info().
		Str("PipelineName", tmpFile.Name()).
		Str("Path", tmpFile.Name()).
		Msg("Creating pipeline archive")

	// File info for the binary executable
	templateArchive := zip.NewWriter(tmpFile)
	zipEntryName := "cloudformation.json"
	bytesWriter, bytesWriterErr := templateArchive.Create(zipEntryName)
	if bytesWriterErr != nil {
		return errors.Wrapf(bytesWriterErr, "Failed to create Zip writer")
	}
	// We need to get the template bytes into a reader...
	jsonTemplateBytes, jsonTemplateBytesErr := cpto.provisionContext.cfTemplate.JSON()
	if jsonTemplateBytesErr != nil {
		return errors.Wrapf(jsonTemplateBytesErr, "Failed to Marshal CloudFormation template")
	}
	bytesReader := bytes.NewReader(jsonTemplateBytes)
	written, writtenErr := io.Copy(bytesWriter, bytesReader)
	if nil != writtenErr {
		return errors.Wrapf(writtenErr, "Failed to copy CloudFormation template to Zip output")
	}
	logger.Debug().
		Int64("WrittenBytes", written).
		Str("ZipName", zipEntryName).
		Msg("Archiving file")

	// If there is a codePipelineEnvironments defined, then we'll need to get all the
	// maps, marshal them to JSON, then add the JSON to the ZIP archive.
	if nil != codePipelineEnvironments {
		for eachEnvironment, eachMap := range codePipelineEnvironments {
			codePipelineParameters := map[string]interface{}{
				"Parameters": eachMap,
			}
			environmentJSON, environmentJSONErr := json.Marshal(codePipelineParameters)
			if nil != environmentJSONErr {
				logger.Error().
					Str("Environment", eachEnvironment).
					Msg("Failed to marshal environment")
				return environmentJSONErr
			}
			var envVarName = fmt.Sprintf("%s.json", eachEnvironment)

			// File info for the binary executable
			binaryWriter, binaryWriterErr := templateArchive.Create(envVarName)
			if binaryWriterErr != nil {
				return binaryWriterErr
			}
			_, writeErr := binaryWriter.Write(environmentJSON)
			if writeErr != nil {
				return writeErr
			}
		}
	}
	archiveCloseErr := templateArchive.Close()
	if nil != archiveCloseErr {
		return archiveCloseErr
	}
	tempfileCloseErr := tmpFile.Close()
	if nil != tempfileCloseErr {
		return tempfileCloseErr
	}
	// Leave it here...
	logger.Info().
		Str("File", filepath.Base(tmpFile.Name())).
		Msg("Created CodePipeline archive")
	// The key is the name + the pipeline name
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// inPlaceUpdatesOp
// perform the inplace update request

type inPlaceUpdatesOp struct {
	provisionWorkflowOp
}

func (ipuo *inPlaceUpdatesOp) Rollback(ctx context.Context, logger *zerolog.Logger) error {
	return nil
}

func (ipuo *inPlaceUpdatesOp) Invoke(ctx context.Context, logger *zerolog.Logger) error {
	if ipuo.provisionContext.noop {
		logger.Info().
			Msg(noopMessage("InPlace Update Check"))
		return nil
	}

	// Let's see if there is a change at all
	awsCloudFormation := cloudformation.New(ipuo.provisionContext.awsSession)
	changeSetRequestName := CloudFormationResourceName(fmt.Sprintf("%sInPlaceChangeSet",
		ipuo.provisionContext.serviceName))
	changes, changesErr := spartaCF.CreateStackChangeSet(changeSetRequestName,
		ipuo.provisionContext.serviceName,
		ipuo.provisionContext.cfTemplate,
		ipuo.provisionContext.s3Uploads[s3UploadCloudFormationStackKey].location,
		ipuo.stackParameters(),
		ipuo.provisionContext.stackTags,
		awsCloudFormation,
		logger)
	if nil != changesErr {
		return changesErr
	}
	if nil == changes || len(changes.Changes) <= 0 {
		return fmt.Errorf("no changes detected")
	}
	s3BucketName := ipuo.s3Bucket()

	//////////////////////////////////////////////////////////////////////////////

	// Either Docker URI or the code URI should be there...
	ecrImageURI := ipuo.provisionContext.stackParameterValues[StackParamCodeImageURI]
	codeKeyName := ipuo.provisionContext.stackParameterValues[StackParamS3CodeKeyName]
	codeKeyVersion := ipuo.provisionContext.stackParameterValues[StackParamS3CodeVersion]

	if ecrImageURI == "" && codeKeyName == "" {
		return errors.Errorf("Failed to find either Code ZIP key or ECR Image tag for inPlace update")
	}

	updateCodeRequests := []*lambda.UpdateFunctionCodeInput{}
	invalidInPlaceRequests := []*cloudformation.Change{}
	for _, eachChange := range changes.Changes {
		resourceChange := eachChange.ResourceChange
		if *resourceChange.Action == "Modify" &&
			*resourceChange.ResourceType == "AWS::Lambda::Function" {
			updateCodeRequest := &lambda.UpdateFunctionCodeInput{
				FunctionName: resourceChange.PhysicalResourceId,
			}
			// Either ZIP or OCI - pick one
			if codeKeyName != "" {
				updateCodeRequest.S3Bucket = aws.String(s3BucketName)
				updateCodeRequest.S3Key = aws.String(codeKeyName)
				updateCodeRequest.S3ObjectVersion = aws.String(codeKeyVersion)
			} else if ecrImageURI != "" {
				updateCodeRequest.ImageUri = aws.String(ecrImageURI)
			}
			updateCodeRequests = append(updateCodeRequests, updateCodeRequest)
		} else {
			invalidInPlaceRequests = append(invalidInPlaceRequests, eachChange)
		}
	}
	if len(invalidInPlaceRequests) != 0 {
		for _, eachInvalidChange := range invalidInPlaceRequests {
			logger.Warn().
				Str("ID", *eachInvalidChange.ResourceChange.LogicalResourceId).
				Str("Action", *eachInvalidChange.ResourceChange.Action).
				Interface("Details", eachInvalidChange.ResourceChange.Details).
				Str("ResourceType", *eachInvalidChange.ResourceChange.ResourceType).
				Msg("Additional change detected for in-place update")
		}
		//return fmt.Errorf("unsupported in-place operations detected")
	}

	logger.Info().
		Int("FunctionCount", len(updateCodeRequests)).
		Msg("Updating Lambda function code")
	logger.Debug().
		Interface("Updates", updateCodeRequests).
		Msg("Update requests")

	updateTaskMaker := func(lambdaSvc *lambda.Lambda, request *lambda.UpdateFunctionCodeInput) taskFunc {
		return func() workResult {
			_, updateResultErr := lambdaSvc.UpdateFunctionCode(request)
			return newTaskResult("", updateResultErr)
		}
	}
	inPlaceUpdateTasks := make([]*workTask,
		len(updateCodeRequests))
	awsLambda := lambda.New(ipuo.provisionContext.awsSession)
	for eachIndex, eachUpdateCodeRequest := range updateCodeRequests {
		updateTask := updateTaskMaker(awsLambda, eachUpdateCodeRequest)
		inPlaceUpdateTasks[eachIndex] = newWorkTask(updateTask)
	}

	// Add the request to delete the change set...
	// TODO: add some retry logic in here to handle failures.
	deleteChangeSetTask := func() workResult {
		_, deleteChangeSetResultErr := spartaCF.DeleteChangeSet(ipuo.provisionContext.serviceName,
			changeSetRequestName,
			awsCloudFormation)
		return newTaskResult("", deleteChangeSetResultErr)
	}
	inPlaceUpdateTasks = append(inPlaceUpdateTasks, newWorkTask(deleteChangeSetTask))
	p := newWorkerPool(inPlaceUpdateTasks, len(inPlaceUpdateTasks))
	_, asyncErrors := p.Run()
	if len(asyncErrors) != 0 {
		return fmt.Errorf("failed to update function code: %v", asyncErrors)
	}
	// Describe the stack so that we can satisfy the contract with the
	// normal path using CloudFormation
	describeStacksInput := &cloudformation.DescribeStacksInput{
		StackName: aws.String(ipuo.provisionContext.serviceName),
	}
	describeStackOutput, describeStackOutputErr := awsCloudFormation.DescribeStacks(describeStacksInput)
	if nil != describeStackOutputErr {
		return describeStackOutputErr
	}
	ipuo.provisionContext.stack = describeStackOutput.Stacks[0]
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// cloudformationStackUpdateOp
// Update the functions via a cloudformation stack opreation
type cloudformationStackUpdateOp struct {
	provisionWorkflowOp
}

func (cfsu *cloudformationStackUpdateOp) Rollback(ctx context.Context, logger *zerolog.Logger) error {
	return nil
}
func (cfsu *cloudformationStackUpdateOp) Invoke(ctx context.Context, logger *zerolog.Logger) error {
	if cfsu.provisionContext.noop {
		logger.Info().
			Msg(noopMessage("CloudFormation Stack update"))

		return nil
	}

	operationTimeout := maximumStackOperationTimeout(cfsu.provisionContext.cfTemplate, logger)
	startTime := time.Now()

	// Regular update, go ahead with the CloudFormation changes
	stack, stackErr := spartaCF.ConvergeStackState(cfsu.provisionContext.serviceName,
		cfsu.provisionContext.cfTemplate,
		cfsu.provisionContext.s3Uploads[s3UploadCloudFormationStackKey].location,
		cfsu.provisionContext.stackParameterValues,
		cfsu.provisionContext.stackTags,
		startTime,
		operationTimeout,
		cfsu.provisionContext.awsSession,
		"▬",
		dividerLength,
		logger)

	if stackErr != nil {
		return stackErr
	}
	cfsu.provisionContext.stack = stack
	return nil
}

////////////////////////////////////////////////////////////////////////////////
//
type outputStackInfoOp struct {
	provisionWorkflowOp
}

func (osi *outputStackInfoOp) Rollback(ctx context.Context, logger *zerolog.Logger) error {
	return nil
}
func (osi *outputStackInfoOp) Invoke(ctx context.Context, logger *zerolog.Logger) error {
	if osi.provisionContext.stack != nil {
		logger.Info().
			Str("StackName", *osi.provisionContext.stack.StackName).
			Str("StackId", *osi.provisionContext.stack.StackId).
			Time("CreationTime", *osi.provisionContext.stack.CreationTime).
			Msg("Stack provisioned")
	}
	return nil
}

/*
type validatePostConditionOp struct {
	provisionWorkflowOp
}

func (vpco *validatePostConditionOp) Rollback(ctx context.Context, logger *zerolog.Logger) error {
	return nil
}
func (vpco *validatePostConditionOp) Invoke(ctx context.Context, logger *zerolog.Logger) error {
	return nil
}
*/
func verifyLambdaPreconditions(lambdaAWSInfo *LambdaAWSInfo, logger *zerolog.Logger) error {

	return nil
}

// Provision compiles, packages, and provisions (either via create or update) a Sparta application.
// The serviceName is the service's logical
// identify and is used to determine create vs update operations.  The compilation options/flags are:
//
// 	TAGS:         -tags lambdabinary
// 	ENVIRONMENT:  GOOS=linux GOARCH=amd64
//
// The compiled binary is packaged with a NodeJS proxy shim to manage AWS Lambda setup & invocation per
// http://docs.aws.amazon.com/lambda/latest/dg/authoring-function-in-nodejs.html
//
// The two files are ZIP'd, posted to S3 and used as an input to a dynamically generated CloudFormation
// template (http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/Welcome.html)
// which creates or updates the service state.
//
func Provision(noop bool,
	templatePath string,
	stackParamValues map[string]string,
	stackTags map[string]string,
	inPlaceUpdates bool,
	codePipelineTrigger string,
	logger *zerolog.Logger) error {

	logger.Info().
		Bool("NOOP", noop).
		Bool("InPlaceUpdates", inPlaceUpdates).
		Str("Template", templatePath).
		Interface("Params", stackParamValues).
		Interface("Tags", stackTags).
		Msg("Provisioning service")

	pc := &provisionContext{
		awsSession:           spartaAWS.NewSession(logger),
		cfTemplatePath:       templatePath,
		cfTemplate:           gof.NewTemplate(),
		codePipelineTrigger:  codePipelineTrigger,
		stackParameterValues: stackParamValues,
		stackTags:            stackTags,
		s3Uploads:            map[string]*s3UploadURL{},
		inPlaceUpdates:       inPlaceUpdates,
		noop:                 noop,
	}

	// Unmarshal the JSON template into the struct
	/* #nosec G304 */
	targetTemplate, targetTemplateErr := gofroot.Open(templatePath)
	if targetTemplateErr != nil {
		return targetTemplateErr
	}
	pc.cfTemplate = targetTemplate

	//////////////////////////////////////////////////////////////////////////////
	// Workflow
	//////////////////////////////////////////////////////////////////////////////
	provisionPipeline := pipeline{}

	// Preconditions

	/* #nosec G104 */
	stagePreconditions := &pipelineStage{}
	stagePreconditions.Append("validatePreconditions",
		&ensureProvisionPreconditionsOp{
			provisionWorkflowOp: provisionWorkflowOp{
				provisionContext: pc,
			}})
	provisionPipeline.Append("preconditions", stagePreconditions)

	// Build Package
	stageBuild := &pipelineStage{}
	if pc.codePipelineTrigger == "" {
		stageBuild.Append("uploadPackages",
			&uploadPackageOp{
				provisionWorkflowOp: provisionWorkflowOp{
					provisionContext: pc,
				}})
		provisionPipeline.Append("upload", stageBuild)
	} else {
		stageBuild.Append("codePipelinePackage",
			&codePipelineTriggerOp{
				provisionWorkflowOp: provisionWorkflowOp{
					provisionContext: pc,
				}})
		provisionPipeline.Append("build", stageBuild)
	}

	// Which mutation to apply?
	stageApply := &pipelineStage{}
	if inPlaceUpdates {
		stageApply.Append("inPlaceUpdates",
			&inPlaceUpdatesOp{
				provisionWorkflowOp: provisionWorkflowOp{
					provisionContext: pc,
				}})
	} else {
		stageApply.Append("cloudformationUpdate",
			&cloudformationStackUpdateOp{
				provisionWorkflowOp: provisionWorkflowOp{
					provisionContext: pc,
				}})
	}
	provisionPipeline.Append("apply", stageApply)

	// Describe tbe output...
	stageDescribe := &pipelineStage{}
	stageDescribe.Append("describeStack",
		&outputStackInfoOp{
			provisionWorkflowOp: provisionWorkflowOp{
				provisionContext: pc,
			}})
	provisionPipeline.Append("describe", stageDescribe)

	// Run
	pipelineContext := context.Background()
	provisionErr := provisionPipeline.Run(pipelineContext, "Provision", logger)
	if provisionErr != nil {
		return provisionErr
	}
	return nil
}
