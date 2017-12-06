---
date: 2016-03-09T19:56:50+01:00
title: FAQ
weight: 10
---

## Development
<hr />

### How do I setup AWS SDK credentials and region?

Sparta relies on standard AWS SDK configuration settings. See the [official documentation](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html) for more information.

During development, configuration is typically done through environment variables:

  - `AWS_ACCESS_KEY_ID`
  - `AWS_SECRET_ACCESS_KEY`
  - `AWS_REGION`

### What are the *Minimum* set of privileges needed for an account to use Sparta?

The absolute minimum set of privileges an account needs is the following [IAM Policy](https://awspolicygen.s3.amazonaws.com/policygen.html):

{{< highlight json >}}
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "Stmt1505975332000",
            "Effect": "Allow",
            "Action": [
                "cloudformation:DescribeStacks",
                "cloudformation:CreateStack",
                "cloudformation:CreateChangeSet",
                "cloudformation:DescribeChangeSet",
                "cloudformation:ExecuteChangeSet",
                "cloudformation:DeleteChangeSet",
                "cloudformation:DeleteStack",
                "iam:GetRole",
                "iam:DeleteRole",
                "iam:DeleteRolePolicy",
                "iam:PutRolePolicy"
            ],
            "Resource": [
                "*"
            ]
        },
        {
            "Sid": "Stmt1505975332000",
            "Effect": "Allow",
            "Action": [
                "s3:PutObject",
                "s3:GetBucketVersioning",
                "s3:DeleteObject"
            ],
            "Resource": [
                "arn:aws:s3:::PROVISION_TARGET_BUCKETNAME"
            ]
        }
    ]
}
{{< /highlight >}}

This set of privileges should be sufficient to deploy a Sparta application similar to [SpartaHelloWorld](https://github.com/mweagle/SpartaHelloWorld). Additional privileges may be required to enable different datasources.

You can view the exact set of AWS API calls by enabling `--level debug` log verbosity. This log level includes all AWS API calls starting with release [0.20.0](https://github.com/mweagle/Sparta/blob/master/CHANGES.md#v0200).

### What flags are defined during AWS AMI compilation?

* **TAGS**:         `-tags lambdabinary`
* **ENVIRONMENT**:  `GOOS=linux GOARCH=amd64`

### What working directory should I use?

Your working directory should be the root of your Sparta application.  Eg, use

{{< highlight go >}}
go run main.go provision --level info --s3Bucket $S3_BUCKET
{{< /highlight >}}

rather than

{{< highlight go >}}
go run ./some/child/path/main.go provision --level info --s3Bucket $S3_BUCKET
{{< /highlight >}}

See [GitHub](https://github.com/mweagle/Sparta/issues/29) for more details.

### How can I test locally?

Local testing is available via the [explore](/docs/local_testing/) command. You can use the standard Go [httptest](https://golang.org/pkg/net/http/httptest/) package for unit tests. See the [explore_test.go](https://github.com/mweagle/Sparta/blob/master/explore_test.go) source for an example.

### How can I make `provision` faster?

Starting with Sparta [v0.11.2](https://github.com/mweagle/Sparta/blob/master/CHANGES.md#v0112), you can supply an optional
_--inplace_ argument to the `provision` command. If this is set when provisioning updates to an existing stack,
your Sparta application will verify that the *only* updates to the CloudFormation stack are code-level updates. If
only code updates are detected, your Sparta application will parallelize [UpdateFunctionCode](http://docs.aws.amazon.com/sdk-for-go/api/service/lambda/#Lambda.UpdateFunctionCode) API calls directly to update the
application code.

Whether _--inplace_ is valid is based on evaluating the [ChangeSet](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-changesets.html) results of the
requested update operation.

*NOTE*: The _inplace_ argument implies that your service state is not reflected in CloudFormation.

## Event Sources - SES
<hr />

### Where does the _SpartaRuleSet_ come from?

SES only permits a single [active receipt rule](http://docs.aws.amazon.com/ses/latest/APIReference/API_SetActiveReceiptRuleSet.html).  Additionally, it's possible that multiple Sparta-based services are handing different SES recipients.

All Sparta-based services share the _SpartaRuleSet_ SES ruleset, and uniquely identify their Rules by including the current servicename as part of the SES [ReceiptRule](http://docs.aws.amazon.com/ses/latest/APIReference/API_CreateReceiptRule.html).

### Why does `provision` not always enable the _SpartaRuleSet_?

Initial _SpartaRuleSet_ will make it the active ruleset, but Sparta assumes that manual updates made outside of the context of the framework were done with good reason and doesn't attempt to override the user setting.

## Operations
<hr />

### How can I provision a service dashboard?

Sparta [v0.13.0](https://github.com/mweagle/Sparta/blob/master/CHANGES.md#v0130) adds support for the provisioning of a
CloudWatch Dashboard that's dynamically created based on your service's topology. The dashboard
is attached to the standard Sparta workflow via a [WorkflowHook](https://godoc.org/github.com/mweagle/Sparta#WorkflowHook) as in:

{{< highlight golang>}}
// Setup the DashboardDecorator lambda hook
workflowHooks := &sparta.WorkflowHooks{
	ServiceDecorator: sparta.DashboardDecorator(lambdaFunctions, 60),
}
{{< /highlight >}}

See the [SpartaXRay](https://github.com/mweagle/SpartaXRay) project for a complete example of provisioning a dashboard as below:

![CloudWatchDashboard](/images/faq/CloudWatchDashboard.jpg)

### How can I monitor my Lambda function?

If you plan on using your Lambdas in production, you'll probably want to be made aware of any excessive errors.

You can easily do this by adding a CloudWatch alarm to your Lambda, in the decorator method.

This example will push a notification to an SNS topic, and you can configure whatever action is appropriate from there.

{{< highlight go >}}
func lambdaDecorator(serviceName string,
	lambdaResourceName string,
	lambdaResource gocf.LambdaFunction,
	resourceMetadata map[string]interface{},
	S3Bucket string,
	S3Key string,
	buildID string,
	cfTemplate *gocf.Template,
	context map[string]interface{},
	logger *logrus.Logger) error {

	// setup CloudWatch alarm
	var alarmDimensions gocf.CloudWatchMetricDimensionList
	alarmDimension := gocf.CloudWatchMetricDimension{Name: gocf.String("FunctionName"), Value: gocf.Ref(lambdaResourceName).String()}
	alarmDimensions = []gocf.CloudWatchMetricDimension{alarmDimension}

	lambdaErrorsAlarm := &gocf.CloudWatchAlarm{
		ActionsEnabled:     gocf.Bool(true),
		AlarmActions:       gocf.StringList(gocf.String("arn:aws:sns:us-east-1:123456789:SNSToNotifyMe")),
		AlarmName:          gocf.String("LambdaErrorAlarm"),
		ComparisonOperator: gocf.String("GreaterThanOrEqualToThreshold"),
		Dimensions:         &alarmDimensions,
		EvaluationPeriods:  gocf.String("1"),
		Period:             gocf.String("300"),
		MetricName:         gocf.String("Errors"),
		Namespace:          gocf.String("AWS/Lambda"),
		Statistic:          gocf.String("Sum"),
		Threshold:          gocf.String("3.0"),
		Unit:               gocf.String("Count"),
	}
	cfTemplate.AddResource("LambdaErrorAlaram", lambdaErrorsAlarm)

	return nil
}
{{< /highlight >}}


### Where can I view my function's `*logger` output?

Each lambda function includes privileges to write to [CloudWatch Logs](https://console.aws.amazon.com/cloudwatch/home).  The `*logrus.logger` output is written (with a brief delay) to a lambda-specific log group.

The CloudWatch log group name includes a sanitized version of your **go** function name & owning service name.

### Where can I view Sparta's golang spawn metrics?

Visit the [CloudWatch Metrics](https://aws.amazon.com/cloudwatch/) AWS console page and select the `Sparta/{SERVICE_NAME}` namespace:

![CloudWatch](/images/faq/CloudWatch_Management_Console.jpg)

Sparta publishes two counters:

  * `ProcessSpawned`: A new **go** process was spawned to handle requests
  * `ProcessReused`: An existing **go** process was used to handle requests.  See also the discussion on AWS Lambda [container reuse](https://aws.amazon.com/blogs/compute/container-reuse-in-lambda/).

### How can I include additional AWS resources as part of my Sparta application?

Define a [TemplateDecorator](https://godoc.org/github.com/mweagle/Sparta#TemplateDecorator) function and annotate the `*gocf.Template` with additional AWS resources.

For more flexibility, use a [WorkflowHook](https://godoc.org/github.com/mweagle/Sparta#WorkflowHooks).

### How can I provide environment variables to lambda functions?

Sparta uses conditional compilation rather than environment variables. See [Managing Environments](/docs/application/environments/) for more information.

### Does Sparta support Versioning & Aliasing?

Yes.

Define a [TemplateDecorator](https://godoc.org/github.com/mweagle/Sparta#TemplateDecorator) function and annotate the `*gocf.Template` with an [AutoIncrementingLambdaVersionInfo](https://godoc.org/github.com/mweagle/Sparta/aws/cloudformation#AutoIncrementingLambdaVersionInfo) resource. During each `provision` operation, the `AutoIncrementingLambdaVersionInfo` resource will dynamically update the CloudFormation template with a new version.

{{< highlight go >}}
autoIncrementingInfo, autoIncrementingInfoErr := spartaCF.AddAutoIncrementingLambdaVersionResource(serviceName,
  lambdaResourceName,
  cfTemplate,
  logger)
{{< /highlight >}}

You can also move the "alias pointer" by referencing one or more of the versions available in the returned struct. For example, to set the alias pointer to the most recent version:

{{< highlight go >}}
// Add an alias to the version we're publishing as part of this `provision` operation
aliasResourceName := sparta.CloudFormationResourceName("Alias", lambdaResourceName)
aliasResource := &gocf.LambdaAlias{
    Name:            gocf.String("MostRecentVersion"),
    FunctionName:    gocf.Ref(lambdaResourceName).String(),
    FunctionVersion: gocf.GetAtt(autoIncrementingInfo.CurrentVersionResourceName, "Version").String(),
}
cfTemplate.AddResource(aliasResourceName, aliasResource)
{{< /highlight >}}

### How do I forward additional metrics?

Sparta-deployed AWS Lambda functions always operate with CloudWatch Metrics `putMetric` privileges.  Your lambda code can call `putMetric` with application-specific data.

### How do I setup alerts on additional metrics?

Define a [TemplateDecorator](https://godoc.org/github.com/mweagle/Sparta#TemplateDecorator) function and annotate the `*gocf.Template` with the needed [AWS::CloudWatch::Alarm](https://godoc.org/github.com/crewjam/go-cloudformation#CloudWatchAlarm) values.  Use [CloudFormationResourceName(prefix, ...parts)](https://godoc.org/github.com/mweagle/Sparta#CloudFormationResourceName) to help generate unique resource names.

### How can I determine the outputs available in sparta.Discover() for dynamic AWS resources?

The list of registered output provider types is defined by `cloudformationTypeMapDiscoveryOutputs` in [cloudformation_resources.go](https://github.com/mweagle/Sparta/blob/master/cloudformation_resources.go).  See the [CloudFormation Resource Types Reference](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-template-resource-type-ref.html) for information on interpreting the values.

## Future
