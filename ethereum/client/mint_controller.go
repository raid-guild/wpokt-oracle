package client

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/dan13ram/wpokt-oracle/ethereum/autogen"
)

type MintControllerContract interface {
	Address() common.Address
	MaxMintLimit(opts *bind.CallOpts) (*big.Int, error)
	FilterFulfillment(opts *bind.FilterOpts, orderID [][32]byte) (MintControllerFulfillmentIterator, error)
	ParseFulfillment(log types.Log) (*autogen.MintControllerFulfillment, error)
}

type MintControllerFulfillmentIterator interface {
	Next() bool
	Event() *autogen.MintControllerFulfillment
	Close() error
	Error() error
}

type mintControllerFulfillmentIterator struct {
	*autogen.MintControllerFulfillmentIterator
}

func (x *mintControllerFulfillmentIterator) Event() *autogen.MintControllerFulfillment {
	return x.MintControllerFulfillmentIterator.Event
}

type mintControllerContract struct {
	*autogen.MintController
	address common.Address
}

func (x *mintControllerContract) Address() common.Address {
	return x.address
}

func (x *mintControllerContract) FilterFulfillment(opts *bind.FilterOpts, orderID [][32]byte) (MintControllerFulfillmentIterator, error) {
	iterator, err := x.MintController.FilterFulfillment(opts, orderID)
	if err != nil {
		return nil, err
	}
	return &mintControllerFulfillmentIterator{iterator}, nil
}

func NewMintControllerContract(address common.Address, client bind.ContractBackend) (MintControllerContract, error) {
	contract, err := autogen.NewMintController(address, client)
	if err != nil {
		return nil, err
	}

	return &mintControllerContract{MintController: contract, address: address}, nil
}
