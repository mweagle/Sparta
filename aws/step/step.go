package step

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	sparta "github.com/mweagle/Sparta"
	spartaCF "github.com/mweagle/Sparta/aws/cloudformation"
	spartaIAM "github.com/mweagle/Sparta/aws/iam"
	gocf "github.com/mweagle/go-cloudformation"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Types of state machines per
// https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-stepfunctions-statemachine.html
const (
	stateMachineStandard = "STANDARD"
	stateMachineExpress  = "EXPRESS"
)

// StateError is the reserved type used for AWS Step function error names
// Ref: https://states-language.net/spec.html#appendix-a
type StateError string

const (
	// StatesAll is a wild-card which matches any Error Name.
	StatesAll StateError = "States.ALL"
	// StatesTimeout is a Task State either ran longer than the
	// “TimeoutSeconds” value, or failed to heartbeat for a time
	// longer than the “HeartbeatSeconds” value.
	StatesTimeout StateError = "States.Timeout"
	// StatesTaskFailed is a Task State failed during the execution
	StatesTaskFailed StateError = "States.TaskFailed"
	// StatesPermissions is a Task State failed because it had
	// insufficient privileges to execute the specified code.
	StatesPermissions StateError = "States.Permissions"
	// StatesResultPathMatchFailure is a Task State’s “ResultPath” field
	// cannot be applied to the input the state received
	StatesResultPathMatchFailure StateError = "States.ResultPathMatchFailure"
	// StatesBranchFailed is a branch of a Parallel state failed
	StatesBranchFailed StateError = "States.BranchFailed"
	// StatesNoChoiceMatched is a Choice state failed to find a match for the
	// condition field extracted from its input
	StatesNoChoiceMatched StateError = "States.NoChoiceMatched"
)

// MachineState is the base state for all AWS Step function
type MachineState interface {
	Name() string
	nodeID() string
	enableEndState(bool)
}

// TransitionState is the generic state according to
// https://states-language.net/spec.html#state-type-table
type TransitionState interface {
	MachineState
	Next(state MachineState) MachineState
	// AdjacentStates returns all the MachineStates that are reachable from
	// the current state
	AdjacentStates() []MachineState
	WithComment(string) TransitionState
	WithInputPath(string) TransitionState
	WithOutputPath(string) TransitionState
}

// Embedding struct for common properties
type baseInnerState struct {
	name              string
	id                int64
	next              MachineState
	comment           string
	inputPath         string
	outputPath        string
	isEndStateInvalid bool
}

func (bis *baseInnerState) nodeID() string {
	return fmt.Sprintf("%s-%d", bis.name, bis.id)
}

func (bis *baseInnerState) enableEndState(isEnabled bool) {
	bis.isEndStateInvalid = !isEnabled
}

// marshalStateJSON for subclass marshalling of state information
func (bis *baseInnerState) marshalStateJSON(stateType string,
	additionalData map[string]interface{}) ([]byte, error) {
	if additionalData == nil {
		additionalData = make(map[string]interface{})
	}
	additionalData["Type"] = stateType
	if bis.next != nil {
		additionalData["Next"] = bis.next.Name()
	}
	if bis.comment != "" {
		additionalData["Comment"] = bis.comment
	}
	if bis.inputPath != "" {
		additionalData["InputPath"] = bis.inputPath
	}
	if bis.outputPath != "" {
		additionalData["OutputPath"] = bis.outputPath
	}
	if !bis.isEndStateInvalid && bis.next == nil {
		additionalData["End"] = true
	}
	// Output the pretty version
	return json.Marshal(additionalData)
}

/*******************************************************************************
 ___ _____ _ _____ ___ ___
/ __|_   _/_\_   _| __/ __|
\__ \ | |/ _ \| | | _|\__ \
|___/ |_/_/ \_\_| |___|___/
/******************************************************************************/

////////////////////////////////////////////////////////////////////////////////
// PassState
////////////////////////////////////////////////////////////////////////////////

// PassState represents a NOP state
type PassState struct {
	baseInnerState
	ResultPath string
	Result     interface{}
}

// WithResultPath is the fluent builder for the result path
func (ps *PassState) WithResultPath(resultPath string) *PassState {
	ps.ResultPath = resultPath
	return ps
}

// WithResult is the fluent builder for the result data
func (ps *PassState) WithResult(result interface{}) *PassState {
	ps.Result = result
	return ps
}

// Next returns the next state
func (ps *PassState) Next(nextState MachineState) MachineState {
	ps.next = nextState
	return ps
}

// AdjacentStates returns nodes reachable from this node
func (ps *PassState) AdjacentStates() []MachineState {
	if ps.next == nil {
		return nil
	}
	return []MachineState{ps.next}
}

// Name returns the name of this Task state
func (ps *PassState) Name() string {
	return ps.name
}

// WithComment returns the TaskState comment
func (ps *PassState) WithComment(comment string) TransitionState {
	ps.comment = comment
	return ps
}

// WithInputPath returns the TaskState input data selector
func (ps *PassState) WithInputPath(inputPath string) TransitionState {
	ps.inputPath = inputPath
	return ps
}

// WithOutputPath returns the TaskState output data selector
func (ps *PassState) WithOutputPath(outputPath string) TransitionState {
	ps.outputPath = outputPath
	return ps
}

// MarshalJSON for custom marshalling
func (ps *PassState) MarshalJSON() ([]byte, error) {
	additionalParams := make(map[string]interface{})
	if ps.ResultPath != "" {
		additionalParams["ResultPath"] = ps.ResultPath
	}
	if ps.Result != nil {
		additionalParams["Result"] = ps.Result
	}
	return ps.marshalStateJSON("Pass", additionalParams)
}

// NewPassState returns a new PassState instance
func NewPassState(name string, resultData interface{}) *PassState {
	return &PassState{
		baseInnerState: baseInnerState{
			name: name,
			id:   rand.Int63(),
		},
		Result: resultData,
	}
}

////////////////////////////////////////////////////////////////////////////////
// ChoiceState
////////////////////////////////////////////////////////////////////////////////

// ChoiceState is a synthetic state that executes a lot of independent
// branches in parallel
type ChoiceState struct {
	baseInnerState
	Choices []ChoiceBranch
	Default TransitionState
}

// WithDefault is the fluent builder for the default state
func (cs *ChoiceState) WithDefault(defaultState TransitionState) *ChoiceState {
	cs.Default = defaultState
	return cs
}

// WithResultPath is the fluent builder for the result path
func (cs *ChoiceState) WithResultPath(resultPath string) *ChoiceState {
	return cs
}

// Name returns the name of this Task state
func (cs *ChoiceState) Name() string {
	return cs.name
}

// WithComment returns the TaskState comment
func (cs *ChoiceState) WithComment(comment string) *ChoiceState {
	cs.comment = comment
	return cs
}

// MarshalJSON for custom marshalling
func (cs *ChoiceState) MarshalJSON() ([]byte, error) {
	/*
		A state in a Parallel state branch “States” field MUST NOT have a “Next” field that targets a field outside of that “States” field. A state MUST NOT have a “Next” field which matches a state name inside a Parallel state branch’s “States” field unless it is also inside the same “States” field.

		Put another way, states in a branch’s “States” field can transition only to each other, and no state outside of that “States” field can transition into it.
	*/
	additionalParams := make(map[string]interface{})
	additionalParams["Choices"] = cs.Choices
	if cs.Default != nil {
		additionalParams["Default"] = cs.Default.Name()
	}
	return cs.marshalStateJSON("Choice", additionalParams)
}

// NewChoiceState returns a "ChoiceState" with the supplied
// information
func NewChoiceState(choiceStateName string, choices ...ChoiceBranch) *ChoiceState {
	return &ChoiceState{
		baseInnerState: baseInnerState{
			name:              choiceStateName,
			id:                rand.Int63(),
			isEndStateInvalid: true,
		},
		Choices: append([]ChoiceBranch{}, choices...),
	}
}

////////////////////////////////////////////////////////////////////////////////
// TaskRetry
////////////////////////////////////////////////////////////////////////////////

// TaskRetry is an action to perform in response to a Task failure
type TaskRetry struct {
	ErrorEquals []StateError `json:",omitempty"`
	//lint:ignore ST1011 we want to give a cue to the client of the units
	IntervalSeconds time.Duration `json:",omitempty"`
	MaxAttempts     int           `json:",omitempty"`
	BackoffRate     float32       `json:",omitempty"`
}

// WithErrors is the fluent builder
func (tr *TaskRetry) WithErrors(errors ...StateError) *TaskRetry {
	if tr.ErrorEquals == nil {
		tr.ErrorEquals = make([]StateError, 0)
	}
	tr.ErrorEquals = append(tr.ErrorEquals, errors...)
	return tr
}

// WithInterval is the fluent builder
func (tr *TaskRetry) WithInterval(interval time.Duration) *TaskRetry {
	tr.IntervalSeconds = interval
	return tr
}

// WithMaxAttempts is the fluent builder
func (tr *TaskRetry) WithMaxAttempts(maxAttempts int) *TaskRetry {
	tr.MaxAttempts = maxAttempts
	return tr
}

// WithBackoffRate is the fluent builder
func (tr *TaskRetry) WithBackoffRate(backoffRate float32) *TaskRetry {
	tr.BackoffRate = backoffRate
	return tr
}

// NewTaskRetry returns a new TaskRetry instance
func NewTaskRetry() *TaskRetry {
	return &TaskRetry{}
}

////////////////////////////////////////////////////////////////////////////////
// TaskCatch
////////////////////////////////////////////////////////////////////////////////

// TaskCatch is an action to handle a failing operation
type TaskCatch struct {
	/*
		The reserved name “States.ALL” appearing in a Retrier’s “ErrorEquals” field is a wild-card and matches any Error Name. Such a value MUST appear alone in the “ErrorEquals” array and MUST appear in the last Catcher in the “Catch” array.
	*/
	errorEquals []StateError
	next        TransitionState
}

// MarshalJSON to prevent inadvertent composition
func (tc *TaskCatch) MarshalJSON() ([]byte, error) {
	catchJSON := map[string]interface{}{
		"ErrorEquals": tc.errorEquals,
		"Next":        tc.next,
	}
	return json.Marshal(catchJSON)
}

// NewTaskCatch returns a new TaskCatch instance
func NewTaskCatch(nextState TransitionState, errors ...StateError) *TaskCatch {
	return &TaskCatch{
		errorEquals: errors,
		next:        nextState,
	}
}

////////////////////////////////////////////////////////////////////////////////
// BaseTask
////////////////////////////////////////////////////////////////////////////////

// BaseTask represents the core BaseTask control flow options.
type BaseTask struct {
	baseInnerState
	ResultPath string
	//lint:ignore ST1011 we want to give a cue to the client of the units
	TimeoutSeconds time.Duration
	//lint:ignore ST1011 we want to give a cue to the client of the units
	HeartbeatSeconds time.Duration
	LambdaDecorator  sparta.TemplateDecorator
	Retriers         []*TaskRetry
	Catchers         []*TaskCatch
}

func (bt *BaseTask) marshalMergedParams(taskResourceType string,
	taskParams interface{}) ([]byte, error) {
	jsonBytes, jsonBytesErr := json.Marshal(taskParams)
	if jsonBytesErr != nil {
		return nil, errors.Wrapf(jsonBytesErr, "attempting to JSON marshal task params")
	}

	var unmarshaled interface{}
	unmarshalErr := json.Unmarshal(jsonBytes, &unmarshaled)
	if unmarshalErr != nil {
		return nil, errors.Wrapf(unmarshalErr, "attempting to unmarshall params")
	}

	mapTyped, mapTypedErr := unmarshaled.(map[string]interface{})
	if !mapTypedErr {
		return nil, errors.Errorf("attempting to type convert unmarshalled params to map[string]interface{}")
	}
	additionalParams := bt.additionalParams()
	additionalParams["Resource"] = taskResourceType
	additionalParams["Parameters"] = mapTyped
	return bt.marshalStateJSON("Task", additionalParams)
}

// attributeMap returns the map of attributes necessary
// for JSON serialization
func (bt *BaseTask) additionalParams() map[string]interface{} {
	additionalParams := make(map[string]interface{})

	if bt.TimeoutSeconds.Seconds() != 0 {
		additionalParams["TimeoutSeconds"] = bt.TimeoutSeconds.Seconds()
	}
	if bt.HeartbeatSeconds.Seconds() != 0 {
		additionalParams["HeartbeatSeconds"] = bt.HeartbeatSeconds.Seconds()
	}
	if bt.ResultPath != "" {
		additionalParams["ResultPath"] = bt.ResultPath
	}
	if len(bt.Retriers) != 0 {
		additionalParams["Retry"] = make([]map[string]interface{}, 0)
	}
	if bt.Catchers != nil {
		catcherMap := make([]map[string]interface{}, len(bt.Catchers))
		for index, eachCatcher := range bt.Catchers {
			catcherMap[index] = map[string]interface{}{
				"ErrorEquals": eachCatcher.errorEquals,
				"Next":        eachCatcher.next.Name(),
			}
		}
		additionalParams["Catch"] = catcherMap
	}
	return additionalParams
}

// Next returns the next state
func (bt *BaseTask) Next(nextState MachineState) MachineState {
	bt.next = nextState
	return nextState
}

// AdjacentStates returns nodes reachable from this node
func (bt *BaseTask) AdjacentStates() []MachineState {
	adjacent := []MachineState{}
	if bt.next != nil {
		adjacent = append(adjacent, bt.next)
	}
	for _, eachCatcher := range bt.Catchers {
		adjacent = append(adjacent, eachCatcher.next)
	}
	return adjacent
}

// Name returns the name of this Task state
func (bt *BaseTask) Name() string {
	return bt.name
}

// WithResultPath is the fluent builder for the result path
func (bt *BaseTask) WithResultPath(resultPath string) *BaseTask {
	bt.ResultPath = resultPath
	return bt
}

// WithTimeout is the fluent builder for BaseTask
func (bt *BaseTask) WithTimeout(timeout time.Duration) *BaseTask {
	bt.TimeoutSeconds = timeout
	return bt
}

// WithHeartbeat is the fluent builder for BaseTask
func (bt *BaseTask) WithHeartbeat(pulse time.Duration) *BaseTask {
	bt.HeartbeatSeconds = pulse
	return bt
}

// WithRetriers is the fluent builder for BaseTask
func (bt *BaseTask) WithRetriers(retries ...*TaskRetry) *BaseTask {
	if bt.Retriers == nil {
		bt.Retriers = make([]*TaskRetry, 0)
	}
	bt.Retriers = append(bt.Retriers, retries...)
	return bt
}

// WithCatchers is the fluent builder for BaseTask
func (bt *BaseTask) WithCatchers(catch ...*TaskCatch) *BaseTask {
	if bt.Catchers == nil {
		bt.Catchers = make([]*TaskCatch, 0)
	}
	bt.Catchers = append(bt.Catchers, catch...)
	return bt
}

// WithComment returns the BaseTask comment
func (bt *BaseTask) WithComment(comment string) TransitionState {
	bt.comment = comment
	return bt
}

// WithInputPath returns the BaseTask input data selector
func (bt *BaseTask) WithInputPath(inputPath string) TransitionState {
	bt.inputPath = inputPath
	return bt
}

// WithOutputPath returns the BaseTask output data selector
func (bt *BaseTask) WithOutputPath(outputPath string) TransitionState {
	bt.outputPath = outputPath
	return bt
}

// MarshalJSON to prevent inadvertent composition
func (bt *BaseTask) MarshalJSON() ([]byte, error) {

	return nil, errors.Errorf("step.BaseTask doesn't support direct JSON serialization. Prefer using an embedding Task type (eg: TaskState, FargateTaskState)")
}

////////////////////////////////////////////////////////////////////////////////
// LambdaTaskState
////////////////////////////////////////////////////////////////////////////////

// LambdaTaskState is the core state, responsible for delegating to a Lambda function
type LambdaTaskState struct {
	BaseTask
	lambdaFn                  *sparta.LambdaAWSInfo
	lambdaLogicalResourceName string
	preexistingDecorator      sparta.TemplateDecorator
}

// NewLambdaTaskState returns a LambdaTaskState instance properly initialized
func NewLambdaTaskState(stateName string, lambdaFn *sparta.LambdaAWSInfo) *LambdaTaskState {
	ts := &LambdaTaskState{
		BaseTask: BaseTask{
			baseInnerState: baseInnerState{
				name: stateName,
				id:   rand.Int63(),
			},
		},
		lambdaFn: lambdaFn,
	}
	ts.LambdaDecorator = func(serviceName string,
		lambdaResourceName string,
		lambdaResource gocf.LambdaFunction,
		resourceMetadata map[string]interface{},
		S3Bucket string,
		S3Key string,
		buildID string,
		cfTemplate *gocf.Template,
		context map[string]interface{},
		logger *logrus.Logger) error {
		if ts.preexistingDecorator != nil {
			preexistingLambdaDecoratorErr := ts.preexistingDecorator(
				serviceName,
				lambdaResourceName,
				lambdaResource,
				resourceMetadata,
				S3Bucket,
				S3Key,
				buildID,
				cfTemplate,
				context,
				logger)
			if preexistingLambdaDecoratorErr != nil {
				return preexistingLambdaDecoratorErr
			}
		}
		// Save the lambda name s.t. we can create the {"Ref"::"lambdaName"} entry...
		ts.lambdaLogicalResourceName = lambdaResourceName
		return nil
	}
	// Make sure this Lambda decorator is included in the list of existing decorators

	// If there already is a decorator, then save it...
	ts.preexistingDecorator = lambdaFn.Decorator
	ts.lambdaFn.Decorators = append(ts.lambdaFn.Decorators,
		sparta.TemplateDecoratorHookFunc(ts.LambdaDecorator))
	return ts
}

// MarshalJSON for custom marshalling, since this will be stringified and we need it
// to turn into a stringified Ref:
func (ts *LambdaTaskState) MarshalJSON() ([]byte, error) {
	additionalParams := ts.BaseTask.additionalParams()
	additionalParams["Resource"] = gocf.GetAtt(ts.lambdaLogicalResourceName, "Arn")
	return ts.marshalStateJSON("Task", additionalParams)
}

////////////////////////////////////////////////////////////////////////////////
// WaitDelay
////////////////////////////////////////////////////////////////////////////////

// WaitDelay is a delay with an interval
type WaitDelay struct {
	baseInnerState
	delay time.Duration
}

// Name returns the WaitDelay name
func (wd *WaitDelay) Name() string {
	return wd.name
}

// Next sets the step after the wait delay
func (wd *WaitDelay) Next(nextState MachineState) MachineState {
	wd.next = nextState
	return wd
}

// AdjacentStates returns nodes reachable from this node
func (wd *WaitDelay) AdjacentStates() []MachineState {
	if wd.next == nil {
		return nil
	}
	return []MachineState{wd.next}
}

// WithComment returns the WaitDelay comment
func (wd *WaitDelay) WithComment(comment string) TransitionState {
	wd.comment = comment
	return wd
}

// WithInputPath returns the TaskState input data selector
func (wd *WaitDelay) WithInputPath(inputPath string) TransitionState {
	wd.inputPath = inputPath
	return wd
}

// WithOutputPath returns the TaskState output data selector
func (wd *WaitDelay) WithOutputPath(outputPath string) TransitionState {
	wd.outputPath = outputPath
	return wd
}

// MarshalJSON for custom marshalling
func (wd *WaitDelay) MarshalJSON() ([]byte, error) {
	additionalParams := make(map[string]interface{})
	additionalParams["Seconds"] = wd.delay.Seconds()
	return wd.marshalStateJSON("Wait", additionalParams)
}

// NewWaitDelayState returns a new WaitDelay pointer instance
func NewWaitDelayState(stateName string, delay time.Duration) *WaitDelay {
	return &WaitDelay{
		baseInnerState: baseInnerState{
			name: stateName,
			id:   rand.Int63(),
		},
		delay: delay,
	}
}

////////////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////////////
// WaitUntil
////////////////////////////////////////////////////////////////////////////////

// WaitUntil is a delay with an absolute time gate
type WaitUntil struct {
	baseInnerState
	Timestamp time.Time
}

// Name returns the WaitDelay name
func (wu *WaitUntil) Name() string {
	return wu.name
}

// Next sets the step after the wait delay
func (wu *WaitUntil) Next(nextState MachineState) MachineState {
	wu.next = nextState
	return wu
}

// AdjacentStates returns nodes reachable from this node
func (wu *WaitUntil) AdjacentStates() []MachineState {
	if wu.next == nil {
		return nil
	}
	return []MachineState{wu.next}
}

// WithComment returns the WaitDelay comment
func (wu *WaitUntil) WithComment(comment string) TransitionState {
	wu.comment = comment
	return wu
}

// WithInputPath returns the TaskState input data selector
func (wu *WaitUntil) WithInputPath(inputPath string) TransitionState {
	wu.inputPath = inputPath
	return wu
}

// WithOutputPath returns the TaskState output data selector
func (wu *WaitUntil) WithOutputPath(outputPath string) TransitionState {
	wu.outputPath = outputPath
	return wu
}

// MarshalJSON for custom marshalling
func (wu *WaitUntil) MarshalJSON() ([]byte, error) {
	additionalParams := make(map[string]interface{})
	additionalParams["Timestamp"] = wu.Timestamp.Format(time.RFC3339)
	return wu.marshalStateJSON("Wait", additionalParams)
}

// NewWaitUntilState returns a new WaitDelay pointer instance
func NewWaitUntilState(stateName string, waitUntil time.Time) *WaitUntil {
	return &WaitUntil{
		baseInnerState: baseInnerState{
			name: stateName,
			id:   rand.Int63(),
		},
		Timestamp: waitUntil,
	}
}

////////////////////////////////////////////////////////////////////////////////

// WaitDynamicUntil is a delay based on a previous response
type WaitDynamicUntil struct {
	baseInnerState
	TimestampPath string
}

// Name returns the WaitDelay name
func (wdu *WaitDynamicUntil) Name() string {
	return wdu.name
}

// Next sets the step after the wait delay
func (wdu *WaitDynamicUntil) Next(nextState MachineState) MachineState {
	wdu.next = nextState
	return wdu
}

// AdjacentStates returns nodes reachable from this node
func (wdu *WaitDynamicUntil) AdjacentStates() []MachineState {
	if wdu.next == nil {
		return nil
	}
	return []MachineState{wdu.next}
}

// WithComment returns the WaitDelay comment
func (wdu *WaitDynamicUntil) WithComment(comment string) TransitionState {
	wdu.comment = comment
	return wdu
}

// WithInputPath returns the TaskState input data selector
func (wdu *WaitDynamicUntil) WithInputPath(inputPath string) TransitionState {
	wdu.inputPath = inputPath
	return wdu
}

// WithOutputPath returns the TaskState output data selector
func (wdu *WaitDynamicUntil) WithOutputPath(outputPath string) TransitionState {
	wdu.outputPath = outputPath
	return wdu
}

// MarshalJSON for custom marshalling
func (wdu *WaitDynamicUntil) MarshalJSON() ([]byte, error) {
	additionalParams := make(map[string]interface{})
	additionalParams["TimestampPath"] = wdu.TimestampPath
	return wdu.marshalStateJSON("Wait", additionalParams)
}

// NewWaitDynamicUntilState returns a new WaitDynamicUntil pointer instance
func NewWaitDynamicUntilState(stateName string, timestampPath string) *WaitDynamicUntil {
	return &WaitDynamicUntil{
		baseInnerState: baseInnerState{
			name: stateName,
			id:   rand.Int63(),
		},
		TimestampPath: timestampPath,
	}
}

/*
Validate-All": {
  "Type": "Map",
  "InputPath": "$.detail",
  "ItemsPath": "$.shipped",
  "MaxConcurrency": 0,
  "Parameters": {
    "parcel.$": "$$.Map.Item.Value",
    "courier.$": "$.delivery-partner"
  },
  "Iterator": {
    "StartAt": "Validate",
    "States": {
      "Validate": {
        "Type": "Task",
        "Resource": "arn:aws:lambda:us-east-1:123456789012:function:ship-val",
        "End": true
      }
    }
  },
  "ResultPath": "$.detail.shipped",
  "End": true
*/

////////////////////////////////////////////////////////////////////////////////
// StateMachine
////////////////////////////////////////////////////////////////////////////////

// StateMachine is the top level item
type StateMachine struct {
	name                 string
	comment              string
	stateDefinitionError error
	machineType          string
	loggingConfiguration *gocf.StepFunctionsStateMachineLoggingConfiguration
	startAt              TransitionState
	uniqueStates         map[string]MachineState
	roleArn              gocf.Stringable
	// internal flag to suppress the automatic "End" property
	// from being serialized for Map states
	disableEndState bool
}

//Comment sets the StateMachine comment
func (sm *StateMachine) Comment(comment string) *StateMachine {
	sm.comment = comment
	return sm
}

//WithRoleArn sets the state machine roleArn
func (sm *StateMachine) WithRoleArn(roleArn gocf.Stringable) *StateMachine {
	sm.roleArn = roleArn
	return sm
}

// validate performs any validation against the state machine
// prior to marshaling
func (sm *StateMachine) validate() []error {
	validationErrors := make([]error, 0)
	if sm.stateDefinitionError != nil {
		validationErrors = append(validationErrors, sm.stateDefinitionError)
	}

	// TODO - add Catcher validator
	/*
		Each Catcher MUST contain a field named “ErrorEquals”, specified exactly as with the Retrier “ErrorEquals” field, and a field named “Next” whose value MUST be a string exactly matching a State Name.

		When a state reports an error and either there is no Retry field, or retries have failed to resolve the error, the interpreter scans through the Catchers in array order, and when the Error Name appears in the value of a Catcher’s “ErrorEquals” field, transitions the machine to the state named in the value of the “Next” field.

		The reserved name “States.ALL” appearing in a Retrier’s “ErrorEquals” field is a wild-card and matches any Error Name. Such a value MUST appear alone in the “ErrorEquals” array and MUST appear in the last Catcher in the “Catch” array.
	*/
	return validationErrors
}

// StateMachineDecorator is a decorator that returns a default
// CloudFormationResource named decorator
func (sm *StateMachine) StateMachineDecorator() sparta.ServiceDecoratorHookFunc {
	cfName := sparta.CloudFormationResourceName("StateMachine", "StateMachine")
	return sm.StateMachineNamedDecorator(cfName)
}

// StateMachineNamedDecorator is the hook exposed by the StateMachine
// to insert the AWS Step function into the CloudFormation template
func (sm *StateMachine) StateMachineNamedDecorator(stepFunctionResourceName string) sparta.ServiceDecoratorHookFunc {
	return func(context map[string]interface{},
		serviceName string,
		template *gocf.Template,
		S3Bucket string,
		S3Key string,
		buildID string,
		awsSession *session.Session,
		noop bool,
		logger *logrus.Logger) error {

		machineErrors := sm.validate()
		if len(machineErrors) != 0 {
			errorText := make([]string, len(machineErrors))
			for index := range machineErrors {
				errorText[index] = machineErrors[index].Error()
			}
			return errors.Errorf("Invalid state machine. Errors: %s",
				strings.Join(errorText, ", "))
		}

		lambdaFunctionResourceNames := []string{}
		for _, eachState := range sm.uniqueStates {
			switch taskState := eachState.(type) {
			case *LambdaTaskState:
				{
					lambdaFunctionResourceNames = append(lambdaFunctionResourceNames,
						taskState.lambdaLogicalResourceName)
				}
			case *MapState:
				{
					for _, eachUniqueState := range taskState.States.uniqueStates {
						switch typedMapState := eachUniqueState.(type) {
						case *LambdaTaskState:
							{
								lambdaFunctionResourceNames = append(lambdaFunctionResourceNames,
									typedMapState.lambdaLogicalResourceName)
							}
						}
					}
				}
			case *ParallelState:
				{
					for _, eachBranch := range taskState.Branches {
						for _, eachUniqueState := range eachBranch.uniqueStates {
							switch typedParallelState := eachUniqueState.(type) {
							case *LambdaTaskState:
								{
									lambdaFunctionResourceNames = append(lambdaFunctionResourceNames,
										typedParallelState.lambdaLogicalResourceName)
								}
							}
						}
					}
				}
			}
		}

		// Assume policy document
		regionalPrincipal := gocf.Join(".",
			gocf.String("states"),
			gocf.Ref("AWS::Region"),
			gocf.String("amazonaws.com"))
		var AssumePolicyDocument = sparta.ArbitraryJSONObject{
			"Version": "2012-10-17",
			"Statement": []sparta.ArbitraryJSONObject{
				{
					"Effect": "Allow",
					"Principal": sparta.ArbitraryJSONObject{
						"Service": regionalPrincipal,
					},
					"Action": []string{"sts:AssumeRole"},
				},
			},
		}
		var iamRoleResourceName string
		if len(lambdaFunctionResourceNames) != 0 {
			statesIAMRole := &gocf.IAMRole{
				AssumeRolePolicyDocument: AssumePolicyDocument,
			}
			statements := make([]spartaIAM.PolicyStatement, 0)
			for _, eachLambdaName := range lambdaFunctionResourceNames {
				statements = append(statements,
					spartaIAM.PolicyStatement{
						Effect: "Allow",
						Action: []string{
							"lambda:InvokeFunction",
						},
						Resource: gocf.GetAtt(eachLambdaName, "Arn").String(),
					},
				)
			}
			iamPolicies := gocf.IAMRolePolicyList{}
			iamPolicies = append(iamPolicies, gocf.IAMRolePolicy{
				PolicyDocument: sparta.ArbitraryJSONObject{
					"Version":   "2012-10-17",
					"Statement": statements,
				},
				PolicyName: gocf.String("StatesExecutionPolicy"),
			})
			statesIAMRole.Policies = &iamPolicies
			iamRoleResourceName = sparta.CloudFormationResourceName("StatesIAMRole",
				"StatesIAMRole")
			template.AddResource(iamRoleResourceName, statesIAMRole)
		}

		// Sweet - serialize it without indentation so that the
		// ConvertToTemplateExpression can actually parse the inline `Ref` objects
		jsonBytes, jsonBytesErr := json.Marshal(sm)
		if jsonBytesErr != nil {
			return errors.Errorf("Failed to marshal: %s", jsonBytesErr.Error())
		}
		logger.WithFields(logrus.Fields{
			"StateMachine": string(jsonBytes),
		}).Debug("State machine definition")

		// Super, now parse this into an Fn::Join representation
		// so that we can get inline expansion of the AWS pseudo params
		smReader := bytes.NewReader(jsonBytes)
		templateExpr, templateExprErr := spartaCF.ConvertToInlineJSONTemplateExpression(smReader, nil)
		if nil != templateExprErr {
			return errors.Errorf("Failed to parser: %s", templateExprErr.Error())
		}

		// Awsome - add an AWS::StepFunction to the template with this info and roll with it...
		stepFunctionResource := &gocf.StepFunctionsStateMachine{
			StateMachineName:     gocf.String(sm.name),
			DefinitionString:     templateExpr,
			LoggingConfiguration: sm.loggingConfiguration,
		}
		if iamRoleResourceName != "" {
			stepFunctionResource.RoleArn = gocf.GetAtt(iamRoleResourceName, "Arn").String()
		} else if sm.roleArn != nil {
			stepFunctionResource.RoleArn = sm.roleArn.String()
		}
		if sm.machineType != "" {
			stepFunctionResource.StateMachineType = gocf.String(sm.machineType)
		}
		template.AddResource(stepFunctionResourceName, stepFunctionResource)
		return nil
	}
}

// MarshalJSON for custom marshalling
func (sm *StateMachine) MarshalJSON() ([]byte, error) {

	// If there aren't any states, then it's the end
	return json.Marshal(&struct {
		Comment string                  `json:",omitempty"`
		StartAt string                  `json:",omitempty"`
		States  map[string]MachineState `json:",omitempty"`
		End     bool                    `json:",omitempty"`
	}{
		Comment: sm.comment,
		StartAt: sm.startAt.Name(),
		States:  sm.uniqueStates,
		End:     (len(sm.uniqueStates) == 1) && !sm.disableEndState,
	})
}

func createStateMachine(stateMachineName string,
	machineType string,
	startState TransitionState) *StateMachine {
	uniqueStates := make(map[string]MachineState)
	pendingStates := []MachineState{startState}
	duplicateStateNames := make(map[string]bool)

	nodeVisited := func(node MachineState) bool {
		if node == nil {
			return true
		}
		_, visited := uniqueStates[node.Name()]
		return visited
	}

	for len(pendingStates) != 0 {
		headState, tailStates := pendingStates[0], pendingStates[1:]
		uniqueStates[headState.Name()] = headState

		switch stateNode := headState.(type) {
		case *ChoiceState:
			for _, eachChoice := range stateNode.Choices {
				if !nodeVisited(eachChoice.nextState()) {
					tailStates = append(tailStates, eachChoice.nextState())
				}
			}
			if !nodeVisited(stateNode.Default) {
				tailStates = append(tailStates, stateNode.Default)
			}

		case TransitionState:
			for _, eachAdjacentState := range stateNode.AdjacentStates() {
				if !nodeVisited(eachAdjacentState) {
					tailStates = append(tailStates, eachAdjacentState)
				}
			}
			// Are there any Catchers in here?
		}
		pendingStates = tailStates
	}

	// Walk all the states and assemble them into the states slice
	sm := &StateMachine{
		name:         stateMachineName,
		startAt:      startState,
		uniqueStates: uniqueStates,
	}
	if machineType != "" {
		sm.machineType = machineType
	}
	// Store duplicate state names
	if len(duplicateStateNames) != 0 {
		sm.stateDefinitionError = fmt.Errorf("duplicate state names: %#v", duplicateStateNames)
	}
	return sm

}

// NewExpressStateMachine returns a new Express StateMachine instance. See
// https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-stepfunctions-statemachine.html
// for more information.
func NewExpressStateMachine(stateMachineName string,
	loggingConfiguration *gocf.StepFunctionsStateMachineLoggingConfiguration,
	startState TransitionState) *StateMachine {

	sm := createStateMachine(stateMachineName,
		stateMachineExpress,
		startState)
	sm.loggingConfiguration = loggingConfiguration
	return sm
}

// NewStateMachine returns a new StateMachine instance
func NewStateMachine(stateMachineName string,
	startState TransitionState) *StateMachine {

	return createStateMachine(stateMachineName,
		stateMachineStandard,
		startState)
}

////////////////////////////////////////////////////////////////////////////////
