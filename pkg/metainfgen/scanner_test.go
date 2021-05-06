package metainfgen

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScanner(t *testing.T) {
	schemas, err := NewScanner("../../internal/testdata/metainfgen")
	assert.Nil(t, err)
	assert.Equal(t, 2, len(schemas))

	cm := schemas[0]
	assert.Equal(t, 2, len(cm.Indexes))
	en := cm.Indexes[0]
	assert.Equal(t, "entity", en.MustGetString("ddoc"))
	assert.Equal(t, []string{"docType", "entity"}, en.MustGetJPtr("/index/fields"))
	assert.Equal(t, "entity", en.MustGetString("name"))
	assert.Equal(t, "json", en.MustGetString("type"))

	tx := cm.Indexes[1]
	assert.Equal(t, "text", tx.MustGetString("ddoc"))
	assert.Equal(t, []string{"docType", "text"}, tx.MustGetJPtr("/index/fields"))
	assert.Equal(t, "text", tx.MustGetString("name"))
	assert.Equal(t, "json", tx.MustGetString("type"))

	in := schemas[1]
	assert.Equal(t, 3, len(in.Indexes))
	aa := in.Indexes[0]
	assert.Equal(t, "additional_assignees", aa.MustGetString("ddoc"))
	assert.Equal(t, []string{"docType", "additional_assignees"}, aa.MustGetJPtr("/index/fields"))
	assert.Equal(t, "additional_assignees", aa.MustGetString("name"))
	assert.Equal(t, "json", aa.MustGetString("type"))

	at := in.Indexes[1]
	assert.Equal(t, "assigned_to", at.MustGetString("ddoc"))
	assert.Equal(t, []string{"docType", "assigned_to"}, at.MustGetJPtr("/index/fields"))
	assert.Equal(t, "assigned_to", at.MustGetString("name"))
	assert.Equal(t, "json", at.MustGetString("type"))

	tl := in.Indexes[2]
	assert.Equal(t, "timelogs", tl.MustGetString("ddoc"))
	assert.Equal(t, []string{"docType", "timelogs"}, tl.MustGetJPtr("/index/fields"))
	assert.Equal(t, "timelogs", tl.MustGetString("name"))
	assert.Equal(t, "json", tl.MustGetString("type"))
}
