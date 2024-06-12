package client

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/dan13ram/wpokt-oracle/ethereum/autogen"
)

type MailboxContract interface {
	Address() common.Address
	FilterDispatch(opts *bind.FilterOpts, sender []common.Address, destination []uint32, recipient [][32]byte) (MailboxDispatchIterator, error)
	ParseDispatch(log types.Log) (*autogen.MailboxDispatch, error)
	ParseDispatchId(log types.Log) (*autogen.MailboxDispatchId, error)
}

type MailboxDispatchIterator interface {
	Next() bool
	Event() *autogen.MailboxDispatch
	Close() error
	Error() error
}

type MailboxDispatchIteratorImpl struct {
	iterator *autogen.MailboxDispatchIterator
}

func (x *MailboxDispatchIteratorImpl) Next() bool {
	return x.iterator.Next()
}

func (x *MailboxDispatchIteratorImpl) Event() *autogen.MailboxDispatch {
	return x.iterator.Event
}

func (x *MailboxDispatchIteratorImpl) Close() error {
	return x.iterator.Close()
}

func (x *MailboxDispatchIteratorImpl) Error() error {
	return x.iterator.Error()
}

type MailboxContractImpl struct {
	contract *autogen.Mailbox
	address  common.Address
}

func (x *MailboxContractImpl) ParseDispatch(log types.Log) (*autogen.MailboxDispatch, error) {
	return x.contract.ParseDispatch(log)
}

func (x *MailboxContractImpl) ParseDispatchId(log types.Log) (*autogen.MailboxDispatchId, error) {
	return x.contract.ParseDispatchId(log)
}

func (x *MailboxContractImpl) Address() common.Address {
	return x.address
}

func (x *MailboxContractImpl) FilterDispatch(opts *bind.FilterOpts, sender []common.Address, destination []uint32, recipient [][32]byte) (MailboxDispatchIterator, error) {
	iterator, err := x.contract.FilterDispatch(opts, sender, destination, recipient)
	if err != nil {
		return nil, err
	}
	return &MailboxDispatchIteratorImpl{iterator: iterator}, nil
}

func NewMailboxContract(address common.Address, client *ethclient.Client) (MailboxContract, error) {
	contract, err := autogen.NewMailbox(address, client)
	if err != nil {
		return nil, err
	}

	return &MailboxContractImpl{contract: contract, address: address}, nil
}
