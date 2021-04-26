package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/KompiTech/fabric-cc-core/v2/internal/initgen"
	"github.com/KompiTech/rmap"
)

func main() {
	registryDir := flag.String("registryDir", "", "directory containing registry definitions for cc-core based chaincode")
	singletonDir := flag.String("singletonDir", "", "directory containing singleton definitions for cc-core based chaincode")
	singletonBlacklistStr := flag.String("singletonBlacklist", "", "optional comma-separated list of singletons to NOT include")

	flag.Parse()

	if *registryDir == "" {
		log.Fatal("registryDir is mandatory argument")
	}

	if *singletonDir == "" {
		log.Fatal("singletonDir is mandatory argument")
	}

	var singletonBlacklist rmap.Rmap
	if *singletonBlacklistStr == "" {
		singletonBlacklist = rmap.NewEmpty()
	} else {
		singletonBlacklist = rmap.NewFromStringSlice(strings.Split(strings.Replace(*singletonBlacklistStr, " ", "", -1), ","))
	}

	initGen := initgen.New(*registryDir, *singletonDir, singletonBlacklist, os.Stdout)
	initGen.Visit()
}
