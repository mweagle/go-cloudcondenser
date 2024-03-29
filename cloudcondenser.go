package cloudcondenser

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	gof "github.com/awslabs/goformation/v5/cloudformation"
	set "github.com/deckarep/golang-set"
	"github.com/pkg/errors"
)

type contextKey int

const (
	// ContextKeyParams is the context key in the evaluation
	// context that stores the map[string]string
	ContextKeyParams contextKey = iota
)

// Utility function to get a set of map keys
func mapKeys(mapType interface{}) (set.Set, error) {
	rv := reflect.ValueOf(mapType)
	if rv.Kind() != reflect.Map {
		return nil, errors.Errorf("Value type is not a map: %T", mapType)
	}
	t := rv.Type()
	if t.Key().Kind() != reflect.String {
		return nil, errors.Errorf("Map key type is not a string: %T", t)
	}
	var result []interface{}
	for _, kv := range rv.MapKeys() {
		result = append(result, kv.String())
	}
	return set.NewSetFromSlice(result), nil
}

// ResourceProvider is the interface that CloudFormationCondenser
// Resources must satisfy. They are responsible for annotating
// the target template with CloudFormation information
type ResourceProvider interface {
	Annotate(ctx context.Context, template *gof.Template) (context.Context, error)
}

////////////////////////////////////////////////////////////////////////////////
// staticResource
////////////////////////////////////////////////////////////////////////////////
type static struct {
	name  string
	cfRes gof.Resource
}

func (res *static) Annotate(ctx context.Context, template *gof.Template) (context.Context, error) {
	template.Resources[res.name] = res.cfRes
	return ctx, nil
}

// Static returns a ResourceProvider for a static resource name
func Static(name string,
	cfRes gof.Resource) ResourceProvider {
	return &static{
		name:  name,
		cfRes: cfRes,
	}
}

var _ ResourceProvider = &static{}

////////////////////////////////////////////////////////////////////////////////
// flatten
////////////////////////////////////////////////////////////////////////////////
type flatten struct {
	namePrefix string
	annotator  ResourceProvider
}

func (fm *flatten) Annotate(ctx context.Context, template *gof.Template) (context.Context, error) {
	generatorTemplate := gof.NewTemplate()

	annotatedCtx, annotationErr := fm.annotator.Annotate(ctx, generatorTemplate)
	if annotationErr != nil {
		return nil, errors.Wrapf(annotationErr, "Failed to call annotation")
	}
	for eachKey, eachProp := range generatorTemplate.Resources {
		keyName := fmt.Sprintf("%s%s", fm.namePrefix, eachKey)
		template.Resources[keyName] = eachProp
	}
	return annotatedCtx, nil
}

var _ ResourceProvider = &flatten{}

// Flatten takes a namePrefix and a generator and promotes the
// returned resources "up" one level so that they are at the normal
// level
func Flatten(namePrefix string, annotator ResourceProvider) ResourceProvider {
	return &flatten{
		namePrefix: namePrefix,
		annotator:  annotator,
	}
}

// ProviderFunc is a wrapper around free functions that satisfies the
// ResourceProvider interface
type ProviderFunc func(ctx context.Context, template *gof.Template) (context.Context, error)

// Annotate satisfies the ResourceProvider interface
func (pfunc ProviderFunc) Annotate(ctx context.Context, template *gof.Template) (context.Context, error) {
	return pfunc(ctx, template)
}

////////////////////////////////////////////////////////////////////////////////

// SafeMerge is a free function that merges src into dest, reporting
// back any conflicting merge operations
func SafeMerge(src *gof.Template, dest *gof.Template) []error {
	mergeErrors := make([]error, 0)
	// Get everything and check it for collisions
	/*
		Mappings:                 map[string]*Mapping{},
		Parameters:               map[string]*Parameter{},
		Resources:                map[string]*Resource{},
		Outputs:                  map[string]*Output{},
		Conditions:               map[string]interface{}{},
	*/
	// Mappings
	srcMappingKeys, srcMappingKeysErr := mapKeys(src.Mappings)
	destMappingKeys, destMappingKeysErr := mapKeys(dest.Mappings)
	if srcMappingKeysErr != nil || destMappingKeysErr != nil {
		if srcMappingKeysErr != nil {
			mergeErrors = append(mergeErrors,
				srcMappingKeysErr)
		}
		if destMappingKeysErr != nil {
			mergeErrors = append(mergeErrors,
				destMappingKeysErr)
		}
	} else {
		collidingKeys := destMappingKeys.Intersect(srcMappingKeys)
		if collidingKeys.Cardinality() > 0 {
			mergeErrors = append(mergeErrors,
				errors.Errorf("Duplicate template.Mappings keynames detected: %s",
					collidingKeys.String()))
		} else {
			for eachKey, eachMapping := range src.Mappings {
				dest.Mappings[eachKey] = eachMapping
			}
		}
	}

	// Parameters
	srcParameterKeys, srcParameterKeysErr := mapKeys(src.Parameters)
	destParameterKeys, destParameterKeysErr := mapKeys(dest.Parameters)
	if srcParameterKeysErr != nil || destParameterKeysErr != nil {
		if srcParameterKeysErr != nil {
			mergeErrors = append(mergeErrors,
				srcParameterKeysErr)
		}
		if destParameterKeysErr != nil {
			mergeErrors = append(mergeErrors,
				destParameterKeysErr)
		}
	} else {
		collidingKeys := destParameterKeys.Intersect(srcParameterKeys)
		if collidingKeys.Cardinality() > 0 {
			mergeErrors = append(mergeErrors,
				errors.Errorf("Duplicate template.Parameters keynames detected: %s",
					collidingKeys.String()))
		} else {
			for eachKey, eachParam := range src.Parameters {
				dest.Parameters[eachKey] = eachParam
			}
		}
	}
	// Resources
	srcResourceKeys, srcResourceKeysErr := mapKeys(src.Resources)
	destResourceKeys, destResourceKeysErr := mapKeys(dest.Resources)
	if srcResourceKeysErr != nil || destResourceKeysErr != nil {
		if srcResourceKeysErr != nil {
			mergeErrors = append(mergeErrors,
				srcResourceKeysErr)
		}
		if destResourceKeysErr != nil {
			mergeErrors = append(mergeErrors,
				destResourceKeysErr)
		}
	} else {
		collidingKeys := destResourceKeys.Intersect(srcResourceKeys)
		if collidingKeys.Cardinality() > 0 {
			mergeErrors = append(mergeErrors,
				errors.Errorf("Duplicate template.Resources keynames detected: %s",
					collidingKeys.String()))
		} else {
			for eachKey, eachResource := range src.Resources {
				dest.Resources[eachKey] = eachResource
			}
		}
	}
	// Outputs
	srcOutputKeys, srcOutputKeysErr := mapKeys(src.Outputs)
	destOutputKeys, destOutputKeysErr := mapKeys(dest.Outputs)
	if srcOutputKeysErr != nil || destOutputKeysErr != nil {
		if srcOutputKeysErr != nil {
			mergeErrors = append(mergeErrors,
				srcOutputKeysErr)
		}
		if destOutputKeysErr != nil {
			mergeErrors = append(mergeErrors,
				destOutputKeysErr)
		}
	} else {
		collidingKeys := destOutputKeys.Intersect(srcOutputKeys)
		if collidingKeys.Cardinality() > 0 {
			mergeErrors = append(mergeErrors,
				errors.Errorf("Duplicate template.Outputs keynames detected: %s",
					collidingKeys.String()))
		} else {
			for eachKey, eachOutput := range src.Outputs {
				dest.Outputs[eachKey] = eachOutput
			}
		}
	}

	// Conditions
	srcConditionKeys, srcConditionKeysErr := mapKeys(src.Conditions)
	destConditionKeys, destConditionKeysErr := mapKeys(dest.Conditions)
	if srcConditionKeysErr != nil || destConditionKeysErr != nil {
		if srcConditionKeysErr != nil {
			mergeErrors = append(mergeErrors,
				srcConditionKeysErr)
		}
		if destConditionKeysErr != nil {
			mergeErrors = append(mergeErrors,
				destConditionKeysErr)
		}
	} else {
		collidingKeys := destConditionKeys.Intersect(srcConditionKeys)
		if collidingKeys.Cardinality() > 0 {
			mergeErrors = append(mergeErrors,
				errors.Errorf("Duplicate template.Conditions keynames detected: %s",
					collidingKeys.String()))
		} else {
			for eachKey, eachCondition := range src.Conditions {
				dest.Conditions[eachKey] = eachCondition
			}
		}
	}
	return mergeErrors
}

////////////////////////////////////////////////////////////////////////////////

// CloudFormationCondenser is the root template type
type CloudFormationCondenser struct {
	Description string
	Resources   []interface{}
}

// Evaluate executes all the registered resource providers
func (cfTemplate *CloudFormationCondenser) Evaluate(ctx context.Context) (*gof.Template, error) {
	evaluationErrors := make([]error, 0)

	// This is ultimately where everything will be merged
	targetTemplate := gof.NewTemplate()

	// Run through them...
	var annotateErr error
	for eachIndex, eachResource := range cfTemplate.Resources {
		annotationTemplate := gof.NewTemplate()

		switch typedValue := eachResource.(type) {
		case gof.Resource:
			cleanTypeName := strings.Replace(typedValue.AWSCloudFormationType(), ":", "", -1)
			cfName := fmt.Sprintf("CloudFormer%s%d", cleanTypeName, eachIndex)
			annotationTemplate.Resources[cfName] = typedValue
		case ResourceProvider:
			ctx, annotateErr = typedValue.Annotate(ctx, annotationTemplate)
		default:
			annotateErr = errors.Errorf("Unsupported Resource type: %#v", typedValue)
		}

		if annotateErr != nil {
			evaluationErrors = append(evaluationErrors, annotateErr)
		} else {
			safeMergeErrors := SafeMerge(annotationTemplate, targetTemplate)
			for _, eachError := range safeMergeErrors {
				indexErr := errors.Wrapf(eachError, "Resource[%d]",
					eachIndex)
				evaluationErrors = append(evaluationErrors, indexErr)
			}
		}
	}
	var evaluationError error
	if len(evaluationErrors) != 0 {
		evaluationError = errors.Errorf("Evaluation errors: %#v", evaluationErrors)
	}
	return targetTemplate, evaluationError
}
