package cloudtest

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"testing"
	"time"

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	awsv2Config "github.com/aws/aws-sdk-go-v2/config"
	awsv2CF "github.com/aws/aws-sdk-go-v2/service/cloudformation"
	awsv2CFTypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	awsv2Lambda "github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/jmespath/go-jmespath"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

////////////////////////////////////////////////////////////////////////////////

var cache *functionCache

func init() {
	// Init the function cache
	cache = &functionCache{
		perStackFunctions: make(map[string][]*awsv2Lambda.GetFunctionOutput),
		freeFunctions:     make(map[string]*awsv2Lambda.GetFunctionOutput),
	}
}

////////////////////////////////////////////////////////////////////////////////

type functionCache struct {
	perStackFunctions map[string][]*awsv2Lambda.GetFunctionOutput
	freeFunctions     map[string]*awsv2Lambda.GetFunctionOutput
	mu                sync.RWMutex
}

func (fc *functionCache) getFunction(t CloudTest, functionName string) *awsv2Lambda.GetFunctionOutput {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	output, outputExists := fc.freeFunctions[functionName]
	if !outputExists {
		// Look it up...
		t.Logf("Looking up function: %s", functionName)
		lambdaSvc := awsv2Lambda.NewFromConfig(t.Config())
		getFunctionInput := &awsv2Lambda.GetFunctionInput{
			FunctionName: awsv2.String(functionName),
		}
		lookupResult, lookupResultErr := lambdaSvc.GetFunction(context.Background(), getFunctionInput)
		if lookupResultErr != nil {
			lookupResult = &awsv2Lambda.GetFunctionOutput{}
		}
		fc.freeFunctions[functionName] = lookupResult
		output = lookupResult
	}

	// Check for an empty output placeholder and if this is a non-empty zero
	// struct then return it.
	if output != nil &&
		output.Configuration != nil &&
		len(*output.Configuration.FunctionName) != 0 {
		return output
	}
	return nil
}

func (fc *functionCache) getStackFunction(t CloudTest,
	stackName string,
	jmesSelector string) *awsv2Lambda.GetFunctionOutput {

	fc.mu.Lock()
	defer fc.mu.Unlock()

	// Do we have something there?
	stackFuncs, stackFuncExists := fc.perStackFunctions[stackName]
	if !stackFuncExists {
		t.Logf("Looking up functions for Stack: %s", stackName)

		// Store an empty map..., we'll overwrite this later if things go well...
		fc.perStackFunctions[stackName] = []*awsv2Lambda.GetFunctionOutput{}

		// Load them all
		// Get all the stack resources, then for each LambdaFunction
		// get the GetFunctionOutput information. For each one, apply the
		// jmesSelector and if it returns an ARN, we're done.
		cloudFormationSvc := awsv2CF.NewFromConfig(t.Config())
		params := &awsv2CF.ListStackResourcesInput{
			StackName: awsv2.String(stackName),
		}
		allLambdaFunctionSummaries := []*awsv2CFTypes.StackResourceSummary{}

		_, listErr := cloudFormationSvc.ListStackResources(context.Background(), params)

		// TODO - LIST THE RESOURCES
		// func(page *cloudformation.ListStackResourcesOutput, lastPage bool) bool {
		// 	for _, eachSummary := range page.StackResourceSummaries {
		// 		if *eachSummary.ResourceType == "AWS::Lambda::Function" {
		// 			allLambdaFunctionSummaries = append(allLambdaFunctionSummaries, eachSummary)
		// 		}
		// 	}
		// 	return true
		// })
		if listErr != nil {
			return nil
		}
		// Great, now for each one, let's get the function info
		lambdaSvc := awsv2Lambda.NewFromConfig(t.Config())
		functionOutput := []*awsv2Lambda.GetFunctionOutput{}
		for _, eachSummary := range allLambdaFunctionSummaries {
			getFunctionInput := &awsv2Lambda.GetFunctionInput{
				FunctionName: eachSummary.PhysicalResourceId,
			}
			getFunctionOutput, getFunctionOutputErr := lambdaSvc.GetFunction(context.Background(), getFunctionInput)
			if getFunctionOutputErr != nil {
				t.Errorf("Failed to get Function info for: %s", *eachSummary.PhysicalResourceId)
				return nil
			}
			functionOutput = append(functionOutput, getFunctionOutput)
		}
		fc.perStackFunctions[stackName] = functionOutput
		stackFuncs = functionOutput
	} else {
		t.Logf("Using cache for stack functions: %s", stackName)
	}

	// Ok, now for each one turn it into JSON, parse it, apply
	// the JMESPath, and if it returns with an ARN, use it...
	for _, eachFunctionOutput := range stackFuncs {
		isMatch, matchErr := IsJMESMatch(jmesSelector, eachFunctionOutput)
		if matchErr != nil {
			t.Errorf("JMES selection error: %s\n", matchErr.Error())
		} else if isMatch {
			return eachFunctionOutput
		}
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////

// IsJMESMatch is a function that accepts a GetFunctionOutput and a
// JMES selector expression. The return value is (isMatch, error). It's exposed
// as a package public to allow for callers to write their own providers
func IsJMESMatch(jmesSelector string, output *awsv2Lambda.GetFunctionOutput) (bool, error) {
	jsonData, jsonDataErr := json.MarshalIndent(output, "", "  ")
	if jsonDataErr != nil {
		return false, nil
	}
	// Unmarshalled data
	var unmarshalledData interface{}
	unmarshalErr := json.Unmarshal(jsonData, &unmarshalledData)
	if unmarshalErr != nil {
		return false, nil
	}
	matchResult, matchResultErr := jmespath.Search(jmesSelector, unmarshalledData)
	if matchResultErr != nil {
		return false, matchResultErr
	} else if matchResult != nil {
		_, isOk := matchResult.(string)
		if isOk {
			return isOk, nil
		}
	}
	return false, nil
}

// CloudTest is the interface passed to testing instances
type CloudTest interface {
	testing.TB
	Config() awsv2.Config
	ZeroLog() *zerolog.Logger
	Context() context.Context
}

type cloudTest struct {
	*testing.T
	awsConfig awsv2.Config
	logger    *zerolog.Logger
	ctx       context.Context
}

func (ct *cloudTest) Config() awsv2.Config {
	return ct.awsConfig
}

func (ct *cloudTest) ZeroLog() *zerolog.Logger {
	return ct.logger
}
func (ct *cloudTest) Context() context.Context {
	return ct.ctx
}

// Trigger is the interface that represents a struct that can trigger
// an event
type Trigger interface {
	Send(CloudTest, *awsv2Lambda.GetFunctionOutput) (interface{}, error)
	Cleanup(CloudTest, *awsv2Lambda.GetFunctionOutput)
}

// LambdaSelector is the interface that provides the awsv2Lambda.GetFunctionOutput
// that will be used for the test
type LambdaSelector interface {
	Select(CloudTest) (*awsv2Lambda.GetFunctionOutput, error)
}

// CloudEvaluator is the interface used to represent a predicate applied
// to a function output
type CloudEvaluator interface {
	Evaluate(CloudTest, *awsv2Lambda.GetFunctionOutput) error
}

// Test is the initial type used to build up a Lambda integration test
type Test struct {
}

// NewTest returns a Test pointer
func NewTest() *Test {
	return &Test{}
}

// Given associates the provided Mutator with the Test
func (ct *Test) Given(trigger Trigger) *TestTrigger {
	return &TestTrigger{
		trigger: trigger,
	}
}

// TestTrigger is the intermediate type that stores the Trigger reference
type TestTrigger struct {
	trigger Trigger
}

// Against accepts the selector against the provided Trigger and returns a
// TestEvaluator
func (ct *TestTrigger) Against(selector LambdaSelector) *TestEvaluator {
	return &TestEvaluator{
		trigger:  ct.trigger,
		selector: selector,
	}
}

// TestEvaluator is the type that stores the Trigger and LambdaSelector
type TestEvaluator struct {
	trigger  Trigger
	selector LambdaSelector
}

// Ensure applies the predicates to the results of the Lambda function and returns
// a TestScenario instance
func (cte *TestEvaluator) Ensure(evaluators ...CloudEvaluator) *TestScenario {
	scenario := &TestScenario{
		trigger:    cte.trigger,
		selector:   cte.selector,
		evaluators: evaluators,
	}
	return scenario
}

// TestScenario is the final type that encapsulates all the state associated
// with a given test
type TestScenario struct {
	trigger    Trigger
	selector   LambdaSelector
	evaluators []CloudEvaluator
	timeout    time.Duration
}

// Within sets the timeout for the test
func (cts *TestScenario) Within(duration time.Duration) {
	cts.timeout = duration
}

// Run actually runs the test in question
func (cts *TestScenario) Run(t *testing.T) {

	// We have a selector, so next up is to provide that to the
	// ensurers in case they need to exit early, we'll have a buffered
	// channel with the expected result
	deadline, _ := t.Deadline()
	errContext, cancelFunc := context.WithDeadline(context.Background(), deadline)
	defer cancelFunc()

	awsConfig, awsConfigErr := awsv2Config.LoadDefaultConfig(context.Background())
	if awsConfigErr != nil {
		t.Fatalf("Failed to create new AWS Session. Error: %s", awsConfigErr.Error())
	}
	zerologger := zerolog.New(os.Stdout).
		With().
		Timestamp().
		Logger().
		Level(zerolog.ErrorLevel)
	ct := &cloudTest{
		t,
		awsConfig,
		&zerologger,
		errContext,
	}

	// Go find the function, using the cached
	// information
	functionOutput, functionOutputErr := cts.selector.Select(ct)
	if functionOutputErr != nil {
		ct.Fatalf("Error finding target function: %v", functionOutputErr)
	}
	if functionOutput == nil {
		ct.Fatalf("No valid function found for selector")
	}

	errGroup, _ := errgroup.WithContext(errContext)

	for _, eachEvaluator := range cts.evaluators {
		curEvaluator := eachEvaluator
		errGroup.Go(
			func() error {
				return curEvaluator.Evaluate(ct, functionOutput)
			})
	}

	// Finally, call the trigger, so that we can trigger everything...
	t.Logf("Calling trigger (ts: %s)", time.Now().Format(time.RFC3339Nano))
	_, mutateErr := cts.trigger.Send(ct, functionOutput)
	if mutateErr != nil {
		cancelFunc()
		ct.Fatalf("Failed to invoke trigger: %#v", mutateErr.Error())
	}
	// Great, wait until all the evaluators are done...
	waitErr := errGroup.Wait()
	if waitErr != nil {
		ct.Fatalf("Failed to ensure result: %#v", waitErr.Error())
	}
	// Cleanup the trigger
	cts.trigger.Cleanup(ct, functionOutput)
}
