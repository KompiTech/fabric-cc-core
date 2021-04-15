package main

import (
	"crypto/x509"
	"log"

	. "github.com/KompiTech/fabric-cc-core/v2/pkg/engine"
	"github.com/KompiTech/rmap"
)

func GetConfiguration() Configuration {
	// this is the simplest possible configuration, no business logic and lists are defined
	return Configuration{
		BusinessExecutor:          BusinessExecutor{},
		FunctionExecutor:          FunctionExecutor{},
		RecursiveResolveWhitelist: rmap.NewEmpty(),
		ResolveBlacklist:          rmap.NewEmpty(),
		ResolveFieldsBlacklist:    rmap.NewEmpty(),
		CurrentIDFunc:             certSubjectCNIDFunc,
		PreviousIDFunc:            nil,
	}
}

// this func is required to get client identity string from cert
var certSubjectCNIDFunc = func(cert *x509.Certificate) (string, error) {
	return cert.Subject.CommonName, nil
}

func main() {
	// create contractapi.ContractChaincode
	cc, err := NewChaincode(GetConfiguration())
	if err != nil {
		log.Panicf("Error creating library chaincode: %v", err)
	}

	// start serving requests
	if err := cc.Start(); err != nil {
		log.Panicf("Error starting library chaincode: %v", err)
	}
}
