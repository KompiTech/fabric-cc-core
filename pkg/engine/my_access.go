package engine

import (
	"fmt"
	"sort"

	"github.com/KompiTech/fabric-cc-core/v2/pkg/kompiguard"
	"github.com/KompiTech/fabric-cc-core/v2/pkg/konst"
	"github.com/KompiTech/rmap"
	"github.com/pkg/errors"
)

// myAccessFunc is implementation of built-in myAccess function to get all accessible assets and functions
/*
{
	"assets_read": ["incident", "k_request"],
	"assets_create": ["incident"],
	"assets_update": ["incident"],
	"assets_delete": ["incident"],
	"functions_query": ["dashboardFce1", "dashboardFce2"],
	"functions_invoke": ["generateBillingReport"],
	"custom_grants": {
		"view_sensitive": ["/user/*"]
	}
}
*/
var myAccessFunc = func(ctx ContextInterface, input rmap.Rmap, output rmap.Rmap) (rmap.Rmap, error) {
	null := rmap.Rmap{}
	reg := ctx.Get("registry").(*Registry)

	thisIdentity, err := reg.GetThisIdentityResolved()
	if err != nil {
		return null, errors.Wrap(err, "reg.GetThisIdentityResolved() failed")
	}

	return accessFuncImpl(ctx, thisIdentity, output)
}

// identityAccessFunc does the same as myAccessFunc but for specified Identity
var identityAccessFunc = func(ctx ContextInterface, input rmap.Rmap, output rmap.Rmap) (rmap.Rmap, error) {
	null := rmap.Rmap{}
	reg := ctx.Get("registry").(*Registry)

	identityFP, err := input.GetString("identity")
	if err != nil {
		return null, ErrorBadRequest("identity' key is missing")

	}

	identityAsset, err := reg.GetAsset(konst.IdentityAssetName, identityFP, true, true)
	if err != nil {
		return null, ErrorBadRequest(fmt.Sprintf("cannot get identity for fingerprint: %s", identityFP))
	}

	return accessFuncImpl(ctx, identityAsset, output)
}

// getCustomGrants collects all unique grants for identity that are not standard asset and function grants (CRUD + function execute)
// returns map with key being grant name and value being list of all objects that have this grant
func getCustomGrants(identity rmap.Rmap) (rmap.Rmap, error) {
	null := rmap.Rmap{}
	stdGrants := map[string]struct{}{ // these grant names are processed elsewhere, so they will be skipped
		"create":  {},
		"read":    {},
		"update":  {},
		"delete":  {},
		"execute": {},
	}

	output := rmap.NewEmpty() // first level: grant name -> map with objects (will be converted to list before returning)

	if !identity.Exists("roles") {
		return output, nil // no roles, nothing can be granted
	}

	roles, err := identity.GetIterable("roles")
	if err != nil {
		return null, err
	}

	for _, roleI := range roles {
		role, err := rmap.NewFromInterface(roleI)
		if err != nil {
			return null, err
		}

		if !role.Exists("grants") {
			continue
		}

		grants, err := role.GetIterable("grants")
		if err != nil {
			return null, err
		}

		for _, grantI := range grants {
			grant, err := rmap.NewFromInterface(grantI)
			if err != nil {
				return null, err
			}

			action, err := grant.GetString("action")
			if err != nil {
				return null, err
			}

			_, isStd := stdGrants[action]
			if isStd {
				continue // skip standard actions
			}

			object, err := grant.GetString("object")
			if err != nil {
				return null, err
			}

			_, outObjExists := output.Mapa[action]
			if !outObjExists {
				output.Mapa[action] = map[string]interface{}{}
			}

			subMap, err := output.GetRmap(action)
			if err != nil {
				return null, err
			}

			subMap.Mapa[object] = ""
		}
	}

	// convert second level maps to sorted list of objects for each action
	for k, actionMapI := range output.Mapa {
		actionMap, err := rmap.NewFromInterface(actionMapI)
		if err != nil {
			return null, err
		}

		keysSlice := actionMap.KeysSliceString()
		sort.Strings(keysSlice)

		output.Mapa[k] = keysSlice
	}

	return output, nil
}

// isIdentitySU checks if identity is SU
func isIdentitySU(identity rmap.Rmap) (bool, error) {
	if !identity.Exists("roles") {
		return false, nil // no roles, cannot be SU
	}

	roles, err := identity.GetIterableRmap("roles")
	if err != nil {
		return false, err
	}

	for _, role := range roles {
		uuid, err := role.GetString("uuid")
		if err != nil {
			return false, err
		}

		if uuid == konst.SuperuserRoleUUID {
			return true, nil
		}
	}

	return false, nil
}

// getStaticSuperUserGrants adds grants to superUser which would not be discovered
func getStaticSuperUserGrants() rmap.Rmap {
	return rmap.NewFromMap(map[string]interface{}{
		"view_sensitive": []string{"/user/*"},
	})
}

func accessFuncImpl(ctx ContextInterface, identity rmap.Rmap, output rmap.Rmap) (rmap.Rmap, error) {
	null := rmap.Rmap{}
	kmpg := kompiguard.KompiGuard{}

	isSU, err := isIdentitySU(identity)
	if err != nil {
		return null, err
	}

	thisFP, err := identity.GetString("fingerprint")
	if err != nil {
		return null, errors.Wrap(err, "could not get fingerprint from Identity")
	}

	reg := ctx.Get("registry").(*Registry)

	allAssets, err := reg.ListItems()
	if err != nil {
		return null, errors.Wrap(err, "reg.ListItems() failed")
	}

	allFunctions := ctx.GetConfiguration().FunctionExecutor.List()

	if !isSU {
		// only load kompiguard if role is not SU
		// because SU can do anything, so no point in asking kompiguard
		kmpg, err = kompiguard.New()
		if err != nil {
			return null, errors.Wrap(err, "kompiguard.New() failed")
		}

		if err := kmpg.LoadRoles(identity); err != nil {
			return null, errors.Wrap(err, "kmpg.LoadRoles() failed")
		}
	}

	output = rmap.NewEmpty()
	var granted bool

	// functions
	for _, typ := range []string{"query", "invoke"} { // funcs have separate query and invoke variant
		for _, funcName := range allFunctions { // iterate through all available CC functions
			obj := fmt.Sprintf("/function/%s/%s", typ, funcName)

			if isSU {
				granted = true
			} else {
				granted, _, err = kmpg.EnforceCustom(obj, thisFP, "execute", nil)
				if err != nil {
					return null, err
				}
			}

			if granted {
				outputKey := "functions_" + typ
				if !output.Exists(outputKey) {
					output.Mapa[outputKey] = []string{}
				}
				outSlice := output.Mapa[outputKey].([]string)
				outSlice = append(outSlice, funcName)
				output.Mapa[outputKey] = outSlice
			}
		}
	}

	// assets
	for _, action := range []string{"create", "read", "update", "delete"} {
		for _, assetName := range allAssets {
			obj := "/" + assetName + "/*"

			if isSU {
				granted = true
			} else {
				granted, _, err = kmpg.EnforceCustom(obj, thisFP, action, nil)
				if err != nil {
					return null, err
				}
			}

			if granted {
				outputKey := "assets_" + action
				if !output.Exists(outputKey) {
					output.Mapa[outputKey] = []string{}
				}
				outSlice := output.Mapa[outputKey].([]string)
				outSlice = append(outSlice, assetName)
				output.Mapa[outputKey] = outSlice
			}
		}
	}

	//sort
	for k, vI := range output.Mapa {
		v := vI.([]string)
		sort.Strings(v)
		output.Mapa[k] = v
	}

	var customGrants rmap.Rmap

	if !isSU {
		customGrants, err = getCustomGrants(identity)
		if err != nil {
			return null, err
		}
	} else {
		customGrants = getStaticSuperUserGrants()
	}

	output.Mapa["custom_grants"] = customGrants.Mapa

	return output, nil
}
