package templates

import (
	"context"

	gocc "github.com/mweagle/go-cloudcondenser"
	gocf "github.com/mweagle/go-cloudformation"
)

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

// MultipleResourceProvider is a free function to return a sample generator instance
func MultipleResourceProvider() gocc.ResourceProvider {
	return &multiProvider{}
}

// freeProvider is a free function that annotates the template
func freeProvider(ctx context.Context, template *gocf.Template) (context.Context, error) {
	template.AddResource("FreeResource", &gocf.S3Bucket{
		BucketName: gocf.String("FreeBucket"),
	})
	return ctx, nil
}

// emptyProvider is a free function that returns no resource
func emptyProvider(ctx context.Context, template *gocf.Template) (context.Context, error) {
	return ctx, nil
}

// DefaultTemplate for a simple test. ResourceProviders have
// access to a context object with command line args for
// internal conditionals
var DefaultTemplate = gocc.CloudFormationCondenser{
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
