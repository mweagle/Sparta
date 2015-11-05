## v0.0.3
  - :checkered_flag: **CHANGES**
    - `sparta.NewLambda(...)` supports either `string` and `sparta.IAMRoleDefinition` types for the IAM role execution value
      - `sparta.IAMRoleDefinition` types implicitly create an [IAM::Role](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-iam-role.html) resource as part of the stack
      - `string` values refer to pre-existing IAM rolenames
    - `S3Permission` type
      - `S3Permission` types denotes an S3 [event source](http://docs.aws.amazon.com/lambda/latest/dg/intro-core-components.html#intro-core-components-event-sources) that should be automatically configured as part of the service definition.
      - S3's [LambdaConfiguration](http://docs.aws.amazon.com/sdk-for-go/api/service/s3.html#type-LambdaFunctionConfiguration) is manabed by a [Lambda custom resource](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/template-custom-resources-lambda.html) dynamically generated as part of in the [CloudFormation template](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/template-custom-resources.html).
      - The subscription management resource is inline NodeJS code and leverages the [cfn-response](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/walkthrough-custom-resources-lambda-cross-stack-ref.html) module.
    - `SNSPermission` type
      - ``SNSPermission` types denote an SNS topic that should should send events to the target Lambda function
      - The subscription management resource is inline NodeJS code and leverages the [cfn-response](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/walkthrough-custom-resources-lambda-cross-stack-ref.html) module.
    - `LambdaPermission` type
      - These denote lambda permissions whose event source subscriptions should not be managed by the service definition.
    - Improved `describe` output CSS and layout
      - Describe now includes push/pull Lambda event sources
    - Fixed latent bug where Lambda functions didn't have CloudFormation::Log privileges
  - :warning: **BREAKING**
    - Changed `LambdaEvent` type to `interface{}`
    - Changed  [AddPermissionInput](http://docs.aws.amazon.com/sdk-for-go/api/service/lambda.html#type-AddPermissionInput) type to _sparta_ types:
      - `LambdaPermission`
      - `S3Permission`

## v0.0.2
  - Update describe command to use [mermaid](https://github.com/knsv/mermaid) for resource dependency tree
    - Previously used [vis.js](http://visjs.org/#)

## v0.0.1
  - Initial release