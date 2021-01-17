package archetype

import (
	"context"
	"fmt"
	_ "net/http/pprof" // include pprop

	awsLambdaEvents "github.com/aws/aws-lambda-go/events"
	sparta "github.com/mweagle/Sparta"
	gocf "github.com/mweagle/go-cloudformation"
	"github.com/rs/zerolog"
)

// ExampleS3Reactor illustrates how to create an S3 event subscriber
func ExampleS3Reactor() {
	inlineReactor := func(ctx context.Context,
		s3Event awsLambdaEvents.S3Event) (interface{}, error) {
		logger, loggerOk := ctx.Value(sparta.ContextKeyLogger).(*zerolog.Logger)
		if loggerOk {
			for _, eachRecord := range s3Event.Records {
				logger.Info().
					Str("EventType", eachRecord.EventName).
					Interface("Entity", eachRecord.S3).
					Msg("Event info")
			}
		}
		return len(s3Event.Records), nil
	}
	// Create the *sparta.LambdaAWSInfo wrapper
	lambdaFn, lambdaFnErr := NewS3Reactor(S3ReactorFunc(inlineReactor),
		gocf.String("MY-S3-BUCKET-TO-REACT"),
		nil)
	fmt.Printf("LambdaFn: %#v, LambdaFnErr: %#v", lambdaFn, lambdaFnErr)
}

// ExampleSNSReactor illustrates how to create an SNS notification subscriber
func ExampleSNSReactor() {
	inlineReactor := func(ctx context.Context, snsEvent awsLambdaEvents.SNSEvent) (interface{}, error) {
		logger, loggerOk := ctx.Value(sparta.ContextKeyLogger).(*zerolog.Logger)
		if loggerOk {
			logger.Info().
				Interface("Event", snsEvent).
				Msg("Event received")
		}
		return &snsEvent, nil
	}
	// Create the *sparta.LambdaAWSInfo wrapper
	lambdaFn, lambdaFnErr := NewSNSReactor(SNSReactorFunc(inlineReactor),
		gocf.String("MY-SNS-TOPIC"),
		nil)
	fmt.Printf("LambdaFn: %#v, LambdaFnErr: %#v", lambdaFn, lambdaFnErr)
}
