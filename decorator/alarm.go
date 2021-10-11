package decorator

import (
	"context"
	"fmt"

	gof "github.com/awslabs/goformation/v5/cloudformation"
	gofcloudwatch "github.com/awslabs/goformation/v5/cloudformation/cloudwatch"
	goflambda "github.com/awslabs/goformation/v5/cloudformation/lambda"
	sparta "github.com/mweagle/Sparta"
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
	thresholdGreaterThanOrEqualToValue float64,
	snsTopic string) sparta.TemplateDecoratorHookFunc {
	alarmDecorator := func(ctx context.Context,
		serviceName string,
		lambdaResourceName string,
		lambdaResource *goflambda.Function,
		resourceMetadata map[string]interface{},
		lambdaFunctionCode *goflambda.Function_Code,
		buildID string,
		template *gof.Template,
		logger *zerolog.Logger) (context.Context, error) {

		periodInSeconds := minutesPerPeriod * 60

		alarm := &gofcloudwatch.Alarm{
			AlarmName: gof.Join("", []string{
				"ERROR Alarm for ",
				gof.Ref(lambdaResourceName)}),

			AlarmDescription: gof.Join(" ", []string{
				"ERROR count for AWS Lambda function",
				gof.Ref(lambdaResourceName),
				"( Stack:",
				gof.Ref("AWS::StackName"),
				") is greater than",
				fmt.Sprintf("%.2f", thresholdGreaterThanOrEqualToValue),
				"over the last",
				fmt.Sprintf("%d", periodInSeconds),
				"seconds",
			}),
			MetricName:         "Errors",
			Namespace:          "AWS/Lambda",
			Statistic:          "Sum",
			Period:             periodInSeconds,
			EvaluationPeriods:  periodWindow,
			Threshold:          thresholdGreaterThanOrEqualToValue,
			ComparisonOperator: "GreaterThanOrEqualToThreshold",
			Dimensions: []gofcloudwatch.Alarm_Dimension{
				{
					Name:  "FunctionName",
					Value: lambdaResourceName,
				},
			},
			TreatMissingData: "notBreaching",
			AlarmActions:     []string{snsTopic},
		}
		// Create the resource, add it...
		alarmResourceName := sparta.CloudFormationResourceName("Alarm",
			lambdaResourceName)
		template.Resources[alarmResourceName] = alarm
		return ctx, nil
	}
	return sparta.TemplateDecoratorHookFunc(alarmDecorator)
}
