package accessor

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	awsv2S3 "github.com/aws/aws-sdk-go-v2/service/s3"

	sparta "github.com/mweagle/Sparta/v3"
	spartaAWS "github.com/mweagle/Sparta/v3/aws"
	spartaCF "github.com/mweagle/Sparta/v3/aws/cloudformation"
	"github.com/rs/zerolog"
)

// S3Accessor to make it a bit easier to work with S3
// as the backing store
type S3Accessor struct {
	testingBucketName        string
	S3BucketResourceOrArnRef string
}

// BucketPrivilege returns a privilege that targets the Bucket
func (svc *S3Accessor) BucketPrivilege(bucketPrivs ...string) sparta.IAMRolePrivilege {
	return sparta.IAMRolePrivilege{
		Actions:  bucketPrivs,
		Resource: svc.S3BucketResourceOrArnRef,
	}
}

// KeysPrivilege returns a privilege that targets the Bucket objects
func (svc *S3Accessor) KeysPrivilege(keyPrivileges ...string) sparta.IAMRolePrivilege {
	return sparta.IAMRolePrivilege{
		Actions:  keyPrivileges,
		Resource: spartaCF.S3AllKeysArnForBucket(svc.S3BucketResourceOrArnRef),
	}
}

func (svc *S3Accessor) s3Svc(ctx context.Context) *awsv2S3.Client {
	logger, _ := ctx.Value(sparta.ContextKeyLogger).(*zerolog.Logger)
	awsConfig, awsConfigErr := spartaAWS.NewConfig(ctx, logger)
	if awsConfigErr != nil {
		return nil
	}
	xrayInit(&awsConfig)
	s3Client := awsv2S3.NewFromConfig(awsConfig)
	return s3Client
}

func (svc *S3Accessor) s3BucketName() string {
	if svc.testingBucketName != "" {
		return svc.testingBucketName
	}
	discover, discoveryInfoErr := sparta.Discover()
	if discoveryInfoErr != nil {
		return ""
	}
	s3BucketRes, s3BucketResExists := discover.Resources[svc.S3BucketResourceOrArnRef]
	if !s3BucketResExists {
		return ""
	}
	return s3BucketRes.ResourceRef
}

// Delete handles deleting the resource
func (svc *S3Accessor) Delete(ctx context.Context, keyPath string) error {
	deleteObjectInput := &awsv2S3.DeleteObjectInput{
		Bucket: awsv2.String(svc.s3BucketName()),
		Key:    awsv2.String(keyPath),
	}
	_, deleteResultErr := svc.
		s3Svc(ctx).
		DeleteObject(ctx, deleteObjectInput)

	return deleteResultErr
}

// DeleteAll handles deleting all the items
func (svc *S3Accessor) DeleteAll(ctx context.Context) error {
	// List each one, delete it

	listObjectInput := &awsv2S3.ListObjectsInput{
		Bucket: awsv2.String(svc.s3BucketName()),
	}

	listObjectResult, listObjectResultErr := svc.
		s3Svc(ctx).
		ListObjects(ctx, listObjectInput)

	if listObjectResultErr != nil {
		return nil
	}
	for _, eachObject := range listObjectResult.Contents {
		deleteErr := svc.Delete(ctx, *eachObject.Key)
		if deleteErr != nil {
			return deleteErr
		}
	}
	return nil
}

// Put handles saving the item
func (svc *S3Accessor) Put(ctx context.Context, keyPath string, object interface{}) error {
	jsonBytes, jsonBytesErr := json.Marshal(object)
	if jsonBytesErr != nil {
		return jsonBytesErr
	}

	logger, _ := ctx.Value(sparta.ContextKeyLogger).(*zerolog.Logger)
	logger.Debug().
		Str("Bytes", string(jsonBytes)).
		Str("KeyPath", keyPath).
		Msg("Saving S3 object")

	bytesReader := bytes.NewReader(jsonBytes)
	putObjectInput := &awsv2S3.PutObjectInput{
		Bucket: awsv2.String(svc.s3BucketName()),
		Key:    awsv2.String(keyPath),
		Body:   bytesReader,
	}
	putObjectResponse, putObjectRespErr := svc.
		s3Svc(ctx).
		PutObject(ctx, putObjectInput)

	logger.Debug().
		Err(putObjectRespErr).
		Interface("Results", putObjectResponse).
		Msg("Save object results")

	return putObjectRespErr
}

// Get handles getting the item
func (svc *S3Accessor) Get(ctx context.Context,
	keyPath string,
	destObject interface{}) error {

	getObjectInput := &awsv2S3.GetObjectInput{
		Bucket: awsv2.String(svc.s3BucketName()),
		Key:    awsv2.String(keyPath),
	}
	getObjectResult, getObjectResultErr := svc.
		s3Svc(ctx).
		GetObject(ctx, getObjectInput)
	if getObjectResultErr != nil {
		return getObjectResultErr
	}
	jsonBytes, jsonBytesErr := ioutil.ReadAll(getObjectResult.Body)
	if jsonBytesErr != nil {
		return jsonBytesErr
	}
	return json.Unmarshal(jsonBytes, destObject)
}

// GetAll handles returning all of the items
func (svc *S3Accessor) GetAll(ctx context.Context,
	ctor NewObjectConstructor) ([]interface{}, error) {

	listObjectInput := &awsv2S3.ListObjectsInput{
		Bucket: awsv2.String(svc.s3BucketName()),
	}

	listObjectResult, listObjectResultErr := svc.
		s3Svc(ctx).
		ListObjects(ctx, listObjectInput)

	if listObjectResultErr != nil {
		return nil, listObjectResultErr
	}
	allObjects := make([]interface{}, 0)
	for _, eachObject := range listObjectResult.Contents {
		objectInstance := ctor()
		entryEntryErr := svc.Get(ctx, *eachObject.Key, objectInstance)
		if entryEntryErr != nil {
			return nil, entryEntryErr
		}
		allObjects = append(allObjects, objectInstance)
	}
	return allObjects, nil
}
