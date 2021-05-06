package main

import (
	"flag"
	"log"

	. "github.com/KompiTech/fabric-cc-core/v2/pkg/metainfgen"
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

	schemas, err := NewScanner(*registryDir)
	if err != nil {
		log.Panic(err.Error())
	}

	wr, err := NewWriter(*outputDir)
	if err != nil {
		log.Panic(err.Error())
	}


	err = wr.WriteIndexFiles(schemas)
	if err != nil {
		log.Panic(err.Error())
	}
}
