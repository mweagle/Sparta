[![Build Status](https://travis-ci.org/crewjam/go-cloudformation.svg?branch=master)](https://travis-ci.org/crewjam/go-cloudformation) [![](https://godoc.org/github.com/crewjam/go-cloudformation?status.png)](https://godoc.org/github.com/crewjam/go-cloudformation)

This package provides a schema and related functions that allow you to parse and serialize CloudFormation templates in golang. The package places an emphasis on type-safety so that the templates it produces are (slightly) more likely to be correct, and maybe you can avoid endless cycles of `UPDATE_ROLLBACK_IN_PROGRESS`.

Parsing example:

```go
t := Template{}
json.NewDecoder(os.Stdin).Decode(&t)
fmt.Printf("DNS name: %s\n", t.Parameters["DnsName"].Default) 
```

Producing Example:

```go
t := NewTemplate()
t.Parameters["DnsName"] = &Parameter{
  Type: "string",
  Default: "example.com",
  Description: "the top level DNS name for the service"
}
t.AddResource("DataBucket", &S3Bucket{
  BucketName: Join("-", String("data"), Ref("DnsName"))
})
json.NewEncoder(os.Stdout).Encoder(t)
```

See the examples directory for a more complete example of producing a
cloudformation template from code.

## Producing the Schema

As far as I can tell, AWS do not produce a structured document that
describes the CloudFormation schema. The names and types for the
various resources and objects are derived from scraping their HTML
documentation (see scraper/). It is mostly, but not entirely,
complete. I've noticed several inconsistencies in the documentation
which suggests that it is constructed by hand. If you run into
problems, please submit a bug (or better yet, a pull request).

## Object Types

Top level objects in CloudFormation are called resources. They have
names like *AWS::S3::Bucket* and appear as values in the "Resources"
mapping. We remove the punctuation and redundant words from the name
to derive a golang structure name like *S3Bucket*.

There are other non-resource structs that are refered to by resources or other non-resource structs. These objects have names with
spaces in them, like "Amazon S3 Versioning Configuration". To derive a golang
type name the non-letter characters and redundant words are removed to get
*S3VersioningConfiguration*.

## Type System

CloudFormation uses three scalar types: *string*, *int* and *bool*. When
they appear as properties we represent them as `*StringExpr`, `*IntegerExpr`,
and `*BoolExpr` respectively. 

```go
type StringExpr struct {
  Func    StringFunc
  Literal string
}

// StringFunc is an interface provided by objects that represent 
// CloudFormation functions that can return a string value.
type StringFunc interface {
  Func
  String() *StringExpr
}
```

These types reflect that fact that a scalar type could be a literal value (`"us-east-1"`) or a JSON dictionary representing a "function call" (`{"Ref": "AWS::Region"}`).

Another vagary of the CloudFormation language is that in cases where
a list of objects is expected, a single object can provided. For example, 
`AutoScalingLaunchConfiguration` has a property `BlockDeviceMappings` which is a list of `AutoScalingBlockDeviceMapping`. Valid CloudFormation documents can specify a single `AutoScalingBlockDeviceMapping` rather than a list. To model this, we use a custom type `AutoScalingBlockDeviceMappingList` which is just a `[]AutoScalingBlockDeviceMapping` with extra functions attached so that a single items an be unserialized. JSON produced by this package will always be in the list-of-objects form, rather than the single object form.

## Known Issues

The `cloudformation.String("foo")` is cumbersome for scalar literals. On balance, I think it is the best way to handle the vagaries of the CloudFormation syntax, but that doesn't make it less kludgy. A similar approach is taken by aws-sdk-go (and is similarly cumbersome).

There are some types that are not parsed fully and appear as `interface{}`.

I worked through public template files I could find, making sure the 
library could accurately serialize and unserialize them. In this process
I discovered some of the idiosyncracies described. This package works for our purposes, but I wouldn't be surprised if there are more idiosyncracies hidden in parts of CloudFormation we are not using. 

Feedback, bug reports and pull requests are gratefully accepted.
