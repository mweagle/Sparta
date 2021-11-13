package archetype

import (
	"context"
	"reflect"
	"runtime"

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	awsv2S3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"

	awsLambdaEvents "github.com/aws/aws-lambda-go/events"
	sparta "github.com/mweagle/Sparta/v3"
	spartaCF "github.com/mweagle/Sparta/v3/aws/cloudformation"
	"github.com/pkg/errors"
)

// ReactorNameProvider is an interface so that a reactor function can
// provide a custom name which prevents collisions
type ReactorNameProvider interface {
	ReactorName() string
}

// S3Reactor represents a lambda function that responds to typical S3 operations
type S3Reactor interface {
	// OnS3Event when an S3 event occurs. Check the event.EventName field
	// for the specific event
	OnS3Event(ctx context.Context, event awsLambdaEvents.S3Event) (interface{}, error)
}

// S3ReactorFunc is a free function that adapts a S3Reactor
// compliant signature into a function that exposes an OnS3Event
// function
type S3ReactorFunc func(ctx context.Context, event awsLambdaEvents.S3Event) (interface{}, error)

// OnS3Event satisfies the S3Reactor interface
func (reactorFunc S3ReactorFunc) OnS3Event(ctx context.Context, event awsLambdaEvents.S3Event) (interface{}, error) {
	return reactorFunc(ctx, event)
}

// ReactorName provides the name of the reactor func
func (reactorFunc S3ReactorFunc) ReactorName() string {
	return runtime.FuncForPC(reflect.ValueOf(reactorFunc).Pointer()).Name()
}

// s3NotificationPrefixFilter is a DRY spec for setting up a notification configuration
// filter
func s3NotificationPrefixBasedPermission(bucketName string, keyPathPrefix string) sparta.S3Permission {

	permission := sparta.S3Permission{
		BasePermission: sparta.BasePermission{
			SourceArn: bucketName,
		},
		Events: []string{"s3:ObjectCreated:*",
			"s3:ObjectRemoved:*"},
	}

	if keyPathPrefix != "" {
		permission.Filter = awsv2S3Types.NotificationConfigurationFilter{
			Key: &awsv2S3Types.S3KeyFilter{
				FilterRules: []awsv2S3Types.FilterRule{
					{
						Name:  "prefix",
						Value: awsv2.String(keyPathPrefix),
					},
				},
			},
		}
	}
	return permission
}

// NewS3Reactor returns an S3 reactor lambda function
func NewS3Reactor(reactor S3Reactor, s3Bucket string, additionalLambdaPermissions []sparta.IAMRolePrivilege) (*sparta.LambdaAWSInfo, error) {
	return NewS3ScopedReactor(reactor, s3Bucket, "", additionalLambdaPermissions)
}

// NewS3ScopedReactor returns an S3 reactor lambda function scoped to the given S3 key prefix
func NewS3ScopedReactor(reactor S3Reactor,
	s3Bucket string,
	keyPathPrefix string,
	additionalLambdaPermissions []sparta.IAMRolePrivilege) (*sparta.LambdaAWSInfo, error) {

	reactorLambda := func(ctx context.Context, event awsLambdaEvents.S3Event) (interface{}, error) {
		return reactor.OnS3Event(ctx, event)
	}

	// Privilege must include access to the S3 bucket for GetObjectRequest
	lambdaFn, lambdaFnErr := sparta.NewAWSLambda(reactorName(reactor),
		reactorLambda,
		sparta.IAMRoleDefinition{})
	if lambdaFnErr != nil {
		return nil, errors.Wrapf(lambdaFnErr, "attempting to create reactor")
	}

	privileges := []sparta.IAMRolePrivilege{{
		Actions:  []string{"s3:GetObject"},
		Resource: spartaCF.S3AllKeysArnForBucket(s3Bucket),
	}}
	if len(additionalLambdaPermissions) != 0 {
		privileges = append(privileges, additionalLambdaPermissions...)
	}

	// IAM Role privileges
	lambdaFn.RoleDefinition.Privileges = privileges

	// Event Triggers
	lambdaFn.Permissions = append(lambdaFn.Permissions,
		s3NotificationPrefixBasedPermission(s3Bucket, keyPathPrefix))

	return lambdaFn, nil
}
