package main

import (
	"flag"
	"log"
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

	metaInfGen := New(*registryDir, *outputDir)
	if err := metaInfGen.Generate(); err != nil {
		log.Fatal(err)
	}
}
