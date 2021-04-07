This is the simplest example demonstrating fundamentals of chaincode development using cc-core. It contains two assets and no business logic.

# Structure

- **asset/** - one YAML file per asset definition. Each file must contain valid JSON schema under **schema** key and additional **destination** key, which can have value either: **state** or **private_data**. This controls where the asset instances are saved
- **singleton/** - one YAML file per singleton. Not used in this example, but kept empty for tools to work properly
- **src/** - Go source for chaincode. Also contains tests.

# Preparation

Assuming that you completed all prerequisites in **{repo-root}/samples/README.txt**, start a new network with:
```
bash start.sh
```

## Detailed description

When executing **start.sh** the following steps are done (implementation in **{repo-root}/scripts/funcs.sh**):

1. Environment variables necessary for **peer** binary to function are exported (function **init_fabric_repo**)
2. File **collection_config.json** is generated from assets definition in **assets/** by **collgen** binary. Since this example uses no assets stored in private data, the generated file contains empty array (function **gen_collection_config**).
3. Directory **META-INF** is generated from assets definition and stored in **src/** by **metainfgen** binary. These definitions are packaged with chaincode and contain indexes for CouchDB to use (function **gen_cc_metainf**).
4. Script **network.sh** from fabric-samples repository is used to create Fabric network inside Docker, create a channel and join all peers to the channel.
5. File **cc_init.json** is generated from asset definition by **initgen** binary. This contains definition of all assets serialized in single JSON message to be used. It also contains a required fingerprint of the first super user in the system. (function **gen_cc_init**)
6. **Init** chaincode method is executed with **cc_init.json** file data as its argument. This bootstraps the network and prepares it for use.
7. **micro-rest** server is built and executed, it provides HTTP API on localhost:10000

## Working with micro-rest

The **micro-rest** is simple proof-of-concept server for experimenting with REST calls to chaincode on a local machine. It uses **peer** binary to interact with chaincode. Try these calls to get familiar.

### Adding an own identity

For cc-core chaincode, it is required for each actor to have an **identity** asset instance matching his fingerprint, otherwise error is returned. Fingerprint generation is customizable and default implementation is to use CommonName of the Subject of client certificate.

To add current actor (don't worry about different actors for now), use this:
```
curl --request POST \
  --url http://localhost:10000/api/v1/identities/me \
  --header 'Content-Type: application/json' \
  --data '{}' | jq
```

You should get a response which looks like this:
```
{
  "result": {
    "docType": "IDENTITY",
    "fingerprint": "Admin@org1.example.com",
    "is_enabled": true,
    "roles": [
      "a00a1f64-01a1-4153-b22e-35cf7026ba7e"
    ],
    "xxx_version": 1
  }
}
```

This means that your identity was successfully created. Also note the **roles** key which contains array with one value. This role **a00a1f64-01a1-4153-b22e-35cf7026ba7e** is special super user role, which was granted to you because it was matched with init data. More info on roles later.

Also note, that this method is safe to call repeatedly - it will just return the identity, if it already exists.

### Creating an asset instance

To create a new asset instance, use this:
```
curl --request POST \
  --url http://localhost:10000/api/v1/assets/author/8cf1c218-462d-4459-8fd1-c6fb5425b30c \
  --header 'Content-Type: application/json' \
  --data '{
	"name": "Edgar Allan Poe"
}' | jq
```

We are using endpoint format **/assets/{name}/{uuid}** to create a new asset instance. Chaincode 'knows' about this asset type, because it was created as part of initialization. Data is validated against JSON schema before saving. You should get a response like this:
```
{
  "result": {
    "docType": "AUTHOR",
    "name": "Edgar Allan Poe",
    "uuid": "8cf1c218-462d-4459-8fd1-c6fb5425b30c",
    "xxx_version": 1
  }
}
```

Also note, that you can omit the **/8cf1c218-462d-4459-8fd1-c6fb5425b30c** part of the URL. In that case, random UUID will be generated for you. We are using constant UUID in this example, because we are going to refer to it.

### Creating an asset instance with reference

If you look at **assets/book.yaml**, you will notice that key **authors** is an array with description containing: 'REF->AUTHOR'. This is magic string telling cc-core, that the key contains reference to another asset. In addition to requiring data to be validated against JSON schema, cc-core will check that the reference target exists.

To demonstrate, let's create a book instance and refer to author, that was created in a previous step:
```
curl --request POST \
  --url http://localhost:10000/api/v1/assets/book/e86aec69-0558-4cdc-8dae-8dda643010ff \
  --header 'Content-Type: application/json' \
  --data '{
	"name": "The Raven",
	"authors": ["8cf1c218-462d-4459-8fd1-c6fb5425b30c"]
}' | jq
```

You should get a response:
```
{
  "result": {
    "authors": [
      "8cf1c218-462d-4459-8fd1-c6fb5425b30c"
    ],
    "docType": "BOOK",
    "name": "The Raven",
    "uuid": "e86aec69-0558-4cdc-8dae-8dda643010ff",
    "xxx_version": 1
  }
}
```

### Updating asset

To update existing book instance by its UUID, use:
```
curl --request PATCH \
  --url http://localhost:10000/api/v1/assets/book/e86aec69-0558-4cdc-8dae-8dda643010ff \
  --header 'Content-Type: application/json' \
  --data '{"name": "The Black Cat"}' | jq
```

Data supplied to the method are understood as JSON Merge Patch - this means that you have to specify only the fields you want to change. You do not have to specify the whole asset. To remove some key, set it to 'null' value.

### Reading and resolving asset

To read existing asset by its UUID, use:
```
curl --request GET \
  --url 'http://localhost:10000/api/v1/assets/book/e86aec69-0558-4cdc-8dae-8dda643010ff' | jq
```

You will get the same response as when you created the book asset:
```
{
  "result": {
    "authors": [
      "8cf1c218-462d-4459-8fd1-c6fb5425b30c"
    ],
    "docType": "BOOK",
    "name": "The Raven",
    "uuid": "e86aec69-0558-4cdc-8dae-8dda643010ff",
    "xxx_version": 1
  }
}
```

A very useful feature of cc-core is ability to resolve references, this means replacing the foreign key with actual reference target, by adding **resolve=true** URL param:
```
curl --request GET \
  --url 'http://localhost:10000/api/v1/assets/book/e86aec69-0558-4cdc-8dae-8dda643010ff?resolve=true' | jq
```

You will get a response with both book and author data:
```
{
  "result": {
    "authors": [
      {
        "docType": "AUTHOR",
        "name": "Edgar Allan Poe",
        "uuid": "8cf1c218-462d-4459-8fd1-c6fb5425b30c",
        "xxx_version": 1
      }
    ],
    "docType": "BOOK",
    "name": "The Raven",
    "uuid": "e86aec69-0558-4cdc-8dae-8dda643010ff",
    "xxx_version": 1
  }
}
```

# Testing

Cc-core also provides completely independent way to write automated tests for your chaincode. It is explained in **src/main_test.go**. These tests are completely self-contained and does not depend on **fabric-samples** repository. However, docker is still required for CouchDB to work.

# Summary

In this example we explained the following topics:

- how are assets defined
- how is the chaincode in docker bootstrapped
- how identity works 
- how to execute REST calls against chaincode (create, update, read) and resolve asset
- how to write automated tests