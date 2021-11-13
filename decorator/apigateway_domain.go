package decorator

import (
	"context"
	"fmt"
	"strings"

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	gofapig "github.com/awslabs/goformation/v5/cloudformation/apigateway"

	gof "github.com/awslabs/goformation/v5/cloudformation"
	goflambda "github.com/awslabs/goformation/v5/cloudformation/lambda"
	gofroute53 "github.com/awslabs/goformation/v5/cloudformation/route53"
	sparta "github.com/mweagle/Sparta/v3"
	spartaCF "github.com/mweagle/Sparta/v3/aws/cloudformation"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

const (
	// APIGatewayMappingEntry is the keyname used to store the API Gateway mappings
	APIGatewayMappingEntry = "APIGatewayMappings"
)

// APIGatewayDomainDecorator returns a ServiceDecoratorHookHandler
// implementation that registers a custom domain for an API Gateway
// service
func APIGatewayDomainDecorator(apiGateway *sparta.API,
	acmCertARN string,
	basePath string,
	domainName string) sparta.ServiceDecoratorHookHandler {

	// Attach the domain decorator to the API GW instance
	domainDecorator := func(ctx context.Context,
		serviceName string,
		template *gof.Template,
		lambdaFunctionCode *goflambda.Function_Code,
		buildID string,
		awsConfig awsv2.Config,
		noop bool,
		logger *zerolog.Logger) (context.Context, error) {

		domainParts := strings.Split(domainName, ".")
		if len(domainParts) != 3 {
			return ctx, errors.Errorf("Invalid domain name supplied to APIGatewayDomainDecorator: %s",
				domainName)
		}
		// Add the mapping
		template.Mappings = map[string]interface{}{
			APIGatewayMappingEntry: spartaCF.APIGatewayMapping,
		}
		// Resource names
		domainInfoResourceName := sparta.CloudFormationResourceName(apiGateway.LogicalResourceName(),
			"Domain")
		basePathMappingResourceName := sparta.CloudFormationResourceName(apiGateway.LogicalResourceName(), "BasePathMapping")
		dnsRecordResourceName := sparta.CloudFormationResourceName(apiGateway.LogicalResourceName(),
			"CloudFrontDNS")

		// Then add all the resources
		domainInfo := &gofapig.DomainName{
			DomainName: domainName,
		}
		apiGatewayType := ""
		apiGWEndpointConfiguration := apiGateway.EndpointConfiguration
		if apiGWEndpointConfiguration != nil && apiGWEndpointConfiguration.Types != nil {
			typesList := apiGWEndpointConfiguration.Types
			if len(typesList) == 1 {
				apiGatewayType = typesList[0]
			} else {
				return ctx, errors.Errorf("Invalid API GW types provided to decorator: %#v",
					apiGWEndpointConfiguration.Types)
			}
		}
		attrName := ""
		switch apiGatewayType {
		case "REGIONAL":
			{
				domainInfo.RegionalCertificateArn = acmCertARN
				domainInfo.EndpointConfiguration = &gofapig.DomainName_EndpointConfiguration{
					Types: []string{
						"REGIONAL",
					}}
				attrName = "RegionalDomainName"
			}
		case "EDGE":
			{
				domainInfo.CertificateArn = acmCertARN
				domainInfo.EndpointConfiguration = &gofapig.DomainName_EndpointConfiguration{
					Types: []string{
						"EDGE",
					}}
				attrName = "DistributionDomainName"
			}
		default:
			return ctx, errors.Errorf("Unsupported API Gateway type: %#v", apiGatewayType)
		}
		template.Resources[domainInfoResourceName] = domainInfo

		basePathMapping := &gofapig.BasePathMapping{
			BasePath:   basePath,
			DomainName: gof.Ref(domainInfoResourceName),
			RestApiId:  gof.Ref(apiGateway.LogicalResourceName()),
		}
		basePathMapping.AWSCloudFormationDependsOn = []string{
			apiGateway.LogicalResourceName(),
		}
		template.Resources[basePathMappingResourceName] = basePathMapping

		// Use the HostedZoneName to create the record
		domainZone := domainParts[1:]
		dnsRecordResource := &gofroute53.RecordSet{
			HostedZoneName: fmt.Sprintf("%s.", strings.Join(domainZone, ".")),
			Name:           fmt.Sprintf("%s.", domainName),
			Type:           "A",
			AliasTarget: &gofroute53.RecordSet_AliasTarget{
				HostedZoneId: gof.FindInMap(APIGatewayMappingEntry,
					gof.Ref("AWS::Region"),
					spartaCF.HostedZoneID),
				DNSName: gof.GetAtt(domainInfoResourceName, attrName),
			},
		}
		template.Resources[dnsRecordResourceName] = dnsRecordResource

		// Add an output...
		template.Outputs["APIGatewayCustomDomain"] = gof.Output{
			Description: "Custom API Gateway Domain",
			Value:       domainName,
		}
		return ctx, nil
	}
	return sparta.ServiceDecoratorHookFunc(domainDecorator)
}
