package engine

import (
	"fmt"
	"strconv"

	. "github.com/KompiTech/fabric-cc-core/v2/pkg/konst"
	"github.com/KompiTech/rmap"
	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/pkg/errors"
)

// makes sure that super user role exists, if not, creates it
func ensureSuperUserRole(reg *Registry) error {
	superuserRole, err := reg.GetAsset(RoleAssetName, SuperuserRoleUUID, false, false)
	if err != nil {
		return errors.Wrap(err, "reg.GetAsset() failed")
	}

	if superuserRole.IsEmpty() {
		// superuser role does not exist, it must be created
		superuserRole, err = reg.MakeAsset(RoleAssetName, SuperuserRoleUUID, -1)
		if err != nil {
			return errors.Wrap(err, "reg.MakeAsset() failed")
		}

		superuserRole.Mapa["name"] = "Superuser"

		if err := reg.PutAsset(superuserRole, true); err != nil {
			return errors.Wrap(err, "reg.PutAsset() failed")
		}
	}

	return nil
}

// makes sure that existingIdentity has SU role granted
// !!! DOES NOT PERFORM SAVING OF THE ASSET !!!
// returns boolean, if it modified the asset
func ensureSuperUserGrant(existingIdentity rmap.Rmap) (bool, error) {
	if existingIdentity.Exists(RolesKey) {
		// check if SU role is already granted
		// roles are unresolved
		rolesIter, err := existingIdentity.GetIterable(RolesKey)
		if err != nil {
			return false, errors.Wrap(err, "existingIdentity.GetIterable() failed")
		}

		grantNeeded := true

		for index, _ := range rolesIter {
			// check if SU role is already granted
			jptr := "/" + RolesKey + "/" + strconv.Itoa(index)

			// use GetJPtrString here to get Nth role element -> does conversion from interface{} for us and handles all errors
			roleUUID, err := existingIdentity.GetJPtrString(jptr)
			if err != nil {
				// decorate error msg with jptr used for extra info
				return false, errors.Wrapf(err, "existingIdentity.GetJPtrString(%s) failed", jptr)
			}

			if roleUUID == SuperuserRoleUUID {
				// grant is not needed
				grantNeeded = false
				break
			}
		}

		if grantNeeded {
			// roles iterable does not contain su role, append to it and save to asset
			rolesIter = append(rolesIter, SuperuserRoleUUID)
			existingIdentity.Mapa[RolesKey] = rolesIter

			return true, nil // it was modified
		} else {
			return false, nil // no modification
		}
	} else {
		// if no roles key is present, it must be created with one element - the magic role UUID
		existingIdentity.Mapa[RolesKey] = []string{SuperuserRoleUUID}

		return true, nil // it was modified
	}
}

// 1. check, if key with init manager configuration exists
// 2. if it does, check if its value matches myFingerprint
// 3. if it does, create SU role with magic UUID if it does not exists
// 4. if fp was matched, delete its object from State
// returns true if SU obj existed and it matches myFingerprint
func processInitManagerObj(ctx ContextInterface, myFingerprint string) (bool, error) {
	reg := ctx.Get(RegistryKey).(*Registry)

	initSuperBootstrapObjBytes, err := ctx.Stub().GetState(InitSuperuserStateKey)
	if err != nil {
		return false, errors.Wrap(err, "ctx.Stub().GetState() failed")
	}

	if len(initSuperBootstrapObjBytes) > 0 {
		// key with superUser exists, load object from bytes
		initSUobj, err := rmap.NewFromBytes(initSuperBootstrapObjBytes)
		if err != nil {
			return false, errors.Wrap(err, "rmap.NewFromBytes() failed")
		}

		// get fingerprint value from stored obj
		initSUfp, err := initSUobj.GetString(InitSuperuserKey)
		if err != nil {
			return false, errors.Wrap(err, "initSUobj.GetString() failed")
		}

		if myFingerprint == initSUfp {
			// bootstrap key matches current FP, superuser will be granted
			// make sure that SU role exists, so it can be granted
			// this already saves the role in state
			if err := ensureSuperUserRole(reg); err != nil {
				return false, errors.Wrap(err, "ensureSuperUserRole() failed")
			}

			// delete bootstrap key in all cases
			// next calls to identityAddMe will skip this if branch to save resources
			if err := ctx.Stub().DelState(InitSuperuserStateKey); err != nil {
				return false, errors.Wrap(err, "ctx.Stub().DelState() failed")
			}

			return true, nil // fp matched, no error
		}
	}

	return false, nil // fp not matched, no error
}

func identityAddMeFrontend(ctx ContextInterface) (string, error) {
	reg := ctx.Get(RegistryKey).(*Registry)

	// determine fingerprint from cert
	myFingerprint, err := GetMyFingerprint(ctx)
	if err != nil {
		return "", err
	}

	// try getting existing identity, do not fail if it doesnt exist
	existingIdentity, err := reg.GetAsset(IdentityAssetName, myFingerprint, false, false)
	if err != nil {
		return "", errors.Wrap(err, "reg.GetAsset() failed")
	}

	// check, if this identity needs SU granted
	suGrantIsNeeded, err := processInitManagerObj(ctx, myFingerprint)
	if err != nil {
		return "", errors.Wrap(err, "processInitManagerObj() failed")
	}

	if !existingIdentity.IsEmpty() {
		// identity already exists
		// but if SU needs granting, it must be saved here
		if suGrantIsNeeded {
			persistIsNeeded, err := ensureSuperUserGrant(existingIdentity)
			if err != nil {
				return "", errors.Wrap(err, "ensureSuperUserGrant() failed")
			}

			if persistIsNeeded {
				// only save modified asset to state if it is actually needed (= modification occurred)
				if err := reg.PutAsset(existingIdentity, false); err != nil {
					return "", errors.Wrap(err, "reg.PutAsset() failed")
				}
			}
		}
		// return identity asset
		return string(existingIdentity.WrappedResultBytes()), nil
	}

	if ctx.GetConfiguration().PreviousIDFunc != nil {
		// only migrate if previous func is defined
		oldIdentity, err := getOldIdentity(ctx)
		if err != nil {
			return "", err
		}

		if !oldIdentity.IsEmpty() {
			// oldIdentity can be migrated
			newIdentity, err := migrateOldIdentity(ctx, oldIdentity)
			if err != nil {
				return "", err
			}

			response := rmap.NewFromMap(map[string]interface{}{
				"migrated": true,
				"result":   newIdentity.Mapa,
			})

			// return newIdentity with info about migration occuring
			return response.String(), nil
		}
	}

	// at this point, new identity asset must be created (migration did not occur)
	identityAsset, err := reg.MakeAsset(IdentityAssetName, myFingerprint, -1)
	if err != nil {
		return "", errors.Wrap(err, "reg.MakeAsset() failed")
	}

	// new identities must be enabled
	identityAsset.Mapa[IsEnabledKey] = true

	if suGrantIsNeeded {
		// grant magic role if this is init superuser
		identityAsset.Mapa[RolesKey] = []string{SuperuserRoleUUID}
	}

	inputBytes, err := ctx.ParamBytes("input")
	if err != nil {
		return "", err
	}

	var input rmap.Rmap
	if len(inputBytes) == 0 {
		input = rmap.NewEmpty()
	} else {
		var err error
		input, err = rmap.NewFromBytes(inputBytes)
		if err != nil {
			return "", errors.Wrap(err, "rmap.NewFromBytes() failed")
		}
	}

	// execute business logic for stage BeforeCreate on identity
	bexec := ctx.GetConfiguration().BusinessExecutor

	identityAsset, err = bexec.ExecuteCustom(ctx, BeforeCreate, identityAsset, &input, identityAsset)
	if err != nil {
		return "", errors.Wrap(err, "bexec.Execute(), stage: BeforeCreate failed")
	}

	if err := reg.PutAsset(identityAsset, true); err != nil {
		return "", errors.Wrap(err, "reg.PutAsset() failed")
	}

	return string(identityAsset.WrappedResultBytes()), nil
}

// deprecated, use assetGet(identity, ...)
func identityGetFrontend(ctx ContextInterface) (string, error) {
	fp, err := ctx.ParamString(FingerprintParam)
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

	return assetGetBackend(ctx, IdentityAssetName, fp, resolve, data, false)
}

func identityMeFrontend(ctx ContextInterface) (string, error) {
	// get fp of user calling this (this user)
	myFP, err := GetMyFingerprint(ctx)
	if err != nil {
		return "", err
	}

	reg := ctx.Get(RegistryKey).(*Registry)

	// fetch identity of this user
	resolve, err := ctx.ParamBool(ResolveParam)
	if err != nil {
		return "", err
	}

	identityAsset, err := reg.GetAsset(IdentityAssetName, myFP, resolve, false)
	if err != nil {
		return "", errors.Wrap(err, "reg.GetAsset() failed")
	}

	if identityAsset.IsEmpty() {
		if ctx.GetConfiguration().PreviousIDFunc != nil {
			// if previous func is defined, attempt to get migration possibility
			oldIdentity, err := getOldIdentity(ctx)
			if err != nil {
				return "", err
			}

			if !oldIdentity.IsEmpty() {
				oldFP, err := oldIdentity.GetString("fingerprint")
				if err != nil {
					return "", err
				}

				// return migration info if old identity was found
				return string(rmap.NewFromMap(map[string]interface{}{"fingerprint": oldFP, "can_migrate": true}).Bytes()), nil
			}
		}

		// identity asset does not exist
		// return particular error message, so frontend can redirect to registration of new user
		return "", ErrorBadRequest(fmt.Sprintf("cannot get identity for fingerprint: %s. Did you call identityAddMe?", myFP))
	}

	// execute AfterGet blogic on identity asset
	identityAsset, err = ctx.GetConfiguration().BusinessExecutor.Execute(ctx, AfterGet, nil, identityAsset)
	if err != nil {
		return "", errors.Wrap(err, "bexec.Execute(), stage: AfterGet failed")
	}

	return string(identityAsset.WrappedResultBytes()), nil
}

// deprecated, use assetUpdate(identity, ...)
func identityUpdateFrontend(ctx ContextInterface) (string, error) {
	fp, err := ctx.ParamString(FingerprintParam)
	if err != nil {
		return "", err
	}

	patch, err := ctx.ParamString(PatchParam)
	if err != nil {
		return "", err
	}

	return assetUpdateBackend(ctx, IdentityAssetName, fp, patch, false)
}

// extra validations required for identity asset regarding SUs
func identityExtraValidate(reg *Registry, identityAssetPre, identityAssetPost, thisIdentity rmap.Rmap) error {
	// particular validations needs to be done:
	// - only superUser can grant superUser to somebody else
	// - only superUser can remove superUser from somebody else (also by disabling the identity)
	// - last superUser cannot remove his superUser (also by disabling the identity)
	// - nobody can disable their own identity
	isEnabled, err := identityAssetPost.GetBool(IsEnabledKey)
	if err != nil {
		return errors.Wrap(err, "identityAssetPost.GetBool() failed")
	}

	if !isEnabled {
		thisID, err := AssetGetID(thisIdentity)
		if err != nil {
			return errors.Wrap(err, "konst.AssetGetID(thisIdentity) failed")
		}

		identityID, err := AssetGetID(identityAssetPost)
		if err != nil {
			return errors.Wrap(err, "konst.AssetGetID(identityAssetPost) failed")
		}

		// client cannot disable its own identity
		if thisID == identityID {
			return ErrorBadRequest("unable to set is_enabled to false on own identity")
		}
	}

	rolesDidExist := identityAssetPre.Exists(RolesKey)
	rolesExist := identityAssetPost.Exists(RolesKey)
	var revokingSU, grantingSU bool
	if !rolesDidExist && !rolesExist {
		// no granted roles are present and nothing is changed
		revokingSU = false
		grantingSU = false
	} else if !rolesDidExist && rolesExist {
		// no roles existed and some were granted
		revokingSU = false
		grantingSU, err = identityAssetPost.Contains(RolesKey, SuperuserRoleUUID)
		if err != nil {
			return errors.Wrap(err, "identityAssetPost.Contains() failed")
		}
	} else if rolesDidExist && !rolesExist {
		// some roles existed and all were removed
		grantingSU = false
		revokingSU, err = identityAssetPre.Contains(RolesKey, SuperuserRoleUUID)
		if err != nil {
			return errors.Wrap(err, "identityAssetPre.Contains() failed")
		}
	} else {
		// some roles existed and they still exist
		suWasGranted, err := identityAssetPre.Contains(RolesKey, SuperuserRoleUUID)
		if err != nil {
			return errors.Wrap(err, "identityAssetPre.Contains() failed")
		}

		suIsGranted, err := identityAssetPost.Contains(RolesKey, SuperuserRoleUUID)
		if err != nil {
			return errors.Wrap(err, "identityAssetPost.Contains() failed")
		}

		grantingSU = !suWasGranted && suIsGranted
		revokingSU = suWasGranted && !suIsGranted
		// SU is granted if role is granted and identity asset is enabled
		grantingSU = isEnabled && !suWasGranted && suIsGranted

		// SU is revoked if role is removed on enabled asset OR if role with SU present is disabled
		revokingSU = (isEnabled && suWasGranted && !suIsGranted) || (!isEnabled && suIsGranted)
	}

	if grantingSU || revokingSU {
		// when revoking or granting SU, current role MUST be SU
		thisIsSU, err := thisIdentity.ContainsJPtrKV(RolesJPtr, "/"+AssetIdKey, SuperuserRoleUUID)
		if err != nil {
			return errors.Wrap(err, "thisIdentity.ContainsJPtr() failed")
		}

		if !thisIsSU {
			return ErrorForbidden("to manage SuperUser role, you must have it granted")
		}

		if revokingSU {
			SUs, _, err := reg.QueryAssets(IdentityAssetName, rmap.NewFromMap(getSUQuery()), "", false, false, PageSize)
			if err != nil {
				return errors.Wrap(err, "reg.QueryAssets() failed")
			}

			if len(SUs) == 1 {
				return ErrorBadRequest("unable to remove last superuser role")
			}
		}
	}

	return nil
}

// deprecated, use assetQuery(identity, ...)
func identityQueryFrontend(ctx ContextInterface) (string, error) {
	query, err := ctx.ParamString(QueryParam)
	if err != nil {
		return "", err
	}

	resolve, err := ctx.ParamBool(ResolveParam)
	if err != nil {
		return "", err
	}

	return assetQueryBackend(ctx, IdentityAssetName, query, resolve, PageSize, false)
}

// deprecated, use assetCreateDirect(identity, ...)
func identityCreateDirectFrontend(ctx ContextInterface) (string, error) {
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

	return assetCreateBackend(ctx, now, IdentityAssetName, data, 1, id, true)
}

// deprecated, use assetUpdateDirect(identity, ...)
func identityUpdateDirectFrontend(ctx ContextInterface) (string, error) {
	id, err := ctx.ParamString(IdParam)
	if err != nil {
		return "", err
	}

	patch, err := ctx.ParamString(PatchParam)
	if err != nil {
		return "", err
	}

	return assetUpdateBackend(ctx, IdentityAssetName, id, patch, true)
}

// attempts to migrate identity identified by eng.PreviousIDFunc to a new one identified by eng.CurrentIDFunc
// return Rmap with new identity
func migrateOldIdentity(ctx ContextInterface, oldIdentity rmap.Rmap) (rmap.Rmap, error) {
	eng := ctx.GetConfiguration()
	reg := ctx.Get("registry").(*Registry)
	null := rmap.Rmap{}

	if eng.PreviousIDFunc == nil {
		// no previous function is defined, did not migrate
		return null, nil
	}

	thisCert, err := cid.GetX509Certificate(ctx.Stub())
	if err != nil {
		return null, errors.Wrap(err, "cid.GetX509Certificate() failed")
	}

	// get fp using current func
	newFP, err := eng.CurrentIDFunc(thisCert)
	if err != nil {
		return null, err
	}

	if oldIdentity.IsEmpty() {
		// nothing was migrated
		return null, nil
	}

	// create new identity asset by copying old and using new fingerprint
	newIdentity := oldIdentity.Copy()
	newIdentity.Mapa["fingerprint"] = newFP
	if err := reg.PutAsset(newIdentity, true); err != nil {
		return null, err
	}

	// disable old identity
	oldIdentity.Mapa["is_enabled"] = false
	if err := reg.PutAsset(oldIdentity, false); err != nil {
		return null, err
	}

	return newIdentity, nil // finished migration
}

func getOldIdentity(ctx ContextInterface) (rmap.Rmap, error) {
	eng := ctx.GetConfiguration()
	reg := ctx.Get("registry").(*Registry)
	null := rmap.Rmap{}
	if eng.PreviousIDFunc == nil {
		// no previous function is defined, did not migrate
		return null, nil
	}

	thisCert, err := cid.GetX509Certificate(ctx.Stub())
	if err != nil {
		return null, errors.Wrap(err, "cid.GetX509Certificate() failed")
	}

	// get fp using previous func
	oldFP, err := (*eng.PreviousIDFunc)(thisCert)
	if err != nil {
		return null, err
	}

	// get old identity asset
	oldIdentity, err := reg.GetAsset(IdentityAssetName, oldFP, false, false)
	if err != nil {
		return null, err
	}

	return oldIdentity, nil
}
