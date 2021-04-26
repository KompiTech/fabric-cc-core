package metainfgen

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/KompiTech/rmap"
)

// handles orchestration of generating index json files in proper directories
type MetaInfGen struct {
	registryDir    string
	outputDir      string
	indexRootDir   string
	indexDir       string
	collectionsDir string
}

func New(registryDir, outputDir string) *MetaInfGen {
	indexRootDir := filepath.Join(outputDir, "statedb", "couchdb")
	indexDir := filepath.Join(indexRootDir, "indexes")           // dir for state indexes
	collectionsDir := filepath.Join(indexRootDir, "collections") // dir for subdir for each collection index

	return &MetaInfGen{
		registryDir:    registryDir,
		outputDir:      outputDir,
		indexRootDir:   indexRootDir,
		indexDir:       indexDir,
		collectionsDir: collectionsDir,
	}
}

func (m *MetaInfGen) Generate() error {
	// create necessary directory structure if not exists
	if err := os.MkdirAll(m.indexDir, 0755); err != nil {
		return err
	}

	if err := os.MkdirAll(m.collectionsDir, 0755); err != nil {
		return err
	}

	// recursively scan everything in registryDir and create relevant jsons
	if err := filepath.Walk(m.registryDir, m.visitRegistry); err != nil {
		return err
	}

	// these indexes are created every time regardless of configuration

	// docType index
	if err := m.ConditionalWrite(IndexFile{
		data:     getIndexMap([]string{docType}, docType),
		fileName: "doctype.json",
	}, ""); err != nil {
		return err
	}

	// docType, entity index
	if err := m.ConditionalWrite(IndexFile{
		data:     getIndexMap([]string{docType, "entity"}, "entity"),
		fileName: "entity.json",
	}, ""); err != nil {
		return err
	}

	return nil
}

// this func gets called for every file in registryDir
func (m *MetaInfGen) visitRegistry(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	// get fileName of registry file and if it should be skipped
	name, skip := GetNameSkip(path, info)
	if skip {
		return nil
	}

	registryMap, err := rmap.NewFromYAMLFile(path)
	if err != nil {
		return nil
	}

	ss := NewSchemaScan()

	// get all index files and their data from schema
	indexFiles, err := ss.GetIndexesFromSchema(registryMap, name)
	if err != nil {
		return nil
	}

	// write all json index produced from schema if they dont exist yet
	for _, indexFile := range indexFiles {
		if err := m.ConditionalWrite(indexFile, ss.collectionName); err != nil {
			return err
		}
	}

	return nil
}

// multiple schemas can define the same single or multi index
// create json only if it doesn't exist
func (m *MetaInfGen) ConditionalWrite(indexFile IndexFile, collectionName string) error {
	// at this point, index for this field is required
	// index filename is by convention same the same as property fileName
	var indexPath string

	if collectionName == "" {
		// all state indexFiles are in one directory
		indexPath = filepath.Join(m.indexDir, indexFile.fileName)
	} else {
		// one directory for each index in private data called after collection fileName, ensure it exists
		collDirPath := filepath.Join(m.collectionsDir, strings.ToUpper(collectionName), "indexes")

		if err := os.MkdirAll(collDirPath, 0755); err != nil {
			return err
		}

		// put index to collection directory
		indexPath = filepath.Join(collDirPath, indexFile.fileName)
	}

	if _, err := os.Stat(indexPath); !os.IsNotExist(err) {
		// if index file already exists, end here
		log.Printf("index %s already exists, skipped", indexPath)
		return nil
	}

	// index json needs to be created at this point
	if err := ioutil.WriteFile(indexPath, indexFile.data.Bytes(), 0644); err != nil {
		return err
	}

	log.Printf("created index %s", indexPath)
	return nil
}
