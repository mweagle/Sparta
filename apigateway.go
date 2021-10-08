package sparta

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	awsv2APIG "github.com/aws/aws-sdk-go-v2/service/apigateway"
	awsv2APIGTypes "github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	gof "github.com/awslabs/goformation/v5/cloudformation"
	gofapig "github.com/awslabs/goformation/v5/cloudformation/apigateway"
	goflambda "github.com/awslabs/goformation/v5/cloudformation/lambda"
	"github.com/rs/zerolog"
)

// APIGateway repreents a type of API Gateway provisoining that can be exported
type APIGateway interface {
	LogicalResourceName() string
	Marshal(serviceName string,
		awsConfig awsv2.Config,
		lambdaFunctionCode *goflambda.Function_Code,
		roleNameMap map[string]string,
		template *gof.Template,
		noop bool,
		logger *zerolog.Logger) error
	Describe(targetNodeName string) (*DescriptionInfo, error)
}

var defaultCORSHeaders = map[string]interface{}{
	"Access-Control-Allow-Headers": "Content-Type,X-Amz-Date,Authorization,X-Api-Key",
	"Access-Control-Allow-Methods": "*",
	"Access-Control-Allow-Origin":  "*",
}

const (
	// OutputAPIGatewayURL is the keyname used in the CloudFormation Output
	// that stores the APIGateway provisioned URL
	// @enum OutputKey
	OutputAPIGatewayURL = "APIGatewayURL"
)

func corsMethodResponseParams(api *API) map[string]bool {

	var userDefinedHeaders map[string]interface{}
	if api != nil &&
		api.CORSOptions != nil {
		userDefinedHeaders = api.CORSOptions.Headers
	}
	if len(userDefinedHeaders) <= 0 {
		userDefinedHeaders = defaultCORSHeaders
	}
	responseParams := make(map[string]bool)
	for eachHeader := range userDefinedHeaders {
		keyName := fmt.Sprintf("method.response.header.%s", eachHeader)
		responseParams[keyName] = true
	}
	return responseParams
}

func corsIntegrationResponseParams(api *API) map[string]string {

	var userDefinedHeaders map[string]interface{}
	if api != nil &&
		api.CORSOptions != nil {
		userDefinedHeaders = api.CORSOptions.Headers
	}
	if len(userDefinedHeaders) <= 0 {
		userDefinedHeaders = defaultCORSHeaders
	}
	responseParams := make(map[string]string)
	for eachHeader, eachHeaderValue := range userDefinedHeaders {
		keyName := fmt.Sprintf("method.response.header.%s", eachHeader)
		switch headerVal := eachHeaderValue.(type) {
		case string:
			responseParams[keyName] = gof.Join("", []string{
				"'",
				headerVal,
				"'",
			})
		default:
			responseParams[keyName] = fmt.Sprintf("'%s'", eachHeaderValue)
		}
	}
	return responseParams
}

// DefaultMethodResponses returns the default set of Method HTTPStatus->Response
// pass through responses.  The successfulHTTPStatusCode param is the single
// 2XX response code to use for the method.
func methodResponses(api *API, userResponses map[int]*Response, corsEnabled bool) []gofapig.Method_MethodResponse {

	var responses []gofapig.Method_MethodResponse
	for eachHTTPStatusCode, eachResponse := range userResponses {
		methodResponseParams := eachResponse.Parameters
		if corsEnabled {
			for eachString, eachBool := range corsMethodResponseParams(api) {
				methodResponseParams[eachString] = eachBool
			}
		}
		// Then transform them all to strings because internet
		methodResponse := gofapig.Method_MethodResponse{
			StatusCode: strconv.Itoa(eachHTTPStatusCode),
		}
		if len(methodResponseParams) != 0 {
			methodResponse.ResponseParameters = methodResponseParams
		}
		responses = append(responses, methodResponse)
	}
	return responses
}

func integrationResponses(api *API, userResponses map[int]*IntegrationResponse, corsEnabled bool) []gofapig.Method_IntegrationResponse {

	var integrationResponses []gofapig.Method_IntegrationResponse

	// We've already populated this entire map in the NewMethod call
	for eachHTTPStatusCode, eachMethodIntegrationResponse := range userResponses {
		responseParameters := eachMethodIntegrationResponse.Parameters
		if corsEnabled {
			for eachKey, eachValue := range corsIntegrationResponseParams(api) {
				responseParameters[eachKey] = eachValue
			}
		}

		integrationResponse := gofapig.Method_IntegrationResponse{
			ResponseTemplates: eachMethodIntegrationResponse.Templates,
			SelectionPattern:  eachMethodIntegrationResponse.SelectionPattern,
			StatusCode:        strconv.Itoa(eachHTTPStatusCode),
		}
		if len(responseParameters) != 0 {
			integrationResponse.ResponseParameters = responseParameters
		}
		integrationResponses = append(integrationResponses, integrationResponse)
	}

	return integrationResponses
}

func methodRequestTemplates(method *Method) (map[string]string, error) {
	supportedTemplates := map[string]string{
		"application/json":                  embeddedMustString("resources/provision/apigateway/inputmapping_json.vtl"),
		"text/plain":                        embeddedMustString("resources/provision/apigateway/inputmapping_default.vtl"),
		"application/x-www-form-urlencoded": embeddedMustString("resources/provision/apigateway/inputmapping_formencoded.vtl"),
		"multipart/form-data":               embeddedMustString("resources/provision/apigateway/inputmapping_default.vtl"),
	}
	if len(method.SupportedRequestContentTypes) <= 0 {
		return supportedTemplates, nil
	}

	// Else, let's go ahead and return only the mappings the user wanted
	userDefinedTemplates := make(map[string]string)
	for _, eachContentType := range method.SupportedRequestContentTypes {
		vtlMapping, vtlMappingExists := supportedTemplates[eachContentType]
		if !vtlMappingExists {
			return nil, fmt.Errorf("unsupported method request template Content-Type provided: %s", eachContentType)
		}
		userDefinedTemplates[eachContentType] = vtlMapping
	}
	return userDefinedTemplates, nil
}

func corsOptionsGatewayMethod(api *API, restAPIID string, resourceID string) *gofapig.Method {
	methodResponse := gofapig.Method_MethodResponse{
		StatusCode:         "200",
		ResponseParameters: corsMethodResponseParams(api),
	}

	integrationResponse := gofapig.Method_IntegrationResponse{
		ResponseTemplates: map[string]string{
			"application/*": "",
			"text/*":        "",
		},
		StatusCode:         "200",
		ResponseParameters: corsIntegrationResponseParams(api),
	}

	methodIntegrationIntegrationResponseList := []gofapig.Method_IntegrationResponse{}
	methodIntegrationIntegrationResponseList = append(methodIntegrationIntegrationResponseList,
		integrationResponse)
	methodResponseList := []gofapig.Method_MethodResponse{}
	methodResponseList = append(methodResponseList, methodResponse)

	corsMethod := &gofapig.Method{
		HttpMethod:        "OPTIONS",
		AuthorizationType: "NONE",
		RestApiId:         restAPIID,
		ResourceId:        resourceID,
		Integration: &gofapig.Method_Integration{
			Type: "MOCK",
			RequestTemplates: map[string]string{
				"application/json": "{\"statusCode\": 200}",
				"text/plain":       "statusCode: 200",
			},
			IntegrationResponses: methodIntegrationIntegrationResponseList,
		},
		MethodResponses: methodResponseList,
	}
	return corsMethod
}

func apiStageInfo(apiName string,
	stageName string,
	awsConfig awsv2.Config,
	noop bool,
	logger *zerolog.Logger) (*awsv2APIGTypes.Stage, error) {

	logger.Info().
		Str("APIName", apiName).
		Str("StageName", stageName).
		Msg("Checking current API Gateway stage status")

	if noop {
		logger.Info().Msg(noopMessage("API Gateway check"))
		return nil, nil
	}
	ctxStageInfo := context.Background()
	svc := awsv2APIG.NewFromConfig(awsConfig)
	restApisInput := &awsv2APIG.GetRestApisInput{
		Limit: awsv2.Int32(500),
	}

	restApisOutput, restApisOutputErr := svc.GetRestApis(ctxStageInfo, restApisInput)
	if nil != restApisOutputErr {
		return nil, restApisOutputErr
	}
	// Find the entry that has this name
	restAPIID := ""
	for _, eachRestAPI := range restApisOutput.Items {
		if *eachRestAPI.Name == apiName {
			if restAPIID != "" {
				return nil, fmt.Errorf("multiple RestAPI matches for API Name: %s", apiName)
			}
			restAPIID = *eachRestAPI.Id
		}
	}
	if restAPIID == "" {
		return nil, nil
	}
	// API exists...does the stage name exist?
	stagesInput := &awsv2APIG.GetStagesInput{
		RestApiId: awsv2.String(restAPIID),
	}
	stagesOutput, stagesOutputErr := svc.GetStages(ctxStageInfo, stagesInput)
	if nil != stagesOutputErr {
		return nil, stagesOutputErr
	}

	// Find this stage name...
	var matchingStageOutput *awsv2APIGTypes.Stage
	for _, eachStage := range stagesOutput.Item {
		if *eachStage.StageName == stageName {
			if nil != matchingStageOutput {
				return nil, fmt.Errorf("multiple stage matches for name: %s", stageName)
			}
			matchingStageOutput = &eachStage
		}
	}
	if nil != matchingStageOutput {
		logger.Info().
			Str("DeploymentId", *matchingStageOutput.DeploymentId).
			Time("LastUpdated", *matchingStageOutput.LastUpdatedDate).
			Time("CreatedDate", *matchingStageOutput.CreatedDate).
			Msg("Checking current APIGateway stage status")
	} else {
		logger.Info().Msg("APIGateway stage has not been deployed")
	}
	return matchingStageOutput, nil
}

////////////////////////////////////////////////////////////////////////////////
//

// APIGatewayIdentity represents the user identity of a request
// made on behalf of the API Gateway
type APIGatewayIdentity struct {
	// Account ID
	AccountID string `json:"accountId"`
	// API Key
	APIKey string `json:"apiKey"`
	// Caller
	Caller string `json:"caller"`
	// Cognito Authentication Provider
	CognitoAuthenticationProvider string `json:"cognitoAuthenticationProvider"`
	// Cognito Authentication Type
	CognitoAuthenticationType string `json:"cognitoAuthenticationType"`
	// CognitoIdentityId
	CognitoIdentityID string `json:"cognitoIdentityId"`
	// CognitoIdentityPoolId
	CognitoIdentityPoolID string `json:"cognitoIdentityPoolId"`
	// Source IP
	SourceIP string `json:"sourceIp"`
	// User
	User string `json:"user"`
	// User Agent
	UserAgent string `json:"userAgent"`
	// User ARN
	UserARN string `json:"userArn"`
}

////////////////////////////////////////////////////////////////////////////////
//

// APIGatewayContext represents the context available to an AWS Lambda
// function that is invoked by an API Gateway integration.
type APIGatewayContext struct {
	// API ID
	APIID string `json:"apiId"`
	// HTTPMethod
	Method string `json:"method"`
	// Request ID
	RequestID string `json:"requestId"`
	// Resource ID
	ResourceID string `json:"resourceId"`
	// Resource Path
	ResourcePath string `json:"resourcePath"`
	// Stage
	Stage string `json:"stage"`
	// User identity
	Identity APIGatewayIdentity `json:"identity"`
}

////////////////////////////////////////////////////////////////////////////////
//

// APIGatewayLambdaJSONEvent provides a pass through mapping
// of all whitelisted Parameters.  The transformation is defined
// by the resources/gateway/inputmapping_json.vtl template.
type APIGatewayLambdaJSONEvent struct {
	// HTTPMethod
	Method string `json:"method"`
	// Body, if available
	Body json.RawMessage `json:"body"`
	// Whitelisted HTTP headers
	Headers map[string]string `json:"headers"`
	// Whitelisted HTTP query params
	QueryParams map[string]string `json:"queryParams"`
	// Whitelisted path parameters
	PathParams map[string]string `json:"pathParams"`
	// Context information - http://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-mapping-template-reference.html#context-variable-reference
	Context APIGatewayContext `json:"context"`
}

////////////////////////////////////////////////////////////////////////////////
//

// Model proxies the AWS SDK's Model data.  See
// http://docs.aws.amazon.com/sdk-for-go/api/service/apigateway.html#Model
//
// TODO: Support Dynamic Model creation
type Model struct {
	Description string `json:",omitempty"`
	Name        string `json:",omitempty"`
	Schema      string `json:",omitempty"`
}

////////////////////////////////////////////////////////////////////////////////
//

// Response proxies the AWS SDK's PutMethodResponseInput data.  See
// http://docs.aws.amazon.com/sdk-for-go/api/service/apigateway.html#PutMethodResponseInput
type Response struct {
	Parameters map[string]bool   `json:",omitempty"`
	Models     map[string]*Model `json:",omitempty"`
}

////////////////////////////////////////////////////////////////////////////////
//

// IntegrationResponse proxies the AWS SDK's IntegrationResponse data.  See
// http://docs.aws.amazon.com/sdk-for-go/api/service/apigateway/#IntegrationResponse
type IntegrationResponse struct {
	Parameters       map[string]string `json:",omitempty"`
	SelectionPattern string            `json:",omitempty"`
	Templates        map[string]string `json:",omitempty"`
}

////////////////////////////////////////////////////////////////////////////////
//

// Integration proxies the AWS SDK's Integration data.  See
// http://docs.aws.amazon.com/sdk-for-go/api/service/apigateway.html#Integration
type Integration struct {
	Parameters         map[string]string
	RequestTemplates   map[string]string
	CacheKeyParameters []string
	CacheNamespace     string
	Credentials        string

	Responses map[int]*IntegrationResponse

	// Typically "AWS", but for OPTIONS CORS support is set to "MOCK"
	integrationType string
}

////////////////////////////////////////////////////////////////////////////////
//

// Method proxies the AWS SDK's Method data.  See
// http://docs.aws.amazon.com/sdk-for-go/api/service/apigateway.html#type-Method
type Method struct {
	authorizationID         string
	httpMethod              string
	defaultHTTPResponseCode int

	APIKeyRequired bool

	// Request data
	Parameters map[string]bool
	Models     map[string]*Model

	// Supported HTTP request Content-Types. Used to limit the amount of VTL
	// injected into the CloudFormation template. Eligible values include:
	// application/json
	// text/plain
	// application/x-www-form-urlencoded
	// multipart/form-data
	SupportedRequestContentTypes []string

	// Response map
	Responses map[int]*Response

	// Integration response map
	Integration Integration
}

////////////////////////////////////////////////////////////////////////////////
//

// Resource proxies the AWS SDK's Resource data.  See
// http://docs.aws.amazon.com/sdk-for-go/api/service/apigateway.html#type-Resource
type Resource struct {
	pathPart     string
	parentLambda *LambdaAWSInfo
	Methods      map[string]*Method
}

// Stage proxies the AWS SDK's Stage data.  See
// http://docs.aws.amazon.com/sdk-for-go/api/service/apigateway.html#type-Stage
type Stage struct {
	name                string
	CacheClusterEnabled bool
	CacheClusterSize    string
	Description         string
	Variables           map[string]string
}

////////////////////////////////////////////////////////////////////////////////
//

// CORSOptions is a struct that clients supply to the API in order to enable
// and parameterize CORS API values
type CORSOptions struct {
	// Headers represent the CORS headers that should be used for an OPTIONS
	// preflight request. These should be of the form key-value as in:
	// "Access-Control-Allow-Headers"="Content-Type,X-Amz-Date,Authorization,X-Api-Key"
	Headers map[string]interface{}
}

////////////////////////////////////////////////////////////////////////////////
//

// API represents the AWS API Gateway data associated with a given Sparta app.  Proxies
// the AWS SDK's CreateRestApiInput data.  See
// http://docs.aws.amazon.com/sdk-for-go/api/service/apigateway.html#type-CreateRestApiInput
type API struct {
	// The API name
	// TODO: bind this to the stack name to prevent provisioning collisions.
	name string
	// Optional stage. If defined, the API will be deployed
	stage *Stage
	// Existing API to CloneFrom
	CloneFrom string
	// APIDescription is the user defined description
	Description string
	// Non-empty map of urlPaths->Resource definitions
	resources map[string]*Resource
	// Should CORS be enabled for this API?
	CORSEnabled bool
	// CORS options - if non-nil, supersedes CORSEnabled
	CORSOptions *CORSOptions
	// Endpoint configuration information
	EndpointConfiguration *gofapig.RestApi_EndpointConfiguration
}

// LogicalResourceName returns the CloudFormation logical
// resource name for this API
func (api *API) LogicalResourceName() string {
	return CloudFormationResourceName("APIGateway", api.name)
}

// RestAPIURL returns the dynamically assigned
// Rest API URL including the scheme
func (api *API) RestAPIURL() string {
	return gof.Join("", []string{
		"https://",
		gof.Ref(api.LogicalResourceName()),
		".execute-api.",
		gof.Ref("AWS::Region"),
		".amazonaws.com",
	})
}

func (api *API) corsEnabled() bool {
	return api.CORSEnabled || (api.CORSOptions != nil)
}

// Describe returns the API for description
func (api *API) Describe(targetNodeName string) (*DescriptionInfo, error) {
	descInfo := &DescriptionInfo{
		Name:  "APIGateway",
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

	// Create the APIGateway virtual node && connect it to the application
	for _, eachResource := range api.resources {
		for eachMethod := range eachResource.Methods {
			// Create the PATH node
			var nodeName = fmt.Sprintf("%s - %s", eachMethod, eachResource.pathPart)
			descInfo.Nodes = append(descInfo.Nodes,
				&DescriptionTriplet{
					SourceNodeName: nodeName,
					DisplayInfo: &DescriptionDisplayInfo{
						SourceNodeColor: nodeColorAPIGateway,
						SourceIcon: &DescriptionIcon{
							Category: "_General",
							Name:     "Internet-alt1_light-bg@4x.png",
						},
					},
					TargetNodeName: nodeNameAPIGateway,
				},
				&DescriptionTriplet{
					SourceNodeName: nodeName,
					TargetNodeName: eachResource.parentLambda.lambdaFunctionName(),
				})
		}
	}
	return descInfo, nil
}

// Marshal marshals the API data to a CloudFormation compatible representation
func (api *API) Marshal(serviceName string,
	awsConfig awsv2.Config,
	lambdaFunctionCode *goflambda.Function_Code,
	roleNameMap map[string]string,
	template *gof.Template,
	noop bool,
	logger *zerolog.Logger) error {

	apiGatewayResourceNameForPath := func(fullPath string) string {
		pathParts := strings.Split(fullPath, "/")
		return CloudFormationResourceName("%sResource", pathParts[0], fullPath)
	}

	// Create an API gateway entry
	apiGatewayRes := &gofapig.RestApi{
		Description:    api.Description,
		FailOnWarnings: false,
		Name:           api.name,
	}
	if api.CloneFrom != "" {
		apiGatewayRes.CloneFrom = api.CloneFrom
	}
	if api.Description == "" {
		apiGatewayRes.Description = fmt.Sprintf("%s RestApi", serviceName)
	} else {
		apiGatewayRes.Description = api.Description
	}
	apiGatewayResName := api.LogicalResourceName()
	// Is there an endpoint type?
	if api.EndpointConfiguration != nil {
		apiGatewayRes.EndpointConfiguration = api.EndpointConfiguration
	}
	template.Resources[apiGatewayResName] = apiGatewayRes
	apiGatewayRestAPIID := gof.Ref(apiGatewayResName)

	// List of all the method resources we're creating s.t. the
	// deployment can DependOn them
	optionsMethodPathMap := make(map[string]bool)
	var apiMethodCloudFormationResources []string
	for eachResourceMethodKey, eachResourceDef := range api.resources {
		// First walk all the user resources and create intermediate paths
		// to repreesent all the resources
		var parentResource string
		pathParts := strings.Split(strings.TrimLeft(eachResourceDef.pathPart, "/"), "/")
		pathAccumulator := []string{"/"}
		for index, eachPathPart := range pathParts {
			pathAccumulator = append(pathAccumulator, eachPathPart)
			resourcePathName := apiGatewayResourceNameForPath(strings.Join(pathAccumulator, "/"))
			if _, exists := template.Resources[resourcePathName]; !exists {
				cfResource := &gofapig.Resource{
					RestApiId: apiGatewayRestAPIID,
					PathPart:  eachPathPart,
				}
				if index <= 0 {
					cfResource.ParentId = gof.GetAtt(apiGatewayResName, "RootResourceId")
				} else {
					cfResource.ParentId = parentResource
				}
				template.Resources[resourcePathName] = cfResource
			}
			parentResource = gof.Ref(resourcePathName)
		}

		// Add the lambda permission
		apiGatewayPermissionResourceName := CloudFormationResourceName("APIGatewayLambdaPerm",
			eachResourceMethodKey)
		lambdaInvokePermission := &goflambda.Permission{
			Action:       "lambda:InvokeFunction",
			FunctionName: gof.GetAtt(eachResourceDef.parentLambda.LogicalResourceName(), "Arn"),
			Principal:    APIGatewayPrincipal,
		}
		template.Resources[apiGatewayPermissionResourceName] = lambdaInvokePermission

		// BEGIN CORS - OPTIONS verb
		// CORS is API global, but it's possible that there are multiple different lambda functions
		// that are handling the same HTTP resource. In this case, track whether we've already created an
		// OPTIONS entry for this path and only append iff this is the first time through
		if api.corsEnabled() {
			methodResourceName := CloudFormationResourceName(fmt.Sprintf("%s-OPTIONS",
				eachResourceDef.pathPart), eachResourceDef.pathPart)
			_, resourceExists := optionsMethodPathMap[methodResourceName]
			if !resourceExists {
				template.Resources[methodResourceName] = corsOptionsGatewayMethod(api,
					apiGatewayRestAPIID,
					parentResource)
				apiMethodCloudFormationResources = append(apiMethodCloudFormationResources, methodResourceName)
				optionsMethodPathMap[methodResourceName] = true
			}
		}
		// END CORS - OPTIONS verb

		// BEGIN - user defined verbs
		for eachMethodName, eachMethodDef := range eachResourceDef.Methods {

			methodRequestTemplates, methodRequestTemplatesErr := methodRequestTemplates(eachMethodDef)
			if methodRequestTemplatesErr != nil {
				return methodRequestTemplatesErr
			}
			apiGatewayMethod := &gofapig.Method{
				HttpMethod: eachMethodName,
				ResourceId: parentResource,
				RestApiId:  apiGatewayRestAPIID,
				Integration: &gofapig.Method_Integration{
					IntegrationHttpMethod: "POST",
					Type:                  "AWS",
					RequestTemplates:      methodRequestTemplates,
					Uri: gof.Join("", []string{
						"arn:aws:apigateway:",
						gof.Ref("AWS::Region"),
						":lambda:path/2015-03-31/functions/",
						gof.GetAtt(eachResourceDef.parentLambda.LogicalResourceName(), "Arn"),
						"/invocations",
					}),
				},
			}
			// Handle authorization
			if eachMethodDef.authorizationID != "" {
				// See https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-apigateway-method.html#cfn-apigateway-method-authorizationtype
				apiGatewayMethod.AuthorizationType = "CUSTOM"
				apiGatewayMethod.AuthorizerId = eachMethodDef.authorizationID
			} else {
				apiGatewayMethod.AuthorizationType = "NONE"
			}
			if len(eachMethodDef.Parameters) != 0 {
				apiGatewayMethod.RequestParameters = eachMethodDef.Parameters
			}

			// Add the integration response RegExps
			apiGatewayMethod.Integration.IntegrationResponses = integrationResponses(api,
				eachMethodDef.Integration.Responses,
				api.corsEnabled())

			// Add outbound method responses
			apiGatewayMethod.MethodResponses = methodResponses(api,
				eachMethodDef.Responses,
				api.corsEnabled())

			prefix := fmt.Sprintf("%s%s", eachMethodDef.httpMethod, eachResourceMethodKey)
			methodResourceName := CloudFormationResourceName(prefix, eachResourceMethodKey, serviceName)
			apiGatewayMethod.AWSCloudFormationDependsOn = []string{
				apiGatewayPermissionResourceName,
			}
			template.Resources[methodResourceName] = apiGatewayMethod

			apiMethodCloudFormationResources = append(apiMethodCloudFormationResources,
				methodResourceName)
		}
	}
	// END
	if nil != api.stage {
		// Is the stack already deployed?
		stageName := api.stage.name
		stageInfo, stageInfoErr := apiStageInfo(api.name,
			stageName,
			awsConfig,
			noop,
			logger)
		if nil != stageInfoErr {
			return stageInfoErr
		}
		if nil == stageInfo {
			// Use a stable identifier so that we can update the existing deployment
			apiDeploymentResName := CloudFormationResourceName("APIGatewayDeployment",
				serviceName)
			apiDeployment := &gofapig.Deployment{
				Description: api.stage.Description,
				RestApiId:   apiGatewayRestAPIID,
				StageName:   stageName,
				StageDescription: &gofapig.Deployment_StageDescription{
					Description: api.stage.Description,
					Variables:   api.stage.Variables,
				},
			}
			if api.stage.CacheClusterEnabled {
				apiDeployment.StageDescription.CacheClusterEnabled =
					api.stage.CacheClusterEnabled
			}
			if api.stage.CacheClusterSize != "" {
				apiDeployment.StageDescription.CacheClusterSize =
					api.stage.CacheClusterSize
			}
			apiDeployment.AWSCloudFormationDependsOn = apiMethodCloudFormationResources
			apiDeployment.AWSCloudFormationDependsOn = append(apiDeployment.AWSCloudFormationDependsOn,
				apiGatewayResName)

			template.Resources[apiDeploymentResName] = apiDeployment

		} else {
			newDeployment := &gofapig.Deployment{
				Description: "Deployment",
				RestApiId:   apiGatewayRestAPIID,
			}
			if stageInfo.StageName != nil {
				newDeployment.StageName = *stageInfo.StageName
			}
			// Use an unstable ID s.t. we can actually create a new deployment event.  Not sure how this
			// is going to work with deletes...
			deploymentResName := CloudFormationResourceName("APIGatewayDeployment")

			newDeployment.AWSCloudFormationDependsOn = apiMethodCloudFormationResources
			newDeployment.AWSCloudFormationDependsOn = append(newDeployment.AWSCloudFormationDependsOn,
				apiGatewayResName)
			template.Resources[deploymentResName] = newDeployment
		}
		// Outputs...
		template.Outputs[OutputAPIGatewayURL] = gof.Output{
			Description: "API Gateway URL",
			Value: gof.Join("", []string{
				"https://",
				apiGatewayRestAPIID,
				".execute-api.",
				gof.Ref("AWS::Region"),
				".amazonaws.com/",
				stageName,
			}),
		}
	}
	return nil
}

// NewAPIGateway returns a new API Gateway structure.  If stage is defined, the API Gateway
// will also be deployed as part of stack creation.
func NewAPIGateway(name string, stage *Stage) *API {
	return &API{
		name:        name,
		stage:       stage,
		resources:   make(map[string]*Resource),
		CORSEnabled: false,
		CORSOptions: nil,
	}
}

// NewStage returns a Stage object with the given name.  Providing a Stage value
// to NewAPIGateway implies that the API Gateway resources should be deployed
// (eg: made publicly accessible).  See
// http://docs.aws.amazon.com/apigateway/latest/developerguide/how-to-deploy-api.html
func NewStage(name string) *Stage {
	return &Stage{
		name:      name,
		Variables: make(map[string]string),
	}
}

// NewResource associates a URL path value with the LambdaAWSInfo golang lambda.  To make
// the Resource available, associate one or more Methods via NewMethod().
func (api *API) NewResource(pathPart string, parentLambda *LambdaAWSInfo) (*Resource, error) {
	// The key is the path+resource, since we want to support POLA scoped
	// security roles based on HTTP method
	resourcesKey := fmt.Sprintf("%s%s", parentLambda.lambdaFunctionName(), pathPart)
	_, exists := api.resources[resourcesKey]
	if exists {
		return nil, fmt.Errorf("path %s already defined for lambda function: %s", pathPart, parentLambda.lambdaFunctionName())
	}
	resource := &Resource{
		pathPart:     pathPart,
		parentLambda: parentLambda,
		Methods:      make(map[string]*Method),
	}
	api.resources[resourcesKey] = resource
	return resource, nil
}

// NewMethod associates the httpMethod name with the given Resource.  The returned Method
// has no authorization requirements. To limit the amount of API gateway resource mappings,
// supply the variadic slice of  possibleHTTPStatusCodeResponses which is the universe
// of all HTTP status codes returned by your Sparta function. If this slice is non-empty,
// Sparta will *ONLY* generate mappings for known codes. This slice need only include the
// codes in addition to the defaultHTTPStatusCode. If the function can only return a single
// value, provide the defaultHTTPStatusCode in the possibleHTTPStatusCodeResponses slice
func (resource *Resource) NewMethod(httpMethod string,
	defaultHTTPStatusCode int,
	possibleHTTPStatusCodeResponses ...int) (*Method, error) {

	if OptionsGlobal.Logger != nil && len(possibleHTTPStatusCodeResponses) != 0 {
		OptionsGlobal.Logger.Debug().Interface(
			"possibleHTTPStatusCodeResponses", possibleHTTPStatusCodeResponses).
			Msg("The set of all HTTP status codes is no longer required for NewMethod(...). Any valid HTTP status code can be returned starting with v1.8.0.")
	}

	// http://docs.aws.amazon.com/apigateway/latest/developerguide/how-to-method-settings.html#how-to-method-settings-console
	keyname := httpMethod
	existingMethod, exists := resource.Methods[keyname]
	if exists {
		return nil, fmt.Errorf("method %s (Auth: %#v) already defined for resource",
			httpMethod,
			existingMethod.authorizationID)
	}
	if defaultHTTPStatusCode == 0 {
		return nil, fmt.Errorf("invalid default HTTP status (%d) code for method", defaultHTTPStatusCode)
	}

	integration := Integration{
		Parameters:       make(map[string]string),
		RequestTemplates: make(map[string]string),
		Responses:        make(map[int]*IntegrationResponse),
		integrationType:  "AWS", // Type used for Lambda integration
	}

	method := &Method{
		httpMethod:              httpMethod,
		defaultHTTPResponseCode: defaultHTTPStatusCode,
		Parameters:              make(map[string]bool),
		Models:                  make(map[string]*Model),
		Responses:               make(map[int]*Response),
		Integration:             integration,
	}

	// Eligible HTTP status codes...
	if len(possibleHTTPStatusCodeResponses) <= 0 {
		// User didn't supply any potential codes, so use the entire set...
		for i := http.StatusOK; i <= http.StatusNetworkAuthenticationRequired; i++ {
			if len(http.StatusText(i)) != 0 {
				possibleHTTPStatusCodeResponses = append(possibleHTTPStatusCodeResponses, i)
			}
		}
	} else {
		// There are some, so include them, plus the default one
		possibleHTTPStatusCodeResponses = append(possibleHTTPStatusCodeResponses,
			defaultHTTPStatusCode)
	}

	// So we need to return everything here, but that means we'll need some other
	// place to mutate the response body...where?
	templateString, templateStringErr := embeddedString("resources/provision/apigateway/outputmapping_json.vtl")

	// Ignore any error when running in AWS, since that version of the binary won't
	// have the embedded asset. This ideally would be done only when we're exporting
	// the Method, but that would involve changing caller behavior since
	// callers currently expect the method.Integration.Responses to be populated
	// when this constructor returns.
	if templateStringErr != nil {
		templateString = embeddedMustString("resources/awsbinary/README.md")
	}

	// TODO - tell the caller that we don't need the list of all HTTP status
	// codes anymore since we've moved everything to overrides in the VTL mapping.

	// Populate Integration.Responses and the method Parameters
	for _, i := range possibleHTTPStatusCodeResponses {
		statusText := http.StatusText(i)
		if statusText == "" {
			return nil, fmt.Errorf("invalid HTTP status code %d provided for method: %s",
				i,
				httpMethod)
		}

		// The integration responses are keyed from supported error codes...
		if defaultHTTPStatusCode == i {
			// Since we pushed this into the VTL mapping, we don't need to create explicit RegExp based
			// mappings for all of the user response codes. It will just work.
			// Ref: https://docs.aws.amazon.com/apigateway/latest/developerguide/handle-errors-in-lambda-integration.html
			method.Integration.Responses[i] = &IntegrationResponse{
				Parameters: make(map[string]string),
				Templates: map[string]string{
					"application/json": templateString,
					"text/*":           "",
				},
				SelectionPattern: "",
			}
		}

		// Then the Method.Responses
		method.Responses[i] = &Response{
			Parameters: make(map[string]bool),
			Models:     make(map[string]*Model),
		}
	}
	resource.Methods[keyname] = method
	return method, nil
}

// NewAuthorizedMethod associates the httpMethod name and authorizationID with
// the given Resource. The authorizerID param is a cloudformation.Strinable
// satisfying value
func (resource *Resource) NewAuthorizedMethod(httpMethod string,
	authorizerID string,
	defaultHTTPStatusCode int,
	possibleHTTPStatusCodeResponses ...int) (*Method, error) {
	if authorizerID == "" {
		return nil, fmt.Errorf("authorizerID must not be `nil` for Authorized Method")
	}
	method, methodErr := resource.NewMethod(httpMethod,
		defaultHTTPStatusCode,
		possibleHTTPStatusCodeResponses...)
	if methodErr == nil {
		method.authorizationID = authorizerID
	}
	return method, methodErr
}
