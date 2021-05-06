package testing

import (
	stdJson "encoding/json"

	"github.com/hyperledger/fabric-protos-go/ledger/queryresult"
)

//mock for CouchDB result
type mockIterator struct {
	Results    [][]byte //array of results
	currentPos int      //current position in iterator
}

//mock for Next, get item from Result
func (mt *mockIterator) Next() (*queryresult.KV, error) {
	elem := mt.Results[mt.currentPos]
	mt.currentPos++
	var decoded map[string]stdJson.RawMessage
	if err := stdJson.Unmarshal(elem, &decoded); err != nil {
		return nil, err
	}
	id := decoded["_id"]

	return &queryresult.KV{
		Namespace: "",
		Key:       string(id),
		Value:     elem,
	}, nil
}

func (mt *mockIterator) Close() error {
	return nil
}

func (mt *mockIterator) HasNext() bool {
	if mt.currentPos >= len(mt.Results) {
		return false
	}
	return true
}
