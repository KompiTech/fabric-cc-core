package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolver_SchemaTransform(t *testing.T) {
	// test correct transformation between asset and schema jptr
	r := resolver{}

	jptr, err := r.getDescriptionJPtr("/attr")
	assert.Nil(t, err)
	assert.Equal(t, "/properties/attr/description", jptr)

	jptr, err = r.getDescriptionJPtr("/arr/0")
	assert.Nil(t, err)
	assert.Equal(t, "/properties/arr/description", jptr)

	jptr, err = r.getDescriptionJPtr("/arr/0/attr")
	assert.Nil(t, err)
	assert.Equal(t, "/properties/arr/items/properties/attr/description", jptr)

	jptr, err = r.getDescriptionJPtr("/arr/0/attr/0/attr2")
	assert.Nil(t, err)
	assert.Equal(t, "/properties/arr/items/properties/attr/items/properties/attr2/description", jptr)
}
