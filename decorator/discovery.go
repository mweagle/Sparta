package decorator

import (
	"context"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	spartaAWS "github.com/mweagle/Sparta/aws"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type discoveryInfo struct {
	serviceID     string
	namespaceName string
	serviceName   string
}

var cachedInfo map[string][]*discoveryInfo

func init() {
	cachedInfo = make(map[string][]*discoveryInfo, 0)
}

func discoveryInfoFromIDs(namespaceID string,
	serviceID string,
	logger *logrus.Logger) (*discoveryInfo, error) {

	// Closure to enforce proper semantic return
	returnWrapper := func(discoveryInfo *discoveryInfo) (*discoveryInfo, error) {
		if discoveryInfo.namespaceName != "" &&
			discoveryInfo.serviceName != "" {
			return discoveryInfo, nil
		}
		return nil, errors.Errorf("Failed to lookup (%s, %s) namespaceID, serviceID pair",
			namespaceID,
			serviceID)
	}
	// Start the lookup logic
	existingInfo, existingInfoOk := cachedInfo[namespaceID]
	logger.WithFields(logrus.Fields{
		"existingInfo": existingInfo,
		"exists":       existingInfoOk}).Info("Cached info")

	if existingInfoOk {
		for _, eachDiscoveryInfo := range existingInfo {
			if eachDiscoveryInfo.serviceID == serviceID {
				return returnWrapper(eachDiscoveryInfo)
			}
		}

	}
	// It doesn't exist, let's see if we can get the data...
	locker := sync.RWMutex{}
	locker.Lock()
	defer locker.Unlock()

	lookupInfo := &discoveryInfo{
		serviceID: serviceID,
	}
	// Issue the queries concurrently
	var wg sync.WaitGroup
	wg.Add(2)
	session := spartaAWS.NewSession(logger)
	cloudmapSvc := servicediscovery.New(session)

	// Go get the namespace info
	go func(svc *servicediscovery.ServiceDiscovery) {
		defer wg.Done()

		params := &servicediscovery.GetNamespaceInput{
			Id: aws.String(namespaceID),
		}
		result, resultErr := cloudmapSvc.GetNamespace(params)
		logger.WithFields(logrus.Fields{
			"result":    result,
			"resultErr": resultErr,
		}).Debug("GetNamespace results")
		if resultErr != nil {
			logger.WithField("Error", resultErr).Error("Failed to lookup service")
		} else {
			lookupInfo.namespaceName = *result.Namespace.Name
		}
	}(cloudmapSvc)

	// Go get the service info
	go func(svc *servicediscovery.ServiceDiscovery) {
		defer wg.Done()

		params := &servicediscovery.GetServiceInput{
			Id: aws.String(serviceID),
		}
		result, resultErr := cloudmapSvc.GetService(params)
		logger.WithFields(logrus.Fields{
			"result":    result,
			"resultErr": resultErr,
		}).Debug("GetService results")
		if resultErr != nil {
			logger.WithField("Error", resultErr).Error("Failed to lookup service")
		} else {
			lookupInfo.serviceName = *result.Service.Name
		}
	}(cloudmapSvc)
	wg.Wait()

	// Push it onto the end of the stack and return the value...
	if existingInfo == nil {
		existingInfo = make([]*discoveryInfo, 0)
	}
	existingInfo = append(existingInfo, lookupInfo)
	cachedInfo[namespaceID] = existingInfo
	return returnWrapper(lookupInfo)
}

// DiscoverInstances returns the HttpInstanceSummary items that match
// the given attribute map
func DiscoverInstances(attributes map[string]string,
	logger *logrus.Logger) ([]*servicediscovery.HttpInstanceSummary, error) {
	return DiscoverInstancesWithContext(context.Background(), attributes, logger)
}

// DiscoverInstancesWithContext returns the HttpInstanceSummary items that match
// the given attribute map for the default service provisioned with this
// application
func DiscoverInstancesWithContext(ctx context.Context,
	attributes map[string]string,
	logger *logrus.Logger) ([]*servicediscovery.HttpInstanceSummary, error) {

	// Get the default discovery info and translate that into name/id pairs...
	namespaceID := os.Getenv(EnvVarCloudMapNamespaceID)
	serviceID := os.Getenv(EnvVarCloudMapServiceID)
	discoveryInfo, discoveryInfoErr := discoveryInfoFromIDs(namespaceID, serviceID, logger)

	logger.WithFields(logrus.Fields{
		"namespaceID":      namespaceID,
		"serviceID":        serviceID,
		"discoveryInfo":    discoveryInfo,
		"discoveryInfoErr": discoveryInfoErr,
	}).Debug("Discovery info lookup results")
	if discoveryInfoErr != nil {
		return nil, discoveryInfoErr
	}
	return DiscoverInstancesInServiceWithContext(ctx,
		discoveryInfo.namespaceName,
		discoveryInfo.serviceName,
		attributes,
		logger)
}

// DiscoverInstancesInServiceWithContext returns the HttpInstanceSummary items that match
// the given attribute map using the supplied context and within the given ServiceID
func DiscoverInstancesInServiceWithContext(ctx context.Context,
	namespaceName string,
	serviceName string,
	attributes map[string]string,
	logger *logrus.Logger) ([]*servicediscovery.HttpInstanceSummary, error) {

	// Great, lookup the instances...
	queryParams := make(map[string]*string, 0)
	for eachKey, eachValue := range attributes {
		queryParams[eachKey] = aws.String(eachValue)
	}

	session := spartaAWS.NewSession(logger)
	cloudmapSvc := servicediscovery.New(session)
	lookupParams := &servicediscovery.DiscoverInstancesInput{
		NamespaceName:   aws.String(namespaceName),
		ServiceName:     aws.String(serviceName),
		QueryParameters: queryParams,
	}
	results, resultsErr := cloudmapSvc.DiscoverInstances(lookupParams)
	if resultsErr != nil {
		return nil, resultsErr
	}
	return results.Instances, nil
}
