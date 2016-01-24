package sparta

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

// Dynamically assigned discover function that is set by Main
var discoverImpl func() (map[string]interface{}, error)

var discoveryCache map[string]map[string]interface{}

// Discover returns metadata information for resources upon which
// the current golang lambda function depends.
func Discover() (map[string]interface{}, error) {
	if nil == discoverImpl {
		return nil, fmt.Errorf("Discovery service has not been initialized")
	}
	return discoverImpl()
}

func initializeDiscovery(serviceName string, lambdaAWSInfos []*LambdaAWSInfo, logger *logrus.Logger) {
	// Setup the discoveryImpl reference
	discoveryCache = make(map[string]map[string]interface{}, 0)
	discoverImpl = func() (map[string]interface{}, error) {
		pc := make([]uintptr, 2)
		entriesWritten := runtime.Callers(2, pc)
		if entriesWritten != 2 {
			return nil, fmt.Errorf("Unsupported call site for sparta.Discover()")
		}

		// The actual caller is sparta.Discover()
		f := runtime.FuncForPC(pc[1])
		golangFuncName := f.Name()

		// Find the LambdaAWSInfo that has this golang function
		// as its target
		lambdaCFResource := ""
		for _, eachLambda := range lambdaAWSInfos {
			if eachLambda.lambdaFnName == golangFuncName {
				lambdaCFResource = eachLambda.logicalName()
			}
		}
		logger.WithFields(logrus.Fields{
			"CallerName":     golangFuncName,
			"CFResourceName": lambdaCFResource,
			"ServiceName":    serviceName,
		}).Debug("Discovery Info")
		if "" == lambdaCFResource {
			return nil, fmt.Errorf("Unsupported call site for sparta.Discover(): %s", golangFuncName)
		}

		emptyConfiguration := make(map[string]interface{}, 0)
		if "" != lambdaCFResource {
			cachedConfig, exists := discoveryCache[lambdaCFResource]
			if exists {
				return cachedConfig, nil
			}

			// Look it up
			awsCloudFormation := cloudformation.New(awsSession(logger))
			params := &cloudformation.DescribeStackResourceInput{
				LogicalResourceId: aws.String(lambdaCFResource),
				StackName:         aws.String(serviceName),
			}
			result, err := awsCloudFormation.DescribeStackResource(params)
			if nil != err {
				discoveryCache[lambdaCFResource] = emptyConfiguration
				return nil, err
			}
			metadata := result.StackResourceDetail.Metadata
			if nil == metadata {
				metadata = aws.String("{}")
			}
			var discoveryInfo map[string]interface{}
			err = json.Unmarshal([]byte(*metadata), &discoveryInfo)
			discoveryCache[lambdaCFResource] = discoveryInfo
			return discoveryInfo, err
		}
		return emptyConfiguration, nil
	}
}
