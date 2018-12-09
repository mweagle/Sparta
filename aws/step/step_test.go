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
