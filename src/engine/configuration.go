package engine

import (
	"crypto/x509"

	"github.com/KompiTech/rmap"
)

type IDFunc func(cert *x509.Certificate) (string, error)

// Configuration is used to manage configuration of this particular dynamic chaincode (previously called Engine)
type Configuration struct {
	BusinessExecutor          BusinessExecutor // Abstraction for executing business logic
	FunctionExecutor          FunctionExecutor // Abstraction for executing generic functionality
	RecursiveResolveWhitelist rmap.Rmap        // Rmap of asset names that have recursive resolve enabled
	ResolveBlacklist          rmap.Rmap        // Rmap of asset names that are forbidden from being resolved
	ResolveFieldsBlacklist    rmap.Rmap        // Rmap of asset name -> list of fields to not resolve
	CurrentIDFunc             IDFunc           // Function to get identity fingerprint
	PreviousIDFunc            *IDFunc          // Previous function to get identity fingerprint when migration is desired
}
