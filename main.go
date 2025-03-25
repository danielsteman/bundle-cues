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

	includes := spec.Include
	for _, path := range includes {
		yml_extension := "yml"
		if len(path) > len(yml_extension) {
			extension := path[len(path)-3:]
			if extension != "yml" {
				log.Fatal("Only yml files can be included")
			}
		} else {
			log.Fatalf("Only yml files can be included, path is too short: %s", path)
		}
	}
	return spec.Include, nil
}

func main() {
	includes, err := getIncludes()
	check(err)
	fmt.Printf("%v\n", includes)

	ctx := cuecontext.New()
	insts := load.Instances([]string{"."}, nil)
	v := ctx.BuildInstance(insts[0])
	fmt.Printf("%v\n", v)

	paths, err := filepath.Glob("*.yml")
	check(err)
	fmt.Printf("%v\n", paths)
}
