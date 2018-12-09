package sparta

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigateway"
	gocf "github.com/mweagle/go-cloudformation"
	"github.com/sirupsen/logrus"
)

/*
"context" : {
  "apiId" : "$util.escapeJavaScript($context.apiId)",
  "method" : "$util.escapeJavaScript($context.httpMethod)",
  "requestId" : "$util.escapeJavaScript($context.requestId)",
  "resourceId" : "$util.escapeJavaScript($context.resourceId)",
  "resourcePath" : "$util.escapeJavaScript($context.resourcePath)",
  "stage" : "$util.escapeJavaScript($context.stage)",
  "identity" : {
    "accountId" : "$util.escapeJavaScript($context.identity.accountId)",
    "apiKey" : "$util.escapeJavaScript($context.identity.apiKey)",
    "caller" : "$util.escapeJavaScript($context.identity.caller)",
    "cognitoAuthenticationProvider" : "$util.escapeJavaScript($context.identity.cognitoAuthenticationProvider)",
    "cognitoAuthenticationType" : "$util.escapeJavaScript($context.identity.cognitoAuthenticationType)",
    "cognitoIdentityId" : "$util.escapeJavaScript($context.identity.cognitoIdentityId)",
    "cognitoIdentityPoolId" : "$util.escapeJavaScript($context.identity.cognitoIdentityPoolId)",
    "sourceIp" : "$util.escapeJavaScript($context.identity.sourceIp)",
    "user" : "$util.escapeJavaScript($context.identity.user)",
    "userAgent" : "$util.escapeJavaScript($context.identity.userAgent)",
    "userArn" : "$util.escapeJavaScript($context.identity.userArn)"
  }
*/

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

func corsIntegrationResponseParams(api *API) map[string]interface{} {

	var userDefinedHeaders map[string]interface{}
	if api != nil &&
		api.CORSOptions != nil {
		userDefinedHeaders = api.CORSOptions.Headers
	}
	if len(userDefinedHeaders) <= 0 {
		userDefinedHeaders = defaultCORSHeaders
	}
	responseParams := make(map[string]interface{})
	for eachHeader, eachHeaderValue := range userDefinedHeaders {
		keyName := fmt.Sprintf("method.response.header.%s", eachHeader)
		switch headerVal := eachHeaderValue.(type) {
		case *gocf.StringExpr:
			responseParams[keyName] = gocf.Join("",
				gocf.String("'"),
				headerVal.String(),
				gocf.String("'"))
		default:
			responseParams[keyName] = fmt.Sprintf("'%s'", eachHeaderValue)
		}
	}
	return responseParams
}

// DefaultMethodResponses returns the default set of Method HTTPStatus->Response
// pass through responses.  The successfulHTTPStatusCode param is the single
// 2XX response code to use for the method.
func methodResponses(api *API, userResponses map[int]*Response, corsEnabled bool) *gocf.APIGatewayMethodMethodResponseList {

	var responses gocf.APIGatewayMethodMethodResponseList
	for eachHTTPStatusCode, eachResponse := range userResponses {
		methodResponseParams := eachResponse.Parameters
		if corsEnabled {
			for eachString, eachBool := range corsMethodResponseParams(api) {
				methodResponseParams[eachString] = eachBool
			}
		}
		// Then transform them all to strings because internet
		methodResponseStringParams := make(map[string]string, len(methodResponseParams))
		for eachKey, eachBool := range methodResponseParams {
			methodResponseStringParams[eachKey] = fmt.Sprintf("%t", eachBool)
		}
		methodResponse := gocf.APIGatewayMethodMethodResponse{
			StatusCode: gocf.String(strconv.Itoa(eachHTTPStatusCode)),
		}
		if len(methodResponseStringParams) != 0 {
			methodResponse.ResponseParameters = methodResponseStringParams
		}
		responses = append(responses, methodResponse)
	}
	return &responses
}

func integrationResponses(api *API, userResponses map[int]*IntegrationResponse,
	corsEnabled bool) *gocf.APIGatewayMethodIntegrationResponseList {

	var integrationResponses gocf.APIGatewayMethodIntegrationResponseList

	// We've already populated this entire map in the NewMethod call
	for eachHTTPStatusCode, eachMethodIntegrationResponse := range userResponses {
		responseParameters := eachMethodIntegrationResponse.Parameters
		if corsEnabled {
			for eachKey, eachValue := range corsIntegrationResponseParams(api) {
				responseParameters[eachKey] = eachValue
			}
		}

		integrationResponse := gocf.APIGatewayMethodIntegrationResponse{
			ResponseTemplates: eachMethodIntegrationResponse.Templates,
			SelectionPattern:  gocf.String(eachMethodIntegrationResponse.SelectionPattern),
			StatusCode:        gocf.String(strconv.Itoa(eachHTTPStatusCode)),
		}
		if len(responseParameters) != 0 {
			integrationResponse.ResponseParameters = responseParameters
		}
		integrationResponses = append(integrationResponses, integrationResponse)
	}

	return &integrationResponses
}

func methodRequestTemplates(method *Method) (map[string]string, error) {
	supportedTemplates := map[string]string{
		"application/json":                  _escFSMustString(false, "/resources/provision/apigateway/inputmapping_json.vtl"),
		"text/plain":                        _escFSMustString(false, "/resources/provision/apigateway/inputmapping_default.vtl"),
		"application/x-www-form-urlencoded": _escFSMustString(false, "/resources/provision/apigateway/inputmapping_formencoded.vtl"),
		"multipart/form-data":               _escFSMustString(false, "/resources/provision/apigateway/inputmapping_default.vtl"),
	}
	if len(method.SupportedRequestContentTypes) <= 0 {
		return supportedTemplates, nil
	}

	// Else, let's go ahead and return only the mappings the user wanted
	userDefinedTemplates := make(map[string]string)
	for _, eachContentType := range method.SupportedRequestContentTypes {
		vtlMapping, vtlMappingExists := supportedTemplates[eachContentType]
		if !vtlMappingExists {
			return nil, fmt.Errorf("Unsupported method request template Content-Type provided: %s", eachContentType)
		}
		userDefinedTemplates[eachContentType] = vtlMapping
	}
	return userDefinedTemplates, nil
}

func corsOptionsGatewayMethod(api *API, restAPIID gocf.Stringable, resourceID gocf.Stringable) *gocf.APIGatewayMethod {
	methodResponse := gocf.APIGatewayMethodMethodResponse{
		StatusCode:         gocf.String("200"),
		ResponseParameters: corsMethodResponseParams(api),
	}

	integrationResponse := gocf.APIGatewayMethodIntegrationResponse{
		ResponseTemplates: map[string]string{
			"application/*": "",
			"text/*":        "",
		},
		StatusCode:         gocf.String("200"),
		ResponseParameters: corsIntegrationResponseParams(api),
	}

	methodIntegrationIntegrationResponseList := gocf.APIGatewayMethodIntegrationResponseList{}
	methodIntegrationIntegrationResponseList = append(methodIntegrationIntegrationResponseList,
		integrationResponse)
	methodResponseList := gocf.APIGatewayMethodMethodResponseList{}
	methodResponseList = append(methodResponseList, methodResponse)

	corsMethod := &gocf.APIGatewayMethod{
		HTTPMethod:        gocf.String("OPTIONS"),
		AuthorizationType: gocf.String("NONE"),
		RestAPIID:         restAPIID.String(),
		ResourceID:        resourceID.String(),
		Integration: &gocf.APIGatewayMethodIntegration{
			Type: gocf.String("MOCK"),
			RequestTemplates: map[string]string{
				"application/json": "{\"statusCode\": 200}",
				"text/plain":       "statusCode: 200",
			},
			IntegrationResponses: &methodIntegrationIntegrationResponseList,
		},
		MethodResponses: &methodResponseList,
	}
	return corsMethod
}

func apiStageInfo(apiName string,
	stageName string,
	session *session.Session,
	noop bool,
	logger *logrus.Logger) (*apigateway.Stage, error) {

	logger.WithFields(logrus.Fields{
		"APIName":   apiName,
		"StageName": stageName,
	}).Info("Checking current API Gateway stage status")

	if noop {
		logger.Info(noopMessage("API Gateway check"))
		return nil, nil
	}

	svc := apigateway.New(session)
	restApisInput := &apigateway.GetRestApisInput{
		Limit: aws.Int64(500),
	}

	restApisOutput, restApisOutputErr := svc.GetRestApis(restApisInput)
	if nil != restApisOutputErr {
		return nil, restApisOutputErr
	}
	// Find the entry that has this name
	restAPIID := ""
	for _, eachRestAPI := range restApisOutput.Items {
		if *eachRestAPI.Name == apiName {
			if restAPIID != "" {
				return nil, fmt.Errorf("Multiple RestAPI matches for API Name: %s", apiName)
			}
			restAPIID = *eachRestAPI.Id
		}
	}
	if "" == restAPIID {
		return nil, nil
	}
	// API exists...does the stage name exist?
	stagesInput := &apigateway.GetStagesInput{
		RestApiId: aws.String(restAPIID),
	}
	stagesOutput, stagesOutputErr := svc.GetStages(stagesInput)
	if nil != stagesOutputErr {
		return nil, stagesOutputErr
	}

	// Find this stage name...
	var matchingStageOutput *apigateway.Stage
	for _, eachStage := range stagesOutput.Item {
		if *eachStage.StageName == stageName {
			if nil != matchingStageOutput {
				return nil, fmt.Errorf("Multiple stage matches for name: %s", stageName)
			}
			matchingStageOutput = eachStage
		}
	}
	if nil != matchingStageOutput {
		logger.WithFields(logrus.Fields{
			"DeploymentId": *matchingStageOutput.DeploymentId,
			"LastUpdated":  matchingStageOutput.LastUpdatedDate,
			"CreatedDate":  matchingStageOutput.CreatedDate,
		}).Info("Checking current APIGateway stage status")
	} else {
		logger.Info("APIGateway stage has not been deployed")
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
	Parameters       map[string]interface{} `json:",omitempty"`
	SelectionPattern string                 `json:",omitempty"`
	Templates        map[string]string      `json:",omitempty"`
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
	authorizationID         gocf.Stringable
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
	// API Description
	Description string
	// Non-empty map of urlPaths->Resource definitions
	resources map[string]*Resource
	// Should CORS be enabled for this API?
	CORSEnabled bool
	// CORS options - if non-nil, supersedes CORSEnabled
	CORSOptions *CORSOptions
}

// LogicalResourceName returns the CloudFormation logical
// resource name for this API
func (api *API) LogicalResourceName() string {
	return CloudFormationResourceName("APIGateway", api.name)
}

// RestAPIURL returns the dynamically assigned
// Rest API URL including the scheme
func (api *API) RestAPIURL() *gocf.StringExpr {
	return gocf.Join("",
		gocf.String("https://"),
		gocf.Ref(api.LogicalResourceName()),
		gocf.String(".execute-api."),
		gocf.Ref("AWS::Region"),
		gocf.String(".amazonaws.com"))
}

func (api *API) corsEnabled() bool {
	return api.CORSEnabled || (api.CORSOptions != nil)
}

// export marshals the API data to a CloudFormation compatible representation
func (api *API) export(serviceName string,
	session *session.Session,
	S3Bucket string,
	S3Key string,
	S3Version string,
	roleNameMap map[string]*gocf.StringExpr,
	template *gocf.Template,
	noop bool,
	logger *logrus.Logger) error {

	apiGatewayResourceNameForPath := func(fullPath string) string {
		pathParts := strings.Split(fullPath, "/")
		return CloudFormationResourceName("%sResource", pathParts[0], fullPath)
	}

	// Create an API gateway entry
	apiGatewayRes := &gocf.APIGatewayRestAPI{
		Description:    gocf.String(api.Description),
		FailOnWarnings: gocf.Bool(false),
		Name:           gocf.String(api.name),
	}
	if "" != api.CloneFrom {
		apiGatewayRes.CloneFrom = gocf.String(api.CloneFrom)
	}
	if "" == api.Description {
		apiGatewayRes.Description = gocf.String(fmt.Sprintf("%s RestApi", serviceName))
	} else {
		apiGatewayRes.Description = gocf.String(api.Description)
	}
	apiGatewayResName := api.LogicalResourceName()
	template.AddResource(apiGatewayResName, apiGatewayRes)
	apiGatewayRestAPIID := gocf.Ref(apiGatewayResName)

	// List of all the method resources we're creating s.t. the
	// deployment can DependOn them
	optionsMethodPathMap := make(map[string]bool)
	var apiMethodCloudFormationResources []string
	for eachResourceMethodKey, eachResourceDef := range api.resources {
		// First walk all the user resources and create intermediate paths
		// to repreesent all the resources
		var parentResource *gocf.StringExpr
		pathParts := strings.Split(strings.TrimLeft(eachResourceDef.pathPart, "/"), "/")
		pathAccumulator := []string{"/"}
		for index, eachPathPart := range pathParts {
			pathAccumulator = append(pathAccumulator, eachPathPart)
			resourcePathName := apiGatewayResourceNameForPath(strings.Join(pathAccumulator, "/"))
			if _, exists := template.Resources[resourcePathName]; !exists {
				cfResource := &gocf.APIGatewayResource{
					RestAPIID: apiGatewayRestAPIID.String(),
					PathPart:  gocf.String(eachPathPart),
				}
				if index <= 0 {
					cfResource.ParentID = gocf.GetAtt(apiGatewayResName, "RootResourceId")
				} else {
					cfResource.ParentID = parentResource
				}
				template.AddResource(resourcePathName, cfResource)
			}
			parentResource = gocf.Ref(resourcePathName).String()
		}

		// Add the lambda permission
		apiGatewayPermissionResourceName := CloudFormationResourceName("APIGatewayLambdaPerm",
			eachResourceMethodKey)
		lambdaInvokePermission := &gocf.LambdaPermission{
			Action:       gocf.String("lambda:InvokeFunction"),
			FunctionName: gocf.GetAtt(eachResourceDef.parentLambda.LogicalResourceName(), "Arn"),
			Principal:    gocf.String(APIGatewayPrincipal),
		}
		template.AddResource(apiGatewayPermissionResourceName, lambdaInvokePermission)

		// BEGIN CORS - OPTIONS verb
		// CORS is API global, but it's possible that there are multiple different lambda functions
		// that are handling the same HTTP resource. In this case, track whether we've already created an
		// OPTIONS entry for this path and only append iff this is the first time through
		if api.corsEnabled() {
			methodResourceName := CloudFormationResourceName(fmt.Sprintf("%s-OPTIONS",
				eachResourceDef.pathPart), eachResourceDef.pathPart)
			_, resourceExists := optionsMethodPathMap[methodResourceName]
			if !resourceExists {
				template.AddResource(methodResourceName, corsOptionsGatewayMethod(api,
					apiGatewayRestAPIID,
					parentResource))
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
			apiGatewayMethod := &gocf.APIGatewayMethod{
				HTTPMethod: gocf.String(eachMethodName),
				ResourceID: parentResource.String(),
				RestAPIID:  apiGatewayRestAPIID.String(),
				Integration: &gocf.APIGatewayMethodIntegration{
					IntegrationHTTPMethod: gocf.String("POST"),
					Type:                  gocf.String("AWS"),
					RequestTemplates:      methodRequestTemplates,
					URI: gocf.Join("",
						gocf.String("arn:aws:apigateway:"),
						gocf.Ref("AWS::Region"),
						gocf.String(":lambda:path/2015-03-31/functions/"),
						gocf.GetAtt(eachResourceDef.parentLambda.LogicalResourceName(), "Arn"),
						gocf.String("/invocations")),
				},
			}
			// Handle authorization
			if eachMethodDef.authorizationID != nil {
				// See https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-apigateway-method.html#cfn-apigateway-method-authorizationtype
				apiGatewayMethod.AuthorizationType = gocf.String("CUSTOM")
				apiGatewayMethod.AuthorizerID = eachMethodDef.authorizationID.String()
			} else {
				apiGatewayMethod.AuthorizationType = gocf.String("NONE")
			}
			if len(eachMethodDef.Parameters) != 0 {
				requestParams := make(map[string]string)
				for eachKey, eachBool := range eachMethodDef.Parameters {
					requestParams[eachKey] = fmt.Sprintf("%t", eachBool)
				}
				apiGatewayMethod.RequestParameters = requestParams
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
			res := template.AddResource(methodResourceName, apiGatewayMethod)
			res.DependsOn = append(res.DependsOn, apiGatewayPermissionResourceName)
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
			session,
			noop,
			logger)
		if nil != stageInfoErr {
			return stageInfoErr
		}
		if nil == stageInfo {
			// Use a stable identifier so that we can update the existing deployment
			apiDeploymentResName := CloudFormationResourceName("APIGatewayDeployment",
				serviceName)
			apiDeployment := &gocf.APIGatewayDeployment{
				Description: gocf.String(api.stage.Description),
				RestAPIID:   apiGatewayRestAPIID.String(),
				StageName:   gocf.String(stageName),
				StageDescription: &gocf.APIGatewayDeploymentStageDescription{
					Description: gocf.String(api.stage.Description),
					Variables:   api.stage.Variables,
				},
			}
			if api.stage.CacheClusterEnabled {
				apiDeployment.StageDescription.CacheClusterEnabled =
					gocf.Bool(api.stage.CacheClusterEnabled)
			}
			if api.stage.CacheClusterSize != "" {
				apiDeployment.StageDescription.CacheClusterSize =
					gocf.String(api.stage.CacheClusterSize)
			}
			deployment := template.AddResource(apiDeploymentResName, apiDeployment)
			deployment.DependsOn = append(deployment.DependsOn, apiMethodCloudFormationResources...)
			deployment.DependsOn = append(deployment.DependsOn, apiGatewayResName)
		} else {
			newDeployment := &gocf.APIGatewayDeployment{
				Description: gocf.String("Deployment"),
				RestAPIID:   apiGatewayRestAPIID.String(),
			}
			if stageInfo.StageName != nil {
				newDeployment.StageName = gocf.String(*stageInfo.StageName)
			}
			// Use an unstable ID s.t. we can actually create a new deployment event.  Not sure how this
			// is going to work with deletes...
			deploymentResName := CloudFormationResourceName("APIGatewayDeployment")
			deployment := template.AddResource(deploymentResName, newDeployment)
			deployment.DependsOn = append(deployment.DependsOn, apiMethodCloudFormationResources...)
			deployment.DependsOn = append(deployment.DependsOn, apiGatewayResName)
		}
		template.Outputs[OutputAPIGatewayURL] = &gocf.Output{
			Description: "API Gateway URL",
			Value: gocf.Join("",
				gocf.String("https://"),
				apiGatewayRestAPIID,
				gocf.String(".execute-api."),
				gocf.Ref("AWS::Region"),
				gocf.String(".amazonaws.com/"),
				gocf.String(stageName)),
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
		return nil, fmt.Errorf("Path %s already defined for lambda function: %s", pathPart, parentLambda.lambdaFunctionName())
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

	// http://docs.aws.amazon.com/apigateway/latest/developerguide/how-to-method-settings.html#how-to-method-settings-console
	keyname := httpMethod
	existingMethod, exists := resource.Methods[keyname]
	if exists {
		return nil, fmt.Errorf("Method %s (Auth: %#v) already defined for resource",
			httpMethod,
			existingMethod.authorizationID)
	}
	if 0 == defaultHTTPStatusCode {
		return nil, fmt.Errorf("Invalid default HTTP status (%d) code for method", defaultHTTPStatusCode)
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

	// Populate Integration.Responses and the method Parameters
	for _, i := range possibleHTTPStatusCodeResponses {
		statusText := http.StatusText(i)
		if "" == statusText {
			return nil, fmt.Errorf("Invalid HTTP status code %d provided for method: %s",
				i,
				httpMethod)
		}

		// The integration responses are keyed from supported error codes...
		// First the Integration Responses...
		regExp := fmt.Sprintf(`"code"\w*:\w*%d`, i)
		if defaultHTTPStatusCode == i {
			regExp = ""
		}

		// So we need to return everything here, but that means we'll need some other
		// place to mutate the response body...where?

		// Ref: https://docs.aws.amazon.com/apigateway/latest/developerguide/handle-errors-in-lambda-integration.html
		method.Integration.Responses[i] = &IntegrationResponse{
			Parameters: make(map[string]interface{}),
			Templates: map[string]string{
				"application/json": _escFSMustString(false, "/resources/provision/apigateway/outputmapping_json.vtl"),
				"text/*":           "",
			},
			SelectionPattern: regExp,
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
	authorizerID gocf.Stringable,
	defaultHTTPStatusCode int,
	possibleHTTPStatusCodeResponses ...int) (*Method, error) {
	if authorizerID == nil {
		return nil, fmt.Errorf("AuthorizerID must not be `nil` for Authorized Method")
	}
	method, methodErr := resource.NewMethod(httpMethod,
		defaultHTTPStatusCode,
		possibleHTTPStatusCodeResponses...)
	if methodErr == nil {
		method.authorizationID = authorizerID
	}
	return method, methodErr
}
