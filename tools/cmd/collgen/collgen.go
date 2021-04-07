package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/KompiTech/rmap"
)

type CollGen struct {
	registryDir string
	template    rmap.Rmap
	output      io.Writer
	outList     []rmap.Rmap
}

func New(registryDir string, template rmap.Rmap, output io.Writer) *CollGen {
	return &CollGen{
		registryDir: registryDir,
		template:    template,
		output:      output,
		outList:     []rmap.Rmap{},
	}
}

func (c *CollGen) Visit() {
	if err := filepath.Walk(c.registryDir, c.visit); err != nil {
		log.Fatalf("filepath.Walk() error: %s", err)
	}

	outBytes, err := json.Marshal(c.outList)
	if err != nil {
		log.Fatal(err)
	}

	written, err := c.output.Write(outBytes)
	if err != nil {
		log.Fatal(err)
	}

	if written != len(outBytes) {
		log.Fatal("written != len(outBytes)")
	}
}

func (c *CollGen) visit(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	if !strings.HasSuffix(info.Name(), ".yaml") {
		log.Printf("%s is not YAML, skipped\n", path)
		return nil
	}

	if strings.HasPrefix(info.Name(), "mock") {
		log.Printf("%s is MOCK, skipped\n", path)
		return nil
	}

	data := rmap.MustNewFromYAMLFile(path)

	if !data.Exists("destination") {
		log.Printf("registry file: %s does not have destination key!", path)
	}

	destination := strings.ToLower(data.MustGetJPtrString("/destination"))

	if !strings.HasPrefix(destination, "private_data") {
		return nil
	}

	collName := strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))

	log.Printf("Found private_data collection: %s\n", collName)

	c.appendToJSON(collName)

	return nil
}

func (c *CollGen) appendToJSON(collectionName string) {
	output := c.template.Copy()
	output.Mapa["name"] = strings.ToUpper(collectionName)

	c.outList = append(c.outList, output)
}
