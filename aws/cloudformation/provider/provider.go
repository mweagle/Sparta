package provider

import (
	"sync"

	gof "github.com/awslabs/goformation/v5/cloudformation"
	"github.com/rs/zerolog"
)

// CustomResourceProvider allows extend the NewResourceByType factory method
// with their own resource types.
type CustomResourceProvider func(customResourceType string) gof.Resource

var registrationMutex sync.Mutex
var customResourceProviders []CustomResourceProvider

// RegisterCustomResourceProvider registers a custom resource provider with
// go-cloudformation. Multiple
// providers may be registered. The first provider that returns a non-nil
// interface will be used and there is no check for a uniquely registered
// resource type.
func RegisterCustomResourceProvider(provider CustomResourceProvider) {
	registrationMutex.Lock()
	defer registrationMutex.Unlock()
	customResourceProviders = append(customResourceProviders, provider)
}

// NewCloudFormationCustomResource returns a new gof.Resource for the given
// custom resource type. There is no guarnatee that a resource
// exists for resourceType
func NewCloudFormationCustomResource(resourceType string, logger *zerolog.Logger) (gof.Resource, error) {
	for _, eachProvider := range customResourceProviders {
		customType := eachProvider(resourceType)
		if nil != customType {
			return customType, nil
		}
	}
	return nil, nil
}
