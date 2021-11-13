package decorator

import (
	"bytes"
	"context"
	"regexp"
	"text/template"

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	gof "github.com/awslabs/goformation/v5/cloudformation"
	gofcloudwatch "github.com/awslabs/goformation/v5/cloudformation/cloudwatch"

	goflambda "github.com/awslabs/goformation/v5/cloudformation/lambda"
	sparta "github.com/mweagle/Sparta/v3"
	spartaCF "github.com/mweagle/Sparta/v3/aws/cloudformation"
	"github.com/rs/zerolog"
)

const (
	// OutputDashboardURL is the keyname used in the CloudFormation Output
	// that stores the CloudWatch Dashboard URL
	// @enum OutputKey
	OutputDashboardURL = "CloudWatchDashboardURL"
)

const (
	headerWidthUnits  = 9
	headerHeightUnits = 6
	metricsPerRow     = 3
	metricWidthUnits  = 6
	metricHeightUnits = 6
)

// widgetExtents represents the extents of various containers in the generated
// dashboard
type widgetExtents struct {
	HeaderWidthUnits  int
	HeaderHeightUnits int
	MetricWidthUnits  int
	MetricHeightUnits int
	MetricsPerRow     int
}

// LambdaTemplateData is the mapping of Sparta public LambdaAWSInfo together
// with the CloudFormationResource name this resource uses
type LambdaTemplateData struct {
	LambdaAWSInfo *sparta.LambdaAWSInfo
	ResourceName  string
}

// DashboardTemplateData is the object supplied to the dashboard template
// to generate the resulting dashboard
type DashboardTemplateData struct {
	// The list of lambda functions
	LambdaFunctions []*LambdaTemplateData
	// SpartaVersion is the Sparta library used to provision this service
	SpartaVersion string
	// SpartaGitHash is the commit hash of this version of the library
	SpartaGitHash    string
	TimeSeriesPeriod int
	Extents          widgetExtents
}

// The default dashboard template
var dashboardTemplate = `
{
    "widgets": [
    {
        "type": "text",
        "x": 0,
        "y": 0,
        "width": << .Extents.HeaderWidthUnits >>,
        "height": << .Extents.HeaderHeightUnits >>,
        "properties": {
						"markdown": "## ![Sparta](https://s3-us-west-2.amazonaws.com/weagle-sparta-public/cloudwatch/SpartaHelmet32.png) { "Ref" : "AWS::StackName" } Summary\n
* ☁️ [CloudFormation Stack](https://{ "Ref" : "AWS::Region" }.console.aws.amazon.com/cloudformation/home?region={ "Ref" : "AWS::Region" }#/stack/detail?stackId={"Ref" : "AWS::StackId"})\n
* ☢️ [XRay](https://{ "Ref" : "AWS::Region" }.console.aws.amazon.com/xray/home?region={ "Ref" : "AWS::Region" }#/service-map)\n
* **Lambda Count** : << len .LambdaFunctions >>\n
* **Sparta Version** : << .SpartaVersion >> ( [<< .SpartaGitHash >>](https://github.com/mweagle/Sparta/commit/<< .SpartaGitHash >>) )\n
  * 🔗 [Sparta Documentation](https://gosparta.io)\n"
		}
    },
    {
        "type": "text",
        "x": << .Extents.HeaderWidthUnits >>,
        "y": 0,
        "width": << .Extents.HeaderWidthUnits >>,
        "height": << .Extents.HeaderHeightUnits >>,
        "properties": {
            "markdown": "## ![Sparta](https://mweagle.github.io/SpartaPublicResources/sparta/SpartaHelmet32.png) { "Ref" : "AWS::StackName" } Logs\n
<<range $index, $eachLambda := .LambdaFunctions>>
* 🔎 [{ "Ref" : "<< $eachLambda.ResourceName >>" }](https://{ "Ref" : "AWS::Region" }.console.aws.amazon.com/cloudwatch/home?region={ "Ref" : "AWS::Region" }#logStream:group=/aws/lambda/{ "Ref" : "<< $eachLambda.ResourceName >>" })\n
<<end>>"
        }
    }<<range $index, $eachLambda := .LambdaFunctions>>,
    {
      "type": "metric",
      "x": <<widgetX $index >>,
      "y": <<widgetY $index >>,
      "width": << $.Extents.MetricWidthUnits >>,
      "height": << $.Extents.MetricHeightUnits >>,
      "properties": {
        "view": "timeSeries",
        "stacked": false,
        "metrics": [
            [ "AWS/Lambda", "Invocations", "FunctionName", "{ "Ref" : "<< $eachLambda.ResourceName >>" }", { "stat": "Sum" }],
						[ ".", "Errors", ".", ".", { "stat": "Sum" }],
						[ ".", "Throttles", ".", ".", { "stat": "Sum" } ]
        ],
        "region": "{ "Ref" : "AWS::Region" }",
        "period": << $.TimeSeriesPeriod >>,
        "title": "λ: { "Ref" : "<< $eachLambda.ResourceName >>" }"
      }
    }<<end>>
  ]
}
`

var templateFuncMap = template.FuncMap{
	// The name "inc" is what the function will be called in the template text.
	"widgetX": func(lambdaIndex int) int {
		return metricWidthUnits * (lambdaIndex % metricsPerRow)
	},
	"widgetY": func(lambdaIndex int) int {
		xRow := 1
		xRow += (int)((float64)(lambdaIndex % metricsPerRow))
		// That's the row
		return headerHeightUnits + (xRow * metricHeightUnits)
	},
}

// DashboardDecorator returns a ServiceDecoratorHook function that
// can be attached the workflow to create a dashboard
func DashboardDecorator(lambdaAWSInfo []*sparta.LambdaAWSInfo,
	timeSeriesPeriod int) sparta.ServiceDecoratorHookFunc {
	return func(ctx context.Context,
		serviceName string,
		cfTemplate *gof.Template,
		lambdaFunctionCode *goflambda.Function_Code,
		buildID string,
		awsConfig awsv2.Config,
		noop bool,
		logger *zerolog.Logger) (context.Context, error) {

		lambdaFunctions := make([]*LambdaTemplateData, len(lambdaAWSInfo))
		for index, eachLambda := range lambdaAWSInfo {
			lambdaFunctions[index] = &LambdaTemplateData{
				LambdaAWSInfo: eachLambda,
				ResourceName:  eachLambda.LogicalResourceName(),
			}
		}
		dashboardTemplateData := &DashboardTemplateData{
			SpartaVersion:    sparta.SpartaVersion,
			SpartaGitHash:    sparta.SpartaGitHash,
			LambdaFunctions:  lambdaFunctions,
			TimeSeriesPeriod: timeSeriesPeriod,
			Extents: widgetExtents{
				HeaderWidthUnits:  headerWidthUnits,
				HeaderHeightUnits: headerHeightUnits,
				MetricWidthUnits:  metricWidthUnits,
				MetricHeightUnits: metricHeightUnits,
				MetricsPerRow:     metricsPerRow,
			},
		}

		dashboardTmpl, dashboardTmplErr := template.New("dashboard").
			Delims("<<", ">>").
			Funcs(templateFuncMap).
			Parse(dashboardTemplate)
		if nil != dashboardTmplErr {
			return ctx, dashboardTmplErr
		}
		var templateResults bytes.Buffer
		evalResultErr := dashboardTmpl.Execute(&templateResults, dashboardTemplateData)
		if nil != evalResultErr {
			return ctx, evalResultErr
		}

		// Raw template output
		logger.Debug().
			Str("Dashboard", templateResults.String()).
			Msg("CloudWatch Dashboard template result")

		// Replace any multiline backtick newlines with nothing, since otherwise
		// the Fn::Joined JSON will be malformed
		reReplace, reReplaceErr := regexp.Compile("\n")
		if nil != reReplaceErr {
			return ctx, reReplaceErr
		}
		escapedBytes := reReplace.ReplaceAll(templateResults.Bytes(), []byte(""))

		logger.Debug().
			Str("Dashboard", string(escapedBytes)).
			Msg("CloudWatch Dashboard post cleanup")

		// Super, now parse this into an Fn::Join representation
		// so that we can get inline expansion of the AWS pseudo params
		templateReader := bytes.NewReader(escapedBytes)
		templateExpr, templateExprErr := spartaCF.ConvertToTemplateExpression(templateReader, nil)
		if nil != templateExprErr {
			return ctx, templateExprErr
		}

		dashboardResource := &gofcloudwatch.Dashboard{
			DashboardName: serviceName,
			DashboardBody: templateExpr,
		}
		dashboardName := sparta.CloudFormationResourceName("Dashboard", "Dashboard")
		cfTemplate.Resources[dashboardName] = dashboardResource

		// Add the output
		cfTemplate.Outputs[OutputDashboardURL] = gof.Output{
			Description: "CloudWatch Dashboard URL",
			Value: gof.Join("", []string{
				"https://",
				gof.Ref("AWS::Region"),
				".console.aws.amazon.com/cloudwatch/home?region=",
				gof.Ref("AWS::Region"),
				"#dashboards:name=",
				gof.Ref(dashboardName),
			}),
		}
		return ctx, nil
	}
}
