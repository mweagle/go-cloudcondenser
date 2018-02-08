package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"testing"

	gocc "github.com/mweagle/go-cloudcondenser"
	gocf "github.com/mweagle/go-cloudformation"
	"github.com/pkg/errors"
)

var condenserTemplates = map[string]gocc.CloudFormationCondenser{
	"simple":             simple,
	"simple-conditional": simpleConditional,
	"multi":              multiTemplate,
	"free":               freeTemplate,
}

func isJSONEqual(t *testing.T, expectedJSON []byte, actualJSON []byte) bool {
	var expected map[string]interface{}
	expectedUnmarshalErr := json.Unmarshal(expectedJSON, &expected)
	if expectedUnmarshalErr != nil {
		t.Fatalf("Failed to unmarshal expected JSON : %s", expectedUnmarshalErr)
	}

	var actual map[string]interface{}
	actualUnmarshalErr := json.Unmarshal(actualJSON, &actual)
	if actualUnmarshalErr != nil {
		t.Fatalf("Failed to unmarshal actual JSON : %s", actualUnmarshalErr)
	}
	return reflect.DeepEqual(expected, actual)
}
func TestCloudCondensor(t *testing.T) {
	for eachExpectedFilename, eachTemplate := range condenserTemplates {
		ctx := context.Background()
		expectedOutputFile := fmt.Sprintf("%s.json", eachExpectedFilename)
		readFile, readFileErr := ioutil.ReadFile(expectedOutputFile)
		if readFileErr != nil {
			t.Fatalf("Failed to read file: %s", expectedOutputFile)
		}
		templateOutput, templateOutputErr := eachTemplate.Evaluate(ctx)
		if templateOutputErr != nil {
			t.Fatalf("Failed to evaluate template :%s", templateOutputErr)
		}
		templateJSON, templateJSONErr := json.Marshal(templateOutput)
		if templateJSONErr != nil {
			t.Fatalf("Failed to marshal JSON : %s", templateJSONErr)
		}
		if !isJSONEqual(t, readFile, templateJSON) {
			t.Fatalf("Failed to verify output for test: %s\nGENERATED:%#v\nEXPECTED: %#v",
				eachExpectedFilename,
				string(readFile),
				string(templateJSON))
		}
	}
	t.Logf("Verified %d templates", len(condenserTemplates))
}

type testContextKey int

const (
	// contextKeyTest is the keyname to verify context handling
	contextKeyTest testContextKey = iota
)

// Verify that the context properly passes information
func provider1(ctx context.Context, template *gocf.Template) (context.Context, error) {
	ctx = context.WithValue(ctx, contextKeyTest, "test")
	return ctx, nil
}

func provider2(ctx context.Context, template *gocf.Template) (context.Context, error) {
	keyValue, keyValueOk := ctx.Value(contextKeyTest).(string)
	if !keyValueOk || keyValue != "test" {
		return nil, errors.Errorf("Failed to verify context values")
	}
	return ctx, nil
}

var contextTemplate = gocc.CloudFormationCondenser{
	Description: "My Stack",
	Resources: []interface{}{
		gocc.ProviderFunc(provider1),
		gocc.ProviderFunc(provider2),
	},
}

func TestCloudCondensorContext(t *testing.T) {
	_, evalErr := contextTemplate.Evaluate(context.Background())
	if evalErr != nil {
		t.Fatalf("Failed to ensure stateful context properly preserved")
	}
}
