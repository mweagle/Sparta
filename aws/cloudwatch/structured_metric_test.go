package cloudwatch

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/xeipuuv/gojsonschema"
)

func TestStructuredMetric(t *testing.T) {

	sink := &bytes.Buffer{}

	// Initialize with high cardinality property
	emMetric, _ := NewEmbeddedMetricWithProperties(map[string]interface{}{
		"testMetric": "42",
	})
	// Add a directive with a namespace that defines a metric. Metric values
	// may also be array
	metricDirective := emMetric.NewMetricDirective("SpecialNamespace",
		map[string]string{"functionVersion": "23"})

	metricDirective.Metrics["invocations"] = MetricValue{
		Unit:  UnitCount,
		Value: 1,
	}
	emMetric.Publish(map[string]interface{}{
		"additional": fmt.Sprintf("high cardinality prop: %d", time.Now().Unix()),
	})

	emMetric.PublishToSink(nil, sink)
	// Verify...
	schemaLoader := gojsonschema.NewReferenceLoader("file://./emf.schema.json")
	documentLoader := gojsonschema.NewBytesLoader(sink.Bytes())

	_, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		t.Fatalf("Failed to produce valid structured metric: %v", err)
	}

}
