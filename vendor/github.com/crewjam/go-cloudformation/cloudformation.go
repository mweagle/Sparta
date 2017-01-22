// Package cloudformation provides a schema and related functions that allow
// you to reason about cloudformation template documents.
//
// Parsing example:
//
// 	t := Template{}
// 	json.NewDecoder(os.Stdin).Decode(&t)
//
// Producing Example:
//
// 	t := NewTemplate()
// 	t.Parameters["DnsName"] = &Parameter{
// 		Type: "string",
// 		Default: "example.com",
// 		Description: "the top level DNS name for the service"
// 	}
// 	t.AddResource("DataBucket", &S3Bucket{
// 		BucketName: Join("-", *String("data"), *Ref("DnsName").String())
// 	})
// 	json.NewEncoder(os.Stdout).Encoder(t)
//
// See the examples directory for a more complete example of producing a
// cloudformation template from code.
//
// Producing the Schema
//
// As far as I can tell, AWS do not produce a structured document that
// describes the Cloudformation schema. The names and types for the
// various resources and objects are derived from scraping their HTML
// documentation (see scraper/). It is mostly, but not entirely,
// complete. I've noticed several inconsistencies in the documentation
// which suggests that it is constructed by hand. If you run into
// problems, please submit a bug (or better yet, a pull request).
//
// Object Types
//
// Top level objects in Cloudformation are called resources. They have
// names like AWS::S3::Bucket and appear as values in the "Resources"
// mapping. We remove the punctuation from the name to derive a golang
// structure name like S3Bucket.
//
// There other non-resource structures that are refered to either by
// resources or by other structures. These objects have names with
// spaces like "Amazon S3 Versioning Configuration". To derive a golang
// type name the non-letter characters are removed to get
// S3VersioningConfiguration.
//
// Type System
//
// Cloudformation uses three scalar types: string, int and bool. When
// they appear as properties we represent them as *StringExpr, *IntegerExpr,
// and *BoolExpr respectively. These types reflect that fact that a
// scalar type could be a literal string, int or bool, or could be a
// JSON dictionary representing a function call. (The *Expr structs have
// custom MarshalJSON and UnmarshalJSON that account for this)
//
// Another vagary of the cloudformation language is that in cases where
// a list of objects is expects, a single object can provided. To account
// for this, whenever a list of objects appears, a custom type *WhateverList
// is used. This allows us to add a custom UnmarshalJSON which transforms
// an object into a list containing an object.
//
package cloudformation
