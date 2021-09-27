package templates

import (
	"context"

	gof "github.com/awslabs/goformation/v5/cloudformation"
	gofiam "github.com/awslabs/goformation/v5/cloudformation/iam"
	gofs3 "github.com/awslabs/goformation/v5/cloudformation/s3"
	gocc "github.com/mweagle/go-cloudcondenser"
)

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

// MultipleResourceProvider is a free function to return a sample generator instance
func MultipleResourceProvider() gocc.ResourceProvider {
	return &multiProvider{}
}

// freeProvider is a free function that annotates the template
func freeProvider(ctx context.Context, template *gof.Template) (context.Context, error) {
	template.Resources["FreeResource"] = &gofs3.Bucket{
		BucketName: "FreeBucket",
	}
	return ctx, nil
}

// emptyProvider is a free function that returns no resource
func emptyProvider(ctx context.Context, template *gof.Template) (context.Context, error) {
	return ctx, nil
}

// DefaultTemplate for a simple test. ResourceProviders have
// access to a context object with command line args for
// internal conditionals
var DefaultTemplate = gocc.CloudFormationCondenser{
	Description: "My Stack",
	Resources: []interface{}{
		// Dynamically assigned resource name
		gofiam.Role{},
		// User defined resource name
		gocc.Static("MyResource", &gofs3.Bucket{
			BucketName: "MyS3Bucket",
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
