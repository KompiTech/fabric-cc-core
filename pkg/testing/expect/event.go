package expect

import (
	convert2 "github.com/KompiTech/fabric-cc-core/v2/internal/testing/convert"
	"github.com/hyperledger/fabric-protos-go/peer"
	g "github.com/onsi/gomega"
)

func EventIs(event *peer.ChaincodeEvent, expectName string, expectPayload interface{}) {
	g.Expect(event.EventName).To(g.Equal(expectName), `event name not match`)

	EventPayloadIs(event, expectPayload)
}

// EventPayloadIs expects peer.ChaincodeEvent payload can be marshaled to
// target interface{} and returns converted value
func EventPayloadIs(event *peer.ChaincodeEvent, target interface{}) interface{} {
	g.Expect(event).NotTo(g.BeNil())
	data, err := convert2.FromBytes(event.Payload, target)
	description := ``
	if err != nil {
		description = err.Error()
	}
	g.Expect(err).To(g.BeNil(), description)
	return data
}
