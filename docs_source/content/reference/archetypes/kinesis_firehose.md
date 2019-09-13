---
date: 2018-11-28 20:03:46
title: Kinesis Firehose
weight: 10
---

There are two ways to create a Firehose Transform reactor that transforms a [KinesisFirehoseEventRecord](https://github.com/aws/aws-lambda-go/blob/master/events/firehose.go#L14) with a [Lambda function](https://aws.amazon.com/blogs/compute/amazon-kinesis-firehose-data-transformation-with-aws-lambda/):

- `NewKinesisFirehoseLambdaTransformer`
  - Transform using a Lambda function
- `NewKinesisFirehoseTransformer`
  - Transform using a go [text/template](https://golang.org/pkg/text/template/) declaration

## NewKinesisFirehoseLambdaTransformer

```go
import (
awsEvents "github.com/aws/aws-lambda-go/events"
  spartaArchetype "github.com/mweagle/Sparta/archetype"
)
// KinesisStream reactor function
func reactorFunc(ctx context.Context,
                    record *awsEvents.KinesisFirehoseEventRecord)
                    (*awsEvents.KinesisFirehoseResponseRecord, error) {
  logger, _ := ctx.Value(sparta.ContextKeyRequestLogger).(*logrus.Entry)

  logger.WithFields(logrus.Fields{
    "Record": record,
  }).Info("Kinesis Firehose Event")

  responseRecord = &awsEvents.KinesisFirehoseResponseRecord{
    RecordID: record.RecordID,
    Result:   awsEvents.KinesisFirehoseTransformedStateOk,
    Data:     record.Data,
  }
  return responseRecord, nil
}

func main() {
  // ...
  handler := spartaArchetype.KinesisFirehoseReactorFunc(reactorFunc)
  lambdaFn, lambdaFnErr := spartaArchetype.NewKinesisFirehoseLambdaTransformer(handler,
    5*time.Minute /* Duration: recommended minimum of 1m */)
  // ...
}
```

This is the lowest level transformation type supported and it enables the most flexibility.

## NewKinesisFirehoseTransformer

Another option for creating Kinesis Firehose Transformers is to leverage the [text/template](https://golang.org/pkg/text/template/) package to define a transformation template. For instance:

```text
{{/* file: transform.template */}}
{{if eq (.Record.Data.JMESPathAsString "sector") "TECHNOLOGY"}}
{
    "region" : "{{ .Record.KinesisEventHeader.Region }}",
    "ticker_symbol" : {{ .Record.Data.JMESPath "ticker_symbol"}}
}
{{else}}
{{ KinesisFirehoseDrop }}
{{end}}
```

A new `*sparta.LambdaAWSInfo` instance can be created from _transform.template_ as in:

```go
func main() {
  // ...
  hooks := &sparta.WorkflowHooks{}
  reactorFunc, reactorFuncErr := archetype.NewKinesisFirehoseTransformer("transform.template",
    5*time.Minute,
    hooks)
  // ...

  var lambdaFunctions []*sparta.LambdaAWSInfo
  lambdaFunctions = append(lambdaFunctions, reactorFunc)
  err := sparta.MainEx(awsName,
    "Simple Sparta application that demonstrates core functionality",
    lambdaFunctions,
    nil,
    nil,
    hooks,
    false)
}
```

The template execution context includes the the following:

### Data Model

- _Data_ (`string`)
  - The [data](https://github.com/aws/aws-lambda-go/blob/master/events/firehose.go#L17) available in the Kinesis Firehose Record. Values can be extracted from the Data content by either [JMESPath](https://github.com/jmespath/go-jmespath) expressions (`JMESPath`, `JMESPathAsString`, `JMESPathAsFormattedString`) or [regexp capture groups](https://golang.org/pkg/regexp/#pkg-overview) (`RegExpGroup`, `RegExpGroupAsString`, `RegExpGroupAsJSON`).
  - See for more information
- _RecordID_ (`string`)
  - The specific record id being processed
- _Metadata_ (`struct`)
  - The [metadata](https://github.com/aws/aws-lambda-go/blob/master/events/firehose.go#L38) associated with the specific record being processed
- _ApproximateArrivalTimestamp_ (`awsEvents.MilliSecondsEpochTime`)
  - The time at which the record arrived
- _KinesisEventHeader_ (`struct`)
  - Metadata associated with the set of records being processed

### Functions

Functions available in the template's [FuncMap](https://golang.org/pkg/text/template/#FuncMap) include:

- `KinesisFirehoseDrop`: indicates that the record should be marked as [KinesisFirehoseTransformedStateDropped](https://github.com/aws/aws-lambda-go/blob/master/events/firehose.go#L24)
- The set of [masterminds/sprig](https://masterminds.github.io/sprig/) functions available in [TxtFuncMap](https://godoc.org/github.com/Masterminds/sprig#TxtFuncMap)
