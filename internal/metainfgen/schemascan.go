package metainfgen

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/KompiTech/rmap"
)

const (
	docType              = "docType"
	indexMagic           = "_INDEX_"
	multiIndexMagicStart = "_MULTI:" // multiindexes must follow format _MULTI:<order>:<index_name>_. order controls only sorting of fields in multiindex and cannot be repeated
	multiIndexMagicEnd   = "_"
	state                = "state"
	description          = "description"
	typ                  = "type"
	schemaRoot           = "/schema/properties"
	properties           = "properties"
)

// scans single schema and produces IndexFiles to be written by MetaInfGen
type SchemaScan struct {
	indexFiles     []IndexFile
	multiIndexes   map[string]map[string]string // name -> order -> field name
	collectionName string
}

func NewSchemaScan() *SchemaScan {
	return &SchemaScan{
		indexFiles:   make([]IndexFile, 0),
		multiIndexes: map[string]map[string]string{},
	}
}

func (s *SchemaScan) GetIndexesFromSchema(registryMap rmap.Rmap, name string) ([]IndexFile, error) {
	if len(s.indexFiles) != 0 || len(s.multiIndexes) != 0 {
		return nil, errors.New("do not reuse this object")
	}

	destination, err := registryMap.GetString("destination")
	if err != nil {
		return nil, err
	}

	if destination == state {
		s.collectionName = ""
	} else {
		s.collectionName = name
	}

	// get schema root and scan for indexes
	properties, err := registryMap.GetJPtrRmap(schemaRoot)
	if err != nil {
		return nil, err
	}

	// iterate in sorted order, so it is deterministic (tests)
	for _, k := range sortedKeys(properties) {
		v := rmap.MustNewFromInterface(properties.Mapa[k])
		if err := s.handleProperty(k, v); err != nil {
			return nil, err
		}
	}

	// single field indexes are already produced to indexFiles
	// create multi indexes
	if err := s.collectMultiIndexes(); err != nil {
		return nil, err
	}

	return s.indexFiles, nil
}

func (s *SchemaScan) handleProperty(propName string, property rmap.Rmap) error {
	// handle recursion first: if type is object, start recursion for all of its properties
	if property.Exists(typ) {
		typ, err := property.GetString(typ)
		if err != nil {
			return err
		}

		if typ == "object" {
			if property.Exists(properties) {
				props, err := property.GetRmap(properties)
				if err != nil {
					return err
				}

				// iterate in sorted order, so it is deterministic (tests)
				for _, k := range sortedKeys(props) {
					v := rmap.MustNewFromInterface(props.Mapa[k])
					if err := s.handleProperty(fmt.Sprintf("%s.%s", propName, k), v); err != nil {
						return err
					}
				}
			}
		}
	}

	// current obj must contain descr, if it is to define any index
	if !property.Exists(description) {
		return nil
	}

	descr, err := property.GetString(description)
	if err != nil {
		return nil
	}

	s.checkForSingleIndex(descr, propName)

	s.checkForMultiIndex(descr, propName)

	return nil
}

func (s *SchemaScan) checkForSingleIndex(descr, propName string) {
	if !strings.Contains(descr, indexMagic) {
		return
	}

	s.indexFiles = append(s.indexFiles, IndexFile{
		data:     getIndexMap([]string{docType, propName}, propName), // get couchDB formatted json with indexed field (docType is always first), call the index after the field
		fileName: propName + ".json",
	})
}

func (s *SchemaScan) checkForMultiIndex(descr, propName string) {
	start := strings.LastIndex(descr, multiIndexMagicStart)
	if start == -1 {
		return // not found magic
	}

	end := strings.LastIndex(descr[start+len(multiIndexMagicStart):], multiIndexMagicEnd)
	if end == -1 {
		log.Printf("multi index defined in: %s, but not terminated, skipped", propName)
		return
	}
	end += start + len(multiIndexMagicStart)

	multiDef := descr[start:end]

	fields := strings.Split(multiDef, ",")

	if len(fields) != 2 {
		log.Printf("multi index defined in: %s has unexpected format: %s, expected: <order>,<index_name>, skipped", propName, multiDef)
		return
	}

	order := fields[0]
	indexName := fields[1]
	fieldName := propName

	indexNameSection, exists := s.multiIndexes[indexName]
	if !exists {
		// create if not exists
		indexNameSection = map[string]string{}
		s.multiIndexes[indexName] = indexNameSection
	}

	_, orderIsUsed := indexNameSection[order]
	if orderIsUsed {
		log.Printf("multi index: %s, fieldName: %s redeclared order: %s, skipped", indexName, fieldName, order)
		return
	}

	indexNameSection[order] = fieldName
}

func (s *SchemaScan) collectMultiIndexes() error {
	// collect everything prepared in s.multiIndexes to form of s.indexFiles

	keys := make([]string, 0, len(s.multiIndexes))
	for k, _ := range s.multiIndexes {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, multiIndexName := range keys {
		orderMap := s.multiIndexes[multiIndexName]

		orderKeys := make([]string, 0, len(orderMap))
		for k, _ := range orderMap {
			orderKeys = append(orderKeys, k)
		}
		sort.Strings(orderKeys)

		fieldNames := make([]string, 0, len(orderMap))
		for _, orderKey := range orderKeys {
			fieldNames = append(fieldNames, orderMap[orderKey])
		}

		if len(fieldNames) < 2 {
			return fmt.Errorf("multi index: %s defines only one field", multiIndexName)
		}

		indexName := "multi." + strings.Join(fieldNames, ".")

		s.indexFiles = append(s.indexFiles, IndexFile{
			data:     getIndexMap(append([]string{docType}, fieldNames...), indexName),
			fileName: indexName + ".json",
		})
	}

	return nil
}

func sortedKeys(m rmap.Rmap) []string {
	keys := make([]string, 0, len(m.Mapa))
	for k, _ := range m.Mapa {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
