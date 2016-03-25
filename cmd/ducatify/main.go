package main

import (
	"flag"
	"io/ioutil"
	"os"

	"github.com/cloudfoundry-incubator/ducatify"

	"gopkg.in/yaml.v2"
)

func main() {
	var diegoManifestPath string
	flag.StringVar(&diegoManifestPath, "diego", "", "path to vanilla diego manifest")
	flag.Parse()
	if diegoManifestPath == "" {
		panic("missing required flag")
	}
	vanillaBytes, err := ioutil.ReadFile(diegoManifestPath)
	if err != nil {
		panic(err)
	}

	transformedBytes, err := transformBytes(vanillaBytes)
	if err != nil {
		panic(err)
	}

	os.Stdout.Write(transformedBytes)
}

func transformBytes(vanillaBytes []byte) ([]byte, error) {
	var manifest map[string]interface{}
	err := yaml.Unmarshal(vanillaBytes, &manifest)
	if err != nil {
		return nil, err
	}

	transformer := &ducatify.Transformer{}
	err = transformer.Transform(manifest)
	if err != nil {
		return nil, err
	}

	transformedBytes, err := yaml.Marshal(manifest)
	if err != nil {
		return nil, err
	}

	return transformedBytes, nil
}
