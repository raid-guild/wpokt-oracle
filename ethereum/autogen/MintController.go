// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package autogen

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// MintControllerMetaData contains all meta data concerning the MintController contract.
var MintControllerMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"mailbox_\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"token_\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"ism_\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"defaultAdmin\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"newLimit_\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"newMintPerSecond_\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"DEFAULT_ADMIN_ROLE\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"MAIL_BOX_ROLE\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"currentMintLimit\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"fulfillOrder\",\"inputs\":[{\"name\":\"metadata\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"message\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"getRoleAdmin\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"grantRole\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"handle\",\"inputs\":[{\"name\":\"\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"_messageBody\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"hasRole\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"initiateOrder\",\"inputs\":[{\"name\":\"destinationDomain\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"recipientAddress\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"messageBody\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"interchainSecurityModule\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIInterchainSecurityModule\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"lastMint\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"lastMintLimit\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"maxMintLimit\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"mintPerSecond\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"renounceRole\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"revokeRole\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setIsm\",\"inputs\":[{\"name\":\"ism_\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"supportsInterface\",\"inputs\":[{\"name\":\"interfaceId\",\"type\":\"bytes4\",\"internalType\":\"bytes4\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"event\",\"name\":\"CurrentMintLimit\",\"inputs\":[{\"name\":\"limit\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"lastMint\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Fulfillment\",\"inputs\":[{\"name\":\"orderId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"message\",\"type\":\"bytes\",\"indexed\":false,\"internalType\":\"bytes\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"MintCooldownSet\",\"inputs\":[{\"name\":\"newLimit\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"newCooldown\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RoleAdminChanged\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"previousAdminRole\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"newAdminRole\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RoleGranted\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"sender\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RoleRevoked\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"sender\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"WarpMint\",\"inputs\":[{\"name\":\"recipient\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"sender\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"InvalidCooldownConfig\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"OverMintLimit\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"SenderMustBeCaller\",\"inputs\":[]}]",
}

// MintControllerABI is the input ABI used to generate the binding from.
// Deprecated: Use MintControllerMetaData.ABI instead.
var MintControllerABI = MintControllerMetaData.ABI

// MintController is an auto generated Go binding around an Ethereum contract.
type MintController struct {
	MintControllerCaller     // Read-only binding to the contract
	MintControllerTransactor // Write-only binding to the contract
	MintControllerFilterer   // Log filterer for contract events
}

// MintControllerCaller is an auto generated read-only Go binding around an Ethereum contract.
type MintControllerCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MintControllerTransactor is an auto generated write-only Go binding around an Ethereum contract.
type MintControllerTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MintControllerFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type MintControllerFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MintControllerSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type MintControllerSession struct {
	Contract     *MintController   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// MintControllerCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type MintControllerCallerSession struct {
	Contract *MintControllerCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// MintControllerTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type MintControllerTransactorSession struct {
	Contract     *MintControllerTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// MintControllerRaw is an auto generated low-level Go binding around an Ethereum contract.
type MintControllerRaw struct {
	Contract *MintController // Generic contract binding to access the raw methods on
}

// MintControllerCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type MintControllerCallerRaw struct {
	Contract *MintControllerCaller // Generic read-only contract binding to access the raw methods on
}

// MintControllerTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type MintControllerTransactorRaw struct {
	Contract *MintControllerTransactor // Generic write-only contract binding to access the raw methods on
}

// NewMintController creates a new instance of MintController, bound to a specific deployed contract.
func NewMintController(address common.Address, backend bind.ContractBackend) (*MintController, error) {
	contract, err := bindMintController(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &MintController{MintControllerCaller: MintControllerCaller{contract: contract}, MintControllerTransactor: MintControllerTransactor{contract: contract}, MintControllerFilterer: MintControllerFilterer{contract: contract}}, nil
}

// NewMintControllerCaller creates a new read-only instance of MintController, bound to a specific deployed contract.
func NewMintControllerCaller(address common.Address, caller bind.ContractCaller) (*MintControllerCaller, error) {
	contract, err := bindMintController(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &MintControllerCaller{contract: contract}, nil
}

// NewMintControllerTransactor creates a new write-only instance of MintController, bound to a specific deployed contract.
func NewMintControllerTransactor(address common.Address, transactor bind.ContractTransactor) (*MintControllerTransactor, error) {
	contract, err := bindMintController(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &MintControllerTransactor{contract: contract}, nil
}

// NewMintControllerFilterer creates a new log filterer instance of MintController, bound to a specific deployed contract.
func NewMintControllerFilterer(address common.Address, filterer bind.ContractFilterer) (*MintControllerFilterer, error) {
	contract, err := bindMintController(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &MintControllerFilterer{contract: contract}, nil
}

// bindMintController binds a generic wrapper to an already deployed contract.
func bindMintController(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := MintControllerMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MintController *MintControllerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _MintController.Contract.MintControllerCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MintController *MintControllerRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MintController.Contract.MintControllerTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MintController *MintControllerRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MintController.Contract.MintControllerTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MintController *MintControllerCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _MintController.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MintController *MintControllerTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MintController.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MintController *MintControllerTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MintController.Contract.contract.Transact(opts, method, params...)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_MintController *MintControllerCaller) DEFAULTADMINROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _MintController.contract.Call(opts, &out, "DEFAULT_ADMIN_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_MintController *MintControllerSession) DEFAULTADMINROLE() ([32]byte, error) {
	return _MintController.Contract.DEFAULTADMINROLE(&_MintController.CallOpts)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_MintController *MintControllerCallerSession) DEFAULTADMINROLE() ([32]byte, error) {
	return _MintController.Contract.DEFAULTADMINROLE(&_MintController.CallOpts)
}

// MAILBOXROLE is a free data retrieval call binding the contract method 0xe1787b90.
//
// Solidity: function MAIL_BOX_ROLE() view returns(bytes32)
func (_MintController *MintControllerCaller) MAILBOXROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _MintController.contract.Call(opts, &out, "MAIL_BOX_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// MAILBOXROLE is a free data retrieval call binding the contract method 0xe1787b90.
//
// Solidity: function MAIL_BOX_ROLE() view returns(bytes32)
func (_MintController *MintControllerSession) MAILBOXROLE() ([32]byte, error) {
	return _MintController.Contract.MAILBOXROLE(&_MintController.CallOpts)
}

// MAILBOXROLE is a free data retrieval call binding the contract method 0xe1787b90.
//
// Solidity: function MAIL_BOX_ROLE() view returns(bytes32)
func (_MintController *MintControllerCallerSession) MAILBOXROLE() ([32]byte, error) {
	return _MintController.Contract.MAILBOXROLE(&_MintController.CallOpts)
}

// CurrentMintLimit is a free data retrieval call binding the contract method 0xa40c3eab.
//
// Solidity: function currentMintLimit() view returns(uint256)
func (_MintController *MintControllerCaller) CurrentMintLimit(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _MintController.contract.Call(opts, &out, "currentMintLimit")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CurrentMintLimit is a free data retrieval call binding the contract method 0xa40c3eab.
//
// Solidity: function currentMintLimit() view returns(uint256)
func (_MintController *MintControllerSession) CurrentMintLimit() (*big.Int, error) {
	return _MintController.Contract.CurrentMintLimit(&_MintController.CallOpts)
}

// CurrentMintLimit is a free data retrieval call binding the contract method 0xa40c3eab.
//
// Solidity: function currentMintLimit() view returns(uint256)
func (_MintController *MintControllerCallerSession) CurrentMintLimit() (*big.Int, error) {
	return _MintController.Contract.CurrentMintLimit(&_MintController.CallOpts)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_MintController *MintControllerCaller) GetRoleAdmin(opts *bind.CallOpts, role [32]byte) ([32]byte, error) {
	var out []interface{}
	err := _MintController.contract.Call(opts, &out, "getRoleAdmin", role)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_MintController *MintControllerSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _MintController.Contract.GetRoleAdmin(&_MintController.CallOpts, role)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_MintController *MintControllerCallerSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _MintController.Contract.GetRoleAdmin(&_MintController.CallOpts, role)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_MintController *MintControllerCaller) HasRole(opts *bind.CallOpts, role [32]byte, account common.Address) (bool, error) {
	var out []interface{}
	err := _MintController.contract.Call(opts, &out, "hasRole", role, account)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_MintController *MintControllerSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _MintController.Contract.HasRole(&_MintController.CallOpts, role, account)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_MintController *MintControllerCallerSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _MintController.Contract.HasRole(&_MintController.CallOpts, role, account)
}

// InterchainSecurityModule is a free data retrieval call binding the contract method 0xde523cf3.
//
// Solidity: function interchainSecurityModule() view returns(address)
func (_MintController *MintControllerCaller) InterchainSecurityModule(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _MintController.contract.Call(opts, &out, "interchainSecurityModule")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// InterchainSecurityModule is a free data retrieval call binding the contract method 0xde523cf3.
//
// Solidity: function interchainSecurityModule() view returns(address)
func (_MintController *MintControllerSession) InterchainSecurityModule() (common.Address, error) {
	return _MintController.Contract.InterchainSecurityModule(&_MintController.CallOpts)
}

// InterchainSecurityModule is a free data retrieval call binding the contract method 0xde523cf3.
//
// Solidity: function interchainSecurityModule() view returns(address)
func (_MintController *MintControllerCallerSession) InterchainSecurityModule() (common.Address, error) {
	return _MintController.Contract.InterchainSecurityModule(&_MintController.CallOpts)
}

// LastMint is a free data retrieval call binding the contract method 0x586fc5b5.
//
// Solidity: function lastMint() view returns(uint256)
func (_MintController *MintControllerCaller) LastMint(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _MintController.contract.Call(opts, &out, "lastMint")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// LastMint is a free data retrieval call binding the contract method 0x586fc5b5.
//
// Solidity: function lastMint() view returns(uint256)
func (_MintController *MintControllerSession) LastMint() (*big.Int, error) {
	return _MintController.Contract.LastMint(&_MintController.CallOpts)
}

// LastMint is a free data retrieval call binding the contract method 0x586fc5b5.
//
// Solidity: function lastMint() view returns(uint256)
func (_MintController *MintControllerCallerSession) LastMint() (*big.Int, error) {
	return _MintController.Contract.LastMint(&_MintController.CallOpts)
}

// LastMintLimit is a free data retrieval call binding the contract method 0x2f99582a.
//
// Solidity: function lastMintLimit() view returns(uint256)
func (_MintController *MintControllerCaller) LastMintLimit(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _MintController.contract.Call(opts, &out, "lastMintLimit")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// LastMintLimit is a free data retrieval call binding the contract method 0x2f99582a.
//
// Solidity: function lastMintLimit() view returns(uint256)
func (_MintController *MintControllerSession) LastMintLimit() (*big.Int, error) {
	return _MintController.Contract.LastMintLimit(&_MintController.CallOpts)
}

// LastMintLimit is a free data retrieval call binding the contract method 0x2f99582a.
//
// Solidity: function lastMintLimit() view returns(uint256)
func (_MintController *MintControllerCallerSession) LastMintLimit() (*big.Int, error) {
	return _MintController.Contract.LastMintLimit(&_MintController.CallOpts)
}

// MaxMintLimit is a free data retrieval call binding the contract method 0x70e2f827.
//
// Solidity: function maxMintLimit() view returns(uint256)
func (_MintController *MintControllerCaller) MaxMintLimit(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _MintController.contract.Call(opts, &out, "maxMintLimit")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MaxMintLimit is a free data retrieval call binding the contract method 0x70e2f827.
//
// Solidity: function maxMintLimit() view returns(uint256)
func (_MintController *MintControllerSession) MaxMintLimit() (*big.Int, error) {
	return _MintController.Contract.MaxMintLimit(&_MintController.CallOpts)
}

// MaxMintLimit is a free data retrieval call binding the contract method 0x70e2f827.
//
// Solidity: function maxMintLimit() view returns(uint256)
func (_MintController *MintControllerCallerSession) MaxMintLimit() (*big.Int, error) {
	return _MintController.Contract.MaxMintLimit(&_MintController.CallOpts)
}

// MintPerSecond is a free data retrieval call binding the contract method 0x272c444d.
//
// Solidity: function mintPerSecond() view returns(uint256)
func (_MintController *MintControllerCaller) MintPerSecond(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _MintController.contract.Call(opts, &out, "mintPerSecond")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MintPerSecond is a free data retrieval call binding the contract method 0x272c444d.
//
// Solidity: function mintPerSecond() view returns(uint256)
func (_MintController *MintControllerSession) MintPerSecond() (*big.Int, error) {
	return _MintController.Contract.MintPerSecond(&_MintController.CallOpts)
}

// MintPerSecond is a free data retrieval call binding the contract method 0x272c444d.
//
// Solidity: function mintPerSecond() view returns(uint256)
func (_MintController *MintControllerCallerSession) MintPerSecond() (*big.Int, error) {
	return _MintController.Contract.MintPerSecond(&_MintController.CallOpts)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_MintController *MintControllerCaller) SupportsInterface(opts *bind.CallOpts, interfaceId [4]byte) (bool, error) {
	var out []interface{}
	err := _MintController.contract.Call(opts, &out, "supportsInterface", interfaceId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_MintController *MintControllerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _MintController.Contract.SupportsInterface(&_MintController.CallOpts, interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_MintController *MintControllerCallerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _MintController.Contract.SupportsInterface(&_MintController.CallOpts, interfaceId)
}

// FulfillOrder is a paid mutator transaction binding the contract method 0xf6853601.
//
// Solidity: function fulfillOrder(bytes metadata, bytes message) returns()
func (_MintController *MintControllerTransactor) FulfillOrder(opts *bind.TransactOpts, metadata []byte, message []byte) (*types.Transaction, error) {
	return _MintController.contract.Transact(opts, "fulfillOrder", metadata, message)
}

// FulfillOrder is a paid mutator transaction binding the contract method 0xf6853601.
//
// Solidity: function fulfillOrder(bytes metadata, bytes message) returns()
func (_MintController *MintControllerSession) FulfillOrder(metadata []byte, message []byte) (*types.Transaction, error) {
	return _MintController.Contract.FulfillOrder(&_MintController.TransactOpts, metadata, message)
}

// FulfillOrder is a paid mutator transaction binding the contract method 0xf6853601.
//
// Solidity: function fulfillOrder(bytes metadata, bytes message) returns()
func (_MintController *MintControllerTransactorSession) FulfillOrder(metadata []byte, message []byte) (*types.Transaction, error) {
	return _MintController.Contract.FulfillOrder(&_MintController.TransactOpts, metadata, message)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_MintController *MintControllerTransactor) GrantRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _MintController.contract.Transact(opts, "grantRole", role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_MintController *MintControllerSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _MintController.Contract.GrantRole(&_MintController.TransactOpts, role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_MintController *MintControllerTransactorSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _MintController.Contract.GrantRole(&_MintController.TransactOpts, role, account)
}

// Handle is a paid mutator transaction binding the contract method 0x56d5d475.
//
// Solidity: function handle(uint32 , bytes32 , bytes _messageBody) returns()
func (_MintController *MintControllerTransactor) Handle(opts *bind.TransactOpts, arg0 uint32, arg1 [32]byte, _messageBody []byte) (*types.Transaction, error) {
	return _MintController.contract.Transact(opts, "handle", arg0, arg1, _messageBody)
}

// Handle is a paid mutator transaction binding the contract method 0x56d5d475.
//
// Solidity: function handle(uint32 , bytes32 , bytes _messageBody) returns()
func (_MintController *MintControllerSession) Handle(arg0 uint32, arg1 [32]byte, _messageBody []byte) (*types.Transaction, error) {
	return _MintController.Contract.Handle(&_MintController.TransactOpts, arg0, arg1, _messageBody)
}

// Handle is a paid mutator transaction binding the contract method 0x56d5d475.
//
// Solidity: function handle(uint32 , bytes32 , bytes _messageBody) returns()
func (_MintController *MintControllerTransactorSession) Handle(arg0 uint32, arg1 [32]byte, _messageBody []byte) (*types.Transaction, error) {
	return _MintController.Contract.Handle(&_MintController.TransactOpts, arg0, arg1, _messageBody)
}

// InitiateOrder is a paid mutator transaction binding the contract method 0x9179f6a0.
//
// Solidity: function initiateOrder(uint32 destinationDomain, bytes32 recipientAddress, bytes messageBody) returns()
func (_MintController *MintControllerTransactor) InitiateOrder(opts *bind.TransactOpts, destinationDomain uint32, recipientAddress [32]byte, messageBody []byte) (*types.Transaction, error) {
	return _MintController.contract.Transact(opts, "initiateOrder", destinationDomain, recipientAddress, messageBody)
}

// InitiateOrder is a paid mutator transaction binding the contract method 0x9179f6a0.
//
// Solidity: function initiateOrder(uint32 destinationDomain, bytes32 recipientAddress, bytes messageBody) returns()
func (_MintController *MintControllerSession) InitiateOrder(destinationDomain uint32, recipientAddress [32]byte, messageBody []byte) (*types.Transaction, error) {
	return _MintController.Contract.InitiateOrder(&_MintController.TransactOpts, destinationDomain, recipientAddress, messageBody)
}

// InitiateOrder is a paid mutator transaction binding the contract method 0x9179f6a0.
//
// Solidity: function initiateOrder(uint32 destinationDomain, bytes32 recipientAddress, bytes messageBody) returns()
func (_MintController *MintControllerTransactorSession) InitiateOrder(destinationDomain uint32, recipientAddress [32]byte, messageBody []byte) (*types.Transaction, error) {
	return _MintController.Contract.InitiateOrder(&_MintController.TransactOpts, destinationDomain, recipientAddress, messageBody)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address account) returns()
func (_MintController *MintControllerTransactor) RenounceRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _MintController.contract.Transact(opts, "renounceRole", role, account)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address account) returns()
func (_MintController *MintControllerSession) RenounceRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _MintController.Contract.RenounceRole(&_MintController.TransactOpts, role, account)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address account) returns()
func (_MintController *MintControllerTransactorSession) RenounceRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _MintController.Contract.RenounceRole(&_MintController.TransactOpts, role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_MintController *MintControllerTransactor) RevokeRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _MintController.contract.Transact(opts, "revokeRole", role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_MintController *MintControllerSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _MintController.Contract.RevokeRole(&_MintController.TransactOpts, role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_MintController *MintControllerTransactorSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _MintController.Contract.RevokeRole(&_MintController.TransactOpts, role, account)
}

// SetIsm is a paid mutator transaction binding the contract method 0x2e779beb.
//
// Solidity: function setIsm(address ism_) returns()
func (_MintController *MintControllerTransactor) SetIsm(opts *bind.TransactOpts, ism_ common.Address) (*types.Transaction, error) {
	return _MintController.contract.Transact(opts, "setIsm", ism_)
}

// SetIsm is a paid mutator transaction binding the contract method 0x2e779beb.
//
// Solidity: function setIsm(address ism_) returns()
func (_MintController *MintControllerSession) SetIsm(ism_ common.Address) (*types.Transaction, error) {
	return _MintController.Contract.SetIsm(&_MintController.TransactOpts, ism_)
}

// SetIsm is a paid mutator transaction binding the contract method 0x2e779beb.
//
// Solidity: function setIsm(address ism_) returns()
func (_MintController *MintControllerTransactorSession) SetIsm(ism_ common.Address) (*types.Transaction, error) {
	return _MintController.Contract.SetIsm(&_MintController.TransactOpts, ism_)
}

// MintControllerCurrentMintLimitIterator is returned from FilterCurrentMintLimit and is used to iterate over the raw logs and unpacked data for CurrentMintLimit events raised by the MintController contract.
type MintControllerCurrentMintLimitIterator struct {
	Event *MintControllerCurrentMintLimit // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MintControllerCurrentMintLimitIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MintControllerCurrentMintLimit)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MintControllerCurrentMintLimit)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MintControllerCurrentMintLimitIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MintControllerCurrentMintLimitIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MintControllerCurrentMintLimit represents a CurrentMintLimit event raised by the MintController contract.
type MintControllerCurrentMintLimit struct {
	Limit    *big.Int
	LastMint *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterCurrentMintLimit is a free log retrieval operation binding the contract event 0x37ce17c2014463ff310cd0d994a30795f5c48b6d6c3104ba83677f096383e3e4.
//
// Solidity: event CurrentMintLimit(uint256 indexed limit, uint256 indexed lastMint)
func (_MintController *MintControllerFilterer) FilterCurrentMintLimit(opts *bind.FilterOpts, limit []*big.Int, lastMint []*big.Int) (*MintControllerCurrentMintLimitIterator, error) {

	var limitRule []interface{}
	for _, limitItem := range limit {
		limitRule = append(limitRule, limitItem)
	}
	var lastMintRule []interface{}
	for _, lastMintItem := range lastMint {
		lastMintRule = append(lastMintRule, lastMintItem)
	}

	logs, sub, err := _MintController.contract.FilterLogs(opts, "CurrentMintLimit", limitRule, lastMintRule)
	if err != nil {
		return nil, err
	}
	return &MintControllerCurrentMintLimitIterator{contract: _MintController.contract, event: "CurrentMintLimit", logs: logs, sub: sub}, nil
}

// WatchCurrentMintLimit is a free log subscription operation binding the contract event 0x37ce17c2014463ff310cd0d994a30795f5c48b6d6c3104ba83677f096383e3e4.
//
// Solidity: event CurrentMintLimit(uint256 indexed limit, uint256 indexed lastMint)
func (_MintController *MintControllerFilterer) WatchCurrentMintLimit(opts *bind.WatchOpts, sink chan<- *MintControllerCurrentMintLimit, limit []*big.Int, lastMint []*big.Int) (event.Subscription, error) {

	var limitRule []interface{}
	for _, limitItem := range limit {
		limitRule = append(limitRule, limitItem)
	}
	var lastMintRule []interface{}
	for _, lastMintItem := range lastMint {
		lastMintRule = append(lastMintRule, lastMintItem)
	}

	logs, sub, err := _MintController.contract.WatchLogs(opts, "CurrentMintLimit", limitRule, lastMintRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MintControllerCurrentMintLimit)
				if err := _MintController.contract.UnpackLog(event, "CurrentMintLimit", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseCurrentMintLimit is a log parse operation binding the contract event 0x37ce17c2014463ff310cd0d994a30795f5c48b6d6c3104ba83677f096383e3e4.
//
// Solidity: event CurrentMintLimit(uint256 indexed limit, uint256 indexed lastMint)
func (_MintController *MintControllerFilterer) ParseCurrentMintLimit(log types.Log) (*MintControllerCurrentMintLimit, error) {
	event := new(MintControllerCurrentMintLimit)
	if err := _MintController.contract.UnpackLog(event, "CurrentMintLimit", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MintControllerFulfillmentIterator is returned from FilterFulfillment and is used to iterate over the raw logs and unpacked data for Fulfillment events raised by the MintController contract.
type MintControllerFulfillmentIterator struct {
	Event *MintControllerFulfillment // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MintControllerFulfillmentIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MintControllerFulfillment)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MintControllerFulfillment)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MintControllerFulfillmentIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MintControllerFulfillmentIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MintControllerFulfillment represents a Fulfillment event raised by the MintController contract.
type MintControllerFulfillment struct {
	OrderId [32]byte
	Message []byte
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterFulfillment is a free log retrieval operation binding the contract event 0x21f27e3543480a9af10f107929b1caf9d017fe7282b1c3d4e8c6960b5464d2d1.
//
// Solidity: event Fulfillment(bytes32 indexed orderId, bytes message)
func (_MintController *MintControllerFilterer) FilterFulfillment(opts *bind.FilterOpts, orderId [][32]byte) (*MintControllerFulfillmentIterator, error) {

	var orderIdRule []interface{}
	for _, orderIdItem := range orderId {
		orderIdRule = append(orderIdRule, orderIdItem)
	}

	logs, sub, err := _MintController.contract.FilterLogs(opts, "Fulfillment", orderIdRule)
	if err != nil {
		return nil, err
	}
	return &MintControllerFulfillmentIterator{contract: _MintController.contract, event: "Fulfillment", logs: logs, sub: sub}, nil
}

// WatchFulfillment is a free log subscription operation binding the contract event 0x21f27e3543480a9af10f107929b1caf9d017fe7282b1c3d4e8c6960b5464d2d1.
//
// Solidity: event Fulfillment(bytes32 indexed orderId, bytes message)
func (_MintController *MintControllerFilterer) WatchFulfillment(opts *bind.WatchOpts, sink chan<- *MintControllerFulfillment, orderId [][32]byte) (event.Subscription, error) {

	var orderIdRule []interface{}
	for _, orderIdItem := range orderId {
		orderIdRule = append(orderIdRule, orderIdItem)
	}

	logs, sub, err := _MintController.contract.WatchLogs(opts, "Fulfillment", orderIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MintControllerFulfillment)
				if err := _MintController.contract.UnpackLog(event, "Fulfillment", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseFulfillment is a log parse operation binding the contract event 0x21f27e3543480a9af10f107929b1caf9d017fe7282b1c3d4e8c6960b5464d2d1.
//
// Solidity: event Fulfillment(bytes32 indexed orderId, bytes message)
func (_MintController *MintControllerFilterer) ParseFulfillment(log types.Log) (*MintControllerFulfillment, error) {
	event := new(MintControllerFulfillment)
	if err := _MintController.contract.UnpackLog(event, "Fulfillment", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MintControllerMintCooldownSetIterator is returned from FilterMintCooldownSet and is used to iterate over the raw logs and unpacked data for MintCooldownSet events raised by the MintController contract.
type MintControllerMintCooldownSetIterator struct {
	Event *MintControllerMintCooldownSet // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MintControllerMintCooldownSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MintControllerMintCooldownSet)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MintControllerMintCooldownSet)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MintControllerMintCooldownSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MintControllerMintCooldownSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MintControllerMintCooldownSet represents a MintCooldownSet event raised by the MintController contract.
type MintControllerMintCooldownSet struct {
	NewLimit    *big.Int
	NewCooldown *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterMintCooldownSet is a free log retrieval operation binding the contract event 0x2c1471be9f5586272d41bc8d7d0ce89b3b10e4693755e30752510cb88d42fdf5.
//
// Solidity: event MintCooldownSet(uint256 newLimit, uint256 newCooldown)
func (_MintController *MintControllerFilterer) FilterMintCooldownSet(opts *bind.FilterOpts) (*MintControllerMintCooldownSetIterator, error) {

	logs, sub, err := _MintController.contract.FilterLogs(opts, "MintCooldownSet")
	if err != nil {
		return nil, err
	}
	return &MintControllerMintCooldownSetIterator{contract: _MintController.contract, event: "MintCooldownSet", logs: logs, sub: sub}, nil
}

// WatchMintCooldownSet is a free log subscription operation binding the contract event 0x2c1471be9f5586272d41bc8d7d0ce89b3b10e4693755e30752510cb88d42fdf5.
//
// Solidity: event MintCooldownSet(uint256 newLimit, uint256 newCooldown)
func (_MintController *MintControllerFilterer) WatchMintCooldownSet(opts *bind.WatchOpts, sink chan<- *MintControllerMintCooldownSet) (event.Subscription, error) {

	logs, sub, err := _MintController.contract.WatchLogs(opts, "MintCooldownSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MintControllerMintCooldownSet)
				if err := _MintController.contract.UnpackLog(event, "MintCooldownSet", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseMintCooldownSet is a log parse operation binding the contract event 0x2c1471be9f5586272d41bc8d7d0ce89b3b10e4693755e30752510cb88d42fdf5.
//
// Solidity: event MintCooldownSet(uint256 newLimit, uint256 newCooldown)
func (_MintController *MintControllerFilterer) ParseMintCooldownSet(log types.Log) (*MintControllerMintCooldownSet, error) {
	event := new(MintControllerMintCooldownSet)
	if err := _MintController.contract.UnpackLog(event, "MintCooldownSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MintControllerRoleAdminChangedIterator is returned from FilterRoleAdminChanged and is used to iterate over the raw logs and unpacked data for RoleAdminChanged events raised by the MintController contract.
type MintControllerRoleAdminChangedIterator struct {
	Event *MintControllerRoleAdminChanged // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MintControllerRoleAdminChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MintControllerRoleAdminChanged)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MintControllerRoleAdminChanged)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MintControllerRoleAdminChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MintControllerRoleAdminChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MintControllerRoleAdminChanged represents a RoleAdminChanged event raised by the MintController contract.
type MintControllerRoleAdminChanged struct {
	Role              [32]byte
	PreviousAdminRole [32]byte
	NewAdminRole      [32]byte
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterRoleAdminChanged is a free log retrieval operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_MintController *MintControllerFilterer) FilterRoleAdminChanged(opts *bind.FilterOpts, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (*MintControllerRoleAdminChangedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var previousAdminRoleRule []interface{}
	for _, previousAdminRoleItem := range previousAdminRole {
		previousAdminRoleRule = append(previousAdminRoleRule, previousAdminRoleItem)
	}
	var newAdminRoleRule []interface{}
	for _, newAdminRoleItem := range newAdminRole {
		newAdminRoleRule = append(newAdminRoleRule, newAdminRoleItem)
	}

	logs, sub, err := _MintController.contract.FilterLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return &MintControllerRoleAdminChangedIterator{contract: _MintController.contract, event: "RoleAdminChanged", logs: logs, sub: sub}, nil
}

// WatchRoleAdminChanged is a free log subscription operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_MintController *MintControllerFilterer) WatchRoleAdminChanged(opts *bind.WatchOpts, sink chan<- *MintControllerRoleAdminChanged, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var previousAdminRoleRule []interface{}
	for _, previousAdminRoleItem := range previousAdminRole {
		previousAdminRoleRule = append(previousAdminRoleRule, previousAdminRoleItem)
	}
	var newAdminRoleRule []interface{}
	for _, newAdminRoleItem := range newAdminRole {
		newAdminRoleRule = append(newAdminRoleRule, newAdminRoleItem)
	}

	logs, sub, err := _MintController.contract.WatchLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MintControllerRoleAdminChanged)
				if err := _MintController.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRoleAdminChanged is a log parse operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_MintController *MintControllerFilterer) ParseRoleAdminChanged(log types.Log) (*MintControllerRoleAdminChanged, error) {
	event := new(MintControllerRoleAdminChanged)
	if err := _MintController.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MintControllerRoleGrantedIterator is returned from FilterRoleGranted and is used to iterate over the raw logs and unpacked data for RoleGranted events raised by the MintController contract.
type MintControllerRoleGrantedIterator struct {
	Event *MintControllerRoleGranted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MintControllerRoleGrantedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MintControllerRoleGranted)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MintControllerRoleGranted)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MintControllerRoleGrantedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MintControllerRoleGrantedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MintControllerRoleGranted represents a RoleGranted event raised by the MintController contract.
type MintControllerRoleGranted struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleGranted is a free log retrieval operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_MintController *MintControllerFilterer) FilterRoleGranted(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*MintControllerRoleGrantedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _MintController.contract.FilterLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &MintControllerRoleGrantedIterator{contract: _MintController.contract, event: "RoleGranted", logs: logs, sub: sub}, nil
}

// WatchRoleGranted is a free log subscription operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_MintController *MintControllerFilterer) WatchRoleGranted(opts *bind.WatchOpts, sink chan<- *MintControllerRoleGranted, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _MintController.contract.WatchLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MintControllerRoleGranted)
				if err := _MintController.contract.UnpackLog(event, "RoleGranted", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRoleGranted is a log parse operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_MintController *MintControllerFilterer) ParseRoleGranted(log types.Log) (*MintControllerRoleGranted, error) {
	event := new(MintControllerRoleGranted)
	if err := _MintController.contract.UnpackLog(event, "RoleGranted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MintControllerRoleRevokedIterator is returned from FilterRoleRevoked and is used to iterate over the raw logs and unpacked data for RoleRevoked events raised by the MintController contract.
type MintControllerRoleRevokedIterator struct {
	Event *MintControllerRoleRevoked // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MintControllerRoleRevokedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MintControllerRoleRevoked)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MintControllerRoleRevoked)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MintControllerRoleRevokedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MintControllerRoleRevokedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MintControllerRoleRevoked represents a RoleRevoked event raised by the MintController contract.
type MintControllerRoleRevoked struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleRevoked is a free log retrieval operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_MintController *MintControllerFilterer) FilterRoleRevoked(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*MintControllerRoleRevokedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _MintController.contract.FilterLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &MintControllerRoleRevokedIterator{contract: _MintController.contract, event: "RoleRevoked", logs: logs, sub: sub}, nil
}

// WatchRoleRevoked is a free log subscription operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_MintController *MintControllerFilterer) WatchRoleRevoked(opts *bind.WatchOpts, sink chan<- *MintControllerRoleRevoked, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _MintController.contract.WatchLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MintControllerRoleRevoked)
				if err := _MintController.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRoleRevoked is a log parse operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_MintController *MintControllerFilterer) ParseRoleRevoked(log types.Log) (*MintControllerRoleRevoked, error) {
	event := new(MintControllerRoleRevoked)
	if err := _MintController.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MintControllerWarpMintIterator is returned from FilterWarpMint and is used to iterate over the raw logs and unpacked data for WarpMint events raised by the MintController contract.
type MintControllerWarpMintIterator struct {
	Event *MintControllerWarpMint // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MintControllerWarpMintIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MintControllerWarpMint)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MintControllerWarpMint)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MintControllerWarpMintIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MintControllerWarpMintIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MintControllerWarpMint represents a WarpMint event raised by the MintController contract.
type MintControllerWarpMint struct {
	Recipient common.Address
	Amount    *big.Int
	Sender    common.Address
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterWarpMint is a free log retrieval operation binding the contract event 0xe8ed11e08f93f4439f5be82aa37d56804a8081a0094b55f98c221ffacb522619.
//
// Solidity: event WarpMint(address indexed recipient, uint256 amount, address indexed sender)
func (_MintController *MintControllerFilterer) FilterWarpMint(opts *bind.FilterOpts, recipient []common.Address, sender []common.Address) (*MintControllerWarpMintIterator, error) {

	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}

	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _MintController.contract.FilterLogs(opts, "WarpMint", recipientRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &MintControllerWarpMintIterator{contract: _MintController.contract, event: "WarpMint", logs: logs, sub: sub}, nil
}

// WatchWarpMint is a free log subscription operation binding the contract event 0xe8ed11e08f93f4439f5be82aa37d56804a8081a0094b55f98c221ffacb522619.
//
// Solidity: event WarpMint(address indexed recipient, uint256 amount, address indexed sender)
func (_MintController *MintControllerFilterer) WatchWarpMint(opts *bind.WatchOpts, sink chan<- *MintControllerWarpMint, recipient []common.Address, sender []common.Address) (event.Subscription, error) {

	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}

	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _MintController.contract.WatchLogs(opts, "WarpMint", recipientRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MintControllerWarpMint)
				if err := _MintController.contract.UnpackLog(event, "WarpMint", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseWarpMint is a log parse operation binding the contract event 0xe8ed11e08f93f4439f5be82aa37d56804a8081a0094b55f98c221ffacb522619.
//
// Solidity: event WarpMint(address indexed recipient, uint256 amount, address indexed sender)
func (_MintController *MintControllerFilterer) ParseWarpMint(log types.Log) (*MintControllerWarpMint, error) {
	event := new(MintControllerWarpMint)
	if err := _MintController.contract.UnpackLog(event, "WarpMint", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
