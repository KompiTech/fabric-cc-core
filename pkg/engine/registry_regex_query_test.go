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

func TestRegistry_FixRegexes(t *testing.T) {
	query := rmap.NewFromMap(map[string]interface{}{
		"selector": map[string]interface{}{
			"something": map[string]interface{}{
				"$regex": "keke",
			},
		},
	})

	err := fixRegexes(query)
	assert.Nil(t, err)
	assert.Equal(t, "(?i)keke", query.MustGetJPtrString("/selector/something/$regex"))
}

func TestRegistry_FixRegexes2(t *testing.T) {
	query := rmap.NewFromMap(map[string]interface{}{
		"selector": map[string]interface{}{
			"something": map[string]interface{}{
				"$regex": "(?i)keke",
			},
		},
	})

	err := fixRegexes(query)
	assert.Nil(t, err)
	assert.Equal(t, "(?i)keke", query.MustGetJPtrString("/selector/something/$regex"))
}
