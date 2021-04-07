package mockblogicfail

import (
	"fmt"

	"github.com/KompiTech/fabric-cc-core/v2/src/engine"
	"github.com/KompiTech/rmap"
)

var Fail = func(ctx engine.ContextInterface, prePatch *rmap.Rmap, postPatch rmap.Rmap) (rmap.Rmap, error) {
	return rmap.Rmap{}, fmt.Errorf("business logic created fail")
}
