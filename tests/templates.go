package tests

import (
	"context"

	gocc "github.com/mweagle/go-cloudcondenser"
	gocf "github.com/mweagle/go-cloudformation"
)

////////////////////////////////////////////////////////////////////////////////
//

// simple template for testing
var simple = gocc.CloudFormationCondenser{
	Description: "My Stack",
	Resources: []interface{}{
		// Dynamically assigned resource name
		gocf.IAMRole{},
		// User defined resource name
		gocc.Static("MyResource", gocf.S3Bucket{
			BucketName: gocf.String("MyS3Bucket"),
		}),
	},
}

////////////////////////////////////////////////////////////////////////////////
//

// emptyProvider is a free function that returns no resource
func emptyProvider(ctx context.Context, template *gocf.Template) (context.Context, error) {
	return ctx, nil
}

// simpleConditional template for testing
var simpleConditional = gocc.CloudFormationCondenser{
	Description: "My Stack",
	Resources: []interface{}{
		// Dynamically assigned resource name
		gocf.IAMRole{},
		// User defined resource name
		gocc.Static("MyResource", gocf.S3Bucket{
			BucketName: gocf.String("MyS3Bucket"),
		}),
		gocc.ProviderFunc(emptyProvider),
	},
}

type multiProvider struct {
}

func (sg *multiProvider) Annotate(ctx context.Context, template *gocf.Template) (context.Context, error) {
	sampleRole := &gocf.IAMRole{
		RoleName: gocf.String("CustomRole"),
	}
	template.AddResource("MultiRole", sampleRole)

	// Add a bucket
	template.AddResource("MultiBucket", &gocf.S3Bucket{
		BucketName: gocf.String("CustomBucket"),
	})
	return ctx, nil
}

////////////////////////////////////////////////////////////////////////////////
//

// MultipleResourceProvider is a free function to return a sample generator instance
func MultipleResourceProvider() gocc.ResourceProvider {
	return &multiProvider{}
}

// multiTemplate for a simple test
var multiTemplate = gocc.CloudFormationCondenser{
	Description: "My Stack",
	Resources: []interface{}{
		// Multiple resources "lifted" to WebsiteResourcesXXX
		// entries. Encapsulate logical sets of related
		// CloudFormation resources
		gocc.Flatten("Rez", MultipleResourceProvider()),
	},
}

////////////////////////////////////////////////////////////////////////////////
//

// freeProvider is a free function that returns one resource
func freeProvider(ctx context.Context, template *gocf.Template) (context.Context, error) {
	template.AddResource("FreeResource", &gocf.IAMRole{
		RoleName: gocf.String("Free"),
	})
	return ctx, nil
}

// simpleConditional template for testing
var freeTemplate = gocc.CloudFormationCondenser{
	Description: "My Stack",
	Resources: []interface{}{
		gocc.ProviderFunc(freeProvider),
	},
}
