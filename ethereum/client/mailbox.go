package client

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

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

type mailboxDispatchIterator struct {
	*autogen.MailboxDispatchIterator
}

func (m *mailboxDispatchIterator) Event() *autogen.MailboxDispatch {
	return m.MailboxDispatchIterator.Event
}

type mailboxContract struct {
	*autogen.Mailbox
	address common.Address
}

func (m *mailboxContract) Address() common.Address {
	return m.address
}

func (m *mailboxContract) FilterDispatch(opts *bind.FilterOpts, sender []common.Address, destination []uint32, recipient [][32]byte) (MailboxDispatchIterator, error) {
	iterator, err := m.Mailbox.FilterDispatch(opts, sender, destination, recipient)
	if err != nil {
		return nil, err
	}
	return &mailboxDispatchIterator{iterator}, nil
}

func NewMailboxContract(address common.Address, client bind.ContractBackend) (MailboxContract, error) {
	contract, err := autogen.NewMailbox(address, client)
	if err != nil {
		return nil, err
	}
	return &mailboxContract{Mailbox: contract, address: address}, nil
}
