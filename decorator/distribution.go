package decorator

import (
	"context"
	"fmt"

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	gof "github.com/awslabs/goformation/v5/cloudformation"
	gofCloudFront "github.com/awslabs/goformation/v5/cloudformation/cloudfront"
	goflambda "github.com/awslabs/goformation/v5/cloudformation/lambda"
	gofRoute53 "github.com/awslabs/goformation/v5/cloudformation/route53"
	sparta "github.com/mweagle/Sparta"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// CloudFrontSiteDistributionDecorator returns a CloudFrontSiteDecorator with
// the default VIP certificate.
// NOTE: The default VIP certificate is expensive. Consider using SNI to
// reduce costs. See https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-cloudfront-distribution-viewercertificate.html#cfn-cloudfront-distribution-viewercertificate-sslsupportmethod
// for more information.
func CloudFrontSiteDistributionDecorator(s3Site *sparta.S3Site,
	subdomain string,
	domainName string,
	acmCertificateARN string) sparta.ServiceDecoratorHookHandler {

	var cert *gofCloudFront.Distribution_ViewerCertificate
	if acmCertificateARN != "" {
		cert = &gofCloudFront.Distribution_ViewerCertificate{
			AcmCertificateArn: acmCertificateARN,
			SslSupportMethod:  "vip",
		}
	}
	return CloudFrontSiteDistributionDecoratorWithCert(s3Site,
		subdomain,
		domainName,
		cert)
}

// CloudFrontSiteDistributionDecoratorWithCert returns a ServiceDecoratorHookHandler
// function that provisions a CloudFront distribution whose origin
// is the supplied S3Site bucket. The supplied viewer certificate
// allows customization of the CloudFront Distribution SSL options.
// See https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-cloudfront-distribution-viewercertificate.html
// for more information.
func CloudFrontSiteDistributionDecoratorWithCert(s3Site *sparta.S3Site,
	subdomain string,
	domainName string,
	cert *gofCloudFront.Distribution_ViewerCertificate) sparta.ServiceDecoratorHookHandler {

	// Setup the CF distro
	distroDecorator := func(ctx context.Context,
		serviceName string,
		template *gof.Template,
		lambdaFunctionCode *goflambda.Function_Code,
		buildID string,
		awsConfig awsv2.Config,
		noop bool,
		logger *zerolog.Logger) (context.Context, error) {

		// Computed name
		bucketName := domainName
		if subdomain != "" {
			bucketName = fmt.Sprintf("%s.%s", subdomain, domainName)
		}

		// If there isn't a domain name, then it's an issue...
		if s3Site.BucketName == "" {
			return ctx, errors.Errorf("CloudFrontDistribution requires an s3Site.BucketName value in the form of a DNS entry")
		}
		if s3Site.BucketName != "" && s3Site.BucketName != bucketName {
			return ctx, errors.Errorf("Mismatch between S3Site.BucketName Literal (%s) and CloudFront DNS entry (%s)",
				s3Site.BucketName,
				bucketName)
		}

		dnsRecordResourceName := sparta.CloudFormationResourceName("DNSRecord",
			"DNSRecord")
		cloudFrontDistroResourceName := sparta.CloudFormationResourceName("CloudFrontDistro",
			"CloudFrontDistro")

		// Use the HostedZoneName to create the record
		hostedZoneName := fmt.Sprintf("%s.", domainName)
		dnsRecordResource := &gofRoute53.RecordSet{
			// // Zone for the mweagle.io
			HostedZoneName: hostedZoneName,
			Name:           bucketName,
			Type:           "A",
			AliasTarget: &gofRoute53.RecordSet_AliasTarget{
				// This HostedZoneID value is required...
				HostedZoneId: "Z2FDTNDATAQYW2",
				DNSName:      gof.GetAtt(cloudFrontDistroResourceName, "DomainName"),
			},
		}
		template.Resources[dnsRecordResourceName] = dnsRecordResource
		// IndexDocument
		indexDocument := "index.html"
		if s3Site.WebsiteConfiguration != nil &&
			s3Site.WebsiteConfiguration.IndexDocument != nil &&
			s3Site.WebsiteConfiguration.IndexDocument.Suffix != nil {
			indexDocument = *s3Site.WebsiteConfiguration.IndexDocument.Suffix
		}
		// Add the distro...
		distroConfig := &gofCloudFront.Distribution_DistributionConfig{
			Aliases:           []string{s3Site.BucketName},
			DefaultRootObject: indexDocument,
			Origins: []gofCloudFront.Distribution_Origin{
				{
					DomainName:     gof.GetAtt(s3Site.CloudFormationS3ResourceName(), "DomainName"),
					Id:             "S3Origin",
					S3OriginConfig: &gofCloudFront.Distribution_S3OriginConfig{},
				},
			},
			Enabled: true,
			DefaultCacheBehavior: &gofCloudFront.Distribution_DefaultCacheBehavior{
				ForwardedValues: &gofCloudFront.Distribution_ForwardedValues{
					QueryString: false,
				},
				TargetOriginId:       "S3Origin",
				ViewerProtocolPolicy: "allow-all",
			},
		}
		// Update the cert...
		distroConfig.ViewerCertificate = cert

		cloudfrontDistro := &gofCloudFront.Distribution{
			DistributionConfig: distroConfig,
		}
		template.Resources[cloudFrontDistroResourceName] = cloudfrontDistro

		// Log the created record
		template.Outputs["CloudFrontDistribution"] = gof.Output{
			Description: "CloudFront Distribution Route53 entry",
			Value:       s3Site.BucketName,
		}
		return ctx, nil
	}
	return sparta.ServiceDecoratorHookFunc(distroDecorator)
}
