package sparta

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
)

type Model struct {
	Description string `json:",omitempty"`
	Name        string `json:",omitempty"`
	Schema      string `json:",omitempty"`
}

type Response struct {
	Parameters map[string]bool  `json:",omitempty"`
	Models     map[string]Model `json:",omitempty"`
}

type IntegrationResponse struct {
	Parameters       map[string]string `json:",omitempty"`
	SelectionPattern string            `json:",omitempty"`
	Templates        map[string]string `json:",omitempty"`
}

type Integration struct {
	Parameters         map[string]string
	CacheKeyParameters []string
	CacheNamespace     string
	Credentials        string
	Responses          map[int]IntegrationResponse
}

func (integration Integration) defaultIntegrationResponse() IntegrationResponse {
	return IntegrationResponse{
		Templates: map[string]string{
			"application/json": "",
			"text/plain":       "",
		},
	}
}

func (integration Integration) MarshalJSON() ([]byte, error) {
	var responses = integration.Responses
	if len(responses) <= 0 {
		responses[http.StatusOK] = integration.defaultIntegrationResponse()
	}

	for eachStatusCode, _ := range responses {
		httpString := http.StatusText(eachStatusCode)
		if "" == httpString {
			return nil, fmt.Errorf("Invalid HTTP status code in Integration Response: %d", eachStatusCode)
		}
	}

	var stringResponses = make(map[string]IntegrationResponse, 0)
	for eachKey, eachValue := range responses {
		stringResponses[strconv.Itoa(eachKey)] = eachValue
	}
	integrationJSON := map[string]interface{}{
		"Responses": stringResponses,
	}
	if len(integration.Parameters) > 0 {
		integrationJSON["Parameters"] = integration.Parameters
	}
	if len(integration.CacheNamespace) > 0 {
		integrationJSON["CacheNamespace"] = integration.CacheNamespace
	}
	if len(integration.Credentials) > 0 {
		integrationJSON["Credentials"] = integration.Credentials
	}
	if len(integration.CacheKeyParameters) > 0 {
		integrationJSON["CacheKeyParameters"] = integration.CacheKeyParameters
	}
	return json.Marshal(integrationJSON)
}

type Method struct {
	authorizationType string
	httpMethod        string
	APIKeyRequired    bool

	// Request data
	Parameters map[string]bool
	Models     map[string]Model

	// Response map
	Responses map[int]Response

	// Integration response map
	Integration Integration
}

func (method Method) defaultResponse() Response {
	contentTypes := []string{"application/json", "text/plain"}
	models := make(map[string]Model, 0)
	for _, eachContentType := range contentTypes {
		description := "Empty model"
		if eachContentType == "application/json" {
			description = "Empty JSON model"
		} else {
			parts := strings.Split(eachContentType, "/")
			if len(parts) == 2 {
				description = fmt.Sprintf("Empty %s model", strings.ToUpper(parts[0]))
			}
		}
		models[eachContentType] = Model{
			Description: description,
			Name:        "Empty",
			Schema:      "",
		}
	}
	return Response{
		Models: models,
	}
}

func (method Method) MarshalJSON() ([]byte, error) {
	responses := method.Responses
	if len(responses) <= 0 {
		responses[http.StatusOK] = method.defaultResponse()
	}
	for eachStatusCode, _ := range responses {
		httpString := http.StatusText(eachStatusCode)
		if "" == httpString {
			return nil, fmt.Errorf("Invalid HTTP status code in Method Response: %d", eachStatusCode)
		}
	}

	var stringResponses = make(map[string]Response, 0)
	for eachKey, eachValue := range responses {
		stringResponses[strconv.Itoa(eachKey)] = eachValue
	}

	return json.Marshal(map[string]interface{}{
		"AuthorizationType": method.authorizationType,
		"HTTPMethod":        method.httpMethod,
		"APIKeyRequired":    method.APIKeyRequired,
		"Parameters":        method.Parameters,
		"Models":            method.Models,
		"Responses":         stringResponses,
		"Integration":       method.Integration,
	})
}

type Resource struct {
	pathPart     string
	parentLambda *LambdaAWSInfo
	Methods      map[string]*Method
}

func (resource Resource) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"PathPart": resource.pathPart,
		"LambdaArn": ArbitraryJSONObject{
			"Fn::GetAtt": []string{resource.parentLambda.logicalName(), "Arn"},
		},
		"Methods": resource.Methods,
	})
}

type Stage struct {
	name                string
	CacheClusterEnabled bool
	CacheClusterSize    string
	Description         string
	Variables           map[string]string
}

func (stage Stage) MarshalJSON() ([]byte, error) {
	stageJSON := map[string]interface{}{
		"Name":                stage.name,
		"CacheClusterEnabled": stage.CacheClusterEnabled,
	}
	if len(stage.CacheClusterSize) > 0 {
		stageJSON["CacheClusterSize"] = stage.CacheClusterSize
	}
	if len(stage.Description) > 0 {
		stageJSON["Description"] = stage.Description
	}
	if len(stage.Variables) > 0 {
		stageJSON["Variables"] = stage.Variables
	}
	return json.Marshal(stageJSON)
}

type API struct {
	name        string
	stage       *Stage
	CloneFrom   string
	Description string
	resources   map[string]*Resource
}

type resourceNode struct {
	PathComponent string
	Children      map[string]*resourceNode
	APIResources  map[string]*Resource
}

func (api API) MarshalJSON() ([]byte, error) {

	// Transform the map of resources into a set of hierarchical resourceNodes
	rootResource := resourceNode{
		PathComponent: "/",
		Children:      make(map[string]*resourceNode, 0),
		APIResources:  make(map[string]*Resource, 0),
	}
	for eachPath, eachResource := range api.resources {
		ctxNode := &rootResource
		pathParts := strings.Split(eachPath, "/")[1:]
		// Start at the root and descend
		for _, eachPathPart := range pathParts {
			_, exists := ctxNode.Children[eachPathPart]
			if !exists {
				childNode := &resourceNode{
					PathComponent: eachPathPart,
					Children:      make(map[string]*resourceNode, 0),
					APIResources:  make(map[string]*Resource, 0),
				}
				fmt.Printf("Inserting part: %s\n", eachPathPart)
				ctxNode.Children[eachPathPart] = childNode
			}
			ctxNode = ctxNode.Children[eachPathPart]
			fmt.Printf("Descending node: %s\n", ctxNode.PathComponent)
		}
		ctxNode.APIResources[eachResource.parentLambda.logicalName()] = eachResource
	}

	apiJSON := map[string]interface{}{
		"Name":      api.name,
		"Resources": rootResource,
	}
	if len(api.CloneFrom) > 0 {
		apiJSON["CloneFrom"] = api.CloneFrom
	}
	if len(api.Description) > 0 {
		apiJSON["Description"] = api.Description
	}
	if nil != api.stage {
		apiJSON["Stage"] = *api.stage
	}
	return json.Marshal(apiJSON)
}

func (api *API) export(S3Bucket string,
	S3Key string,
	roleNameMap map[string]interface{},
	resources ArbitraryJSONObject,
	logger *logrus.Logger) error {

	lambdaResourceName, err := ensureConfiguratorLambdaResource(APIGatewayPrincipal, "*", resources, S3Bucket, S3Key, logger)
	if nil != err {
		return err
	}

	// Unmarshal everything to JSON
	apiGatewayInvoker := ArbitraryJSONObject{
		"Type":    "AWS::CloudFormation::CustomResource",
		"Version": "1.0",
		"Properties": ArbitraryJSONObject{
			"ServiceToken": ArbitraryJSONObject{
				"Fn::GetAtt": []string{lambdaResourceName, "Arn"},
			},
			"API": *api,
		},
		"DependsOn": []string{lambdaResourceName},
	}

	apiGatewayInvokerResName := CloudFormationResourceName("APIGateway", api.name)
	resources[apiGatewayInvokerResName] = apiGatewayInvoker
	return nil
}

func (api *API) logicalName() string {
	return CloudFormationResourceName("APIGateway", api.name, api.stage.name)
}

func NewAPIGateway(name string, stage *Stage) *API {
	return &API{
		name:      name,
		stage:     stage,
		resources: make(map[string]*Resource, 0),
	}
}

func NewStage(name string) *Stage {
	return &Stage{
		name:      name,
		Variables: make(map[string]string, 0),
	}
}

func (api *API) NewResource(pathPart string, parentLambda *LambdaAWSInfo) (*Resource, error) {
	_, exists := api.resources[pathPart]
	if exists {
		errMsg := fmt.Sprintf("Path %s already defined for lambda function: %s", pathPart, parentLambda.lambdaFnName)
		return nil, errors.New(errMsg)
	}
	resource := &Resource{
		pathPart:     pathPart,
		parentLambda: parentLambda,
		Methods:      make(map[string]*Method, 0),
	}
	api.resources[pathPart] = resource
	return resource, nil
}

func (resource *Resource) NewMethod(httpMethod string) (*Method, error) {
	authorizationType := "NONE"

	// http://docs.aws.amazon.com/apigateway/latest/developerguide/how-to-method-settings.html#how-to-method-settings-console
	keyname := httpMethod
	_, exists := resource.Methods[keyname]
	if exists {
		errMsg := fmt.Sprintf("Method %s (Auth: %s) already defined for resource", httpMethod, authorizationType)
		return nil, errors.New(errMsg)
	}
	integration := Integration{
		Parameters: make(map[string]string, 0),
		Responses:  make(map[int]IntegrationResponse, 0),
	}

	method := &Method{
		authorizationType: authorizationType,
		httpMethod:        httpMethod,
		Parameters:        make(map[string]bool, 0),
		Models:            make(map[string]Model, 0),
		Responses:         make(map[int]Response, 0),
		Integration:       integration,
	}
	resource.Methods[keyname] = method
	return method, nil
}

func (resource *Resource) NewAuthorizedMethod(httpMethod string, authorizationType string) (*Method, error) {
	method, err := resource.NewMethod(httpMethod)
	if nil != err {
		method.authorizationType = authorizationType
	}
	return method, err
}
