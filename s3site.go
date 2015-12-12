package sparta

import (
	"fmt"

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
	roleNameMap map[string]interface{},
	resources ArbitraryJSONObject,
	outputs ArbitraryJSONObject,
	logger *logrus.Logger) error {

	websiteConfig := s3Site.WebsiteConfiguration
	if nil == websiteConfig {
		websiteConfig = &s3.WebsiteConfiguration{
			ErrorDocument: &s3.ErrorDocument{
				Key: aws.String("error.html"),
			},
			IndexDocument: &s3.IndexDocument{
				Suffix: aws.String("index.html"),
			},
		}
	}

	//////////////////////////////////////////////////////////////////////////////
	// 1 - Create the S3 bucket.  The "BucketName" property is empty s.t.
	// AWS will assign a unique one.
	s3BucketSite := ArbitraryJSONObject{
		"Type": "AWS::S3::Bucket",
		// http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-attribute-deletionpolicy.html
		"DeletionPolicy": "Delete",
		"Properties": ArbitraryJSONObject{
			"AccessControl": "PublicRead",
			"WebsiteConfiguration": ArbitraryJSONObject{
				"ErrorDocument": *(websiteConfig.ErrorDocument.Key),
				"IndexDocument": *(websiteConfig.IndexDocument.Suffix),
			},
		},
	}
	s3BucketResourceName := stableCloudformationResourceName("Site")
	resources[s3BucketResourceName] = s3BucketSite

	// Include the WebsiteURL in the outputs
	outputs[OutputS3SiteURL] = ArbitraryJSONObject{
		"Description": "S3 website URL",
		"Value": ArbitraryJSONObject{
			"Fn::GetAtt": []string{s3BucketResourceName, "WebsiteURL"},
		},
	}

	// Represents the S3 ARN that is provisioned
	s3SiteBucketResourceValue := ArbitraryJSONObject{
		"Fn::Join": []interface{}{"",
			[]interface{}{
				"arn:aws:s3:::",
				ArbitraryJSONObject{
					"Ref": s3BucketResourceName,
				},
			}}}

	s3SiteBucketAllKeysResourceValue := ArbitraryJSONObject{
		"Fn::Join": []interface{}{"",
			[]interface{}{
				"arn:aws:s3:::",
				ArbitraryJSONObject{
					"Ref": s3BucketResourceName,
				},
				"/*",
			}}}

	//////////////////////////////////////////////////////////////////////////////
	// 2 - Add a bucket policy to enable anonymous access, as the PublicRead
	// canned ACL doesn't seem to do what is implied.
	// TODO - determine if this is needed or if PublicRead is being misued
	s3SiteBucketPolicy := ArbitraryJSONObject{
		"Type": "AWS::S3::BucketPolicy",
		"Properties": ArbitraryJSONObject{
			"Bucket": ArbitraryJSONObject{"Ref": s3BucketResourceName},
			"PolicyDocument": ArbitraryJSONObject{
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
		},
	}
	s3BucketPolicyResourceName := stableCloudformationResourceName("S3SiteBucketPolicy")
	resources[s3BucketPolicyResourceName] = s3SiteBucketPolicy

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
	iamPolicy := ArbitraryJSONObject{"Type": "AWS::IAM::Role",
		"DependsOn": []string{s3BucketResourceName},
		"Properties": ArbitraryJSONObject{
			"AssumeRolePolicyDocument": AssumePolicyDocument,
			"Policies": []ArbitraryJSONObject{
				{
					"PolicyName": "S3SiteMgmnt",
					"PolicyDocument": ArbitraryJSONObject{
						"Version":   "2012-10-17",
						"Statement": statements,
					},
				},
			},
		},
	}
	iamRoleName := stableCloudformationResourceName("S3SiteIAMRole")
	resources[iamRoleName] = iamPolicy
	iamRoleRef := ArbitraryJSONObject{
		"Fn::GetAtt": []string{iamRoleName, "Arn"},
	}

	//////////////////////////////////////////////////////////////////////////////
	// 4 - Create the lambda function definition that executes with the
	// dynamically provisioned IAM policy
	customResourceHandlerDef := ArbitraryJSONObject{
		"Type":      "AWS::Lambda::Function",
		"DependsOn": []string{s3BucketResourceName, iamRoleName},
		"Properties": ArbitraryJSONObject{
			"Code": ArbitraryJSONObject{
				"S3Bucket": S3Bucket,
				"S3Key":    S3Key,
			},
			"Role":    iamRoleRef,
			"Handler": nodeJSHandlerName("s3Site"),
			"Runtime": "nodejs",
			"Timeout": "120",
			// Default is 128, but we're buffering everything in memory...
			"MemorySize": 256,
		},
	}
	lambdaResourceName := stableCloudformationResourceName("S3SiteCreator")
	resources[lambdaResourceName] = customResourceHandlerDef

	//////////////////////////////////////////////////////////////////////////////
	// 5 - Create the custom resource that invokes the site bootstrapper lambda to
	// actually populate the S3 with content
	customResourceName := stableCloudformationResourceName("S3SiteInvoker")
	s3SiteCustomResource := ArbitraryJSONObject{
		"Type":    "AWS::CloudFormation::CustomResource",
		"Version": "1.0",
		"Properties": ArbitraryJSONObject{
			"ServiceToken": ArbitraryJSONObject{
				"Fn::GetAtt": []string{lambdaResourceName, "Arn"},
			},
			"TargetBucket": s3SiteBucketResourceValue,
			"SourceKey":    S3ResourcesKey,
			"SourceBucket": S3Bucket,
		},
		"DependsOn": []string{lambdaResourceName},
	}
	resources[customResourceName] = s3SiteCustomResource

	return nil
}

// NewS3Site returns a new S3Site pointer initialized with the
// static resources at the supplied path.  If resources is a directory,
// the contents will be recursively archived and used to populate
// the new S3 bucket.
func NewS3Site(resources string) (*S3Site, error) {
	site := &S3Site{
		resources: resources,
	}
	return site, nil
}
