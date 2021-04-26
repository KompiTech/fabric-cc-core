package engine

import (
	"fmt"
	"strings"
)

// route calls appropriate frontend method and returns result, of returns error if name is not valid
func route(ctx ContextInterface) (string, error) {
	fcn, args := ctx.GetStub().GetFunctionAndParameters()
	uerr := fmt.Errorf("invalid function %s passed with args %v", fcn, args)
	var ret string
	var err error

	// when prefix is matched, it is also trimmed from fcn
	matchPrefix := func(prefix string) bool {
		if strings.HasPrefix(fcn, prefix) {
			fcn = strings.TrimPrefix(fcn, prefix)
			return true
		}
		return false
	}

	// match prefix with case insensitive first letter
	matchPrefixI := func(lowerCasePrefix string) bool {
		if matchPrefix(lowerCasePrefix) {
			return true
		}

		upperCasePrefix := strings.ToUpper(string(lowerCasePrefix[0])) + lowerCasePrefix[1:]
		if matchPrefix(upperCasePrefix) {
			return true
		}

		return false
	}

	isDirect := func() bool {
		return matchPrefix("Direct")
	}

	isEmpty := func() bool {
		return len(fcn) == 0
	}

	if matchPrefixI("asset") {
		if matchPrefix("Create") {
			if isDirect() && isEmpty() {
				ret, err = assetCreateDirectFrontend(ctx)
			} else if isEmpty() {
				ret, err = assetCreateFrontend(ctx)
			} else {
				err = uerr
			}
		} else if matchPrefix("Delete") {
			if isDirect() && isEmpty() {
				ret, err = assetDeleteDirectFrontend(ctx)
			} else if isEmpty() {
				ret, err = assetDeleteFrontend(ctx)
			} else {
				err = uerr
			}
		} else if matchPrefix("Get") {
			if isDirect() && isEmpty() {
				ret, err = assetGetDirectFrontend(ctx)
			} else if isEmpty() {
				ret, err = assetGetFrontend(ctx)
			} else {
				err = uerr
			}
		} else if matchPrefix("History") && isEmpty() {
			ret, err = assetHistoryFrontend(ctx)
		} else if matchPrefix("Migrate") && isEmpty() {
			ret, err = assetMigrateFrontend(ctx)
		} else if matchPrefix("Update") {
			if isDirect() && isEmpty() {
				ret, err = assetUpdateDirectFrontend(ctx)
			} else if isEmpty() {
				ret, err = assetUpdateFrontend(ctx)
			} else {
				err = uerr
			}
		} else if matchPrefix("Query") {
			if isDirect() && isEmpty() {
				ret, err = assetQueryDirectFrontend(ctx)
			} else if isEmpty() {
				ret, err = assetQueryFrontend(ctx)
			} else {
				err = uerr
			}
		} else {
			err = uerr
		}
	} else if matchPrefixI("changelog") {
		if matchPrefix("Get") && isEmpty() {
			ret, err = changelogGetFrontend(ctx)
		} else if matchPrefix("List") && isEmpty() {
			ret, err = changelogListFrontend(ctx)
		} else {
			err = uerr
		}
	} else if matchPrefixI("function") {
		if matchPrefix("Invoke") && isEmpty() {
			ret, err = functionInvokeFrontend(ctx)
		} else if matchPrefix("Query") && isEmpty() {
			ret, err = functionQueryFrontend(ctx)
		} else {
			err = uerr
		}
	} else if matchPrefixI("identity") {
		if matchPrefix("AddMe") && isEmpty() {
			ret, err = identityAddMeFrontend(ctx)
		} else if matchPrefix("Create") && isDirect() && isEmpty() {
			ret, err = identityCreateDirectFrontend(ctx)
		} else if matchPrefix("Get") && isEmpty() {
			ret, err = identityGetFrontend(ctx)
		} else if matchPrefix("Me") && isEmpty() {
			ret, err = identityMeFrontend(ctx)
		} else if matchPrefix("Update") {
			if isDirect() && isEmpty() {
				ret, err = identityUpdateDirectFrontend(ctx)
			} else if isEmpty() {
				ret, err = identityUpdateFrontend(ctx)
			} else {
				err = uerr
			}
		} else if matchPrefix("Query") && isEmpty() {
			ret, err = identityQueryFrontend(ctx)
		} else {
			err = uerr
		}
	} else if matchPrefixI("init") && isEmpty() {
		err = initChaincode(ctx)
	} else if matchPrefixI("registry") {
		if matchPrefix("Get") && isEmpty() {
			ret, err = registryGetFrontend(ctx)
		} else if matchPrefix("Upsert") && isEmpty() {
			ret, err = registryUpsertFrontend(ctx)
		} else if matchPrefix("List") && isEmpty() {
			ret, err = registryListFrontend(ctx)
		} else {
			err = uerr
		}
	} else if matchPrefixI("role") {
		if matchPrefix("Get") && isEmpty() {
			ret, err = roleGetFrontend(ctx)
		} else if matchPrefix("Create") && isEmpty() {
			ret, err = roleCreateFrontend(ctx)
		} else if matchPrefix("Update") && isEmpty() {
			ret, err = roleUpdateFrontend(ctx)
		} else if matchPrefix("Query") && isEmpty() {
			ret, err = roleQueryFrontend(ctx)
		} else {
			err = uerr
		}
	} else if matchPrefix("singleton") {
		if matchPrefix("Get") && isEmpty() {
			ret, err = singletonGetFrontend(ctx)
		} else if matchPrefix("Upsert") && isEmpty() {
			ret, err = singletonUpsertFrontend(ctx)
		} else if matchPrefix("List") && isEmpty() {
			ret, err = singletonListFrontend(ctx)
		} else {
			err = uerr
		}
	} else {
		err = uerr
	}

	return ret, err
}
