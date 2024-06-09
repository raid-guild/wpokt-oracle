package client

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

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

type MintControllerFulfillmentIteratorImpl struct {
	iterator *autogen.MintControllerFulfillmentIterator
}

func (x *MintControllerFulfillmentIteratorImpl) Next() bool {
	return x.iterator.Next()
}

func (x *MintControllerFulfillmentIteratorImpl) Event() *autogen.MintControllerFulfillment {
	return x.iterator.Event
}

func (x *MintControllerFulfillmentIteratorImpl) Close() error {
	return x.iterator.Close()
}

func (x *MintControllerFulfillmentIteratorImpl) Error() error {
	return x.iterator.Error()
}

type MintControllerContractImpl struct {
	contract *autogen.MintController
	address  common.Address
}

func (x *MintControllerContractImpl) Address() common.Address {
	return x.address
}

func (x *MintControllerContractImpl) ParseFulfillment(log types.Log) (*autogen.MintControllerFulfillment, error) {
	return x.contract.ParseFulfillment(log)
}

func (x *MintControllerContractImpl) FilterFulfillment(opts *bind.FilterOpts, orderID [][32]byte) (MintControllerFulfillmentIterator, error) {
	iterator, err := x.contract.FilterFulfillment(opts, orderID)
	if err != nil {
		return nil, err
	}
	return &MintControllerFulfillmentIteratorImpl{iterator: iterator}, nil
}

func (x *MintControllerContractImpl) MaxMintLimit(opts *bind.CallOpts) (*big.Int, error) {
	return x.contract.MaxMintLimit(opts)
}

func NewMintControllerContract(address common.Address, client *ethclient.Client) (MintControllerContract, error) {
	contract, err := autogen.NewMintController(address, client)
	if err != nil {
		return nil, err
	}

	return &MintControllerContractImpl{contract: contract, address: address}, nil
}
