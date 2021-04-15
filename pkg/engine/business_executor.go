package engine

import (
	"strings"

	. "github.com/KompiTech/fabric-cc-core/v2/pkg/konst"
	"github.com/KompiTech/rmap"
	"github.com/pkg/errors"
)

// BusinessPolicyMember is signature for business logic function
type BusinessPolicyMember = func(ctx ContextInterface, prePatch *rmap.Rmap, postPatch rmap.Rmap) (rmap.Rmap, error)

// FuncKey is the key for policy member - it defines the asset name and its version to execute on
// version -1 has special meaning -> wildcard -> execute on ALL versions (do not use in production or the migrations of business logic will be impossible)
type FuncKey struct {
	Name    string
	Version int
}

// businessPolicy defines which functions should be executed on what asset versions and when
// level 1: {assetName, assetVersion}
// level 2: Stage (execution stage)
// level 3: list of functions to execute in this order
type businessPolicy = map[FuncKey]map[Stage][]BusinessPolicyMember

type StageMembers = map[Stage][]BusinessPolicyMember

// BusinessExecutor holds configuration for business logic policy and can execute it on asset instance
type BusinessExecutor struct {
	policy businessPolicy
}

func NewBusinessExecutor() *BusinessExecutor {
	return &BusinessExecutor{
		policy: businessPolicy{},
	}
}

func (be *BusinessExecutor) SetPolicy(key FuncKey, members StageMembers) {
	be.policy[key] = members
}

func (be BusinessExecutor) GetPolicy(key FuncKey, stage Stage) []BusinessPolicyMember {
	var allPolicies map[Stage][]BusinessPolicyMember
	var policy []BusinessPolicyMember
	var exists bool

	// sanitize key just in case
	key.Name = strings.ToLower(key.Name)

	allPolicies, exists = be.policy[key]
	if !exists {
		// if policy is not found by key, try wildcard key
		key.Version = -1

		var wildcardExists bool
		allPolicies, wildcardExists = be.policy[key]
		if !wildcardExists {
			// not even wildcard policy exists
			return []BusinessPolicyMember{}
		}
	}

	policy, exists = allPolicies[stage]
	if !exists {
		// asset has policy, but not for this stage
		return []BusinessPolicyMember{}
	}

	return policy
}

func (be BusinessExecutor) getPolicyForAsset(asset rmap.Rmap, stage Stage) ([]BusinessPolicyMember, error) {
	isAsset, err := IsAsset(asset)
	if err != nil {
		return nil, errors.Wrap(err, "asset.IsAsset() failed")
	}

	if !isAsset {
		// if asset does not contain all service keys, business logic cannot be executed
		// this can happen if it was previously filtered by KompiGuard
		// this is not an error
		return nil, nil
	}

	docType, err := AssetGetDocType(asset)
	if err != nil {
		return nil, errors.Wrap(err, "asset.GetDocType() failed")
	}

	version, err := AssetGetVersion(asset)
	if err != nil {
		return nil, errors.Wrap(err, "asset.GetVersion() failed")
	}

	funcKey := FuncKey{
		Name:    docType,
		Version: version,
	}

	return be.GetPolicy(funcKey, stage), nil
}

func (be BusinessExecutor) executePolicy(ctx ContextInterface, policy []BusinessPolicyMember, assetPrePatch *rmap.Rmap, assetPostPatch rmap.Rmap) (rmap.Rmap, error) {
	// execute all required business logic funcs in order
	// each func receives the same assetPrePatch
	// assetPostPatch argument is result of previous func execution
	next := assetPostPatch
	for index, fnc := range policy {
		tmpNext, err := fnc(ctx, assetPrePatch, next)
		if err != nil {
			return rmap.Rmap{}, errors.Wrapf(err, "func index: %d returned error", index)
		}
		next = tmpNext
	}
	return next, nil
}

// Execute loads policy from asset and executes it on asset
func (be BusinessExecutor) Execute(ctx ContextInterface, stage Stage, assetPrePatch *rmap.Rmap, assetPostPatch rmap.Rmap) (rmap.Rmap, error) {
	policy, err := be.getPolicyForAsset(assetPostPatch, stage)
	if err != nil {
		return rmap.Rmap{}, errors.Wrap(err, "be.getPolicyForAsset() failed")
	}

	result, err := be.executePolicy(ctx, policy, assetPrePatch, assetPostPatch)
	if err != nil {
		return rmap.Rmap{}, errors.Wrapf(err, "be.executePolicy() failed on stage: %d", stage)
	}

	return result, nil
}

// ExecuteCustom loads policy from asset but executes on custom assetPrePatch and assetPostPatch
func (be BusinessExecutor) ExecuteCustom(ctx ContextInterface, stage Stage, policySource rmap.Rmap, assetPrePatch *rmap.Rmap, assetPostPatch rmap.Rmap) (rmap.Rmap, error) {
	policy, err := be.getPolicyForAsset(policySource, stage)
	if err != nil {
		return rmap.Rmap{}, errors.Wrap(err, "be.getPolicyForAsset() failed")
	}

	result, err := be.executePolicy(ctx, policy, assetPrePatch, assetPostPatch)
	if err != nil {
		return rmap.Rmap{}, errors.Wrapf(err, "be.executePolicy() failed on stage: %d", stage)
	}

	return result, nil
}

// ExecuteCustomPolicy executes some policy, nothing is inferred from anything
func (be BusinessExecutor) ExecuteCustomPolicy(ctx ContextInterface, policy []BusinessPolicyMember, assetPrePatch *rmap.Rmap, assetPostPatch rmap.Rmap) (rmap.Rmap, error) {
	result, err := be.executePolicy(ctx, policy, assetPrePatch, assetPostPatch)
	if err != nil {
		return rmap.Rmap{}, errors.Wrap(err, "be.executePolicy() failed")
	}

	return result, nil
}
