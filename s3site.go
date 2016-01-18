package sparta

import (
	"fmt"
	"os"
	"path/filepath"

	gocf "github.com/mweagle/go-cloudformation"

	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	// OutputS3SiteURL is the keyname used in the CloudFormation Output
	// that stores the S3 backed static site provisioned with this Sparta application
	// @enum OutputKey
	OutputS3SiteURL = "S3SiteURL"
)

// Create the resource, which will be part of the stack definition and use a CustomResource
// to copy the content.  Which means we need PutItem access to the target Bucket.  Use
// Cloudformation to create a random bucketname:
// http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-s3-bucket.html

// Need to create the S3 target bucket
// http://docs.aws.amazon.com/AWSJavaScriptSDK/latest/AWS/S3.html#putBucketWebsite-property

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
}

// export marshals the API data to a CloudFormation compatible representation
func (s3Site *S3Site) export(S3Bucket string,
	S3Key string,
	S3ResourcesKey string,
	apiGatewayOutputs map[string]*gocf.Output,
	roleNameMap map[string]*gocf.StringExpr,
	template *gocf.Template,
	logger *logrus.Logger) error {

	websiteConfig := s3Site.WebsiteConfiguration
	if nil == websiteConfig {
		websiteConfig = &s3.WebsiteConfiguration{}
	}

	//////////////////////////////////////////////////////////////////////////////
	// 1 - Create the S3 bucket.  The "BucketName" property is empty s.t.
	// AWS will assign a unique one.
	if nil == websiteConfig.ErrorDocument {
		websiteConfig.ErrorDocument = &s3.ErrorDocument{
			Key: aws.String("error.html"),
		}
	}
	if nil == websiteConfig.IndexDocument {
		websiteConfig.IndexDocument = &s3.IndexDocument{
			Suffix: aws.String("index.html"),
		}
	}

	s3WebsiteConfig := &gocf.S3WebsiteConfigurationProperty{
		ErrorDocument: gocf.String(aws.StringValue(websiteConfig.ErrorDocument.Key)),
		IndexDocument: gocf.String(aws.StringValue(websiteConfig.IndexDocument.Suffix)),
	}
	s3Bucket := &gocf.S3Bucket{
		AccessControl:        gocf.String("PublicRead"),
		WebsiteConfiguration: s3WebsiteConfig,
	}
	s3BucketResourceName := stableCloudformationResourceName("Site")
	cfResource := template.AddResource(s3BucketResourceName, s3Bucket)
	cfResource.DeletionPolicy = "Delete"

	template.Outputs[OutputS3SiteURL] = &gocf.Output{
		Description: "S3 Website URL",
		Value:       gocf.GetAtt(s3BucketResourceName, "WebsiteURL"),
	}

	// Represents the S3 ARN that is provisioned
	s3SiteBucketResourceValue := gocf.Join("",
		gocf.String("arn:aws:s3:::"),
		gocf.Ref(s3BucketResourceName))
	s3SiteBucketAllKeysResourceValue := gocf.Join("",
		gocf.String("arn:aws:s3:::"),
		gocf.Ref(s3BucketResourceName),
		gocf.String("/*"))

	//////////////////////////////////////////////////////////////////////////////
	// 2 - Add a bucket policy to enable anonymous access, as the PublicRead
	// canned ACL doesn't seem to do what is implied.
	// TODO - determine if this is needed or if PublicRead is being misued
	s3SiteBucketPolicy := &gocf.S3BucketPolicy{
		Bucket: gocf.Ref(s3BucketResourceName).String(),
		PolicyDocument: ArbitraryJSONObject{
			"Version": "2012-10-17",
			"Statement": []ArbitraryJSONObject{
				{
					"Sid":    "PublicReadGetObject",
					"Effect": "Allow",
					"Principal": ArbitraryJSONObject{
						"AWS": "*",
					},
					"Action":   "s3:GetObject",
					"Resource": s3SiteBucketAllKeysResourceValue,
				},
			},
		},
	}
	s3BucketPolicyResourceName := stableCloudformationResourceName("S3SiteBucketPolicy")
	template.AddResource(s3BucketPolicyResourceName, s3SiteBucketPolicy)

	//////////////////////////////////////////////////////////////////////////////
	// 3 - Create the IAM role for the lambda function
	// The lambda function needs to download the posted resource content, as well
	// as manage the S3 bucket that hosts the site.
	statements := CommonIAMStatements["core"]
	statements = append(statements, ArbitraryJSONObject{
		"Action":   []string{"s3:ListBucket"},
		"Effect":   "Allow",
		"Resource": s3SiteBucketResourceValue,
	})
	statements = append(statements, ArbitraryJSONObject{
		"Action":   []string{"s3:DeleteObject", "s3:PutObject"},
		"Effect":   "Allow",
		"Resource": s3SiteBucketAllKeysResourceValue,
	})
	statements = append(statements, ArbitraryJSONObject{
		"Action":   []string{"s3:GetObject"},
		"Effect":   "Allow",
		"Resource": fmt.Sprintf("arn:aws:s3:::%s/%s", S3Bucket, S3ResourcesKey),
	})

	iamS3Role := &gocf.IAMRole{
		AssumeRolePolicyDocument: AssumePolicyDocument,
		Policies: &gocf.IAMPoliciesList{
			gocf.IAMPolicies{
				ArbitraryJSONObject{
					"Version":   "2012-10-17",
					"Statement": statements,
				},
				gocf.String("S3SiteMgmnt"),
			},
		},
	}

	iamRoleName := stableCloudformationResourceName("S3SiteIAMRole")
	cfResource = template.AddResource(iamRoleName, iamS3Role)
	cfResource.DependsOn = append(cfResource.DependsOn, s3BucketResourceName)
	iamRoleRef := gocf.GetAtt(iamRoleName, "Arn")

	//////////////////////////////////////////////////////////////////////////////
	// 4 - Create the lambda function definition that executes with the
	// dynamically provisioned IAM policy
	customResourceHandlerDef := gocf.LambdaFunction{
		Code: &gocf.LambdaFunctionCode{
			S3Bucket: gocf.String(S3Bucket),
			S3Key:    gocf.String(S3Key),
		},
		Description: gocf.String("Manage static S3 site resources"),
		Handler:     gocf.String(nodeJSHandlerName("s3Site")),
		Role:        iamRoleRef,
		Runtime:     gocf.String("nodejs"),
		Timeout:     gocf.Integer(30),
		// Default is 128, but we're buffering everything in memory, in NodeJS
		MemorySize: gocf.Integer(256),
	}
	lambdaResourceName := stableCloudformationResourceName("S3SiteCreator")
	cfResource = template.AddResource(lambdaResourceName, customResourceHandlerDef)
	cfResource.DependsOn = append(cfResource.DependsOn, s3BucketResourceName, iamRoleName)

	//////////////////////////////////////////////////////////////////////////////
	// 5 - Create the custom resource that invokes the site bootstrapper lambda to
	// actually populate the S3 with content
	customResourceName := stableCloudformationResourceName("S3SiteInvoker")
	newResource, err := newCloudFormationResource("Custom::SpartaS3SiteManager", logger)
	if nil != err {
		return err
	}
	customResource := newResource.(*cloudformationS3SiteManager)
	customResource.ServiceToken = gocf.GetAtt(lambdaResourceName, "Arn")
	customResource.TargetBucket = s3SiteBucketResourceValue
	customResource.SourceKey = gocf.String(S3ResourcesKey)
	customResource.SourceBucket = gocf.String(S3Bucket)
	customResource.APIGateway = apiGatewayOutputs

	cfResource = template.AddResource(customResourceName, customResource)
	cfResource.DependsOn = append(cfResource.DependsOn, lambdaResourceName)

	return nil
}

// NewS3Site returns a new S3Site pointer initialized with the
// static resources at the supplied path.  If resources is a directory,
// the contents will be recursively archived and used to populate
// the new S3 bucket.
func NewS3Site(resources string) (*S3Site, error) {
	absPath, err := filepath.Abs(resources)
	if nil != err {
		return nil, err
	}
	_, err = os.Stat(absPath)
	if nil != err {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("Path does not exist: %s", absPath)
		}
	}
	site := &S3Site{
		resources: resources,
	}
	return site, nil
}
