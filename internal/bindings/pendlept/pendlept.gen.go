// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package pendlept

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

// PendlePTMetaData contains all meta data concerning the PendlePT contract.
var PendlePTMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"SY\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"YT\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"approve\",\"inputs\":[{\"name\":\"spender\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"balanceOf\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"decimals\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint8\",\"internalType\":\"uint8\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"expiry\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isExpired\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"}]",
}

// PendlePTABI is the input ABI used to generate the binding from.
// Deprecated: Use PendlePTMetaData.ABI instead.
var PendlePTABI = PendlePTMetaData.ABI

// PendlePT is an auto generated Go binding around an Ethereum contract.
type PendlePT struct {
	PendlePTCaller     // Read-only binding to the contract
	PendlePTTransactor // Write-only binding to the contract
	PendlePTFilterer   // Log filterer for contract events
}

// PendlePTCaller is an auto generated read-only Go binding around an Ethereum contract.
type PendlePTCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PendlePTTransactor is an auto generated write-only Go binding around an Ethereum contract.
type PendlePTTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PendlePTFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type PendlePTFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PendlePTSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type PendlePTSession struct {
	Contract     *PendlePT         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// PendlePTCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type PendlePTCallerSession struct {
	Contract *PendlePTCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// PendlePTTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type PendlePTTransactorSession struct {
	Contract     *PendlePTTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// PendlePTRaw is an auto generated low-level Go binding around an Ethereum contract.
type PendlePTRaw struct {
	Contract *PendlePT // Generic contract binding to access the raw methods on
}

// PendlePTCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type PendlePTCallerRaw struct {
	Contract *PendlePTCaller // Generic read-only contract binding to access the raw methods on
}

// PendlePTTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type PendlePTTransactorRaw struct {
	Contract *PendlePTTransactor // Generic write-only contract binding to access the raw methods on
}

// NewPendlePT creates a new instance of PendlePT, bound to a specific deployed contract.
func NewPendlePT(address common.Address, backend bind.ContractBackend) (*PendlePT, error) {
	contract, err := bindPendlePT(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &PendlePT{PendlePTCaller: PendlePTCaller{contract: contract}, PendlePTTransactor: PendlePTTransactor{contract: contract}, PendlePTFilterer: PendlePTFilterer{contract: contract}}, nil
}

// NewPendlePTCaller creates a new read-only instance of PendlePT, bound to a specific deployed contract.
func NewPendlePTCaller(address common.Address, caller bind.ContractCaller) (*PendlePTCaller, error) {
	contract, err := bindPendlePT(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &PendlePTCaller{contract: contract}, nil
}

// NewPendlePTTransactor creates a new write-only instance of PendlePT, bound to a specific deployed contract.
func NewPendlePTTransactor(address common.Address, transactor bind.ContractTransactor) (*PendlePTTransactor, error) {
	contract, err := bindPendlePT(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &PendlePTTransactor{contract: contract}, nil
}

// NewPendlePTFilterer creates a new log filterer instance of PendlePT, bound to a specific deployed contract.
func NewPendlePTFilterer(address common.Address, filterer bind.ContractFilterer) (*PendlePTFilterer, error) {
	contract, err := bindPendlePT(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &PendlePTFilterer{contract: contract}, nil
}

// bindPendlePT binds a generic wrapper to an already deployed contract.
func bindPendlePT(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := PendlePTMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PendlePT *PendlePTRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _PendlePT.Contract.PendlePTCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PendlePT *PendlePTRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PendlePT.Contract.PendlePTTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PendlePT *PendlePTRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PendlePT.Contract.PendlePTTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PendlePT *PendlePTCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _PendlePT.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PendlePT *PendlePTTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PendlePT.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PendlePT *PendlePTTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PendlePT.Contract.contract.Transact(opts, method, params...)
}

// SY is a free data retrieval call binding the contract method 0xafd27bf5.
//
// Solidity: function SY() view returns(address)
func (_PendlePT *PendlePTCaller) SY(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _PendlePT.contract.Call(opts, &out, "SY")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// SY is a free data retrieval call binding the contract method 0xafd27bf5.
//
// Solidity: function SY() view returns(address)
func (_PendlePT *PendlePTSession) SY() (common.Address, error) {
	return _PendlePT.Contract.SY(&_PendlePT.CallOpts)
}

// SY is a free data retrieval call binding the contract method 0xafd27bf5.
//
// Solidity: function SY() view returns(address)
func (_PendlePT *PendlePTCallerSession) SY() (common.Address, error) {
	return _PendlePT.Contract.SY(&_PendlePT.CallOpts)
}

// YT is a free data retrieval call binding the contract method 0x781c18db.
//
// Solidity: function YT() view returns(address)
func (_PendlePT *PendlePTCaller) YT(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _PendlePT.contract.Call(opts, &out, "YT")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// YT is a free data retrieval call binding the contract method 0x781c18db.
//
// Solidity: function YT() view returns(address)
func (_PendlePT *PendlePTSession) YT() (common.Address, error) {
	return _PendlePT.Contract.YT(&_PendlePT.CallOpts)
}

// YT is a free data retrieval call binding the contract method 0x781c18db.
//
// Solidity: function YT() view returns(address)
func (_PendlePT *PendlePTCallerSession) YT() (common.Address, error) {
	return _PendlePT.Contract.YT(&_PendlePT.CallOpts)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_PendlePT *PendlePTCaller) BalanceOf(opts *bind.CallOpts, account common.Address) (*big.Int, error) {
	var out []interface{}
	err := _PendlePT.contract.Call(opts, &out, "balanceOf", account)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_PendlePT *PendlePTSession) BalanceOf(account common.Address) (*big.Int, error) {
	return _PendlePT.Contract.BalanceOf(&_PendlePT.CallOpts, account)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_PendlePT *PendlePTCallerSession) BalanceOf(account common.Address) (*big.Int, error) {
	return _PendlePT.Contract.BalanceOf(&_PendlePT.CallOpts, account)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_PendlePT *PendlePTCaller) Decimals(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _PendlePT.contract.Call(opts, &out, "decimals")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_PendlePT *PendlePTSession) Decimals() (uint8, error) {
	return _PendlePT.Contract.Decimals(&_PendlePT.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_PendlePT *PendlePTCallerSession) Decimals() (uint8, error) {
	return _PendlePT.Contract.Decimals(&_PendlePT.CallOpts)
}

// Expiry is a free data retrieval call binding the contract method 0xe184c9be.
//
// Solidity: function expiry() view returns(uint256)
func (_PendlePT *PendlePTCaller) Expiry(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _PendlePT.contract.Call(opts, &out, "expiry")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Expiry is a free data retrieval call binding the contract method 0xe184c9be.
//
// Solidity: function expiry() view returns(uint256)
func (_PendlePT *PendlePTSession) Expiry() (*big.Int, error) {
	return _PendlePT.Contract.Expiry(&_PendlePT.CallOpts)
}

// Expiry is a free data retrieval call binding the contract method 0xe184c9be.
//
// Solidity: function expiry() view returns(uint256)
func (_PendlePT *PendlePTCallerSession) Expiry() (*big.Int, error) {
	return _PendlePT.Contract.Expiry(&_PendlePT.CallOpts)
}

// IsExpired is a free data retrieval call binding the contract method 0x2f13b60c.
//
// Solidity: function isExpired() view returns(bool)
func (_PendlePT *PendlePTCaller) IsExpired(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _PendlePT.contract.Call(opts, &out, "isExpired")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsExpired is a free data retrieval call binding the contract method 0x2f13b60c.
//
// Solidity: function isExpired() view returns(bool)
func (_PendlePT *PendlePTSession) IsExpired() (bool, error) {
	return _PendlePT.Contract.IsExpired(&_PendlePT.CallOpts)
}

// IsExpired is a free data retrieval call binding the contract method 0x2f13b60c.
//
// Solidity: function isExpired() view returns(bool)
func (_PendlePT *PendlePTCallerSession) IsExpired() (bool, error) {
	return _PendlePT.Contract.IsExpired(&_PendlePT.CallOpts)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_PendlePT *PendlePTTransactor) Approve(opts *bind.TransactOpts, spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _PendlePT.contract.Transact(opts, "approve", spender, amount)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_PendlePT *PendlePTSession) Approve(spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _PendlePT.Contract.Approve(&_PendlePT.TransactOpts, spender, amount)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_PendlePT *PendlePTTransactorSession) Approve(spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _PendlePT.Contract.Approve(&_PendlePT.TransactOpts, spender, amount)
}
