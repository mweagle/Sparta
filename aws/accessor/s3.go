package accessor

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	sparta "github.com/mweagle/Sparta"
	spartaAWS "github.com/mweagle/Sparta/aws"
	spartaCF "github.com/mweagle/Sparta/aws/cloudformation"
	gocf "github.com/mweagle/go-cloudformation"
	"github.com/rs/zerolog"
)

// S3Accessor to make it a bit easier to work with S3
// as the backing store
type S3Accessor struct {
	testingBucketName    string
	S3BucketResourceName string
}

// BucketPrivilege returns a privilege that targets the Bucket
func (svc *S3Accessor) BucketPrivilege(bucketPrivs ...string) sparta.IAMRolePrivilege {
	return sparta.IAMRolePrivilege{
		Actions:  bucketPrivs,
		Resource: spartaCF.S3ArnForBucket(gocf.Ref(svc.S3BucketResourceName)),
	}
}

// KeysPrivilege returns a privilege that targets the Bucket objects
func (svc *S3Accessor) KeysPrivilege(keyPrivileges ...string) sparta.IAMRolePrivilege {
	return sparta.IAMRolePrivilege{
		Actions:  keyPrivileges,
		Resource: spartaCF.S3AllKeysArnForBucket(gocf.Ref(svc.S3BucketResourceName)),
	}
}

func (svc *S3Accessor) s3Svc(ctx context.Context) *s3.S3 {
	logger, _ := ctx.Value(sparta.ContextKeyLogger).(*zerolog.Logger)
	sess := spartaAWS.NewSession(logger)
	s3Client := s3.New(sess)
	xrayInit(s3Client.Client)
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
	s3BucketRes, s3BucketResExists := discover.Resources[svc.S3BucketResourceName]
	if !s3BucketResExists {
		return ""
	}
	return s3BucketRes.ResourceRef
}

// Delete handles deleting the resource
func (svc *S3Accessor) Delete(ctx context.Context, keyPath string) error {
	deleteObjectInput := &s3.DeleteObjectInput{
		Bucket: aws.String(svc.s3BucketName()),
		Key:    aws.String(keyPath),
	}
	_, deleteResultErr := svc.
		s3Svc(ctx).
		DeleteObjectWithContext(ctx, deleteObjectInput)

	return deleteResultErr
}

// DeleteAll handles deleting all the items
func (svc *S3Accessor) DeleteAll(ctx context.Context) error {
	// List each one, delete it

	listObjectInput := &s3.ListObjectsInput{
		Bucket: aws.String(svc.s3BucketName()),
	}

	listObjectResult, listObjectResultErr := svc.
		s3Svc(ctx).
		ListObjectsWithContext(ctx, listObjectInput)

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
	putObjectInput := &s3.PutObjectInput{
		Bucket: aws.String(svc.s3BucketName()),
		Key:    aws.String(keyPath),
		Body:   bytesReader,
	}
	putObjectResponse, putObjectRespErr := svc.
		s3Svc(ctx).
		PutObjectWithContext(ctx, putObjectInput)

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

	getObjectInput := &s3.GetObjectInput{
		Bucket: aws.String(svc.s3BucketName()),
		Key:    aws.String(keyPath),
	}
	getObjectResult, getObjectResultErr := svc.
		s3Svc(ctx).
		GetObjectWithContext(ctx, getObjectInput)
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

	listObjectInput := &s3.ListObjectsInput{
		Bucket: aws.String(svc.s3BucketName()),
	}

	listObjectResult, listObjectResultErr := svc.
		s3Svc(ctx).
		ListObjectsWithContext(ctx, listObjectInput)

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
