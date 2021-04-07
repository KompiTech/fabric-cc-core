package konst

const (
	AssetVersionKey     = "xxx_version" // which key in asset stores version
	AssetIdKey          = "uuid"        // which key in asset stores primary key
	AssetDocTypeKey     = "docType"     // which key in asset stores document type
	AssetFingerprintKey = "fingerprint" // which key stores fingerprint for identity assets

	ChangelogItemPrefix = "XXXCHANGELOG" // prefix for changelog key

	IdentityAssetKeyPrefix = "IDENTITY" // prefix for identity key
	RoleAssetKeyPrefix     = "ROLE"     // prefix for role key

	InitSuperuserStateKey = "INIT_MANAGER"                         // state key with initial superuser's fingerprint
	SuperuserRoleUUID     = "a00a1f64-01a1-4153-b22e-35cf7026ba7e" // magic UUID for Superuser role (must match KompiGuard model!)
	InitSingletonsKey     = "singletons"                           // key in init data that contains singletons
	InitRegistriesKey     = "registries"                           // key in init data that contains registries
	InitSuperuserKey      = "init_manager"                         // key in init data that contains first superuser's fingerprint

	RegistryItemPrefix       = "REGISTRY"             // prefix for all registryItem state keys
	SingletonItemPrefix      = "SINGLETON"            // prefix for all singleton state keys
	LatestVersionMapKey      = "LATEST_VERSION_MAP"   // key for version with map registry name -> latest version (deprecated)
	LatestSingletonMapKey    = "LATEST_SINGLETON_MAP" // key for singleton with map registry name -> latest version (deprecated)
	LatestRegistryItemPrefix = "LATEST_REGISTRY_OBJ"  // prefix for latest registry object's key for some asset name
	LatestSingletonPrefix    = "LATEST_SINGLETON_OBJ" // prefix for latest singleton key for some singleton

	SingletonCasbinObject = "singleton"
	SingletonVersionKey   = "version"
	SingletonNameKey      = "name"

	IdentityAssetName = "identity" // name of asset storing identity
	IdentityRolesKey  = "roles"    // key in identity asset with refs to roles
	RoleAssetName     = "role"     // name of asset storing role

	QueryFieldsKey   = "fields"
	QuerySelectorKey = "selector"
	QueryBookmarkKey = "bookmark"
	QueryLimitKey    = "limit"
	QuerySortKey     = "sort"

	RefDescriptionPrefix       = "REF->"     // prefix of description of field containing reference
	EntityRefDescriptionPrefix = "ENTITYREF" // prefix of description of field containing entityref

	RegistryItemDestinationKey = "destination" // key in registryItem that stores destination location
	RegistryItemSchemaKey      = "schema"      // key in registryItem that stores schema
	RegistryCasbinObject       = "registry"    // casbin object name for registry operations
	RegistryItemVersionKey     = "version"
	RegistryItemNameKey        = "name"

	LatestObjNameKey    = "name"    // key in latestObj that stores name
	LatestObjVersionKey = "version" // key in latestObj that stores version

	ChangelogCreateOperation = "create" // label for changelog when something new is created
	ChangelogUpdateOperation = "update" // label for changelog when something is updated
	ChangelogTimestampKey    = "timestamp"
	ChangelogTxIdKey         = "txid"
	ChangelogChangesKey      = "changes"
	ChangelogAssetNameKey    = "assetName"
	ChangelogVersionKey      = "version"
	ChangelogOperationKey    = "operation"
	ChangelogCasbinObject    = "changelog" // object name for changelog in Casbin

	FunctionCasbinName = "function" // function object name in Casbin (full casbin object name is /function/{invoke,query}/<name>)
	FunctionInvokeVerb = "invoke"   // function invoke name in Casbin
	FunctionQueryVerb  = "query"    // function query name in Casbin

	HistoryItemIsDeleteKey  = "is_delete"
	HistoryItemTimestampKey = "timestamp"
	HistoryItemTxIdKey      = "txid"
	HistoryItemValueKey     = "value"

	StateDestinationValue = "state" // value of destination that is considered state

	SchemaDefinitionsJPtr          = "/definitions" // json pointer of definitions inside JSONSchema
	SchemaPropertiesJPtr           = "/properties"  // json pointer of properties inside JSONSchema
	SchemaPropertiesKey            = "properties"
	SchemaDescriptionJPtr          = "/description"
	SchemaRequiredKey              = "required"                     // key in schema containing required properties
	SchemaTypeJPtr                 = "/schema/type"                 // jptr for root schema type
	SchemaAdditionalPropertiesJPtr = "/schema/additionalProperties" // jptr for additionalProperties attribute of schema
	SchemaStringType               = "string"                       // string type in JSONSchema
	SchemaObjectType               = "object"                       // object type in JSONSchema
	SchemaArrayItemsTypeJPtr       = "/items/type"                  // JSONPtr for type of items in array

	RegistryKey = "registry" // key in context that contains *Registry
	EngineKey   = "engine"   // key in context that contains Engine

	PageSize = 10 // size of returned array in query operations

	OutputResultKey   = "result"   // key under which result is wrapped in output
	OutputBookmarkKey = "bookmark" // key on output with bookmark

	ZeroByte      = "\x00" // zero byte used as separator in composite keys
	JPtrSeparator = "/"    // what separates elements in JSONPointer

	// parameter names for CC methods
	NameParam        = "name"
	DataParam        = "data"
	VersionParam     = "version"
	IdParam          = "id"
	ResolveParam     = "resolve"
	QueryParam       = "query"
	PatchParam       = "patch"
	FingerprintParam = "fingerprint"
	InputParam       = "input"
	NumberParam      = "number"

	MyAccessFuncName         = "myAccess"       // name of myAccess built-in function
	UserAccessFuncName       = "identityAccess" // name of userAccess built-in function
	UpsertRegistriesFuncName = "upsertRegistries"
	UpsertSingletonsFuncName = "upsertSingletons"
)

// ServiceKeys returns "const []string" with service keys for asset
func ServiceKeys() []string {
	return []string{AssetVersionKey, AssetIdKey, AssetDocTypeKey}
}

// builtin schema for registry item
const RegistryItemSchema = `{
  "description": "A definition of single asset type",
  "properties": {
    "destination": {
      "description": "Definition where the asset instances of the name should be stored. Possible: state or private_data",
      "pattern": "(^state$)|(^private_data)",
      "type": "string"
    },
	"schema": {
	  "description": "JSONSchema document describing the asset instances",
	  "type": "object"
	}
  },
  "required": [
	"destination", "schema"
  ],
  "additionalProperties": false
}`

// builtin schema for singleton
const SingletonItemSchema = `{
  "description": "A definition of singleton",
  "properties": {
    "value": {
      "description": "Value of singleton",
      "type": "object"
    }
  },
  "required": [
    "value"
  ],
  "additionalProperties": false
}`

// hardcoded schema for identity asset WITH ALL KEYS (do not inject service keys here)
// additionalProperties is set to true to allow extension of identity with custom data for applications
const IdentitySchema = `{
  "description": "Identity stores data related to current user's identity",
  "properties": {
    "docType": {
      "type": "string",
      "pattern": "^IDENTITY$"
    },
    "fingerprint": {
      "description": "ID for identity asset. For historic reasons is always called fingerprint.",
      "type": "string"
    },
    "org_name": {
      "description": "Copied from cert.Issuer.CommonName",
      "type": "string"
    },	
    "users": {
      "description": "REF->USER user details",
      "type": "array",
	  "uniqueItems": true,
      "items": {
        "type": "string"
      }
    },
    "roles": {
      "description": "REF->ROLE granted roles",
      "type": "array",
      "uniqueItems": true,
      "items": {
        "type": "string"
      }
    },
    "overrides": {
      "description": "Overrides for this Identity",
      "type": "array",
      "items": { "$ref": "#/definitions/override" }
    },
    "is_enabled": {
      "description": "If this is set to false, this identity cannot access anything",
      "type": "boolean"
    },
    "xxx_version": {
	  "type": "integer",
	  "minimum": 1
  	}
  },
  "required": [
    "fingerprint", "is_enabled", "docType", "xxx_version"
  ],
  "additionalProperties": true
}`

// builtin schema for role asset WITHOUT SERVICE KEYS
const RoleSchema = `{
  "description": "Role stores grants and overrides",
  "type": "object",
  "properties": {
    "name": {
      "description": "Name of the role",
      "type": "string"
    },
    "grants": { "$ref": "#/definitions/grants" },
    "overrides": { "$ref": "#/definitions/overrides" }
  },
  "required": [
    "name"
  ],
  "additionalProperties": false
}`

// SchemaDefinitions are reusable JSONSchema definitions injected to every JSONSchema under key "definitions".
const SchemaDefinitions = `{
  "uuid": {
    "description": "UUID identifier",
    "type": "string",
    "pattern": "^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}"
  },
  "fingerprint": {
    "description": "Fingerprint of some Identity",
    "type": "string",
    "pattern": "^[0-9a-f]{128}$"
  },
  "object": {
    "type": "string",
    "description": "Target of grant or override. Usual format is /<asset_name>/<primary_identifier>, but * can be used for wildcarding"
  },
  "effect": {
    "type": "string",
    "description": "Effect for action",
    "pattern": "(^allow$)|(^deny$)"
  },
  "action": {
    "type": "string",
    "description": "Name of action",
    "pattern": "^[a-z-_]*$"
  },
  "grant": {
    "description": "Grant of some action on some object. Effect is always implicitly grant.",
    "type": "object",
    "properties": {
      "object": { "$ref": "#/definitions/object" },
      "action": { "$ref": "#/definitions/action" }
    },
    "required": [
      "object", "action"
    ],
    "additionalProperties": false
  },
  "override": {
    "type": "object",
    "properties": {
      "action": { "$ref": "#/definitions/action" },
      "effect": { "$ref": "#/definitions/effect" },
      "subject": { "$ref": "#/definitions/fingerprint" }
    },
    "required": [
      "action", "subject", "effect"
    ],
    "additionalProperties": false
  },
  "grants": {
    "description": "Array of grants",
    "type": "array",
    "items": { "$ref": "#/definitions/grant" }
  },
  "overrides": {
    "description": "Array of overrides",
    "type": "array",
    "items": { "$ref": "#/definitions/override" }
  },
  "entity": {
    "description": "Specification of target entity, format <name>:<uuid>",
    "type": "string",
    "pattern": "^.*:.*$"
  }
}`

// ServiceKeyJSONSchema is part of the schema for validating service keys. It must be injected programatically into existing schema "properties" key. Keys must match to those above!!!
const SchemaServiceKeys = `{
  "uuid": { "$ref": "#/definitions/uuid" },
  "xxx_version": {
    "type": "integer",
    "minimum": 1
  },
  "docType": {
    "type": "string",
    "pattern": "^[A-Z0-9-_]"
  }
}`

// schema for Init data
const InstantiateJSONSchema = `{
  "description": "A definition of this chaincode configuration",
  "properties": {
    "init_manager": {
      "description": "Fingerprint of initial manager",
      "type": "string"
    },
    "registries": {
      "description": "Key: registry name, Value: registry object",
      "type": "object"
    },
    "singletons": {
      "description": "Key: singleton name, Value: singleton object",
      "type": "object"
    }
  },
  "additionalProperties": false
}`
