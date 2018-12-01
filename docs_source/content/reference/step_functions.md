---
date: 2017-10-31 18:20:05
title: Step Functions
weight: 400
---

AWS [Step Functions](https://aws.amazon.com/step-functions/) are a powerful way to express long-running, complex workflows comprised of Lambda functions. With Sparta 0.20.2, you can build a State Machine as part of your application. This section walks through the three steps necessary to provision a sample "Roll Die" state machine using a single Lambda function. See [SpartaStep](https://github.com/mweagle/SpartaStep) for the full source.

## Lambda Functions

The first step is to define the core Lambda function [Task](http://docs.aws.amazon.com/step-functions/latest/dg/amazon-states-language-task-state.html) that will be our Step function's core logic. In this example, we'll define a _rollDie_ function:

```go
type lambdaRollResponse struct {
  Roll int `json:"roll"`
}

// Standard AWS λ function
func lambdaRollDie(ctx context.Context) (lambdaRollResponse, error) {
  return lambdaRollResponse{
    Roll: rand.Intn(5) + 1,
  }, nil
}
```

### State Machine

Our state machine is simple: we want to keep rolling the die until we get a "good" result, with a delay in between rolls:

![State Machine](/images/step_functions/roll_die.jpg)

To do this, we use the new `github.com/mweagle/Sparta/aws/step` functions to define the other states.

```go
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
```

The Sparta state types correspond to their [AWS States Spec](https://states-language.net/spec.html) equivalents:

  - `successState` : [SucceedState](https://states-language.net/spec.html#succeed-state)
  - `delayState` : a specialized [WaitState](https://states-language.net/spec.html#wait-state)
  - `choiceState`: [ChoiceState](https://states-language.net/spec.html#choice-state)

The `choiceState` is the most interesting state: based on the JSON response of the `lambdaRollDie`, it will either transition
to a delay or the success end state.

See [godoc](https://godoc.org/github.com/mweagle/Sparta/aws/step) for the complete set of types.

The `lambdaTaskState` uses a normal Sparta function as in:

```go
lambdaFn := sparta.HandleAWSLambda("StepRollDie",
  lambdaRollDie,
  sparta.IAMRoleDefinition{})

lambdaFn.Options.MemorySize = 128
lambdaFn.Options.Tags = map[string]string{
  "myAccounting": "tag",
}
```

The final step is to hook up the state transitions for states that don't implicitly include them, and create the State Machine:

```go
// Hook up the transitions
lambdaTaskState.Next(choiceState)
delayState.Next(lambdaTaskState)

// Startup the machine with a user-scoped name for account uniqueness
stateMachineName := spartaCF.UserScopedStackName("StateMachine")
startMachine := step.NewStateMachine(stateMachineName, lambdaTaskState)
```

At this point we have a potentially well-formed [Lambda-powered](http://docs.aws.amazon.com/step-functions/latest/dg/tutorial-creating-lambda-state-machine.html) State Machine.
The final step is to attach this machine to the normal service definition.

### Service Decorator

The return type from `step.NewStateMachine(...)` is a `*step.StateMachine` instance that exposes a [ServiceDecoratorHook](https://godoc.org/github.com/mweagle/Sparta#ServiceDecoratorHook).
Adding the hook to your service's Workflow Hooks (similar to provisioning a service-scoped [CloudWatch Dashboard](https://github.com/mweagle/Sparta/blob/master/CHANGES.md#v0130))
will include it in the CloudFormation template serialization:

```go
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
```

With the decorator attached, the next service `provision` request will include the state machine as above.

```text

$ go run main.go provision --s3Bucket weagle
INFO[0000] ════════════════════════════════════════════════
INFO[0000] ╔═╗┌─┐┌─┐┬─┐┌┬┐┌─┐   Version : 1.0.2
INFO[0000] ╚═╗├─┘├─┤├┬┘ │ ├─┤   SHA     : b37b93e
INFO[0000] ╚═╝┴  ┴ ┴┴└─ ┴ ┴ ┴   Go      : go1.9.2
INFO[0000] ════════════════════════════════════════════════
INFO[0000] Service: SpartaStep-mweagle                   LinkFlags= Option=provision UTC="2018-01-29T14:33:36Z"
INFO[0000] ════════════════════════════════════════════════
INFO[0000] Provisioning service                          BuildID=f7ade93d3900ab4b01c468c1723dedac24cbfa93 CodePipelineTrigger= InPlaceUpdates=false NOOP=false Tags=
INFO[0000] Verifying IAM Lambda execution roles
INFO[0000] IAM roles verified                            Count=1
INFO[0000] Checking S3 versioning                        Bucket=weagle VersioningEnabled=true
INFO[0000] Checking S3 region                            Bucket=weagle Region=us-west-2
INFO[0000] Running `go generate`
INFO[0000] Compiling binary                              Name=Sparta.lambda.amd64
INFO[0010] Creating code ZIP archive for upload          TempName=./.sparta/SpartaStep_mweagle-code.zip
INFO[0010] Lambda code archive size                      Size="13 MB"
INFO[0010] Uploading local file to S3                    Bucket=weagle Key=SpartaStep-mweagle/SpartaStep_mweagle-code.zip Path=./.sparta/SpartaStep_mweagle-code.zip Size="13 MB"
INFO[0020] Calling WorkflowHook                          ServiceDecoratorHook="github.com/mweagle/Sparta/aws/step.(*StateMachine).StateMachineDecorator.func1" WorkflowHookContext="map[]"
INFO[0020] Uploading local file to S3                    Bucket=weagle Key=SpartaStep-mweagle/SpartaStep_mweagle-cftemplate.json Path=./.sparta/SpartaStep_mweagle-cftemplate.json Size="3.7 kB"
INFO[0021] Creating stack                                StackID="arn:aws:cloudformation:us-west-2:123412341234:stack/SpartaStep-mweagle/6ff65180-0501-11e8-935b-50a68d01a629"
INFO[0094] CloudFormation provisioning metrics:
INFO[0094] Operation duration                            Duration=54.73s Resource=SpartaStep-mweagle Type="AWS::CloudFormation::Stack"
INFO[0094] Operation duration                            Duration=19.02s Resource=IAMRole49969e8a894b9eeea02a4936fb9519f2bd67dbe6 Type="AWS::IAM::Role"
INFO[0094] Operation duration                            Duration=18.69s Resource=StatesIAMRolee00aa3484b0397c676887af695abfd160104318a Type="AWS::IAM::Role"
INFO[0094] Operation duration                            Duration=2.60s Resource=StateMachine59f153f18068faa0b7fb588350be79df422ba5ef Type="AWS::StepFunctions::StateMachine"
INFO[0094] Operation duration                            Duration=2.28s Resource=StepRollDieLambda7d9f8ab476995f16b91b154f68e5f5cc42601ebf Type="AWS::Lambda::Function"
INFO[0094] Stack provisioned                             CreationTime="2018-01-29 14:33:56.7 +0000 UTC" StackId="arn:aws:cloudformation:us-west-2:123412341234:stack/SpartaStep-mweagle/6ff65180-0501-11e8-935b-50a68d01a629" StackName=SpartaStep-mweagle
INFO[0094] ════════════════════════════════════════════════
INFO[0094] SpartaStep-mweagle Summary
INFO[0094] ════════════════════════════════════════════════
INFO[0094] Verifying IAM roles                           Duration (s)=0
INFO[0094] Verifying AWS preconditions                   Duration (s)=0
INFO[0094] Creating code bundle                          Duration (s)=10
INFO[0094] Uploading code                                Duration (s)=10
INFO[0094] Ensuring CloudFormation stack                 Duration (s)=73
INFO[0094] Total elapsed time                            Duration (s)=94
```

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
