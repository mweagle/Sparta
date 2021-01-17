package sparta

import (
	"archive/zip"
	"context"

	"github.com/aws/aws-sdk-go/aws/session"
	gocf "github.com/mweagle/go-cloudformation"
	"github.com/rs/zerolog"
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
type TemplateDecorator func(ctx context.Context,
	serviceName string,
	lambdaResourceName string,
	lambdaResource gocf.LambdaFunction,
	resourceMetadata map[string]interface{},
	lambdaFunctionCode *gocf.LambdaFunctionCode,
	buildID string,
	template *gocf.Template,
	logger *zerolog.Logger) (context.Context, error)

// TemplateDecoratorHookFunc is the adapter to transform an existing
// TemplateHook into a TemplateDecoratorHandler satisfier
type TemplateDecoratorHookFunc func(ctx context.Context,
	serviceName string,
	lambdaResourceName string,
	lambdaResource gocf.LambdaFunction,
	resourceMetadata map[string]interface{},
	lambdaFunctionCode *gocf.LambdaFunctionCode,
	buildID string,
	template *gocf.Template,
	logger *zerolog.Logger) (context.Context, error)

// DecorateTemplate calls tdhf(...) to satisfy TemplateDecoratorHandler
func (tdhf TemplateDecoratorHookFunc) DecorateTemplate(ctx context.Context,
	serviceName string,
	lambdaResourceName string,
	lambdaResource gocf.LambdaFunction,
	resourceMetadata map[string]interface{},
	lambdaFunctionCode *gocf.LambdaFunctionCode,
	buildID string,
	template *gocf.Template,
	logger *zerolog.Logger) (context.Context, error) {
	return tdhf(ctx,
		serviceName,
		lambdaResourceName,
		lambdaResource,
		resourceMetadata,
		lambdaFunctionCode,
		buildID,
		template,
		logger)
}

// TemplateDecoratorHandler is the interface type to indicate a template
// decoratorHook
type TemplateDecoratorHandler interface {
	DecorateTemplate(ctx context.Context,
		serviceName string,
		lambdaResourceName string,
		lambdaResource gocf.LambdaFunction,
		resourceMetadata map[string]interface{},
		lambdaFunctionCode *gocf.LambdaFunctionCode,
		buildID string,
		template *gocf.Template,
		logger *zerolog.Logger) (context.Context, error)
}

////////////////////////////////////////////////////////////////////////////////
// WorkflowHandler

// WorkflowHook defines a user function that should be called at a specific
// point in the larger Sparta workflow. The first argument is a map that
// is shared across all LifecycleHooks and which Sparta treats as an opaque
// value.
type WorkflowHook func(ctx context.Context,
	serviceName string,
	S3Bucket gocf.Stringable,
	buildID string,
	awsSession *session.Session,
	noop bool,
	logger *zerolog.Logger) (context.Context, error)

// WorkflowHookFunc is the adapter to transform an existing
// WorkflowHook into a WorkflowHookHandler satisfier
type WorkflowHookFunc func(ctx context.Context,
	serviceName string,
	S3Bucket gocf.Stringable,
	buildID string,
	awsSession *session.Session,
	noop bool,
	logger *zerolog.Logger) (context.Context, error)

// DecorateWorkflow calls whf(...) to satisfy WorkflowHookHandler
func (whf WorkflowHookFunc) DecorateWorkflow(ctx context.Context,
	serviceName string,
	S3Bucket gocf.Stringable,
	buildID string,
	awsSession *session.Session,
	noop bool,
	logger *zerolog.Logger) (context.Context, error) {
	return whf(ctx,
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
	DecorateWorkflow(ctx context.Context,
		serviceName string,
		S3Bucket gocf.Stringable,
		buildID string,
		awsSession *session.Session,
		noop bool,
		logger *zerolog.Logger) (context.Context, error)
}

////////////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////////////
// ArchiveHandler

// ArchiveHook provides callers an opportunity to insert additional
// files into the ZIP archive deployed to S3
type ArchiveHook func(ctx context.Context,
	serviceName string,
	zipWriter *zip.Writer,
	awsSession *session.Session,
	noop bool,
	logger *zerolog.Logger) (context.Context, error)

// ArchiveHookFunc is the adapter to transform an existing
// ArchiveHook into a WorkflowHookHandler satisfier
type ArchiveHookFunc func(ctx context.Context,
	serviceName string,
	zipWriter *zip.Writer,
	awsSession *session.Session,
	noop bool,
	logger *zerolog.Logger) (context.Context, error)

// DecorateArchive calls whf(...) to satisfy ArchiveHookHandler
func (ahf ArchiveHookFunc) DecorateArchive(ctx context.Context,
	serviceName string,
	zipWriter *zip.Writer,
	awsSession *session.Session,
	noop bool,
	logger *zerolog.Logger) (context.Context, error) {
	return ahf(ctx,
		serviceName,
		zipWriter,
		awsSession,
		noop,
		logger)
}

// ArchiveHookHandler is the interface type to indicate a workflow
// hook
type ArchiveHookHandler interface {
	DecorateArchive(ctx context.Context,
		serviceName string,
		zipWriter *zip.Writer,
		awsSession *session.Session,
		noop bool,
		logger *zerolog.Logger) (context.Context, error)
}

////////////////////////////////////////////////////////////////////////////////
// ServiceDecoratorHandler

// ServiceDecoratorHook defines a user function that is called a single
// time in the marshall workflow.
type ServiceDecoratorHook func(ctx context.Context,
	serviceName string,
	template *gocf.Template,
	lambdaFunctionCode *gocf.LambdaFunctionCode,
	buildID string,
	awsSession *session.Session,
	noop bool,
	logger *zerolog.Logger) (context.Context, error)

// ServiceDecoratorHookFunc is the adapter to transform an existing
// ArchiveHook into a WorkflowHookHandler satisfier
type ServiceDecoratorHookFunc func(ctx context.Context,
	serviceName string,
	template *gocf.Template,
	lambdaFunctionCode *gocf.LambdaFunctionCode,
	buildID string,
	awsSession *session.Session,
	noop bool,
	logger *zerolog.Logger) (context.Context, error)

// DecorateService calls sdhf(...) to satisfy ServiceDecoratorHookHandler
func (sdhf ServiceDecoratorHookFunc) DecorateService(ctx context.Context,
	serviceName string,
	template *gocf.Template,
	lambdaFunctionCode *gocf.LambdaFunctionCode,
	buildID string,
	awsSession *session.Session,
	noop bool,
	logger *zerolog.Logger) (context.Context, error) {
	return sdhf(ctx,
		serviceName,
		template,
		lambdaFunctionCode,
		buildID,
		awsSession,
		noop,
		logger)
}

// ServiceDecoratorHookHandler is the interface type to indicate a workflow
// hook
type ServiceDecoratorHookHandler interface {
	DecorateService(ctx context.Context,
		serviceName string,
		template *gocf.Template,
		lambdaFunctionCode *gocf.LambdaFunctionCode,
		buildID string,
		awsSession *session.Session,
		noop bool,
		logger *zerolog.Logger) (context.Context, error)
}

////////////////////////////////////////////////////////////////////////////////
// ServiceValidationHookHandler

// ServiceValidationHook defines a user function that is called a single
// after all template annotations have been performed. It is where
// policies should be applied
type ServiceValidationHook func(ctx context.Context,
	serviceName string,
	template *gocf.Template,
	lambdaFunctionCode *gocf.LambdaFunctionCode,
	buildID string,
	awsSession *session.Session,
	noop bool,
	logger *zerolog.Logger) (context.Context, error)

// ServiceValidationHookFunc is the adapter to transform an existing
// ArchiveHook into a WorkflowHookHandler satisfier
type ServiceValidationHookFunc func(ctx context.Context,
	serviceName string,
	template *gocf.Template,
	lambdaFunctionCode *gocf.LambdaFunctionCode,
	buildID string,
	awsSession *session.Session,
	noop bool,
	logger *zerolog.Logger) (context.Context, error)

// ValidateService calls sdhf(...) to satisfy ServiceValidationHookHandler
func (sdhf ServiceValidationHookFunc) ValidateService(ctx context.Context,
	serviceName string,
	template *gocf.Template,
	lambdaFunctionCode *gocf.LambdaFunctionCode,
	buildID string,
	awsSession *session.Session,
	noop bool,
	logger *zerolog.Logger) (context.Context, error) {
	return sdhf(ctx,
		serviceName,
		template,
		lambdaFunctionCode,
		buildID,
		awsSession,
		noop,
		logger)
}

// ServiceValidationHookHandler is the interface type to indicate a workflow
// hook
type ServiceValidationHookHandler interface {
	ValidateService(ctx context.Context,
		serviceName string,
		template *gocf.Template,
		lambdaFunctionCode *gocf.LambdaFunctionCode,
		buildID string,
		awsSession *session.Session,
		noop bool,
		logger *zerolog.Logger) (context.Context, error)
}

////////////////////////////////////////////////////////////////////////////////
// RollbackHandler

// RollbackHook provides callers an opportunity to handle failures
// associated with failing to perform the requested operation
type RollbackHook func(ctx context.Context,
	serviceName string,
	awsSession *session.Session,
	noop bool,
	logger *zerolog.Logger)

// RollbackHookFunc the adapter to transform an existing
// RollbackHook into a RollbackHookHandler satisfier
type RollbackHookFunc func(ctx context.Context,
	serviceName string,
	awsSession *session.Session,
	noop bool,
	logger *zerolog.Logger)

// Rollback calls sdhf(...) to satisfy ArchiveHookHandler
func (rhf RollbackHookFunc) Rollback(ctx context.Context,
	serviceName string,
	awsSession *session.Session,
	noop bool,
	logger *zerolog.Logger) (context.Context, error) {
	rhf(ctx,
		serviceName,
		awsSession,
		noop,
		logger)
	return ctx, nil
}

// RollbackHookHandler is the interface type to indicate a workflow
// hook
type RollbackHookHandler interface {
	Rollback(ctx context.Context,
		serviceName string,
		awsSession *session.Session,
		noop bool,
		logger *zerolog.Logger) (context.Context, error)
}
