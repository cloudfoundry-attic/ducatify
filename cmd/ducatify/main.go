package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/cloudfoundry-incubator/candiedyaml"
	"github.com/cloudfoundry-incubator/ducatify"
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

func getSystemDomain(cfCreds map[interface{}]interface{}) (string, error) {
	apiVal, exists := cfCreds["api"]
	if !exists {
		return "", fmt.Errorf("missing expected config in cfCreds: api")
	}
	api, ok := apiVal.(string)
	if !ok {
		return "", fmt.Errorf("api key not a string")
	}
	if !strings.HasPrefix(api, "api.") {
		return "", fmt.Errorf("unable to parse api key to extract system domain")
	}
	return strings.TrimPrefix(api, "api."), nil
}

func transformBytes(vanillaBytes, cfCredBytes []byte) ([]byte, error) {
	var manifest map[interface{}]interface{}
	err := candiedyaml.Unmarshal(vanillaBytes, &manifest)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling yaml: %s", err)
	}

	var cfCreds map[interface{}]interface{}
	err = candiedyaml.Unmarshal(cfCredBytes, &cfCreds)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling yaml: %s", err)
	}

	systemDomain, err := getSystemDomain(cfCreds)
	if err != nil {
		return nil, fmt.Errorf("getting system domain: %s", err)
	}

	transformer := ducatify.New()
	err = transformer.Transform(manifest, cfCreds, systemDomain)
	if err != nil {
		return nil, fmt.Errorf("transforming: %s", err)
	}

	transformedBytes, err := candiedyaml.Marshal(manifest)
	if err != nil {
		return nil, fmt.Errorf("re-marshalling yaml: %s", err)
	}

	return transformedBytes, nil
}
