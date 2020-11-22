package decorator

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	sparta "github.com/mweagle/Sparta"
	spartaCF "github.com/mweagle/Sparta/aws/cloudformation"
	gocf "github.com/mweagle/go-cloudformation"
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
	acmCertARN gocf.Stringable,
	basePath string,
	domainName string) sparta.ServiceDecoratorHookHandler {

	// Attach the domain decorator to the API GW instance
	domainDecorator := func(ctx context.Context,
		serviceName string,
		template *gocf.Template,
		lambdaFunctionCode *gocf.LambdaFunctionCode,
		buildID string,
		awsSession *session.Session,
		noop bool,
		logger *zerolog.Logger) (context.Context, error) {

		domainParts := strings.Split(domainName, ".")
		if len(domainParts) != 3 {
			return ctx, errors.Errorf("Invalid domain name supplied to APIGatewayDomainDecorator: %s",
				domainName)
		}
		// Add the mapping
		template.Mappings = map[string]*gocf.Mapping{
			APIGatewayMappingEntry: spartaCF.APIGatewayMapping,
		}
		// Resource names
		domainInfoResourceName := sparta.CloudFormationResourceName(apiGateway.LogicalResourceName(),
			"Domain")
		basePathMappingResourceName := sparta.CloudFormationResourceName(apiGateway.LogicalResourceName(), "BasePathMapping")
		dnsRecordResourceName := sparta.CloudFormationResourceName(apiGateway.LogicalResourceName(),
			"CloudFrontDNS")

		// Then add all the resources
		domainInfo := &gocf.APIGatewayDomainName{
			DomainName: gocf.String(domainName),
		}
		apiGatewayType := ""
		apiGWEndpointConfiguration := apiGateway.EndpointConfiguration
		if apiGWEndpointConfiguration != nil && apiGWEndpointConfiguration.Types != nil {
			typesList := apiGWEndpointConfiguration.Types
			if len(typesList.Literal) == 1 {
				apiGatewayType = typesList.Literal[0].Literal
			} else {
				return ctx, errors.Errorf("Invalid API GW types provided to decorator: %#v",
					apiGWEndpointConfiguration.Types)
			}
		}
		attrName := ""
		switch apiGatewayType {
		case "REGIONAL":
			{
				domainInfo.RegionalCertificateArn = acmCertARN.String()
				domainInfo.EndpointConfiguration = &gocf.APIGatewayDomainNameEndpointConfiguration{
					Types: gocf.StringList(gocf.String("REGIONAL")),
				}
				attrName = "RegionalDomainName"
			}
		case "EDGE":
			{
				domainInfo.CertificateArn = acmCertARN.String()
				domainInfo.EndpointConfiguration = &gocf.APIGatewayDomainNameEndpointConfiguration{
					Types: gocf.StringList(gocf.String("EDGE")),
				}
				attrName = "DistributionDomainName"
			}
		default:
			return ctx, errors.Errorf("Unsupported API Gateway type: %#v", apiGatewayType)
		}
		template.AddResource(domainInfoResourceName, domainInfo)

		basePathMapping := gocf.APIGatewayBasePathMapping{
			BasePath:   gocf.String(basePath),
			DomainName: gocf.Ref(domainInfoResourceName).String(),
			RestAPIID:  gocf.Ref(apiGateway.LogicalResourceName()).String(),
		}
		mappingResource := template.AddResource(basePathMappingResourceName, basePathMapping)
		mappingResource.DependsOn = []string{domainInfoResourceName,
			apiGateway.LogicalResourceName()}

		// Use the HostedZoneName to create the record
		domainZone := domainParts[1:]
		dnsRecordResource := &gocf.Route53RecordSet{
			HostedZoneName: gocf.String(fmt.Sprintf("%s.", strings.Join(domainZone, "."))),
			Name:           gocf.String(fmt.Sprintf("%s.", domainName)),
			Type:           gocf.String("A"),
			AliasTarget: &gocf.Route53RecordSetAliasTarget{
				HostedZoneID: gocf.FindInMap(APIGatewayMappingEntry,
					gocf.Ref("AWS::Region"),
					gocf.String(spartaCF.HostedZoneID)),
				DNSName: gocf.GetAtt(domainInfoResourceName, attrName).String(),
			},
		}
		template.AddResource(dnsRecordResourceName, dnsRecordResource)

		// Add an output...
		template.Outputs["APIGatewayCustomDomain"] = &gocf.Output{
			Description: "Custom API Gateway Domain",
			Value:       gocf.String(domainName),
		}
		return ctx, nil
	}
	return sparta.ServiceDecoratorHookFunc(domainDecorator)
}
