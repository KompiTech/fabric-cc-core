package testing

import (
	"errors"
	//"github.com/golang/protobuf/proto"
	//"github.com/hyperledger/fabric/msp"
	//"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	//pmsp "github.com/hyperledger/fabric-sdk-go/pkg/msp"
)

// TransformCreator transforms arbitrary tx creator (pmsp.SerializedIdentity etc)  to mspID string, certPEM []byte,
func TransformCreator(txCreator ...interface{}) (mspID string, certPEM []byte, err error) {
	if len(txCreator) == 1 {
		p := txCreator[0]
		switch p.(type) {

		case *Identity:
			x := p.(*Identity)
			return x.GetMSPIdentifier(), x.GetPEM(), nil
		}
	}

	return ``, nil, errors.New("unknown args type to cckit.MockStub.From func")
}
