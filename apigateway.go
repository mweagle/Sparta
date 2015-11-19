package sparta

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Sirupsen/logrus"
)

type Method struct {
	authorizationType string `json:"AuthorizationType,omitempty"`
	httpMethod        string `json:"HTTPMethod,omitempty"`
	APIKeyRequired    bool
	RequestModels     map[string]string
	RequestParameters map[string]string
}

func (method Method) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"AuthorizationType": method.authorizationType,
		"HTTPMethod":        method.httpMethod,
		"APIKeyRequired":    method.APIKeyRequired,
		"RequestModels":     method.RequestModels,
		"RequestParameters": method.RequestParameters,
	})
}

type Resource struct {
	pathPart     string `json:"PathPart,omitempty"`
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
	name                string `json:"Name,omitempty"`
	CacheClusterEnabled bool
	CacheClusterSize    string
	Description         string
	Variables           map[string]string
}

func (stage Stage) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"Name":                stage.name,
		"CacheClusterEnabled": stage.CacheClusterEnabled,
		"CacheClusterSize":    stage.CacheClusterSize,
		"Description":         stage.Description,
		"Variables":           stage.Variables,
	})
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
	marshalMap := make(map[string]interface{})
	marshalMap["Name"] = api.name
	marshalMap["CloneFrom"] = api.CloneFrom
	marshalMap["Description"] = api.Description
	marshalMap["Resources"] = rootResource
	if nil != api.stage {
		marshalMap["Stage"] = *api.stage
	}
	return json.Marshal(marshalMap)
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

func (resource *Resource) NewMethod(httpMethod string, authorizationType string) (*Method, error) {
	// http://docs.aws.amazon.com/apigateway/latest/developerguide/how-to-method-settings.html#how-to-method-settings-console
	if "" == authorizationType {
		authorizationType = "NONE"
	}
	keyname := fmt.Sprintf("%s%s", httpMethod, authorizationType)
	_, exists := resource.Methods[keyname]
	if exists {
		errMsg := fmt.Sprintf("Method %s (Auth: %s) already defined for resource", httpMethod, authorizationType)
		return nil, errors.New(errMsg)
	}
	method := &Method{
		authorizationType: authorizationType,
		httpMethod:        httpMethod,
		RequestModels:     make(map[string]string, 0),
		RequestParameters: make(map[string]string, 0),
	}
	resource.Methods[keyname] = method
	return method, nil
}
