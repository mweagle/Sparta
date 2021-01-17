package decorator

import (
	"context"
	"fmt"

	sparta "github.com/mweagle/Sparta"
	gocf "github.com/mweagle/go-cloudformation"
	"github.com/rs/zerolog"
)

// CloudWatchErrorAlarmDecorator returns a TemplateDecoratorHookFunc
// that associates a CloudWatch Lambda Error count alarm with the given
// lambda function. The four parameters are periodWindow, minutes per period
// the strict lower bound value, and the SNS topic to which alerts should be
// sent. See the CloudWatch alarm resource type in the official
// AWS documentation at https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-cw-alarm.html
// for more information
func CloudWatchErrorAlarmDecorator(periodWindow int,
	minutesPerPeriod int,
	thresholdGreaterThanOrEqualToValue int,
	snsTopic gocf.Stringable) sparta.TemplateDecoratorHookFunc {
	alarmDecorator := func(ctx context.Context,
		serviceName string,
		lambdaResourceName string,
		lambdaResource gocf.LambdaFunction,
		resourceMetadata map[string]interface{},
		lambdaFunctionCode *gocf.LambdaFunctionCode,
		buildID string,
		template *gocf.Template,
		logger *zerolog.Logger) (context.Context, error) {

		periodInSeconds := minutesPerPeriod * 60

		alarm := &gocf.CloudWatchAlarm{
			AlarmName: gocf.Join("",
				gocf.String("ERROR Alarm for "),
				gocf.Ref(lambdaResourceName)),
			AlarmDescription: gocf.Join(" ",
				gocf.String("ERROR count for AWS Lambda function"),
				gocf.Ref(lambdaResourceName),
				gocf.String("( Stack:"),
				gocf.Ref("AWS::StackName"),
				gocf.String(") is greater than"),
				gocf.String(fmt.Sprintf("%d", thresholdGreaterThanOrEqualToValue)),
				gocf.String("over the last"),
				gocf.String(fmt.Sprintf("%d", periodInSeconds)),
				gocf.String("seconds"),
			),
			MetricName:         gocf.String("Errors"),
			Namespace:          gocf.String("AWS/Lambda"),
			Statistic:          gocf.String("Sum"),
			Period:             gocf.Integer(int64(periodInSeconds)),
			EvaluationPeriods:  gocf.Integer(int64(periodWindow)),
			Threshold:          gocf.Integer(int64(thresholdGreaterThanOrEqualToValue)),
			ComparisonOperator: gocf.String("GreaterThanOrEqualToThreshold"),
			Dimensions: &gocf.CloudWatchAlarmDimensionList{
				gocf.CloudWatchAlarmDimension{
					Name:  gocf.String("FunctionName"),
					Value: gocf.Ref(lambdaResourceName).String(),
				},
			},
			TreatMissingData: gocf.String("notBreaching"),
			AlarmActions: gocf.StringList(
				snsTopic,
			),
		}
		// Create the resource, add it...
		alarmResourceName := sparta.CloudFormationResourceName("Alarm",
			lambdaResourceName)
		template.AddResource(alarmResourceName, alarm)
		return ctx, nil
	}
	return sparta.TemplateDecoratorHookFunc(alarmDecorator)
}
