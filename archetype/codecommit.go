package archetype

import (
	"context"
	"reflect"
	"runtime"

	awsLambdaEvents "github.com/aws/aws-lambda-go/events"
	sparta "github.com/mweagle/Sparta"
	gocf "github.com/mweagle/go-cloudformation"
	"github.com/pkg/errors"
)

// CodeCommitReactor represents a lambda function that responds to CodeCommit events
type CodeCommitReactor interface {
	// OnCodeCommitEvent when an SNS event occurs. Check the codeCommitEvent field
	// for the specific event
	OnCodeCommitEvent(ctx context.Context, codeCommitEvent awsLambdaEvents.CodeCommitEvent) (interface{}, error)
}

// CodeCommitReactorFunc is a free function that adapts a CodeCommitReactor
// compliant signature into a function that exposes an OnEvent
// function
type CodeCommitReactorFunc func(ctx context.Context,
	codeCommitEvent awsLambdaEvents.CodeCommitEvent) (interface{}, error)

// OnCodeCommitEvent satisfies the CodeCommitReactor interface
func (reactorFunc CodeCommitReactorFunc) OnCodeCommitEvent(ctx context.Context,
	codeCommitEvent awsLambdaEvents.CodeCommitEvent) (interface{}, error) {
	return reactorFunc(ctx, codeCommitEvent)
}

// ReactorName provides the name of the reactor func
func (reactorFunc CodeCommitReactorFunc) ReactorName() string {
	return runtime.FuncForPC(reflect.ValueOf(reactorFunc).Pointer()).Name()
}

// NewCodeCommitReactor returns an SNS reactor lambda function
func NewCodeCommitReactor(reactor CodeCommitReactor,
	repositoryName gocf.Stringable,
	branches []string,
	events []string,
	additionalLambdaPermissions []sparta.IAMRolePrivilege) (*sparta.LambdaAWSInfo, error) {

	reactorLambda := func(ctx context.Context, codeCommitEvent awsLambdaEvents.CodeCommitEvent) (interface{}, error) {
		return reactor.OnCodeCommitEvent(ctx, codeCommitEvent)
	}

	lambdaFn, lambdaFnErr := sparta.NewAWSLambda(reactorName(reactor),
		reactorLambda,
		sparta.IAMRoleDefinition{})
	if lambdaFnErr != nil {
		return nil, errors.Wrapf(lambdaFnErr, "attempting to create reactor")
	}

	lambdaFn.Permissions = append(lambdaFn.Permissions, sparta.CodeCommitPermission{
		BasePermission: sparta.BasePermission{
			SourceArn: repositoryName,
		},
		RepositoryName: repositoryName.String(),
		Branches:       branches,
		Events:         events,
	})
	if len(additionalLambdaPermissions) != 0 {
		lambdaFn.RoleDefinition.Privileges = additionalLambdaPermissions
	}
	return lambdaFn, nil
}
