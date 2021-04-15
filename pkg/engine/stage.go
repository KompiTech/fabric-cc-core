package engine

// Stage is symbolic for stage when should function(s) execute
type Stage int

// Definition of business logic stages
const (
	FirstCreate  Stage = iota // executed before asset is created from input data. Only service keys are set. Useful for setting defaults which are overwritten by input data.
	BeforeCreate              // executed after FirstCreate, after applying input data. Useful for logic that needs to be persisted to state.
	AfterCreate               // executed after BeforeCreate, after saving asset to state. Changes made here will NOT be persisted to state, but only returned to user. Useful for rendering stuff for user.

	FirstUpdate  // executed after asset was loaded from state, but before input data patch was applied.
	BeforeUpdate // executed after FirstUpdate, after applying input data
	AfterUpdate  // executed after AfterUpdate, after saving asset to state

	AfterGet // execute after asset was loaded from state. "render" method. "Data" parameter for assetGet is present in prePatch value. WARNING: resolve parameter that client sent has effect on asset being present in postPatch. You must write code to be compatible with both resolve true/false variants.

	BeforeQuery // execute before query was executed. Query is available in postPatch param, all modifications performed will be reflected on executed query. Policy MUST be defined with version -1 for given asset name, because no particular version
	AfterQuery  // execute after asset was loaded from state and is about to be put to "results" array

	BeforeDelete // execute before asset is deleted. Useful for ensuring reference integrity or disabling delete

	AfterResolve // execute after asset is resolved. "render" method, useful for field ACL

	// special stage which has asset present in prePatch param, and the patch present in postPatch param. Result of this execution is not saved. Should be used for patch validation only.
	PatchCreate
	PatchUpdate
)
