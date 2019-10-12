package step

import (
	"context"
	"math/rand"
	"testing"
	"time"

	sparta "github.com/mweagle/Sparta"
	spartaCF "github.com/mweagle/Sparta/aws/cloudformation"
	spartaTesting "github.com/mweagle/Sparta/testing"
	gocf "github.com/mweagle/go-cloudformation"
	"github.com/sirupsen/logrus"
)

func testStepProvision(t *testing.T,
	lambdaFns []*sparta.LambdaAWSInfo,
	stateMachine *StateMachine) {

	// Add the state machine to the deployment...
	workflowHooks := &sparta.WorkflowHooks{
		ServiceDecorators: []sparta.ServiceDecoratorHookHandler{
			stateMachine.StateMachineDecorator(),
		},
	}
	spartaTesting.ProvisionEx(t, lambdaFns, nil, nil, workflowHooks, false, nil)
}

// Standard AWS λ function
func helloWorld(ctx context.Context,
	props map[string]interface{}) (map[string]interface{}, error) {
	sparta.Logger().WithFields(logrus.Fields{
		"Woot": "Found",
	}).Warn("Lambda called")

	return map[string]interface{}{
		"hello": "world",
	}, nil
}

type lambdaRollResponse struct {
	Roll int `json:"roll"`
}

// Standard AWS λ function
func lambdaRollDie(ctx context.Context) (lambdaRollResponse, error) {
	return lambdaRollResponse{
		Roll: rand.Intn(5) + 1,
	}, nil
}

func TestAWSStepFunction(t *testing.T) {
	// Normal Sparta lambda function
	lambdaFn, _ := sparta.NewAWSLambda(sparta.LambdaName(helloWorld),
		helloWorld,
		sparta.IAMRoleDefinition{})

	// // Create a Choice state
	lambdaTaskState := NewLambdaTaskState("lambdaHelloWorld", lambdaFn)
	delayState := NewWaitDelayState("holdUpNow", 3*time.Second)
	successState := NewSuccessState("success")

	// Hook them up..
	lambdaTaskState.Next(delayState)
	delayState.Next(successState)

	// Startup the machine.
	startMachine := NewStateMachine("SampleStepFunction", lambdaTaskState)

	testStepProvision(t,
		[]*sparta.LambdaAWSInfo{lambdaFn},
		startMachine)
}

func TestRollDieChoice(t *testing.T) {
	lambdaFn, _ := sparta.NewAWSLambda("StepRollDie",
		lambdaRollDie,
		sparta.IAMRoleDefinition{})

	// Make all the Step states
	lambdaTaskState := NewLambdaTaskState("lambdaRollDie", lambdaFn)
	successState := NewSuccessState("success")
	delayState := NewWaitDelayState("tryAgainShortly", 3*time.Second)
	lambdaChoices := []ChoiceBranch{
		&Not{
			Comparison: &NumericGreaterThan{
				Variable: "$.roll",
				Value:    3,
			},
			Next: delayState,
		},
	}
	choiceState := NewChoiceState("checkRoll",
		lambdaChoices...).
		WithDefault(successState)

	// Hook up the transitions
	lambdaTaskState.Next(choiceState)
	delayState.Next(lambdaTaskState)

	// Startup the machine.
	stateMachineName := spartaCF.UserScopedStackName("TestStateMachine")
	startMachine := NewStateMachine(stateMachineName, lambdaTaskState)

	testStepProvision(t,
		[]*sparta.LambdaAWSInfo{lambdaFn},
		startMachine)
}

func TestDynamoDB(t *testing.T) {
	dynamoDbParams := DynamoDBGetItemParameters{
		TableName:       gocf.String("MY_TABLE"),
		AttributesToGet: []string{"attr1", "attr2"},
	}
	dynamoState := NewDynamoDBGetItemState("testState", dynamoDbParams)
	stateJSON, stateJSONErr := dynamoState.MarshalJSON()
	if stateJSONErr != nil {
		t.Fatalf("Failed to create JSON: %s", stateJSONErr)
	}
	t.Logf("JSON DATA:\n%s", string(stateJSON))
}

func createDataLambda(ctx context.Context,
	props map[string]interface{}) (map[string]interface{}, error) {

	return map[string]interface{}{
		"ship-date": "2016-03-14T01:59:00Z",
		"detail": map[string]interface{}{
			"delivery-partner": "UQS",
			"shipped": []map[string]interface{}{
				{
					"prod":      "R31",
					"dest-code": 9511,
					"quantity":  1344,
				},
				{
					"prod":      "S39",
					"dest-code": 9511,
					"quantity":  40,
				},
				{
					"prod":      "R31",
					"dest-code": 9833,
					"quantity":  12,
				},
				{
					"prod":      "R40",
					"dest-code": 9860,
					"quantity":  887,
				},
				{
					"prod":      "R40",
					"dest-code": 9511,
					"quantity":  1220,
				},
			},
		},
	}, nil
}

// Standard AWS λ function
func applyCallback(ctx context.Context,
	props map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{
		"Message": "Hello",
		"Event":   props,
	}, nil
}

func TestMapState(t *testing.T) {
	// Make the Map state
	lambdaMapFn, _ := sparta.NewAWSLambda("mapLambdaCallback",
		applyCallback,
		sparta.IAMRoleDefinition{})
	lambdaMapTaskState := NewLambdaTaskState("lambdaMapData", lambdaMapFn)
	mapMachine := NewStateMachine("mapStateName", lambdaMapTaskState)
	mapState := NewMapState("mapResults", mapMachine)
	successState := NewSuccessState("success")
	mapState.Next(successState)

	// Then the start state to produce some data
	lambdaProducerFn, _ := sparta.NewAWSLambda("produceData",
		createDataLambda,
		sparta.IAMRoleDefinition{})
	lambdaProducerTaskState := NewLambdaTaskState("lambdaProduceData", lambdaProducerFn)

	// Hook up the transitions
	stateMachineName := spartaCF.UserScopedStackName("TestMapStateMachine")
	lambdaProducerTaskState.Next(mapState)
	stateMachine := NewStateMachine(stateMachineName, lambdaProducerTaskState)
	// Startup the machine.
	testStepProvision(t,
		[]*sparta.LambdaAWSInfo{lambdaMapFn, lambdaProducerFn},
		stateMachine)
}

func TestParallelState(t *testing.T) {
	// Make the Map state
	lambdaMapFn, _ := sparta.NewAWSLambda("parallelLambdaCallback",
		applyCallback,
		sparta.IAMRoleDefinition{})
	lambdaMapTaskState := NewLambdaTaskState("lambdaMapData", lambdaMapFn)
	parallelMachine := NewStateMachine("mapStateName", lambdaMapTaskState)
	parallelState := NewParallelState("mapResults", parallelMachine)
	successState := NewSuccessState("success")
	parallelState.Next(successState)

	// Then the start state to produce some data
	lambdaProducerFn, _ := sparta.NewAWSLambda("produceData",
		createDataLambda,
		sparta.IAMRoleDefinition{})
	lambdaProducerTaskState := NewLambdaTaskState("lambdaProduceData", lambdaProducerFn)

	// Hook up the transitions
	stateMachineName := spartaCF.UserScopedStackName("TestParallelStateMachine")
	lambdaProducerTaskState.Next(parallelState)
	stateMachine := NewStateMachine(stateMachineName, lambdaProducerTaskState)
	// Startup the machine.
	testStepProvision(t,
		[]*sparta.LambdaAWSInfo{lambdaMapFn, lambdaProducerFn},
		stateMachine)
}
