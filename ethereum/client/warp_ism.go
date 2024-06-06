package client

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/dan13ram/wpokt-oracle/ethereum/autogen"
)

type DomainData struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}

type WarpISMContract interface {
	ValidatorCount(opts *bind.CallOpts) (*big.Int, error)
	Eip712Domain(opts *bind.CallOpts) (DomainData, error)
}

type WarpISMContractImpl struct {
	contract *autogen.WarpISM
}

func (x *WarpISMContractImpl) ValidatorCount(opts *bind.CallOpts) (*big.Int, error) {
	return x.contract.ValidatorCount(opts)
}

func (x *WarpISMContractImpl) Eip712Domain(opts *bind.CallOpts) (DomainData, error) {
	return x.contract.Eip712Domain(opts)
}

func NewWarpISMContract(contract *autogen.WarpISM) WarpISMContract {
	return &WarpISMContractImpl{contract: contract}
}
