package engine

import (
	"strings"

	. "github.com/KompiTech/rmap"
)

// CouchDB-specific fixes go here

func fixQueryForCouchDB(q Rmap) error {
	// find all $regex operators and make them case-insensitive for CouchDB
	rs, err := findRegexOperators(q.Mapa, []string{})
	if err != nil {
		return err
	}

	if err := addRegexPrefix(q, rs); err != nil {
		return err
	}

	// if there are sort operators, make the direction of sort lowercase - CouchDB does not like uppercase

	return nil
}

func fixSortOperators(q Rmap) error {
	if q.Exists("sort") {

	}
}

// returns list of all jsonptrs that have the $regex key
func findRegexOperators(q map[string]interface{}, path []string) ([]string, error) {
	res := []string{}

	for k, v := range q {
		if k == "$regex" {
			_, isString := v.(string)
			if isString {
				res = append(res, "/"+strings.Join(append(path, k), "/"))
			}
		}

		if subMap, isMap := v.(map[string]interface{}); isMap {
			rop, err := findRegexOperators(subMap, append(path, k))
			if err != nil {
				return nil, err
			}

			res = append(res, rop...)
		}
	}

	return res, nil
}

// adds (?i) to each $regex value (if it doesn't already start with it) to make it case-insensitive
func addRegexPrefix(q Rmap, jptrs []string) error {
	for _, jptr := range jptrs {
		oldValue, err := q.GetJPtrString(jptr)
		if err != nil {
			return err
		}

		oldValue = strings.TrimSpace(oldValue)

		if !strings.HasPrefix(oldValue, regexCaseInsensitive) {
			// prefix is not present -> add it
			newValue := regexCaseInsensitive + oldValue

			if err := q.SetJPtr(jptr, newValue); err != nil {
				return err
			}
		}
	}

	return nil
}
