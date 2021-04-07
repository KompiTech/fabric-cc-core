package konst

const (
	RolesKey     = "roles"
	GrantsKey    = "grants"
	ObjectKey    = "object"
	SubjectKey   = "subject"
	EffectKey    = "effect"
	ActionKey    = "action"
	OverridesKey = "overrides"
	IsEnabledKey = "is_enabled"

	RolesJPtr     = "/roles"
	GrantsJPtr    = "/grants"
	ObjectJPtr    = "/object"
	SubjectJPtr   = "/subject"
	EffectJPtr    = "/effect"
	OctionJPtr    = "/action"
	OverridesJPtr = "/overrides"
	IsEnabledJPtr = "/is_enabled"

	ReadAction         = "read"
	ReadDirectAction   = "read_direct"
	CreateAction       = "create"
	CreateDirectAction = "create_direct"
	UpdateAction       = "update"
	UpdateDirectAction = "update_direct"
	DeleteAction       = "delete"
	DeleteDirectAction = "delete_direct"
	ExecuteAction      = "execute"
	UpsertAction       = "upsert"
)

// casbinModel uses RBAC with deny-override, superuser and wildcards
// RBAC - subjects are assigned to roles
// deny-override - if policy effect has "deny", then it "wins" against "allow"
// superuser - role with UUID a00a1f64-01a1-4153-b22e-35cf7026ba7e is superuser and can do anything
// wildcards - when object is for example /incident/* then ALL incidents are matched
const CasbinModel = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act, eft

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow)) && !some(where (p.eft == deny))

[matchers]
m = (g(r.sub, p.sub) && keyMatch(r.obj, p.obj) && r.act == p.act) || g(r.sub, "a00a1f64-01a1-4153-b22e-35cf7026ba7e")
`
