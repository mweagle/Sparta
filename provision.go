// +build !lambdabinary

package sparta

import (
	"archive/zip"
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	spartaCF "github.com/mweagle/Sparta/aws/cloudformation"
	spartaS3 "github.com/mweagle/Sparta/aws/s3"
	spartaZip "github.com/mweagle/Sparta/zip"
	"net/url"
	"path/filepath"

	"github.com/mweagle/cloudformationresources"
	"io"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	gocf "github.com/crewjam/go-cloudformation"
	spartaAWS "github.com/mweagle/Sparta/aws"
)

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS
////////////////////////////////////////////////////////////////////////////////
func spartaTagName(baseKey string) string {
	return fmt.Sprintf("io:gosparta:%s", baseKey)
}

// SpartaTagHomeKey is the keyname used in the CloudFormation Output
// that stores the Sparta home URL.
// @enum OutputKey
var SpartaTagHomeKey = spartaTagName("home")

// SpartaTagVersionKey is the keyname used in the CloudFormation Output
// that stores the Sparta version used to provision/update the service.
// @enum OutputKey
var SpartaTagVersionKey = spartaTagName("version")

// SpartaTagBuildIDKey is the keyname used in the CloudFormation Output
// that stores the user-supplied or automatically generated BuildID
// for this run
var SpartaTagBuildIDKey = spartaTagName("buildId")

// SpartaTagBuildTagsKey is the keyname used in the CloudFormation Output
// that stores the optional user-supplied golang build tags
var SpartaTagBuildTagsKey = spartaTagName("buildTags")

// The basename of the scripts that are embedded into CONSTANTS.go
// by `esc` during the generate phase.  In order to export these, there
// MUST be a corresponding PROXIED_MODULES entry for the base filename
// in resources/index.js
var customResourceScripts = []string{"sparta_utils.js",
	"golang-constants.json"}

var golangCustomResourceTypes = []string{
	cloudformationresources.SESLambdaEventSource,
	cloudformationresources.S3LambdaEventSource,
	cloudformationresources.SNSLambdaEventSource,
	cloudformationresources.CloudWatchLogsLambdaEventSource,
	cloudformationresources.ZipToS3Bucket,
}

type finalizerFunction func(logger *logrus.Logger)

// The relative path of the custom scripts that is used
// to create the filename relative path when creating the custom archive
const provisioningResourcesRelPath = "/resources/provision"

////////////////////////////////////////////////////////////////////////////////
// Type that encapsulates an S3 URL with accessors to return either the
// full URL or just the valid S3 Keyname
type s3UploadURL struct {
	location string
}

func (s3URL *s3UploadURL) url() string {
	return s3URL.location
}
func (s3URL *s3UploadURL) keyName() string {
	// Find the hostname in the URL, then strip it out
	urlParts, _ := url.Parse(s3URL.location)
	return strings.TrimPrefix(urlParts.Path, "/")
}

func newS3UploadURL(s3URL string) *s3UploadURL {
	return &s3UploadURL{location: s3URL}
}

////////////////////////////////////////////////////////////////////////////////

// Represents data associated with provisioning the S3 Site iff defined
type s3SiteContext struct {
	s3Site      *S3Site
	s3UploadURL *s3UploadURL
}

// Type of a workflow step.  Each step is responsible
// for returning the next step or an error if the overall
// workflow should stop.
type workflowStep func(ctx *workflowContext) (workflowStep, error)

////////////////////////////////////////////////////////////////////////////////
// Workflow context
// The workflow context is created by `provision` and provided to all
// functions that constitute the provisioning workflow.
type workflowContext struct {
	// Is this is a -dry-run?
	noop bool
	// Canonical basename of the service.  Also used as the CloudFormation
	// stack name
	serviceName string
	// Service description
	serviceDescription string
	// The slice of Lambda functions that constitute the service
	lambdaAWSInfos []*LambdaAWSInfo
	// Optional APIGateway definition to associate with this service
	api *API
	// Optional S3 site data to provision together with this service
	s3SiteContext *s3SiteContext
	// CloudFormation Template
	cfTemplate *gocf.Template
	// Cached IAM role name map.  Used to support dynamic and static IAM role
	// names.  Static ARN role names are checked for existence via AWS APIs
	// prior to CloudFormation provisioning.
	lambdaIAMRoleNameMap map[string]*gocf.StringExpr
	// The user-supplied S3 bucket where service artifacts should be posted.
	s3Bucket string
	// Is versioning enabled for s3 Bucket?
	s3BucketVersioningEnabled bool
	// The user-supplied or automatically generated BuildID
	buildID string
	// Code pipeline S3 trigger keyname
	codePipelineTrigger string
	// Optional user-supplied build tags
	buildTags string
	// Optional link flags
	linkFlags string
	// The time when we started s.t. we can filter stack events
	buildTime time.Time
	// Information about the ZIP archive that contains the LambdaCode source
	s3CodeZipURL *s3UploadURL
	// AWS Session to be used for all API calls made in the process of provisioning
	// this service.
	awsSession *session.Session
	// IO writer for autogenerated template results
	templateWriter io.Writer
	// User supplied workflow hooks
	workflowHooks *WorkflowHooks
	// Context to pass between workflow operations
	workflowHooksContext map[string]interface{}
	// Preconfigured logger
	logger *logrus.Logger
	// Optional rollback functions that workflow steps may append to if they
	// have made mutations during provisioning.
	rollbackFunctions []spartaS3.RollbackFunction

	// Optional finalizer functions that are unconditionally executed following
	// workflow completion, success or failure
	finalizerFunctions []finalizerFunction
}

// Register a rollback function in the event that the provisioning
// function failed.
func (ctx *workflowContext) registerRollback(userFunction spartaS3.RollbackFunction) {
	if nil == ctx.rollbackFunctions || len(ctx.rollbackFunctions) <= 0 {
		ctx.rollbackFunctions = make([]spartaS3.RollbackFunction, 0)
	}
	ctx.rollbackFunctions = append(ctx.rollbackFunctions, userFunction)
}

// Register a rollback function in the event that the provisioning
// function failed.
func (ctx *workflowContext) registerFinalizer(userFunction finalizerFunction) {
	if nil == ctx.finalizerFunctions || len(ctx.finalizerFunctions) <= 0 {
		ctx.finalizerFunctions = make([]finalizerFunction, 0)
	}
	ctx.finalizerFunctions = append(ctx.finalizerFunctions, userFunction)
}

// Register a finalizer that cleans up local artifacts
func (ctx *workflowContext) registerFileCleanupFinalizer(localPath string) {
	cleanup := func(logger *logrus.Logger) {
		errRemove := os.Remove(localPath)
		if nil != errRemove {
			logger.WithFields(logrus.Fields{
				"Path":  localPath,
				"Error": errRemove,
			}).Warn("Failed to cleanup intermediate artifact")
		} else {
			logger.WithFields(logrus.Fields{
				"Path": localPath,
			}).Debug("Build artifact deleted")
		}
	}
	ctx.registerFinalizer(cleanup)
}

// Run any provided rollback functions
func (ctx *workflowContext) rollback() {
	// Run each cleanup function concurrently.  If there's an error
	// all we're going to do is log it as a warning, since at this
	// point there's nothing to do...
	var wg sync.WaitGroup
	wg.Add(len(ctx.rollbackFunctions))

	// Include the user defined rollback if there is one...
	if ctx.workflowHooks != nil && ctx.workflowHooks.Rollback != nil {
		wg.Add(1)
		go func(hook RollbackHook, context map[string]interface{},
			serviceName string,
			awsSession *session.Session,
			noop bool,
			logger *logrus.Logger) {
			// Decrement the counter when the goroutine completes.
			defer wg.Done()
			hook(context, serviceName, awsSession, noop, logger)
		}(ctx.workflowHooks.Rollback,
			ctx.workflowHooksContext,
			ctx.serviceName,
			ctx.awsSession,
			ctx.noop,
			ctx.logger)
	}

	ctx.logger.WithFields(logrus.Fields{
		"RollbackCount": len(ctx.rollbackFunctions),
	}).Info("Invoking rollback functions")

	for _, eachCleanup := range ctx.rollbackFunctions {
		go func(cleanupFunc spartaS3.RollbackFunction, goLogger *logrus.Logger) {
			// Decrement the counter when the goroutine completes.
			defer wg.Done()
			// Fetch the URL.
			err := cleanupFunc(goLogger)
			if nil != err {
				ctx.logger.WithFields(logrus.Fields{
					"Error": err,
				}).Warning("Failed to cleanup resource")
			}
		}(eachCleanup, ctx.logger)
	}
	wg.Wait()
}

////////////////////////////////////////////////////////////////////////////////
// Private - START
//

// Encapsulate calling a workflow hook
func callWorkflowHook(hook WorkflowHook, ctx *workflowContext) error {
	if nil == hook {
		return nil
	}
	// Run the hook
	hookName := runtime.FuncForPC(reflect.ValueOf(hook).Pointer()).Name()
	ctx.logger.WithFields(logrus.Fields{
		"WorkflowHook":        hookName,
		"WorkflowHookContext": ctx.workflowHooksContext,
	}).Info("Calling WorkflowHook")

	return hook(ctx.workflowHooksContext,
		ctx.serviceName,
		ctx.s3Bucket,
		ctx.buildID,
		ctx.awsSession,
		ctx.noop,
		ctx.logger)
}

func versionAwareS3KeyName(s3DefaultKey string, s3VersioningEnabled bool, logger *logrus.Logger) (string, error) {
	versionKeyName := s3DefaultKey
	if !s3VersioningEnabled {
		var extension = path.Ext(s3DefaultKey)
		var prefixString = strings.TrimSuffix(s3DefaultKey, extension)

		hash := sha1.New()
		salt := fmt.Sprintf("%s-%d", s3DefaultKey, time.Now().UnixNano())
		hash.Write([]byte(salt))
		versionKeyName = fmt.Sprintf("%s-%s%s",
			prefixString,
			hex.EncodeToString(hash.Sum(nil)),
			extension)

		logger.WithFields(logrus.Fields{
			"Default":      s3DefaultKey,
			"Extension":    extension,
			"PrefixString": prefixString,
			"Unique":       versionKeyName,
		}).Debug("Created unique S3 keyname")
	}
	return versionKeyName, nil
}

// Upload a local file to S3.  Returns the full S3 URL to the file that was
// uploaded. If the target bucket does not have versioning enabled,
// this function will automatically make a new key to ensure uniqueness
func uploadLocalFileToS3(localPath string, s3ObjectKey string, ctx *workflowContext) (string, error) {

	// If versioning is enabled, use a stable name, otherwise use a name
	// that's dynamically created. By default assume that the bucket is
	// enabled for versioning
	if "" == s3ObjectKey {
		defaultS3KeyName := path.Join(ctx.serviceName, filepath.Base(localPath))
		s3KeyName, s3KeyNameErr := versionAwareS3KeyName(defaultS3KeyName,
			ctx.s3BucketVersioningEnabled,
			ctx.logger)
		if nil != s3KeyNameErr {
			return "", s3KeyNameErr
		}
		s3ObjectKey = s3KeyName
	}

	s3URL := ""
	if ctx.noop {
		ctx.logger.WithFields(logrus.Fields{
			"Bucket": ctx.s3Bucket,
			"Key":    s3ObjectKey,
			"File":   filepath.Base(localPath),
		}).Info("Bypassing S3 upload due to --noop")
		s3URL = fmt.Sprintf("https://%s-s3.amazonaws.com/%s", ctx.s3Bucket, s3ObjectKey)
	} else {
		// Make sure we mark things for cleanup in case there's a problem
		ctx.registerFileCleanupFinalizer(localPath)
		// Then upload it
		uploadLocation, uploadURLErr := spartaS3.UploadLocalFileToS3(localPath,
			ctx.awsSession,
			ctx.s3Bucket,
			s3ObjectKey,
			ctx.logger)
		if nil != uploadURLErr {
			return "", uploadURLErr
		}
		s3URL = uploadLocation
		ctx.registerRollback(spartaS3.CreateS3RollbackFunc(ctx.awsSession, uploadLocation))
	}
	return s3URL, nil
}

// Private - END
////////////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////////////
// Workflow steps
////////////////////////////////////////////////////////////////////////////////

// Verify & cache the IAM rolename to ARN mapping
func verifyIAMRoles(ctx *workflowContext) (workflowStep, error) {
	// The map is either a literal Arn from a pre-existing role name
	// or a gocf.RefFunc() value.
	// Don't verify them, just create them...
	ctx.logger.Info("Verifying IAM Lambda execution roles")
	ctx.lambdaIAMRoleNameMap = make(map[string]*gocf.StringExpr, 0)
	svc := iam.New(ctx.awsSession)

	// Assemble all the RoleNames and validate the inline IAMRoleDefinitions
	var allRoleNames []string
	for _, eachLambdaInfo := range ctx.lambdaAWSInfos {
		if "" != eachLambdaInfo.RoleName {
			allRoleNames = append(allRoleNames, eachLambdaInfo.RoleName)
		}
		// Custom resources?
		for _, eachCustomResource := range eachLambdaInfo.customResources {
			if "" != eachCustomResource.roleName {
				allRoleNames = append(allRoleNames, eachCustomResource.roleName)
			}
		}

		// Validate the IAMRoleDefinitions associated
		if nil != eachLambdaInfo.RoleDefinition {
			logicalName := eachLambdaInfo.RoleDefinition.logicalName(ctx.serviceName, eachLambdaInfo.lambdaFunctionName())
			_, exists := ctx.lambdaIAMRoleNameMap[logicalName]
			if !exists {
				// Insert it into the resource creation map and add
				// the "Ref" entry to the hashmap
				ctx.cfTemplate.AddResource(logicalName,
					eachLambdaInfo.RoleDefinition.toResource(eachLambdaInfo.EventSourceMappings, eachLambdaInfo.Options, ctx.logger))

				ctx.lambdaIAMRoleNameMap[logicalName] = gocf.GetAtt(logicalName, "Arn")
			}
		}

		// And the custom resource IAMRoles as well...
		for _, eachCustomResource := range eachLambdaInfo.customResources {
			if nil != eachCustomResource.roleDefinition {
				customResourceLogicalName := eachCustomResource.roleDefinition.logicalName(ctx.serviceName,
					eachCustomResource.userFunctionName)

				_, exists := ctx.lambdaIAMRoleNameMap[customResourceLogicalName]
				if !exists {
					ctx.cfTemplate.AddResource(customResourceLogicalName,
						eachCustomResource.roleDefinition.toResource(nil, eachCustomResource.options, ctx.logger))
					ctx.lambdaIAMRoleNameMap[customResourceLogicalName] = gocf.GetAtt(customResourceLogicalName, "Arn")
				}
			}
		}
	}

	// Then check all the RoleName literals
	for _, eachRoleName := range allRoleNames {
		_, exists := ctx.lambdaIAMRoleNameMap[eachRoleName]
		if !exists {
			// Check the role
			params := &iam.GetRoleInput{
				RoleName: aws.String(eachRoleName),
			}
			ctx.logger.Debug("Checking IAM RoleName: ", eachRoleName)
			resp, err := svc.GetRole(params)
			if err != nil {
				ctx.logger.Error(err.Error())
				return nil, err
			}
			// Cache it - we'll need it later when we create the
			// CloudFormation template which needs the execution Arn (not role)
			ctx.lambdaIAMRoleNameMap[eachRoleName] = gocf.String(*resp.Role.Arn)
		}
	}

	ctx.logger.WithFields(logrus.Fields{
		"Count": len(ctx.lambdaIAMRoleNameMap),
	}).Info("IAM roles verified")

	return verifyAWSPreconditions, nil
}

// Verify that everything is setup in AWS before we start building things
func verifyAWSPreconditions(ctx *workflowContext) (workflowStep, error) {
	// Get the S3 bucket and see if it has versioning enabled
	isEnabled, versioningPolicyErr := spartaS3.BucketVersioningEnabled(ctx.awsSession, ctx.s3Bucket, ctx.logger)
	if nil != versioningPolicyErr {
		return nil, versioningPolicyErr
	}
	ctx.logger.WithFields(logrus.Fields{
		"VersioningEnabled": isEnabled,
		"Bucket":            ctx.s3Bucket,
	}).Info("Checking S3 versioning")
	ctx.s3BucketVersioningEnabled = isEnabled
	if "" != ctx.codePipelineTrigger && !isEnabled {
		return nil, fmt.Errorf("Bucket (%s) for CodePipeline trigger doesn't have a versioning policy enabled", ctx.s3Bucket)
	}

	// If there are codePipeline environments defined, warn if they don't include
	// the same keysets
	if nil != codePipelineEnvironments {
		mapKeys := func(inboundMap map[string]string) []string {
			keys := make([]string, len(inboundMap))
			i := 0
			for k := range inboundMap {
				keys[i] = k
				i++
			}
			return keys
		}
		aggregatedKeys := make([][]string, len(codePipelineEnvironments))
		i := 0
		for _, eachEnvMap := range codePipelineEnvironments {
			aggregatedKeys[i] = mapKeys(eachEnvMap)
			i++
		}
		i = 0
		keysEqual := true
		for _, eachKeySet := range aggregatedKeys {
			j := 0
			for _, eachKeySetTest := range aggregatedKeys {
				if j != i {
					if !reflect.DeepEqual(eachKeySet, eachKeySetTest) {
						keysEqual = false
					}
				}
				j++
			}
			i++
		}
		if !keysEqual {
			// Setup an interface with the fields so that the log message
			fields := make(logrus.Fields, len(codePipelineEnvironments))
			for eachEnv, eachEnvMap := range codePipelineEnvironments {
				fields[eachEnv] = eachEnvMap
			}
			ctx.logger.WithFields(fields).Warn("CodePipeline environments do not define equivalent environment keys")
		}
	}

	return createPackageStep(), nil
}

// Return a string representation of a JS function call that can be exposed
// to AWS Lambda
func createNewNodeJSProxyEntry(lambdaInfo *LambdaAWSInfo, logger *logrus.Logger) string {
	logger.WithFields(logrus.Fields{
		"FunctionName": lambdaInfo.lambdaFunctionName(),
	}).Info("Registering Sparta function")

	// We do know the CF resource name here - could write this into
	// index.js and expose a GET localhost:9000/lambdaMetadata
	// which wraps up DescribeStackResource for the running
	// lambda function
	primaryEntry := fmt.Sprintf("exports[\"%s\"] = createForwarder(\"/%s\");\n",
		lambdaInfo.jsHandlerName(),
		lambdaInfo.lambdaFunctionName())
	return primaryEntry
}

func createUserCustomResourceEntry(customResource *customResourceInfo, logger *logrus.Logger) string {
	// The resource name is a :: delimited one, so let's sanitize that
	// to make it a valid JS identifier
	logger.WithFields(logrus.Fields{
		"UserFunction":       customResource.userFunctionName,
		"NodeJSFunctionName": customResource.jsHandlerName(),
	}).Debug("Registering User CustomResource function")

	primaryEntry := fmt.Sprintf("exports[\"%s\"] = createForwarder(\"/%s\");\n",
		customResource.jsHandlerName(),
		customResource.userFunctionName)
	return primaryEntry
}

func createNewSpartaCustomResourceEntry(resourceName string, logger *logrus.Logger) string {
	// The resource name is a :: delimited one, so let's sanitize that
	// to make it a valid JS identifier
	jsName := javascriptExportNameForCustomResourceType(resourceName)
	logger.WithFields(logrus.Fields{
		"Resource":           resourceName,
		"NodeJSFunctionName": jsName,
	}).Debug("Registering Sparta CustomResource function")

	primaryEntry := fmt.Sprintf("exports[\"%s\"] = createForwarder(\"/%s\");\n",
		jsName,
		resourceName)
	return primaryEntry
}

func logFilesize(message string, filePath string, logger *logrus.Logger) {
	// Binary size
	stat, err := os.Stat(filePath)
	if err == nil {
		logger.WithFields(logrus.Fields{
			"KB": stat.Size() / 1024,
			"MB": stat.Size() / (1024 * 1024),
		}).Info(message)
	}
}

func buildGoBinary(executableOutput string, buildTags string, linkFlags string, logger *logrus.Logger) error {
	// Go generate
	cmd := exec.Command("go", "generate")
	if logger.Level == logrus.DebugLevel {
		cmd = exec.Command("go", "generate", "-v", "-x")
	}
	cmd.Env = os.Environ()
	commandString := fmt.Sprintf("%s", cmd.Args)
	logger.Info(fmt.Sprintf("Running `%s`", strings.Trim(commandString, "[]")))
	goGenerateErr := runOSCommand(cmd, logger)
	if nil != goGenerateErr {
		return goGenerateErr
	}

	// TODO: Smaller binaries via linker flags
	// Ref: https://blog.filippo.io/shrink-your-go-binaries-with-this-one-weird-trick/
	allBuildTags := fmt.Sprintf("lambdabinary %s", buildTags)

	buildArgs := []string{
		"build",
		"-o",
		executableOutput,
		"-tags",
		allBuildTags,
	}
	// Append all the linker flags
	if len(linkFlags) != 0 {
		buildArgs = append(buildArgs, "-ldflags", linkFlags)
	}
	buildArgs = append(buildArgs, ".")
	cmd = exec.Command("go", buildArgs...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "GOOS=linux", "GOARCH=amd64")
	logger.WithFields(logrus.Fields{
		"Name": executableOutput,
	}).Info("Compiling binary")
	return runOSCommand(cmd, logger)
}

func writeNodeJSShim(serviceName string,
	executableOutput string,
	lambdaAWSInfos []*LambdaAWSInfo,
	zipWriter *zip.Writer,
	logger *logrus.Logger) error {

	// Add the string literal adapter, which requires us to add exported
	// functions to the end of index.js.  These NodeJS exports will be
	// linked to the AWS Lambda NodeJS function name, and are basically
	// automatically generated pass through proxies to the golang HTTP handler.
	nodeJSWriter, err := zipWriter.Create("index.js")
	if err != nil {
		return errors.New("Failed to create ZIP entry: index.js")
	}
	nodeJSSource := _escFSMustString(false, "/resources/index.js")
	nodeJSSource += "\n// DO NOT EDIT - CONTENT UNTIL EOF IS AUTOMATICALLY GENERATED\n"

	handlerNames := make(map[string]bool, 0)
	for _, eachLambda := range lambdaAWSInfos {
		if _, exists := handlerNames[eachLambda.jsHandlerName()]; !exists {
			nodeJSSource += createNewNodeJSProxyEntry(eachLambda, logger)
			handlerNames[eachLambda.jsHandlerName()] = true
		}

		// USER DEFINED RESOURCES
		for _, eachCustomResource := range eachLambda.customResources {
			if _, exists := handlerNames[eachCustomResource.jsHandlerName()]; !exists {
				nodeJSSource += createUserCustomResourceEntry(eachCustomResource, logger)
				handlerNames[eachCustomResource.jsHandlerName()] = true
			}
		}
	}
	// SPARTA CUSTOM RESOURCES
	for _, eachCustomResourceName := range golangCustomResourceTypes {
		nodeJSSource += createNewSpartaCustomResourceEntry(eachCustomResourceName, logger)
	}

	// Finally, replace
	// 	SPARTA_BINARY_NAME = 'Sparta.lambda.amd64';
	// with the service binary name
	nodeJSSource += fmt.Sprintf("SPARTA_BINARY_NAME='%s';\n", executableOutput)
	// And the service name
	nodeJSSource += fmt.Sprintf("SPARTA_SERVICE_NAME='%s';\n", serviceName)
	logger.WithFields(logrus.Fields{
		"index.js": nodeJSSource,
	}).Debug("Dynamically generated NodeJS adapter")

	stringReader := strings.NewReader(nodeJSSource)
	_, copyErr := io.Copy(nodeJSWriter, stringReader)
	return copyErr
}

func writeCustomResources(zipWriter *zip.Writer,
	logger *logrus.Logger) error {
	for _, eachName := range customResourceScripts {
		resourceName := fmt.Sprintf("%s/%s", provisioningResourcesRelPath, eachName)
		resourceContent := _escFSMustString(false, resourceName)
		stringReader := strings.NewReader(resourceContent)
		embedWriter, errCreate := zipWriter.Create(eachName)
		if nil != errCreate {
			return errCreate
		}
		logger.WithFields(logrus.Fields{
			"Name": eachName,
		}).Debug("Script name")

		_, copyErr := io.Copy(embedWriter, stringReader)
		if nil != copyErr {
			return copyErr
		}
	}
	return nil
}

// Build and package the application
func createPackageStep() workflowStep {

	return func(ctx *workflowContext) (workflowStep, error) {

		// PreBuild Hook
		if ctx.workflowHooks != nil {
			preBuildErr := callWorkflowHook(ctx.workflowHooks.PreBuild, ctx)
			if nil != preBuildErr {
				return nil, preBuildErr
			}
		}
		sanitizedServiceName := sanitizedName(ctx.serviceName)
		executableOutput := fmt.Sprintf("%s.lambda.amd64", sanitizedServiceName)
		buildErr := buildGoBinary(executableOutput, ctx.buildTags, ctx.linkFlags, ctx.logger)
		if nil != buildErr {
			return nil, buildErr
		}
		// Cleanup the temporary binary
		defer func() {
			errRemove := os.Remove(executableOutput)
			if nil != errRemove {
				ctx.logger.WithFields(logrus.Fields{
					"File":  executableOutput,
					"Error": errRemove,
				}).Warn("Failed to delete binary")
			}
		}()

		// Binary size
		logFilesize("Executable binary size", executableOutput, ctx.logger)

		// PostBuild Hook
		if ctx.workflowHooks != nil {
			postBuildErr := callWorkflowHook(ctx.workflowHooks.PostBuild, ctx)
			if nil != postBuildErr {
				return nil, postBuildErr
			}
		}
		tmpFile, err := temporaryFile(fmt.Sprintf("%s-code.zip", sanitizedServiceName))
		if err != nil {
			return nil, err
		}
		ctx.logger.WithFields(logrus.Fields{
			"TempName": tmpFile.Name(),
		}).Info("Creating code ZIP archive for upload")

		lambdaArchive := zip.NewWriter(tmpFile)

		// Archive Hook
		if ctx.workflowHooks != nil && ctx.workflowHooks.Archive != nil {
			archiveErr := ctx.workflowHooks.Archive(ctx.workflowHooksContext,
				ctx.serviceName,
				lambdaArchive,
				ctx.awsSession,
				ctx.noop,
				ctx.logger)
			if nil != archiveErr {
				return nil, archiveErr
			}
		}

		// File info for the binary executable
		readerErr := spartaZip.AddToZip(lambdaArchive,
			executableOutput,
			"",
			ctx.logger)
		if nil != readerErr {
			return nil, readerErr
		}

		// Add the string literal adapter, which requires us to add exported
		// functions to the end of index.js.  These NodeJS exports will be
		// linked to the AWS Lambda NodeJS function name, and are basically
		// automatically generated pass through proxies to the golang HTTP handler.
		shimErr := writeNodeJSShim(ctx.serviceName,
			executableOutput,
			ctx.lambdaAWSInfos,
			lambdaArchive,
			ctx.logger)
		if nil != shimErr {
			return nil, shimErr
		}

		// Next embed the custom resource scripts into the package.
		ctx.logger.Debug("Embedding CustomResource scripts")
		customResourceErr := writeCustomResources(lambdaArchive, ctx.logger)
		if nil != customResourceErr {
			return nil, customResourceErr
		}
		archiveCloseErr := lambdaArchive.Close()
		if nil != archiveCloseErr {
			return nil, archiveCloseErr
		}
		tempfileCloseErr := tmpFile.Close()
		if nil != tempfileCloseErr {
			return nil, tempfileCloseErr
		}
		return createUploadStep(tmpFile.Name()), nil
	}
}

// Given the zipped binary in packagePath, upload the primary code bundle
// and optional S3 site resources iff they're defined.
func createUploadStep(packagePath string) workflowStep {
	return func(ctx *workflowContext) (workflowStep, error) {
		var uploadErrors []error
		var wg sync.WaitGroup

		// We always need to upload the primary binary
		wg.Add(1)
		go func() {
			defer wg.Done()
			logFilesize("Lambda function deployment package size", packagePath, ctx.logger)

			// Create the S3 key...
			zipS3URL, zipS3URLErr := uploadLocalFileToS3(packagePath, "", ctx)
			if nil != zipS3URLErr {
				uploadErrors = append(uploadErrors, zipS3URLErr)
			} else {
				ctx.s3CodeZipURL = newS3UploadURL(zipS3URL)
			}
		}()

		// S3 site to compress & upload
		if nil != ctx.s3SiteContext.s3Site {
			wg.Add(1)
			go func() {
				defer wg.Done()

				tempName := fmt.Sprintf("%s-S3Site.zip", ctx.serviceName)
				tmpFile, err := temporaryFile(tempName)
				if err != nil {
					uploadErrors = append(uploadErrors,
						errors.New("Failed to create temporary S3 site archive file"))
					return
				}

				// Add the contents to the Zip file
				zipArchive := zip.NewWriter(tmpFile)
				absResourcePath, err := filepath.Abs(ctx.s3SiteContext.s3Site.resources)
				if nil != err {
					uploadErrors = append(uploadErrors, err)
					return
				}

				ctx.logger.WithFields(logrus.Fields{
					"S3Key":  path.Base(tmpFile.Name()),
					"Source": absResourcePath,
				}).Info("Creating S3Site archive")

				err = spartaZip.AddToZip(zipArchive, absResourcePath, absResourcePath, ctx.logger)
				if nil != err {
					uploadErrors = append(uploadErrors, err)
					return
				}
				zipArchive.Close()

				// Upload it & save the key
				s3SiteLambdaZipURL, s3SiteLambdaZipURLErr := uploadLocalFileToS3(tmpFile.Name(), "", ctx)
				if s3SiteLambdaZipURLErr != nil {
					uploadErrors = append(uploadErrors, s3SiteLambdaZipURLErr)
				} else {
					ctx.s3SiteContext.s3UploadURL = newS3UploadURL(s3SiteLambdaZipURL)
				}
				ctx.registerFileCleanupFinalizer(tmpFile.Name())
			}()
		}
		wg.Wait()

		if len(uploadErrors) > 0 {
			errorText := "Encountered multiple errors during upload:\n"
			for _, eachError := range uploadErrors {
				errorText += fmt.Sprintf("%s%s\n", errorText, eachError.Error())
				return nil, errors.New(errorText)
			}
		}
		return ensureCloudFormationStack(), nil
	}
}

func annotateDiscoveryInfo(template *gocf.Template, logger *logrus.Logger) *gocf.Template {
	for eachResourceID, eachResource := range template.Resources {
		// Only apply this to lambda functions
		if eachResource.Properties.CfnResourceType() == "AWS::Lambda::Function" {

			// Update the metdata with a reference to the output of each
			// depended on item...
			for _, eachDependsKey := range eachResource.DependsOn {
				dependencyOutputs, _ := outputsForResource(template, eachDependsKey, logger)
				if nil != dependencyOutputs && len(dependencyOutputs) != 0 {
					logger.WithFields(logrus.Fields{
						"Resource":  eachDependsKey,
						"DependsOn": eachResource.DependsOn,
						"Outputs":   dependencyOutputs,
					}).Debug("Resource metadata")
					safeMetadataInsert(eachResource, eachDependsKey, dependencyOutputs)
				}
			}
			// Also include standard AWS outputs at a resource level if a lambda
			// needs to self-discover other resources
			safeMetadataInsert(eachResource, TagLogicalResourceID, gocf.String(eachResourceID))
			safeMetadataInsert(eachResource, TagStackRegion, gocf.Ref("AWS::Region"))
			safeMetadataInsert(eachResource, TagStackID, gocf.Ref("AWS::StackId"))
			safeMetadataInsert(eachResource, TagStackName, gocf.Ref("AWS::StackName"))
		}
	}
	return template
}

// createCodePipelineTriggerPackage handles marshaling the template, zipping
// the config files in the package, and the
func createCodePipelineTriggerPackage(cfTemplateJSON []byte, ctx *workflowContext) (string, error) {
	sanitizedServiceName := sanitizedName(ctx.serviceName)
	tmpFile, err := temporaryFile(fmt.Sprintf("%s-pipeline.zip", sanitizedServiceName))
	if err != nil {
		return "", err
	}
	templateArchive := zip.NewWriter(tmpFile)
	ctx.logger.WithFields(logrus.Fields{
		"Path": tmpFile.Name(),
	}).Info("Creating CodePipeline archive")

	// File info for the binary executable
	zipEntryName := "cloudformation.json"
	bytesWriter, bytesWriterErr := templateArchive.Create(zipEntryName)
	if bytesWriterErr != nil {
		return "", bytesWriterErr
	}

	bytesReader := bytes.NewReader(cfTemplateJSON)
	written, writtenErr := io.Copy(bytesWriter, bytesReader)
	if nil != writtenErr {
		return "", writtenErr
	}
	ctx.logger.WithFields(logrus.Fields{
		"WrittenBytes": written,
		"ZipName":      zipEntryName,
	}).Debug("Archiving file")

	// If there is a codePipelineEnvironments defined, then we'll need to get all the
	// maps, marshal them to JSON, then add the JSON to the ZIP archive.
	if nil != codePipelineEnvironments {
		for eachEnvironment, eachMap := range codePipelineEnvironments {
			codePipelineParameters := map[string]interface{}{
				"Parameters": eachMap,
			}
			environmentJSON, environmentJSONErr := json.Marshal(codePipelineParameters)
			if nil != environmentJSONErr {
				ctx.logger.Error("Failed to Marshal CodePipeline environment: " + eachEnvironment)
				return "", environmentJSONErr
			}
			var envVarName = fmt.Sprintf("%s.json", eachEnvironment)

			// File info for the binary executable
			binaryWriter, binaryWriterErr := templateArchive.Create(envVarName)
			if binaryWriterErr != nil {
				return "", binaryWriterErr
			}
			_, writeErr := binaryWriter.Write(environmentJSON)
			if writeErr != nil {
				return "", writeErr
			}
		}
	}
	archiveCloseErr := templateArchive.Close()
	if nil != archiveCloseErr {
		return "", archiveCloseErr
	}
	tempfileCloseErr := tmpFile.Close()
	if nil != tempfileCloseErr {
		return "", tempfileCloseErr
	}
	return uploadLocalFileToS3(tmpFile.Name(), ctx.codePipelineTrigger, ctx)
}

func applyCloudFormationOperation(ctx *workflowContext) (workflowStep, error) {

	stackTags := map[string]string{
		SpartaTagHomeKey:    "http://gosparta.io",
		SpartaTagVersionKey: SpartaVersion,
		SpartaTagBuildIDKey: ctx.buildID,
	}
	if len(ctx.buildTags) != 0 {
		stackTags[SpartaTagBuildTagsKey] = ctx.buildTags
	}
	// Generate the CF template...
	cfTemplate, err := json.Marshal(ctx.cfTemplate)
	if err != nil {
		ctx.logger.Error("Failed to Marshal CloudFormation template: ", err.Error())
		return nil, err
	}

	// Consistent naming of template
	sanitizedServiceName := sanitizedName(ctx.serviceName)
	templateName := fmt.Sprintf("%s-cftemplate.json", sanitizedServiceName)
	templateFile, templateFileErr := temporaryFile(templateName)
	if nil != templateFileErr {
		return nil, templateFileErr
	}
	_, writeErr := templateFile.Write(cfTemplate)
	if nil != writeErr {
		return nil, writeErr
	}
	templateFile.Close()

	// Log the template if needed
	if nil != ctx.templateWriter || ctx.logger.Level <= logrus.DebugLevel {

		templateBody := string(cfTemplate)
		formatted, formattedErr := json.MarshalIndent(templateBody, "", " ")
		if nil != formattedErr {
			return nil, formattedErr
		}
		ctx.logger.WithFields(logrus.Fields{
			"Body": string(formatted),
		}).Debug("CloudFormation template body")
		if nil != ctx.templateWriter {
			io.WriteString(ctx.templateWriter, string(formatted))
		}
	}

	if "" == ctx.codePipelineTrigger {
		if ctx.noop {
			ctx.logger.WithFields(logrus.Fields{
				"Bucket":       ctx.s3Bucket,
				"TemplateName": templateName,
			}).Info("Bypassing Stack creation due to -n/-noop command line argument")
		} else {
			// Dump the template to a file, then upload it...
			uploadURL, uploadURLErr := uploadLocalFileToS3(templateFile.Name(), "", ctx)
			if nil != uploadURLErr {
				return nil, uploadURLErr
			}

			stack, stackErr := spartaCF.ConvergeStackState(ctx.serviceName,
				ctx.cfTemplate,
				uploadURL,
				stackTags,
				ctx.buildTime,
				ctx.awsSession,
				ctx.logger)
			if nil != stackErr {
				return nil, stackErr
			}
			ctx.logger.WithFields(logrus.Fields{
				"StackName":    *stack.StackName,
				"StackId":      *stack.StackId,
				"CreationTime": *stack.CreationTime,
			}).Info("Stack provisioned")
		}
	} else {
		// Cleanup the template...
		ctx.registerFileCleanupFinalizer(templateFile.Name())
		_, urlErr := createCodePipelineTriggerPackage(cfTemplate, ctx)
		if nil != urlErr {
			return nil, urlErr
		}
	}
	return nil, nil
}

func annotateCodePipelineEnvironments(lambdaAWSInfo *LambdaAWSInfo, logger *logrus.Logger) {
	if nil != codePipelineEnvironments {
		if nil == lambdaAWSInfo.Options {
			lambdaAWSInfo.Options = defaultLambdaFunctionOptions()
		}
		if nil == lambdaAWSInfo.Options.Environment {
			lambdaAWSInfo.Options.Environment = make(map[string]*gocf.StringExpr, 0)
		}
		for _, eachEnvironment := range codePipelineEnvironments {

			logger.WithFields(logrus.Fields{
				"Environment":    eachEnvironment,
				"LambdaFunction": lambdaAWSInfo.lambdaFunctionName(),
			}).Debug("Annotating Lambda environment for CodePipeline")

			for eachKey := range eachEnvironment {
				lambdaAWSInfo.Options.Environment[eachKey] = gocf.Ref(eachKey).String()
			}
		}
	}
}

func ensureCloudFormationStack() workflowStep {
	return func(ctx *workflowContext) (workflowStep, error) {
		// PreMarshall Hook
		if ctx.workflowHooks != nil {
			preMarshallErr := callWorkflowHook(ctx.workflowHooks.PreMarshall, ctx)
			if nil != preMarshallErr {
				return nil, preMarshallErr
			}
		}

		// Add the "Parameters" to the template...
		if nil != codePipelineEnvironments {
			ctx.cfTemplate.Parameters = make(map[string]*gocf.Parameter, 0)
			for _, eachEnvironment := range codePipelineEnvironments {
				for eachKey := range eachEnvironment {
					ctx.cfTemplate.Parameters[eachKey] = &gocf.Parameter{
						Type:    "String",
						Default: "",
					}
				}
			}
		}

		for _, eachEntry := range ctx.lambdaAWSInfos {
			annotateCodePipelineEnvironments(eachEntry, ctx.logger)

			err := eachEntry.export(ctx.serviceName,
				ctx.s3Bucket,
				ctx.s3CodeZipURL.keyName(),
				ctx.buildID,
				ctx.lambdaIAMRoleNameMap,
				ctx.cfTemplate,
				ctx.workflowHooksContext,
				ctx.logger)
			if nil != err {
				return nil, err
			}
		}
		// If there's an API gateway definition, include the resources that provision it. Since this export will likely
		// generate outputs that the s3 site needs, we'll use a temporary outputs accumulator, pass that to the S3Site
		// if it's defined, and then merge it with the normal output map.
		apiGatewayTemplate := gocf.NewTemplate()

		if nil != ctx.api {
			err := ctx.api.export(
				ctx.serviceName,
				ctx.awsSession,
				ctx.s3Bucket,
				ctx.s3CodeZipURL.keyName(),
				ctx.lambdaIAMRoleNameMap,
				apiGatewayTemplate,
				ctx.noop,
				ctx.logger)
			if nil == err {
				err = safeMergeTemplates(apiGatewayTemplate, ctx.cfTemplate, ctx.logger)
			}
			if nil != err {
				return nil, fmt.Errorf("Failed to export APIGateway template resources")
			}
		}
		// If there's a Site defined, include the resources the provision it
		if nil != ctx.s3SiteContext.s3Site {
			ctx.s3SiteContext.s3Site.export(ctx.serviceName,
				ctx.s3Bucket,
				ctx.s3CodeZipURL.keyName(),
				ctx.s3SiteContext.s3UploadURL.keyName(),
				apiGatewayTemplate.Outputs,
				ctx.lambdaIAMRoleNameMap,
				ctx.cfTemplate,
				ctx.logger)
		}
		// Service decorator?
		// If there's an API gateway definition, include the resources that provision it. Since this export will likely
		// generate outputs that the s3 site needs, we'll use a temporary outputs accumulator, pass that to the S3Site
		// if it's defined, and then merge it with the normal output map.-
		if nil != ctx.workflowHooks && nil != ctx.workflowHooks.ServiceDecorator {
			hookName := runtime.FuncForPC(reflect.ValueOf(ctx.workflowHooks.ServiceDecorator).Pointer()).Name()
			ctx.logger.WithFields(logrus.Fields{
				"WorkflowHook":        hookName,
				"WorkflowHookContext": ctx.workflowHooksContext,
			}).Info("Calling WorkflowHook")

			serviceTemplate := gocf.NewTemplate()
			decoratorError := ctx.workflowHooks.ServiceDecorator(
				ctx.workflowHooksContext,
				ctx.serviceName,
				serviceTemplate,
				ctx.s3Bucket,
				ctx.buildID,
				ctx.awsSession,
				ctx.noop,
				ctx.logger,
			)
			if nil != decoratorError {
				return nil, decoratorError
			}
			mergeErr := safeMergeTemplates(serviceTemplate, ctx.cfTemplate, ctx.logger)
			if nil != mergeErr {
				return nil, mergeErr
			}
		}
		ctx.cfTemplate = annotateDiscoveryInfo(ctx.cfTemplate, ctx.logger)

		// PostMarshall Hook
		if ctx.workflowHooks != nil {
			postMarshallErr := callWorkflowHook(ctx.workflowHooks.PostMarshall, ctx)
			if nil != postMarshallErr {
				return nil, postMarshallErr
			}
		}
		return applyCloudFormationOperation(ctx)
	}
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
// More information on golang 1.5's support for vendor'd resources is documented at
//
//  https://docs.google.com/document/d/1Bz5-UB7g2uPBdOx-rw5t9MxJwkfpx90cqG9AFL0JAYo/edit
//  https://medium.com/@freeformz/go-1-5-s-vendor-experiment-fd3e830f52c3#.voiicue1j
//
// type Configuration struct {
//     Val   string
//     Proxy struct {
//         Address string
//         Port    string
//     }
// }
func Provision(noop bool,
	serviceName string,
	serviceDescription string,
	lambdaAWSInfos []*LambdaAWSInfo,
	api *API,
	site *S3Site,
	s3Bucket string,
	buildID string,
	codePipelineTrigger string,
	buildTags string,
	linkerFlags string,
	templateWriter io.Writer,
	workflowHooks *WorkflowHooks,
	logger *logrus.Logger) error {

	err := validateSpartaPreconditions(lambdaAWSInfos, logger)
	if nil != err {
		return err
	}
	startTime := time.Now()

	ctx := &workflowContext{
		noop:               noop,
		serviceName:        serviceName,
		serviceDescription: serviceDescription,
		lambdaAWSInfos:     lambdaAWSInfos,
		api:                api,
		s3SiteContext: &s3SiteContext{
			s3Site: site,
		},
		cfTemplate:                gocf.NewTemplate(),
		s3Bucket:                  s3Bucket,
		s3BucketVersioningEnabled: false,
		buildID:                   buildID,
		codePipelineTrigger:       codePipelineTrigger,
		buildTags:                 buildTags,
		linkFlags:                 linkerFlags,
		buildTime:                 time.Now(),
		awsSession:                spartaAWS.NewSession(logger),
		templateWriter:            templateWriter,
		workflowHooks:             workflowHooks,
		workflowHooksContext:      make(map[string]interface{}, 0),
		logger:                    logger,
	}
	ctx.cfTemplate.Description = serviceDescription

	// Update the context iff it exists
	if nil != workflowHooks && nil != workflowHooks.Context {
		for eachKey, eachValue := range workflowHooks.Context {
			ctx.workflowHooksContext[eachKey] = eachValue
		}
	}

	ctx.logger.WithFields(logrus.Fields{
		"BuildID":             buildID,
		"NOOP":                noop,
		"Tags":                ctx.buildTags,
		"CodePipelineTrigger": ctx.codePipelineTrigger,
	}).Info("Provisioning service")

	if len(lambdaAWSInfos) <= 0 {
		return errors.New("No lambda functions provided to Sparta.Provision()")
	}

	// Start the workflow
	for step := verifyIAMRoles; step != nil; {
		next, err := step(ctx)
		if err != nil {
			ctx.rollback()
			// Workflow step?
			ctx.logger.Error(err)
			return err
		}
		if next == nil {
			elapsed := time.Since(startTime)
			ctx.logger.WithFields(logrus.Fields{
				"Seconds": fmt.Sprintf("%.f", elapsed.Seconds()),
			}).Info("Elapsed time")
			break
		} else {
			step = next
		}
	}
	// When we're done, execute any finalizers
	if nil != ctx.finalizerFunctions {
		ctx.logger.WithFields(logrus.Fields{
			"FinalizerCount": len(ctx.finalizerFunctions),
		}).Debug("Invoking finalizer functions")
		for _, eachFinalizer := range ctx.finalizerFunctions {
			eachFinalizer(ctx.logger)
		}
	}
	return nil
}
