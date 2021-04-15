package engine

import (
	. "github.com/KompiTech/fabric-cc-core/v2/pkg/konst"
	"github.com/pkg/errors"
)

// deprecated, use assetCreate(role, ...)
func roleCreateFrontend(ctx ContextInterface) (string, error) {
	now, err := ctx.Time()
	if err != nil {
		return "", errors.Wrap(err, "ctx.Time() failed")
	}

	data, err := ctx.ParamString(DataParam)
	if err != nil {
		return "", err
	}

	id, err := ctx.ParamString(IdParam)
	if err != nil {
		return "", err
	}

	return assetCreateBackend(ctx, now, RoleAssetName, data, -1, id, false)
}

// deprecated, use assetUpdate(role, ...)
func roleUpdateFrontend(ctx ContextInterface) (string, error) {
	id, err := ctx.ParamString(IdParam)
	if err != nil {
		return "", err
	}

	patch, err := ctx.ParamString(PatchParam)
	if err != nil {
		return "", err
	}

	return assetUpdateBackend(ctx, RoleAssetName, id, patch, false)
}

// deprecated, use assetGet(role, ...)
func roleGetFrontend(ctx ContextInterface) (string, error) {
	id, err := ctx.ParamString(IdParam)
	if err != nil {
		return "", err
	}

	data, err := ctx.ParamString(DataParam)
	if err != nil {
		return "", err
	}

	return assetGetBackend(ctx, RoleAssetName, id, false, data, false)
}

// deprecated, use assetQuery(role, ...)
func roleQueryFrontend(ctx ContextInterface) (string, error) {
	query, err := ctx.ParamString(QueryParam)
	if err != nil {
		return "", err
	}

	return assetQueryBackend(ctx, RoleAssetName, query, false, PageSize, false)
}
