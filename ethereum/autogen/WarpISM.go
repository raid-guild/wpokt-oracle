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

// WarpISMMetaData contains all meta data concerning the WarpISM contract.
var WarpISMMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"name_\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"version_\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"initialOwner_\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"DIGEST_TYPE_HASH\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"SIGNATURE_SIZE\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"addValidator\",\"inputs\":[{\"name\":\"validator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"eip712Domain\",\"inputs\":[],\"outputs\":[{\"name\":\"fields\",\"type\":\"bytes1\",\"internalType\":\"bytes1\"},{\"name\":\"name\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"version\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"chainId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"verifyingContract\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"salt\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"extensions\",\"type\":\"uint256[]\",\"internalType\":\"uint256[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getDigest\",\"inputs\":[{\"name\":\"message\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"digest\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getSignatures\",\"inputs\":[{\"name\":\"metadata\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"signatures\",\"type\":\"bytes[]\",\"internalType\":\"bytes[]\"}],\"stateMutability\":\"pure\"},{\"type\":\"function\",\"name\":\"owner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"removeValidator\",\"inputs\":[{\"name\":\"validator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"renounceOwnership\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setSignerThreshold\",\"inputs\":[{\"name\":\"signatureRatio\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"signerThreshold\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"transferOwnership\",\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"validatorCount\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"validators\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"verify\",\"inputs\":[{\"name\":\"metadata\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"message\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"success\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"event\",\"name\":\"EIP712DomainChanged\",\"inputs\":[],\"anonymous\":false},{\"type\":\"event\",\"name\":\"NewValidator\",\"inputs\":[{\"name\":\"validator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipTransferred\",\"inputs\":[{\"name\":\"previousOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RemovedValidator\",\"inputs\":[{\"name\":\"validator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SignerThresholdSet\",\"inputs\":[{\"name\":\"ratio\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"BelowMinThreshold\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ECDSAInvalidSignature\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ECDSAInvalidSignatureLength\",\"inputs\":[{\"name\":\"length\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"ECDSAInvalidSignatureS\",\"inputs\":[{\"name\":\"s\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"InvalidAddValidator\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidRemoveValidator\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidShortString\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidSignatureLength\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidSignatureRatio\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidSignatures\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NonZero\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"OwnableInvalidOwner\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"OwnableUnauthorizedAccount\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"StringTooLong\",\"inputs\":[{\"name\":\"str\",\"type\":\"string\",\"internalType\":\"string\"}]}]",
}

// WarpISMABI is the input ABI used to generate the binding from.
// Deprecated: Use WarpISMMetaData.ABI instead.
var WarpISMABI = WarpISMMetaData.ABI

// WarpISM is an auto generated Go binding around an Ethereum contract.
type WarpISM struct {
	WarpISMCaller     // Read-only binding to the contract
	WarpISMTransactor // Write-only binding to the contract
	WarpISMFilterer   // Log filterer for contract events
}

// WarpISMCaller is an auto generated read-only Go binding around an Ethereum contract.
type WarpISMCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// WarpISMTransactor is an auto generated write-only Go binding around an Ethereum contract.
type WarpISMTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// WarpISMFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type WarpISMFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// WarpISMSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type WarpISMSession struct {
	Contract     *WarpISM          // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// WarpISMCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type WarpISMCallerSession struct {
	Contract *WarpISMCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts  // Call options to use throughout this session
}

// WarpISMTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type WarpISMTransactorSession struct {
	Contract     *WarpISMTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// WarpISMRaw is an auto generated low-level Go binding around an Ethereum contract.
type WarpISMRaw struct {
	Contract *WarpISM // Generic contract binding to access the raw methods on
}

// WarpISMCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type WarpISMCallerRaw struct {
	Contract *WarpISMCaller // Generic read-only contract binding to access the raw methods on
}

// WarpISMTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type WarpISMTransactorRaw struct {
	Contract *WarpISMTransactor // Generic write-only contract binding to access the raw methods on
}

// NewWarpISM creates a new instance of WarpISM, bound to a specific deployed contract.
func NewWarpISM(address common.Address, backend bind.ContractBackend) (*WarpISM, error) {
	contract, err := bindWarpISM(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &WarpISM{WarpISMCaller: WarpISMCaller{contract: contract}, WarpISMTransactor: WarpISMTransactor{contract: contract}, WarpISMFilterer: WarpISMFilterer{contract: contract}}, nil
}

// NewWarpISMCaller creates a new read-only instance of WarpISM, bound to a specific deployed contract.
func NewWarpISMCaller(address common.Address, caller bind.ContractCaller) (*WarpISMCaller, error) {
	contract, err := bindWarpISM(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &WarpISMCaller{contract: contract}, nil
}

// NewWarpISMTransactor creates a new write-only instance of WarpISM, bound to a specific deployed contract.
func NewWarpISMTransactor(address common.Address, transactor bind.ContractTransactor) (*WarpISMTransactor, error) {
	contract, err := bindWarpISM(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &WarpISMTransactor{contract: contract}, nil
}

// NewWarpISMFilterer creates a new log filterer instance of WarpISM, bound to a specific deployed contract.
func NewWarpISMFilterer(address common.Address, filterer bind.ContractFilterer) (*WarpISMFilterer, error) {
	contract, err := bindWarpISM(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &WarpISMFilterer{contract: contract}, nil
}

// bindWarpISM binds a generic wrapper to an already deployed contract.
func bindWarpISM(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := WarpISMMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_WarpISM *WarpISMRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _WarpISM.Contract.WarpISMCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_WarpISM *WarpISMRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _WarpISM.Contract.WarpISMTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_WarpISM *WarpISMRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _WarpISM.Contract.WarpISMTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_WarpISM *WarpISMCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _WarpISM.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_WarpISM *WarpISMTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _WarpISM.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_WarpISM *WarpISMTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _WarpISM.Contract.contract.Transact(opts, method, params...)
}

// DIGESTTYPEHASH is a free data retrieval call binding the contract method 0xf682a929.
//
// Solidity: function DIGEST_TYPE_HASH() view returns(bytes32)
func (_WarpISM *WarpISMCaller) DIGESTTYPEHASH(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _WarpISM.contract.Call(opts, &out, "DIGEST_TYPE_HASH")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DIGESTTYPEHASH is a free data retrieval call binding the contract method 0xf682a929.
//
// Solidity: function DIGEST_TYPE_HASH() view returns(bytes32)
func (_WarpISM *WarpISMSession) DIGESTTYPEHASH() ([32]byte, error) {
	return _WarpISM.Contract.DIGESTTYPEHASH(&_WarpISM.CallOpts)
}

// DIGESTTYPEHASH is a free data retrieval call binding the contract method 0xf682a929.
//
// Solidity: function DIGEST_TYPE_HASH() view returns(bytes32)
func (_WarpISM *WarpISMCallerSession) DIGESTTYPEHASH() ([32]byte, error) {
	return _WarpISM.Contract.DIGESTTYPEHASH(&_WarpISM.CallOpts)
}

// SIGNATURESIZE is a free data retrieval call binding the contract method 0x308ff0c9.
//
// Solidity: function SIGNATURE_SIZE() view returns(uint256)
func (_WarpISM *WarpISMCaller) SIGNATURESIZE(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _WarpISM.contract.Call(opts, &out, "SIGNATURE_SIZE")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SIGNATURESIZE is a free data retrieval call binding the contract method 0x308ff0c9.
//
// Solidity: function SIGNATURE_SIZE() view returns(uint256)
func (_WarpISM *WarpISMSession) SIGNATURESIZE() (*big.Int, error) {
	return _WarpISM.Contract.SIGNATURESIZE(&_WarpISM.CallOpts)
}

// SIGNATURESIZE is a free data retrieval call binding the contract method 0x308ff0c9.
//
// Solidity: function SIGNATURE_SIZE() view returns(uint256)
func (_WarpISM *WarpISMCallerSession) SIGNATURESIZE() (*big.Int, error) {
	return _WarpISM.Contract.SIGNATURESIZE(&_WarpISM.CallOpts)
}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_WarpISM *WarpISMCaller) Eip712Domain(opts *bind.CallOpts) (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	var out []interface{}
	err := _WarpISM.contract.Call(opts, &out, "eip712Domain")

	outstruct := new(struct {
		Fields            [1]byte
		Name              string
		Version           string
		ChainId           *big.Int
		VerifyingContract common.Address
		Salt              [32]byte
		Extensions        []*big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Fields = *abi.ConvertType(out[0], new([1]byte)).(*[1]byte)
	outstruct.Name = *abi.ConvertType(out[1], new(string)).(*string)
	outstruct.Version = *abi.ConvertType(out[2], new(string)).(*string)
	outstruct.ChainId = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.VerifyingContract = *abi.ConvertType(out[4], new(common.Address)).(*common.Address)
	outstruct.Salt = *abi.ConvertType(out[5], new([32]byte)).(*[32]byte)
	outstruct.Extensions = *abi.ConvertType(out[6], new([]*big.Int)).(*[]*big.Int)

	return *outstruct, err

}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_WarpISM *WarpISMSession) Eip712Domain() (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	return _WarpISM.Contract.Eip712Domain(&_WarpISM.CallOpts)
}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_WarpISM *WarpISMCallerSession) Eip712Domain() (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	return _WarpISM.Contract.Eip712Domain(&_WarpISM.CallOpts)
}

// GetDigest is a free data retrieval call binding the contract method 0x7d6e9e37.
//
// Solidity: function getDigest(bytes message) view returns(bytes32 digest)
func (_WarpISM *WarpISMCaller) GetDigest(opts *bind.CallOpts, message []byte) ([32]byte, error) {
	var out []interface{}
	err := _WarpISM.contract.Call(opts, &out, "getDigest", message)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetDigest is a free data retrieval call binding the contract method 0x7d6e9e37.
//
// Solidity: function getDigest(bytes message) view returns(bytes32 digest)
func (_WarpISM *WarpISMSession) GetDigest(message []byte) ([32]byte, error) {
	return _WarpISM.Contract.GetDigest(&_WarpISM.CallOpts, message)
}

// GetDigest is a free data retrieval call binding the contract method 0x7d6e9e37.
//
// Solidity: function getDigest(bytes message) view returns(bytes32 digest)
func (_WarpISM *WarpISMCallerSession) GetDigest(message []byte) ([32]byte, error) {
	return _WarpISM.Contract.GetDigest(&_WarpISM.CallOpts, message)
}

// GetSignatures is a free data retrieval call binding the contract method 0x6b45b4e3.
//
// Solidity: function getSignatures(bytes metadata) pure returns(bytes[] signatures)
func (_WarpISM *WarpISMCaller) GetSignatures(opts *bind.CallOpts, metadata []byte) ([][]byte, error) {
	var out []interface{}
	err := _WarpISM.contract.Call(opts, &out, "getSignatures", metadata)

	if err != nil {
		return *new([][]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([][]byte)).(*[][]byte)

	return out0, err

}

// GetSignatures is a free data retrieval call binding the contract method 0x6b45b4e3.
//
// Solidity: function getSignatures(bytes metadata) pure returns(bytes[] signatures)
func (_WarpISM *WarpISMSession) GetSignatures(metadata []byte) ([][]byte, error) {
	return _WarpISM.Contract.GetSignatures(&_WarpISM.CallOpts, metadata)
}

// GetSignatures is a free data retrieval call binding the contract method 0x6b45b4e3.
//
// Solidity: function getSignatures(bytes metadata) pure returns(bytes[] signatures)
func (_WarpISM *WarpISMCallerSession) GetSignatures(metadata []byte) ([][]byte, error) {
	return _WarpISM.Contract.GetSignatures(&_WarpISM.CallOpts, metadata)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_WarpISM *WarpISMCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _WarpISM.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_WarpISM *WarpISMSession) Owner() (common.Address, error) {
	return _WarpISM.Contract.Owner(&_WarpISM.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_WarpISM *WarpISMCallerSession) Owner() (common.Address, error) {
	return _WarpISM.Contract.Owner(&_WarpISM.CallOpts)
}

// SignerThreshold is a free data retrieval call binding the contract method 0xa4a4f390.
//
// Solidity: function signerThreshold() view returns(uint256)
func (_WarpISM *WarpISMCaller) SignerThreshold(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _WarpISM.contract.Call(opts, &out, "signerThreshold")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SignerThreshold is a free data retrieval call binding the contract method 0xa4a4f390.
//
// Solidity: function signerThreshold() view returns(uint256)
func (_WarpISM *WarpISMSession) SignerThreshold() (*big.Int, error) {
	return _WarpISM.Contract.SignerThreshold(&_WarpISM.CallOpts)
}

// SignerThreshold is a free data retrieval call binding the contract method 0xa4a4f390.
//
// Solidity: function signerThreshold() view returns(uint256)
func (_WarpISM *WarpISMCallerSession) SignerThreshold() (*big.Int, error) {
	return _WarpISM.Contract.SignerThreshold(&_WarpISM.CallOpts)
}

// ValidatorCount is a free data retrieval call binding the contract method 0x0f43a677.
//
// Solidity: function validatorCount() view returns(uint256)
func (_WarpISM *WarpISMCaller) ValidatorCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _WarpISM.contract.Call(opts, &out, "validatorCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ValidatorCount is a free data retrieval call binding the contract method 0x0f43a677.
//
// Solidity: function validatorCount() view returns(uint256)
func (_WarpISM *WarpISMSession) ValidatorCount() (*big.Int, error) {
	return _WarpISM.Contract.ValidatorCount(&_WarpISM.CallOpts)
}

// ValidatorCount is a free data retrieval call binding the contract method 0x0f43a677.
//
// Solidity: function validatorCount() view returns(uint256)
func (_WarpISM *WarpISMCallerSession) ValidatorCount() (*big.Int, error) {
	return _WarpISM.Contract.ValidatorCount(&_WarpISM.CallOpts)
}

// Validators is a free data retrieval call binding the contract method 0xfa52c7d8.
//
// Solidity: function validators(address ) view returns(bool)
func (_WarpISM *WarpISMCaller) Validators(opts *bind.CallOpts, arg0 common.Address) (bool, error) {
	var out []interface{}
	err := _WarpISM.contract.Call(opts, &out, "validators", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Validators is a free data retrieval call binding the contract method 0xfa52c7d8.
//
// Solidity: function validators(address ) view returns(bool)
func (_WarpISM *WarpISMSession) Validators(arg0 common.Address) (bool, error) {
	return _WarpISM.Contract.Validators(&_WarpISM.CallOpts, arg0)
}

// Validators is a free data retrieval call binding the contract method 0xfa52c7d8.
//
// Solidity: function validators(address ) view returns(bool)
func (_WarpISM *WarpISMCallerSession) Validators(arg0 common.Address) (bool, error) {
	return _WarpISM.Contract.Validators(&_WarpISM.CallOpts, arg0)
}

// Verify is a free data retrieval call binding the contract method 0xf7e83aee.
//
// Solidity: function verify(bytes metadata, bytes message) view returns(bool success)
func (_WarpISM *WarpISMCaller) Verify(opts *bind.CallOpts, metadata []byte, message []byte) (bool, error) {
	var out []interface{}
	err := _WarpISM.contract.Call(opts, &out, "verify", metadata, message)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Verify is a free data retrieval call binding the contract method 0xf7e83aee.
//
// Solidity: function verify(bytes metadata, bytes message) view returns(bool success)
func (_WarpISM *WarpISMSession) Verify(metadata []byte, message []byte) (bool, error) {
	return _WarpISM.Contract.Verify(&_WarpISM.CallOpts, metadata, message)
}

// Verify is a free data retrieval call binding the contract method 0xf7e83aee.
//
// Solidity: function verify(bytes metadata, bytes message) view returns(bool success)
func (_WarpISM *WarpISMCallerSession) Verify(metadata []byte, message []byte) (bool, error) {
	return _WarpISM.Contract.Verify(&_WarpISM.CallOpts, metadata, message)
}

// AddValidator is a paid mutator transaction binding the contract method 0x4d238c8e.
//
// Solidity: function addValidator(address validator) returns()
func (_WarpISM *WarpISMTransactor) AddValidator(opts *bind.TransactOpts, validator common.Address) (*types.Transaction, error) {
	return _WarpISM.contract.Transact(opts, "addValidator", validator)
}

// AddValidator is a paid mutator transaction binding the contract method 0x4d238c8e.
//
// Solidity: function addValidator(address validator) returns()
func (_WarpISM *WarpISMSession) AddValidator(validator common.Address) (*types.Transaction, error) {
	return _WarpISM.Contract.AddValidator(&_WarpISM.TransactOpts, validator)
}

// AddValidator is a paid mutator transaction binding the contract method 0x4d238c8e.
//
// Solidity: function addValidator(address validator) returns()
func (_WarpISM *WarpISMTransactorSession) AddValidator(validator common.Address) (*types.Transaction, error) {
	return _WarpISM.Contract.AddValidator(&_WarpISM.TransactOpts, validator)
}

// RemoveValidator is a paid mutator transaction binding the contract method 0x40a141ff.
//
// Solidity: function removeValidator(address validator) returns()
func (_WarpISM *WarpISMTransactor) RemoveValidator(opts *bind.TransactOpts, validator common.Address) (*types.Transaction, error) {
	return _WarpISM.contract.Transact(opts, "removeValidator", validator)
}

// RemoveValidator is a paid mutator transaction binding the contract method 0x40a141ff.
//
// Solidity: function removeValidator(address validator) returns()
func (_WarpISM *WarpISMSession) RemoveValidator(validator common.Address) (*types.Transaction, error) {
	return _WarpISM.Contract.RemoveValidator(&_WarpISM.TransactOpts, validator)
}

// RemoveValidator is a paid mutator transaction binding the contract method 0x40a141ff.
//
// Solidity: function removeValidator(address validator) returns()
func (_WarpISM *WarpISMTransactorSession) RemoveValidator(validator common.Address) (*types.Transaction, error) {
	return _WarpISM.Contract.RemoveValidator(&_WarpISM.TransactOpts, validator)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_WarpISM *WarpISMTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _WarpISM.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_WarpISM *WarpISMSession) RenounceOwnership() (*types.Transaction, error) {
	return _WarpISM.Contract.RenounceOwnership(&_WarpISM.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_WarpISM *WarpISMTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _WarpISM.Contract.RenounceOwnership(&_WarpISM.TransactOpts)
}

// SetSignerThreshold is a paid mutator transaction binding the contract method 0x251b8192.
//
// Solidity: function setSignerThreshold(uint256 signatureRatio) returns()
func (_WarpISM *WarpISMTransactor) SetSignerThreshold(opts *bind.TransactOpts, signatureRatio *big.Int) (*types.Transaction, error) {
	return _WarpISM.contract.Transact(opts, "setSignerThreshold", signatureRatio)
}

// SetSignerThreshold is a paid mutator transaction binding the contract method 0x251b8192.
//
// Solidity: function setSignerThreshold(uint256 signatureRatio) returns()
func (_WarpISM *WarpISMSession) SetSignerThreshold(signatureRatio *big.Int) (*types.Transaction, error) {
	return _WarpISM.Contract.SetSignerThreshold(&_WarpISM.TransactOpts, signatureRatio)
}

// SetSignerThreshold is a paid mutator transaction binding the contract method 0x251b8192.
//
// Solidity: function setSignerThreshold(uint256 signatureRatio) returns()
func (_WarpISM *WarpISMTransactorSession) SetSignerThreshold(signatureRatio *big.Int) (*types.Transaction, error) {
	return _WarpISM.Contract.SetSignerThreshold(&_WarpISM.TransactOpts, signatureRatio)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_WarpISM *WarpISMTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _WarpISM.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_WarpISM *WarpISMSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _WarpISM.Contract.TransferOwnership(&_WarpISM.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_WarpISM *WarpISMTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _WarpISM.Contract.TransferOwnership(&_WarpISM.TransactOpts, newOwner)
}

// WarpISMEIP712DomainChangedIterator is returned from FilterEIP712DomainChanged and is used to iterate over the raw logs and unpacked data for EIP712DomainChanged events raised by the WarpISM contract.
type WarpISMEIP712DomainChangedIterator struct {
	Event *WarpISMEIP712DomainChanged // Event containing the contract specifics and raw log

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
func (it *WarpISMEIP712DomainChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(WarpISMEIP712DomainChanged)
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
		it.Event = new(WarpISMEIP712DomainChanged)
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
func (it *WarpISMEIP712DomainChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *WarpISMEIP712DomainChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// WarpISMEIP712DomainChanged represents a EIP712DomainChanged event raised by the WarpISM contract.
type WarpISMEIP712DomainChanged struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterEIP712DomainChanged is a free log retrieval operation binding the contract event 0x0a6387c9ea3628b88a633bb4f3b151770f70085117a15f9bf3787cda53f13d31.
//
// Solidity: event EIP712DomainChanged()
func (_WarpISM *WarpISMFilterer) FilterEIP712DomainChanged(opts *bind.FilterOpts) (*WarpISMEIP712DomainChangedIterator, error) {

	logs, sub, err := _WarpISM.contract.FilterLogs(opts, "EIP712DomainChanged")
	if err != nil {
		return nil, err
	}
	return &WarpISMEIP712DomainChangedIterator{contract: _WarpISM.contract, event: "EIP712DomainChanged", logs: logs, sub: sub}, nil
}

// WatchEIP712DomainChanged is a free log subscription operation binding the contract event 0x0a6387c9ea3628b88a633bb4f3b151770f70085117a15f9bf3787cda53f13d31.
//
// Solidity: event EIP712DomainChanged()
func (_WarpISM *WarpISMFilterer) WatchEIP712DomainChanged(opts *bind.WatchOpts, sink chan<- *WarpISMEIP712DomainChanged) (event.Subscription, error) {

	logs, sub, err := _WarpISM.contract.WatchLogs(opts, "EIP712DomainChanged")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(WarpISMEIP712DomainChanged)
				if err := _WarpISM.contract.UnpackLog(event, "EIP712DomainChanged", log); err != nil {
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

// ParseEIP712DomainChanged is a log parse operation binding the contract event 0x0a6387c9ea3628b88a633bb4f3b151770f70085117a15f9bf3787cda53f13d31.
//
// Solidity: event EIP712DomainChanged()
func (_WarpISM *WarpISMFilterer) ParseEIP712DomainChanged(log types.Log) (*WarpISMEIP712DomainChanged, error) {
	event := new(WarpISMEIP712DomainChanged)
	if err := _WarpISM.contract.UnpackLog(event, "EIP712DomainChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// WarpISMNewValidatorIterator is returned from FilterNewValidator and is used to iterate over the raw logs and unpacked data for NewValidator events raised by the WarpISM contract.
type WarpISMNewValidatorIterator struct {
	Event *WarpISMNewValidator // Event containing the contract specifics and raw log

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
func (it *WarpISMNewValidatorIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(WarpISMNewValidator)
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
		it.Event = new(WarpISMNewValidator)
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
func (it *WarpISMNewValidatorIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *WarpISMNewValidatorIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// WarpISMNewValidator represents a NewValidator event raised by the WarpISM contract.
type WarpISMNewValidator struct {
	Validator common.Address
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterNewValidator is a free log retrieval operation binding the contract event 0x29b4645f23b856eccf12b3b38e036c3221ca1b5a9afa2a83aea7ead34e47987c.
//
// Solidity: event NewValidator(address indexed validator)
func (_WarpISM *WarpISMFilterer) FilterNewValidator(opts *bind.FilterOpts, validator []common.Address) (*WarpISMNewValidatorIterator, error) {

	var validatorRule []interface{}
	for _, validatorItem := range validator {
		validatorRule = append(validatorRule, validatorItem)
	}

	logs, sub, err := _WarpISM.contract.FilterLogs(opts, "NewValidator", validatorRule)
	if err != nil {
		return nil, err
	}
	return &WarpISMNewValidatorIterator{contract: _WarpISM.contract, event: "NewValidator", logs: logs, sub: sub}, nil
}

// WatchNewValidator is a free log subscription operation binding the contract event 0x29b4645f23b856eccf12b3b38e036c3221ca1b5a9afa2a83aea7ead34e47987c.
//
// Solidity: event NewValidator(address indexed validator)
func (_WarpISM *WarpISMFilterer) WatchNewValidator(opts *bind.WatchOpts, sink chan<- *WarpISMNewValidator, validator []common.Address) (event.Subscription, error) {

	var validatorRule []interface{}
	for _, validatorItem := range validator {
		validatorRule = append(validatorRule, validatorItem)
	}

	logs, sub, err := _WarpISM.contract.WatchLogs(opts, "NewValidator", validatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(WarpISMNewValidator)
				if err := _WarpISM.contract.UnpackLog(event, "NewValidator", log); err != nil {
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

// ParseNewValidator is a log parse operation binding the contract event 0x29b4645f23b856eccf12b3b38e036c3221ca1b5a9afa2a83aea7ead34e47987c.
//
// Solidity: event NewValidator(address indexed validator)
func (_WarpISM *WarpISMFilterer) ParseNewValidator(log types.Log) (*WarpISMNewValidator, error) {
	event := new(WarpISMNewValidator)
	if err := _WarpISM.contract.UnpackLog(event, "NewValidator", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// WarpISMOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the WarpISM contract.
type WarpISMOwnershipTransferredIterator struct {
	Event *WarpISMOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *WarpISMOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(WarpISMOwnershipTransferred)
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
		it.Event = new(WarpISMOwnershipTransferred)
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
func (it *WarpISMOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *WarpISMOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// WarpISMOwnershipTransferred represents a OwnershipTransferred event raised by the WarpISM contract.
type WarpISMOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_WarpISM *WarpISMFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*WarpISMOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _WarpISM.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &WarpISMOwnershipTransferredIterator{contract: _WarpISM.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_WarpISM *WarpISMFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *WarpISMOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _WarpISM.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(WarpISMOwnershipTransferred)
				if err := _WarpISM.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_WarpISM *WarpISMFilterer) ParseOwnershipTransferred(log types.Log) (*WarpISMOwnershipTransferred, error) {
	event := new(WarpISMOwnershipTransferred)
	if err := _WarpISM.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// WarpISMRemovedValidatorIterator is returned from FilterRemovedValidator and is used to iterate over the raw logs and unpacked data for RemovedValidator events raised by the WarpISM contract.
type WarpISMRemovedValidatorIterator struct {
	Event *WarpISMRemovedValidator // Event containing the contract specifics and raw log

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
func (it *WarpISMRemovedValidatorIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(WarpISMRemovedValidator)
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
		it.Event = new(WarpISMRemovedValidator)
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
func (it *WarpISMRemovedValidatorIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *WarpISMRemovedValidatorIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// WarpISMRemovedValidator represents a RemovedValidator event raised by the WarpISM contract.
type WarpISMRemovedValidator struct {
	Validator common.Address
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterRemovedValidator is a free log retrieval operation binding the contract event 0xb625c55cf7e37b54fcd18bc4edafdf3f4f9acd59a5ec824c77c795dcb2d65070.
//
// Solidity: event RemovedValidator(address indexed validator)
func (_WarpISM *WarpISMFilterer) FilterRemovedValidator(opts *bind.FilterOpts, validator []common.Address) (*WarpISMRemovedValidatorIterator, error) {

	var validatorRule []interface{}
	for _, validatorItem := range validator {
		validatorRule = append(validatorRule, validatorItem)
	}

	logs, sub, err := _WarpISM.contract.FilterLogs(opts, "RemovedValidator", validatorRule)
	if err != nil {
		return nil, err
	}
	return &WarpISMRemovedValidatorIterator{contract: _WarpISM.contract, event: "RemovedValidator", logs: logs, sub: sub}, nil
}

// WatchRemovedValidator is a free log subscription operation binding the contract event 0xb625c55cf7e37b54fcd18bc4edafdf3f4f9acd59a5ec824c77c795dcb2d65070.
//
// Solidity: event RemovedValidator(address indexed validator)
func (_WarpISM *WarpISMFilterer) WatchRemovedValidator(opts *bind.WatchOpts, sink chan<- *WarpISMRemovedValidator, validator []common.Address) (event.Subscription, error) {

	var validatorRule []interface{}
	for _, validatorItem := range validator {
		validatorRule = append(validatorRule, validatorItem)
	}

	logs, sub, err := _WarpISM.contract.WatchLogs(opts, "RemovedValidator", validatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(WarpISMRemovedValidator)
				if err := _WarpISM.contract.UnpackLog(event, "RemovedValidator", log); err != nil {
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

// ParseRemovedValidator is a log parse operation binding the contract event 0xb625c55cf7e37b54fcd18bc4edafdf3f4f9acd59a5ec824c77c795dcb2d65070.
//
// Solidity: event RemovedValidator(address indexed validator)
func (_WarpISM *WarpISMFilterer) ParseRemovedValidator(log types.Log) (*WarpISMRemovedValidator, error) {
	event := new(WarpISMRemovedValidator)
	if err := _WarpISM.contract.UnpackLog(event, "RemovedValidator", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// WarpISMSignerThresholdSetIterator is returned from FilterSignerThresholdSet and is used to iterate over the raw logs and unpacked data for SignerThresholdSet events raised by the WarpISM contract.
type WarpISMSignerThresholdSetIterator struct {
	Event *WarpISMSignerThresholdSet // Event containing the contract specifics and raw log

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
func (it *WarpISMSignerThresholdSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(WarpISMSignerThresholdSet)
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
		it.Event = new(WarpISMSignerThresholdSet)
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
func (it *WarpISMSignerThresholdSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *WarpISMSignerThresholdSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// WarpISMSignerThresholdSet represents a SignerThresholdSet event raised by the WarpISM contract.
type WarpISMSignerThresholdSet struct {
	Ratio *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterSignerThresholdSet is a free log retrieval operation binding the contract event 0x030f8d880b0bf7b6311c00e028809570118cbccc8302e08c8776ad77d10f505e.
//
// Solidity: event SignerThresholdSet(uint256 indexed ratio)
func (_WarpISM *WarpISMFilterer) FilterSignerThresholdSet(opts *bind.FilterOpts, ratio []*big.Int) (*WarpISMSignerThresholdSetIterator, error) {

	var ratioRule []interface{}
	for _, ratioItem := range ratio {
		ratioRule = append(ratioRule, ratioItem)
	}

	logs, sub, err := _WarpISM.contract.FilterLogs(opts, "SignerThresholdSet", ratioRule)
	if err != nil {
		return nil, err
	}
	return &WarpISMSignerThresholdSetIterator{contract: _WarpISM.contract, event: "SignerThresholdSet", logs: logs, sub: sub}, nil
}

// WatchSignerThresholdSet is a free log subscription operation binding the contract event 0x030f8d880b0bf7b6311c00e028809570118cbccc8302e08c8776ad77d10f505e.
//
// Solidity: event SignerThresholdSet(uint256 indexed ratio)
func (_WarpISM *WarpISMFilterer) WatchSignerThresholdSet(opts *bind.WatchOpts, sink chan<- *WarpISMSignerThresholdSet, ratio []*big.Int) (event.Subscription, error) {

	var ratioRule []interface{}
	for _, ratioItem := range ratio {
		ratioRule = append(ratioRule, ratioItem)
	}

	logs, sub, err := _WarpISM.contract.WatchLogs(opts, "SignerThresholdSet", ratioRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(WarpISMSignerThresholdSet)
				if err := _WarpISM.contract.UnpackLog(event, "SignerThresholdSet", log); err != nil {
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

// ParseSignerThresholdSet is a log parse operation binding the contract event 0x030f8d880b0bf7b6311c00e028809570118cbccc8302e08c8776ad77d10f505e.
//
// Solidity: event SignerThresholdSet(uint256 indexed ratio)
func (_WarpISM *WarpISMFilterer) ParseSignerThresholdSet(log types.Log) (*WarpISMSignerThresholdSet, error) {
	event := new(WarpISMSignerThresholdSet)
	if err := _WarpISM.contract.UnpackLog(event, "SignerThresholdSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
