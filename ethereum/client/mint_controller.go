package client

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/dan13ram/wpokt-oracle/ethereum/autogen"
)

type MintControllerContract interface {
	MaxMintLimit(opts *bind.CallOpts) (*big.Int, error)
}

type MintControllerContractImpl struct {
	contract *autogen.MintController
}

func (x *MintControllerContractImpl) MaxMintLimit(opts *bind.CallOpts) (*big.Int, error) {
	return x.contract.MaxMintLimit(opts)
}

func NewMintControllerContract(contract *autogen.MintController) MintControllerContract {
	return &MintControllerContractImpl{contract: contract}
}
