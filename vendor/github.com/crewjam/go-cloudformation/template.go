package cloudformation

import (
	"encoding/json"
	"fmt"
)

// NewTemplate returns a new empty Template initialized with some
// default values.
func NewTemplate() *Template {
	return &Template{
		AWSTemplateFormatVersion: "2010-09-09",
		Mappings:                 map[string]*Mapping{},
		Parameters:               map[string]*Parameter{},
		Resources:                map[string]*Resource{},
		Outputs:                  map[string]*Output{},
		Conditions:               map[string]interface{}{},
	}
}

// Template represents a cloudformation template.
type Template struct {
	AWSTemplateFormatVersion string                 `json:",omitempty"`
	Description              string                 `json:",omitempty"`
	Mappings                 map[string]*Mapping    `json:",omitempty"`
	Parameters               map[string]*Parameter  `json:",omitempty"`
	Resources                map[string]*Resource   `json:",omitempty"`
	Outputs                  map[string]*Output     `json:",omitempty"`
	Conditions               map[string]interface{} `json:",omitempty"`
}

// AddResource adds the resource to the template as name, displacing
// any resource with the same name that already exists.
func (t *Template) AddResource(name string, resource ResourceProperties) *Resource {
	templateResource := &Resource{Properties: resource}
	t.Resources[name] = templateResource
	return templateResource
}

// Mapping matches a key to a corresponding set of named values. For example,
// if you want to set values based on a region, you can create a mapping that
// uses the region name as a key and contains the values you want to specify
// for each specific region. You use the Fn::FindInMap intrinsic function to
// retrieve values in a map.
//
// See http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/mappings-section-structure.html
type Mapping map[string]map[string]string

// Parameter represents a parameter to the template.
//
// You can use the optional Parameters section to pass values into your
// template when you create a stack. With parameters, you can create templates
// that are customized each time you create a stack. Each parameter must
// contain a value when you create a stack. You can specify a default value to
// make the parameter optional.
//
// See http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/parameters-section-structure.html
type Parameter struct {
	Type                  string       `json:",omitempty"`
	Default               string       `json:",omitempty"`
	NoEcho                *BoolExpr    `json:",omitempty"`
	AllowedValues         []string     `json:",omitempty"`
	AllowedPattern        string       `json:",omitempty"`
	MinLength             *IntegerExpr `json:",omitempty"`
	MaxLength             *IntegerExpr `json:",omitempty"`
	MinValue              *IntegerExpr `json:",omitempty"`
	MaxValue              *IntegerExpr `json:",omitempty"`
	Description           string       `json:",omitempty"`
	ConstraintDescription string       `json:",omitempty"`
}

// OutputExport represents the name of the resource output that should
// be used for cross stack references.
//
// See http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/walkthrough-crossstackref.html
type OutputExport struct {
	Name Stringable `json:",omitempty"`
}

// Output represents a template output
//
// The optional Outputs section declares output values that you want to view from the
// AWS CloudFormation console or that you want to return in response to describe stack calls.
// For example, you can output the Amazon S3 bucket name for a stack so that you can easily find it.
//
// See http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/outputs-section-structure.html
type Output struct {
	Description string        `json:",omitempty"`
	Value       interface{}   `json:",omitempty"`
	Export      *OutputExport `json:",omitempty"`
}

// ResourceProperties is an interface that is implemented by resource objects.
type ResourceProperties interface {
	CfnResourceType() string
}

// Resource represents a resource in a cloudformation template. It contains resource
// metadata and, in Properties, a struct that implements ResourceProperties which
// contains the properties of the resource.
type Resource struct {
	CreationPolicy *CreationPolicy
	DeletionPolicy string
	DependsOn      []string
	Metadata       map[string]interface{}
	UpdatePolicy   *UpdatePolicy
	Condition      string
	Properties     ResourceProperties
}

// MarshalJSON returns a JSON representation of the object
func (r Resource) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type           string
		CreationPolicy *CreationPolicy        `json:",omitempty"`
		DeletionPolicy string                 `json:",omitempty"`
		DependsOn      []string               `json:",omitempty"`
		Metadata       map[string]interface{} `json:",omitempty"`
		UpdatePolicy   *UpdatePolicy          `json:",omitempty"`
		Condition      string                 `json:",omitempty"`
		Properties     ResourceProperties
	}{
		Type:           r.Properties.CfnResourceType(),
		CreationPolicy: r.CreationPolicy,
		DeletionPolicy: r.DeletionPolicy,
		DependsOn:      r.DependsOn,
		Metadata:       r.Metadata,
		UpdatePolicy:   r.UpdatePolicy,
		Condition:      r.Condition,
		Properties:     r.Properties,
	})
}

// UnmarshalJSON sets the object from the provided JSON representation
func (r *Resource) UnmarshalJSON(buf []byte) error {
	m := map[string]interface{}{}
	if err := json.Unmarshal(buf, &m); err != nil {
		return err
	}

	typeName := m["Type"].(string)
	r.DependsOn, _ = m["DependsOn"].([]string)
	r.Metadata, _ = m["Metadata"].(map[string]interface{})
	r.DeletionPolicy, _ = m["DeletionPolicy"].(string)
	r.Properties = NewResourceByType(typeName)
	if r.Properties == nil {
		return fmt.Errorf("unknown resource type: %s", typeName)
	}

	propertiesBuf, err := json.Marshal(m["Properties"])
	if err != nil {
		return err
	}
	if err := json.Unmarshal(propertiesBuf, r.Properties); err != nil {
		return err
	}
	return nil
}
