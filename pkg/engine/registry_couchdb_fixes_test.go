package engine

import (
	"testing"

	"github.com/KompiTech/rmap"
	"github.com/stretchr/testify/assert"
)

func TestRegistry_FindRegexOperators(t *testing.T) {
	query := rmap.NewFromMap(map[string]interface{}{
		"selector": map[string]interface{}{
			"something": map[string]interface{}{
				"$regex": "keke",
			},
		},
	})

	rs, err := findRegexOperators(query.Mapa, []string{})
	assert.Nil(t, err)
	assert.Equal(t, rs, []string{"/selector/something/$regex"})
}

func TestRegistry_FixQuery1(t *testing.T) {
	query := rmap.NewFromMap(map[string]interface{}{
		"selector": map[string]interface{}{
			"something": map[string]interface{}{
				"$regex": "keke",
			},
		},
		"sort": []interface{}{"something", "something2"},
	})

	err := fixQueryForCouchDB(query)
	assert.Nil(t, err)
	assert.Equal(t, "(?i)keke", query.MustGetJPtrString("/selector/something/$regex"))
	assert.Equal(t, []interface{}{"something", "something2"}, query.MustGetIterable("sort"))
}

func TestRegistry_FixQuery2(t *testing.T) {
	query := rmap.NewFromMap(map[string]interface{}{
		"selector": map[string]interface{}{
			"something": map[string]interface{}{
				"$regex": "(?i)keke",
			},
		},
		"sort": []interface{}{map[string]interface{}{"field1": "DESC"}, map[string]interface{}{"field2": "AsC"}},
	})

	err := fixQueryForCouchDB(query)
	assert.Nil(t, err)
	assert.Equal(t, "(?i)keke", query.MustGetJPtrString("/selector/something/$regex"))
	assert.Equal(t, []interface{}{map[string]interface{}{"field1": "desc"}, map[string]interface{}{"field2": "asc"}}, query.MustGetIterable("sort"))
}
