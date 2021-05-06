package metainfgen

import (
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/KompiTech/rmap"
	"github.com/stretchr/testify/assert"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

var testdir string

func prepareEnv() {
	testdir = filepath.Join("/tmp", "metainfgen_test-" + randStringRunes(10))
	if err := os.MkdirAll(testdir, 0755); err != nil {
		panic(err)
	}
}

func teardownEnv() {
	if err := os.RemoveAll(testdir); err != nil {
		panic(err)
	}
}

func TestWriter(t *testing.T) {
	prepareEnv()

	schemas, err := NewScanner("../../internal/testdata/metainfgen")
	assert.Nil(t, err)

	wr, err := NewWriter(testdir)
	assert.Nil(t, err)

	err = wr.WriteIndexFiles(schemas)
	assert.Nil(t, err)

	statePath := filepath.Join(testdir, "statedb/couchdb/indexes")

	aa, err := rmap.NewFromYAMLFile(filepath.Join(statePath, "additional_assignees.json"))
	assert.Nil(t, err)
	assert.Equal(t, "additional_assignees", aa.MustGetString("ddoc"))
	assert.Equal(t, "additional_assignees", aa.MustGetString("name"))
	assert.Equal(t, "json", aa.MustGetString("type"))
	assert.Equal(t, []interface{}{"docType", "additional_assignees"}, aa.MustGetJPtr("/index/fields"))

	at, err := rmap.NewFromYAMLFile(filepath.Join(statePath, "assigned_to.json"))
	assert.Nil(t, err)
	assert.Equal(t, "assigned_to", at.MustGetString("ddoc"))
	assert.Equal(t, "assigned_to", at.MustGetString("name"))
	assert.Equal(t, "json", at.MustGetString("type"))
	assert.Equal(t, []interface{}{"docType", "assigned_to"}, at.MustGetJPtr("/index/fields"))

	dt, err := rmap.NewFromYAMLFile(filepath.Join(statePath, "doctype.json"))
	assert.Nil(t, err)
	assert.Equal(t, "docType", dt.MustGetString("ddoc"))
	assert.Equal(t, "docType", dt.MustGetString("name"))
	assert.Equal(t, "json", dt.MustGetString("type"))
	assert.Equal(t, []interface{}{"docType"}, dt.MustGetJPtr("/index/fields"))

	tl, err := rmap.NewFromYAMLFile(filepath.Join(statePath, "timelogs.json"))
	assert.Nil(t, err)
	assert.Equal(t, "timelogs", tl.MustGetString("ddoc"))
	assert.Equal(t, "timelogs", tl.MustGetString("name"))
	assert.Equal(t, "json", tl.MustGetString("type"))
	assert.Equal(t, []interface{}{"docType", "timelogs"}, tl.MustGetJPtr("/index/fields"))

	uuid, err := rmap.NewFromYAMLFile(filepath.Join(statePath, "uuid.json"))
	assert.Nil(t, err)
	assert.Equal(t, "uuid", uuid.MustGetString("ddoc"))
	assert.Equal(t, "uuid", uuid.MustGetString("name"))
	assert.Equal(t, "json", uuid.MustGetString("type"))
	assert.Equal(t, []interface{}{"docType", "uuid"}, uuid.MustGetJPtr("/index/fields"))

	collectionPath := filepath.Join(testdir, "statedb/couchdb/collections/MOCKCOMMENT/indexes")

	entity, err := rmap.NewFromYAMLFile(filepath.Join(collectionPath, "entity.json"))
	assert.Nil(t, err)
	assert.Equal(t, "entity", entity.MustGetString("ddoc"))
	assert.Equal(t, "entity", entity.MustGetString("name"))
	assert.Equal(t, "json", entity.MustGetString("type"))
	assert.Equal(t, []interface{}{"docType", "entity"}, entity.MustGetJPtr("/index/fields"))

	text, err := rmap.NewFromYAMLFile(filepath.Join(collectionPath, "text.json"))
	assert.Nil(t, err)
	assert.Equal(t, "text", text.MustGetString("ddoc"))
	assert.Equal(t, "text", text.MustGetString("name"))
	assert.Equal(t, "json", text.MustGetString("type"))
	assert.Equal(t, []interface{}{"docType", "text"}, text.MustGetJPtr("/index/fields"))

	uuid2, err := rmap.NewFromYAMLFile(filepath.Join(collectionPath, "uuid.json"))
	assert.Nil(t, err)
	assert.Equal(t, "uuid", uuid2.MustGetString("ddoc"))
	assert.Equal(t, "uuid", uuid2.MustGetString("name"))
	assert.Equal(t, "json", uuid2.MustGetString("type"))
	assert.Equal(t, []interface{}{"docType", "uuid"}, uuid2.MustGetJPtr("/index/fields"))

	doctype2, err := rmap.NewFromYAMLFile(filepath.Join(collectionPath, "doctype.json"))
	assert.Nil(t, err)
	assert.Equal(t, "docType", doctype2.MustGetString("ddoc"))
	assert.Equal(t, "docType", doctype2.MustGetString("name"))
	assert.Equal(t, "json", doctype2.MustGetString("type"))
	assert.Equal(t, []interface{}{"docType"}, doctype2.MustGetJPtr("/index/fields"))

	teardownEnv()
}
