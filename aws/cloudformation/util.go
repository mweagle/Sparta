package cloudformation

import (
	"fmt"

	gocf "github.com/crewjam/go-cloudformation"
)

// S3AllKeysArnForBucket returns a CloudFormation-compatible Arn expression
// (string or Ref) for all bucket keys (`/*`).  The bucket
// parameter may be either a string or an interface{} ("Ref: "myResource")
// value
func S3AllKeysArnForBucket(bucket interface{}) *gocf.StringExpr {
	arnParts := []gocf.Stringable{gocf.String("arn:aws:s3:::")}

	switch bucket.(type) {
	case string:
		// Don't be smart if the Arn value is a user supplied literal
		arnParts = append(arnParts, gocf.String(bucket.(string)))
	case *gocf.StringExpr:
		arnParts = append(arnParts, bucket.(*gocf.StringExpr))
	case gocf.RefFunc:
		arnParts = append(arnParts, bucket.(gocf.RefFunc).String())
	default:
		panic(fmt.Sprintf("Unsupported SourceArn value type: %+v", bucket))
	}
	arnParts = append(arnParts, gocf.String("/*"))
	return gocf.Join("", arnParts...).String()
}

// S3ArnForBucket returns a CloudFormation-compatible Arn expression
// (string or Ref) suitable for template reference.  The bucket
// parameter may be either a string or an interface{} ("Ref: "myResource")
// value
func S3ArnForBucket(bucket interface{}) *gocf.StringExpr {
	arnParts := []gocf.Stringable{gocf.String("arn:aws:s3:::")}

	switch bucket.(type) {
	case string:
		// Don't be smart if the Arn value is a user supplied literal
		arnParts = append(arnParts, gocf.String(bucket.(string)))
	case *gocf.StringExpr:
		arnParts = append(arnParts, bucket.(*gocf.StringExpr))
	case gocf.RefFunc:
		arnParts = append(arnParts, bucket.(gocf.RefFunc).String())
	default:
		panic(fmt.Sprintf("Unsupported SourceArn value type: %+v", bucket))
	}
	return gocf.Join("", arnParts...).String()
}

// MapToResourceTags transforms a go map[string]string to a CloudFormation-compliant
// Tags representation.  See http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-resource-tags.html
func MapToResourceTags(tagMap map[string]string) []interface{} {
	var tags []interface{}
	for eachKey, eachValue := range tagMap {
		tags = append(tags, map[string]interface{}{
			"Key":   eachKey,
			"Value": eachValue,
		})
	}
	return tags
}
