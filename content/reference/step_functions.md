---
date: 2017-10-31 18:20:05
title: Step Functions
weight: 10
---

# Introduction

AWS [Step Functions](https://aws.amazon.com/step-functions/) are a powerful way to express long-running, complex workflows comprised of Lambda functions. With Sparta 0.20.2, you can build a State Machine as part of your application. This section walks through the three steps necessary to provision a sample "Roll Die" state machine using a single Lambda function. See [SpartaStep](https://github.com/mweagle/SpartaStep) for the full source.

### Lambda Functions

The first step is to define the core Lambda function [Task](http://docs.aws.amazon.com/step-functions/latest/dg/amazon-states-language-task-state.html) that will be our Step function's core logic. In this example, we'll define a _rollDie_ function:

{{< highlight go >}}
// Standard AWS λ function
func lambdaRollDie(w http.ResponseWriter, r *http.Request) {
  ...
	// Return a randomized value in the range [1, 6]
	rollBytes, rollBytesErr := json.Marshal(&struct {
		Roll int `json:"roll"`
	}{
		Roll: rand.Intn(5) + 1,
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

The Sparta state types correspond to their [AWS States Spec](https://states-language.net/spec.html) equivalents:

  - `successState` : [SucceedState](https://states-language.net/spec.html#succeed-state)
  - `delayState` : a specialized [WaitState](https://states-language.net/spec.html#wait-state)
  - `choiceState`: [ChoiceState](https://states-language.net/spec.html#choice-state)

The `choiceState` is the most interesting state: based on the JSON response of the `lambdaRollDie`, it will either transition
to a delay or the success end state.

See [godoc](https://godoc.org/github.com/mweagle/Sparta/aws/step) for the complete set of types.

The `lambdaTaskState` uses a normal Sparta function as in:

{{< highlight go >}}
lambdaFn := sparta.HandleAWSLambda("StepRollDie",
  lambdaRollDie,
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

// Startup the machine with a user-scoped name for account uniqueness
stateMachineName := spartaCF.UserScopedStackName("StateMachine")
startMachine := step.NewStateMachine(stateMachineName, lambdaTaskState)
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

{{< highlight text >}}

INFO[0000] ══════════════════════════════════════════════════════════════
INFO[0000]    _______  ___   ___  _________
INFO[0000]   / __/ _ \/ _ | / _ \/_  __/ _ |     Version : 0.20.2
INFO[0000]  _\ \/ ___/ __ |/ , _/ / / / __ |     SHA     : 740028b
INFO[0000] /___/_/  /_/ |_/_/|_| /_/ /_/ |_|     Go      : go1.9.1
INFO[0000]
INFO[0000] ══════════════════════════════════════════════════════════════
INFO[0000] Service: SpartaStep-mweagle                   LinkFlags= Option=provision UTC="2017-11-01T02:03:04Z"
INFO[0000] ══════════════════════════════════════════════════════════════
INFO[0000] Provisioning service                          BuildID=69ecad9a90c763922e292cd22d63b6874dd3195c CodePipelineTrigger= InPlaceUpdates=false NOOP=false Tags=
INFO[0000] Verifying IAM Lambda execution roles
INFO[0000] IAM roles verified                            Count=1
INFO[0000] Checking S3 versioning                        Bucket=XXXXXXXXX VersioningEnabled=true
INFO[0000] Running `go generate`
INFO[0000] Compiling binary                              Name=Sparta.lambda.amd64
INFO[0010] Executable binary size                        KB=20162 MB=19
INFO[0010] Creating code ZIP archive for upload          TempName=./.sparta/SpartaStep_mweagle-code.zip
INFO[0011] Creating NodeJS/Sparta proxy function         FunctionName=StepRollDie ScriptName=StepRollDie
INFO[0011] Lambda code archive size                      KB=20261 MB=19
INFO[0011] Uploading local file to S3                    Bucket=XXXXXXXXX Key=SpartaStep-mweagle/SpartaStep_mweagle-code.zip Path=./.sparta/SpartaStep_mweagle-code.zip
INFO[0026] Calling WorkflowHook                          WorkflowHook="github.com/mweagle/Sparta/aws/step.(*StateMachine).StateMachineDecorator.func1" WorkflowHookContext="map[]"
INFO[0026] Uploading local file to S3                    Bucket=XXXXXXXXX Key=SpartaStep-mweagle/SpartaStep_mweagle-cftemplate.json Path=./.sparta/SpartaStep_mweagle-cftemplate.json
INFO[0027] Creating stack                                StackID="arn:aws:cloudformation:us-west-2:000000000000:stack/SpartaStep-mweagle/dbc121a0-bea8-11e7-8184-503acbd4dcfd"
INFO[0045] Waiting for CloudFormation operation to complete
INFO[0067] Waiting for CloudFormation operation to complete
INFO[0085] Stack provisioned                             CreationTime="2017-11-01 02:03:30.945 +0000 UTC" StackId="arn:aws:cloudformation:us-west-2:000000000000:stack/SpartaStep-mweagle/dbc121a0-bea8-11e7-8184-503acbd4dcfd" StackName=SpartaStep-mweagle
INFO[0085] ──────────────────────────────────────────────────────────────
INFO[0085] SpartaStep-mweagle Summary (2017-10-31T19:04:29-07:00)
INFO[0085] ──────────────────────────────────────────────────────────────
INFO[0085] Verifying IAM roles                           Duration (s)=0
INFO[0085] Verifying AWS preconditions                   Duration (s)=0
INFO[0085] Creating code bundle                          Duration (s)=11
INFO[0085] Uploading code                                Duration (s)=15
INFO[0085] Ensuring CloudFormation stack                 Duration (s)=59
INFO[0085] Total elapsed time                            Duration (s)=85
INFO[0085] ──────────────────────────────────────────────────────────────
{{< /highlight >}}

### Testing

With the stack provisioned, the final step is to test the State Machine and see how lucky our die roll is. Navigate to the **Step Functions**
service dashboard in the AWS Console and find the State Machine that was provisioned. Click the **New Execution** button and accept the default JSON.
This sample state machine doesn't interrogate the incoming data so the initial JSON is effectively ignored.

For this test the first roll was a `4`, so there was only 1 path through the state machine. Depending on your
die roll, you may see different state machine paths through the `WaitState`.

![State Machine](/images/step_functions/step_execution.jpg)

## Wrapping Up

AWS Step Functions are a powerful tool that allows you to orchestrate long running workflows using AWS Lambda. State functions
are useful to implement the Saga pattern as in [here](http://theburningmonk.com/2017/07/applying-the-saga-pattern-with-aws-lambda-and-step-functions/) and
[here](https://read.acloud.guru/how-the-saga-pattern-manages-failures-with-aws-lambda-and-step-functions-bc8f7129f900). They can also be used
to compose Lambda functions into more complex workflows that include [parallel](https://states-language.net/spec.html#parallel-state) operations
on shared data.

# Notes
  * Minimal State machine validation is done at this time. See [Tim Bray](https://www.tbray.org/ongoing/When/201x/2016/12/01/J2119-Validator) for more information.
  * Value interrogation is defined by [JSONPath](http://goessner.net/articles/JsonPath/) expressions
