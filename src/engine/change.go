package engine

import (
	. "github.com/KompiTech/fabric-cc-core/v2/src/konst"

	"github.com/KompiTech/rmap"
)

type Change struct {
	AssetName string
	Version   int
	Operation string
}

func (c Change) IsEmpty() bool {
	if c.AssetName == "" && c.Version == 0 && c.Operation == "" {
		return true
	}
	return false
}

func (c Change) Rmap() rmap.Rmap {
	m := map[string]interface{}{
		ChangelogAssetNameKey: c.AssetName,
		ChangelogVersionKey:   c.Version,
		ChangelogOperationKey: c.Operation,
	}
	return rmap.NewFromMap(m)
}
