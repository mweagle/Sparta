//go:build !lambdabinary
// +build !lambdabinary

package sparta

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	gof "github.com/awslabs/goformation/v5/cloudformation"
	gofiam "github.com/awslabs/goformation/v5/cloudformation/iam"
	goflambda "github.com/awslabs/goformation/v5/cloudformation/lambda"
	gofs3 "github.com/awslabs/goformation/v5/cloudformation/s3"
	cfCustomResources "github.com/mweagle/Sparta/aws/cloudformation/resources"
	spartaIAM "github.com/mweagle/Sparta/aws/iam"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
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

// export marshals the API data to a CloudFormation compatible representation
func (s3Site *S3Site) export(serviceName string,
	binaryName string,
	s3ArtifactBucket string,
	s3CodeResource *goflambda.Function_Code,
	s3ResourcesKey string,
	apiGatewayOutputs map[string]gof.Output,
	roleNameMap map[string]string,
	template *gof.Template,
	logger *zerolog.Logger) error {

	if s3Site.WebsiteConfiguration == nil {
		s3Site.WebsiteConfiguration = &s3.WebsiteConfiguration{
			ErrorDocument: &s3.ErrorDocument{
				Key: aws.String("error.html"),
			},
			IndexDocument: &s3.IndexDocument{
				Suffix: aws.String("index.html"),
			},
		}
	}
	// Ensure everything is set
	if s3Site.WebsiteConfiguration.ErrorDocument == nil {
		s3Site.WebsiteConfiguration.ErrorDocument = &s3.ErrorDocument{
			Key: aws.String("error.html"),
		}
	}
	if s3Site.WebsiteConfiguration.IndexDocument == nil {
		s3Site.WebsiteConfiguration.IndexDocument = &s3.IndexDocument{
			Suffix: aws.String("index.html"),
		}
	}

	//////////////////////////////////////////////////////////////////////////////
	// 1 - Create the S3 bucket.  The "BucketName" property is empty s.t.
	// AWS will assign a unique one.

	s3WebsiteConfig := &gofs3.Bucket_WebsiteConfiguration{
		ErrorDocument: *s3Site.WebsiteConfiguration.ErrorDocument.Key,
		IndexDocument: *s3Site.WebsiteConfiguration.IndexDocument.Suffix,
	}
	s3Bucket := &gofs3.Bucket{
		AccessControl:        "PublicRead",
		WebsiteConfiguration: s3WebsiteConfig,
	}
	s3Bucket.BucketName = s3Site.BucketName
	s3Bucket.AWSCloudFormationDeletionPolicy = "Delete"
	s3BucketResourceName := s3Site.CloudFormationS3ResourceName()
	template.Resources[s3BucketResourceName] = s3Bucket

	template.Outputs[OutputS3SiteURL] = gof.Output{
		Description: "S3 Website URL",
		Value:       gof.GetAtt(s3BucketResourceName, "WebsiteURL"),
	}

	// Represents the S3 ARN that is provisioned
	s3SiteBucketResourceValue := gof.Join("", []string{
		"arn:aws:s3:::",
		s3BucketResourceName,
	})
	s3SiteBucketAllKeysResourceValue := gof.Join("", []string{
		"arn:aws:s3:::",
		gof.Ref(s3BucketResourceName),
		"/*"})

	//////////////////////////////////////////////////////////////////////////////
	// 2 - Add a bucket policy to enable anonymous access, as the PublicRead
	// canned ACL doesn't seem to do what is implied.
	// TODO - determine if this is needed or if PublicRead is being misued
	s3SiteBucketPolicy := &gofs3.BucketPolicy{
		Bucket: gof.Ref(s3BucketResourceName),
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
	template.Resources[s3BucketPolicyResourceName] = s3SiteBucketPolicy

	//////////////////////////////////////////////////////////////////////////////
	// 3 - Create the IAM role for the lambda function
	// The lambda function needs to download the posted resource content, as well
	// as manage the S3 bucket that hosts the site.
	statements := CommonIAMStatements.Core
	statements = append(statements, spartaIAM.PolicyStatement{
		Action: []string{"s3:ListBucket",
			"s3:ListObjectsPages"},
		Effect:   "Allow",
		Resource: s3SiteBucketResourceValue,
	})
	statements = append(statements, spartaIAM.PolicyStatement{
		Action: []string{"s3:DeleteObject",
			"s3:PutObject",
			"s3:DeleteObjects"},
		Effect:   "Allow",
		Resource: s3SiteBucketAllKeysResourceValue,
	})
	statements = append(statements, spartaIAM.PolicyStatement{
		Action: []string{"s3:GetObject"},
		Effect: "Allow",
		Resource: gof.Join("", []string{
			"arn:aws:s3:::",
			s3ArtifactBucket,
			"/",
			s3ResourcesKey}),
	})

	iamPolicyList := []gofiam.Role_Policy{}
	iamPolicyList = append(iamPolicyList,
		gofiam.Role_Policy{
			PolicyDocument: ArbitraryJSONObject{
				"Version":   "2012-10-17",
				"Statement": statements,
			},
			PolicyName: "S3SiteMgmnt",
		},
	)

	iamS3Role := &gofiam.Role{
		AssumeRolePolicyDocument: AssumePolicyDocument,
		Policies:                 iamPolicyList,
	}

	iamRoleName := stableCloudformationResourceName("S3SiteIAMRole")
	iamS3Role.AWSCloudFormationDependsOn = []string{s3BucketResourceName}
	template.Resources[iamRoleName] = iamS3Role
	iamRoleRef := gof.GetAtt(iamRoleName, "Arn")

	// Create the IAM role and CustomAction handler to do the work

	//////////////////////////////////////////////////////////////////////////////
	// 4 - Create the lambda function definition that executes with the
	// dynamically provisioned IAM policy.  This is similar to what happens in
	// EnsureCustomResourceHandler, but due to the more complex IAM rules
	// there's a bit of duplication
	//	handlerName := lambdaExportNameForCustomResourceType(cloudformationresources.ZipToS3Bucket)
	logger.Debug().
		Interface("CustomResourceType", cfCustomResources.ZipToS3Bucket).
		Msg("Sparta CloudFormation custom resource handler info")

	// Since this is a custom resource command, stuff the type in the environment
	userDispatchMap := map[string]string{
		EnvVarCustomResourceTypeName: cfCustomResources.ZipToS3Bucket,
	}
	lambdaEnv, lambdaEnvErr := lambdaFunctionEnvironment(userDispatchMap,
		cfCustomResources.ZipToS3Bucket,
		nil,
		logger)
	if lambdaEnvErr != nil {
		return errors.Wrapf(lambdaEnvErr, "Failed to create S3 site resource")
	}
	customResourceHandlerDef := &goflambda.Function{
		Code:        s3CodeResource,
		Description: customResourceDescription(serviceName, "S3 static site"),
		Role:        iamRoleRef,
		MemorySize:  256,
		Timeout:     180,
		// Let AWS assign the function name
		/*
			FunctionName: lambdaFunctionName.String(),
		*/
		Environment: lambdaEnv,
	}
	if s3CodeResource.ImageUri != "" {
		customResourceHandlerDef.PackageType = "Image"
	} else {
		customResourceHandlerDef.Runtime = string(Go1LambdaRuntime)
		customResourceHandlerDef.Handler = binaryName
	}
	lambdaResourceName := stableCloudformationResourceName("S3SiteCreator")
	customResourceHandlerDef.AWSCloudFormationDependsOn = []string{s3BucketResourceName,
		iamRoleName}
	template.Resources[lambdaResourceName] = customResourceHandlerDef

	//////////////////////////////////////////////////////////////////////////////
	// 5 - Create the custom resource that invokes the site bootstrapper lambda to
	// actually populate the S3 with content
	customResourceName := CloudFormationResourceName("S3SiteBuilder")
	newResource, err := newCloudFormationResource(cfCustomResources.ZipToS3Bucket, logger)
	if nil != err {
		return errors.Wrapf(err, "Failed to create ZipToS3Bucket CustomResource")
	}
	zipResource, zipResourceOK := newResource.(*cfCustomResources.ZipToS3BucketResource)
	if !zipResourceOK {
		return errors.Errorf("Failed to type assert *cfCustomResources.ZipToS3BucketResource custom resource")
	}
	zipResource.ServiceToken = gof.GetAtt(lambdaResourceName, "Arn")
	zipResource.SrcKeyName = s3ResourcesKey
	zipResource.SrcBucket = s3ArtifactBucket
	zipResource.DestBucket = gof.Ref(s3BucketResourceName)

	// Build the manifest data with any output info...
	manifestData := make(map[string]interface{})
	for eachKey, eachOutput := range apiGatewayOutputs {
		manifestData[eachKey] = map[string]interface{}{
			"Description": eachOutput.Description,
			"Value":       eachOutput.Value,
		}
	}
	if len(s3Site.UserManifestData) != 0 {
		manifestData["userdata"] = s3Site.UserManifestData
	}

	zipResource.Manifest = manifestData
	zipResource.AWSCloudFormationDependsOn = []string{
		lambdaResourceName,
		s3BucketResourceName,
	}
	template.Resources[customResourceName] = zipResource
	return nil
}

// NewS3Site returns a new S3Site pointer initialized with the
// static resources at the supplied path.  If resources is a directory,
// the contents will be recursively archived and used to populate
// the new S3 bucket.
func NewS3Site(resources string) (*S3Site, error) {
	// We'll ensure its valid during the build step, since
	// there could be a go:generate command in the source that
	// actually builds it.
	site := &S3Site{
		resources:        resources,
		UserManifestData: map[string]interface{}{},
	}
	return site, nil
}
