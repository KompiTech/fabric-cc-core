package engine

import (
	"github.com/KompiTech/rmap"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type mainContract struct {
	contractapi.Contract
}

// KompiTech is func is just to make Fabric happy, otherwise the contract will not work
// routing is done in route() method which is called by unknownTransactionHandler
// standard fabric public functions are not used because of more flexibility of handling errors and tracing provided
func (mc *mainContract) KompiTech(ctx ContextInterface) (string, error) {
	return rmap.NewFromMap(map[string]interface{}{"Powered by": "KompiTech"}).String(), nil
}

func NewChaincode(config Configuration) (*contractapi.ContractChaincode, error) {
	contract := new(mainContract)
	contract.TransactionContextHandler = new(Context)
	contract.BeforeTransaction = makeInitializationFunc(config)
	contract.UnknownTransaction = unknownTransactionHandler
	return contractapi.NewChaincode(contract)
}
