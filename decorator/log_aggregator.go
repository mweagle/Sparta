package decorator

import (
	"context"
	"fmt"

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	gof "github.com/awslabs/goformation/v5/cloudformation"
	gofiam "github.com/awslabs/goformation/v5/cloudformation/iam"
	gofkinesis "github.com/awslabs/goformation/v5/cloudformation/kinesis"
	goflambda "github.com/awslabs/goformation/v5/cloudformation/lambda"
	goflogs "github.com/awslabs/goformation/v5/cloudformation/logs"
	sparta "github.com/mweagle/Sparta/v3"
	spartaIAM "github.com/mweagle/Sparta/v3/aws/iam"
	spartaIAMBuilder "github.com/mweagle/Sparta/v3/aws/iam/builder"

	"github.com/rs/zerolog"
)

// LogAggregatorAssumePolicyDocument is the document for LogSubscription filters
var LogAggregatorAssumePolicyDocument = sparta.ArbitraryJSONObject{
	"Version": "2012-10-17",
	"Statement": []sparta.ArbitraryJSONObject{
		{
			"Action": []string{"sts:AssumeRole"},
			"Effect": "Allow",
			"Principal": sparta.ArbitraryJSONObject{
				"Service": []string{
					"logs.us-west-2.amazonaws.com",
				},
			},
		},
	},
}

/*
Inspired by

https://theburningmonk.com/2018/07/centralised-logging-for-aws-lambda-revised-2018/

Create a new LogAggregatorDecorator and then hook up the decorator to the
desired lambda functions as in:

decorator := spartaDecorators.NewLogAggregatorDecorator(kinesisResource, kinesisMapping, loggingRelay)

// Add the decorator to each function
for _, eachLambda := range lambdaFunctions {
	if eachLambda.Decorators == nil {
		eachLambda.Decorators = make([]sparta.TemplateDecoratorHandler, 0)
	}
	eachLambda.Decorators = append(eachLambda.Decorators, decorator)
}

// Add the decorator to the service
workflowHooks.ServiceDecorators = []sparta.ServiceDecoratorHookHandler{decorator}
*/

func logAggregatorResName(baseName string) string {
	return sparta.CloudFormationResourceName(fmt.Sprintf("LogAggregator%s", baseName),
		baseName)
}

// LogAggregatorDecorator is the decorator that
// satisfies both the ServiceDecoratorHandler and TemplateDecoratorHandler
// interfaces. It ensures that each lambda function has a CloudWatch logs
// subscription that forwards to a Kinesis stream. That stream is then
// subscribed to by the relay lambda function. Only log statements
// of level info or higher are published to Kinesis.
type LogAggregatorDecorator struct {
	kinesisStreamResourceName string
	iamRoleNameResourceName   string
	kinesisResource           *gofkinesis.Stream
	kinesisMapping            *sparta.EventSourceMapping
	logRelay                  *sparta.LambdaAWSInfo
}

// Ensure compliance
var _ sparta.ServiceDecoratorHookHandler = (*LogAggregatorDecorator)(nil)
var _ sparta.TemplateDecoratorHandler = (*LogAggregatorDecorator)(nil)

// KinesisLogicalResourceName returns the name of the Kinesis stream that will be provisioned
// by this Decorator
func (lad *LogAggregatorDecorator) KinesisLogicalResourceName() string {
	return lad.kinesisStreamResourceName
}

// DecorateService annotates the service with the Kinesis hook
func (lad *LogAggregatorDecorator) DecorateService(ctx context.Context,
	serviceName string,
	template *gof.Template,
	lambdaFunctionCode *goflambda.Function_Code,
	buildID string,
	awsConfig awsv2.Config,
	noop bool,
	logger *zerolog.Logger) (context.Context, error) {

	// Create the Kinesis Stream
	template.Resources[lad.kinesisStreamResourceName] = lad.kinesisResource

	// Create the IAM role
	putRecordPriv := spartaIAMBuilder.Allow("kinesis:PutRecord").
		ForResource().
		Attr(lad.kinesisStreamResourceName, "Arn").
		ToPolicyStatement()
	passRolePriv := spartaIAMBuilder.Allow("iam:PassRole").
		ForResource().
		Literal("arn:aws:iam::").
		AccountID(":").
		Literal("role/").
		Literal(lad.iamRoleNameResourceName).
		ToPolicyStatement()

	statements := make([]spartaIAM.PolicyStatement, 0)
	statements = append(statements,
		putRecordPriv,
		passRolePriv,
	)
	iamPolicyList := []gofiam.Role_Policy{
		{
			PolicyDocument: sparta.ArbitraryJSONObject{
				"Version":   "2012-10-17",
				"Statement": statements,
			},
			PolicyName: "LogAggregatorPolicy",
		},
	}

	iamLogAggregatorRole := &gofiam.Role{
		RoleName:                 lad.iamRoleNameResourceName,
		AssumeRolePolicyDocument: LogAggregatorAssumePolicyDocument,
		Policies:                 iamPolicyList,
	}
	template.Resources[lad.iamRoleNameResourceName] = iamLogAggregatorRole
	return ctx, nil
}

// DecorateTemplate annotates the lambda with the log forwarding sink info
func (lad *LogAggregatorDecorator) DecorateTemplate(ctx context.Context,
	serviceName string,
	lambdaResourceName string,
	lambdaResource *goflambda.Function,
	resourceMetadata map[string]interface{},
	lambdaFunctionCode *goflambda.Function_Code,
	buildID string,
	template *gof.Template,
	logger *zerolog.Logger) (context.Context, error) {

	// The relay function should consume the stream
	if lad.logRelay.LogicalResourceName() == lambdaResourceName {
		// Need to add a Lambda EventSourceMapping
		eventSourceMappingResourceName := sparta.CloudFormationResourceName("LogAggregator",
			"EventSourceMapping",
			lambdaResourceName)

		template.Resources[eventSourceMappingResourceName] = &goflambda.EventSourceMapping{
			StartingPosition: lad.kinesisMapping.StartingPosition,
			BatchSize:        lad.kinesisMapping.BatchSize,
			EventSourceArn:   gof.GetAtt(lad.kinesisStreamResourceName, "Arn"),
			FunctionName:     gof.GetAtt(lambdaResourceName, "Arn"),
		}

	} else {
		// The other functions should publish their logs to the stream
		subscriptionName := logAggregatorResName(fmt.Sprintf("Lambda%s", lambdaResourceName))
		subscriptionFilterRes := &goflogs.SubscriptionFilter{
			DestinationArn: gof.GetAtt(lad.kinesisStreamResourceName, "Arn"),
			RoleArn:        gof.GetAtt(lad.iamRoleNameResourceName, "Arn"),
			LogGroupName: gof.Join("", []string{
				"/aws/lambda/",
				gof.Ref(lambdaResourceName),
			}),
			FilterPattern: "{$.level = info || $.level = warning || $.level = error }",
		}
		template.Resources[subscriptionName] = subscriptionFilterRes
	}
	return ctx, nil
}

// NewLogAggregatorDecorator returns a ServiceDecoratorHook that registers a Kinesis
// stream lambda log aggregator
func NewLogAggregatorDecorator(
	kinesisResource *gofkinesis.Stream,
	kinesisMapping *sparta.EventSourceMapping,
	relay *sparta.LambdaAWSInfo) *LogAggregatorDecorator {

	return &LogAggregatorDecorator{
		kinesisStreamResourceName: logAggregatorResName("Kinesis"),
		kinesisResource:           kinesisResource,
		kinesisMapping:            kinesisMapping,
		iamRoleNameResourceName:   logAggregatorResName("IAMRole"),
		logRelay:                  relay,
	}
}
