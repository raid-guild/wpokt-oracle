package client

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/dan13ram/wpokt-oracle/ethereum/autogen"
)

type MailboxContract interface {
	FilterDispatch(opts *bind.FilterOpts, sender []common.Address, destination []uint32, recipient [][32]byte) (MailboxDispatchIterator, error)
	ParseDispatch(log types.Log) (*autogen.MailboxDispatch, error)
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
}

func (x *MailboxContractImpl) ParseDispatch(log types.Log) (*autogen.MailboxDispatch, error) {
	return x.contract.ParseDispatch(log)
}

func (x *MailboxContractImpl) FilterDispatch(opts *bind.FilterOpts, sender []common.Address, destination []uint32, recipient [][32]byte) (MailboxDispatchIterator, error) {
	iterator, err := x.contract.FilterDispatch(opts, sender, destination, recipient)
	if err != nil {
		return nil, err
	}
	return &MailboxDispatchIteratorImpl{iterator: iterator}, nil
}

func NewMailboxContract(contract *autogen.Mailbox) MailboxContract {
	return &MailboxContractImpl{contract: contract}
}
