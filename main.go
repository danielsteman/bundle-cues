package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
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
			found := false
			for j := 0; j < len(into.Content); j += 2 {
				intoKey := into.Content[j]
				intoVal := into.Content[j+1]
				if nodesEqual(fromKey, intoKey) {
					found = true
					if err := recursiveMerge(fromVal, intoVal); err != nil {
						return fmt.Errorf("error merging key %q: %w", fromKey.Value, err)
					}
					break
				}
			}
			if !found {
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
		return errors.New("can only merge mapping, sequence, or document nodes")
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

	mergedYAML, err := yaml.Marshal(finalConfig)
	if err != nil {
		log.Fatalf("Failed to marshal final config: %v", err)
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

	dataVal := ctx.CompileBytes(mergedYAML)
	if err := dataVal.Err(); err != nil {
		log.Fatalf("CUE compile error in merged data: %v", err)
	}

	result := schemaVal.Unify(dataVal)
	if err := result.Err(); err != nil {
		log.Fatalf("Validation failed: %v", err)
	}

	fmt.Println("Validation successful!")
}

func main() {
	includes, err := getIncludes()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found includes: %v\n\n", includes)
	finalConfig, err := unifyConfigs(includes)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Final merged config as *yaml.Node:\n%+v\n\n", finalConfig)
	ctx := cuecontext.New()
	insts := load.Instances([]string{"."}, nil)
	if len(insts) > 0 && insts[0] != nil {
		val := ctx.BuildInstance(insts[0])
		if val.Err() != nil {
			log.Printf("CUE validation error: %v", val.Err())
		}
	}
}
