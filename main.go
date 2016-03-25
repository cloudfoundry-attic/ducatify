package main

import (
	"flag"
	"fmt"
	"io/ioutil"
)

func main() {
	var diegoManifestPath string
	flag.StringVar(&diegoManifestPath, "diego", "", "path to vanilla diego manifest")
	flag.Parse()
	if diegoManifestPath == "" {
		panic("missing required flag")
	}
	diegoBytes, err := ioutil.ReadFile(diegoManifestPath)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(diegoBytes))
}
