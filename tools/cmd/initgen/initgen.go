package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/KompiTech/rmap"
)

type InitGen struct {
	registryDir          string
	singletonDir         string
	singletonBlacklist   rmap.Rmap
	discoveredRegistries rmap.Rmap
	discoveredSingletons rmap.Rmap
	encoder              *json.Encoder
}

func New(registryDir, singletonDir string, singletonBlacklist rmap.Rmap, output io.Writer) *InitGen {
	return &InitGen{
		registryDir:          registryDir,
		singletonDir:         singletonDir,
		singletonBlacklist:   singletonBlacklist,
		discoveredRegistries: rmap.NewEmpty(),
		discoveredSingletons: rmap.NewEmpty(),
		encoder:              json.NewEncoder(output),
	}
}

func (i *InitGen) Visit() {
	if err := filepath.Walk(i.registryDir, i.visitRegistry); err != nil {
		log.Fatalf("filepath.Walk() error: %s", err)
	}

	if err := filepath.Walk(i.singletonDir, i.visitSingleton); err != nil {
		log.Fatalf("filepath.Walk() error: %s", err)
	}

	err := i.encoder.Encode(map[string]interface{}{
		"registries": i.discoveredRegistries.Mapa,
		"singletons": i.discoveredSingletons.Mapa,
	})
	if err != nil {
		log.Fatalf("i.encoder.Encode() error: %s", err)
	}
}

func (i *InitGen) visitRegistry(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	name, skip := GetNameSkip(path, info)
	if skip {
		return nil
	}

	if i.discoveredRegistries.Exists(name) {
		log.Fatalf("registry name: %s was already discovered", name)
	}

	i.discoveredRegistries.Mapa[name] = rmap.MustNewFromYAMLFile(path).Mapa

	return nil
}

func (i *InitGen) visitSingleton(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	name, skip := GetNameSkip(path, info)
	if skip {
		return nil
	}

	if i.singletonBlacklist.Exists(name) {
		log.Printf("%s is BLACKLISTED, skipped\n", name)
		return nil
	}

	if i.discoveredSingletons.Exists(name) {
		log.Fatalf("singleton name: %s was already discovered", name)
	}

	i.discoveredSingletons.Mapa[name] = rmap.MustNewFromYAMLFile(path).Mapa

	return nil
}
