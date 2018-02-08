package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	templates "github.com/mweagle/go-cloudcondenser/cmd/templates"
)

type contextKey int

const (
	// ContextKeyParams are the parameters made available to the
	// evaluation context via the command line args
	ContextKeyParams contextKey = iota
)

// Custom flags sink to translate NAME=VALUE pairs into values
type flagsMap struct {
	args map[string]string
}

func (i *flagsMap) String() string {
	return fmt.Sprintf("%#v", i)
}

func (i *flagsMap) Set(value string) error {
	parts := strings.Split(value, "=")
	if i.args == nil {
		i.args = make(map[string]string, 0)
	}
	if len(parts) > 1 {
		i.args[parts[0]] = parts[1]
	} else {
		i.args[parts[0]] = parts[0]
	}
	return nil
}

var myFlags flagsMap

func main() {
	// Grab all the command line options and stuff them
	// into a map. We'll put that into the context in case
	// ResourceProviders need to make conditional switches
	flag.Var(&myFlags, "param", "Set a NAME=VALUE value pair that will be published into the Evaluation(context) map")
	flag.Parse()

	// Build it...
	ctx := context.Background()
	ctx = context.WithValue(ctx, ContextKeyParams, myFlags.args)
	outputTemplate, outputErr := templates.DefaultTemplate.Evaluate(ctx)
	if outputErr != nil {
		fmt.Printf("Failed to execute template: %s\n", outputErr.Error())
		os.Exit(1)
	}
	// Output this
	jsonOut, _ := json.MarshalIndent(outputTemplate, "", " ")
	fmt.Println(string(jsonOut))
}
