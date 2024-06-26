package client

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/dan13ram/wpokt-oracle/ethereum/autogen"
	"github.com/dan13ram/wpokt-oracle/ethereum/util"
)

type WarpISMContract interface {
	ValidatorCount(opts *bind.CallOpts) (*big.Int, error)
	SignerThreshold(opts *bind.CallOpts) (*big.Int, error)
	Eip712Domain(opts *bind.CallOpts) (util.DomainData, error)
}

type warpISMContract struct {
	*autogen.WarpISM
	address common.Address
}

func (x *warpISMContract) Address() common.Address {
	return x.address
}

func (x *warpISMContract) Eip712Domain(opts *bind.CallOpts) (util.DomainData, error) {
	return x.WarpISM.Eip712Domain(opts)
}

func NewWarpISMContract(address common.Address, client bind.ContractBackend) (WarpISMContract, error) {
	contract, err := autogen.NewWarpISM(address, client)
	if err != nil {
		return nil, err
	}

	return &warpISMContract{WarpISM: contract, address: address}, nil
}
