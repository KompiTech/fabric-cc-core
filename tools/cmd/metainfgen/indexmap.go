package main

import "github.com/KompiTech/rmap"

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
