package engine

import (
	"fmt"
	"strings"
	"time"

	"github.com/KompiTech/fabric-cc-core/v2/pkg/kompiguard"
	. "github.com/KompiTech/fabric-cc-core/v2/pkg/konst"
	"github.com/KompiTech/rmap"
	"github.com/pkg/errors"
)

func assetGetBackend(ctx ContextInterface, name, id string, resolve bool, data string, isDirect bool) (string, error) {
	// get desired asset
	asset, err := ctx.GetRegistry().GetAsset(name, id, resolve, true)
	if err != nil {
		return "", errors.Wrap(err, "reg.GetAsset() failed")
	}

	docType, err := AssetGetDocType(asset)
	if err != nil {
		return "", errors.Wrap(err, "asset.GetDocType() failed")
	}
	docType = strings.ToLower(docType)

	if isDirect {
		// when isDirect, explicit permission is required
		thisIdentity, err := ctx.GetRegistry().GetThisIdentityResolved()
		if err != nil {
			return "", errors.Wrap(err, "reg.GetThisIdentityResolved() failed")
		}

		kmpg, err := kompiguard.New()
		if err != nil {
			return "", errors.Wrap(err, "kompiguard.New() failed")
		}

		granted, reason, err := kmpg.EnforceAsset(asset, thisIdentity, ReadDirectAction)
		if err != nil {
			return "", errors.Wrap(err, "kmpg.EnforceCustom()")
		}

		if !granted {
			return "", ErrorForbidden(reason)
		}
	}

	// built-in assets need to be protected in non-direct mode
	if (docType == IdentityAssetName || docType == RoleAssetName) && !isDirect {
		// get identity of this user
		thisIdentity, err := ctx.GetRegistry().GetThisIdentityResolved()
		if err != nil {
			return "", errors.Wrap(err, "reg.GetThisIdentityResolved() failed")
		}

		assetID, err := AssetGetID(asset)
		if err != nil {
			return "", errors.Wrap(err, "AssetGetID(asset) failed")
		}

		thisIdentityID, err := AssetGetID(thisIdentity)
		if err != nil {
			return "", errors.Wrap(err, "AssetGetID(thisIdentity) failed")
		}

		if !(docType == IdentityAssetName && assetID == thisIdentityID) {
			// when client reads anything else than his own identity, standard access control is used
			kmpg, err := kompiguard.New()
			if err != nil {
				return "", errors.Wrap(err, "kompiguard.New() failed")
			}

			granted, reason, err := kmpg.EnforceAsset(asset, thisIdentity, ReadAction)
			if err != nil {
				return "", err
			}

			if !granted {
				// no, return permission denied message
				return "", ErrorForbidden(reason)
			}
		}
	}

	if !isDirect {
		// execute business logic AfterGet when not direct
		var dataR rmap.Rmap

		if len(data) == 0 {
			dataR = rmap.NewEmpty()
		} else {
			dataR, err = rmap.NewFromString(data)
			if err != nil {
				return "", errors.Wrap(err, "rmap.NewFromBytes() failed")
			}
		}

		asset, err = ctx.GetConfiguration().BusinessExecutor.Execute(ctx, AfterGet, &dataR, asset)
		if err != nil {
			return "", errors.Wrap(err, "bexec.Execute(), stage: AfterGet failed")
		}
	}

	return string(asset.WrappedResultBytes()), nil
}

func assetUpdateBackend(ctx ContextInterface, name, id string, patchBytes string, isDirect bool) (string, error) {
	// load json patch
	patch, err := rmap.NewFromString(patchBytes)
	if err != nil {
		return "", errors.Wrap(err, "rmap.NewFromBytes() failed")
	}

	// get asset that client wants to update
	assetPre, err := ctx.GetRegistry().GetAsset(name, id, false, true)
	if err != nil {
		return "", errors.Wrap(err, "reg.GetAsset() failed")
	}

	docType, err := AssetGetDocType(assetPre)
	if err != nil {
		return "", errors.Wrap(err, "assetPre.GetDocType() failed")
	}

	var requiredAction string

	if !isDirect {
		if docType == IdentityAssetName {
			// when updating identity in non-direct mode, require update action
			requiredAction = UpdateAction
		} else if docType == RoleAssetName {
			var isSystemRole bool

			if assetPre.Exists(RoleIsSystemKey) {
				// system role key is present, read value
				isSystemRole, err = assetPre.GetBool(RoleIsSystemKey)
				if err != nil {
					return "", err
				}
			} else {
				// no system role key, ordinary role
				isSystemRole = false
			}

			// when updating role in non-direct mode, require either update action or update_system (depends if particular role is system or not)
			if isSystemRole {
				requiredAction = UpdateSystemAction
			} else {
				requiredAction = UpdateAction
			}
		}
	} else {
		// when updating anything in direct mode, require update_direct action
		requiredAction = UpdateDirectAction
	}

	// enforce built-in kompiguard if required
	if requiredAction != "" {
		if err := enforceAssetAccess(ctx.GetRegistry(), assetPre, requiredAction); err != nil {
			return "", err
		}
	}

	if patch.IsEmpty() {
		return "", ErrorBadRequest("patch is empty")
	}

	if HasServiceKeys(patch) {
		return "", ErrorBadRequest("patch contains service key(s)")
	}

	if !isDirect {
		// execute PatchUpdate and FirstUpdate blogic if not running in direct mode
		_, err = ctx.GetConfiguration().BusinessExecutor.ExecuteCustom(ctx, PatchUpdate, assetPre, &assetPre, patch)
		if err != nil {
			return "", errors.Wrap(err, "bexec.ExecuteCustom(), stage: PatchUpdate failed")
		}

		assetPre, err = ctx.GetConfiguration().BusinessExecutor.Execute(ctx, FirstUpdate, nil, assetPre)
		if err != nil {
			return "", errors.Wrap(err, "bexec.Execute(), stage: FirstUpdate failed")
		}
	}

	// make copy before patching, so changes can be tracked
	assetPost := assetPre.Copy()

	// apply JSON patch
	if err := assetPost.ApplyMergePatch(patch); err != nil {
		return "", errors.Wrap(err, "assetPre.ApplyMergePatch() failed")
	}

	if !isDirect {
		// execute BeforeUpdate blogic if not running in direct mode
		assetPost, err = ctx.GetConfiguration().BusinessExecutor.Execute(ctx, BeforeUpdate, &assetPre, assetPost)
		if err != nil {
			return "", errors.Wrap(err, "bexec.Execute(), stage: BeforeUpdate failed")
		}
	}

	if docType == IdentityAssetName {
		// execute extra validations for identity in direct and not direct mode
		// this protects system from last superuser taking his permission away
		thisIdentity, err := ctx.GetRegistry().GetThisIdentityResolved()
		if err != nil {
			return "", errors.Wrap(err, "reg.GetThisIdentityResolved() failed")
		}

		if err := identityExtraValidate(ctx.GetRegistry(), assetPre, assetPost, thisIdentity); err != nil {
			return "", errors.Wrap(err, "identityExtraValidate() failed")
		}
	}

	// persist changes to modified asset
	if err := ctx.GetRegistry().putAsset(assetPost, false, isDirect); err != nil {
		return "", errors.Wrap(err, "reg.PutAsset() failed")
	}

	if !isDirect {
		// execute AfterUpdate blogic if not running in direct mode
		assetPost, err = ctx.GetConfiguration().BusinessExecutor.Execute(ctx, AfterUpdate, &assetPre, assetPost)
		if err != nil {
			return "", errors.Wrap(err, "bexec.Execute(), stage: AfterUpdate failed")
		}
	}

	return string(assetPost.WrappedResultBytes()), nil
}

func assetUpdateFrontend(ctx ContextInterface) (string, error) {
	name, err := ctx.ParamString(NameParam)
	if err != nil {
		return "", err
	}

	id, err := ctx.ParamString(IdParam)
	if err != nil {
		return "", err
	}

	patch, err := ctx.ParamString(PatchParam)
	if err != nil {
		return "", err
	}

	return assetUpdateBackend(ctx, name, id, patch, false)
}

func assetUpdateDirectFrontend(ctx ContextInterface) (string, error) {
	name, err := ctx.ParamString(NameParam)
	if err != nil {
		return "", err
	}

	id, err := ctx.ParamString(IdParam)
	if err != nil {
		return "", err
	}

	patch, err := ctx.ParamString(PatchParam)
	if err != nil {
		return "", err
	}

	return assetUpdateBackend(ctx, name, id, patch, true)
}

func assetGetFrontend(ctx ContextInterface) (string, error) {
	name, err := ctx.ParamString(NameParam)
	if err != nil {
		return "", err
	}

	id, err := ctx.ParamString(IdParam)
	if err != nil {
		return "", err
	}

	resolve, err := ctx.ParamBool(ResolveParam)
	if err != nil {
		return "", err
	}

	data, err := ctx.ParamString(DataParam)
	if err != nil {
		return "", err
	}

	return assetGetBackend(ctx, name, id, resolve, data, false)
}

func assetGetDirectFrontend(ctx ContextInterface) (string, error) {
	name, err := ctx.ParamString(NameParam)
	if err != nil {
		return "", err
	}

	id, err := ctx.ParamString(IdParam)
	if err != nil {
		return "", err
	}

	resolve, err := ctx.ParamBool(ResolveParam)
	if err != nil {
		return "", err
	}

	return assetGetBackend(ctx, name, id, resolve, rmap.NewEmpty().String(), true)
}

func assetCreateBackend(ctx ContextInterface, now time.Time, name string, data string, version int, id string, isDirect bool) (string, error) {
	name = strings.ToLower(name)
	var err error
	var patch rmap.Rmap

	if len(data) == 0 {
		patch = rmap.NewEmpty()
	} else {
		patch, err = rmap.NewFromString(data)
		if err != nil {
			return "", errors.Wrap(err, "rmap.NewFromBytes() failed")
		}
	}

	if name == IdentityAssetName {
		// identity can only be created using direct == true and its ID cannot be autogenerated
		// identity must always be version 1
		if !isDirect {
			return "", ErrorBadRequest("identity can only be created with identityCreateDirect method")
		}

		if version != 1 {
			return "", ErrorBadRequest("identity must be created with version 1")
		}

		if patch.Exists(AssetFingerprintKey) {
			// client sent some fingerprint in data
			fpInData, err := patch.GetString(AssetFingerprintKey)
			if err != nil {
				return "", errors.Wrap(err, "patch.GetString() failed")
			}

			if id != "" {
				// client also sent some fingerprint in CC param, they must match
				if id != fpInData {
					return "", ErrorBadRequest(fmt.Sprintf("fingerprint from CC param: %s does not match fingerprint in data: %s", id, fpInData))
				}
			}
			id = fpInData
			delete(patch.Mapa, AssetFingerprintKey) // delete fingerprint from patch, it will be added when creating asset
		} else {
			// client did not send fingerprint in patch
			if id == "" {
				return "", ErrorBadRequest(fmt.Sprintf("identity fingerprint cannot be autogenerated, you must send one in data or in CC param"))
			}
		}
	} else {
		// other assets than identity
		// determine ID of newly created asset
		// if there is ID inside patch, it must match supplied ID param or ID param must be set to empty
		// if patch does not contain ID, then it is either autogenerated or value is used
		if patch.Exists(AssetIdKey) {
			// client send some id in data
			idInData, err := patch.GetString(AssetIdKey)
			if err != nil {
				return "", errors.Wrap(err, "patch.GetString() failed")
			}

			if id != "" {
				// client also sent some id in CC param, they must match
				if id != idInData {
					return "", ErrorBadRequest(fmt.Sprintf("id from CC param: %s does not match id in data: %s", id, idInData))
				}
			}

			delete(patch.Mapa, AssetIdKey) // delete id from patch, it will be added when creating asset
			id = idInData
		} else {
			// client did not sent id in data
			if id == "" {
				// client wants to autogenerate id
				id, err = MakeUUID(now)
				if err != nil {
					return "", errors.Wrap(err, "engine.MakeUUID() failed")
				}
			}
		}
	}

	if HasServiceKeys(patch) {
		return "", ErrorBadRequest("patch contains service key(s)")
	}

	// create asset with only service keys
	newAsset, err := ctx.GetRegistry().MakeAsset(name, id, version)
	if err != nil {
		return "", errors.Wrap(err, "reg.MakeAsset() failed")
	}

	var requiredAction string

	if !isDirect {
		if name == IdentityAssetName {
			// when create role or identity in non-direct mode, require create action
			requiredAction = CreateAction
		} else if name == RoleAssetName {
			var isSystem bool
			// when creating system role, require create_system action
			if patch.Exists(RoleIsSystemKey) {
				isSystem, err = patch.GetBool(RoleIsSystemKey)
				if err != nil {
					return "", err
				}
			} else {
				isSystem = false
			}

			if !isSystem {
				requiredAction = CreateAction
			} else {
				requiredAction = CreateSystemAction
			}
		}
	} else {
		// when creating anything in direct mode, require create_direct action
		requiredAction = CreateDirectAction
	}

	if requiredAction != "" {
		if err := enforceAssetAccess(ctx.Get("registry").(*Registry), newAsset, requiredAction); err != nil {
			return "", err
		}
	}

	if err := ctx.GetRegistry().MarkAssetAsExisting(name, id, patch); err != nil {
		return "", errors.Wrap(err, "reg.MarkAssetAsExisting() failed")
	}

	if !isDirect {
		// execute blogic in not direct mode
		_, err = ctx.GetConfiguration().BusinessExecutor.ExecuteCustom(ctx, PatchCreate, newAsset, &newAsset, patch)
		if err != nil {
			return "", errors.Wrap(err, "bexec.Execute(), stage: PatchCreate failed")
		}

		_, err = ctx.GetConfiguration().BusinessExecutor.Execute(ctx, FirstCreate, nil, newAsset)
		if err != nil {
			return "", errors.Wrap(err, "bexec.Execute(), stage: FirstCreate failed")
		}
	}

	if err := newAsset.ApplyMergePatch(patch); err != nil {
		return "", errors.Wrap(err, "newAsset.ApplyMergePatch() failed")
	}

	if !isDirect {
		newAsset, err = ctx.GetConfiguration().BusinessExecutor.Execute(ctx, BeforeCreate, nil, newAsset)
		if err != nil {
			return "", errors.Wrap(err, "bexec.Execute(), stage: BeforeCreate failed")
		}
	}

	if err := ctx.Get("registry").(*Registry).putAsset(newAsset, true, isDirect); err != nil {
		return "", errors.Wrap(err, "reg.PutAsset() failed")
	}

	if !isDirect {
		newAsset, err = ctx.GetConfiguration().BusinessExecutor.Execute(ctx, AfterCreate, nil, newAsset)
		if err != nil {
			return "", errors.Wrap(err, "bexec.Execute(), stage: AfterCreate failed")
		}
	}

	return string(newAsset.WrappedResultBytes()), nil
}

func assetCreateDirectFrontend(ctx ContextInterface) (string, error) {
	now, err := ctx.Time()
	if err != nil {
		return "", errors.Wrap(err, "ctx.Time() failed")
	}

	name, err := ctx.ParamString(NameParam)
	if err != nil {
		return "", err
	}

	data, err := ctx.ParamString(DataParam)
	if err != nil {
		return "", err
	}

	version, err := ctx.ParamInt(VersionParam)
	if err != nil {
		return "", err
	}

	id, err := ctx.ParamString(IdParam)
	if err != nil {
		return "", err
	}

	return assetCreateBackend(ctx, now, name, data, version, id, true)
}

func assetCreateFrontend(ctx ContextInterface) (string, error) {
	now, err := ctx.Time()
	if err != nil {
		return "", errors.Wrap(err, "ctx.Time() failed")
	}

	name, err := ctx.ParamString(NameParam)
	if err != nil {
		return "", err
	}

	data, err := ctx.ParamString(DataParam)
	if err != nil {
		return "", err
	}

	version, err := ctx.ParamInt(VersionParam)
	if err != nil {
		return "", err
	}

	id, err := ctx.ParamString(IdParam)
	if err != nil {
		return "", err
	}

	return assetCreateBackend(ctx, now, name, data, version, id, false)
}

func assetQueryFrontend(ctx ContextInterface) (string, error) {
	name, err := ctx.ParamString(NameParam)
	if err != nil {
		return "", err
	}

	query, err := ctx.ParamString(QueryParam)
	if err != nil {
		return "", err
	}

	resolve, err := ctx.ParamBool(ResolveParam)
	if err != nil {
		return "", err
	}

	return assetQueryBackend(ctx, name, query, resolve, PageSize, false)
}

func assetQueryDirectFrontend(ctx ContextInterface) (string, error) {
	name, err := ctx.ParamString(NameParam)
	if err != nil {
		return "", err
	}

	query, err := ctx.ParamString(QueryParam)
	if err != nil {
		return "", err
	}

	resolve, err := ctx.ParamBool(ResolveParam)
	if err != nil {
		return "", err
	}

	return assetQueryBackend(ctx, name, query, resolve, PageSize, true)
}

func assetQueryBackend(ctx ContextInterface, name string, queryBytes string, resolve bool, pageSize int, isDirect bool) (string, error) {
	docType := strings.ToLower(name)
	var query rmap.Rmap

	if len(queryBytes) == 0 {
		query = rmap.NewEmpty()
	} else {
		var err error
		query, err = rmap.NewFromString(queryBytes)
		if err != nil {
			return "", errors.Wrap(err, "rmap.NewFromBytes() failed")
		}
	}

	if isDirect {
		// when isDirect, explicit permission is required
		thisIdentity, err := ctx.GetRegistry().GetThisIdentityResolved()
		if err != nil {
			return "", errors.Wrap(err, "reg.GetThisIdentityResolved() failed")
		}

		kmpg, err := kompiguard.New()
		if err != nil {
			return "", errors.Wrap(err, "kompiguard.New() failed")
		}

		myFP, err := AssetGetID(thisIdentity)
		if err != nil {
			return "", errors.Wrap(err, "AssetGetID(thisIdentity) failed")
		}

		if err := kmpg.LoadRoles(thisIdentity); err != nil {
			return "", errors.Wrap(err, "kmpg.LoadRoles()")
		}

		granted, reason, err := kmpg.EnforceCustom("/"+docType, myFP, "query_direct", nil)
		if err != nil {
			return "", errors.Wrap(err, "kmpg.EnforceCustom()")
		}

		if !granted {
			return "", ErrorForbidden(reason)
		}
	}

	if !isDirect {
		// execute blogic stage BeforeQuery
		// since version does not makes sense here, -1 is used
		var err error
		query, err = ctx.GetConfiguration().BusinessExecutor.ExecuteCustomPolicy(ctx,
			ctx.GetConfiguration().BusinessExecutor.GetPolicy(FuncKey{Name: strings.ToLower(name), Version: -1}, BeforeQuery),
			nil, query)
		if err != nil {
			return "", errors.Wrap(err, "bexec.ExecuteCustomPolicy() failed")
		}
	}

	bookmark := ""
	if query.Exists(QueryBookmarkKey) {
		var err error
		bookmark, err = query.GetString(QueryBookmarkKey)
		if err != nil {
			return "", errors.Wrap(err, "query.GetString()")
		}
	}

	assets, bookmark, err := ctx.GetRegistry().QueryAssets(name, query, bookmark, resolve, true, pageSize)
	if err != nil {
		return "", errors.Wrap(err, "reg.QueryAssets() failed")
	}

	if (docType == IdentityAssetName || docType == RoleAssetName) && !isDirect {
		thisIdentity, err := ctx.GetRegistry().GetThisIdentityResolved()
		if err != nil {
			return "", errors.Wrap(err, "reg.GetThisIdentityResolved() failed")
		}

		kmpg, err := kompiguard.New()
		if err != nil {
			return "", errors.Wrap(err, "kompiguard.New() failed")
		}

		assets, err = kmpg.FilterAssets(assets, thisIdentity, "read")
		if err != nil {
			return "", errors.Wrap(err, "kompiguard.New().FilterAssets() failed")
		}
	}

	outputSlice := make([]interface{}, 0, len(assets))
	for _, asset := range assets {
		if !isDirect {
			var err error
			var assetTmp rmap.Rmap
			assetTmp, err = ctx.GetConfiguration().BusinessExecutor.Execute(ctx, AfterQuery, nil, asset)
			if err != nil {
				return "", errors.Wrap(err, "bexec.Execute(), stage: AfterQuery failed")
			}
			asset = assetTmp
		}

		outputSlice = append(outputSlice, asset.Mapa)
	}

	output := rmap.NewFromMap(map[string]interface{}{
		OutputResultKey:   outputSlice,
		OutputBookmarkKey: bookmark,
	})

	return string(output.Bytes()), nil
}

func assetMigrateFrontend(ctx ContextInterface) (string, error) {
	name, err := ctx.ParamString(NameParam)
	if err != nil {
		return "", err
	}

	id, err := ctx.ParamString(IdParam)
	if err != nil {
		return "", err
	}

	version, err := ctx.ParamInt(VersionParam)
	if err != nil {
		return "", err
	}

	patchB, err := ctx.ParamString(PatchParam)
	if err != nil {
		return "", err
	}

	reg := ctx.Get(RegistryKey).(*Registry)

	asset, err := reg.GetAsset(name, id, false, true)
	if err != nil {
		return "", errors.Wrap(err, "reg.GetAsset() failed")
	}

	if err := enforceAssetAccess(reg, asset, "migrate"); err != nil {
		return "", err
	}

	thisVersion, err := AssetGetVersion(asset)
	if err != nil {
		return "", errors.Wrap(err, "asset.GetVersion() failed")
	}

	_, targetVersion, err := reg.GetItem(name, version)
	if err != nil {
		return "", errors.Wrap(err, "reg.GetItem() failed")
	}

	if thisVersion == targetVersion {
		return "", ErrorBadRequest("unable to migrate to the same version of asset")
	}

	asset.Mapa[AssetVersionKey] = targetVersion

	patch, err := rmap.NewFromString(patchB)
	if err != nil {
		return "", errors.Wrap(err, "rmap.NewFromBytes() failed")
	}

	if HasServiceKeys(patch) {
		return "", ErrorBadRequest("patch contains service key(s)")
	}

	if err := asset.ApplyMergePatchBytes(patch.Bytes()); err != nil {
		return "", errors.Wrap(err, "asset.ApplyMergePatchBytes() failed")
	}

	if err := reg.PutAsset(asset, false); err != nil {
		return "", errors.Wrap(err, "reg.PutAsset() failed")
	}

	return string(asset.WrappedResultBytes()), nil
}

func assetDeleteFrontend(ctx ContextInterface) (string, error) {
	name, err := ctx.ParamString(NameParam)
	if err != nil {
		return "", err
	}

	id, err := ctx.ParamString(IdParam)
	if err != nil {
		return "", err
	}

	return assetDeleteBackend(ctx, name, id, false)
}

func assetDeleteDirectFrontend(ctx ContextInterface) (string, error) {
	name, err := ctx.ParamString(NameParam)
	if err != nil {
		return "", err
	}

	id, err := ctx.ParamString(IdParam)
	if err != nil {
		return "", err
	}

	return assetDeleteBackend(ctx, name, id, true)
}

func assetDeleteBackend(ctx ContextInterface, name, id string, isDirect bool) (string, error) {
	asset, err := ctx.GetRegistry().GetAsset(name, id, false, true)
	if err != nil {
		return "", errors.Wrap(err, "reg.GetAsset() failed")
	}

	if isDirect {
		// when isDirect, explicit permission is required
		thisIdentity, err := ctx.GetRegistry().GetThisIdentityResolved()
		if err != nil {
			return "", errors.Wrap(err, "reg.GetThisIdentityResolved() failed")
		}

		kmpg, err := kompiguard.New()
		if err != nil {
			return "", errors.Wrap(err, "kompiguard.New() failed")
		}

		granted, reason, err := kmpg.EnforceAsset(asset, thisIdentity, "delete_direct")
		if err != nil {
			return "", errors.Wrap(err, "kmpg.EnforceCustom()")
		}

		if !granted {
			return "", ErrorForbidden(reason)
		}
	}

	if !isDirect {
		_, err = ctx.GetConfiguration().BusinessExecutor.Execute(ctx, BeforeDelete, nil, asset)
		if err != nil {
			return "", errors.Wrap(err, "bexec.Execute(), stage: BeforeDelete failed")
		}
	}

	if err := ctx.GetRegistry().DeleteAsset(asset); err != nil {
		return "", errors.Wrap(err, "reg.DeleteAsset() failed")
	}

	output := rmap.NewFromMap(map[string]interface{}{
		"ok":            true,
		OutputResultKey: true,
	})

	return string(output.Bytes()), nil
}

func assetHistoryFrontend(ctx ContextInterface) (string, error) {
	name, err := ctx.ParamString(NameParam)
	if err != nil {
		return "", err
	}

	id, err := ctx.ParamString(IdParam)
	if err != nil {
		return "", err
	}

	reg := ctx.Get(RegistryKey).(*Registry)

	asset, err := reg.GetAsset(name, id, false, true)
	if err != nil {
		return "", errors.Wrap(err, "reg.GetAsset() failed")
	}

	if err := enforceAssetAccess(reg, asset, "get_history"); err != nil {
		return "", err
	}

	hItems, err := reg.GetAssetHistory(asset)
	if err != nil {
		return "", errors.Wrap(err, "reg.GetAssetHistory() failed")
	}

	output := rmap.NewFromMap(map[string]interface{}{
		"result": hItems,
	})

	return string(output.Bytes()), nil
}
