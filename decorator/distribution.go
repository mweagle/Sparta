package decorator

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	sparta "github.com/mweagle/Sparta"
	gocf "github.com/mweagle/go-cloudformation"
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
	acmCertificateARN gocf.Stringable) sparta.ServiceDecoratorHookHandler {

	var cert *gocf.CloudFrontDistributionViewerCertificate
	if acmCertificateARN != nil {
		cert = &gocf.CloudFrontDistributionViewerCertificate{
			AcmCertificateArn: acmCertificateARN.String(),
			SslSupportMethod:  gocf.String("vip"),
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
	cert *gocf.CloudFrontDistributionViewerCertificate) sparta.ServiceDecoratorHookHandler {

	// Setup the CF distro
	distroDecorator := func(ctx context.Context,
		serviceName string,
		template *gocf.Template,
		lambdaFunctionCode *gocf.LambdaFunctionCode,
		buildID string,
		awsSession *session.Session,
		noop bool,
		logger *zerolog.Logger) (context.Context, error) {

		// Computed name
		bucketName := domainName
		if subdomain != "" {
			bucketName = fmt.Sprintf("%s.%s", subdomain, domainName)
		}

		// If there isn't a domain name, then it's an issue...
		if s3Site.BucketName == nil {
			return ctx, errors.Errorf("CloudFrontDistribution requires an s3Site.BucketName value in the form of a DNS entry")
		}
		if s3Site.BucketName.Literal != "" && s3Site.BucketName.Literal != bucketName {
			return ctx, errors.Errorf("Mismatch between S3Site.BucketName Literal (%s) and CloudFront DNS entry (%s)",
				s3Site.BucketName.Literal,
				bucketName)
		}

		dnsRecordResourceName := sparta.CloudFormationResourceName("DNSRecord",
			"DNSRecord")
		cloudFrontDistroResourceName := sparta.CloudFormationResourceName("CloudFrontDistro",
			"CloudFrontDistro")

		// Use the HostedZoneName to create the record
		hostedZoneName := fmt.Sprintf("%s.", domainName)
		dnsRecordResource := &gocf.Route53RecordSet{
			// // Zone for the mweagle.io
			HostedZoneName: gocf.String(hostedZoneName),
			Name:           gocf.String(bucketName),
			Type:           gocf.String("A"),
			AliasTarget: &gocf.Route53RecordSetAliasTarget{
				// This HostedZoneID value is required...
				HostedZoneID: gocf.String("Z2FDTNDATAQYW2"),
				DNSName:      gocf.GetAtt(cloudFrontDistroResourceName, "DomainName"),
			},
		}
		template.AddResource(dnsRecordResourceName, dnsRecordResource)
		// IndexDocument
		indexDocument := gocf.String("index.html")
		if s3Site.WebsiteConfiguration != nil &&
			s3Site.WebsiteConfiguration.IndexDocument != nil &&
			s3Site.WebsiteConfiguration.IndexDocument.Suffix != nil {
			indexDocument = gocf.String(*s3Site.WebsiteConfiguration.IndexDocument.Suffix)
		}
		// Add the distro...
		distroConfig := &gocf.CloudFrontDistributionDistributionConfig{
			Aliases:           gocf.StringList(s3Site.BucketName),
			DefaultRootObject: indexDocument,
			Origins: &gocf.CloudFrontDistributionOriginList{
				gocf.CloudFrontDistributionOrigin{
					DomainName:     gocf.GetAtt(s3Site.CloudFormationS3ResourceName(), "DomainName"),
					ID:             gocf.String("S3Origin"),
					S3OriginConfig: &gocf.CloudFrontDistributionS3OriginConfig{},
				},
			},
			Enabled: gocf.Bool(true),
			DefaultCacheBehavior: &gocf.CloudFrontDistributionDefaultCacheBehavior{
				ForwardedValues: &gocf.CloudFrontDistributionForwardedValues{
					QueryString: gocf.Bool(false),
				},
				TargetOriginID:       gocf.String("S3Origin"),
				ViewerProtocolPolicy: gocf.String("allow-all"),
			},
		}
		// Update the cert...
		distroConfig.ViewerCertificate = cert

		cloudfrontDistro := &gocf.CloudFrontDistribution{
			DistributionConfig: distroConfig,
		}
		template.AddResource(cloudFrontDistroResourceName, cloudfrontDistro)

		// Log the created record
		template.Outputs["CloudFrontDistribution"] = &gocf.Output{
			Description: "CloudFront Distribution Route53 entry",
			Value:       s3Site.BucketName,
		}
		return ctx, nil
	}
	return sparta.ServiceDecoratorHookFunc(distroDecorator)
}
