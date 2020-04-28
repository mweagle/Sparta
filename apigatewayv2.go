package sparta

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	gocf "github.com/mweagle/go-cloudformation"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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
	APIDescription            string
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
	template *gocf.Template,
	S3Bucket string,
	S3Key string,
	buildID string,
	awsSession *session.Session,
	noop bool,
	logger *logrus.Logger) error {

	// Create the table...
	dynamoDBResourceName := apigd.logicalResourceName()
	dynamoDBResource := &gocf.DynamoDBTable{
		AttributeDefinitions: &gocf.DynamoDBTableAttributeDefinitionList{
			gocf.DynamoDBTableAttributeDefinition{
				AttributeName: gocf.String(apigd.propertyName),
				AttributeType: gocf.String("S"),
			},
		},
		KeySchema: &gocf.DynamoDBTableKeySchemaList{
			gocf.DynamoDBTableKeySchema{
				AttributeName: gocf.String(apigd.propertyName),
				KeyType:       gocf.String("HASH"),
			},
		},
		SSESpecification: &gocf.DynamoDBTableSSESpecification{
			SSEEnabled: gocf.Bool(true),
		},
		ProvisionedThroughput: &gocf.DynamoDBTableProvisionedThroughput{
			ReadCapacityUnits:  gocf.Integer(apigd.readCapacity),
			WriteCapacityUnits: gocf.Integer(apigd.writeCapacity),
		},
	}
	template.AddResource(dynamoDBResourceName, dynamoDBResource)
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
			env = make(map[string]*gocf.StringExpr)
		}
		env[apigd.envTableKeyName] = gocf.Ref(apigd.logicalResourceName()).String()
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
func (apiv2 *APIV2) Description(targetNodeName string) (*DescriptionInfo, error) {
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
				Name:     "Amazon-API-Gateway_light-bg.svg",
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
						Name:     "Internet-alt1_light-bg.svg",
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

// Marshal marshals the API V2 Gateway instance to the given template instance
func (apiv2 *APIV2) Marshal(serviceName string,
	session *session.Session,
	S3Bucket string,
	S3Key string,
	S3Version string,
	roleNameMap map[string]*gocf.StringExpr,
	template *gocf.Template,
	noop bool,
	logger *logrus.Logger) error {

	apiV2Entry := &gocf.APIGatewayV2API{
		APIKeySelectionExpression: marshalString(apiv2.APIKeySelectionExpression),
		Description:               marshalString(apiv2.APIDescription),
		DisableSchemaValidation:   marshalBool(apiv2.DisableSchemaValidation),
		Name:                      marshalString(apiv2.name),
		ProtocolType:              marshalString(string(Websocket)),
		RouteSelectionExpression:  marshalString(apiv2.routeSelectionExpression),
		Version:                   marshalString(apiv2.Version),
	}
	// Add it
	apiV2Resource := template.AddResource(apiv2.LogicalResourceName(), apiV2Entry)
	apiV2Resource.DependsOn = []string{}
	allRouteResources := []string{}

	// Alright, setup the route
	for eachExpression, eachRoute := range apiv2.routes {
		routeResourceName := CloudFormationResourceName("Route", string(eachExpression))
		allRouteResources = append(allRouteResources, routeResourceName)

		routeIntegrationResourceName := CloudFormationResourceName("RouteIntg", string(eachExpression))
		routeEntry := &gocf.APIGatewayV2Route{
			APIID:                            gocf.Ref(apiv2.LogicalResourceName()).String(),
			APIKeyRequired:                   marshalBool(eachRoute.APIKeyRequired),
			AuthorizerID:                     marshalStringExpr(eachRoute.AuthorizerID),
			AuthorizationScopes:              marshalStringList(eachRoute.AuthorizationScopes),
			AuthorizationType:                marshalString(eachRoute.AuthorizationType),
			ModelSelectionExpression:         marshalString(eachRoute.ModelSelectionExpression),
			OperationName:                    marshalString(eachRoute.OperationName),
			RequestModels:                    marshalInterface(eachRoute.RequestModels),
			RequestParameters:                marshalInterface(eachRoute.RequestParameters),
			RouteKey:                         marshalString(string(eachRoute.routeKey)),
			RouteResponseSelectionExpression: marshalString(eachRoute.RouteResponseSelectionExpression),
			Target: gocf.Join("/",
				gocf.String("integrations"),
				gocf.Ref(routeIntegrationResourceName)),
		}

		// Add the route resource
		template.AddResource(routeResourceName, routeEntry)
		//apiV2Resource.DependsOn = append(apiV2Resource.DependsOn, routeResourceName)

		// Add the integration
		routeIntegration := &gocf.APIGatewayV2Integration{
			APIID:                   gocf.Ref(apiv2.LogicalResourceName()).String(),
			ConnectionType:          marshalString(eachRoute.Integration.ConnectionType),
			ContentHandlingStrategy: marshalString(eachRoute.Integration.ContentHandlingStrategy),
			CredentialsArn:          marshalStringExpr(eachRoute.Integration.CredentialsArn),
			Description:             marshalString(eachRoute.Integration.Description),
			IntegrationMethod:       marshalString(eachRoute.Integration.IntegrationMethod),
			IntegrationType:         marshalString(eachRoute.Integration.IntegrationType),
			IntegrationURI: gocf.Join("",
				gocf.String("arn:aws:apigateway:"),
				gocf.Ref("AWS::Region"),
				gocf.String(":lambda:path/2015-03-31/functions/"),
				gocf.GetAtt(eachRoute.lambdaFn.LogicalResourceName(), "Arn"),
				gocf.String("/invocations")),
			PassthroughBehavior: marshalString(eachRoute.Integration.PassthroughBehavior),
			// TODO - auto create this...
			RequestParameters:           marshalInterface(eachRoute.Integration.RequestParameters),
			RequestTemplates:            marshalInterface(eachRoute.Integration.RequestTemplates),
			TemplateSelectionExpression: marshalString(eachRoute.Integration.TemplateSelectionExpression),
			TimeoutInMillis:             marshalInt(eachRoute.Integration.TimeoutInMillis),
		}
		template.AddResource(routeIntegrationResourceName, routeIntegration)

		// Add the lambda permission
		apiGatewayPermissionResourceName := CloudFormationResourceName("APIV2GatewayLambdaPerm",
			string(eachExpression))
		lambdaInvokePermission := &gocf.LambdaPermission{
			Action:       gocf.String("lambda:InvokeFunction"),
			FunctionName: gocf.GetAtt(eachRoute.lambdaFn.LogicalResourceName(), "Arn"),
			Principal:    gocf.String(APIGatewayPrincipal),
		}
		template.AddResource(apiGatewayPermissionResourceName, lambdaInvokePermission)
	}

	// Add the Stage and Deploy...
	stageResourceName := CloudFormationResourceName("APIV2GatewayStage", "APIV2GatewayStage")
	// Use an unstable ID s.t. we can actually create a new deployment event.
	deploymentResName := CloudFormationResourceName("APIV2GatewayDeployment")

	// Unstable name to trigger a deployment
	newDeployment := &gocf.APIGatewayV2Deployment{
		APIID:       gocf.Ref(apiv2.LogicalResourceName()).String(),
		Description: marshalString(apiv2.stage.Description),
	}
	// Use an unstable ID s.t. we can actually create a new deployment event.  Not sure how this
	// is going to work with deletes...
	deployment := template.AddResource(deploymentResName, newDeployment)
	deployment.DeletionPolicy = "Retain"
	deployment.DependsOn = append(deployment.DependsOn, allRouteResources...)

	// Add the stage...
	stageResource := &gocf.APIGatewayV2Stage{
		APIID:                gocf.Ref(apiv2.LogicalResourceName()).String(),
		DeploymentID:         gocf.Ref(deploymentResName).String(),
		StageName:            marshalString(apiv2.stage.name),
		AccessLogSettings:    apiv2.stage.AccessLogSettings,
		ClientCertificateID:  marshalStringExpr(apiv2.stage.ClientCertificateID),
		DefaultRouteSettings: apiv2.stage.DefaultRouteSettings,
		Description:          marshalString(apiv2.stage.Description),
		RouteSettings:        marshalInterface(apiv2.stage.RouteSettings),
		StageVariables:       marshalInterface(apiv2.stage.StageVariables),
		//Tags:                 marshalInterface(apiv2.stage.Tags),
	}
	template.AddResource(stageResourceName, stageResource)

	// Outputs...
	template.Outputs[OutputAPIGatewayURL] = &gocf.Output{
		Description: "API Gateway Websocket URL",
		Value: gocf.Join("",
			gocf.String("wss://"),
			gocf.Ref(apiv2.LogicalResourceName()),
			gocf.String(".execute-api."),
			gocf.Ref("AWS::Region"),
			gocf.String(".amazonaws.com/"),
			gocf.String(apiv2.stage.name)),
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
	AuthorizerID                     gocf.Stringable
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
	AccessLogSettings    *gocf.APIGatewayV2StageAccessLogSettings
	ClientCertificateID  gocf.Stringable
	DefaultRouteSettings *gocf.APIGatewayV2StageRouteSettings
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
	CredentialsArn          gocf.Stringable
	Description             string
	IntegrationMethod       string
	IntegrationType         string
	//IntegrationUri              string
	PassthroughBehavior         string
	RequestParameters           interface{}
	RequestTemplates            interface{}
	TemplateSelectionExpression string
	TimeoutInMillis             int64
}
