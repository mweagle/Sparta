package archetype

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/Masterminds/sprig"
	awsEvents "github.com/aws/aws-lambda-go/events"
	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	sparta "github.com/mweagle/Sparta/v3"
	"github.com/mweagle/Sparta/v3/archetype/xformer"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

func dropError() error {
	return errors.New("KinesisFirehoseDrop")
}

// TemplateFileName is the name of the file in the ZIP archive
const TemplateFileName = "xform.template"
const xformResourcePrefix = "firehosexform_"
const envVarKinesisFirehoseTransformName = "SPARTA_KINESIS_FIREHOSE_TRANSFORM"

// KinesisFirehoseReactor represents a lambda function that responds to Dynamo  messages
type KinesisFirehoseReactor interface {
	// OnKinesisFirehoseRecord when an Kinesis reocrd occurs.
	OnKinesisFirehoseRecord(ctx context.Context,
		record *awsEvents.KinesisFirehoseEventRecord) (*awsEvents.KinesisFirehoseResponseRecord, error)
}

// KinesisFirehoseReactorFunc is a free function that adapts a KinesisFirehoseReactor
// compliant signature into a function that exposes an OnEvent
// function
type KinesisFirehoseReactorFunc func(ctx context.Context,
	kinesisRecord *awsEvents.KinesisFirehoseEventRecord) (*awsEvents.KinesisFirehoseResponseRecord, error)

// OnKinesisFirehoseRecord satisfies the KinesisFirehoseReactor interface
func (reactorFunc KinesisFirehoseReactorFunc) OnKinesisFirehoseRecord(ctx context.Context,
	kinesisRecord *awsEvents.KinesisFirehoseEventRecord) (*awsEvents.KinesisFirehoseResponseRecord, error) {
	return reactorFunc(ctx, kinesisRecord)
}

// ReactorName provides the name of the reactor func
func (reactorFunc KinesisFirehoseReactorFunc) ReactorName() string {
	return runtime.FuncForPC(reflect.ValueOf(reactorFunc).Pointer()).Name()
}

// NewKinesisFirehoseLambdaTransformer returns a new firehose proocessor that supports
// transforming records.
func NewKinesisFirehoseLambdaTransformer(reactor KinesisFirehoseReactor,
	timeout time.Duration) (*sparta.LambdaAWSInfo, error) {

	reactorLambda := func(ctx context.Context,
		kinesisFirehoseEvent awsEvents.KinesisFirehoseEvent) (interface{}, error) {
		// Apply the transform to each record and see
		// what it says

		response := &awsEvents.KinesisFirehoseResponse{
			Records: make([]awsEvents.KinesisFirehoseResponseRecord,
				len(kinesisFirehoseEvent.Records)),
		}

		var responseRecord *awsEvents.KinesisFirehoseResponseRecord
		var responseRecordErr error
		for eachIndex, eachRecord := range kinesisFirehoseEvent.Records {
			responseRecord, responseRecordErr = reactor.OnKinesisFirehoseRecord(ctx, &eachRecord)
			if responseRecordErr != nil {
				return nil, errors.Wrapf(responseRecordErr, "Failed to transform record")
			}
			if responseRecord == nil {
				responseRecord = &awsEvents.KinesisFirehoseResponseRecord{
					RecordID: eachRecord.RecordID,
					Result:   awsEvents.KinesisFirehoseTransformedStateDropped,
					Data:     eachRecord.Data,
				}
			}
			response.Records[eachIndex] = *responseRecord
		}
		return response, nil
	}

	lambdaFn, lambdaFnErr := sparta.NewAWSLambda(reactorName(reactor),
		reactorLambda,
		sparta.IAMRoleDefinition{})
	return lambdaFn, lambdaFnErr
}

// NewKinesisFirehoseTransformer returns a new firehose proocessor that supports
// transforming records.
func NewKinesisFirehoseTransformer(xformFilePath string,
	timeout time.Duration,
	hooks *sparta.WorkflowHooks) (*sparta.LambdaAWSInfo, error) {

	baseName := filepath.Base(xformFilePath)
	archiveEntryName := sparta.CloudFormationResourceName(xformResourcePrefix, xformFilePath)
	lambdaName := fmt.Sprintf("Firehose%s", baseName)

	// Return a lambda function that applies the XForm transformation
	reactorLambda := func(ctx context.Context,
		kinesisEvent awsEvents.KinesisFirehoseEvent) (*awsEvents.KinesisFirehoseResponse, error) {
		return lambdaXForm(ctx, kinesisEvent)
	}
	lambdaFn, lambdaFnErr := sparta.NewAWSLambda(lambdaName,
		reactorLambda,
		sparta.IAMRoleDefinition{})
	if lambdaFnErr != nil {
		return nil, errors.Wrapf(lambdaFnErr, "attempting to create Kinesis Firehose reactor")
	}

	// Borrow the resource name creator to get a name for the archive
	lambdaFn.Options.Environment[envVarKinesisFirehoseTransformName] = archiveEntryName
	lambdaFn.Options.Timeout = (int)(timeout.Milliseconds() / 1000)

	// Create the decorator that adds the file to the ZIP archive using
	// the transform name...
	archiveDecorator := func(ctx context.Context,
		serviceName string,
		zipWriter *zip.Writer,
		awsConfig awsv2.Config,
		noop bool,
		logger *zerolog.Logger) (context.Context, error) {
		fileInfo, fileInfoErr := os.Stat(xformFilePath)
		if fileInfoErr != nil {
			return ctx, errors.Wrapf(fileInfoErr, "Failed to get fileInfo for Kinesis Firehose transform")
		}
		// G304: Potential file inclusion via variable
		/* #nosec */
		fileReader, fileReaderErr := os.Open(xformFilePath)
		if fileReaderErr != nil {
			return ctx, errors.Wrapf(fileReaderErr, "Failed to open Kinesis Firehose transform file")
		}
		/* #nosec */
		defer func() {
			closeErr := fileReader.Close()

			if closeErr != nil {
				logger.Warn().
					Err(closeErr).
					Msg("Failed to close file reader")
			}
		}()

		fileHeader, fileHeaderErr := zip.FileInfoHeader(fileInfo)
		if fileHeaderErr != nil {
			return ctx, errors.Wrapf(fileHeaderErr, "Failed to detect ZIP header for Kinesis Firehose transform")
		}

		fileHeader.Name = archiveEntryName
		fileHeader.Method = zip.Deflate

		// Copy it...
		writer, writerErr := zipWriter.CreateHeader(fileHeader)
		if writerErr != nil {
			return ctx, errors.Wrapf(fileHeaderErr, "Failed to create ZIP header for Kinesis Firehose transform")
		}
		_, copyErr := io.Copy(writer, fileReader)
		return ctx, copyErr
	}
	// Done...
	hooks.Archives = append(hooks.Archives, sparta.ArchiveHookFunc(archiveDecorator))
	return lambdaFn, nil
}

// ApplyTransformToKinesisFirehoseEvent is the generic transformation function that applies
// a template.Template transformation to each
func ApplyTransformToKinesisFirehoseEvent(ctx context.Context,
	templateBytes []byte,
	kinesisEvent awsEvents.KinesisFirehoseEvent) (*awsEvents.KinesisFirehoseResponse, error) {

	logger, loggerOk := ctx.Value(sparta.ContextKeyLogger).(*zerolog.Logger)
	if loggerOk {
		logger.Info().Msg("Hello world structured log message")
	}

	funcMap := sprig.TxtFuncMap()
	funcMap["KinesisFirehoseDrop"] = interface{}(func() (string, error) {
		return "", dropError()
	})

	// Setup the function map that knows how to do the JMESPath
	// given the map...
	transform, transformErr := template.
		New("xformer").
		Funcs(funcMap).
		Parse(string(templateBytes))
	if transformErr != nil {
		return nil, errors.Wrapf(transformErr, "Attempting to create template")
	}

	response := &awsEvents.KinesisFirehoseResponse{
		Records: make([]awsEvents.KinesisFirehoseResponseRecord, len(kinesisEvent.Records)),
	}
	headerInfo := &xformer.KinesisEventHeaderInfo{
		InvocationID: kinesisEvent.InvocationID,

		DeliveryStreamArn:      kinesisEvent.DeliveryStreamArn,
		SourceKinesisStreamArn: kinesisEvent.SourceKinesisStreamArn,
		Region:                 kinesisEvent.Region,
	}

	for eachIndex, eachRecord := range kinesisEvent.Records {
		xformedRecord := awsEvents.KinesisFirehoseResponseRecord{
			RecordID: eachRecord.RecordID,
			Result:   awsEvents.KinesisFirehoseTransformedStateDropped,
			Data:     eachRecord.Data,
		}
		xform, xformErr := xformer.NewKinesisFirehoseEventXFormer(headerInfo, &eachRecord)
		if xformErr == nil {
			dataMap := map[string]interface{}{
				"Record": xform,
			}
			var outputBuffer bytes.Buffer
			templateErr := transform.Execute(&outputBuffer, dataMap)
			if templateErr != nil {
				// Is the fail value "KinesisFirehoseDrop" ?
				if !strings.Contains(templateErr.Error(), dropError().Error()) {
					xformedRecord.Result = awsEvents.KinesisFirehoseTransformedStateProcessingFailed
				}
			} else if xform.Error() != nil {
				xformedRecord.Result = awsEvents.KinesisFirehoseTransformedStateProcessingFailed
			} else {
				if loggerOk && logger.GetLevel() >= (zerolog.DebugLevel) {
					logger.Debug().
						Str("input", string(eachRecord.Data)).
						Str("output", outputBuffer.String()).
						Msg("Transformation result")
				}

				xformedRecord.Data = outputBuffer.Bytes()
				xformedRecord.Result = awsEvents.KinesisFirehoseTransformedStateOk
			}
		}
		// Save it...
		response.Records[eachIndex] = xformedRecord
	}
	return response, nil
}
