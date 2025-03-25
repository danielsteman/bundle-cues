package main

import (
	"fmt"
	"io"
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

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func getIncludes() ([]string, error) {
	file, err := os.Open("databricks.yml")
	check(err)
	defer file.Close()
	data, err := io.ReadAll(file)
	check(err)

	spec := BundleSpec{}
	err = yaml.Unmarshal(data, &spec)
	check(err)

	include_patterns := spec.Include

	var include_paths []string
	var seen map[string]bool

	for _, path := range include_patterns {
		yml_extension := "yml"
		if len(path) > len(yml_extension) {
			extension := path[len(path)-3:]
			if extension != "yml" {
				log.Fatal("Only yml files can be included")
			}

			glob_paths, err := filepath.Glob(path)
			check(err)

			for _, path := range glob_paths {
				if seen[path] == false {
					include_paths = append(include_paths, path)
				}
			}

		} else {
			log.Fatalf("Only yml files can be included, path is too short: %s", path)
		}
	}
	return include_paths, nil
}

func main() {
	includes, err := getIncludes()
	check(err)
	fmt.Printf("%v\n", includes)

	ctx := cuecontext.New()
	insts := load.Instances([]string{"."}, nil)
	ctx.BuildInstance(insts[0])

}
