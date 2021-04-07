package reusable

import (
	"github.com/KompiTech/fabric-cc-core/v2/src/engine"
	"github.com/KompiTech/fabric-cc-core/v2/src/kompiguard"
	"github.com/KompiTech/fabric-cc-core/v2/src/konst"
	"github.com/KompiTech/rmap"
	"github.com/pkg/errors"
)

func authorize(ctx engine.ContextInterface, asset rmap.Rmap, action string) (rmap.Rmap, error) {
	thisIdentity, err := ctx.Get("registry").(*engine.Registry).GetThisIdentityResolved()
	if err != nil {
		return rmap.Rmap{}, errors.Wrap(err, "reg.GetThisIdentityResolved() failed")
	}

	kmpg, err := kompiguard.New()
	if err != nil {
		return rmap.Rmap{}, errors.Wrap(err, "kompiguard.New() failed")
	}

	granted, reason, err := kmpg.EnforceAsset(asset, thisIdentity, action)
	if err != nil {
		return rmap.Rmap{}, errors.Wrap(err, "kompiguard.New().EnforceAsset() failed")
	}

	if !granted {
		return rmap.Rmap{}, engine.ErrorForbidden(reason)
	}

	return asset, nil
}

// EnforceRead requires read action granted on asset
var EnforceRead = func(ctx engine.ContextInterface, prePatch *rmap.Rmap, postPatch rmap.Rmap) (rmap.Rmap, error) {
	return authorize(ctx, postPatch, "read")
}

// EnforceCreate requires create action granted on asset
var EnforceCreate = func(ctx engine.ContextInterface, prePatch *rmap.Rmap, postPatch rmap.Rmap) (rmap.Rmap, error) {
	return authorize(ctx, postPatch, "create")
}

// EnforceUpdate requires update action granted on asset
var EnforceUpdate = func(ctx engine.ContextInterface, prePatch *rmap.Rmap, postPatch rmap.Rmap) (rmap.Rmap, error) {
	return authorize(ctx, postPatch, "update")
}

// EnforceDelete requires delete action granted on asset
var EnforceDelete = func(ctx engine.ContextInterface, prePatch *rmap.Rmap, postPatch rmap.Rmap) (rmap.Rmap, error) {
	return authorize(ctx, postPatch, "delete")
}

// FilterRead replaces asset with censored version if read action is not granted
var FilterRead = func(ctx engine.ContextInterface, prePatch *rmap.Rmap, postPatch rmap.Rmap) (rmap.Rmap, error) {
	_, err := authorize(ctx, postPatch, "read")

	if err != nil {
		// not granted, replace
		return rmap.NewFromMap(map[string]interface{}{
			konst.AssetIdKey: postPatch.Mapa[konst.AssetIdKey],
			"error":          "permission denied",
		}), nil
	}
	// granted, return original
	return postPatch, nil
}

// Deny is used whenever we just want to disable some functionality
var Deny = func(ctx engine.ContextInterface, prePatch *rmap.Rmap, postPatch rmap.Rmap) (rmap.Rmap, error) {
	return rmap.Rmap{}, engine.ErrorForbidden("operation is denied")
}
