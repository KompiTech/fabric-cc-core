package mockdataafterresolve

import (
	"fmt"

	"github.com/KompiTech/fabric-cc-core/v2/pkg/engine"
	"github.com/KompiTech/rmap"
	"github.com/pkg/errors"
)

var TestPassing = func(ctx engine.ContextInterface, prePatch *rmap.Rmap, postPatch rmap.Rmap) (rmap.Rmap, error) {
	if prePatch != nil {
		val, err := prePatch.GetString("magic")
		if err != nil {
			return rmap.Rmap{}, errors.Wrap(err, "prePatch.GetString() failed")
		}

		if val == "value" {
			return rmap.Rmap{}, fmt.Errorf("magic value found, param passing works")
		}
	}

	return postPatch, nil
}
