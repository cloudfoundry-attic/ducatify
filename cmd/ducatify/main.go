package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/cloudfoundry-incubator/ducatify"

	"gopkg.in/yaml.v2"
)

func main() {
	var diegoManifestPath string
	var cfCredsPath string

	flag.StringVar(&diegoManifestPath, "diego", "", "path to vanilla diego manifest")
	flag.StringVar(&cfCredsPath, "cfCreds", "", "path to cf creds config")
	flag.Parse()

	if diegoManifestPath == "" {
		log.Fatalf("missing required flag 'diego'")
	}

	if cfCredsPath == "" {
		log.Fatalf("missing required flag 'cfCreds'")
	}

	vanillaBytes, err := ioutil.ReadFile(diegoManifestPath)
	if err != nil {
		log.Fatalf("reading diego manifest: %s", err)
	}

	cfCredBytes, err := ioutil.ReadFile(cfCredsPath)
	if err != nil {
		log.Fatalf("reading cf creds config: %s", err)
	}

	transformedBytes, err := transformBytes(vanillaBytes, cfCredBytes)
	if err != nil {
		log.Fatalf("%s", err)
	}

	os.Stdout.Write(transformedBytes)
}

func transformBytes(vanillaBytes, cfCredBytes []byte) ([]byte, error) {
	var manifest map[interface{}]interface{}
	err := yaml.Unmarshal(vanillaBytes, &manifest)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling yaml: %s", err)
	}

	var acceptanceJobConfig map[interface{}]interface{}
	err = yaml.Unmarshal(cfCredBytes, &acceptanceJobConfig)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling yaml: %s", err)
	}

	transformer := ducatify.New()
	err = transformer.Transform(manifest, acceptanceJobConfig)
	if err != nil {
		return nil, fmt.Errorf("transforming: %s", err)
	}

	transformedBytes, err := yaml.Marshal(manifest)
	if err != nil {
		return nil, fmt.Errorf("re-marshalling yaml: %s", err)
	}

	return transformedBytes, nil
}
