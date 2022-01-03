package engine

import (
	"fmt"
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
	return fixSortOperators(q)
}

func fixSortOperators(q Rmap) error {
	if !q.Exists("sort") {
		return nil
	}

	// values can be: ["field1", "field2", ...]
	// or: [{"field1": "ASC"}, {"field2": "DESC"}, ...]
	// we care only about object form

	iter, err := q.GetIterable("sort")
	if err != nil {
		return err
	}

	for _, sortI := range iter {
		sort, ok := sortI.(map[string]interface{})
		if !ok {
			// if no object is found, stop the search (not possible to be there later at this point)
			break
		}

		sortRM := NewFromMap(sort)

		// only one key per sort obj
		if len(sort) > 1 {
			return fmt.Errorf("invalid sort: %s", sortRM.String())
		}

		var k string

		// to get the key ... we dont know the name, so need to iterate
		for k = range sort {
			break
		}

		oldVal, err := sortRM.GetString(k)
		if err != nil {
			return err
		}

		sort[k] = strings.ToLower(oldVal)
	}

	return nil
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
