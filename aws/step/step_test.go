package step

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	sparta "github.com/mweagle/Sparta"
	"github.com/sirupsen/logrus"
)

func TestAWSStepFunction(t *testing.T) {
	// Normal Sparta lambda function
	lambdaFn := sparta.HandleAWSLambda(sparta.LambdaName(helloWorld),
		helloWorld,
		sparta.IAMRoleDefinition{})

	// // Create a Choice state
	lambdaTaskState := NewTaskState("lambdaHelloWorld", lambdaFn)
	delayState := NewWaitDelayState("holdUpNow", 3*time.Second)
	successState := NewSuccessState("success")

	// Hook them up..
	lambdaTaskState.Next(delayState)
	delayState.Next(successState)

	// Startup the machine.
	startMachine := NewStateMachine("SampleStepFunction", lambdaTaskState)

	// Add the state machine to the deployment...
	workflowHooks := &sparta.WorkflowHooks{
		ServiceDecorator: startMachine.StateMachineDecorator(),
	}

	// Test it...
	logger, _ := sparta.NewLogger("info")
	var templateWriter bytes.Buffer
	err := sparta.Provision(true,
		"SampleStepFunction",
		"",
		[]*sparta.LambdaAWSInfo{lambdaFn},
		nil,
		nil,
		os.Getenv("S3_BUCKET"),
		false,
		false,
		"testBuildID",
		"",
		"",
		"",
		&templateWriter,
		workflowHooks,
		logger)
	if nil != err {
		t.Fatal(err.Error())
	}
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
