package initgen

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

func GetNameSkip(path string, info os.FileInfo) (string, bool) {
	if !strings.HasSuffix(info.Name(), ".yaml") {
		log.Printf("%s is not YAML, skipped\n", path)
		return "", true
	}

	name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))

	if strings.HasPrefix(name, "mock") {
		log.Printf("%s is MOCK, skipped\n", path)
		return "", true
	}

	log.Printf("Processing path: %s, name: %s\n", path, name)
	return name, false
}
