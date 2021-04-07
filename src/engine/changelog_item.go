package engine

import (
	"time"

	. "github.com/KompiTech/fabric-cc-core/v2/src/konst"
	"github.com/KompiTech/rmap"
)

type ChangelogItem struct {
	Timestamp time.Time
	TxId      string
	Changes   []Change
}

func (ci ChangelogItem) Rmap() rmap.Rmap {
	m := map[string]interface{}{
		ChangelogTimestampKey: ci.Timestamp.Format(time.RFC3339),
		ChangelogTxIdKey:      ci.TxId,
	}
	changes := make([]interface{}, len(ci.Changes))
	for index, change := range ci.Changes {
		changes[index] = change.Rmap().Mapa
	}
	m[ChangelogChangesKey] = changes
	return rmap.NewFromMap(m)
}

func (ci ChangelogItem) IsEmpty() bool {
	if len(ci.Changes) == 0 {
		return true
	}
	return false
}
