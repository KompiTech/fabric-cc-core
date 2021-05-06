package main

import (
	"flag"
	"log"

	metainfgen2 "github.com/KompiTech/fabric-cc-core/v2/pkg/metainfgen"
)

func main() {
	registryDir := flag.String("registryDir", "", "directory containing registry definitions for cc-core based chaincode")
	outputDir := flag.String("outputDir", "", "directory where to create")

	flag.Parse()

	if *registryDir == "" {
		log.Fatal("registryDir is mandatory argument")
	}

	if *outputDir == "" {
		log.Fatal("outputDir is mandatory argument")
	}

	metaInfGen := metainfgen2.New(*registryDir, *outputDir)
	if err := metaInfGen.Generate(); err != nil {
		log.Fatal(err)
	}
}
