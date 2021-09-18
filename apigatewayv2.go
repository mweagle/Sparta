package sparta

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	gof "github.com/awslabs/goformation/v5/cloudformation"
	gofapigv2 "github.com/awslabs/goformation/v5/cloudformation/apigatewayv2"
	gofddb "github.com/awslabs/goformation/v5/cloudformation/dynamodb"
	goflambda "github.com/awslabs/goformation/v5/cloudformation/lambda"
	gocf "github.com/mweagle/go-cloudformation"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// Ref: https://github.com/aws-samples/simple-websockets-chat-app

// APIV2RouteSelectionExpression represents a route selection
type APIV2RouteSelectionExpression string

// APIV2Protocol is the type of API V2 protocols
type APIV2Protocol string

const (
	// Websocket represents the only supported V2 protocol
	Websocket APIV2Protocol = "WEBSOCKET"
)

// APIV2 contains the information necessary for the routes in here.
// Please tell me they can use the same routes...
// They cannot
type APIV2 struct {
	protocol                  APIV2Protocol
	name                      string
	routeSelectionExpression  string
	stage                     *APIV2Stage
	APIKeySelectionExpression string
	Description               string
	DisableSchemaValidation   bool
	Tags                      map[string]interface{}
	Version                   string
	// Routes mapping selection expression to Route handler
	routes map[APIV2RouteSelectionExpression]*APIV2Route
}

// APIV2GatewayDecorator is the compound decorator that handles both
// the DDB table creation and the lambda decorator...winning.
type APIV2GatewayDecorator struct {
	envTableKeyName string
	propertyName    string
	readCapacity    int64
	writeCapacity   int64
}

func (apigd *APIV2GatewayDecorator) logicalResourceName() string {
	return CloudFormationResourceName("WSSConnectionTable",
		"WSSConnectionTable")
}

// DecorateService handles inserting the DDB Table
func (apigd *APIV2GatewayDecorator) DecorateService(context map[string]interface{},
	serviceName string,
	template *gof.Template,
	S3Bucket string,
	S3Key string,
	buildID string,
	awsSession *session.Session,
	noop bool,
	logger *zerolog.Logger) error {

	// Create the table...
	dynamoDBResourceName := apigd.logicalResourceName()
	dynamoDBResource := &gofddb.Table{
		AttributeDefinitions: []gofddb.Table_AttributeDefinition{
			gofddb.Table_AttributeDefinition{
				AttributeName: apigd.propertyName,
				AttributeType: "S",
			},
		},
		KeySchema: []gofddb.Table_KeySchema{
			gofddb.Table_KeySchema{
				AttributeName: apigd.propertyName,
				KeyType:       "HASH",
			},
		},
		SSESpecification: &gofddb.Table_SSESpecification{
			SSEEnabled: true,
		},
		ProvisionedThroughput: &gofddb.Table_ProvisionedThroughput{
			ReadCapacityUnits:  apigd.readCapacity,
			WriteCapacityUnits: apigd.writeCapacity,
		},
	}
	template.Resources[dynamoDBResourceName] = dynamoDBResource
	return nil
}

// AnnotateLambdas handles hooking up the lambda perms
func (apigd *APIV2GatewayDecorator) AnnotateLambdas(lambdaFns []*LambdaAWSInfo) error {

	var ddbPermissions = []IAMRolePrivilege{
		{
			Actions: []string{"dynamodb:GetItem",
				"dynamodb:DeleteItem",
				"dynamodb:PutItem",
				"dynamodb:Scan",
				"dynamodb:Query",
				"dynamodb:UpdateItem",
				"dynamodb:BatchWriteItem",
				"dynamodb:BatchGetItem"},
			Resource: gocf.Join("",
				gocf.String("arn:"),
				gocf.Ref("AWS::Partition"),
				gocf.String(":dynamodb:"),
				gocf.Ref("AWS::Region"),
				gocf.String(":"),
				gocf.Ref("AWS::AccountId"),
				gocf.String(":table/"),
				gocf.Ref(apigd.logicalResourceName())),
		},
		{
			Actions: []string{"dynamodb:GetItem",
				"dynamodb:DeleteItem",
				"dynamodb:PutItem",
				"dynamodb:Scan",
				"dynamodb:Query",
				"dynamodb:UpdateItem",
				"dynamodb:BatchWriteItem",
				"dynamodb:BatchGetItem"},
			Resource: gocf.Join("",
				gocf.String("arn:"),
				gocf.Ref("AWS::Partition"),
				gocf.String(":dynamodb:"),
				gocf.Ref("AWS::Region"),
				gocf.String(":"),
				gocf.Ref("AWS::AccountId"),
				gocf.String(":table/"),
				gocf.Ref(apigd.logicalResourceName()),
				gocf.String("/index/*")),
		},
	}

	for _, eachLambda := range lambdaFns {
		// Add the permission
		eachLambda.RoleDefinition.Privileges = append(eachLambda.RoleDefinition.Privileges,
			ddbPermissions...)

		// Add the env
		env := eachLambda.Options.Environment
		if env == nil {
			env = make(map[string]string)
		}
		env[apigd.envTableKeyName] = gof.Ref(apigd.logicalResourceName())
		eachLambda.Options.Environment = env
	}
	return nil
}

// NewConnectionTableDecorator returns a *APIV2GatewayDecorator that handles
// creating the DynamoDDB table and hooking up all the lambda permissions
func (apiv2 *APIV2) NewConnectionTableDecorator(envTableNameKey string,
	propertyName string,
	readCapacity int64,
	writeCapacity int64) (*APIV2GatewayDecorator, error) {

	return &APIV2GatewayDecorator{
		envTableKeyName: envTableNameKey,
		propertyName:    propertyName,
		readCapacity:    readCapacity,
		writeCapacity:   writeCapacity,
	}, nil
}

// NewAPIV2Route returns a new Route
func (apiv2 *APIV2) NewAPIV2Route(routeKey APIV2RouteSelectionExpression,
	lambdaFn *LambdaAWSInfo) (*APIV2Route, error) {

	_, exists := apiv2.routes[routeKey]
	if exists {
		return nil, errors.Errorf("APIV2 Route for expression `%s` already exists",
			routeKey)
	}
	route := &APIV2Route{
		routeKey: routeKey,
		lambdaFn: lambdaFn,
		Integration: &APIV2Integration{
			IntegrationType: "AWS_PROXY",
		},
	}
	apiv2.routes[routeKey] = route
	return route, nil
}

// LogicalResourceName returns the logical resoource name of this API V2 Gateway
// instance
func (apiv2 *APIV2) LogicalResourceName() string {
	return CloudFormationResourceName("APIGateway",
		fmt.Sprintf("v2%s", apiv2.name))
}

// Describe satisfies the API interface
func (apiv2 *APIV2) Describe(targetNodeName string) (*DescriptionInfo, error) {
	descInfo := &DescriptionInfo{
		Name:  "APIGatewayV2",
		Nodes: make([]*DescriptionTriplet, 0),
	}

	descInfo.Nodes = append(descInfo.Nodes, &DescriptionTriplet{
		SourceNodeName: nodeNameAPIGateway,
		DisplayInfo: &DescriptionDisplayInfo{
			SourceNodeColor: nodeColorAPIGateway,
			SourceIcon: &DescriptionIcon{
				Category: "Mobile",
				Name:     "Amazon-API-Gateway_light-bg@4x.png",
			},
		},
		TargetNodeName: targetNodeName,
	})

	for eachRouteExpr, eachRoute := range apiv2.routes {
		opName := ""
		if eachRoute.OperationName != "" {
			opName = fmt.Sprintf(" - %s", eachRoute.OperationName)
		}
		var nodeName = fmt.Sprintf("%s%s", eachRouteExpr, opName)

		descInfo.Nodes = append(descInfo.Nodes,
			&DescriptionTriplet{
				SourceNodeName: nodeNameAPIGateway,
				DisplayInfo: &DescriptionDisplayInfo{
					SourceNodeColor: nodeColorAPIGateway,
					SourceIcon: &DescriptionIcon{
						Category: "_General",
						Name:     "Internet-alt1_light-bg@4x.png",
					},
				},
				TargetNodeName: nodeName,
			},
			&DescriptionTriplet{
				SourceNodeName: nodeName,
				TargetNodeName: eachRoute.lambdaFn.lambdaFunctionName(),
			})
	}
	return descInfo, nil
}

// Marshal the API V2 Gateway instance to the given template instane
func (apiv2 *APIV2) Marshal(serviceName string,
	session *session.Session,
	lambdaFunctionCode *goflambda.Function_Code,
	roleNameMap map[string]string,
	template *gof.Template,
	noop bool,
	logger *zerolog.Logger) error {

	apiV2Entry := &gofapigv2.Api{
		ApiKeySelectionExpression: apiv2.APIKeySelectionExpression,
		Description:               apiv2.Description,
		DisableSchemaValidation:   apiv2.DisableSchemaValidation,
		Name:                      apiv2.name,
		ProtocolType:              string(Websocket),
		RouteSelectionExpression:  apiv2.routeSelectionExpression,
		Version:                   apiv2.Version,
	}
	// Add it
	template.Resources[apiv2.LogicalResourceName()] = apiV2Entry

	allRouteResources := []string{}

	// Alright, setup the route
	for eachExpression, eachRoute := range apiv2.routes {
		routeResourceName := CloudFormationResourceName("Route", string(eachExpression))
		allRouteResources = append(allRouteResources, routeResourceName)

		routeIntegrationResourceName := CloudFormationResourceName("RouteIntg", string(eachExpression))
		routeEntry := &gofapigv2.Route{
			ApiId:                            gof.Ref(apiv2.LogicalResourceName()),
			ApiKeyRequired:                   eachRoute.APIKeyRequired,
			AuthorizerId:                     eachRoute.AuthorizerID,
			AuthorizationScopes:              eachRoute.AuthorizationScopes,
			AuthorizationType:                eachRoute.AuthorizationType,
			ModelSelectionExpression:         eachRoute.ModelSelectionExpression,
			OperationName:                    eachRoute.OperationName,
			RequestModels:                    eachRoute.RequestModels,
			RequestParameters:                eachRoute.RequestParameters,
			RouteKey:                         string(eachRoute.routeKey),
			RouteResponseSelectionExpression: eachRoute.RouteResponseSelectionExpression,
			Target: gof.Join("/", []string{
				"integrations",
				gof.Ref(routeIntegrationResourceName),
			}),
		}

		// Add the route resource
		template.Resources[routeResourceName] = routeEntry

		// Add the integration
		routeIntegration := &gofapigv2.Integration{
			ApiId:                   gof.Ref(apiv2.LogicalResourceName()),
			ConnectionType:          eachRoute.Integration.ConnectionType,
			ContentHandlingStrategy: eachRoute.Integration.ContentHandlingStrategy,
			CredentialsArn:          eachRoute.Integration.CredentialsArn,
			Description:             eachRoute.Integration.Description,
			IntegrationMethod:       eachRoute.Integration.IntegrationMethod,
			IntegrationType:         eachRoute.Integration.IntegrationType,
			IntegrationUri: gof.Join("", []string{
				"arn:aws:apigateway:",
				gof.Ref("AWS::Region"),
				":lambda:path/2015-03-31/functions/",
				gof.GetAtt(eachRoute.lambdaFn.LogicalResourceName(), "Arn"),
				"/invocations",
			}),
			PassthroughBehavior: eachRoute.Integration.PassthroughBehavior,
			// TODO - auto create this...
			RequestParameters:           eachRoute.Integration.RequestParameters,
			RequestTemplates:            eachRoute.Integration.RequestTemplates,
			TemplateSelectionExpression: eachRoute.Integration.TemplateSelectionExpression,
			TimeoutInMillis:             eachRoute.Integration.TimeoutInMillis,
		}
		template.Resources[routeIntegrationResourceName] = routeIntegration

		// Add the lambda permission
		apiGatewayPermissionResourceName := CloudFormationResourceName("APIV2GatewayLambdaPerm",
			string(eachExpression))
		lambdaInvokePermission := &goflambda.Permission{
			Action:       "lambda:InvokeFunction",
			FunctionName: gof.GetAtt(eachRoute.lambdaFn.LogicalResourceName(), "Arn"),
			Principal:    APIGatewayPrincipal,
		}
		template.Resources[apiGatewayPermissionResourceName] = lambdaInvokePermission
	}

	// Add the Stage and Deploy...
	stageResourceName := CloudFormationResourceName("APIV2GatewayStage", "APIV2GatewayStage")
	// Use an unstable ID s.t. we can actually create a new deployment event.
	deploymentResName := CloudFormationResourceName("APIV2GatewayDeployment")

	// Unstable name to trigger a deployment
	newDeployment := &gofapigv2.Deployment{
		ApiId:       gof.Ref(apiv2.LogicalResourceName()),
		Description: apiv2.stage.Description,
	}
	// Use an unstable ID s.t. we can actually create a new deployment event.  Not sure how this
	// is going to work with deletes...
	newDeployment.AWSCloudFormationDeletionPolicy = "Retain"
	newDeployment.AWSCloudFormationDependsOn = allRouteResources

	template.Resources[deploymentResName] = newDeployment

	// Add the stage...
	stageResource := &gofapigv2.Stage{
		ApiId:                gof.Ref(apiv2.LogicalResourceName()),
		DeploymentId:         gof.Ref(deploymentResName),
		StageName:            apiv2.stage.name,
		AccessLogSettings:    apiv2.stage.AccessLogSettings,
		ClientCertificateId:  apiv2.stage.ClientCertificateID,
		DefaultRouteSettings: apiv2.stage.DefaultRouteSettings,
		Description:          apiv2.stage.Description,
		RouteSettings:        apiv2.stage.RouteSettings,
		StageVariables:       apiv2.stage.StageVariables,
		//Tags:                 apiv2.stage.Tags,
	}
	template.Resources[stageResourceName] = stageResource

	// Outputs...
	template.Outputs[OutputAPIGatewayURL] = gof.Output{
		Description: "API Gateway Websocket URL",
		Value: gof.Join("", []string{
			"wss://",
			gof.Ref(apiv2.LogicalResourceName()),
			".execute-api.",
			gof.Ref("AWS::Region"),
			".amazonaws.com/",
			apiv2.stage.name,
		}),
	}
	return nil
}

// NewAPIV2 returns a new API V2 Gateway instance
func NewAPIV2(protocol APIV2Protocol,
	name string,
	routeSelectionExpression string,
	stage *APIV2Stage) (*APIV2, error) {
	return &APIV2{
		protocol:                 protocol,
		name:                     name,
		routeSelectionExpression: routeSelectionExpression,
		stage:                    stage,
		routes:                   make(map[APIV2RouteSelectionExpression]*APIV2Route),
		Tags:                     make(map[string]interface{}),
	}, nil
}

// APIV2Route represents a V2 route
type APIV2Route struct {
	routeKey                         APIV2RouteSelectionExpression
	APIKeyRequired                   bool
	AuthorizationScopes              []string
	AuthorizationType                string
	AuthorizerID                     string
	ModelSelectionExpression         string
	OperationName                    string
	RequestModels                    interface{}
	RequestParameters                interface{}
	RouteResponseSelectionExpression string
	Integration                      *APIV2Integration
	lambdaFn                         *LambdaAWSInfo
}

// APIV2Stage represents the deployment stage
type APIV2Stage struct {
	AccessLogSettings    *gofapigv2.Stage_AccessLogSettings
	ClientCertificateID  string
	DefaultRouteSettings *gofapigv2.Stage_RouteSettings
	Description          string
	RouteSettings        interface{}
	name                 string
	StageVariables       interface{}
	Tags                 interface{}
}

// NewAPIV2Stage returns a new APIV2Stage entry
func NewAPIV2Stage(stageName string) (*APIV2Stage, error) {
	return &APIV2Stage{
		name: stageName,
	}, nil
}

// APIV2Integration is the integration type for an APIV2Route
// entry
type APIV2Integration struct {
	//ApiID                       string
	ConnectionType          string
	ContentHandlingStrategy string
	CredentialsArn          string
	Description             string
	IntegrationMethod       string
	IntegrationType         string
	//IntegrationUri              string
	PassthroughBehavior         string
	RequestParameters           interface{}
	RequestTemplates            interface{}
	TemplateSelectionExpression string
	TimeoutInMillis             int
}
