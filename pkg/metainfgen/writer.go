package metainfgen

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Writer takes Scanner output and creates resulting META-INF structure to be packaged/used by testing CouchDB
type Writer struct {
	stateIndexPath string
	privateIndexPath string
}

func NewWriter(outputPath string) (*Writer, error) {
	indexRootPath := filepath.Join(outputPath, "statedb", "couchdb") // common dir

	wrt := &Writer{
		stateIndexPath: filepath.Join(indexRootPath, "indexes"), // dir for state indexes
		privateIndexPath: filepath.Join(indexRootPath, "collections"), // dir for private data indexes (subdir is created for each collection name)
	}

	return wrt, nil
}

// make sure directories in outputPath are created and writable
func (w Writer) ensureMetaInfStructure() error {
	if err := os.MkdirAll(w.stateIndexPath, 0755); err != nil {
		return err
	}

	if err := os.MkdirAll(w.privateIndexPath, 0755); err != nil {
		return err
	}

	return nil
}

func (w Writer) writeDefaultIndexes(path string) error {
	// docType index
	if err := ioutil.WriteFile(filepath.Join(path, "doctype.json"), getIndexMap([]string{docType}, docType).Bytes(), 0644); err != nil {
		return err
	}

	// docType, uuid index
	if err := ioutil.WriteFile(filepath.Join(path, "uuid.json"), getIndexMap([]string{docType, "uuid"}, "uuid").Bytes(), 0644); err != nil {
		return err
	}

	return nil
}

func (w *Writer) WriteIndexFiles(schemas []Schema) error {
	if err := w.ensureMetaInfStructure(); err != nil {
		return err
	}

	if err := w.writeDefaultIndexes(w.stateIndexPath); err != nil {
		return err
	}

	// go through all schemas
	for _, sch := range schemas {
		if sch.Destination != stateDestination {
			// write default indexes for each collection
			privateDirPath := filepath.Join(w.privateIndexPath, strings.ToUpper(sch.Name), "indexes")
			if err := os.MkdirAll(privateDirPath, 0755); err != nil {
				return err
			}
			if err := w.writeDefaultIndexes(privateDirPath); err != nil {
				return err
			}
		}

		// go through all indexes in schema
		for _, idx := range sch.Indexes {
			var writePath string

			fileName, err := idx.GetString("name")
			if err != nil {
				return nil
			}
			fileName += ".json"

			if sch.Destination == stateDestination {
				// state index -> write to stateIndexPath/<name>.yaml
				writePath = filepath.Join(w.stateIndexPath, fileName)
			} else {
				// private data destination -> write to privateIndexPath/<collection_name(uppercase)>/indexes/<name>.yaml
				// create directory if not existing already
				privateDirPath := filepath.Join(w.privateIndexPath, strings.ToUpper(sch.Name), "indexes")
				if err := os.MkdirAll(privateDirPath, 0755); err != nil {
					return err
				}

				writePath = filepath.Join(privateDirPath, fileName)
			}

			if _, err := os.Stat(writePath); !os.IsNotExist(err) {
				// if index file already exists, skip
				log.Printf("index %s already exists, skipped", writePath)
				continue
			}

			if err := ioutil.WriteFile(writePath, idx.Bytes(), 0644); err != nil {
				return err
			}

			log.Printf("Wrote index file: %s", writePath)
		}
	}

	return nil
}

