[![Build Status](https://travis-ci.org/mweagle/go-cloudcondenser.svg?branch=master)]

# go-cloudcondensor

Compose and compile CloudFormation templates written in Go.

# Overview

## Define Template

```go
var DefaultTemplate = gocc.CloudformationCondenser{
	Description: "My Stack",
	Resources: []interface{}{
		// Dynamically assigned resource name
		gocf.IAMRole{},
		// User defined resource name
		gocc.Static("MyResource", gocf.S3Bucket{
			BucketName: gocf.String("MyS3Bucket"),
		}),
		// Multiple resources "flattened" to WebsiteResourcesXXX
		// entries. Encapsulate logical sets of related
		// CloudFormation resources
		gocc.Flatten("WebsiteResources", MultipleResourceProvider()),
		// Include free functions to annotate the template
		gocc.ProviderFunc(freeProvider),
		// Providers can conditionally update the template
		gocc.ProviderFunc(emptyProvider),
	},
}
```

## Evaluate Template

```
	ctx := context.Background()
	outputTemplate, outputErr := DefaultTemplate.Evaluate(ctx)
```

## Convert to CLI

See _./cmd/main.go_ for a simple CLI app that produces
JSON output from a `gocc.CloudformationCondenser` instance:

## Results

```json
{
  "AWSTemplateFormatVersion": "2010-09-09",
  "Resources": {
    "CloudFormerAWSIAMRole0": {
      "Type": "AWS::IAM::Role",
      "Properties": {}
    },
    "FreeResource": {
      "Type": "AWS::S3::Bucket",
      "Properties": {
        "BucketName": "FreeBucket"
      }
    },
    "MyResource": {
      "Type": "AWS::S3::Bucket",
      "Properties": {
        "BucketName": "MyS3Bucket"
      }
    },
    "WebsiteResourcesMultiBucket": {
      "Type": "AWS::S3::Bucket",
      "Properties": {
        "BucketName": "CustomBucket"
      }
    },
    "WebsiteResourcesMultiRole": {
      "Type": "AWS::IAM::Role",
      "Properties": {
        "RoleName": "CustomRole"
      }
    }
  }
}
```
