package kompiguard

import (
	"fmt"
	"strings"

	. "github.com/KompiTech/fabric-cc-core/v2/pkg/konst"
	"github.com/KompiTech/rmap"
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/pkg/errors"
)

type KompiGuard struct {
	enforcer *casbin.Enforcer
}

func New() (KompiGuard, error) {
	m, err := model.NewModelFromString(CasbinModel)
	if err != nil {
		return KompiGuard{}, errors.Wrap(err, "model.NewModelFromString() failed")
	}

	enf, err := casbin.NewEnforcer(m)
	if err != nil {
		return KompiGuard{}, errors.Wrap(err, "casbin.NewEnforcer() failed")
	}

	return KompiGuard{enf}, nil
}

// LoadRoles loads roles from identity asset into enforcer
// identityAsset must be resolved
func (k KompiGuard) LoadRoles(thisIdentity rmap.Rmap) error {
	isEnabled, err := thisIdentity.GetBool(IsEnabledKey)
	if err != nil {
		return errors.Wrap(err, "thisIdentity.GetBool() failed")
	}

	if !isEnabled {
		fp, err := AssetGetID(thisIdentity)
		if err != nil {
			return errors.Wrap(err, "AssetGetID(thisIdentity) failed")
		}

		return fmt.Errorf("current identity: %s is not enabled", fp)
	}

	if !thisIdentity.Exists(RolesKey) {
		// identity does not define any roles, finished
		return nil
	}

	roleList, err := thisIdentity.GetIterable(RolesKey)
	if err != nil {
		return errors.Wrap(err, "thisIdentity.GetJPtrIterable() failed")
	}

	subject, err := AssetGetID(thisIdentity)
	if err != nil {
		return errors.Wrap(err, "AssetGetID(thisIdentity) failed")
	}

	for _, roleI := range roleList {
		// for each role assigned to user:
		// load all grants from each role
		// load all roles
		role, err := rmap.NewFromInterface(roleI)
		if err != nil {
			return errors.Wrap(err, "NewFromInterface() failed: did you resolve the role asset?")
		}

		roleUUID, err := AssetGetID(role)
		if err != nil {
			return errors.Wrap(err, "role.GetID() failed")
		}

		// add mapping of user-role to enforcer
		_, err = k.enforcer.AddRoleForUser(subject, roleUUID)
		if err != nil {
			return errors.Wrap(err, "k.enforcer.AddRoleForUser() failed")
		}

		if !role.Exists(GrantsKey) {
			// role does not have grants key, finished
			continue
		}

		grantsList, err := role.GetIterable(GrantsKey)
		if err != nil {
			return errors.Wrap(err, "role.GetJPtrIterable() failed")
		}

		for _, grantI := range grantsList {
			grant, err := rmap.NewFromInterface(grantI)
			if err != nil {
				return errors.Wrap(err, "NewFromInterface() failed")
			}

			object, err := grant.GetString(ObjectKey)
			if err != nil {
				return errors.Wrap(err, "grant.GetJPtrString() failed")
			}

			action, err := grant.GetString(ActionKey)
			if err != nil {
				return errors.Wrap(err, "grant.GetJPtrString() failed")
			}

			// add every grant to Enforcer
			_, err = k.enforcer.AddPermissionForUser(roleUUID, object, action, "allow")
			if err != nil {
				return errors.Wrap(err, "k.enforcer.AddPermissionForUser() failed")
			}
		}
	}

	return nil
}

// Load overrides from asset into enforcer
func (k KompiGuard) loadOverrides(asset rmap.Rmap) error {
	if !asset.Exists(OverridesKey) {
		// no overrides defined on asset, finished
		return nil
	}

	object, err := AssetGetCasbinObject(asset)
	if err != nil {
		return errors.Wrap(err, "asset.GetCasbinObject() failed")
	}

	overridesList, err := asset.GetIterable(OverridesKey)
	if err != nil {
		return errors.Wrap(err, "asset.GetJPtrIterable() failed")
	}

	for _, overrideI := range overridesList {
		override, err := rmap.NewFromInterface(overrideI)
		if err != nil {
			return errors.Wrap(err, "NewFromInterface() failed")
		}

		subject, err := override.GetString(SubjectKey)
		if err != nil {
			return errors.Wrap(err, "override.GetJPtrString() failed")
		}
		subject = strings.ToLower(subject)

		action, err := override.GetString(ActionKey)
		if err != nil {
			return errors.Wrap(err, "override.GetJPtrString() failed")
		}
		action = strings.ToLower(action)

		effect, err := override.GetString(EffectKey)
		if err != nil {
			return errors.Wrap(err, "override.GetJPtrString() failed")
		}
		effect = strings.ToLower(effect)

		if _, err := k.enforcer.AddPermissionForUser(subject, object, action, effect); err != nil {
			return errors.Wrap(err, "k.enforcer.AddPermissionForUser() failed")
		}
	}

	return nil
}

// EnforceAsset checks if action is allowed on asset for fingerprint
// object is derived from asset parameter, it must be a valid asset
// overrides are loaded from asset, if present
func (k KompiGuard) EnforceAsset(asset, thisIdentity rmap.Rmap, action string) (bool, string, error) {
	object, err := AssetGetCasbinObject(asset)
	if err != nil {
		return false, "", errors.Wrap(err, "AssetGetCasbinObject(asset) failed")
	}

	subject, err := AssetGetID(thisIdentity)
	if err != nil {
		return false, "", errors.Wrap(err, "thisIdentity.GetID() failed")
	}

	if err := k.LoadRoles(thisIdentity); err != nil {
		return false, "", errors.Wrap(err, "k.LoadRoles() failed")
	}

	return k.EnforceCustom(object, subject, action, &asset)
}

// EnforceCustom checks if enforcer allows action on object for subject
// if asset argument is present, overrides are loaded from it
// returns (granted, reason, error)
// granted - true or false if operation was granted
// reason - if granted == false && err != nil then it contains which permission is required for action to be granted
// error - some other error has occurred
func (k KompiGuard) EnforceCustom(object, subject, action string, asset *rmap.Rmap) (bool, string, error) {
	action = strings.ToLower(action)

	if asset != nil {
		if err := k.loadOverrides(*asset); err != nil {
			return false, "", errors.Wrap(err, "k.loadOverrides() failed")
		}
	}

	// try enforcing with mixed case object
	granted, err := k.enforcer.Enforce(subject, object, action)
	if err != nil {
		return false, "", errors.Wrap(err, "k.enforcer.Enforce() failed")
	}

	if !granted {
		// try enforcing with lowercase object for backwards compatibility with existing objects
		grantedLower, err := k.enforcer.Enforce(subject, strings.ToLower(object), action)
		if err != nil {
			return false, "", errors.Wrap(err, "k.enforcer.Enforce() failed")
		}

		if !grantedLower {
			// still denied, return error with original mixed case
			return false, fmt.Sprintf("permission denied, sub: %s, obj: %s, act: %s", subject, object, action), nil
		}
	}

	return true, "", nil
}

// FilterAssets iterates through slice of assets on input and checks if enforcer grants action for fingerprint
// if action is granted, asset is present in output without changes
// if action is not granted, asset is replaced by censored object
func (k KompiGuard) FilterAssets(input []rmap.Rmap, identityAsset rmap.Rmap, action string) ([]rmap.Rmap, error) {
	if err := k.LoadRoles(identityAsset); err != nil {
		return nil, errors.Wrap(err, "k.LoadRoles() failed")
	}

	subject, err := AssetGetID(identityAsset)
	if err != nil {
		return nil, errors.Wrap(err, "AssetGetID() failed")
	}

	output := make([]rmap.Rmap, 0, len(input))
	for _, asset := range input {
		object, err := AssetGetCasbinObject(asset)
		if err != nil {
			return nil, errors.Wrap(err, "asset.GetCasbinObject() failed")
		}

		docType, err := AssetGetDocType(asset)
		if err != nil {
			return nil, errors.Wrap(err, "asset.GetDocType() failed")
		}

		var granted bool
		var reason string
		insertee := asset

		if docType == IdentityAssetName {
			assetID, err := AssetGetID(asset)
			if err != nil {
				return nil, errors.Wrap(err, "asset.GetID() failed")
			}

			if assetID == subject {
				// if filtering own identity asset, allow
				granted = true
			}
		}

		if !granted {
			// if not previously granted by own identity, evaluate standard access control
			granted, reason, err = k.EnforceCustom(object, subject, ReadAction, &asset)
			if err != nil {
				return nil, errors.Wrap(err, "k.EnforceCustom() failed")
			}
		}

		if !granted {
			// replace with just id and error message if not allowed to read
			idKey, err := AssetGetIDKey(asset)
			if err != nil {
				return nil, errors.Wrap(err, "asset.GetIDKey() failed")
			}

			id, err := AssetGetID(asset)
			if err != nil {
				return nil, errors.Wrap(err, "asset.GetID() failed")
			}

			insertee = rmap.NewFromMap(map[string]interface{}{
				idKey:   id,
				"error": reason,
			})
		}

		output = append(output, insertee)
	}

	return output, nil
}
