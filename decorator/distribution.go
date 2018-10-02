package decorator

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	sparta "github.com/mweagle/Sparta"
	gocf "github.com/mweagle/go-cloudformation"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// CloudFrontSiteDistributionDecorator returns a ServiceDecoratorHookHandler
// function that provisions a CloudFront distribution whose origin
// is the supplied S3Site bucket. If the acmCertificateARN
// value is non-nil, the CloudFront distribution will support SSL
// access via the ViewerCertificate struct
func CloudFrontSiteDistributionDecorator(s3Site *sparta.S3Site,
	subdomain string,
	domainName string,
	acmCertificateARN gocf.Stringable) sparta.ServiceDecoratorHookHandler {

	// If there isn't a BucketName, then there's a problem...
	bucketName := domainName
	if subdomain != "" {
		bucketName = fmt.Sprintf("%s.%s", subdomain, domainName)
	}
	// If there is a name set, but it doesn't match what we're going to setup, then it's
	// an eror

	// Setup the CF distro
	distroDecorator := func(context map[string]interface{},
		serviceName string,
		template *gocf.Template,
		S3Bucket string,
		S3Key string,
		buildID string,
		awsSession *session.Session,
		noop bool,
		logger *logrus.Logger) error {

		// If there isn't a domain name, then it's an issue...
		if s3Site.BucketName == nil {
			return errors.Errorf("CloudFrontDistribution requires an s3Site.BucketName value in the form of a DNS entry")
		}
		if s3Site.BucketName.Literal != "" && s3Site.BucketName.Literal != bucketName {
			return errors.Errorf("Mismatch between S3Site.BucketName literal (%s) and CloudFront DNS entry (%s)",
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
		if acmCertificateARN != nil {
			distroConfig.ViewerCertificate = &gocf.CloudFrontDistributionViewerCertificate{
				AcmCertificateArn: acmCertificateARN.String(),
				SslSupportMethod:  gocf.String("vip"),
			}
		}

		cloudfrontDistro := &gocf.CloudFrontDistribution{
			DistributionConfig: distroConfig,
		}
		template.AddResource(cloudFrontDistroResourceName, cloudfrontDistro)

		// Log the created record
		template.Outputs["CloudFrontDistribution"] = &gocf.Output{
			Description: "CloudFront Distribution Route53 entry",
			Value:       s3Site.BucketName,
		}
		return nil
	}
	return sparta.ServiceDecoratorHookFunc(distroDecorator)
}
