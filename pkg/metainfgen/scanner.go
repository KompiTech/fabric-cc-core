package metainfgen

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/KompiTech/rmap"
)

// Scanner reads all YAML files in directory and produces Schema object for each schema encountered
// it does not handle duplicates
type Scanner struct {
	Path string
}

func NewScanner(path string) ([]Schema, error) {
	scanner := &Scanner{
		Path: path,
	}

	return scanner.scan()
}

func (s *Scanner) scan() ([]Schema, error) {
	pwd, err := os.Getwd()
	log.Print(pwd)

	files, err := ioutil.ReadDir(s.Path)
	if err != nil {
		return nil, err
	}

	schemas := []Schema{}

	for _, info := range files {
		if !strings.HasSuffix(info.Name(), ".yaml") {
			continue // not a YAML
		}

		data, err := rmap.NewFromYAMLFile(filepath.Join(s.Path, info.Name()))
		if err != nil {
			return nil, err
		}

		// attempt schema parse
		sch, err := NewSchema(strings.TrimSuffix(info.Name(), ".yaml"), data)
		if err != nil {
			return nil, err
		}

		schemas = append(schemas, sch)
	}

	return schemas, nil
}
