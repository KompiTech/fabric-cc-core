package testing

import (
	"crypto/x509"
	"encoding/pem"
	"time"

	identity2 "github.com/KompiTech/fabric-cc-core/v2/internal/testing/identity"
	msppb "github.com/hyperledger/fabric-protos-go/msp"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/pkg/errors"
)

type (
	Identities map[string]*Identity

	// implements msp.SigningIdentity
	Identity struct {
		MspId       string
		Certificate *x509.Certificate
	}
)

func MustIdentityFromPem(mspID string, certPEM []byte) *Identity {
	if id, err := IdentityFromPem(mspID, certPEM); err != nil {
		panic(err)
	} else {
		return id
	}
}

func IdentityFromPem(mspID string, certPEM []byte) (*Identity, error) {
	certIdentity, err := identity2.New(mspID, certPEM)
	if err != nil {
		return nil, err
	}
	return NewIdentity(mspID, certIdentity.Cert), nil
}

// IdentitiesFromPem returns CertIdentity (MSP ID and X.509 cert) converted PEM content
func IdentitiesFromPem(mspID string, certPEMs map[string][]byte) (ids Identities, err error) {
	identities := make(Identities)
	for role, certPEM := range certPEMs {
		if identities[role], err = IdentityFromPem(mspID, certPEM); err != nil {
			return
		}
	}
	return identities, nil
}

// IdentitiesFromFiles returns map of CertIdentity, loaded from PEM files
func IdentitiesFromFiles(mspID string, files map[string]string, getContent identity2.GetContent) (Identities, error) {
	contents := make(map[string][]byte)
	for key, filename := range files {
		content, err := getContent(filename)
		if err != nil {
			return nil, err
		}
		contents[key] = content
	}
	return IdentitiesFromPem(mspID, contents)
}

// IdentityFromFile returns Identity struct containing mspId and certificate
func IdentityFromFile(mspID string, file string, getContent identity2.GetContent) (*Identity, error) {
	content, err := getContent(file)
	if err != nil {
		return nil, err
	}

	return IdentityFromPem(mspID, content)
}

//  MustIdentitiesFromFiles
func MustIdentitiesFromFiles(mspID string, files map[string]string, getContent identity2.GetContent) Identities {
	ids, err := IdentitiesFromFiles(mspID, files, getContent)
	if err != nil {
		panic(err)
	} else {
		return ids
	}
}

func NewIdentity(mspID string, cert *x509.Certificate) *Identity {
	return &Identity{
		MspId:       mspID,
		Certificate: cert,
	}
}

func (i *Identity) Anonymous() bool {
	return false
}

// ExpiresAt returns date of certificate expiration
func (i *Identity) ExpiresAt() time.Time {
	return i.Certificate.NotAfter
}

func (i *Identity) GetIdentifier() *msp.IdentityIdentifier {
	return &msp.IdentityIdentifier{
		MSPID: i.MspId,
		ID:    i.Certificate.Subject.CommonName,
	}
}

// GetMSPIdentifier returns current MspID of identity
func (i *Identity) GetMSPIdentifier() string {
	return i.MspId
}

func (i *Identity) Validate() error {
	return nil
}

func (i *Identity) GetOrganizationalUnits() []*OUIdentifier {
	return nil
}

func (i *Identity) Verify(msg []byte, sig []byte) error {
	return nil
}

func (i *Identity) Serialize() ([]byte, error) {
	pb := &pem.Block{Bytes: i.Certificate.Raw, Type: "CERTIFICATE"}
	pemBytes := pem.EncodeToMemory(pb)
	if pemBytes == nil {
		return nil, errors.New("encoding of identity failed")
	}

	return proto.Marshal(&msppb.SerializedIdentity{Mspid: i.MspId, IdBytes: pemBytes})
}

func (i *Identity) SatisfiesPrincipal(principal *msppb.MSPPrincipal) error {
	return nil
}

func (i *Identity) Sign(msg []byte) ([]byte, error) {
	return nil, nil
}

func (i *Identity) GetPublicVersion() msp.Identity {
	return nil
}

// ==== additional method ===

func (i *Identity) GetSubject() string {
	return identity2.GetDN(&i.Certificate.Subject)
}

func (i *Identity) GetID() string {
	return identity2.IDByCert(i.Certificate)
}

// GetPEM certificate encoded to PEM
func (i *Identity) GetPEM() []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  `CERTIFICATE`,
		Bytes: i.Certificate.Raw,
	})
}

// OUIdentifier represents an organizational unit and
// its related chain of trust identifier.
type OUIdentifier struct {
	// CertifiersIdentifier is the hash of certificates chain of trust
	// related to this organizational unit
	CertifiersIdentifier []byte
	// OrganizationUnitIdentifier defines the organizational unit under the
	// MSP identified with MSPIdentifier
	OrganizationalUnitIdentifier string
}
