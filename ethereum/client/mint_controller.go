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

type mintControllerFulfillmentIterator struct {
	iterator *autogen.MintControllerFulfillmentIterator
}

func (x *mintControllerFulfillmentIterator) Next() bool {
	return x.iterator.Next()
}

func (x *mintControllerFulfillmentIterator) Event() *autogen.MintControllerFulfillment {
	return x.iterator.Event
}

func (x *mintControllerFulfillmentIterator) Close() error {
	return x.iterator.Close()
}

func (x *mintControllerFulfillmentIterator) Error() error {
	return x.iterator.Error()
}

type mintControllerContract struct {
	contract *autogen.MintController
	address  common.Address
}

func (x *mintControllerContract) Address() common.Address {
	return x.address
}

func (x *mintControllerContract) ParseFulfillment(log types.Log) (*autogen.MintControllerFulfillment, error) {
	return x.contract.ParseFulfillment(log)
}

func (x *mintControllerContract) FilterFulfillment(opts *bind.FilterOpts, orderID [][32]byte) (MintControllerFulfillmentIterator, error) {
	iterator, err := x.contract.FilterFulfillment(opts, orderID)
	if err != nil {
		return nil, err
	}
	return &mintControllerFulfillmentIterator{iterator: iterator}, nil
}

func (x *mintControllerContract) MaxMintLimit(opts *bind.CallOpts) (*big.Int, error) {
	return x.contract.MaxMintLimit(opts)
}

func NewMintControllerContract(address common.Address, client *ethclient.Client) (MintControllerContract, error) {
	contract, err := autogen.NewMintController(address, client)
	if err != nil {
		return nil, err
	}

	return &mintControllerContract{contract: contract, address: address}, nil
}
