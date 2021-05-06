package metainfgen

import (
	"testing"

	"github.com/KompiTech/rmap"
	"github.com/stretchr/testify/assert"
)

const sampleSch = `
destination: state
schema:
  title: MockIncident
  description: An Incident is ITSM managed item
  type: object
  properties:
    description:
      description: Description of stuff
      type: string
    zzzz:
      description: _INDEX_
      type: string
    assigned_to:
      description: aaaaaa_INDEX_bbbbbbbbb
      type: string
    additional_assignees:
      description: _INDEX_xxxxxxxxxxxxxxxx
      type: string
    timelogs:
      description: xxxxxxxxxxxxxxxxx_INDEX_
      type: string
    deep:
      type: object
      properties:
        level2:
          description: _INDEX_
          type: string
    multiField1:
      type: string
      description: _MULTI:0,multi1_
    multiField2:
      type: string
      description: _MULTI:1,multi1_
  required:
    - description
  additionalProperties: false
`

func TestSchema(t *testing.T) {
	rm := rmap.MustNewFromYAMLBytes([]byte(sampleSch))

	sch, err := NewSchema("mockincident", rm)
	assert.Nil(t, err)
	assert.Equal(t, "state", sch.Destination)
	assert.Equal(t, 6, len(sch.Indexes))

	aa := sch.Indexes[0]
	assert.Equal(t, "additional_assignees", aa.MustGetString("ddoc"))
	assert.Equal(t, []string{"docType", "additional_assignees"}, aa.MustGetJPtr("/index/fields"))
	assert.Equal(t, "additional_assignees", aa.MustGetString("name"))
	assert.Equal(t, "json", aa.MustGetString("type"))

	at := sch.Indexes[1]
	assert.Equal(t, "assigned_to", at.MustGetString("ddoc"))
	assert.Equal(t, []string{"docType", "assigned_to"}, at.MustGetJPtr("/index/fields"))
	assert.Equal(t, "assigned_to", at.MustGetString("name"))
	assert.Equal(t, "json", at.MustGetString("type"))

	l2 := sch.Indexes[2]
	assert.Equal(t, "deep.level2", l2.MustGetString("ddoc"))
	assert.Equal(t, []string{"docType", "deep.level2"}, l2.MustGetJPtr("/index/fields"))
	assert.Equal(t, "deep.level2", l2.MustGetString("name"))
	assert.Equal(t, "json", l2.MustGetString("type"))

	tl := sch.Indexes[3]
	assert.Equal(t, "timelogs", tl.MustGetString("ddoc"))
	assert.Equal(t, []string{"docType", "timelogs"}, tl.MustGetJPtr("/index/fields"))
	assert.Equal(t, "timelogs", tl.MustGetString("name"))
	assert.Equal(t, "json", tl.MustGetString("type"))

	zz := sch.Indexes[4]
	assert.Equal(t, "zzzz", zz.MustGetString("ddoc"))
	assert.Equal(t, []string{"docType", "zzzz"}, zz.MustGetJPtr("/index/fields"))
	assert.Equal(t, "zzzz", zz.MustGetString("name"))
	assert.Equal(t, "json", zz.MustGetString("type"))

	mi := sch.Indexes[5]
	assert.Equal(t, "multi.multiField1.multiField2", mi.MustGetString("ddoc"))
	assert.Equal(t, []string{"docType", "multiField1", "multiField2"}, mi.MustGetJPtr("/index/fields"))
	assert.Equal(t, "multi.multiField1.multiField2", mi.MustGetString("name"))
	assert.Equal(t, "json", mi.MustGetString("type"))
}