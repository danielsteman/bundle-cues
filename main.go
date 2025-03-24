package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"gopkg.in/yaml.v3"
)

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func resolveIncludes() {
	file, err := os.Open("databricks.yml")
	check(err)
	defer file.Close()
	data, err := io.ReadAll(file)
	check(err)

	m := make(map[string]interface{})
	err = yaml.Unmarshal(data, &m)
	check(err)
}

func main() {
    resolveIncludes() 
        
	ctx := cuecontext.New()
	insts := load.Instances([]string{"."}, nil)
	v := ctx.BuildInstance(insts[0])
	fmt.Printf("%v\n", v)
}
