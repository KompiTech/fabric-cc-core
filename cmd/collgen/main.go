package main

import (
	"flag"
	"log"
	"os"

	"github.com/KompiTech/fabric-cc-core/v2/internal/collgen"
	"github.com/KompiTech/rmap"
)

func main() {
	registryDir := flag.String("registryDir", "", "directory containing registry definitions for cc-core based chaincode")
	templateFile := flag.String("templateFile", "", "optional YAML file which contents will be copied to each produced JSON")

	flag.Parse()

	if *registryDir == "" {
		log.Fatal("registryDir is mandatory argument")
	}

	var template rmap.Rmap
	if *templateFile == "" {
		template = rmap.NewEmpty()
	} else {
		template = rmap.MustNewFromYAMLFile(*templateFile)
	}

	collGen := collgen.New(*registryDir, template, os.Stdout)
	collGen.Visit()
}
