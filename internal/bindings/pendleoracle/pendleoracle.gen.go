// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package pendleoracle

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

// PendleOracleMetaData contains all meta data concerning the PendleOracle contract.
var PendleOracleMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"getOracleState\",\"inputs\":[{\"name\":\"market\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"duration\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"outputs\":[{\"name\":\"increaseCardinalityRequired\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"cardinalityRequired\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"oldestObservationSatisfied\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getPtToAssetRate\",\"inputs\":[{\"name\":\"market\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"duration\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"}]",
}

// PendleOracleABI is the input ABI used to generate the binding from.
// Deprecated: Use PendleOracleMetaData.ABI instead.
var PendleOracleABI = PendleOracleMetaData.ABI

// PendleOracle is an auto generated Go binding around an Ethereum contract.
type PendleOracle struct {
	PendleOracleCaller     // Read-only binding to the contract
	PendleOracleTransactor // Write-only binding to the contract
	PendleOracleFilterer   // Log filterer for contract events
}

// PendleOracleCaller is an auto generated read-only Go binding around an Ethereum contract.
type PendleOracleCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PendleOracleTransactor is an auto generated write-only Go binding around an Ethereum contract.
type PendleOracleTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PendleOracleFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type PendleOracleFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PendleOracleSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type PendleOracleSession struct {
	Contract     *PendleOracle     // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// PendleOracleCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type PendleOracleCallerSession struct {
	Contract *PendleOracleCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts       // Call options to use throughout this session
}

// PendleOracleTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type PendleOracleTransactorSession struct {
	Contract     *PendleOracleTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts       // Transaction auth options to use throughout this session
}

// PendleOracleRaw is an auto generated low-level Go binding around an Ethereum contract.
type PendleOracleRaw struct {
	Contract *PendleOracle // Generic contract binding to access the raw methods on
}

// PendleOracleCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type PendleOracleCallerRaw struct {
	Contract *PendleOracleCaller // Generic read-only contract binding to access the raw methods on
}

// PendleOracleTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type PendleOracleTransactorRaw struct {
	Contract *PendleOracleTransactor // Generic write-only contract binding to access the raw methods on
}

// NewPendleOracle creates a new instance of PendleOracle, bound to a specific deployed contract.
func NewPendleOracle(address common.Address, backend bind.ContractBackend) (*PendleOracle, error) {
	contract, err := bindPendleOracle(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &PendleOracle{PendleOracleCaller: PendleOracleCaller{contract: contract}, PendleOracleTransactor: PendleOracleTransactor{contract: contract}, PendleOracleFilterer: PendleOracleFilterer{contract: contract}}, nil
}

// NewPendleOracleCaller creates a new read-only instance of PendleOracle, bound to a specific deployed contract.
func NewPendleOracleCaller(address common.Address, caller bind.ContractCaller) (*PendleOracleCaller, error) {
	contract, err := bindPendleOracle(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &PendleOracleCaller{contract: contract}, nil
}

// NewPendleOracleTransactor creates a new write-only instance of PendleOracle, bound to a specific deployed contract.
func NewPendleOracleTransactor(address common.Address, transactor bind.ContractTransactor) (*PendleOracleTransactor, error) {
	contract, err := bindPendleOracle(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &PendleOracleTransactor{contract: contract}, nil
}

// NewPendleOracleFilterer creates a new log filterer instance of PendleOracle, bound to a specific deployed contract.
func NewPendleOracleFilterer(address common.Address, filterer bind.ContractFilterer) (*PendleOracleFilterer, error) {
	contract, err := bindPendleOracle(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &PendleOracleFilterer{contract: contract}, nil
}

// bindPendleOracle binds a generic wrapper to an already deployed contract.
func bindPendleOracle(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := PendleOracleMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PendleOracle *PendleOracleRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _PendleOracle.Contract.PendleOracleCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PendleOracle *PendleOracleRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PendleOracle.Contract.PendleOracleTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PendleOracle *PendleOracleRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PendleOracle.Contract.PendleOracleTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PendleOracle *PendleOracleCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _PendleOracle.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PendleOracle *PendleOracleTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PendleOracle.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PendleOracle *PendleOracleTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PendleOracle.Contract.contract.Transact(opts, method, params...)
}

// GetOracleState is a free data retrieval call binding the contract method 0x873e9600.
//
// Solidity: function getOracleState(address market, uint32 duration) view returns(bool increaseCardinalityRequired, uint16 cardinalityRequired, bool oldestObservationSatisfied)
func (_PendleOracle *PendleOracleCaller) GetOracleState(opts *bind.CallOpts, market common.Address, duration uint32) (struct {
	IncreaseCardinalityRequired bool
	CardinalityRequired         uint16
	OldestObservationSatisfied  bool
}, error) {
	var out []interface{}
	err := _PendleOracle.contract.Call(opts, &out, "getOracleState", market, duration)

	outstruct := new(struct {
		IncreaseCardinalityRequired bool
		CardinalityRequired         uint16
		OldestObservationSatisfied  bool
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.IncreaseCardinalityRequired = *abi.ConvertType(out[0], new(bool)).(*bool)
	outstruct.CardinalityRequired = *abi.ConvertType(out[1], new(uint16)).(*uint16)
	outstruct.OldestObservationSatisfied = *abi.ConvertType(out[2], new(bool)).(*bool)

	return *outstruct, err

}

// GetOracleState is a free data retrieval call binding the contract method 0x873e9600.
//
// Solidity: function getOracleState(address market, uint32 duration) view returns(bool increaseCardinalityRequired, uint16 cardinalityRequired, bool oldestObservationSatisfied)
func (_PendleOracle *PendleOracleSession) GetOracleState(market common.Address, duration uint32) (struct {
	IncreaseCardinalityRequired bool
	CardinalityRequired         uint16
	OldestObservationSatisfied  bool
}, error) {
	return _PendleOracle.Contract.GetOracleState(&_PendleOracle.CallOpts, market, duration)
}

// GetOracleState is a free data retrieval call binding the contract method 0x873e9600.
//
// Solidity: function getOracleState(address market, uint32 duration) view returns(bool increaseCardinalityRequired, uint16 cardinalityRequired, bool oldestObservationSatisfied)
func (_PendleOracle *PendleOracleCallerSession) GetOracleState(market common.Address, duration uint32) (struct {
	IncreaseCardinalityRequired bool
	CardinalityRequired         uint16
	OldestObservationSatisfied  bool
}, error) {
	return _PendleOracle.Contract.GetOracleState(&_PendleOracle.CallOpts, market, duration)
}

// GetPtToAssetRate is a free data retrieval call binding the contract method 0xabca0eab.
//
// Solidity: function getPtToAssetRate(address market, uint32 duration) view returns(uint256)
func (_PendleOracle *PendleOracleCaller) GetPtToAssetRate(opts *bind.CallOpts, market common.Address, duration uint32) (*big.Int, error) {
	var out []interface{}
	err := _PendleOracle.contract.Call(opts, &out, "getPtToAssetRate", market, duration)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetPtToAssetRate is a free data retrieval call binding the contract method 0xabca0eab.
//
// Solidity: function getPtToAssetRate(address market, uint32 duration) view returns(uint256)
func (_PendleOracle *PendleOracleSession) GetPtToAssetRate(market common.Address, duration uint32) (*big.Int, error) {
	return _PendleOracle.Contract.GetPtToAssetRate(&_PendleOracle.CallOpts, market, duration)
}

// GetPtToAssetRate is a free data retrieval call binding the contract method 0xabca0eab.
//
// Solidity: function getPtToAssetRate(address market, uint32 duration) view returns(uint256)
func (_PendleOracle *PendleOracleCallerSession) GetPtToAssetRate(market common.Address, duration uint32) (*big.Int, error) {
	return _PendleOracle.Contract.GetPtToAssetRate(&_PendleOracle.CallOpts, market, duration)
}
