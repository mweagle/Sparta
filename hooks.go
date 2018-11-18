package sparta

import (
	"archive/zip"

	"github.com/aws/aws-sdk-go/aws/session"
	gocf "github.com/mweagle/go-cloudformation"
	"github.com/sirupsen/logrus"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES
////////////////////////////////////////////////////////////////////////////////

// TemplateDecorator allows Lambda functions to annotate the CloudFormation
// template definition.  Both the resources and the outputs params
// are initialized to an empty ArbitraryJSONObject and should
// be populated with valid CloudFormation ArbitraryJSONObject values.  The
// CloudFormationResourceName() function can be used to generate
// logical CloudFormation-compatible resource names.
// See http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-template-resource-type-ref.html and
// http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/outputs-section-structure.html for
// more information.
type TemplateDecorator func(serviceName string,
	lambdaResourceName string,
	lambdaResource gocf.LambdaFunction,
	resourceMetadata map[string]interface{},
	S3Bucket string,
	S3Key string,
	buildID string,
	template *gocf.Template,
	context map[string]interface{},
	logger *logrus.Logger) error

// TemplateDecoratorHookFunc is the adapter to transform an existing
// ArchiveHook into a WorkflowHookHandler satisfier
type TemplateDecoratorHookFunc func(serviceName string,
	lambdaResourceName string,
	lambdaResource gocf.LambdaFunction,
	resourceMetadata map[string]interface{},
	S3Bucket string,
	S3Key string,
	buildID string,
	template *gocf.Template,
	context map[string]interface{},
	logger *logrus.Logger) error

// DecorateTemplate calls tdhf(...) to satisfy TemplateDecoratorHandler
func (tdhf TemplateDecoratorHookFunc) DecorateTemplate(serviceName string,
	lambdaResourceName string,
	lambdaResource gocf.LambdaFunction,
	resourceMetadata map[string]interface{},
	S3Bucket string,
	S3Key string,
	buildID string,
	template *gocf.Template,
	context map[string]interface{},
	logger *logrus.Logger) error {
	return tdhf(serviceName,
		lambdaResourceName,
		lambdaResource,
		resourceMetadata,
		S3Bucket,
		S3Key,
		buildID,
		template,
		context,
		logger)
}

// TemplateDecoratorHandler is the interface type to indicate a template
// decoratorHook
type TemplateDecoratorHandler interface {
	DecorateTemplate(serviceName string,
		lambdaResourceName string,
		lambdaResource gocf.LambdaFunction,
		resourceMetadata map[string]interface{},
		S3Bucket string,
		S3Key string,
		buildID string,
		template *gocf.Template,
		context map[string]interface{},
		logger *logrus.Logger) error
}

////////////////////////////////////////////////////////////////////////////////
// WorkflowHandler

// WorkflowHook defines a user function that should be called at a specific
// point in the larger Sparta workflow. The first argument is a map that
// is shared across all LifecycleHooks and which Sparta treats as an opaque
// value.
type WorkflowHook func(context map[string]interface{},
	serviceName string,
	S3Bucket string,
	buildID string,
	awsSession *session.Session,
	noop bool,
	logger *logrus.Logger) error

// WorkflowHookFunc is the adapter to transform an existing
// WorkflowHook into a WorkflowHookHandler satisfier
type WorkflowHookFunc func(context map[string]interface{},
	serviceName string,
	S3Bucket string,
	buildID string,
	awsSession *session.Session,
	noop bool,
	logger *logrus.Logger) error

// DecorateWorkflow calls whf(...) to satisfy WorkflowHookHandler
func (whf WorkflowHookFunc) DecorateWorkflow(context map[string]interface{},
	serviceName string,
	S3Bucket string,
	buildID string,
	awsSession *session.Session,
	noop bool,
	logger *logrus.Logger) error {
	return whf(context,
		serviceName,
		S3Bucket,
		buildID,
		awsSession,
		noop,
		logger)
}

// WorkflowHookHandler is the interface type to indicate a workflow
// hook
type WorkflowHookHandler interface {
	DecorateWorkflow(context map[string]interface{},
		serviceName string,
		S3Bucket string,
		buildID string,
		awsSession *session.Session,
		noop bool,
		logger *logrus.Logger) error
}

////////////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////////////
// ArchiveHandler

// ArchiveHook provides callers an opportunity to insert additional
// files into the ZIP archive deployed to S3
type ArchiveHook func(context map[string]interface{},
	serviceName string,
	zipWriter *zip.Writer,
	awsSession *session.Session,
	noop bool,
	logger *logrus.Logger) error

// ArchiveHookFunc is the adapter to transform an existing
// ArchiveHook into a WorkflowHookHandler satisfier
type ArchiveHookFunc func(context map[string]interface{},
	serviceName string,
	zipWriter *zip.Writer,
	awsSession *session.Session,
	noop bool,
	logger *logrus.Logger) error

// DecorateArchive calls whf(...) to satisfy ArchiveHookHandler
func (ahf ArchiveHookFunc) DecorateArchive(context map[string]interface{},
	serviceName string,
	zipWriter *zip.Writer,
	awsSession *session.Session,
	noop bool,
	logger *logrus.Logger) error {
	return ahf(context,
		serviceName,
		zipWriter,
		awsSession,
		noop,
		logger)
}

// ArchiveHookHandler is the interface type to indicate a workflow
// hook
type ArchiveHookHandler interface {
	DecorateArchive(context map[string]interface{},
		serviceName string,
		zipWriter *zip.Writer,
		awsSession *session.Session,
		noop bool,
		logger *logrus.Logger) error
}

////////////////////////////////////////////////////////////////////////////////
// ServiceDecoratorHandler

// ServiceDecoratorHook defines a user function that is called a single
// time in the marshall workflow.
type ServiceDecoratorHook func(context map[string]interface{},
	serviceName string,
	template *gocf.Template,
	S3Bucket string,
	S3Key string,
	buildID string,
	awsSession *session.Session,
	noop bool,
	logger *logrus.Logger) error

// ServiceDecoratorHookFunc is the adapter to transform an existing
// ArchiveHook into a WorkflowHookHandler satisfier
type ServiceDecoratorHookFunc func(context map[string]interface{},
	serviceName string,
	template *gocf.Template,
	S3Bucket string,
	S3Key string,
	buildID string,
	awsSession *session.Session,
	noop bool,
	logger *logrus.Logger) error

// DecorateService calls sdhf(...) to satisfy ArchiveHookHandler
func (sdhf ServiceDecoratorHookFunc) DecorateService(context map[string]interface{},
	serviceName string,
	template *gocf.Template,
	S3Bucket string,
	S3Key string,
	buildID string,
	awsSession *session.Session,
	noop bool,
	logger *logrus.Logger) error {
	return sdhf(context,
		serviceName,
		template,
		S3Bucket,
		S3Key,
		buildID,
		awsSession,
		noop,
		logger)
}

// ServiceDecoratorHookHandler is the interface type to indicate a workflow
// hook
type ServiceDecoratorHookHandler interface {
	DecorateService(context map[string]interface{},
		serviceName string,
		template *gocf.Template,
		S3Bucket string,
		S3Key string,
		buildID string,
		awsSession *session.Session,
		noop bool,
		logger *logrus.Logger) error
}

////////////////////////////////////////////////////////////////////////////////
// ServiceValidationHookHandler

// ServiceValidationHook defines a user function that is called a single
// after all template annotations have been performed. It is where
// policies should be applied
type ServiceValidationHook func(context map[string]interface{},
	serviceName string,
	template *gocf.Template,
	S3Bucket string,
	S3Key string,
	buildID string,
	awsSession *session.Session,
	noop bool,
	logger *logrus.Logger) error

// ServiceValidationHookFunc is the adapter to transform an existing
// ArchiveHook into a WorkflowHookHandler satisfier
type ServiceValidationHookFunc func(context map[string]interface{},
	serviceName string,
	template *gocf.Template,
	S3Bucket string,
	S3Key string,
	buildID string,
	awsSession *session.Session,
	noop bool,
	logger *logrus.Logger) error

// ValidateService calls sdhf(...) to satisfy ServiceValidationHookHandler
func (sdhf ServiceValidationHookFunc) ValidateService(context map[string]interface{},
	serviceName string,
	template *gocf.Template,
	S3Bucket string,
	S3Key string,
	buildID string,
	awsSession *session.Session,
	noop bool,
	logger *logrus.Logger) error {
	return sdhf(context,
		serviceName,
		template,
		S3Bucket,
		S3Key,
		buildID,
		awsSession,
		noop,
		logger)
}

// ServiceValidationHookHandler is the interface type to indicate a workflow
// hook
type ServiceValidationHookHandler interface {
	ValidateService(context map[string]interface{},
		serviceName string,
		template *gocf.Template,
		S3Bucket string,
		S3Key string,
		buildID string,
		awsSession *session.Session,
		noop bool,
		logger *logrus.Logger) error
}

////////////////////////////////////////////////////////////////////////////////
// RollbackHandler

// RollbackHook provides callers an opportunity to handle failures
// associated with failing to perform the requested operation
type RollbackHook func(context map[string]interface{},
	serviceName string,
	awsSession *session.Session,
	noop bool,
	logger *logrus.Logger)

// RollbackHookFunc the adapter to transform an existing
// RollbackHook into a RollbackHookHandler satisfier
type RollbackHookFunc func(context map[string]interface{},
	serviceName string,
	awsSession *session.Session,
	noop bool,
	logger *logrus.Logger)

// Rollback calls sdhf(...) to satisfy ArchiveHookHandler
func (rhf RollbackHookFunc) Rollback(context map[string]interface{},
	serviceName string,
	awsSession *session.Session,
	noop bool,
	logger *logrus.Logger) error {
	rhf(context,
		serviceName,
		awsSession,
		noop,
		logger)
	return nil
}

// RollbackHookHandler is the interface type to indicate a workflow
// hook
type RollbackHookHandler interface {
	Rollback(context map[string]interface{},
		serviceName string,
		awsSession *session.Session,
		noop bool,
		logger *logrus.Logger) error
}
