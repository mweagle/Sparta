package archetype

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"runtime"

	sparta "github.com/mweagle/Sparta"
	"github.com/pkg/errors"
)

// EventBridge represents a lambda function that responds to CW messages
type EventBridge interface {
	// OnLogMessage when an SNS event occurs. Check the snsEvent field
	// for the specific event
	OnEventBridgeBroadcast(ctx context.Context,
		msg json.RawMessage) (interface{}, error)
}

// EventBridgeReactorFunc is a free function that adapts a EventBridge
// compliant signature into a function that exposes an OnEvent
// function
type EventBridgeReactorFunc func(ctx context.Context,
	msg json.RawMessage) (interface{}, error)

// OnEventBridgeBroadcast satisfies the EventBridge interface
func (reactorFunc EventBridgeReactorFunc) OnEventBridgeBroadcast(ctx context.Context,
	msg json.RawMessage) (interface{}, error) {
	return reactorFunc(ctx, msg)
}

// ReactorName provides the name of the reactor func
func (reactorFunc EventBridgeReactorFunc) ReactorName() string {
	return runtime.FuncForPC(reflect.ValueOf(reactorFunc).Pointer()).Name()
}

// Standardized event bridge reactor ctor
func newEventBridgeEventReactor(reactor EventBridge,
	eventPattern map[string]interface{},
	scheduleExpression string,
	additionalLambdaPermissions []sparta.IAMRolePrivilege) (*sparta.LambdaAWSInfo, error) {

	reactorLambda := func(ctx context.Context, msg json.RawMessage) (interface{}, error) {
		return reactor.OnEventBridgeBroadcast(ctx, msg)
	}

	lambdaFn, lambdaFnErr := sparta.NewAWSLambda(reactorName(reactor),
		reactorLambda,
		sparta.IAMRoleDefinition{})
	if lambdaFnErr != nil {
		return nil, errors.Wrapf(lambdaFnErr, "attempting to create reactor")
	}
	if len(additionalLambdaPermissions) != 0 {
		lambdaFn.RoleDefinition.Privileges = additionalLambdaPermissions
	}

	eventBridgePermission := sparta.EventBridgePermission{}
	eventBridgePermission.Rule = &sparta.EventBridgeRule{
		Description:        fmt.Sprintf("EventBridge rule for %s", lambdaFn.LogicalResourceName()),
		EventPattern:       eventPattern,
		ScheduleExpression: scheduleExpression,
	}
	lambdaFn.Permissions = append(lambdaFn.Permissions, eventBridgePermission)
	return lambdaFn, nil
}

// NewEventBridgeEventReactor returns an EventBridge reactor function
// that responds to events (as opposed to schedules)
func NewEventBridgeEventReactor(reactor EventBridge,
	eventPattern map[string]interface{},
	additionalLambdaPermissions []sparta.IAMRolePrivilege) (*sparta.LambdaAWSInfo, error) {

	if len(eventPattern) <= 0 {
		return nil, errors.Errorf("EventBridge eventPattern map must not be empty")
	}
	return newEventBridgeEventReactor(reactor,
		eventPattern,
		"",
		additionalLambdaPermissions)
}

// NewEventBridgeScheduledReactor returns an EventBridge reactor function
// that responds to events (as opposed to schedules)
func NewEventBridgeScheduledReactor(reactor EventBridge,
	scheduleExpression string,
	additionalLambdaPermissions []sparta.IAMRolePrivilege) (*sparta.LambdaAWSInfo, error) {

	if len(scheduleExpression) <= 0 {
		return nil, errors.Errorf("EventBridge scheduleExpression must not be empty")
	}
	return newEventBridgeEventReactor(reactor,
		nil,
		scheduleExpression,
		additionalLambdaPermissions)
}
