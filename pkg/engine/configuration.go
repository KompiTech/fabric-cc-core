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

	// SchemaDefinitionCompatibility is legacy setting, to allow the chaincode to work with older JSONSchemas (draft-07 and older) that are using reusable definitions.
	// Previously, any location for the definitions can be used, but JSONSchema newer than draft-07 allows only "$defs" key to be used.
	// To allow chaincode to work with these older schemas, set the value of SchemaDefinitionCompatibility member to name under which the definitions are stored in schema.
	// Chaincode will then convert its key to standard $defs key to allow use of newer JSONSchema library.
	// These replacements are done only at runtime and not persisted anywhere.
	// If SchemaDefinitionCompatibility is empty string, then no replacements are done.
	// This is the default setting, as the issue only occurs with legacy schemas.
	SchemaDefinitionCompatibility string
}
