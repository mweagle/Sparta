# Change Notes

## v1.13.1 - The post:Invent Edition 🗓

- :warning: **BREAKING**
- :checkered_flag: **CHANGES**
- :bug: **FIXED**
  - [Correct documentation links](https://github.com/mweagle/Sparta/issues/160)

## v1.13.0 - The pre:Invent Edition 🗓

- :warning: **BREAKING**
- :checkered_flag: **CHANGES**

  - Updated [go-cloudformation](https://github.com/mweagle/go-cloudformation) dependency to expose:
    - [LambdaEventInvokeConfig](https://godoc.org/github.com/mweagle/go-cloudformation#LambdaEventInvokeConfig) for success/failure handlers
      - See the [blog post](https://aws.amazon.com/blogs/compute/introducing-aws-lambda-destinations/) for more information
      - Destinations can be connected via [TemplateDecorators](https://godoc.org/github.com/mweagle/Sparta#TemplateDecorator) or [ServiceDecoratorHook](https://godoc.org/github.com/mweagle/Sparta#ServiceDecoratorHook)
    - [LambdaEventSourceMapping](https://godoc.org/github.com/mweagle/go-cloudformation#LambdaEventSourceMapping) for updated stream controls
      - See the [blog post](https://aws.amazon.com/blogs/compute/new-aws-lambda-scaling-controls-for-kinesis-and-dynamodb-event-sources/) for more information
  - Added [cloudwatch.EmbeddedMetric](https://godoc.org/github.com/mweagle/Sparta/aws/cloudwatch#EmbeddedMetric) to support publishing CloudWatch [Embedded Metrics](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/CloudWatch_Embedded_Metric_Format_Specification.html) via logs

    - See the [blog post](https://aws.amazon.com/about-aws/whats-new/2019/11/amazon-cloudwatch-launches-embedded-metric-format/) for more information
    - Usage:

      ```go
      metricDirective := emMetric.NewMetricDirective("SpecialNamespace",
        // Initialize with metric dimensions
        map[string]string{"functionVersion": os.Getenv("AWS_LAMBDA_FUNCTION_VERSION")})
      // Record a metric value
      metricDirective.Metrics["invocations"] = cloudwatch.MetricValue{
        Unit:  cloudwatch.UnitCount,
        Value: 1,
      }
      // Publish optionally accepts additional high-cardinality KV properties
      emMetric.Publish(nil)
      ```

- :bug: **FIXED**
  - [Set executable bit on Sparta binary in ZIP archive](https://github.com/mweagle/Sparta/issues/158)

## v1.12.0 - The Mapping Edition 🗺

- :warning: **BREAKING**
- :checkered_flag: **CHANGES**
  - Added [step.MapState](https://godoc.org/github.com/mweagle/Sparta/aws/step#MapState) to support creating AWS Step functions that support the new [Map State](https://states-language.net/spec.html#map-state)
    - See the [blog post](https://aws.amazon.com/blogs/aws/new-step-functions-support-for-dynamic-parallelism/) for more details
    - Also the corresponding sample application in the [Sparta Step](https://github.com/mweagle/SpartaStep/blob/master/parallel/main.go) project.
- :bug: **FIXED**
  - Fixed latent issue in [step.ParallelState](https://godoc.org/github.com/mweagle/Sparta/aws/step#ParallelState) that prevented `Branches` field from being properly marshaled.

## v1.11.0 - The Firehose Edition 🚒

- :warning: **BREAKING**
- :checkered_flag: **CHANGES**
  - Added [archetype.NewKinesisFirehoseTransformer](https://godoc.org/github.com/mweagle/Sparta/archetype#NewKinesisFirehoseTransformer) and [archetype.NewKinesisFirehoseLambdaTransformer](https://godoc.org/github.com/mweagle/Sparta/archetype#NewKinesisFirehoseLambdaTransformer) to support Kinesis Firehose Lambda Transforms
    - See the [documentation](http://gosparta.io/reference/archetypes/kinesis_firehose/) for more details
    - Also the corresponding sample application at the [SpartaXForm repo](https://github.com/mweagle/SpartaXForm).
- :bug: **FIXED**

## v1.10.0 - The Load Balancer Edition ⚖️

- :warning: **BREAKING**
- :checkered_flag: **CHANGES**
  - Added [ApplicationLoadBalancerDecorator](https://godoc.org/github.com/mweagle/Sparta/decorator#ApplicationLoadBalancerDecorator) type to support Lambda functions as load balancer targets.
    - See the [documentation](http://gosparta.io/reference/decorators/application_load_balancer/) for more details
    - Also the corresponding sample application at the [SpartaALB repo](https://github.com/mweagle/SpartaALB).
- :bug: **FIXED**

## v1.9.4 - The Sockets Edition 🔌

- :warning: **BREAKING**
  - Update `sparta.Main` and `sparta.MainEx` to accept new _APIGateway_ interface type rather than concrete _API_ type. This should be backward compatible for most usage and was done to support the new APIV2 Gateway type.
- :checkered_flag: **CHANGES**
  - Added [API V2] type to provision WebSocket APIs
    - See the [documentation](http://gosparta.io/reference/apiv2gateway/) for more details
  - Update to `go` [modules](https://github.com/golang/go/wiki/Modules)
- :bug: **FIXED**

## v1.9.3 - The Discovery Edition ☁️🔍

- :warning: **BREAKING**
- :checkered_flag: **CHANGES**
  - Added [Cloud Map](https://aws.amazon.com/cloud-map/) discovery publisher
    - See the [documentation](https://gosparta.io/reference/decorators/cloudmap/)
  - Added `panic` recover handler to more gracefully handle exceptions
  - Include AWS Session in context with key `sparta.ContextKeyAWSSession`
- :bug: **FIXED**

  - [Update to support new Amazon Linux AMI](https://github.com/mweagle/Sparta/issues/145)
  - Fixed latent issue where `env` specified log level wasn't respected at lambda execution time

## v1.9.2 - The Names Edition 📛

- :warning: **BREAKING**
- :checkered_flag: **CHANGES**

  - Added `API.EndpointConfiguration` field to [API](https://godoc.org/github.com/mweagle/Sparta#API).
    - This field exposes the [EndpointConfiguration](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-apigateway-restapi-endpointconfiguration.html) property to specify either _EDGE_ or _REGIONAL_ API types.
  - Added `decorator.APIGatewayDomainDecorator` to associate a custom domain with an API Gateway instance

    - Usage:

      ```go
        hooks := &sparta.WorkflowHooks{}
        serviceDecorator := spartaDecorators.APIGatewayDomainDecorator(apiGateway,
          gocf.String(acmCertARNLiteral),
          "", // Optional base path value
          "subdomain.mydomain.net")
        hooks.ServiceDecorators = []sparta.ServiceDecoratorHookHandler{
          serviceDecorator,
        }
      ```

    - See [apigateway_domain_test](https://github.com/mweagle/Sparta/blob/master/decorator/dashboard.go) for a complete example.
    - See the [AWS Documentation](https://docs.aws.amazon.com/apigateway/latest/developerguide/how-to-custom-domains.html) for more information.

- :bug: **FIXED**
  - [Support custom domains](https://github.com/mweagle/Sparta/issues/91)

## v1.9.1 - The CodeCommitment Edition 💕

- :warning: **BREAKING**
- :checkered_flag: **CHANGES**

  - Added `CodeCommitPermission` type to support CodeCommit [notifications](https://docs.aws.amazon.com/codecommit/latest/userguide/how-to-repository-email.html)
  - There is an _archetype_ constructor that encapsulates this type of Lambda reactor.

    - Usage:

      ```go
      func echoCodeCommit(ctx context.Context,
        event awsLambdaEvents.CodeCommitEvent) (interface{}, error) {
        // ...
        return &event, nil
      }
      func main() {
        // ...
        reactor, reactorErr := spartaArchetype.NewCodeCommitReactor(spartaArchetype.CodeCommitReactorFunc(echoCodeCommit),
            gocf.String("TestCodeCommitRepo"),
            nil,
            nil,
            nil)
        ...
      }
      ```

  - Updated to [staticcheck.io](https://staticcheck.io/)

- :bug: **FIXED**
  - [Add CodeCommit support](https://github.com/mweagle/Sparta/issues/86)
  - [Fixed broken link to AWS documentation](https://github.com/mweagle/Sparta/pull/136)
  - [RegisterLambdaUtilizationMetricPublisher Name ref obsolete](https://github.com/mweagle/Sparta/issues/130)
  - [Document archetype constructors](https://github.com/mweagle/Sparta/issues/119)

## v1.9.0 - The LayerCake Edition 🍰

- :warning: **BREAKING**
- :checkered_flag: **CHANGES**

  - Added `LambdaAWSInfo.Layers` field to support [Lambda Layers](https://docs.aws.amazon.com/lambda/latest/dg/configuration-layers.html)

    - Usage:

      ```go
      lambdaRole := sparta.IAMRoleDefinition{
        Privileges: []sparta.IAMRolePrivilege{
          iamBuilder.Allow("lambda:GetLayerVersion").
            ForResource().
            Literal("*").
            ToPrivilege(),
        },
      }
      lambdaFn, lambdaFnErr := sparta.NewAWSLambda("Hello World",
        helloWorld,
        lambdaRole)
      lambdaFn.Layers = []gocf.Stringable{
        gocf.String("arn:aws:lambda:us-west-2:123412341234:layer:ffmpeg:1"),
      }
      ```

  - Added `WithCondition` to [IAM Builder](https://godoc.org/github.com/mweagle/Sparta/aws/iam/builder)
  - Added `s3Site.UserManifestData` map property to allow for custom user data to be included in _MANIFEST.json_ content that is deployed to an S3 Site bucket.
    - Userdata is scoped to a **userdata** keyname in _MANIFEST.json_
    - See the [SpartaAmplify](https://github.com/mweagle/SpartaAmplify) sample app for a complete example.
  - Added `github.com/mweagle/Sparta/system.RunAndCaptureOSCommand`
    - This is convenience function to support alternative `io.Writer` sinks for _stdout_ and _stderr_.
  - Minor usability improvements to `--status` report output

- :bug: **FIXED**
  - [overview page is broken](https://github.com/mweagle/Sparta/issues/133)

## v1.8.0 - The #postReInvent Edition ⌛️

- :warning: **BREAKING**
  - Renamed `archetype.CloudWatchLogsReactor` to `archetype.CloudWatchReactor`
    - Also changed `OnLogMessage` to `OnCloudWatchMessage`
      - I consistently forget the fact that CloudWatch is more than logs
    - Moved the internal `cloudwatchlogs` package to the `cloudwatch/logs` import path
  - Renamed fluent typenames in _github.com/mweagle/Sparta/aws/iam/builder_ to support Principal-based builders
  - Renamed `step.NewTaskState` to `step.NewLambdaTaskState` to enable type specific [Step function services](https://docs.aws.amazon.com/step-functions/latest/dg/concepts-connectors.html).
  - Simplified versioning Lambda resource so that the [Lambda::Version](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-lambda-version.html) resource is orphaned (via [DeletionPolicy](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-attribute-deletionpolicy.html)) rather than the prior implementation, which fetched all versions from the provisioned template and accumulated them over time.
    - This also obsoleted the `ContextKeyLambdaVersions` constant
- :checkered_flag: **CHANGES**

  - More documentation
  - Added Step function [service integrations](https://docs.aws.amazon.com/step-functions/latest/dg/connectors-supported-services.html)
    - See the [SpartaStepServicefull](https://github.com/mweagle/SpartaStepServicefull) project for an example of a service that:
      - Provisions no Lambda functions
      - Dockerizes itself
      - Pushes that image to ECR
      - Uses the resulting ECR Image URL as a Fargate Task in a Step function:
      - <div align="center"><img src="https://raw.githubusercontent.com/mweagle/Sparta/master/docs_source/static/site/1.8.0/step_functions_fargate.jpg" />
  - Added _github.com/mweagle/Sparta/aws/iam/builder.IAMBuilder::ForPrincipals_ fluent builder. Example usage:

    ```go
      "Statement": []spartaIAM.PolicyStatement{
        iamBuilder.Allow("sts:AssumeRole").
          ForPrincipals("states.amazonaws.com").
          ToPolicyStatement(),
    ```

  - Upgraded to `docker login --password-stdin` for local authentication. Previously used `docker login --password`. Example:

    ```plain
    INFO[0005] df64d3292fd6: Preparing
    INFO[0006] denied: Your Authorization Token has expired. Please run 'aws ecr get-login --no-include-email' to fetch a new one.
    INFO[0006] ECR push failed - reauthorizing               Error="exit status 1"
    INFO[0006] Login Succeeded
    INFO[0006] The push refers to repository [123412341234.dkr.ecr.us-west-2.amazonaws.com/argh]
    ```

    - See the [Docker docs](https://docs.docker.com/engine/reference/commandline/login/#parent-command)

  - Include `docker -v` output in log when calling [BuildDockerImage](https://godoc.org/github.com/mweagle/Sparta/docker#BuildDockerImage)
  - Added `StateMachineNamedDecorator(stepFunctionResourceName)` to supply the name of the Step function
  - Migrated all API-Gateway integration mappings to use the [mapping override](https://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-mapping-template-reference.html) support in VTL.

    - This reduces the number of API-Gateway RegExp-based integration mappings and relies on a Lambda function returning a shape that matches the default _application/json_ expectations:

      ```json
      {
        "code" : int,
        "body" : ...,
        "headers": {
          "x-lowercase-header" : "foo",
        }
      }
      ```

    - The default shape can be customized by providing custom mapping templates to the [IntegrationResponses](https://godoc.org/github.com/mweagle/Sparta#IntegrationResponse)

  - [rest.MethodHandler:Headers](https://godoc.org/github.com/mweagle/Sparta/archetype/rest#MethodHandler.Headers) has been deprecated.
    - Moving all header management to VTL eliminated the need to explicitly declare headers.
  - Added `spartaDecorators.PublishAllResourceOutputs(cfResourceName, gocf.ResourceProperties)` which adds all the associated resource `Ref` and `Att` values to the Stack Outputs
    - The set of `Att` values is extracted from the [CloudFormation Resource Specification](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/cfn-resource-specification.html) via the [go-cloudformation](https://github.com/mweagle/go-cloudformation) project.

- :bug: **FIXED**
  - API Gateway custom headers were not being properly returned
  - [RegisterLambdaUtilizationMetricPublisher Name ref obsolete](https://github.com/mweagle/Sparta/issues/130)

## v1.7.3 - The Documentation Edition 📚

- :warning: **BREAKING**
  - Renamed `archetype.NewCloudWatchLogsReactor` to `archetype.NewCloudWatchReactor`
- :checkered_flag: **CHANGES**
  - Moved all documentation into the _master_ branch to make it a bit easier to update docs together with code.
    - See _/docs_source/content/meta/\_index.md_ for how to edit, preview, and submit.
  - Added `archetype.NewCloudWatchScheduledReactor` and `archetype.NewCloudWatchEventedReactor`
- :bug: **FIXED**

## v1.7.2 - The Cloud Drift Edition v2 🌬☁️

- :warning: **BREAKING**
- :checkered_flag: **CHANGES**

  - Moved `decorator.DriftDetector` to `validator.DriftDetector` and changed signature to [ServiceValidationHookHandler](https://godoc.org/github.com/mweagle/Sparta#ServiceValidationHookHandler)

    - Clearly I was too focused on enabling drift detection than enabling it in an appropriate place.
    - Updated usage:

      ```go
        import (
          "github.com/mweagle/Sparta/validator"
        )
        workflowHooks := &sparta.WorkflowHooks{
          Validators: []sparta.ServiceValidationHookHandler{
            validator.DriftDetector(true),
          },
        }
      ```

  - Added `LambdaFuncName` to output when stack drift detected.

    - Example:

      ```plain
      WARN[0013] Stack drift detected                          Actual=debug Expected=info LambdaFuncName="Hello World" PropertyPath=/Environment/Variables/SPARTA_LOG_LEVEL Relation=NOT_EQUAL Resource=HelloWorldLambda80576f7b21690b0cb485a6b69c927aac972cd693
      ```

- :bug: **FIXED**

## v1.7.1 - The Cloud Drift Edition 🌬☁️

- :warning: **BREAKING**
- :checkered_flag: **CHANGES**

  - Added `decorator.DriftDetector` to optionally prevent operations in the presence of [CloudFormation Drift](https://aws.amazon.com/blogs/aws/new-cloudformation-drift-detection/).

    - Usage:

      ```go
      workflowHooks := &sparta.WorkflowHooks{
        PreBuilds: []sparta.WorkflowHookHandler{
          decorator.DriftDetector(false),
        },
      }
      ```

    - Sample output:

      ```text
      INFO[0001] Calling WorkflowHook                          Phase=PreBuild WorkflowHookContext="map[]"
      INFO[0001] Waiting for drift detection to complete       Status=DETECTION_IN_PROGRESS
      ERRO[0012] Stack drift detected                          Actual=debug Expected=info PropertyPath=/Environment/Variables/SPARTA_LOG_LEVEL Relation=NOT_EQUAL Resource=HelloWorldLambda80576f7b21690b0cb485a6b69c927aac972cd693
      INFO[0012] Invoking rollback functions
      ERRO[0012] Failed to provision service: DecorateWorkflow returned an error: stack MyHelloWorldStack-mweagle prevented update due to drift being detected
      ```

  - Usability improvements when errors produced. Previously the usage instructions were output on every failed command. Now they are only displayed if there are CLI argument validation errors.
  - Usability improvement to log individual [validation errors](https://godoc.org/gopkg.in/go-playground/validator.v9) if the CLI arguments are invalid.

- :bug: **FIXED**
  - Fixed latent issue where Sparta misreported its internal version

## v1.7.0 - The Time Machine Edition 🕰

- :warning: **BREAKING**
- :checkered_flag: **CHANGES**

  - Added `LambdaAWSInfo.Interceptors` support

    - `Interceptors` are functions (`func(context.Context, json.RawMessage) context.Context`) called in the normal event handling lifecycle to support cross cutting concerns. They are the runtime analog to `WorkflowHooks`.
    - The following stages are supported:
      - _Begin_: Called as soon as Sparta determines which user-function to invoke
      - _BeforeSetup_: Called before Sparta creates your lambda's `context` value
      - _AfterSetup_: Called after Sparta creates your lambda's `context` value
      - _BeforeDispatch_: Called before Sparta invokes your lambda function
      - _AfterDispatch_: Called after Sparta invokes your lambda function
      - _Complete_: Called immediately before Sparta returns your function return value(s) to AWS
    - The first interceptor is `interceptor.RegisterXRayInterceptor(ctx, options)` which creates a custom [XRay Segment](https://docs.aws.amazon.com/xray/latest/devguide/xray-sdk-go-segment.html) spanning your lambda's execution and supports:
      - Including the service BuildID in the [Trace Annotation](https://docs.aws.amazon.com/xray/latest/devguide/xray-api-segmentdocuments.html#api-segmentdocuments-annotations)
      - Optionally including the incoming event, all log statements (_trace_ and higher), and AWS request-id as [Trace Metadata](https://docs.aws.amazon.com/xray/latest/devguide/xray-api-segmentdocuments.html#api-segmentdocuments-metadata) **ONLY** in the case when your lambda function returns an error.
        - Log messages are stored in a [ring buffer](https://golang.org/pkg/container/ring/) and limited to 1024 entries.
    - This data is associated with XRay Traces in the console. Example:

      - <div align="center"><img src="https://raw.githubusercontent.com/mweagle/Sparta/master/docs_source/static/site/1.7.0/XRaySegment.jpg" />
        </div>

    - See the [SpartaXRayInterceptor](https://github.com/mweagle/SpartaXRayInterceptor) repo for a complete sample

    - Go back in time to when you wish you had enabled debug-level logging before the error ever occurred.

  - Expose `sparta.ProperName` as framework name literal
  - Add lightweight Key-Value interface and S3 and DynamoDB implementations to support [SpartaTodoBackend](https://github.com/mweagle/SpartaTodoBackend/)
    - The DynamoDB provider uses [dynamodbattribute](https://docs.aws.amazon.com/sdk-for-go/api/service/dynamodb/dynamodbattribute/) to map `go` structs to attributes.
    - See the [aws.accessor](https://godoc.org/github.com/mweagle/Sparta/aws/accessor) docs

- :bug: **FIXED**

## v1.6.0 - The REST Edition 😴

- :warning: **BREAKING**
  - Eliminate pre 1.0 GM Sparta function signature: `type LambdaFunction func(*json.RawMessage, *LambdaContext, http.ResponseWriter, *logrus.Logger)` 🎉
    - See the [AWS Docs](https://docs.aws.amazon.com/lambda/latest/dg/go-programming-model-handler-types.html) for officially supported signatures
  - Changed API Gateway response mapping to support body and header return values.
    - API Gateway lambda functions should use `aws/apigateway.NewResponse` to produce a new `Response` type with struct fields that are properly interpreted by the new `$input.json('$.body')` mapping expression.
    - The change was driven by the [SpartaTodoBackend](https://github.com/mweagle/SpartaTodoBackend) service's need to return both a body and HTTP location header.
      - See the [response](https://github.com/mweagle/SpartaTodoBackend/blob/master/service/todos.go#L79) for an example
- :checkered_flag: **CHANGES**

  - Add more _go_ idiomatic `sparta.NewAWSLambda(...) (*sparta.LambdaAWSInfo, error)` constructor
    - The existing `sparta.HandleAWSLambda` function is deprecated and will be removed in a subsequent release
  - Added _Sparta/archetype/rest_ package to streamline REST-based Sparta services.

    - This package includes a fluent builder (`MethodHandler`) and constructor function (`RegisterResource`) that transforms a _rest.Resource_ implementing struct into an API Gateway resource.
    - Usage:

      ```go
      // File: resource.go
      // TodoItemResource is the /todo/{id} resource
      type TodoItemResource struct {
      }
      // ResourceDefinition returns the Sparta REST definition for the Todo item
      func (svc *TodoItemResource) ResourceDefinition() (spartaREST.ResourceDefinition, error) {

        return spartaREST.ResourceDefinition{
          URL: todoItemURL,
          MethodHandlers: spartaREST.MethodHandlerMap{
            ...
          }
        }, nil
      }

      // File: main.go
      func() {
        myResource := &TodoItemResource{}
        resourceMap, resourcesErr := spartaREST.RegisterResource(apiGatewayInstance, myResource)
      }
      ```

    - Sample fluent method builder:

      ```go
        // GET
        http.MethodGet: spartaREST.NewMethodHandler(svc.Get, http.StatusOK).
          StatusCodes(http.StatusInternalServerError).
          Privileges(svc.S3Accessor.KeysPrivilege("s3:GetObject"),
                      svc.S3Accessor.BucketPrivilege("s3:ListBucket")),
      ```

    - See [SpartaTodoBackend](https://github.com/mweagle/SpartaTodoBackend) for a complete example
      - The _SpartaTodoBackend_ is a self-deploying CORS-accessible service that satisfies the [TodoBackend](https://www.todobackend.com/) online tests

  - Added _Sparta/aws/accessor_ package to streamline S3-backed service creation.
    - Embed a `services.S3Accessor` type to enable utility methods for:
      - `Put`
      - `Get`
      - `GetAll`
      - `Delete`
      - `DeleteAll`
  - Added [prealloc](https://github.com/alexkohler/prealloc) check to ensure that slices are preallocated when possible

- :bug: **FIXED**
  - Fix latent issue where CloudWatch Log ARN was malformed [commit](https://github.com/mweagle/Sparta/commit/5800553983ed16e6c5e4a622559909c050c00219)

## v1.5.0 - The Observability Edition 🔭

- :warning: **BREAKING**
- :checkered_flag: **CHANGES**

  - Expose `sparta.InstanceID()` that returns a random instance identifier for a single Lambda container instance
    - The _instanceID_ field is also included in the [ContextLogger](https://godoc.org/github.com/mweagle/Sparta#pkg-constants)
  - Add a self-monitoring function that publishes container-level metrics to CloudWatch.

    - Usage:

      ```go
        import spartaCloudWatch "github.com/mweagle/Sparta/aws/cloudwatch"
        func main() {
          ...
          spartaCloudWatch.RegisterLambdaUtilizationMetricPublisher(map[string]string{
            "BuildId":    sparta.StampedBuildID,
          })
          ...
        }
      ```

    - The optional `map[string]string` parameter is the custom Name-Value pairs to use as a [CloudWatch Dimension](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/cloudwatch_concepts.html#Dimension)
    - <div align="center"><img src="https://raw.githubusercontent.com/mweagle/Sparta/master/docs_source/static/site/1.5.0/CloudWatch_Management_Console.jpg" />

  - Add `WorkflowHooks.Validators` to support policy-based validation of the materialized template.
    - Each validator receives a complete read-only copy of the template
  - Add [magefile](https://magefile.org/) actions in _github.com/mweagle/Sparta/magefile_ to support cross platform scripting.

    - A Sparta service can use a standard _magefile.go_ as in:

      ```go
      // +build mage

      package main

      import (
        spartaMage "github.com/mweagle/Sparta/magefile"
      )

      // Provision the service
      func Provision() error {
        return spartaMage.Provision()
      }

      // Describe the stack by producing an HTML representation of the CloudFormation
      // template
      func Describe() error {
        return spartaMage.Describe()
      }

      // Delete the service, iff it exists
      func Delete() error {
        return spartaMage.Delete()
      }

      // Status report if the stack has been provisioned
      func Status() error {
        return spartaMage.Status()
      }

      // Version information
      func Version() error {
        return spartaMage.Version()
      }
      ```

    which exposes the most common Sparta command line options.

    - Usage: `mage status`:

      ```plain
      $ mage status
      INFO[0000] ════════════════════════════════════════════════
      INFO[0000] ╔═╗╔═╗╔═╗╦═╗╔╦╗╔═╗   Version : 1.5.0
      INFO[0000] ╚═╗╠═╝╠═╣╠╦╝ ║ ╠═╣   SHA     : 8f199e1
      INFO[0000] ╚═╝╩  ╩ ╩╩╚═ ╩ ╩ ╩   Go      : go1.11.1
      INFO[0000] ════════════════════════════════════════════════
      INFO[0000] Service: MyHelloWorldStack-mweagle            LinkFlags= Option=status UTC="2018-10-20T04:46:57Z"
      INFO[0000] ════════════════════════════════════════════════
      INFO[0001] StackId                                       Id="arn:aws:cloudformation:us-west-2:************:stack/MyHelloWorldStack-mweagle/5817dff0-c5f1-11e8-b43a-503ac9841a99"
      INFO[0001] Stack status                                  State=UPDATE_COMPLETE
      INFO[0001] Created                                       Time="2018-10-02 03:14:59.127 +0000 UTC"
      INFO[0001] Last Update                                   Time="2018-10-19 03:23:00.048 +0000 UTC"
      INFO[0001] Tag                                           io:gosparta:buildId=7ee3e1bc52f15c4a636e05061eaec7b748db22a9
      ```

- :bug: **FIXED**
  - Fix latent issue where multiple [archetype](https://godoc.org/github.com/mweagle/Sparta/archetype) handlers of the same type would collide.

## v1.4.0

- :warning: **BREAKING**
  - Moved `sparta.LambdaVersioningDecorator` to `decorator.LambdaVersioningDecorator`
  - Updated [cloudformation.ConvergeStackState](https://godoc.org/github.com/mweagle/Sparta/aws/cloudformation#ConvergeStackState) to accept a timeout parameter
  - Updated [ServiceDecorator.DecorateService](https://godoc.org/github.com/mweagle/Sparta#ServiceDecoratorHookFunc.DecorateService) to accept the S3Key parameter
    - This allows `ServiceDecorators` to add their own Lambda-backed CloudFormation custom resources and have them instantiated at AWS Lambda runtime. (eg: CloudFormation [Lambda-backed custom resources](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/template-custom-resources-lambda.html) ). See next section for more information.
- :checkered_flag: **CHANGES**

  - Simplified CustomResource creation and dispatch logic

    - The benefit of this is that users can define new `CustomResourceCommand` implementing CustomResources and have them roundtripped and instantiated at AWS Lambda execution time. 🎉
    - I'll write up more documentation, but the steps to defining your own Lambda-backed custom resource:

      1. Create a resource that embeds [gocf.CloudFormationCustomResource](https://godoc.org/github.com/mweagle/go-cloudformation#CloudFormationCustomResource) and your custom event properties:

         ```go
           type HelloWorldResourceRequest struct {
             Message *gocf.StringExpr
           }
           type HelloWorldResource struct {
             gocf.CloudFormationCustomResource
             HelloWorldResourceRequest
           }
         ```

      1. Register the custom resource provider with [RegisterCustomResourceProvider](https://godoc.org/github.com/mweagle/go-cloudformation#RegisterCustomResourceProvider)
      1. Implement [CustomResourceCommand](https://godoc.org/github.com/mweagle/Sparta/aws/cloudformation/resources#CustomResourceCommand)

    - At provisioning time, an instance of your CustomResource will be created and the appropriate functions will be called with the incoming [CloudFormationLambdaEvent](https://godoc.org/github.com/mweagle/Sparta/aws/cloudformation/resources#CloudFormationLambdaEvent).
      - Unmarshal the `event.ResourceProperties` map into your command handler instance and perform the requested operation.

  - Added a set of `archetype.*` convenience functions to create `sparta.LambdaAWSInfo` for specific event types.

    - The `archetype.*` package exposes creation functions to simplify common lambda types. Sample S3 _Reactor_ handler:

      ```go
        func echoS3Event(ctx context.Context, s3Event awsLambdaEvents.S3Event) (interface{}, error) {
          // Respond to s3:ObjectCreated:*", "s3:ObjectRemoved:*" S3 events
        }
        func main() {
          lambdaFn, _ := spartaArchetype.NewS3Reactor(spartaArchetype.S3ReactorFunc(echoS3Event),
            gocf.String("MY_S3_BUCKET"),
            nil)
            // ...
        }
      ```

  - Added `--nocolor` command line option to suppress colorized output. Default value: `false`.
  - When a service `provision` fails, only report resources that failed to succeed.
    - Previously, resources that were cancelled due to other resource failures were also logged as _ERROR_ statements.
  - Added `decorator.CloudWatchErrorAlarmDecorator(...)` to create per-Lambda CloudWatch Alarms.

    - Sample usage:

      ```go
        lambdaFn.Decorators = []sparta.TemplateDecoratorHandler{
          spartaDecorators.CloudWatchErrorAlarmDecorator(1, // Number of periods
            1, // Number of minutes per period
            1, // GreaterThanOrEqualToThreshold value
            gocf.String("SNS_TOPIC_ARN_OR_RESOURCE_REF")),
        }
      ```

  - Added `decorator.NewLogAggregatorDecorator` which forwards all CloudWatch log messages to a Kinesis stream.
    - See [SpartaPProf](https://github.com/mweagle/SpartaPProf) for an example of forwarding CloudWatch log messages to Google StackDriver
  - Added [decorator.CloudFrontSiteDistributionDecorator](https://godoc.org/github.com/mweagle/Sparta/decorator#CloudFrontSiteDistributionDecorator) to provision a CloudFront distribution with a custom Route53 name and optional SSL support.

    - Sample usage:

      ```go
      func distroHooks(s3Site *sparta.S3Site) *sparta.WorkflowHooks {
        hooks := &sparta.WorkflowHooks{}
        siteHookDecorator := spartaDecorators.CloudFrontSiteDistributionDecorator(s3Site,
          "subdomainNameHere",
          "myAWSHostedZone.com",
          "arn:aws:acm:us-east-1:OPTIONAL-ACM-CERTIFICATE-FOR-SSL")
        hooks.ServiceDecorators = []sparta.ServiceDecoratorHookHandler{
          siteHookDecorator,
        }
        return hooks
      }
      ```

    - Supply the `WorkflowHooks` struct to `MainEx` to annotate your service with an example CloudFront distribution. Note that CF distributions introduce a significant provisioning delay.
    - See [SpartaHTML](https://github.com/mweagle/SpartaHTML) for more

  - Added `decorator.S3ArtifactPublisherDecorator` to publish an arbitrary JSON file to an S3 location
    - This is implemented as Sparta-backed CustomResource
  - Added `status` command to produce a report of a provisioned service. Sample usage:

    ```bash
    $ go run main.go status --redact
    INFO[0000] ════════════════════════════════════════════════
    INFO[0000] ╔═╗╔═╗╔═╗╦═╗╔╦╗╔═╗   Version : 1.4.0
    INFO[0000] ╚═╗╠═╝╠═╣╠╦╝ ║ ╠═╣   SHA     : 3681d28
    INFO[0000] ╚═╝╩  ╩ ╩╩╚═ ╩ ╩ ╩   Go      : go1.11.1
    INFO[0000] ════════════════════════════════════════════════
    INFO[0000] Service: SpartaPProf-mweagle                  LinkFlags= Option=status UTC="2018-10-05T12:24:57Z"
    INFO[0000] ════════════════════════════════════════════════
    INFO[0000] StackId                                       Id="arn:aws:cloudformation:us-west-2:************:stack/SpartaPProf-mweagle/da781540-c764-11e8-9bf1-0aceeffcea3c"
    INFO[0000] Stack status                                  State=CREATE_COMPLETE
    INFO[0000] Created                                       Time="2018-10-03 23:34:21.142 +0000 UTC"
    INFO[0000] Tag                                           io:gosparta:buildTags=googlepprof
    INFO[0000] Tag                                           io:gosparta:buildId=c3fbe8c289c3184efec842dca56b9bf541f39d21
    INFO[0000] Output                                        HelloWorldFunctionARN="arn:aws:lambda:us-west-2:************:function:SpartaPProf-mweagle_Hello_World"
    INFO[0000] Output                                        KinesisLogConsumerFunctionARN="arn:aws:lambda:us-west-2:************:function:SpartaPProf-mweagle_KinesisLogConsumer"
    ```

  - Replaced _Makefile_ with [magefile](https://magefile.org/) to better support cross platform builds.

    - This is an internal only change and does not impact users
    - For **CONTRIBUTORS**, to use the new _mage_ targets:

      ```plain
      $> go get -u github.com/magefile/mage
      $> mage -l

      Targets:
        build                           the application
        clean                           the working directory
        describe                        runs the `TestDescribe` test to generate a describe HTML output file at graph.html
        ensureAllPreconditions          ensures that the source passes *ALL* static `ensure*` precondition steps
        ensureFormatted                 ensures that the source code is formatted with goimports
        ensureLint                      ensures that the source is `golint`ed
        ensureSpelling                  ensures that there are no misspellings in the source
        ensureStaticChecks              ensures that the source code passes static code checks
        ensureTravisBuildEnvironment    is the command that sets up the Travis environment to run the build.
        ensureVet                       ensures that the source has been `go vet`ted
        generateBuildInfo               creates the automatic buildinfo.go file so that we can stamp the SHA into the binaries we build...
        generateConstants               runs the set of commands that update the embedded CONSTANTS for both local and AWS Lambda execution
        installBuildRequirements        installs or updates the dependent packages that aren't referenced by the source, but are needed to build the Sparta source
        publish                         the latest source
        test                            runs the Sparta tests
        testCover                       runs the test and opens up the resulting report
        travisBuild                     is the task to build in the context of a Travis CI pipeline
      ```

  - Added [misspell](https://github.com/client9/misspell) static check as part of `mage test` to catch misspellings

- :bug: **FIXED**

## v1.3.0

- :warning: **BREAKING**
- :checkered_flag: **CHANGES**
  - Update branchname and release tag to support Go 1.11 [modules](https://github.com/golang/go/wiki/Modules).
- :bug: **FIXED**
  - Fixed `panic` when extracting [lambda function name](https://github.com/mweagle/Sparta/commit/c10a7a88c403ecf5b1f06784f0027fb35e0220a7).

## v1.2.1

- :warning: **BREAKING**
- :checkered_flag: **CHANGES**
  - Added `decorator.LogAggregatorDecorator`
    - This is a decorator that:
      1. Creates a [CloudWatchLogs Subscription Filter](https://t.co/C0cbo99Tsr) for the Lambda functions
      1. Creates a Kinesis sink with the user defined shard count to receive the log events.
      1. Subscribes the relay lambda function to the Kinesis stream
      1. See [SpartaPProf](https://github.com/mweagle/SpartaPProf) for an example that relays log entries to Google StackDriver.
  - Added `decorator.PublishAttOutputDecorator` and `decorator.PublishRefOutputDecorator` as convenience functions to update the Stack [Outputs](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/outputs-section-structure.html) section.
  - Added `RuntimeLoggerHook` to [WorkflowHooks](https://godoc.org/github.com/mweagle/Sparta#WorkflowHooks) to support logrus logger [hooks](https://github.com/sirupsen/logrus#hooks).
  - Added `IsExecutingInLambda () bool` to return execution environment
- :bug: **FIXED**
  - [`$GOPATH` is no longer present by default](https://github.com/mweagle/Sparta/issues/111)
  - [`gas` was replaced by `gosec`](https://github.com/mweagle/Sparta/issues/112)
  - [`tview.ANSIIWriter` has been renamed to `tview.ANSIWriter`](https://github.com/mweagle/Sparta/issues/110)

## v1.2

- :warning: **BREAKING**
- :checkered_flag: **CHANGES**

  - Added support for SQS event triggers.

    - SQS event sources use the same [EventSourceMappings](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-lambda-eventsourcemapping.html) entry that is used by DynamoDB and Kinesis. For example:

      ```go
      lambdaFn.EventSourceMappings = append(lambdaFn.EventSourceMappings,
          &sparta.EventSourceMapping{
            EventSourceArn: gocf.GetAtt(sqsResourceName, "Arn"),
            BatchSize:      2,
          })
      ```

      - Where `sqsResourceName` is the name of a CloudFormation resource provisioned by the stack
      - Use the [aws.SQSEvent](https://godoc.org/github.com/aws/aws-lambda-go/events#SQSEvent) value type as the incoming message

    - See the [SpartaSQS](https://github.com/mweagle/SpartaSQS) project for a complete example

  - Migrated `describe` command to use [Cytoscape.JS](http://js.cytoscape.org/) library
    - Cytoscape supports several layout algorithms and per-service node icons.
  - Added `APIGatewayEnvelope` type to allow struct embedding and overriding of the `Body` field. Example:

    ```go
    // FeedbackBody is the typed body submitted in a FeedbackRequest
    type FeedbackBody struct {
      Language string `json:"lang"`
      Comment  string `json:"comment"`
    }

    // FeedbackRequest is the typed input to the
    // onFeedbackDetectSentiment
    type FeedbackRequest struct {
      spartaEvents.APIGatewayEnvelope
      Body FeedbackBody `json:"body"`
    }
    ```

  - The previous [APIGatewayRequest](https://godoc.org/github.com/mweagle/Sparta/aws/events#APIGatewayRequest) remains unchanged:

    ```go
    type APIGatewayRequest struct {
      APIGatewayEnvelope
      Body interface{} `json:"body"`
    }
    ```

- :bug: **FIXED**
  - Fixed latent bug where dynamically created DynamoDB and Kinesis Event Source mappings had insufficient IAM privileges
  - Fixed latent bug where the [S3Site](https://godoc.org/github.com/mweagle/Sparta#S3Site) source directory was validated before `go:generate` could have been executed. This resulted in cases where fresh-cloned repositories would not self-deploy.
    - The filepath existence requirement was moved further into the provision workflow to support inline JS build operations.

## v1.1.1

- :warning: **BREAKING**
- :checkered_flag: **CHANGES**
  - Re-implemented the `explore` command.
    - The `explore` command provides a terminal-based UI to interactively submit events to provisioned Lambda functions.
    - The set of JSON files are determined by walking the working directory for all _\*.json_ files
    - _Example_: <div align="center"><img src="https://raw.githubusercontent.com/mweagle/Sparta/master/docs_source/static/site/1.1.1/explore.jpg" />
  - Eliminate redundant `Statement` entries in [AssumeRolePolicyDocument](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-iam-role.html)
  - Add `sparta.StampedBuildID` global variable to access the _BuildID_ value (either user defined or automatically generated)
  - Added `-z/--timestamps` command line flag to optionally include UTC timestamp prefix on every log line.
  - Prefer `git rev-parse HEAD` value for fallback BuildID value iff `--buildID` isn't provided as a _provision_ command line argument. If an error is detected calling `git`, the previous randomly initialized buffer behavior is used.
- :bug: **FIXED**

## v1.1.0

- :warning: **BREAKING**
  - Removed `lambdabinary` build tags from [BuildDockerImage](https://godoc.org/github.com/mweagle/Sparta/docker#BuildDockerImage)
    - AWS native support for **Go** in AWS caused a significant difference in standard vs `lambdabinary` build targets executed which prevented custom application options from being respected.
- :checkered_flag: **CHANGES**

  - Change [EventSourceMapping.EventSourceArn](https://godoc.org/github.com/mweagle/Sparta#EventSourceMapping) from string to `interface{}` type.

    - This change was to allow for provisioning of Pull-based event sources being provisioned in the same Sparta application as the lambda definition.
    - For example, to reference a DynamoDB Stream created by in a [ServiceDecoratorHook](https://godoc.org/github.com/mweagle/Sparta#ServiceDecoratorHook) for the _myDynamoDBResourceName_ resource you can now use:

    ```go
    lambdaFn.EventSourceMappings = append(lambdaFn.EventSourceMappings,
      &sparta.EventSourceMapping{
        EventSourceArn:   gocf.GetAtt(myDynamoDBResourceName, "StreamArn"),
        StartingPosition: "TRIM_HORIZON",
        BatchSize:        10,
      })
    ```

  - Updated `describe` output format and upgraded to latest versions of static HTML assets.
    - _Example_: <div align="center"><img src="https://raw.githubusercontent.com/mweagle/Sparta/master/docs_source/static/site/1.1.0/describe.jpg" />
      </div>
  - Delegate CloudFormation template aggregation to [go-cloudcondenser](https://github.com/mweagle/go-cloudcondenser)
  - Exposed [ReservedConcurrentExecutions](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-lambda-function.html#cfn-lambda-function-reservedconcurrentexecutions) option for Lambda functions.
  - Exposed [DeadLetterConfigArn](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-lambda-function.html#cfn-lambda-function-deadletterconfig) property to support custom DLQ destinations.
  - Added IAM `sparta.IAMRolePrivilege` fluent builder type in the _github.com/mweagle/Sparta/aws/iam/builder_. Sample usage:

    ```go
    iambuilder.Allow("ssm:GetParameter").ForResource().
      Literal("arn:aws:ssm:").
      Region(":").
      AccountID(":").
      Literal("parameter/MyReservedParameter").
      ToPrivilege()
    ```

  - Remove _io:gosparta:home_ and _io:gosparta:sha_ Tags from Lambda functions
  - Standardize on Lambda function naming in AWS Console
  - Reduced AWS Go binary size by 20% or more by including the `-s` and `-w` [link flags](https://golang.org/cmd/link/)
    - See [Shrink your Go Binaries with this One Weird Trick](https://blog.filippo.io/shrink-your-go-binaries-with-this-one-weird-trick/) for more information
  - Added `github.com/mweagle/Sparta/aws/cloudformation.UserAccountScopedStackName` to produce CloudFormation Stack names that are namespaced by AWS account username
  - Ensure `Pre` and `Post` deploy hooks are granted proper permissions
    - See [SpartaSafeDeploy](https://github.com/mweagle/SpartaSafeDeploy) for more information.
  - Added [Sparta/aws/apigateway.Error](https://godoc.org/github.com/mweagle/Sparta/aws/apigateway#Error) to support returning custom API Gateway errors
    - See [SpartaHTML](https://github.com/mweagle/SpartaHTML) for example usage
  - API Gateway `error` responses are now converted to JSON objects via a Body Mapping template:

    ```go
    "application/json": "$input.path('$.errorMessage')",
    ```

    - See the [AWS docs](https://docs.aws.amazon.com/apigateway/latest/developerguide/handle-errors-in-lambda-integration.html) for more information

  - Added check for Linux only package [sysinfo](github.com/zcalusic/sysinfo). This Linux-only package is ignored by `go get` because of build tags and cannot be safely imported. An error will be shown if the package cannot be found:

    ```plain
    ERRO[0000] Failed to validate preconditions: Please run
    `go get -v github.com/zcalusic/sysinfo` to install this Linux-only package.
    This package is used when cross-compiling your AWS Lambda binary and cannot
    be safely imported across platforms. When you `go get` the package, you may
    see errors as in `undefined: syscall.Utsname`. These are expected and can be
    ignored
    ```

  - Added additional build-time static analysis check for suspicious coding practices with [gas](https://github.com/GoASTScanner/gas)

- :bug: **FIXED**
  - [101](https://github.com/mweagle/Sparta/issues/101)
  - Fixed latent bug where `NewAuthorizedMethod` didn't properly preserve the [AuthorizerID](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-apigateway-method.html#cfn-apigateway-method-authorizationtype) when serializing to CloudFormation. This also forced a change to the function signature to accept a `gocf.Stringable` satisfying type for the [authorizerID](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-apigateway-authorizer.html).

## v1.0.1

- :warning: **BREAKING**
- :checkered_flag: **CHANGES**
  - Added [events](https://github.com/mweagle/Sparta/blob/master/aws/events/event.go) package for Sparta specific event types.
    - Initial top level event is `APIGatewayRequest` type for responding to API-Gateway mediated requests.
  - Prefer stamping `buildID` into binary rather than providing as environment variable. Previously the stamped buildID was the `env.SPARTA_BUILD_ID` mutable variable.
  - Remove dependency on [go-validator](github.com/asaskevich/govalidator)
- :bug: **FIXED**
  - Fixed latent bug where [Discovery](https://godoc.org/github.com/mweagle/Sparta#Discover) wasn't properly initialized in AWS Lambda execution context
  - Fixed latent bug where [CommandLineOptions](https://github.com/mweagle/Sparta/blob/master/sparta_main.go#L72) weren't properly defined in AWS build target
    - Affected [SpartaCodePipeline](https://github.com/mweagle/SpartaCodePipeline) project

## v1.0.0

## 🎉 AWS Lambda for Go Support 🎉

- Sparta Go function signature has been changed to **ONLY** support the official AWS Lambda Go signatures

  - `func ()`
  - `func () error`
  - `func (TIn) error`
  - `func () (TOut, error)`
  - `func (context.Context) error`
  - `func (context.Context, TIn) error`
  - `func (context.Context) (TOut, error)`
  - `func (context.Context, TIn) (TOut, error)`

- See the lambda.Start [docs](https://godoc.org/github.com/aws/aws-lambda-go/lambda#Start) or the related [AWS Blog Post](https://aws.amazon.com/blogs/compute/announcing-go-support-for-aws-lambda/) for more information.
- _ALL_ Sparta Go Lambda function targets **MUST** now use the `sparta.HandleAWSLambda` creation function, a function pointer that satisfies one of the supported signatures.
- Providing an invalid signature such as `func() string` will produce a `provision` time error as in:

  ```plain
  Error: Invalid lambda returns: Hello World. Error: handler returns a single value, but it does not implement error
  ```

- :warning: **BREAKING**

  - Removed `sparta.NewLambda` constructor
  - Removed `sparta.NewServeMuxLambda` proxying function
  - Removed `sparta.LambdaFunction` type
  - `ContextKeyLambdaContext` is no longer published into the context. Prefer the official AWS [FromContext()](https://godoc.org/github.com/aws/aws-lambda-go/lambdacontext#LambdaContext) function to access the AWS Go Lambda context.
  - Moved [DashboardDecorator](https://github.com/mweagle/SpartaXRay) to `decorators` namespace
  - Removed `explore` command line option as proxying tier is no longer supported
  - Changed all `logrus` imports to proper [lowercase format](https://github.com/sirupsen/logrus#logrus-)

- :checkered_flag: **CHANGES**

  - All decorators are now implemented as slices.
    - Existing single-valued fields remain supported, but deprecated
    - There are convenience types to adapt free functions to their `*Handler` interface versions:
      - `TemplateDecoratorHookFunc`
      - `WorkflowHookFunc`
      - `ArchiveHookFunc`
      - `ServiceDecoratorHookFunc`
      - `RollbackHookFunc`
  - Added `CodeDeployServiceUpdateDecorator` to support [safe AWS Lambda deploys](https://github.com/awslabs/serverless-application-model/blob/master/docs/safe_lambda_deployments.rst)
    - Safe lambda deploys are implemented via [ServiceDecoratorHooks](https://godoc.org/github.com/mweagle/Sparta#WorkflowHooks)
    - See [SpartaSafeDeploy](https://github.com/mweagle/SpartaSafeDeploy) for a complete example
  - Added **requestID** and **lambdaARN** request-scoped [\*logrus.Entry](https://godoc.org/github.com/sirupsen/logrus#Entry) to `context` argument.

    - This can be accessed as in:

    ```go
      contextLogger, contextLoggerOk := ctx.Value(sparta.ContextKeyRequestLogger).(*logrus.Entry)
      if contextLoggerOk {
        contextLogger.Info("Request scoped log")
      }
    ```

    - The existing `*logrus.Logger` entry is also available in the `context` via:

    ```go
      logger, loggerOk := ctx.Value(sparta.ContextKeyLogger).(*logrus.Logger)
    ```

  - [NewMethod](https://godoc.org/github.com/mweagle/Sparta#Resource.NewMethod) now accepts variadic parameters to limit how many API Gateway integration mappings are defined
  - Added `SupportedRequestContentTypes` to [NewMethod](https://godoc.org/github.com/mweagle/Sparta#Resource.NewMethod) to limit API Gateway generated content.
  - Added `apiGateway.CORSOptions` field to configure _CORS_ settings
  - Added `Add S3Site.CloudFormationS3ResourceName()`

    - This value can be used to scope _CORS_ access to a dynamoc S3 website as in:

    ```go
    apiGateway.CORSOptions = &sparta.CORSOptions{
      Headers: map[string]interface{}{
        "Access-Control-Allow-Origin":  gocf.GetAtt(s3Site.CloudFormationS3ResourceName(),
        "WebsiteURL"),
      }
    }
    ```

    - Improved CLI usability in consistency of named outputs, formatting.

  - :bug: **FIXED**
    - Fix latent bug where `provision` would not consistently create new [API Gateway Stage](https://docs.aws.amazon.com/apigateway/latest/developerguide/stages.html) events.

## v0.30.1

- :warning: **BREAKING**
- :checkered_flag: **CHANGES**
  - Improved API-Gateway CORS support. The following customizations are opt-in:
    - Parameterize CORS headers returned by _OPTIONS_ via [API.CORSOptions](https://godoc.org/github.com/mweagle/Sparta#API)
    - Add `SupportedRequestContentTypes` to [Method](https://godoc.org/github.com/mweagle/Sparta#Method) struct. This is a slice of supported content types that define what API-Gateway _Content-Type_ values are supported. Limiting the set of supported content types reduces CloudFormation template size.
    - Add variadic `possibleHTTPStatusCodeResponses` values to [NewMethod](https://godoc.org/github.com/mweagle/Sparta#Resource.NewMethod). If defined, Sparta will ONLY generate [IntegrationResponse](https://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-integration-settings-integration-response.html) entries for the possible codes (including the default HTTP status code). The previous, and default behavior, is to generate IntegrationResponse entries for _ALL_ valid HTTP status codes.
  - Include per-resource CloudFormation provisioning times in output log
  - Humanize magnitude output values and times with [go-humanize](https://github.com/dustin/go-humanize)
  - Replace CloudFormation polling log output with [spinner](https://github.com/briandowns/spinner)
    - This feedback is only available in normal CLI output. JSON formatted output remains unchanged.
  - Usability improvements for Windows based builds
- :bug: **FIXED**
  - Re-enable `cloudformation:DescribeStacks` and `cloudformation:DescribeStackResource` privileges to support HTML based deployments

## v0.30.0

- :warning: **BREAKING**
  - `Tags` for dependent resources no longer available via [sparta.Discover](https://godoc.org/github.com/mweagle/Sparta#Discover)
  - Remove public sparta `Tag*` constants that were previously reserved for Discover support.
- :checkered_flag: **CHANGES**
  - Change [sparta.Discover](https://godoc.org/github.com/mweagle/Sparta#Discover) to use _Environment_ data rather than CloudFormation API calls.
  - See [SpartaDynamoDB](https://github.com/mweagle/SpartaDynamoDB) for sample usage of multiple lambda functions depending on a single, dynamically provisioned Dynamo table.
  - Include **BuildID** in Lambda environment via `SPARTA_BUILD_ID` environment variable.
- :bug: **FIXED**
  - Correct CLI typo

## v0.20.4

- :warning: **BREAKING**
  - Changed `step.NewStateMachine` signature to include _StateMachineName_ as first argument per [Nov 15th, 2017 release](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/ReleaseHistory.html)
- :checkered_flag: **CHANGES**

  - Add `profile` command

    - Profile snapshots are enabled via:

    ```go
    sparta.ScheduleProfileLoop(nil, 5*time.Second, 30*time.Second, "heap")
    ```

    - Profile snapshots are published to S3 and are locally aggregated across all lambda instance publishers. To view the ui, run the `profile` Sparta command.
      - For more information, please see [The new pprof user interface - ⭐️](https://rakyll.org/pprof-ui/), [Profiling Go programs with pprof](https://jvns.ca/blog/2017/09/24/profiling-go-with-pprof/), or the [Go blog](https://blog.golang.org/profiling-go-programs)
    - See the [SpartaPProf](https://github.com/mweagle/SpartaPProf) sample for a service that installs profiling hooks.
    - Ensure you have the latest `pprof` UI via _go get -u -v github.com/google/pprof_
    - The standard [profile names](https://golang.org/pkg/runtime/pprof/#Profile) are available, as well as a _cpu_ type implied by a non-zero `time.Duration` supplied as the third parameter to `ScheduleProfileLoop`.

  - Eliminate unnecessary logging in AWS lambda environment
  - Log NodeJS [process.uptime()](https://nodejs.org/api/process.html#process_process_uptime)

- :bug: **FIXED**
  - Added more constructive message when working directory for `go build` doesn't contain `main` package.

## v0.20.3

- :warning: **BREAKING**
- :checkered_flag: **CHANGES**
- :bug: **FIXED**
  - Fixed `explore` interactive debugging instructions

## v0.20.2

- :warning: **BREAKING**
- :checkered_flag: **CHANGES**

  - Added support for [Step functions](https://aws.amazon.com/step-functions/faqs/).
    - Step functions are expressed via a combination of: states, `NewStateMachine`, and adding a `StateMachineDecorator` as a [service hook](https://godoc.org/github.com/mweagle/Sparta#ServiceDecoratorHook).
    - See the [SpartaStep](https://github.com/mweagle/SpartaStep) sample for a service that provisions a simple roll die state machine.
  - Usability improvements & enhancements for CLI log output. Text-formatted output now includes cleaner header as in:

    ```plain
    INFO[0000] ══════════════════════════════════════════════════════════════
    INFO[0000]    _______  ___   ___  _________
    INFO[0000]   / __/ _ \/ _ | / _ \/_  __/ _ |     Version : 0.20.2
    INFO[0000]  _\ \/ ___/ __ |/ , _/ / / / __ |     SHA     : 740028b
    INFO[0000] /___/_/  /_/ |_/_/|_| /_/ /_/ |_|     Go      : go1.9.1
    INFO[0000]
    INFO[0000] ══════════════════════════════════════════════════════════════
    INFO[0000] Service: SpartaStep-mweagle                   LinkFlags= Option=provision UTC="2017-11-01T01:14:31Z"
    INFO[0000] ══════════════════════════════════════════════════════════════
    ```

  - Added [megacheck](https://github.com/dominikh/go-tools/tree/master/cmd/megacheck) to compile pipeline. Fixed issues.
  - Corrected inline Go examples to use proper function references & signatures.

- :bug: **FIXED**
  - Handle case where multipart forms with empty values weren't handled [https://github.com/mweagle/Sparta/issues/74](https://github.com/mweagle/Sparta/issues/74)

## v0.20.1

- :warning: **BREAKING**
- :checkered_flag: **CHANGES**
  - Add `sparta.LambdaName` to return the reflection-discovered name of an `http.HandleFunc` instance.
- :bug: **FIXED**
  - Fixed issue with `--describe` not rendering CloudFormation template properly
  - Better handle failures when [posting body](https://github.com/mweagle/Sparta/pull/72) - thanks [@nylar](https://github.com/nylar)

## v0.20.0

### :star: Deprecation Notice

The `sparta.LambdaFunc` signature is officially deprecated in favor of `http.HandlerFunc` and will be removed in an upcoming release. See below for more information

- :warning: **BREAKING**
  - Changed `NewLambdaHTTPHandler` to `NewServeMuxLambda`
  - Remove obsolete `InvokeID` from [LambdaContext](https://godoc.org/github.com/mweagle/Sparta#LambdaContext)
  - Changed `codePipelineTrigger` CLI arg name to `codePipelinePackage`
- :checkered_flag: **CHANGES**

  - Eliminated NodeJS cold start `cp & chmod` penalty! :fire:
    - Prior to this release, the NodeJS proxying code would copy the embedded binary to _/tmp_ and add the executable flag prior to actually launching the binary. This had a noticeable performance penalty for startup.
    - This release embeds the application or library in a _./bin_ directory with the file permissions set so that there is no additional filesystem overhead on cold-start. h/t to [StackOverflow](https://stackoverflow.com/questions/41651134/cant-run-binary-from-within-python-aws-lambda-function) for the tips.
  - Migrated all IPC calls to [protocolBuffers](https://developers.google.com/protocol-buffers/).
    - Message definitions are in the [proxy](https://github.com/mweagle/Sparta/tree/master/proxy) directory.
  - The client-side log level (eg: `--level debug`) is carried into the AWS Lambda Code package.
    - Provisioning a service with `--level debug` will log everything at `logger.Debug` level and higher **including all AWS API** calls made both at `provision` and Lambda execution time.
    - Help resolve "Works on My Stack" syndrome.
  - HTTP handler `panic` events are now recovered and the traceback logged for both NodeJS and `cgo` deployments
  - Introduced `sparta.HandleAWSLambda`

    - `sparta.HandleAWSLambda` accepts standard `http.RequestFunc` signatures as in:

      ```go
      func helloWorld(w http.ResponseWriter, r *http.Request) {
        ...
      }

      lambdaFn := sparta.HandleAWSLambda("Hello HTTP World",
        http.HandlerFunc(helloWorld),
        sparta.IAMRoleDefinition{})
      ```

    - This allows you to [chain middleware](https://github.com/justinas/alice) for a lambda function as if it were a standard HTTP handler. Say, for instance: [X-Ray](https://github.com/aws/aws-xray-sdk-go).
    - The legacy [sparta.LambdaFunction](https://godoc.org/github.com/mweagle/Sparta#LambdaFunction) is still supported, but marked for deprecation. You will see a log warning as in:

      ```plain
      WARN[0045] DEPRECATED: sparta.LambdaFunc() signature provided. Please migrate to http.HandlerFunc()
      ```

    - _LambdaContext_ and _\*logrus.Logger_ are now available in the [requext.Context()](https://golang.org/pkg/net/http/#Request.Context) via:
      - `sparta.ContextKeyLogger` => `*logrus.Logger`
      - `sparta.ContextKeyLambdaContext` => `*sparta.LambdaContext`
    - Example:
      - `loggerVal, loggerValOK := r.Context().Value(sparta.ContextKeyLogger).(*logrus.Logger)`

  - Added support for [CodePipeline](https://aws.amazon.com/about-aws/whats-new/2016/11/aws-codepipeline-introduces-aws-cloudformation-deployment-action/)
    - See the [SpartaCodePipeline](https://github.com/mweagle/SpartaCodePipeline) project for a complete example and the related [post](https://medium.com/@mweagle/serverless-serverfull-and-weaving-pipelines-c9f83eec9227).
  - Upgraded NodeJS to [nodejs6.10](http://docs.aws.amazon.com/lambda/latest/dg/API_CreateFunction.html#SSS-CreateFunction-request-Runtime) runtime
  - Parity between NodeJS and Python/`cgo` startup output
  - Both NodeJS and `cgo` based Sparta applications now log equivalent system information.

    - Example:

      ```json
      {
        "level": "info",
        "msg": "SystemInfo",
        "systemInfo": {
          "sysinfo": {
            "version": "0.9.1",
            "timestamp": "2017-09-16T17:07:34.491807588Z"
          },
          "node": {
            "hostname": "ip-10-25-51-97",
            "machineid": "0046d1358d2346adbf8851e664b30d25",
            "hypervisor": "xenhvm",
            "timezone": "UTC"
          },
          "os": {
            "name": "Amazon Linux AMI 2017.03",
            "vendor": "amzn",
            "version": "2017.03",
            "architecture": "amd64"
          },
          "kernel": {
            "release": "4.9.43-17.38.amzn1.x86_64",
            "version": "#1 SMP Thu Aug 17 00:20:39 UTC 2017",
            "architecture": "x86_64"
          },
          "product": {},
          "board": {},
          "chassis": {},
          "bios": {},
          "cpu": {
            "vendor": "GenuineIntel",
            "model": "Intel(R) Xeon(R) CPU E5-2680 v2 @ 2.80GHz",
            "cache": 25600,
            "threads": 2
          },
          "memory": {}
        },
        "time": "2017-09-16T17:07:34Z"
      }
      ```

- :bug: **FIXED**
  - There were more than a few

## v0.13.2

- :warning: **BREAKING**
- :checkered_flag: **CHANGES**
  - Changed how Lambda [FunctionName](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-lambda-function.html#cfn-lambda-function-functionname) values are defined so that function name uniqueness is preserved for free, imported free, and struct-defined functions
- :bug: **FIXED**

## v0.13.1

- :warning: **BREAKING**
- :checkered_flag: **CHANGES**
  - Changed how Lambda [FunctionName](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-lambda-function.html#cfn-lambda-function-functionname) values are defined so that same-named functions provisioned across multiple stacks remain unique. This is done by prefixing the function name with the CloudFormation StackName.
  - Cleaned up S3 upload log statements to prefer relative paths iff applicable
- :bug: **FIXED**
  - [Cloudformation lambda function name validation error](https://github.com/mweagle/Sparta/issues/63)
  - [64](https://github.com/mweagle/Sparta/issues/64)

## v0.13.0

- :warning: **BREAKING**
  - Removed `sparta.NewNamedLambda`. Stable, user-defined function names can be supplied via the [SpartaOptions.Name](https://godoc.org/github.com/mweagle/Sparta#SpartaOptions) field.
- :checkered_flag: **CHANGES**

  - [CloudWatch Dashboard Support!](http://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/CloudWatch_Dashboards.html)

    - You can provision a CloudWatch dashboard that provides a single overview and link portal for your Lambda-based service. Use the new `sparta.DashboardDecorator` function to automatically create a dashboard. This leverages the existing [WorkflowHooks](https://godoc.org/github.com/mweagle/Sparta#WorkflowHooks) functionality.
    - Example:

    ```go
    // Setup the DashboardDecorator lambda hook
    workflowHooks := &sparta.WorkflowHooks{
      ServiceDecorator: sparta.DashboardDecorator(lambdaFunctions, 60),
    }
    ```

    - Where the `60` value is the CloudWatch time series [period](http://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/cloudwatch_concepts.html).
    - The CloudWatch Dashboard URL will be included in your stack's Outputs as in:

    ```plain
    INFO[0064] Stack output                                  Description="CloudWatch Dashboard URL" Key=CloudWatchDashboardURL Value="https://us-west-2.console.aws.amazon.com/cloudwatch/home?region=us-west-2#dashboards:name=SpartaXRay-mweagle"
    ```

    - _Example_: <div align="center"><img src="https://raw.githubusercontent.com/mweagle/Sparta/master/docs_source/static/site/0.13.0/dashboard.jpg" />

      </div>

    - For more info, see the [AWS Blog Post](https://aws.amazon.com/blogs/aws/new-api-cloudformation-support-for-amazon-cloudwatch-dashboards/)
    - The [SpartaXRay](https://github.com/mweagle/SpartaXRay) sample application has additional code samples.

  - [XRay](http://docs.aws.amazon.com/xray/latest/devguide/xray-services-lambda.html) support added
    - added `LambdaFunctionOptions.TracingConfig` field to [LambdaFunctionOptions](https://godoc.org/github.com/mweagle/Sparta#LambdaFunctionOptions)
    - added XRay IAM privileges to default IAM role settings:
      - _xray:PutTraceSegments_
      - _xray:PutTelemetryRecords_
    - See [AWS blog](https://aws.amazon.com/blogs/aws/aws-lambda-support-for-aws-x-ray/) for more information
  - added [LambdaFunctionOptions.Tags](https://godoc.org/github.com/mweagle/Sparta#LambdaFunctionOptions) to support tagging AWS Lambda functions
  - added _SpartaGitHash_ output to both CLI and CloudWatch Dashboard output. This is in addition to the _SpartaVersion_ value (which I occasionally have failed to update).

- :bug: **FIXED**
  - Fixed latent issue where `SpartaOptions.Name` field wasn't consistently used for function names.

## v0.12.1

- :warning: **BREAKING**
- :checkered_flag: **CHANGES**
  - added _Sparta/aws/cloudformation.UserScopedStackName()_ to generate username-suffixed CloudFormation StackNames
- :bug: **FIXED**

## v0.12.0

- :warning: **BREAKING**
  - Replaced all [https://github.com/crewjam/go-cloudformation](https://github.com/crewjam/go-cloudformation) references with [https://github.com/mweagle/go-cloudformation](https://github.com/mweagle/go-cloudformation) references
    - This is mostly internal facing, but impacts advanced usage via [ServiceDecoratorHook](https://godoc.org/github.com/mweagle/Sparta#ServiceDecoratorHook) users. Clients may
      need to update the types used to create [alternative topologies](http://gosparta.io/docs/alternative_topologies/).
- :checkered_flag: **CHANGES**
- :bug: **FIXED**
  - Fixed latent issue where CGO-enabled services that reference `cgo.NewSession()` would not build properly
  - Fixed latent issue where S3 backed sites (eg: [SpartaHugo](https://github.com/mweagle/SpartaHugo)) would not refresh on update.
  - [55](https://github.com/mweagle/Sparta/issues/55)

## v0.11.2

- :warning: **BREAKING**
- :checkered_flag: **CHANGES**

  - Added `--inplace/-c` command line option to support safe, concurrent updating of Lambda code packages

    - If enabled _AND_ the stack update [changeset](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-changesets.html) reports _only_ modifications to Lambda functions, then Sparta will use the AWS Lambda API to [update the function code](http://docs.aws.amazon.com/sdk-for-go/api/service/lambda/#Lambda.UpdateFunctionCode).
    - If enabled _AND_ additional mutations are reported, you'll see an error as in:

    ```plain
    ERRO[0022] Unsupported in-place operations detected:
      Add for IAMRole9fd267df3a3d0a144ae11a64c7fb9b7ffff3fb6c (ResourceType: AWS::IAM::Role),
      Add for mainhelloWorld2Lambda32fcf388f6b20e86feb93e990fa8decc5b3f9095 (ResourceType: AWS::Lambda::Function)
    ```

  - Prefer [NewRecorder](https://golang.org/pkg/net/http/httptest/#NewRecorder) to internal type for CGO marshalling
  - Added `--format/-f` command line flag `[text, txt, json]` to specify logfile output format. Default is `text`.
    - See [logrus.Formatters](https://github.com/sirupsen/logrus#formatters)

- :bug: **FIXED**
  - [45](https://github.com/mweagle/Sparta/issues/45)

## v0.11.1

- :warning: **BREAKING**
- :checkered_flag: **CHANGES**

  - Support Go 1.8 newly optional _GOPATH_ environment variable
  - Python proxied `cgo` builds now preserve the transformed source in the _./sparta_ scratch space directory.
  - Sparta assigned AWS Lambda function names now strip the leading SCM prefix. Example:

  ```bash
  github.com/mweagle/SpartaPython.HelloWorld
  ```

  becomes:

  ```bash
  mweagle/SpartaPython.HelloWorld
  ```

  - Upgrade to Mermaid [7.0.0](https://github.com/knsv/mermaid/releases/tag/7.0.0)
  - Use stable _PolicyName_ in `IAM::Role` definitions to minimize CloudFormation resource update churn

- :bug: **FIXED**
  - Fixed latent bug where S3 bucket version check didn't respect `--noop` mode.
  - Fixed latent `cgo` bug where command line arguments weren't properly parsed

## v0.11.0

- :warning: **BREAKING**
- :checkered_flag: **CHANGES**
  - :tada: Python CGO support added. See the [https://github.com/mweagle/SpartaPython](https://github.com/mweagle/SpartaPython) project for example usage!
    - In preliminary testing, the Python CGO package provides significant cold start and hot-execution performance benefits.
  - Migrated dependency management to [dep](https://github.com/golang/dep)
- :bug: **FIXED**
  - Fixed latent bug where DynamoDB EventSource mappings ResourceARNs weren't properly serialized.
  - Fixed latent bug where code pushed to S3 version-enabled buckets didn't use the latest `VersionID` in the AWS [Lambda Code](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-lambda-function-code.html) value.

## v0.10.0

- :warning: **BREAKING**
  - `describe` option now requires `-b/--s3Bucket` argument
  - Changed signature of `aws/s3/CreateS3RollbackFunc` to accept full S3 URL, including `versionId` query param
  - Signatures for `sparta.Provision` and `sparta.Discover` updated with new arguments
- :checkered_flag: **CHANGES**

  - Add `-p/--codePipelineTrigger` command line option to generate CodePipeline deployment package
  - Add `sparta.RegisterCodePipelineEnvironment` to define environment variables in support of [CloudFormation Deployments](https://aws.amazon.com/about-aws/whats-new/2016/11/aws-codepipeline-introduces-aws-cloudformation-deployment-action/). Example:

  ```go
  func init() {
    sparta.RegisterCodePipelineEnvironment("test", map[string]string{
      "MESSAGE": "Hello Test!",
    })
    sparta.RegisterCodePipelineEnvironment("production", map[string]string{
      "MESSAGE": "Hello Production!",
    })
  }
  ```

  - Add support for `Environment` and `KmsKeyArn` properties to [LambdaFunctionOptions](https://godoc.org/github.com/mweagle/Sparta#LambdaFunctionOptions). See [AWS](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-lambda-function.html) for more information.
  - Move all build artifacts to _./sparta_ directory
  - `-n/--noop` argument orphans S3 artifacts in _./sparta_ directory
  - Add support for S3 version policy enabled buckets
    - Artifacts pushed to S3 version-enabled buckets now use stable object keys. Rollback functions target specific versions if available.
  - Cleanup log statements
  - Add `sparta/aws/session.NewSessionWithLevel()` to support [AWS LogLevel](http://docs.aws.amazon.com/sdk-for-go/api/aws/#LogLevelType) parameter

- :bug: **FIXED**
  - [34](https://github.com/mweagle/Sparta/issues/34)
  - [37](https://github.com/mweagle/Sparta/issues/37)
  - [38](https://github.com/mweagle/Sparta/issues/38)

## v0.9.3

- :warning: **BREAKING**
- :checkered_flag: **CHANGES**

  - Added [LambdaFunctionOptions.SpartaOptions](https://godoc.org/github.com/mweagle/Sparta#SpartaOptions) struct
    - The primary use case is to support programmatically generated lambda functions that must be disambiguated by their Sparta name. Sparta defaults to reflection based function name identification.
  - Added `--ldflags` support to support lightweight [dynamic string variables](https://golang.org/cmd/link/)
    - Usage:
      `go run main.go provision --level info --s3Bucket $(S3_BUCKET) --ldflags "-X main.dynamicValue=SampleValue"`

- :bug: **FIXED**

## v0.9.2

- :warning: **BREAKING**
- :checkered_flag: **CHANGES**
  - Move Sparta-related provisioning values from stack [Outputs](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/outputs-section-structure.html) to [Tags](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/cfn-console-add-tags.html).
  - Add support for go [BuildTags](https://golang.org/pkg/go/build/) to support environment settings.
  - Added [Sparta/aws/cloudformation](https://godoc.org/github.com/mweagle/Sparta/aws/cloudformation) functions to support stack creation.
  - Added [Sparta/aws/s3](https://godoc.org/github.com/mweagle/Sparta/aws/cloudformation) functions to encapsulate common S3 operations.
  - Added [Sparta/zip](https://godoc.org/github.com/mweagle/Sparta/zip) functions to expose common ZIP related functions.
  - Legibility enhancements for `describe` output
  - `sparta.CloudFormationResourceName` proxies to `github.com/mweagle/Sparta/aws/cloudformation.CloudFormationResourceName`. The `sparta` package function is _deprecated_ and will be removed in a subsequent release.
- :bug: **FIXED**
  - Fixed latent bug in `github.com/mweagle/Sparta/zip.AddToZip` where the supplied ZipWriter was incorrectly closed on function exit.
  - Fixed latent parsing _userdata_ input
  - Fixed latent issue where empty [ChangeSets](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-changesets-execute.html) were applied rather than deleted.

## v0.9.1

- :warning: **BREAKING**
- :checkered_flag: **CHANGES**
  - Improved `describe` output. Includes APIGateway resources and more consistent UI.
  - Additive changes to [WorkflowHooks](https://godoc.org/github.com/mweagle/Sparta#WorkflowHooks)
    - `Context` property to set the initial context for Workflow hook execution
    - [ServiceDecorator](https://godoc.org/github.com/mweagle/Sparta#ServiceDecorator) type to define service-scoped AWS resources. Previously, template decoration was bound to specific Lambda functions.
  - Published related [SpartaVault](https://github.com/mweagle/SpartaVault): use AWS KMS to encrypt secrets as Go variables. See the [KMS Docs](http://docs.aws.amazon.com/kms/latest/developerguide/workflow.html) for information.
  - Add Godeps support
- :bug: **FIXED**
  - Fixed latent bug when adding custom resources to the ZIP archive via [ArchiveHook](https://godoc.org/github.com/mweagle/Sparta#ArchiveHook). ArchiveHook is now called after core Sparta assets are injected into archive.

## v0.9.0

- :warning: **BREAKING**

  - `NewMethod` and `NewAuthorizedMethod` for APIGateway definitions have been changed to include new, final parameter that marks the _default_ integration response code.

    - Prior to this change, Sparta would automatically use `http.StatusOK` for all non-POST requests, and `http.StatusCreated` for POST requests. The change allows you to control whitelisted headers to be returned through APIGateway as in:

    ```go
    // API response struct
    type helloWorldResponse struct {
      Location string `json:"location"`
      Body     string `json:"body"`
    }
    //
    // Promote the location key value to an HTTP header
    //
    apiGWMethod, _ := apiGatewayResource.NewMethod("GET", http.StatusOK)
    apiGWMethod.Responses[http.StatusOK].Parameters = map[string]bool{
      "method.response.header.Location": true,
    }
    apiGWMethod.Integration.Responses[http.StatusOK].Parameters["method.response.header.Location"] = "integration.response.body.location"
    ```

- :checkered_flag: **CHANGES**

  - (@sdbeard) Added [sparta.NewNamedLambda](https://godoc.org/github.com/mweagle/Sparta#NewNamedLambda) that allows you to set stable AWS Lambda [FunctionNames](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-lambda-function.html#cfn-lambda-function-functionname)
  - Added [spartaCF.AddAutoIncrementingLambdaVersionResource](https://godoc.org/github.com/mweagle/Sparta/aws/cloudformation#AddAutoIncrementingLambdaVersionResource) to support Lambda function versions. Should be called from a TemplateDecorator. Usage:

    ```go
    autoIncrementingInfo, autoIncrementingInfoErr := spartaCF.AddAutoIncrementingLambdaVersionResource(serviceName,
      lambdaResourceName,
      cfTemplate,
      logger)
    if nil != autoIncrementingInfoErr {
      return autoIncrementingInfoErr
    }
    ```

  - Added new [CloudWatch Metrics](http://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/CW_Support_For_AWS.html.html#cfn-lambda-function-functionname) for lambda execution
  - Removed all NodeJS shim `dependencies` from _./resources/provision/package.json_
  - Added utility CloudFormation script _./aws/cloudformation/cli/describe.go_ which produces a JSON serialization of a [DescribeStacksOutput](https://godoc.org/github.com/aws/aws-sdk-go/service/cloudformation#DescribeStacksOutput) struct for build-time discovery of cluster-scoped resources.
  - Relaxed constraint that an API GW resource is bound to single Sparta lambda function. You can now register per-HTTP method name lambda functions for the same API GW resource.
  - Added [Contributors](https://github.com/mweagle/Sparta#contributors) section to README

- :bug: **FIXED**
  - [19](https://github.com/mweagle/Sparta/issues/19)
  - [15](https://github.com/mweagle/Sparta/issues/15)
  - [16](https://github.com/mweagle/Sparta/issues/16)

## v0.8.0

- :warning: **BREAKING**
  - `TemplateDecorator` signature changed to include `context map[string]interface{}` to support sharing state across `WorkflowHooks` (below).
- :checkered_flag: **CHANGES**
  - Add `SpartaBuildID` stack output with build ID
  - `WorkflowHooks`
    - WorkflowHooks enable an application to customize the ZIP archive used as the AWS Lambda target rather than needing to embed resources inside their Go binary
    - They may also be used for Docker-based mixed topologies. See
  - Add optional `-i/--buildID` parameter for `provision`.
    - The parameter will be added to the stack outputs
    - A random value will be used if non is provided on the command line
  - Artifacts posted to S3 are now scoped by `serviceName`
  - Add `sparta.MainEx` for non-breaking signature extension
- :bug: **FIXED**

  - (@sdbeard) Fixed latent bug in Kinesis event source subscriptions that caused `ValidationError`s during provisioning:

    ```bash
    ERRO[0028] ValidationError: [/Resources/IAMRole3dbc1b4199ad659e6267d25cfd8cc63b4124530d/Type/Policies/0/PolicyDocument/Statement/5/Resource] 'null' values are not allowed in templates
        status code: 400, request id: ed5fae8e-7103-11e6-8d13-b943b498f5a2
    ```

  - Fixed latent bug in [ConvertToTemplateExpression](https://godoc.org/github.com/mweagle/Sparta/aws/cloudformation#ConvertToTemplateExpression) when parsing input with multiple AWS JSON fragments.
  - Fixed latent bug in [sparta.Discover](https://godoc.org/github.com/mweagle/Sparta#Discover) which prevented dependent resources from being discovered at Lambda execution time.
  - Fixed latent bug in [explore.NewAPIGatewayRequest](https://godoc.org/github.com/mweagle/Sparta/explore#NewAPIGatewayRequest) where whitelisted param keynames were unmarshalled to `method.request.TYPE.VALUE` rather than `TYPE`.

## v0.7.1

- :warning: **BREAKING**
- :checkered_flag: **CHANGES**
  - Upgrade to latest [go-cloudformation](https://github.com/crewjam/go-cloudformation) that required internal [refactoring](https://github.com/mweagle/Sparta/pull/9).
- :bug: **FIXED**
  - N/A

## v0.7.0

- :warning: **BREAKING**
  - `TemplateDecorator` signature changed to include `serviceName`, `S3Bucket`, and `S3Key` to allow for decorating CloudFormation with [UserData](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/user-data.html) to support [alternative topology](http://gosparta.io/docs/alternative_topologies/) deployments.
  - `CommonIAMStatements` changed from `map[string][]iamPolicyStatement` to struct with named fields.
  - `PushSourceConfigurationActions` changed from `map[string][]string` to struct with named fields.
  - Eliminated [goptions](https://github.com/voxelbrain/goptions)
- :checkered_flag: **CHANGES**
  - Moved CLI parsing to [Cobra](https://github.com/spf13/cobra)
    - Applications can extend the set of flags for existing Sparta commands (eg, `provision` can include `--subnetIDs`) as well as add their own top level commands to the `CommandLineOptions` exported values. See [SpartaCICD](https://github.com/mweagle/SpartaCICD) for an example.
  - Added _Sparta/aws/cloudformation_ `ConvertToTemplateExpression` to convert string value into [Fn::Join](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/intrinsic-function-reference-join.html) compatible representation. Parses inline AWS references and supports user-defined [template](https://golang.org/pkg/text/template/) properties.
  - Added `sparta/aws/iam` _PolicyStatement_ type
  - Upgraded `describe` output to use [Mermaid 6.0.0](https://github.com/knsv/mermaid/releases/tag/6.0.0)
  - All [goreportcard](https://goreportcard.com/report/github.com/mweagle/Sparta) issues fixed.
- :bug: **FIXED**
  - Fixed latent VPC provisioning bug where VPC/Subnet IDs couldn't be provided to template serialization.

## v0.6.0

- :warning: **BREAKING**
  - `TemplateDecorator` signature changed to include `map[string]string` to allow for decorating CloudFormation resource metadata
- :checkered_flag: **CHANGES**
  - All NodeJS CustomResources moved to _go_
  - Add support for user-defined CloudFormation CustomResources via `LambdaAWSInfo.RequireCustomResource`
  - `DiscoveryInfo` struct now includes `TagLogicalResourceID` field with CloudFormation Resource ID of calling lambda function
- :bug: **FIXED**
  - N/A

## v0.5.5

This release includes a major internal refactoring to move the current set of NodeJS [Lambda-backed CloudFormation CustomResources](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/template-custom-resources-lambda.html) to Sparta Go functions. The two migrated CustomActions are:

- The S3 event source configuration
- Provisioning an S3-static site

Both are implemented using [cloudformationresources](https://github.com/mweagle/cloudformationresources). There are no changes to the calling code and no regressions are expected.

- :warning: **BREAKING**
  - APIGateway provisioning now only creates a single discovery file: _MANIFEST.json_ at the site root.
- :checkered_flag: **CHANGES**
  - VPC support! Added [LambdaFunctionVPCConfig](https://godoc.org/github.com/crewjam/go-cloudformation#LambdaFunctionVPCConfig) to [LambdaFunctionsOptions](https://godoc.org/github.com/mweagle/Sparta#LambdaFunctionOptions) struct.
  - Updated NodeJS runtime to [nodejs4.3](http://docs.aws.amazon.com/lambda/latest/dg/programming-model.html)
  - CloudFormation updates are now done via [Change Sets](https://aws.amazon.com/blogs/aws/new-change-sets-for-aws-cloudformation/), rather than [UpdateStack](http://docs.aws.amazon.com/sdk-for-go/api/service/cloudformation/CloudFormation.html#UpdateStack-instance_method).
  - APIGateway and CloudWatchEvents are now configured using [CloudFormation](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/ReleaseHistory.html). They were previously implemented using NodeJS CustomResources.
- :bug: **FIXED**
  - Fixed latent issue where `IAM::Role` resources didn't use stable CloudFormation resource names
  - Fixed latent issue where names & descriptions of Lambda functions weren't consistent
  - [1](https://github.com/mweagle/SpartaApplication/issues/1)

## v0.5.4

- :warning: **BREAKING**
  - N/A
- :checkered_flag: **CHANGES**
  - Run `go generate` as part of the _provision_ step
- :bug: **FIXED**
  - N/A

## v0.5.3

- :warning: **BREAKING**
  - N/A
- :checkered_flag: **CHANGES**
  - N/A
- :bug: **FIXED**
  - [6](https://github.com/mweagle/Sparta/issues/6)

## v0.5.2

- :warning: **BREAKING**
  - N/A
- :checkered_flag: **CHANGES**
  - Added [cloudwatchlogs.Event](https://godoc.org/github.com/mweagle/Sparta/aws/cloudwatchlogs#Event) to support unmarshaling CloudWatchLogs data

## v0.5.1

- :warning: **BREAKING**
  - N/A
- :checkered_flag: **CHANGES**
  - Added [LambdaAWSInfo.URLPath](https://godoc.org/github.com/mweagle/Sparta#LambdaAWSInfo.URLPath) to enable _localhost_ testing
    - See _explore_test.go_ for example
- :bug: **FIXED**
  - [8](https://github.com/mweagle/Sparta/issues/8)

## v0.5.0

- :warning: **BREAKING**
  - N/A
- :checkered_flag: **CHANGES**
  - Added [sparta.CloudWatchLogsPermission](https://godoc.org/github.com/mweagle/Sparta#CloudWatchLogsPermission) type to support lambda invocation in response to log events.
  - Fixed latent bug on Windows where temporary archives weren't properly deleted
  - The `GO15VENDOREXPERIMENT=1` environment variable for cross compilation is now inherited from the current session.
    - Sparta previously always added it to the environment variables during compilation.
  - Hooked AWS SDK logger so that Sparta `--level debug` log level includes AWS SDK status
    - Also include `debug` level message listing AWS SDK version for diagnostic info
  - Log output includes lambda deployment [package size](http://docs.aws.amazon.com/lambda/latest/dg/limits.html)

## v0.4.0

- :warning: **BREAKING**
  - Change `sparta.Discovery()` return type from `map[string]interface{}` to `sparta.DiscoveryInfo`.
    - This type provides first class access to service-scoped and `DependsOn`-related resource information
- :checkered_flag: **CHANGES**
  - N/A

## v0.3.0

- :warning: **BREAKING**
  - Enforce that a single **Go** function cannot be associated with more than 1 `sparta.LamddaAWSInfo` struct.
    - This was done so that `sparta.Discovery` can reliably use the enclosing **Go** function name for discovery.
  - Enforce that a non-nil `*sparta.API` value provided to `sparta.Main()` includes a non-empty set of resources and methods
- :checkered_flag: **CHANGES**
  type
  - This type can be used to enable [CloudWatch Events](https://aws.amazon.com/blogs/aws/new-cloudwatch-events-track-and-respond-to-changes-to-your-aws-resources/)
    - See the [SpartaApplication](https://github.com/mweagle/SpartaApplication/blob/master/application.go#L381) example app for a sample usage.
  - `sparta.Discovery` now returns the following CloudFormation [Pseudo Parameters](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/pseudo-parameter-reference.html):
    - _StackName_
    - _StackID_
    - _Region_
  - Upgrade to Mermaid [0.5.7](https://github.com/knsv/mermaid/releases/tag/0.5.7) to fix `describe` rendering failure on Chrome.

## v0.2.0

- :warning: **BREAKING**

  - Changed `NewRequest` to `NewLambdaRequest` to support mock API gateway requests being made in `explore` mode
  - `TemplateDecorator` signature changed to support [go-cloudformation](https://github.com/crewjam/go-cloudformation) representation of the CloudFormation JSON template.
    - /ht @crewjam for [go-cloudformation](https://github.com/crewjam/go-cloudformation)
  - Use `sparta.EventSourceMapping` rather than [aws.CreateEventSourceMappingInput](http://docs.aws.amazon.com/sdk-for-go/api/service/lambda.html#type-CreateEventSourceMappingInput) type for `LambdaAWSInfo.EventSourceMappings` slice
  - Add dependency on [crewjam/go-cloudformation](https://github.com/crewjam/go-cloudformation) for CloudFormation template creation
    - /ht @crewjam for the great library
  - CloudWatch log output no longer automatically uppercases all first order child key names.

- :checkered_flag: **CHANGES**

  - :boom: Add `LambdaAWSInfo.DependsOn` slice
    - Lambda functions can now declare explicit dependencies on resources added via a `TemplateDecorator` function
    - The `DependsOn` value should be the dependency's logical resource name. Eg, the value returned from `CloudFormationResourceName(...)`.
  - :boom: Add `sparta.Discovery()` function

    - To be called from a **Go** lambda function (Eg, `func echoEvent(*json.RawMessage, *LambdaContext, http.ResponseWriter, *logrus.Logger)`), it returns the Outputs (both [Fn::Att](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/intrinsic-function-reference-getatt.html) and [Ref](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/intrinsic-function-reference-ref.html) ) values of dynamically generated CloudFormation resources that are declared as explicit `DependsOn` of the current function.
    - Sample output return value:

      ```json
      {
        "SESMessageStoreBucketa622fdfda5789d596c08c79124f12b978b3da772": {
          "DomainName": "spartaapplication-sesmessagestorebucketa622fdfda5-1rhh9ckj38gt4.s3.amazonaws.com",
          "Ref": "spartaapplication-sesmessagestorebucketa622fdfda5-1rhh9ckj38gt4",
          "Tags": [
            {
              "Key": "sparta:logicalBucketName",
              "Value": "Special"
            }
          ],
          "Type": "AWS::S3::Bucket",
          "WebsiteURL": "http://spartaapplication-sesmessagestorebucketa622fdfda5-1rhh9ckj38gt4.s3-website-us-west-2.amazonaws.com"
        },
        "golangFunc": "main.echoSESEvent"
      }
      ```

    - See the [SES EventSource docs](http://gosparta.io/docs/eventsources/ses/) for more information.

  - Added `TS` (UTC TimeStamp) field to startup message
  - Improved stack provisioning performance
  - Fixed latent issue where CloudFormation template wasn't deleted from S3 on stack provisioning failure.
  - Refactor AWS runtime requirements into `lambdaBinary` build tag scope to support Windows builds.
  - Add `SESPermission` type to support triggering Lambda functions in response to inbound email
    - See _doc_sespermission_test.go_ for an example
    - Storing the message body to S3 is done by assigning the `MessageBodyStorage` field.
  - Add `NewAPIGatewayRequest` to support _localhost_ API Gateway mock requests

## v0.1.5

- :warning: **BREAKING**
  - N/A
- :checkered_flag: **CHANGES**
  - Add [S3 Object Expiration](http://docs.aws.amazon.com/AmazonS3/latest/dev/how-to-set-lifecycle-configuration-intro.html) warning message if the target bucket doesn't specify one.
  - Replace internal CloudFormation polling loop with [WaitUntilStackCreateComplete](https://godoc.org/github.com/aws/aws-sdk-go/service/cloudformation#CloudFormation.WaitUntilStackCreateComplete) and [WaitUntilStackUpdateComplete](https://godoc.org/github.com/aws/aws-sdk-go/service/cloudformation#CloudFormation.WaitUntilStackUpdateComplete)

## v0.1.4

- :warning: **BREAKING**
  - N/A
- :checkered_flag: **CHANGES**
  - Reduce deployed binary size by excluding Sparta embedded resources from deployed binary via build tags.

## v0.1.3

- :warning: **BREAKING**
  - API Gateway responses are only transformed into a standard format in the case of a go lambda function returning an HTTP status code >= 400
    - Previously all responses were wrapped which prevented integration with other services.
- :checkered_flag: **CHANGES**

  - Default [integration mappings](http://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-mapping-template-reference.html) now defined for:
    - _application/json_
    - _text/plain_
    - _application/x-www-form-urlencoded_
    - _multipart/form-data_
    - Depending on the content-type, the **Body** value of the incoming event will either be a `string` or a `json.RawMessage` type.
  - CloudWatch log files support spawned golang binary JSON formatted logfiles
  - CloudWatch log output includes environment. Sample:

    ```JSON
      {
          "AWS_SDK": "2.2.25",
          "NODE_JS": "v0.10.36",
          "OS": {
              "PLATFORM": "linux",
              "RELEASE": "3.14.48-33.39.amzn1.x86_64",
              "TYPE": "Linux",
              "UPTIME": 4755.330878024
          }
      }
    ```

## v0.1.2

- :warning: **BREAKING**
  - N/A
- :checkered_flag: **CHANGES**
  - Added `explore.NewRequest` to support _localhost_ testing of lambda functions.
    - Clients can supply optional **event** data similar to the AWS Console feature.
    - See [explore_test](https://github.com/mweagle/Sparta/blob/master/explore_test.go) for an example.

## v0.1.1

- :warning: **BREAKING**
  - `sparta.Main()` signature changed to accept optional `S3Site` pointer
- :checkered_flag: **CHANGES**

  - Updated `describe` CSS font styles to eliminate clipping
  - Support `{Ref: 'MyDynamicResource'}` for _SourceArn_ values. Example:

    ```javascript
    lambdaFn.Permissions = append(lambdaFn.Permissions, sparta.SNSPermission{
      BasePermission: sparta.BasePermission{
        SourceArn: sparta.ArbitraryJSONObject{"Ref": snsTopicName},
      },
    })
    ```

    - Where _snsTopicName_ is a CloudFormation resource name representing a resource added to the template via a [TemplateDecorator](https://godoc.org/github.com/mweagle/Sparta#TemplateDecorator).

  - Add CloudWatch metrics to help track [container reuse](https://aws.amazon.com/blogs/compute/container-reuse-in-lambda/).
    - Metrics are published to **Sparta/<SERVICE_NAME>** namespace.
    - MetricNames: `ProcessCreated`, `ProcessReused`, `ProcessTerminated`.

## v0.1.0

- :warning: **BREAKING**
  - `sparta.Main()` signature changed to accept optional `S3Site` pointer
- :checkered_flag: **CHANGES**
  - Added `S3Site` type and optional static resource provisioning as part of `provision`
    - See the [SpartaHTML](https://github.com/mweagle/SpartaHTML) application for a complete example
  - Added `API.CORSEnabled` option (defaults to _false_).
    - If defined, all APIGateway methods will have [CORS Enabled](http://docs.aws.amazon.com/apigateway/latest/developerguide/how-to-cors.html).
  - Update logging to use structured fields rather than variadic, concatenation
  - Reimplement `explore` command line option.
    - The `explore` command line option creates a _localhost_ server to which requests can be sent for testing. The POST request body **MUST** be _application/json_, with top level `event` and `context` keys for proper unmarshaling.
  - Expose NewLambdaHTTPHandler() which can be used to generate an _httptest_

## v0.0.7

- :warning: **BREAKING**
  - N/A
- :checkered_flag: **CHANGES**
  - Documentation moved to [gosparta.io](http://gosparta.io)
    compliant value for `go test` integration.
    - Add [context](http://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-mapping-template-reference.html) struct to APIGatewayLambdaJSONEvent
    - Default description based on _Go_ function name for AWS Lambda if none provided
    - Added [SNS Event](https://github.com/mweagle/Sparta/blob/master/aws/sns/events.go) types for unmarshaling
    - Added [DynamoDB Event](https://github.com/mweagle/Sparta/blob/master/aws/dynamodb/events.go) types for unmarshaling
    - Added [Kinesis Event](https://github.com/mweagle/Sparta/blob/master/aws/kinesis/events.go) types for unmarshaling
    - Fixed latent issue where `IAMRoleDefinition` CloudFormation names would collide if they had the same Permission set.
    - Remove _API Gateway_ view from `describe` if none is defined.

## v0.0.6

- :warning: **BREAKING**
  - Changed:
    - `type LambdaFunction func(*json.RawMessage, *LambdaContext, *http.ResponseWriter, *logrus.Logger)`
      - **TO**
    - `type LambdaFunction func(*json.RawMessage, *LambdaContext, http.ResponseWriter, *logrus.Logger)`
    - See also [FAQ: When should I use a pointer to an interface?](https://golang.org/doc/faq#pointer_to_interface).
- Add _.travis.yml_ for CI support.
- :checkered_flag: **CHANGES**
  - Added [LambdaAWSInfo.Decorator](https://github.com/mweagle/Sparta/blob/master/sparta.go#L603) field (type [TemplateDecorator](https://github.com/mweagle/Sparta/blob/master/sparta.go#L192) ). If defined, the template decorator will be called during CloudFormation template creation and enables a Sparta lambda function to annotate the CloudFormation template with additional Resources or Output entries.
    - See [TestDecorateProvision](https://github.com/mweagle/Sparta/blob/master/provision_test.go#L44) for an example.
  - Improved API Gateway `describe` output.
  - Added [method response](http://docs.aws.amazon.com/apigateway/api-reference/resource/method-response/) support.
    - The [DefaultMethodResponses](https://godoc.org/github.com/mweagle/Sparta#DefaultMethodResponses) map is used if [Method.Responses](https://godoc.org/github.com/mweagle/Sparta#Method) is empty (`len(Responses) <= 0`) at provision time.
    - The default response map defines `201` for _POST_ methods, and `200` for all other methods. An API Gateway method may only support a single 2XX status code.
  - Added [integration response](http://docs.aws.amazon.com/apigateway/api-reference/resource/integration-response/) support for to support HTTP status codes defined in [status.go](https://golang.org/src/net/http/status.go).
    - The [DefaultIntegrationResponses](https://godoc.org/github.com/mweagle/Sparta#DefaultIntegrationResponses) map is used if [Integration.Responses](https://godoc.org/github.com/mweagle/Sparta#Integration) is empty (`len(Responses) <= 0`) at provision time.
    - The mapping uses regular expressions based on the standard _golang_ [HTTP StatusText](https://golang.org/src/net/http/status.go) values.
  - Added `SpartaHome` and `SpartaVersion` template [outputs](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/outputs-section-structure.html).

## v0.0.5

- :warning: **BREAKING**
  - Changed `Sparta.Main()` signature to accept API pointer as fourth argument. Parameter is optional.
- :checkered_flag: **CHANGES**
  - Preliminary support for API Gateway provisioning
    - See API type for more information.
  - `describe` output includes:
    - Dynamically generated CloudFormation Template
    - API Gateway json
    - Lambda implementation of `CustomResources` for push source configuration promoted from inline [ZipFile](http://docs.aws.amazon.com/lambda/latest/dg/API_FunctionCode.html) JS code to external JS files that are proxied via _index.js_ exports.
    - [Fixed latent bug](https://github.com/mweagle/Sparta/commit/684b48eb0c2356ba332eee6054f4d57fc48e1419) where remote push source registrations were deleted during stack updates.

## v0.0.3

- :warning: **BREAKING**
  - Changed `LambdaEvent` type to `json.RawMessage`
  - Changed [AddPermissionInput](http://docs.aws.amazon.com/sdk-for-go/api/service/lambda.html#type-AddPermissionInput) type to _sparta_ types:
    - `LambdaPermission`
    - `S3Permission`
    - `SNSPermission`
- :checkered_flag: **CHANGES**
  - `sparta.NewLambda(...)` supports either `string` or `sparta.IAMRoleDefinition` types for the IAM role execution value
    - `sparta.IAMRoleDefinition` types implicitly create an [IAM::Role](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-iam-role.html) resource as part of the stack
    - `string` values refer to pre-existing IAM rolenames
  - `S3Permission` type
    - `S3Permission` types denotes an S3 [event source](http://docs.aws.amazon.com/lambda/latest/dg/intro-core-components.html#intro-core-components-event-sources) that should be automatically configured as part of the service definition.
    - S3's [LambdaConfiguration](http://docs.aws.amazon.com/sdk-for-go/api/service/s3.html#type-LambdaFunctionConfiguration) is managed by a [Lambda custom resource](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/template-custom-resources-lambda.html) dynamically generated as part of in the [CloudFormation template](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/template-custom-resources.html).
    - The subscription management resource is inline NodeJS code and leverages the [cfn-response](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/walkthrough-custom-resources-lambda-cross-stack-ref.html) module.
  - `SNSPermission` type
    - `SNSPermission` types denote an SNS topic that should should send events to the target Lambda function
    - An SNS Topic's [subscriber list](http://docs.aws.amazon.com/AWSJavaScriptSDK/latest/AWS/SNS.html#subscribe-property) is managed by a [Lambda custom resource](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/template-custom-resources-lambda.html) dynamically generated as part of in the [CloudFormation template](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/template-custom-resources.html).
  - The subscription management resource is inline NodeJS code and leverages the [cfn-response](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/walkthrough-custom-resources-lambda-cross-stack-ref.html) module.
  - `LambdaPermission` type
    - These denote Lambda Permissions whose event source subscriptions should **NOT** be managed by the service definition.
  - Improved `describe` output CSS and layout
    - Describe now includes push/pull Lambda event sources
  - Fixed latent bug where Lambda functions didn't have CloudFormation::Log privileges

## v0.0.2

- Update describe command to use [mermaid](https://github.com/knsv/mermaid) for resource dependency tree
  - Previously used [vis.js](http://visjs.org/#)

## v0.0.1

- Initial release
