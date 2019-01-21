package sparta

import (
	"github.com/aws/aws-sdk-go/service/s3"
	gocf "github.com/mweagle/go-cloudformation"
)

func stableCloudformationResourceName(prefix string) string {
	return CloudFormationResourceName(prefix, prefix)
}

// S3Site provisions a new, publicly available S3Bucket populated by the
// contents of the resources directory.
// http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/quickref-s3.html#scenario-s3-bucket-website-customdomain
type S3Site struct {
	// Directory or filepath (uncompressed) of contents to use to initialize
	// S3 bucket hosting site.
	resources string
	// If nil, defaults to ErrorDocument: error.html and IndexDocument: index.html
	WebsiteConfiguration *s3.WebsiteConfiguration
	// BucketName is the name of the bucket to create. Required
	// to specify a CloudFront Distribution
	BucketName *gocf.StringExpr
	// UserManifestData is a map of optional data to include
	// in the MANIFEST.json data at the site root. These optional
	// values will be scoped to a `userdata` key in the MANIFEST.json
	// object
	UserManifestData map[string]interface{}
}

// CloudFormationS3ResourceName returns the stable CloudformationResource name that
// can be used by callers to get S3 resource outputs for API Gateway configuration
func (s3Site *S3Site) CloudFormationS3ResourceName() string {
	return stableCloudformationResourceName("S3Site")
}
