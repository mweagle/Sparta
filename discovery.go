package sparta

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
)

// Dynamically assigned discover function that is set by Main
var discoverImpl func() (*DiscoveryInfo, error)

var cachedDiscoveryInfo *DiscoveryInfo

////////////////////////////////////////////////////////////////////////////////
// START - DiscoveryResource
//

// DiscoveryResource stores information about a CloudFormation resource
// that the calling Go function `DependsOn`.
type DiscoveryResource struct {
	ResourceID   string
	ResourceRef  string
	ResourceType string
	Properties   map[string]string
}

//
// END - DiscoveryResource
////////////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////////////
// START - DiscoveryInfo
//

// DiscoveryInfo encapsulates information returned by `sparta.Discovery()`
// to enable a runtime function to discover information about its
// AWS environment or resources that the function created explicit
// `DependsOn` relationships
type DiscoveryInfo struct {
	// Current logical resource ID
	ResourceID string
	// Current AWS region
	Region string
	// Current Stack ID
	StackID string
	// StackName (eg, Sparta service name)
	StackName string
	// Map of resources this Go function has explicit `DependsOn` relationship
	Resources map[string]DiscoveryResource
}

//
// START - DiscoveryInfo
////////////////////////////////////////////////////////////////////////////////

// Discover returns metadata information for resources upon which
// the current golang lambda function depends. It's a reflection-based
// pass-through to DiscoverByName
func Discover() (*DiscoveryInfo, error) {
	if nil == discoverImpl {
		return nil, fmt.Errorf("Discovery service has not been initialized")
	}
	return discoverImpl()
}

func initializeDiscovery(logger *logrus.Logger) {
	// Setup the discoveryImpl reference
	discoverImpl = func() (*DiscoveryInfo, error) {
		// Cached info?
		if cachedDiscoveryInfo != nil {
			return cachedDiscoveryInfo, nil
		}
		// Initialize the cache
		cachedDiscoveryInfo = &DiscoveryInfo{}

		// Get the serialized discovery info the environment string
		discoveryInfo := os.Getenv(spartaEnvVarDiscoveryInformation)
		decoded, decodedErr := base64.StdEncoding.DecodeString(discoveryInfo)
		logger.WithFields(logrus.Fields{
			"DecodeData":  string(decoded),
			"DecodeError": decodedErr,
		}).Debug("Decode result")
		if decodedErr == nil {
			// Unmarshal it...
			unmarshalErr := json.Unmarshal(decoded, cachedDiscoveryInfo)
			if unmarshalErr != nil {
				logger.WithFields(logrus.Fields{
					"Raw":           string(decoded),
					"DiscoveryInfo": cachedDiscoveryInfo,
					"Error":         unmarshalErr,
				}).Error("Failed to unmarshal discovery info")
			}
			decodedErr = unmarshalErr
		}
		return cachedDiscoveryInfo, decodedErr
	}
}
