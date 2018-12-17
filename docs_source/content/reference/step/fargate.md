---
date: 2018-12-16 07:02:28
title: Fargate
weight: 20
---

{{< tweet 1073347546553217024 >}}

While Serverless and FaaS are often used interchangeably, there are types of
workloads that are more challenging to move to FaaS. Perhaps due to third
party libraries, latency, or storage requirements, the FaaS model
isn't an ideal fit. An example that is commonly provided is the need
to run [ffmpeg](https://www.ffmpeg.org/).

To benefit from the serverless model in these cases, Sparta provides
the ability to leverage the [Fargate](https://aws.amazon.com/fargate) service
to run Containers without needing to manage servers.

There are several steps to _Fargate-ifying_ your application and Sparta exposes
functions and hooks to make that operation scoped to a `provision` operation.

Those steps include:

  1. Make the application Task-aware
  1. Package your application in a Docker image
  1. Push the Docker image to [Amazon ECR](https://aws.amazon.com/ecr/)
  1. Reference the ECR URL in a Fargate Task
  1. Provision an ECS cluster that hosts the Task

This overview is based on the [SpartaStepServicefull](https://github.com/mweagle/SpartaStepServicefull)
project. The implementation uses a combination of [ServiceDecoratorHookHandlers](https://godoc.org/github.com/mweagle/Sparta#ServiceDecoratorHookHandler)
to achieve the end result.

Please see [servicefull_build.go](https://github.com/mweagle/SpartaStepServicefull/blob/master/bootstrap/servicefull_build.go)
for the most up-to-date version of code samples.

## Task Awareness

The first step is to provide an opportunity for our application to behave
differently when run as a Fargate task. To do this we add a new
application subcommand option that augments the standard `Main` behavior:

```go
// Add a hook to do something
fargateTask := &cobra.Command{
  Use:   "fargateTask",
  Short: "Sample Fargate task",
  Long:  `Sample Fargate task that simply logs a message"`,
  RunE: func(cmd *cobra.Command, args []string) error {
    fmt.Printf("Insert your Fargate code here! 🎉")
    return nil
  },
}
// Register the command with the Sparta root dispatcher. This
// command `fargateTask` matches the command line option in the
// Dockerfile that is used to build the image.
sparta.CommandLineOptions.Root.AddCommand(fargateTask)
```

This subcommand is defined in the [servicefull_task](https://github.com/mweagle/SpartaStepServicefull/blob/master/bootstrap/servicefull_task.go)
file. Note that the file uses `go` [build tags](https://dave.cheney.net/2013/10/12/how-to-use-conditional-compilation-with-the-go-build-tool)
so that the new **fargateTask** subcommand is only available when the
build target includes the _lambdaBinary_ flag:

```go
// +build lambdabinary

package bootstrap
```

We can now package our Task-aware executable and deploy it to the cloud.

## Package

The first step is to create a version of your application that
can support a Fargate task. This is done in the `ecrImageBuilderDecorator`
function which delegates the compiling and image creation to Sparta:

```go
// Always build the image
buildErr := spartaDocker.BuildDockerImage(serviceName,
  "",
  dockerTags,
  logger)
```

The second empty argument above is an optional _Dockerfile_ path. The sample
project uses the default _Dockerfile_ filename and defines that at the root
of the repository. The full _Dockerfile_ is:

```docker
FROM alpine:3.8
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

# Sparta provides the SPARTA_DOCKER_BINARY argument to the builder
# in order to embed the binary.
# Ref: https://docs.docker.com/engine/reference/builder/
ARG SPARTA_DOCKER_BINARY
ADD $SPARTA_DOCKER_BINARY /SpartaServicefull
CMD ["/SpartaServicefull", "fargateTask"]
```

The `BuildDockerImage` function supplies the transient binary executable
path to docker via the **SPARTA_DOCKER_BINARY** [ARG](https://docs.docker.com/engine/reference/commandline/build/#set-build-time-variables---build-arg)
value.


The `CMD` instruction includes our previously registered **fargateTask**
subcommand name to invoke the Task-appropriate codepath at runtime.

The log output includes the docker build info:
```
INFO[0002] Calling WorkflowHook
  ServiceDecoratorHook=github.com/mweagle/SpartaStepServicefull/bootstrap.ecrImageBuilderDecorator.func1
  WorkflowHookContext="map[]"
INFO[0002] Docker version 18.09.0, build 4d60db4
INFO[0002] Running `go generate`
INFO[0002] Compiling binary
  Name=ServicefulStepFunction-1544976454011339000-docker.lambda.amd64
INFO[0003] Creating Docker image
  Tags="map[servicefulstepfunction:adc67a77aef22b6dab9c6156d13853e2cfe06488.1544976453]"
 NFO[0004] Sending build context to Docker daemon  35.43MB
INFO[0004] Step 1/5 : FROM alpine:3.8
INFO[0004]  ---> 196d12cf6ab1
INFO[0004] Step 2/5 : RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
INFO[0004]  ---> Using cache
INFO[0004]  ---> 99402375b7f2
INFO[0004] Step 3/5 : ARG SPARTA_DOCKER_BINARY
INFO[0004]  ---> Using cache
INFO[0004]  ---> a44d27522c40
INFO[0004] Step 4/5 : ADD $SPARTA_DOCKER_BINARY /SpartaServicefull
INFO[0005]  ---> 87ffd10e9901
INFO[0005] Step 5/5 : CMD ["/SpartaServicefull", "fargateTask"]
INFO[0005]  ---> Running in 0a3b503201c7
INFO[0005] Removing intermediate container 0a3b503201c7
INFO[0005]  ---> 7cb1b2261a92
INFO[0005] Successfully built 7cb1b2261a92
INFO[0005] Successfully tagged
  servicefulstepfunction:adc67a77aef22b6dab9c6156d13853e2cfe06488.1544976453
```

## Push to ECR

The next step is to push the locally built image to the Elastic
Container Registry. The push will return either the ECR URL
which will be used as Fargate Task [image](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ecs-taskdefinition-containerdefinitions.html#cfn-ecs-taskdefinition-containerdefinition-image)
property or an error:

```go
// Push the image to ECR & store the URL s.t. we can properly annotate
// the CloudFormation template
ecrURLPush, pushImageErr := spartaDocker.PushDockerImageToECR(buildTag,
  ecrRepositoryName,
  awsSession,
  logger)
```

The ECR push URL is stored in the `context` variable so that a downstream
Fargate cluster builder knows the image to use:

```go
context[contextKeyImageURL] = ecrURLPush
```

## State Machine

The Step Function definition indirectly references the Fargate
Task via task specific [parameters](https://docs.aws.amazon.com/step-functions/latest/dg/connectors-ecs.html):

```go
fargateParams := spartaStep.FargateTaskParameters{
  LaunchType:     "FARGATE",
  Cluster:        gocf.Ref(resourceNames.ECSCluster).String(),
  TaskDefinition: gocf.Ref(resourceNames.ECSTaskDefinition).String(),
  NetworkConfiguration: &spartaStep.FargateNetworkConfiguration{
    AWSVPCConfiguration: &gocf.ECSServiceAwsVPCConfiguration{
      Subnets: gocf.StringList(
        gocf.Ref(resourceNames.PublicSubnetAzs[0]).String(),
        gocf.Ref(resourceNames.PublicSubnetAzs[1]).String(),
      ),
      AssignPublicIP: gocf.String("ENABLED"),
    },
  },
}
fargateState := spartaStep.NewFargateTaskState("Run Fargate Task", fargateParams)
```

The **ECSCluster** and **ECSTaskDefinition** are resources that are provisioned
by the `fargateClusterDecorator` decorator function.

## Fargate Cluster

The final step is to provision the ECS cluster that supports the Fargate
task. This is encapsulated in the `fargateClusterDecorator` which creates
the required set of CloudFormation resources. The set of CloudFormation
resource names is represented in the `stackResourceNames` struct:

```go
type stackResourceNames struct {
  StepFunction              string
  SNSTopic                  string
  ECSCluster                string
  ECSRunTaskRole            string
  ECSTaskDefinition         string
  ECSTaskDefinitionLogGroup string
  ECSTaskDefinitionRole     string
  VPC                       string
  InternetGateway           string
  AttachGateway             string
  RouteViaIgw               string
  PublicRouteViaIgw         string
  ECSSecurityGroup          string
  PublicSubnetAzs           []string
}
```

The [ECS Task Definition](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task_definitions.html)
is of particular interest and is where the inline created **ECR_URL** is used to
define a FARGATE task.

### ECS Task Definition

```go
imageURL, _ := context[contextKeyImageURL].(string)
if imageURL == "" {
  return errors.Errorf("Failed to get image URL from context with key %s",
    contextKeyImageURL)
}
...
// Create the ECS task definition
ecsTaskDefinition := &gocf.ECSTaskDefinition{
  ExecutionRoleArn:        gocf.GetAtt(resourceNames.ECSTaskDefinitionRole, "Arn"),
  RequiresCompatibilities: gocf.StringList(gocf.String("FARGATE")),
  CPU:                     gocf.String("256"),
  Memory:                  gocf.String("512"),
  NetworkMode:             gocf.String("awsvpc"),
  ContainerDefinitions: &gocf.ECSTaskDefinitionContainerDefinitionList{
    gocf.ECSTaskDefinitionContainerDefinition{
      Image:     gocf.String(imageURL),
      Name:      gocf.String("sparta-servicefull"),
      Essential: gocf.Bool(true),
      LogConfiguration: &gocf.ECSTaskDefinitionLogConfiguration{
        LogDriver: gocf.String("awslogs"),
        // Options Ref: https://docs.aws.amazon.com/AmazonECS/latest/developerguide/AWS_Fargate.html
        Options: map[string]interface{}{
          "awslogs-region": gocf.Ref("AWS::Region"),
          "awslogs-group": strings.Join([]string{"",
            sparta.ProperName,
            serviceName}, "/"),
          "awslogs-stream-prefix": serviceName,
          "awslogs-create-group":  "true",
        },
      },
    },
  },
}
```

## Configuration

The final step is to provide the three decorators to the
[WorkflowHooks](https://godoc.org/github.com/mweagle/Sparta#WorkflowHooks) structure:

```go
workflowHooks := &sparta.WorkflowHooks{
  ServiceDecorators: []sparta.ServiceDecoratorHookHandler{
    ecrImageBuilderDecorator("spartadocker"),
    // Then build the state machine
    stateMachine.StateMachineDecorator(),
    // Then the ECS cluster that supports the Fargate task
    fargateClusterDecorator(resourceNames),
  },
}
```

## Provisioning

The provisioning workflow for this service is the same as a Lambda-based one:

```shell
$ go run main.provision --s3Bucket $MY_S3_BUCKET
```

Output:


```
INFO[0000] ════════════════════════════════════════════════
INFO[0000] ╔═╗╔═╗╔═╗╦═╗╔╦╗╔═╗   Version : 1.8.0
INFO[0000] ╚═╗╠═╝╠═╣╠╦╝ ║ ╠═╣   SHA     : 597d3ba
INFO[0000] ╚═╝╩  ╩ ╩╩╚═ ╩ ╩ ╩   Go      : go1.11.1
INFO[0000] ════════════════════════════════════════════════
INFO[0000] Service: ServicefulStepFunction
  LinkFlags= Option=provision UTC="2018-12-16T16:07:31Z"
INFO[0000] ════════════════════════════════════════════════
INFO[0000] Using `git` SHA for StampedBuildID
  Command="git rev-parse HEAD" SHA=adc67a77aef22b6dab9c6156d13853e2cfe06488
INFO[0000] Provisioning service
  BuildID=adc67a77aef22b6dab9c6156d13853e2cfe06488
  CodePipelineTrigger=
  InPlaceUpdates=false
  NOOP=false Tags=
WARN[0000] No lambda functions provided to Sparta.Provision()
INFO[0000] Verifying IAM Lambda execution roles
INFO[0000] IAM roles verified                            Count=0
```


## Result

The end result is a Step function that uses our `go` binary, Step functions,
and SNS rather than Lambda functions:

![Step Function](https://raw.githubusercontent.com/mweagle/Sparta/master/docs_source/static/site/1.8.0/step_functions_fargate.jpg "Step Function")
