package xformer

import (
	"encoding/json"
	"fmt"
	"regexp"

	awsEvents "github.com/aws/aws-lambda-go/events"
	jmesPath "github.com/jmespath/go-jmespath"
)

type xformRegexpResults struct {
	re            *regexp.Regexp
	reErr         error
	capturedNames map[string]string
}

// XFormableData is the struct that supports transforming the record
type XFormableData struct {
	rawData    []byte
	stringData string
	jsonData   map[string]interface{}
	xformErr   error
	regexps    map[string]*xformRegexpResults
}

func (xf *XFormableData) regexpGroupValue(regexpString string, groupName string) interface{} {
	results, resultsExist := xf.regexps[regexpString]
	if !resultsExist {
		if xf.stringData == "" {
			xf.stringData = string(xf.rawData)
		}
		// Run it...
		re, reErr := regexp.Compile(regexpString)
		xformResults := &xformRegexpResults{
			re:    re,
			reErr: reErr,
		}
		// Parse it...
		if reErr == nil {
			xformResults.capturedNames = make(map[string]string)
			matches := xformResults.re.FindStringSubmatch(xf.stringData)
			if len(matches) > 0 {
				for i, eachName := range xformResults.re.SubexpNames() {
					// Ignore the whole regexp match and unnamed groups
					if i == 0 || eachName == "" {
						continue
					}
					xformResults.capturedNames[eachName] = matches[i]
				}
			}
		}
		xf.regexps[regexpString] = xformResults
		results = xformResults
	}
	if results == nil || results.reErr != nil {
		return nil
	}
	return results.capturedNames[groupName]
}

// RegExpGroup captures the output as a String
func (xf *XFormableData) RegExpGroup(regexpString string, groupName string) interface{} {
	value := xf.regexpGroupValue(regexpString, groupName)
	if value == nil {
		return ""
	}
	return xf.RegExpGroupAsFormattedString(regexpString, groupName, "%s")
}

// RegExpGroupAsFormattedString captures the output as a String
func (xf *XFormableData) RegExpGroupAsFormattedString(regexpString string,
	groupName string,
	formatSpecifier string) interface{} {
	value := xf.regexpGroupValue(regexpString, groupName)
	if value == nil {
		return ""
	}
	return fmt.Sprintf(formatSpecifier, value)
}

// RegExpGroupAsJSON captures the output as a JSON blob
func (xf *XFormableData) RegExpGroupAsJSON(regexpString string, groupName string) interface{} {
	value := xf.regexpGroupValue(regexpString, groupName)
	if value == nil {
		return nil
	}
	marshalResult, marshalResultErr := json.Marshal(value)
	if marshalResultErr != nil {
		xf.xformErr = marshalResultErr
		return nil
	}
	return string(marshalResult)
}

func (xf *XFormableData) jmesValue(pathSpec string) interface{} {
	if xf.jsonData == nil {
		xf.jsonData = make(map[string]interface{})
		unmarshalErr := json.Unmarshal(xf.rawData, &xf.jsonData)
		if unmarshalErr != nil {
			xf.xformErr = unmarshalErr
			return nil
		}
	}

	precompiled, compileErr := jmesPath.Compile(pathSpec)
	if compileErr != nil {
		xf.xformErr = compileErr
		return nil
	}
	result, resultErr := precompiled.Search(xf.jsonData)
	if resultErr != nil {
		xf.xformErr = compileErr
		return nil
	}
	return result
}

// JMESPath runs the query
func (xf *XFormableData) JMESPath(pathSpec string) interface{} {
	value := xf.jmesValue(pathSpec)
	if value == nil {
		return nil
	}
	marshalResult, marshalResultErr := json.Marshal(value)
	if marshalResultErr != nil {
		xf.xformErr = marshalResultErr
		return nil
	}
	return string(marshalResult)
}

// JMESPathAsString returns the value at pathSpec using a string formatter
func (xf *XFormableData) JMESPathAsString(pathSpec string) interface{} {
	return xf.JMESPathAsFormattedString(pathSpec, "%s")
}

// JMESPathAsFormattedString returns the value at pathSpec using the provided go
// format specifier
func (xf *XFormableData) JMESPathAsFormattedString(pathSpec string, formatSpecifier string) interface{} {
	value := xf.jmesValue(pathSpec)
	if value == nil {
		return nil
	}
	return fmt.Sprintf(formatSpecifier, value)
}

// KinesisEventHeaderInfo is the common header info in the KinesisFirehoseInput event
type KinesisEventHeaderInfo struct {
	InvocationID           string
	DeliveryStreamArn      string
	SourceKinesisStreamArn string
	Region                 string
}

// XFormer is the transformer
type XFormer struct {
	Data                        *XFormableData
	KinesisEventHeader          *KinesisEventHeaderInfo
	RecordID                    string
	Metadata                    *awsEvents.KinesisFirehoseRecordMetadata
	ApproximateArrivalTimestamp awsEvents.MilliSecondsEpochTime
}

// Error returns the potentially nil error from applying the
// transform
func (xf *XFormer) Error() error {
	return xf.Data.xformErr
}

// NewKinesisFirehoseEventXFormer returns a new XFormer pointer initialized with the
// awsEvents.KinesisFirehoseEventRecord data
func NewKinesisFirehoseEventXFormer(kinesisEventHeader *KinesisEventHeaderInfo,
	record *awsEvents.KinesisFirehoseEventRecord) (*XFormer, error) {
	return &XFormer{
		Data: &XFormableData{
			rawData: record.Data,
			regexps: make(map[string]*xformRegexpResults),
		},
		KinesisEventHeader:          kinesisEventHeader,
		RecordID:                    record.RecordID,
		Metadata:                    &record.KinesisFirehoseRecordMetadata,
		ApproximateArrivalTimestamp: record.ApproximateArrivalTimestamp,
	}, nil
}
