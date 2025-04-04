package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type BundleSpec struct {
	Include []string `yaml:"include"`
}

func getIncludes() ([]string, error) {
	data, err := os.ReadFile("databricks.yml")
	if err != nil {
		return nil, fmt.Errorf("failed to read databricks.yml: %w", err)
	}
	spec := BundleSpec{}
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("failed to unmarshal databricks.yml: %w", err)
	}
	var includePaths []string
	seen := make(map[string]bool)
	for _, pattern := range spec.Include {
		if len(pattern) < 4 {
			return nil, fmt.Errorf("include pattern is too short: %q", pattern)
		}
		if filepath.Ext(pattern) != ".yml" {
			return nil, fmt.Errorf("only .yml files can be included: %q", pattern)
		}
		globPaths, err := filepath.Glob(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid glob pattern %q: %w", pattern, err)
		}
		for _, p := range globPaths {
			if !seen[p] {
				seen[p] = true
				includePaths = append(includePaths, p)
			}
		}
	}
	return includePaths, nil
}

func nodesEqual(l, r *yaml.Node) bool {
	if l.Kind == yaml.ScalarNode && r.Kind == yaml.ScalarNode {
		return l.Value == r.Value
	}
	panic("equals on non-scalars not implemented!")
}

func recursiveMerge(from, into *yaml.Node) error {
	if from.Kind != into.Kind {
		return fmt.Errorf("cannot merge nodes of different kinds (%v vs %v)", from.Kind, into.Kind)
	}

	switch from.Kind {
	case yaml.MappingNode:
		for i := 0; i < len(from.Content); i += 2 {
			fromKey := from.Content[i]
			fromVal := from.Content[i+1]

			// Try to find matching key in 'into'
			var found bool
			for j := 0; j < len(into.Content); j += 2 {
				intoKey := into.Content[j]
				intoVal := into.Content[j+1]

				if nodesEqual(fromKey, intoKey) {
					found = true
					// If both values are mappings, recursively merge
					if fromVal.Kind == yaml.MappingNode && intoVal.Kind == yaml.MappingNode {
						if err := recursiveMerge(fromVal, intoVal); err != nil {
							return fmt.Errorf("error merging map for key %q: %w", fromKey.Value, err)
						}
					} else if fromVal.Kind == yaml.SequenceNode && intoVal.Kind == yaml.SequenceNode {
						// Optionally deduplicate or just append
						intoVal.Content = append(intoVal.Content, fromVal.Content...)
					} else {
						// Replace the value (scalar or different kinds)
						into.Content[j+1] = fromVal
					}
					break
				}
			}

			if !found {
				// Key doesn't exist in 'into', append it
				into.Content = append(into.Content, fromKey, fromVal)
			}
		}

	case yaml.SequenceNode:
		into.Content = append(into.Content, from.Content...)

	case yaml.DocumentNode:
		if len(from.Content) == 0 || len(into.Content) == 0 {
			return errors.New("unexpected empty content in document node")
		}
		return recursiveMerge(from.Content[0], into.Content[0])

	default:
		// For ScalarNode or other unsupported kinds
		return fmt.Errorf("cannot merge node kind %v", from.Kind)
	}

	return nil
}

func unifyConfigs(includePaths []string) (*yaml.Node, error) {
	var master yaml.Node
	baseData, err := os.ReadFile("databricks.yml")
	if err != nil {
		return nil, fmt.Errorf("failed to read databricks.yml: %w", err)
	}
	if err := yaml.Unmarshal(baseData, &master); err != nil {
		return nil, fmt.Errorf("failed to unmarshal databricks.yml: %w", err)
	}
	for _, path := range includePaths {
		fmt.Printf("Merging file: %s\n", path)
		var override yaml.Node
		bs, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read include file %q: %w", path, err)
		}
		if err := yaml.Unmarshal(bs, &override); err != nil {
			return nil, fmt.Errorf("failed to unmarshal include file %q: %w", path, err)
		}
		if err := recursiveMerge(&override, &master); err != nil {
			return nil, fmt.Errorf("failed to merge file %q: %w", path, err)
		}
	}
	return &master, nil
}

func validate(schemaPath string) {
	includes, err := getIncludes()
	if err != nil {
		log.Fatal(err)
	}

	finalConfig, err := unifyConfigs(includes)
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < len(finalConfig.Content[0].Content); i += 2 {
		key := finalConfig.Content[0].Content[i]
		fmt.Printf("Top-level key: %s\n", key.Value)
	}

	mergedYAML, err := yaml.Marshal(finalConfig)
	if err != nil {
		log.Fatalf("Failed to marshal final config: %v", err)
	}
	fmt.Println(string(mergedYAML))

	intermediate := make(map[string]interface{})
	if err := yaml.Unmarshal(mergedYAML, &intermediate); err != nil {
		log.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	jsonBytes, err := json.Marshal(intermediate)
	if err != nil {
		log.Fatalf("Failed to convert YAML to JSON: %v", err)
	}

	ctx := cuecontext.New()

	schemaInsts := load.Instances([]string{schemaPath}, nil)
	if len(schemaInsts) == 0 || schemaInsts[0] == nil {
		log.Fatalf("No CUE instance found for schema %q", schemaPath)
	}
	schemaVal := ctx.BuildInstance(schemaInsts[0])
	if err := schemaVal.Err(); err != nil {
		log.Fatalf("CUE build error in schema: %v", err)
	}

	dataVal := ctx.CompileBytes(jsonBytes)
	if err := dataVal.Err(); err != nil {
		log.Fatalf("CUE compile error in merged data: %v", err)
	}

	result := schemaVal.Unify(dataVal)
	if err := result.Err(); err != nil {
		log.Fatalf("Validation failed: %v", err)
	}

	fmt.Println("Validation successful!")
}

var validateCmd = &cobra.Command{
	Use:   "validate [schema]",
	Short: "Validate your databricks asset bundle config against a CUE schema",
	Long: `Validates the final merged configuration (from "databricks.yml" plus any 
".yml" includes) against a user-provided CUE schema file. The schema 
can contain custom rules and constraints for your bundle.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		schemaPath := args[0]

		validate(schemaPath)
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}
