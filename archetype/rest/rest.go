package rest

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/mweagle/Sparta"
	"github.com/pkg/errors"
)

var allHTTPMethods = strings.Join([]string{
	http.MethodGet,
	http.MethodHead,
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
	http.MethodConnect,
	http.MethodOptions,
	http.MethodTrace,
}, " ")

// MethodHandlerMap is a map of http method names to their handlers
type MethodHandlerMap map[string]*MethodHandler

// MethodHandler represents a handler for a given HTTP method
type MethodHandler struct {
	DefaultCode int
	statusCodes []int
	Handler     interface{}
	privileges  []sparta.IAMRolePrivilege
	options     *sparta.LambdaFunctionOptions
	headers     []string
}

// StatusCodes is a fluent builder to append additional HTTP status codes
// for the given MethodHandler. It's primarily used to disamgiguate
// input from the NewMethodHandler constructor
func (mh *MethodHandler) StatusCodes(codes ...int) *MethodHandler {
	if mh.statusCodes == nil {
		mh.statusCodes = make([]int, 0)
	}
	for _, eachCode := range codes {
		mh.statusCodes = append(mh.statusCodes, eachCode)
	}
	return mh
}

// Options is a fluent builder that allows customizing the lambda execution
// options for the given function
func (mh *MethodHandler) Options(options *sparta.LambdaFunctionOptions) *MethodHandler {
	mh.options = options
	return mh
}

// Privileges is the fluent builder to associated IAM privileges with this
// HTTP handler
func (mh *MethodHandler) Privileges(privileges ...sparta.IAMRolePrivilege) *MethodHandler {
	if mh.privileges == nil {
		mh.privileges = make([]sparta.IAMRolePrivilege, 0)
	}
	for _, eachPrivilege := range privileges {
		mh.privileges = append(mh.privileges, eachPrivilege)
	}
	return mh
}

// Headers is the fluent builder that defines what headers this method returns
func (mh *MethodHandler) Headers(headerNames ...string) *MethodHandler {
	if mh.headers == nil {
		mh.headers = make([]string, 0)
	}
	for _, eachHeader := range headerNames {
		mh.headers = append(mh.headers, eachHeader)
	}
	return mh
}

// NewMethodHandler is a constructor function to return a new MethodHandler
// pointer instance.
func NewMethodHandler(handler interface{}, defaultCode int) *MethodHandler {
	return &MethodHandler{
		DefaultCode: defaultCode,
		Handler:     handler,
	}
}

// ResourceDefinition represents a set of handlers for a given URL path
type ResourceDefinition struct {
	URL            string
	MethodHandlers MethodHandlerMap
}

// Resource defines the interface an object must define in order to
// provide a ResourceDefinition
type Resource interface {
	ResourceDefinition() (ResourceDefinition, error)
}

// RegisterResource creates a set of lambda handlers for the given resource
// and registers them with the apiGateway. The sparta Lambda handler returned
// slice is eligible
func RegisterResource(apiGateway *sparta.API, resource Resource) ([]*sparta.LambdaAWSInfo, error) {

	definition, definitionErr := resource.ResourceDefinition()
	if definitionErr != nil {
		return nil, errors.Wrapf(definitionErr, "requesting ResourceDefinition from provider")
	}

	urlParts, urlPartsErr := url.Parse(definition.URL)
	if urlPartsErr != nil {
		return nil, errors.Wrapf(urlPartsErr, "parsing REST URL: %s", definition.URL)
	}
	// Any query params?
	queryParams, queryParamsErr := url.ParseQuery(urlParts.RawQuery)
	if nil != queryParamsErr {
		return nil, errors.Wrap(queryParamsErr, "parsing REST URL query params")
	}

	// Any path params?
	pathParams := []string{}
	pathParts := strings.Split(urlParts.Path, "/")
	for _, eachPathPart := range pathParts {
		trimmedPathPart := strings.Trim(eachPathPart, "{}")
		if trimmedPathPart != eachPathPart {
			pathParams = append(pathParams, trimmedPathPart)
		}
	}

	// Local function to produce a friendlyname for the provider
	lambdaName := func(methodName string) string {
		nameValue := fmt.Sprintf("%T_%s", resource, methodName)
		return strings.Trim(nameValue, "_-.()*")
	}

	// Local function to handle registering the function with API Gateway
	createAPIGEntry := func(methodName string,
		methodHandler *MethodHandler,
		handler *sparta.LambdaAWSInfo) error {
		apiGWResource, apiGWResourceErr := apiGateway.NewResource(definition.URL, handler)
		if apiGWResourceErr != nil {
			return errors.Wrapf(apiGWResourceErr, "attempting to create API Gateway Resource")
		}
		statusCodes := methodHandler.statusCodes
		if statusCodes == nil {
			statusCodes = []int{}
		}
		// We only return http.StatusOK
		apiMethod, apiMethodErr := apiGWResource.NewMethod(methodName,
			methodHandler.DefaultCode,
			statusCodes...)
		if apiMethodErr != nil {
			return apiMethodErr
		}
		// Do anything smart with the URL? Split the URL into components to first see
		// if it's a URL template
		for _, eachPathPart := range pathParams {
			apiMethod.Parameters[fmt.Sprintf("method.request.path.%s", eachPathPart)] = true
		}

		// Then parse it to see what's up with the query param names
		for eachQueryParam := range queryParams {
			apiMethod.Parameters[fmt.Sprintf("method.request.querystring.%s", eachQueryParam)] = true
		}
		// Any headers?
		for _, eachHeader := range methodHandler.headers {
			// Make this an optional header on the method response
			methodHeaderKey := fmt.Sprintf("method.response.header.%s", eachHeader)

			for _, eachResponse := range apiMethod.Responses {
				eachResponse.Parameters[methodHeaderKey] = false
			}
			// Add it to the integration mappings
			// Then ensure every integration response knows how to pass it along...
			inputSelector := fmt.Sprintf("integration.response.header.%s", eachHeader)
			for _, eachIntegrationResponse := range apiMethod.Integration.Responses {
				if len(eachIntegrationResponse.Parameters) <= 0 {
					eachIntegrationResponse.Parameters = make(map[string]interface{})
				}
				eachIntegrationResponse.Parameters[methodHeaderKey] = inputSelector
			}
		}

		return nil
	}
	resourceMap := make(map[string]*sparta.LambdaAWSInfo, 0)

	// Great, walk the map of handlers
	for eachMethod, eachMethodDefinition := range definition.MethodHandlers {
		if !strings.Contains(allHTTPMethods, eachMethod) {
			return nil, errors.Errorf("unsupported HTTP method name: `%s %s`. Supported: %s",
				eachMethod,
				definition.URL,
				allHTTPMethods)
		}
		lambdaFn, lambdaFnErr := sparta.NewAWSLambda(lambdaName(eachMethod),
			eachMethodDefinition.Handler,
			sparta.IAMRoleDefinition{})

		if lambdaFnErr != nil {
			return nil, errors.Wrapf(lambdaFnErr,
				"attempting to register url `%s %s`", eachMethod, definition.URL)
		}

		resourceMap[eachMethod] = lambdaFn

		// Any options?
		if eachMethodDefinition.options != nil {
			lambdaFn.Options = eachMethodDefinition.options
		}

		// Any privs?
		if len(eachMethodDefinition.privileges) != 0 {
			lambdaFn.RoleDefinition.Privileges = eachMethodDefinition.privileges
		}

		// Register the route...
		apiGWRegistrationErr := createAPIGEntry(eachMethod, eachMethodDefinition, lambdaFn)
		if apiGWRegistrationErr != nil {
			return nil, errors.Wrapf(apiGWRegistrationErr, "attemping to create resource for method: %s", http.MethodHead)
		}
	}
	if len(resourceMap) <= 0 {
		return nil, errors.Errorf("No resource methodHandlers found for resource: %T", resource)
	}
	// Convert this into a slice and return it...
	lambdaResourceHandlers := make([]*sparta.LambdaAWSInfo,
		len(resourceMap), len(resourceMap))
	lambdaIndex := 0
	for _, eachLambda := range resourceMap {
		lambdaResourceHandlers[lambdaIndex] = eachLambda
		lambdaIndex++
	}
	return lambdaResourceHandlers, nil
}
