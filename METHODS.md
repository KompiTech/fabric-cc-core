# Chaincode methods

Common constraints:
- All method arguments are mandatory
- All arguments are passed as strings
- All strings should be lowercase
- All methods return string that contains JSON document

## Asset family

These methods enable CRUD operations with assets.

### assetCreate

Create a new asset instance. If instance with UUID exists, error is returned.

Arguments:

- **name** - name of asset type
- **data** - JSON data of asset instance
- **version** - integer describing desired version. Use -1 to use latest available
- **id** - desired UUID of asset instance. Use empty string "" to autogenerate

MicroREST routes:

- POST /api/v1/assets/{name}
- POST /api/v1/assets/{name}?version={version} 
- POST /api/v1/assets/{name}/{id}
- POST /api/v1/assets/{name}/{id}?version={version}

**data** are in body

### assetGet

Returns data for asset instance

Arguments:

- **name** - name of asset type
- **id** - UUID of asset instance
- **resolve** - should the asset be resolved? (use "true" or "false")
- **data** - optional data (use empty JSON object "{}" as default)

MicroREST routes:

- GET /api/v1/assets/{name}/{id}
- GET /api/v1/assets/{name}/{id}?resolve={resolve}

**data** passing is not implemented (always passes {})

### assetUpdate

Update existing asset instance. Data are understood as JSONPatch which is applied to instance

Arguments:

- **name** - name of asset type
- **id** - UUID of asset instance
- **patch** - JSONPatch data

MicroREST routes:

- PATCH /api/v1/assets/{name}/{id}

**patch** is in body

### assetDelete

Delete existing asset instance.

Arguments:

- **name** - name of asset type
- **id** - UUID of asset instance

MicroREST routes:

- not yet implemented

### assetMigrate

Change asset version. Asset after migrating must validate JSONSchema for given target version

Arguments:

- **name** - name of asset type
- **id** - UUID of asset instance
- **version** - target version number, use -1 for latest
- **patch** - JSONPatch that will be applied to asset instance

MicroREST routes:

- PATCH /api/v1/assets/migrate/{name}/{id}?version={version}

**patch** is in body

### assetHistory

Returns history log for some asset instance

Arguments:

- **name** - name of asset type
- **id** - UUID of asset instance

MicroREST routes:

- not yet implemented

### assetQuery

Perform rich query on some asset type. Returns max 10 asset instances per page. To get additional pages, add **bookmark** key with bookmark value returned by previous call to **query**. If result array contains less than 10 instances, you are at last page.

Arguments:

- **name** - name of asset type
- **query** - JSON document containing CouchDB query to use (usually **selector** key, see CouchDB docs)
- **resolve** - should the assets be resolved? (use "true" or "false")

MicroREST routes:

- OPTIONS /api/v1/assets/{name}
- OPTIONS /api/v1/assets/{name}?resolve={resolve}

**query** is in body

## Function family

Allows invocation of chaincode functions

### functionInvoke / functionQuery

Only difference is that **functionInvoke** can modify state, **functionQuery** never modifies state 

Arguments:

- **name** - name of function
- **input** - JSON document with input data passed to function

MicroREST routes:

- POST /api/v1/functions-query/{name}
- POST /api/v1/functions-invoke/{name}

**input** is in body

## Identity family

Allows management of identity. Also contains deprecated functions for backward compatibility, that are not explained here

### identityAddMe

Creates identity asset for current client identity. If identity asset does exist, it is returned -> this method is safe to call repeatedly.

Arguments:

- **input** - optional JSON document. Use empty JSON document as default

MicroREST routes:

- POST /api/v1/identities/me

**input** is in body

## Registry family

Enables operation with asset registry, to define new asset classes or modify existing.

### registryGet

Returns document with **schema** and **destination** keys, describing asset class.

Arguments:

- **name** - name of asset class
- **version** - desired version number, use -1 for latest

MicroREST routes:

- GET /api/v1/registries/{name}
- GET /api/v1/registries/{name}?version={version}

### registryList

Returns list of all available asset classes with all versions

Arguments: none

MicroREST routes:

- not yet implemented

### registryUpsert

Creates a new asset class or new version of existing asset class. Previous versions are never overwritten or modified.

Arguments:

- **name** - name of asset class
- **data** - JSON document with mandatory keys: **schema**, **destination** ("state" or "private_data")

MicroREST routes:

- POST /api/v1/registries/{name}

**data** is in body

## Singleton family

Allows creating, upserting and reading singletons. Deletion is not allowed.

### singletonGet

Get existing singleton data

Arguments:

- **name** - name of singleton
- **version** - desired version number, use -1 for latest

MicroREST routes:

- GET /api/v1/singletons/{name}
- GET /api/v1/singletons/{name}?version={version}

### singletonList

List all existing singletons in all versions

Arguments: none

MicroREST routes:

- not yet implemented

### singletonUpsert

Creates a new singleton or new version of existing one. Previous versions are never modified.

Arguments:

- **name** - name of singleton
- **data** - JSON document describing the singleton. It is necessary to wrap the singleton into another JSON document with only "value" key.

MicroREST routes:

- POST /api/v1/singletons/{name}

**data** is in body