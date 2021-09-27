package tests

import (
	"context"

	gof "github.com/awslabs/goformation/v5/cloudformation"
	gofiam "github.com/awslabs/goformation/v5/cloudformation/iam"
	gofs3 "github.com/awslabs/goformation/v5/cloudformation/s3"
	gocc "github.com/mweagle/go-cloudcondenser"
)

////////////////////////////////////////////////////////////////////////////////
//

// simple template for testing
var simple = gocc.CloudFormationCondenser{
	Description: "My Stack",
	Resources: []interface{}{
		// Dynamically assigned resource name
		&gofiam.Role{},
		// User defined resource name
		gocc.Static("MyResource", &gofs3.Bucket{
			BucketName: "MyS3Bucket",
		}),
	},
}

////////////////////////////////////////////////////////////////////////////////
//

// emptyProvider is a free function that returns no resource
func emptyProvider(ctx context.Context, template *gof.Template) (context.Context, error) {
	return ctx, nil
}

// simpleConditional template for testing
var simpleConditional = gocc.CloudFormationCondenser{
	Description: "My Stack",
	Resources: []interface{}{
		// Dynamically assigned resource name
		&gofiam.Role{},
		// User defined resource name
		gocc.Static("MyResource", &gofs3.Bucket{
			BucketName: "MyS3Bucket",
		}),
		gocc.ProviderFunc(emptyProvider),
	},
}

type multiProvider struct {
}

func (sg *multiProvider) Annotate(ctx context.Context, template *gof.Template) (context.Context, error) {
	sampleRole := &gofiam.Role{
		RoleName: "CustomRole",
	}
	template.Resources["MultiRole"] = sampleRole

	// Add a bucket
	template.Resources["MultiBucket"] = &gofs3.Bucket{
		BucketName: "CustomBucket",
	}
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
func freeProvider(ctx context.Context, template *gof.Template) (context.Context, error) {
	template.Resources["FreeResource"] = &gofiam.Role{
		RoleName: "Free",
	}
	return ctx, nil
}

// simpleConditional template for testing
var freeTemplate = gocc.CloudFormationCondenser{
	Description: "My Stack",
	Resources: []interface{}{
		gocc.ProviderFunc(freeProvider),
	},
}
