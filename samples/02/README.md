Let's use the same assets as in example 01, but add some extra functionality:

- only admin can create or update authors and books, other users can only read. Nobody can delete.
- lets assume author contains sensitive data and needs to be stored in Fabric's private data  
- add configuration with banned words in books titles (book cannot contain these words in name)
- add function to get number of books written by each author
- change the schema at runtime

# Preparation

It is required to completely tear down the network from previous example, use:
```
bash stop.sh
```

After that, start a new network with
```
bash start.sh
```

# Private data

If you look at **assets/author.yaml**, you will note that **destination** key now has value 'private_data'. This is all you need to do, to store this asset in private data collection. You can work with all private data assets in completely the same way as state assets.

# Business logic example - access control

Lets try to create an asset as different identity. You can use HTTP header 'X-Fabric-User' to select. This depends on identities created by Fabric's test network.
Possible choices are:
- Admin@org1.example.com (the default that is used when no header is present. Also is super user)
- User1@org1.example.com
- Admin@org2.example.com
- User1@org2.example.com

First, create an identity as User1@org1.example.com:
```
curl --request POST \
  --url http://localhost:10000/api/v1/identities/me \
  --header 'Content-Type: application/json' \
  --header 'X-Fabric-User: User1@org1.example.com' \
  --data '{}' | jq
```

In response, you can see that no roles are granted (**roles** key is not present):
```
{
  "result": {
    "docType": "IDENTITY",
    "fingerprint": "User1@org1.example.com",
    "is_enabled": true,
    "xxx_version": 1
  }
}
```

When trying to create an author asset:
```
curl --request POST \
  --url http://localhost:10000/api/v1/assets/author \
  --header 'Content-Type: application/json' \
  --header 'X-Fabric-User: User1@org1.example.com' \
  --data '{
	"name": "Edgar Allan Poe"
}' | jq
```

You will get a following response:
```
{
  "error": "Error: endorsement failure during invoke. response: status:500 message:\"bexec.Execute(), stage: BeforeCreate failed: be.executePolicy() failed on stage: 1: func index: 0 returned error: permission denied, sub: User1@org1.example.com, obj: /author/e11d9f7b-b683-e89c-1217-b6987925bd83, act: create|||403\" \n"
}
```

This is the error being produced by reusable.EnforceCreate business logic. For more info, check the **{repo-root}/src/main.go** and also tests in **{repo-root}/src/kompiguard_test.go**.

# Managing roles

Assume we want to allow User1@org1.example.com to create book assets, we first need to create identity for super user (we will need it to grant the roles):
```
curl --request POST \
  --url http://localhost:10000/api/v1/identities/me \
  --header 'Content-Type: application/json' \
  --data '{}' | jq
```

Next, the role asset with list of grants needs to be created:
```
curl --request POST \
  --url http://localhost:10000/api/v1/assets/role/e86aec69-0558-4cdc-8dae-8dda643010ff \
  --header 'Content-Type: application/json' \
  --data '{"name": "Author creator", "grants": [{"object": "/author/*", "action": "create"}]}' | jq
```

Response will be:
```
{
  "result": {
    "docType": "ROLE",
    "grants": [
      {
        "action": "create",
        "object": "/author/*"
      }
    ],
    "name": "Author creator",
    "uuid": "e86aec69-0558-4cdc-8dae-8dda643010ff",
    "xxx_version": 1
  }
}
```

The object is in format /{asset name}/{asset uuid}. Wildcard '*' can be used instead of UUID to grant for all assets. Role is hardcoded asset that cannot be modified. Schema can be found in **{repo-root}/src/konst/engine.go** as RoleSchema.

Next, role can be granted to identity of User1@org1.example.com. We do this by updating the identity as ordinary asset:
```
curl --request PATCH \
  --url http://localhost:10000/api/v1/assets/identity/User1@org1.example.com \
  --header 'Content-Type: application/json' \
  --data '{
	"roles": ["e86aec69-0558-4cdc-8dae-8dda643010ff"]
}' | jq
```

Response confirms the update:
```
{
  "result": {
    "docType": "IDENTITY",
    "fingerprint": "User1@org1.example.com",
    "is_enabled": true,
    "roles": [
      "e86aec69-0558-4cdc-8dae-8dda643010ff"
    ],
    "xxx_version": 1
  }
}
```

Identity assets are also hardcoded like roles, and their schema can be found in **{repo-root}/src/konst/engine.go** as IdentitySchema. Important difference is that schema does allow additional keys to be present. This enables particular chaincode applications that use cc-core as a library to extend this asset and store for example user info.

After this update, User1@org1.example.com can create author asset:
```
curl --request POST \
  --url http://localhost:10000/api/v1/assets/author \
  --header 'Content-Type: application/json' \
  --header 'X-Fabric-User: User1@org1.example.com' \
  --data '{
	"name": "The Raven"
}' | jq
```

To allow creation of books, you can either add the grant to the role, or create and grant second role.

# Working with singletons

Singletons are objects that are stored separately from assets and have no schema. They can be used to hold configuration that can be easily changed at runtime. They require **/singleton/{name}** object with read, create or update action to be accessed by other users.

File **singleton/banned_words.yaml** is processed by tools and already loaded in chaincode as part of Init payload. For implementation side of the business logic, check **src/main.go**, variable CheckForBannedWords.

We can try to create an author which name contains a banned word 'example':
```
curl --request POST \
  --url http://localhost:10000/api/v1/assets/author \
  --header 'Content-Type: application/json' \
  --data '{
	"name": "example Edgar Allan Poe"
}' | jq
```

Error will be returned, showing match against singleton configuration:
```
{
    "error": "Error: endorsement failure during invoke. response: status:500 message:\"bexec.Execute(), stage: BeforeCreate failed: be.executePolicy() failed on stage: 1: func index: 1 returned error: name contains banned word(s)\" \n"
}
```

Singletons can be easily updated at runtime. This will always replace the singleton with new data (as compared to JSON Patch in case of updating assets). Also, this singleton will have incremented version.

Let's demonstrate this by removing the word 'example' from the list:
```
curl --request POST \
  --url http://localhost:10000/api/v1/singletons/banned_words \
  --header 'Content-Type: application/json' \
  --data '{
	"value": {
	  "banned_words": [
	    "word"
	  ]
	}
}' | jq
```

Response shows the new list and incremented version:
```
{
  "result": {
    "name": "banned_words",
    "value": {
      "banned_words": [
        "word"
      ]
    },
    "version": 2
  }
}
```

Author with a previously banned word in name can now be created:
```
curl --request POST \
  --url http://localhost:10000/api/v1/assets/author \
  --header 'Content-Type: application/json' \
  --data '{
	"name": "example Edgar Allan Poe"
}' | jq
```

# Working with functions

If you require a general functionality added to chaincode that is not part of CRUD operations, functions are ideal for this. Each one is identified by name and there are 2 ways to execute them:
* query - function is read only and state will not be changed
* invoke - function is allowed to change state

You can use functions for example for:
* reporting and returning statistic data
* encapsulating more complicated functionality (for example create an Invoice asset with multiple InvoiceLine assets in one transaction)

In this example, we will create a sort of view, that returns number of books written by each unique author. Implementation can be found in **src/main.go** along with comments. To test the function, lets add some additional books and author:

```
curl --request POST \
  --url http://localhost:10000/api/v1/assets/author/b9e45c0d-4cd8-4f0a-bad6-5a5a370ed66a \
  --header 'Content-Type: application/json' \
  --data '{
	"name": "George Orwell"
}' | jq
```

```
curl --request POST \
  --url http://localhost:10000/api/v1/assets/book/ \
  --header 'Content-Type: application/json' \
  --data '{
	"name": "Animal Farm",
	"authors": ["b9e45c0d-4cd8-4f0a-bad6-5a5a370ed66a"]
}' | jq
```

```
curl --request POST \
  --url http://localhost:10000/api/v1/assets/book/ \
  --header 'Content-Type: application/json' \
  --data '{
	"name": "1984",
	"authors": ["b9e45c0d-4cd8-4f0a-bad6-5a5a370ed66a"]
}' | jq
```

Now, function can be executed (query mode is enough, because no state is modified, but invoke will also work):
```
curl --request GET \
  --url http://localhost:10000/api/v1/functions-query/authorStats \
  --header 'Content-Type: application/json' \
  --data '{}' | jq
```

From the result, we can see function correctly counting. Author "example Edgar Allan Poe" is not present, because there are no books linked:
```
{
  "result": {
    "George Orwell": 2
  }
}
```

# Asset version management

Let's assume the library is running for some time with data present. New requirement comes, to also store the birth year of each Author. CC-core allows to do this without disrupting existing data. Each asset instance contains key "xxx_version" which describes its version. We can easily add the new version at runtime to registry with the new mandatory "birth_year" field:
```
curl --request POST \
  --url http://localhost:10000/api/v1/registries/author \
  --header 'Content-Type: application/json' \
  --data '{
	"destination": "private_data",
	"schema": {
	  "additionalProperties": false,
      "description": "An author is a person",
      "properties": {
        "name": {
          "type": "string"
        },
        "birth_year": {
          "type": "integer"
        }
      },
      "required": [
        "name",
        "birth_year"
      ],
      "title": "Author",
      "type": "object"
	  }
}' | jq
```

We can see the new version in the response:
```
{
  "result": {
    "destination": "private_data",
    "name": "author",
    "schema": {
      "additionalProperties": false,
      "description": "An author is a person",
      "properties": {
        "birth_year": {
          "type": "integer"
        },
        "name": {
          "type": "string"
        }
      },
      "required": [
        "name",
        "birth_year"
      ],
      "title": "Author",
      "type": "object"
    },
    "version": 2
  }
}
```

When getting the Author that was created previously:
```
curl --request GET \
  --url http://localhost:10000/api/v1/assets/author/b9e45c0d-4cd8-4f0a-bad6-5a5a370ed66a | jq
```

We can see that its version is still 1:
```
{
  "result": {
    "docType": "AUTHOR",
    "name": "George Orwell",
    "uuid": "b9e45c0d-4cd8-4f0a-bad6-5a5a370ed66a",
    "xxx_version": 1
  }
}
```

CC-core never migrates existing assets automatically. To migrate an existing author to a new version, we must provide a target version and migration patch (new asset instance must validate the new schema):
```
curl --request POST \
  --url 'http://localhost:10000/api/v1/assets/migrate/author/b9e45c0d-4cd8-4f0a-bad6-5a5a370ed66a?version=2' \
  --header 'Content-Type: application/json' \
  --data '{"birth_year": 1903}' | jq
```

Response will confirm, that asset is now stored as version 2:
```
{
  "result": {
    "birth_year": 1903,
    "docType": "AUTHOR",
    "name": "George Orwell",
    "uuid": "b9e45c0d-4cd8-4f0a-bad6-5a5a370ed66a",
    "xxx_version": 2
  }
}
```

This pattern allows incremental migration of assets, without disturbing service.
