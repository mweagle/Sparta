## v0.1.4
- Reduce deployed binary size by excluding Sparta embedded resources from deployed binary via build tags.
- :warning: **BREAKING**
  - N/A

## v0.1.3
- :checkered_flag: **CHANGES**
  - Default [integration mappings](http://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-mapping-template-reference.html) now defined for:
    * _application/json_
    * _text/plain_
    * _application/x-www-form-urlencoded_
    * _multipart/form-data_
    - Depending on the content-type, the **Body** value of the incoming event will either be a `string` or a `json.RawMessage` type.
  - CloudWatch log files support spawned golang binary JSON formatted logfiles
  - CloudWatch log output includes environment.  Sample:

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
- :warning: **BREAKING**
  - API Gateway responses are only transformed into a standard format in the case of a go lambda function returning an HTTP status code >= 400
    - Previously all responses were wrapped which prevented integration with other services.

## v0.1.2
- :checkered_flag: **CHANGES**
  - Added `explore.NewRequest` to support _localhost_ testing of lambda functions.  
    - Clients can supply optional **event** data similar to the AWS Console feature.
    - See [explore_test](https://github.com/mweagle/Sparta/blob/master/explore_test.go) for an example.
- :warning: **BREAKING**
  - N/A

## v0.1.1
- :checkered_flag: **CHANGES**
  - Updated `describe` CSS font styles to eliminate clipping
  - Support `{Ref: 'MyDynamicResource'}` for _SourceArn_ values.  Example:

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
- :warning: **BREAKING**
  - `sparta.Main()` signature changed to accept optional `S3Site` pointer

## v0.1.0
- :checkered_flag: **CHANGES**
  - Added `S3Site` type and optional static resource provisioning as part of `provision`
    - See the [SpartaHTML](https://github.com/mweagle/SpartaHTML) application for a complete example
  - Added `API.CORSEnabled` option (defaults to _false_).  
    - If defined, all APIGateway methods will have [CORS Enabled](http://docs.aws.amazon.com/apigateway/latest/developerguide/how-to-cors.html).
  - Update logging to use structured fields rather than variadic, concatenation
  - Reimplement `explore` command line option.
    - The `explore` command line option creates a _localhost_ server to which requests can be sent for testing.  The POST request body **MUST** be _application/json_, with top level `event` and `context` keys for proper unmarshaling.
  - Expose NewLambdaHTTPHandler() which can be used to generate an _httptest_
- :warning: **BREAKING**
  - `sparta.Main()` signature changed to accept optional `S3Site` pointer

## v0.0.7
- :checkered_flag: **CHANGES**
  - Documentation moved to [gosparta.io](http://gosparta.io)
 compliant value for `go test` integration.
    - Add [context](http://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-mapping-template-reference.html) struct to APIGatewayLambdaJSONEvent
    - Default description based on *Go* function name for AWS Lambda if none provided
    - Added [SNS Event](https://github.com/mweagle/Sparta/blob/master/aws/sns/events.go) types for unmarshaling
    - Added [DynamoDB Event](https://github.com/mweagle/Sparta/blob/master/aws/dynamodb/events.go) types for unmarshaling
    - Added [Kinesis Event](https://github.com/mweagle/Sparta/blob/master/aws/kinesis/events.go) types for unmarshaling
    - Fixed latent issue where `IAMRoleDefinition` CloudFormation names would collide if they had the same Permission set.
    - Remove _API Gateway_ view from `describe` if none is defined.
- :warning: **BREAKING**
  - N/A

## v0.0.6
- Add _.travis.yml_ for CI support.
- :checkered_flag: **CHANGES**
    - Added [LambdaAWSInfo.Decorator](https://github.com/mweagle/Sparta/blob/master/sparta.go#L603) field (type [TemplateDecorator](https://github.com/mweagle/Sparta/blob/master/sparta.go#L192) ). If defined, the template decorator will be called during CloudFormation template creation and enables a Sparta lambda function to annotate the CloudFormation template with additional Resources or Output entries.
      - See [TestDecorateProvision](https://github.com/mweagle/Sparta/blob/master/provision_test.go#L44) for an example.
    - Improved API Gateway `describe` output.
    - Added [method response](http://docs.aws.amazon.com/apigateway/api-reference/resource/method-response/) support.  
      - The [DefaultMethodResponses](https://godoc.org/github.com/mweagle/Sparta#DefaultMethodResponses) map is used if [Method.Responses](https://godoc.org/github.com/mweagle/Sparta#Method) is empty  (`len(Responses) <= 0`) at provision time.
      - The default response map defines `201` for _POST_ methods, and `200` for all other methods. An API Gateway method may only support a single 2XX status code.
    - Added [integration response](http://docs.aws.amazon.com/apigateway/api-reference/resource/integration-response/) support for to support HTTP status codes defined in [status.go](https://golang.org/src/net/http/status.go).
      - The [DefaultIntegrationResponses](https://godoc.org/github.com/mweagle/Sparta#DefaultIntegrationResponses) map is used if [Integration.Responses](https://godoc.org/github.com/mweagle/Sparta#Integration) is empty  (`len(Responses) <= 0`) at provision time.
      - The mapping uses regular expressions based on the standard _golang_ [HTTP StatusText](https://golang.org/src/net/http/status.go) values.
    - Added `SpartaHome` and `SpartaVersion` template [outputs](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/outputs-section-structure.html).
- :warning: **BREAKING**
  - Changed:
    - `type LambdaFunction func(*json.RawMessage, *LambdaContext, *http.ResponseWriter, *logrus.Logger)`
      - **TO**
    - `type LambdaFunction func(*json.RawMessage, *LambdaContext, http.ResponseWriter, *logrus.Logger)`
    - See also [FAQ: When should I use a pointer to an interface?](https://golang.org/doc/faq#pointer_to_interface).

## v0.0.5
- :checkered_flag: **CHANGES**
  - Preliminary support for API Gateway provisioning
    - See API type for more information.
  - `describe` output includes:
    - Dynamically generated CloudFormation Template
    - API Gateway json
    - Lambda implementation of `CustomResources` for push source configuration promoted from inline [ZipFile](http://docs.aws.amazon.com/lambda/latest/dg/API_FunctionCode.html) JS code to external JS files that are proxied via _index.js_ exports.
    - [Fixed latent bug](https://github.com/mweagle/Sparta/commit/684b48eb0c2356ba332eee6054f4d57fc48e1419) where remote push source registrations were deleted during stack updates.
- :warning: **BREAKING**
  - Changed `Sparta.Main()` signature to accept API pointer as fourth argument.  Parameter is optional.

## v0.0.3
  - :checkered_flag: **CHANGES**
    - `sparta.NewLambda(...)` supports either `string` or `sparta.IAMRoleDefinition` types for the IAM role execution value
      - `sparta.IAMRoleDefinition` types implicitly create an [IAM::Role](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-iam-role.html) resource as part of the stack
      - `string` values refer to pre-existing IAM rolenames
    - `S3Permission` type
      - `S3Permission` types denotes an S3 [event source](http://docs.aws.amazon.com/lambda/latest/dg/intro-core-components.html#intro-core-components-event-sources) that should be automatically configured as part of the service definition.
      - S3's [LambdaConfiguration](http://docs.aws.amazon.com/sdk-for-go/api/service/s3.html#type-LambdaFunctionConfiguration) is managed by a [Lambda custom resource](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/template-custom-resources-lambda.html) dynamically generated as part of in the [CloudFormation template](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/template-custom-resources.html).
      - The subscription management resource is inline NodeJS code and leverages the [cfn-response](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/walkthrough-custom-resources-lambda-cross-stack-ref.html) module.
    - `SNSPermission` type
      - ``SNSPermission` types denote an SNS topic that should should send events to the target Lambda function
      - An SNS Topic's [subscriber list](http://docs.aws.amazon.com/AWSJavaScriptSDK/latest/AWS/SNS.html#subscribe-property) is managed by a [Lambda custom resource](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/template-custom-resources-lambda.html) dynamically generated as part of in the [CloudFormation template](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/template-custom-resources.html).
     - The subscription management resource is inline NodeJS code and leverages the [cfn-response](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/walkthrough-custom-resources-lambda-cross-stack-ref.html) module.
    - `LambdaPermission` type
      - These denote Lambda Permissions whose event source subscriptions should **NOT** be managed by the service definition.
    - Improved `describe` output CSS and layout
      - Describe now includes push/pull Lambda event sources
    - Fixed latent bug where Lambda functions didn't have CloudFormation::Log privileges
  - :warning: **BREAKING**
    - Changed `LambdaEvent` type to `json.RawMessage`
    - Changed  [AddPermissionInput](http://docs.aws.amazon.com/sdk-for-go/api/service/lambda.html#type-AddPermissionInput) type to _sparta_ types:
      - `LambdaPermission`
      - `S3Permission`
      - `SNSPermission`

## v0.0.2
  - Update describe command to use [mermaid](https://github.com/knsv/mermaid) for resource dependency tree
    - Previously used [vis.js](http://visjs.org/#)

## v0.0.1
  - Initial release
