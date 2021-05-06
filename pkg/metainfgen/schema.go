package metainfgen

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/KompiTech/rmap"
)

const (
	docType              = "docType"
	indexMagic           = "_INDEX_"
	multiIndexMagicStart = "_MULTI:" // multi indexes must follow format _MULTI:<order>:<index_name>_. order controls only sorting of fields in multiindex and cannot be repeated
	multiIndexMagicEnd   = "_"
	description          = "description"
	typ                  = "type"
	schemaRoot           = "/schema/properties"
	properties           = "properties"
	stateDestination     = "state"
)

// Schema is parsed YAML schema containing all indexes discovered
type Schema struct {
	Name        string
	Destination string
	Indexes []rmap.Rmap
	multiIndexes map[string]map[string]string // name -> order -> field name
}

// NewSchema parses YAML data from argument and returns Schema obj
func NewSchema(name string, schema rmap.Rmap) (Schema, error) {
	sch := &Schema{
		Name: name,
		Indexes: []rmap.Rmap{},
		multiIndexes: map[string]map[string]string{},
	}

	// scan schema for all indexes and create single indexes
	if err := sch.scan(schema); err != nil {
		return *sch, err
	}

	// create multi indexes for data prepared by scan()
	if err := sch.generateMultiIndexes(); err != nil {
		return *sch, err
	}

	return *sch, nil
}

func (s *Schema) scan(schema rmap.Rmap) error {
	destination, err := schema.GetString("destination")
	if err != nil {
		return err
	}

	s.Destination = destination

	// get schema root and scan for indexes
	props, err := schema.GetJPtrRmap(schemaRoot)
	if err != nil {
		return err
	}

	// scan all schema attributes, iterate in sorted order, so it is deterministic (tests)
	for _, k := range sortedKeys(props) {
		v := rmap.MustNewFromInterface(props.Mapa[k])
		if err := s.handleProperty(k, v); err != nil {
			return err
		}
	}

	return nil
}

func (s *Schema) handleProperty(propName string, property rmap.Rmap) error {
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

	// current obj must contain description, if it is to define any index
	if !property.Exists(description) {
		return nil
	}

	descr, err := property.GetString(description)
	if err != nil {
		return nil
	}

	if strings.Contains(descr, indexMagic) {
		// found single index
		idx := []string{docType, propName}
		s.Indexes = append(s.Indexes, getIndexMap(idx, propName))
		log.Printf("Found index: %v, on property: %s, on schema: %s", idx, propName, s.Name)
	} else if multiStart := strings.LastIndex(descr, multiIndexMagicStart); multiStart != -1 {
		// found multi index
		s.handleMultiIndex(descr, propName)
	}

	return nil
}


func (s *Schema) handleMultiIndex(descr, propName string) {
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

func (s *Schema) generateMultiIndexes() error {
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

		idx := append([]string{docType}, fieldNames...)
		s.Indexes = append(s.Indexes, getIndexMap(idx, indexName))
		log.Printf("Found multi index: %v, name: %s on schema: %s", idx, indexName, s.Name)
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

// get map in couchdb format for some fields to index
func getIndexMap(fields []string, name string) rmap.Rmap {
	return rmap.NewFromMap(map[string]interface{}{
		"index": map[string]interface{}{
			"fields": fields,
		},
		"ddoc": name,
		"name": name,
		"type": "json",
	})
}
