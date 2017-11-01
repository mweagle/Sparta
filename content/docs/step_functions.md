---
date: 2017-10-31 18:20:05
title: Step Functions
weight: 10
menu:
  main:
    parent: Documentation
    identifier: step-functions
    weight: 100
---

# Introduction

AWS [Step Functions](https://aws.amazon.com/step-functions/) are a powerful way to express long-running, complex workflows comprised of Lambda functions. With Sparta 0.20.2, you can build a State Machine as part of your application. This section walks through the three steps necessary to provision a sample "Roll Die" state machine using a single Lambda function. See [SpartaStep](https://github.com/mweagle/SpartaStep) for the full source.

### Lambda Functions

The first step is to define the core Lambda function [Task](http://docs.aws.amazon.com/step-functions/latest/dg/amazon-states-language-task-state.html) that will be our Step function's core logic. In this example, we'll define a _rollDie_ function:

{{< highlight go >}}
// Standard AWS λ function
func lambdaRollDie(w http.ResponseWriter, r *http.Request) {
  ...
	// Return a randomized value in the range [0, 6]
	rollBytes, rollBytesErr := json.Marshal(&struct {
		Roll int `json:"roll"`
	}{
		Roll: rand.Int31n(1,7),
	})
	if rollBytesErr != nil {
		http.Error(w, rollBytesErr.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(rollBytes)
}
{{< /highlight >}}

### State Machine

Our state machine is simple: we want to keep rolling the die until we get a "good" result, with a delay in between rolls:

![State Machine](/images/step_functions/roll_die.jpg)

To do this, we use the new `github.com/mweagle/Sparta/aws/step` functions to define the other states.

{{< highlight go >}}
// Make all the Step states
lambdaTaskState := step.NewTaskState("lambdaRollDie", lambdaFn)
successState := step.NewSuccessState("success")
delayState := step.NewWaitDelayState("tryAgainShortly", 3*time.Second)
lambdaChoices := []step.ChoiceBranch{
  &step.Not{
    Comparison: &step.NumericGreaterThan{
      Variable: "$.roll",
      Value:    3,
    },
    Next: delayState,
  },
}
choiceState := step.NewChoiceState("checkRoll",
  lambdaChoices...).
  WithDefault(successState)
{{< /highlight >}}

The Sparta state types correspond to their [AWS States Spec](https://states-language.net) equivalents:

  - `successState` : [SucceedState](https://states-language.net/spec.html#succeed-state)
  - `delayState` : a specialized [WaitState](https://states-language.net/spec.html#wait-state)
  - `choiceState`: [ChoiceState](https://states-language.net/spec.html#choice-state)

See [godoc](https://godoc.org/github.com/mweagle/Sparta/aws/step) for the complete set of types.

The `lambdaTaskState` uses a normal Sparta function as in:

{{< highlight go >}}
lambdaFn := sparta.HandleAWSLambda("StepRollDie",
  http.HandlerFunc(lambdaRollDie),
  sparta.IAMRoleDefinition{})

lambdaFn.Options.MemorySize = 128
lambdaFn.Options.Tags = map[string]string{
  "myAccounting": "tag",
}
{{< /highlight >}}

The final step is to hook up the state transitions for states that don't implicitly include them, and create the State Machine:

{{< highlight go >}}
// Hook up the transitions
lambdaTaskState.Next(choiceState)
delayState.Next(lambdaTaskState)

// Startup the machine.
startMachine := step.NewStateMachine(lambdaTaskState)
{{< /highlight >}}

At this point we have a potentially well-formed [Lambda-powered](http://docs.aws.amazon.com/step-functions/latest/dg/tutorial-creating-lambda-state-machine.html) State Machine.
The final step is to attach this machine to the normal service definition.

### Service Decorator

The return type from `step.NewStateMachine(...)` is a `*step.StateMachine` instance that exposes a [ServiceDecoratorHook](https://godoc.org/github.com/mweagle/Sparta#ServiceDecoratorHook).
Adding the hook to your service's Workflow Hooks (similar to provisioning a service-scoped [CloudWatch Dashboard](https://github.com/mweagle/Sparta/blob/master/CHANGES.md#v0130))
will include it in the CloudFormation template serialization:

{{< highlight go >}}
// Setup the hook to annotate
workflowHooks := &sparta.WorkflowHooks{
  ServiceDecorator: startMachine.StateMachineDecorator(),
}
userStackName := spartaCF.UserScopedStackName("SpartaStep")
err := sparta.MainEx(userStackName,
  "Simple Sparta application that demonstrates AWS Step functions",
  lambdaFunctions,
  nil,
  nil,
  workflowHooks,
  false)
{{< /highlight >}}

With the decorator attached, the next service `provision` request will include the state machine as above.

### Testing

## Wrapping Up

AWS Step Functions are a powerful tool that allows you to orchestrate long running workflows using AWS Lambda.

# Notes
  * Minimal State machine validation is done at this time. See [Tim Bray](https://www.tbray.org/ongoing/When/201x/2016/12/01/J2119-Validator) for more information.
  * Value interrogation is defined by [JSONPath](http://goessner.net/articles/JsonPath/) expressions
