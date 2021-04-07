module github.com/KompiTech/fabric-cc-core/v2

go 1.14

require (
	github.com/KompiTech/rmap v1.12.0
	github.com/casbin/casbin/v2 v2.23.0
	github.com/golang/protobuf v1.4.3
	github.com/google/go-cmp v0.5.2 // indirect
	github.com/hashicorp/golang-lru v0.5.4
	github.com/hyperledger/fabric-chaincode-go v0.0.0-20201119163726-f8ef75b17719
	github.com/hyperledger/fabric-contract-api-go v1.1.1
	github.com/hyperledger/fabric-protos-go v0.0.0-20210127161553-4f432a78f286
	github.com/hyperledger/fabric-sdk-go v1.0.0-rc1
	github.com/json-iterator/go v1.1.10
	github.com/onsi/ginkgo v1.14.2
	github.com/onsi/gomega v1.10.4
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
)

replace (
	github.com/cucumber/godog => github.com/cucumber/godog v0.8.1
	github.com/golang/protobuf => github.com/golang/protobuf v1.3.3
	github.com/qri-io/jsonschema => github.com/qri-io/jsonschema v0.1.2
	google.golang.org/genproto => google.golang.org/genproto v0.0.0-20200218151345-dad8c97a84f5
	google.golang.org/grpc => google.golang.org/grpc v1.27.1
)
