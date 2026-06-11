// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package vault

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

// LibWithdrawQueueWithdrawRequest is an auto generated low-level Go binding around an user-defined struct.
type LibWithdrawQueueWithdrawRequest struct {
	Owner    common.Address
	Receiver common.Address
	Shares   *big.Int
}

// VaultMetaData contains all meta data concerning the Vault contract.
var VaultMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"fulfillWithdraw\",\"inputs\":[{\"name\":\"id\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"guardCheckpoint\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"harvestAll\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"idleAssets\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"idleReserveBps\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint16\",\"internalType\":\"uint16\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isQuarantined\",\"inputs\":[{\"name\":\"strategyId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"maxRebalanceDelta\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint16\",\"internalType\":\"uint16\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"nextWithdrawRequestId\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"paused\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"pendingWithdrawShares\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"rebalance\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setAllocation\",\"inputs\":[{\"name\":\"strategyIds\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"},{\"name\":\"bps\",\"type\":\"uint16[]\",\"internalType\":\"uint16[]\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"strategies\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"strategyCap\",\"inputs\":[{\"name\":\"strategyId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint16\",\"internalType\":\"uint16\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"strategyTotalAssets\",\"inputs\":[{\"name\":\"strategyId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"targetAllocation\",\"inputs\":[{\"name\":\"strategyId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint16\",\"internalType\":\"uint16\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"totalAssets\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"withdrawRequest\",\"inputs\":[{\"name\":\"id\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structLibWithdrawQueue.WithdrawRequest\",\"components\":[{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"receiver\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"shares\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}],\"stateMutability\":\"view\"}]",
}

// VaultABI is the input ABI used to generate the binding from.
// Deprecated: Use VaultMetaData.ABI instead.
var VaultABI = VaultMetaData.ABI

// Vault is an auto generated Go binding around an Ethereum contract.
type Vault struct {
	VaultCaller     // Read-only binding to the contract
	VaultTransactor // Write-only binding to the contract
	VaultFilterer   // Log filterer for contract events
}

// VaultCaller is an auto generated read-only Go binding around an Ethereum contract.
type VaultCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// VaultTransactor is an auto generated write-only Go binding around an Ethereum contract.
type VaultTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// VaultFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type VaultFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// VaultSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type VaultSession struct {
	Contract     *Vault            // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// VaultCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type VaultCallerSession struct {
	Contract *VaultCaller  // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// VaultTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type VaultTransactorSession struct {
	Contract     *VaultTransactor  // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// VaultRaw is an auto generated low-level Go binding around an Ethereum contract.
type VaultRaw struct {
	Contract *Vault // Generic contract binding to access the raw methods on
}

// VaultCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type VaultCallerRaw struct {
	Contract *VaultCaller // Generic read-only contract binding to access the raw methods on
}

// VaultTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type VaultTransactorRaw struct {
	Contract *VaultTransactor // Generic write-only contract binding to access the raw methods on
}

// NewVault creates a new instance of Vault, bound to a specific deployed contract.
func NewVault(address common.Address, backend bind.ContractBackend) (*Vault, error) {
	contract, err := bindVault(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Vault{VaultCaller: VaultCaller{contract: contract}, VaultTransactor: VaultTransactor{contract: contract}, VaultFilterer: VaultFilterer{contract: contract}}, nil
}

// NewVaultCaller creates a new read-only instance of Vault, bound to a specific deployed contract.
func NewVaultCaller(address common.Address, caller bind.ContractCaller) (*VaultCaller, error) {
	contract, err := bindVault(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &VaultCaller{contract: contract}, nil
}

// NewVaultTransactor creates a new write-only instance of Vault, bound to a specific deployed contract.
func NewVaultTransactor(address common.Address, transactor bind.ContractTransactor) (*VaultTransactor, error) {
	contract, err := bindVault(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &VaultTransactor{contract: contract}, nil
}

// NewVaultFilterer creates a new log filterer instance of Vault, bound to a specific deployed contract.
func NewVaultFilterer(address common.Address, filterer bind.ContractFilterer) (*VaultFilterer, error) {
	contract, err := bindVault(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &VaultFilterer{contract: contract}, nil
}

// bindVault binds a generic wrapper to an already deployed contract.
func bindVault(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := VaultMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Vault *VaultRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Vault.Contract.VaultCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Vault *VaultRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Vault.Contract.VaultTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Vault *VaultRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Vault.Contract.VaultTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Vault *VaultCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Vault.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Vault *VaultTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Vault.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Vault *VaultTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Vault.Contract.contract.Transact(opts, method, params...)
}

// IdleAssets is a free data retrieval call binding the contract method 0xe16b03a3.
//
// Solidity: function idleAssets() view returns(uint256)
func (_Vault *VaultCaller) IdleAssets(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Vault.contract.Call(opts, &out, "idleAssets")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// IdleAssets is a free data retrieval call binding the contract method 0xe16b03a3.
//
// Solidity: function idleAssets() view returns(uint256)
func (_Vault *VaultSession) IdleAssets() (*big.Int, error) {
	return _Vault.Contract.IdleAssets(&_Vault.CallOpts)
}

// IdleAssets is a free data retrieval call binding the contract method 0xe16b03a3.
//
// Solidity: function idleAssets() view returns(uint256)
func (_Vault *VaultCallerSession) IdleAssets() (*big.Int, error) {
	return _Vault.Contract.IdleAssets(&_Vault.CallOpts)
}

// IdleReserveBps is a free data retrieval call binding the contract method 0x97b4d1bb.
//
// Solidity: function idleReserveBps() view returns(uint16)
func (_Vault *VaultCaller) IdleReserveBps(opts *bind.CallOpts) (uint16, error) {
	var out []interface{}
	err := _Vault.contract.Call(opts, &out, "idleReserveBps")

	if err != nil {
		return *new(uint16), err
	}

	out0 := *abi.ConvertType(out[0], new(uint16)).(*uint16)

	return out0, err

}

// IdleReserveBps is a free data retrieval call binding the contract method 0x97b4d1bb.
//
// Solidity: function idleReserveBps() view returns(uint16)
func (_Vault *VaultSession) IdleReserveBps() (uint16, error) {
	return _Vault.Contract.IdleReserveBps(&_Vault.CallOpts)
}

// IdleReserveBps is a free data retrieval call binding the contract method 0x97b4d1bb.
//
// Solidity: function idleReserveBps() view returns(uint16)
func (_Vault *VaultCallerSession) IdleReserveBps() (uint16, error) {
	return _Vault.Contract.IdleReserveBps(&_Vault.CallOpts)
}

// IsQuarantined is a free data retrieval call binding the contract method 0x5af11d65.
//
// Solidity: function isQuarantined(bytes32 strategyId) view returns(bool)
func (_Vault *VaultCaller) IsQuarantined(opts *bind.CallOpts, strategyId [32]byte) (bool, error) {
	var out []interface{}
	err := _Vault.contract.Call(opts, &out, "isQuarantined", strategyId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsQuarantined is a free data retrieval call binding the contract method 0x5af11d65.
//
// Solidity: function isQuarantined(bytes32 strategyId) view returns(bool)
func (_Vault *VaultSession) IsQuarantined(strategyId [32]byte) (bool, error) {
	return _Vault.Contract.IsQuarantined(&_Vault.CallOpts, strategyId)
}

// IsQuarantined is a free data retrieval call binding the contract method 0x5af11d65.
//
// Solidity: function isQuarantined(bytes32 strategyId) view returns(bool)
func (_Vault *VaultCallerSession) IsQuarantined(strategyId [32]byte) (bool, error) {
	return _Vault.Contract.IsQuarantined(&_Vault.CallOpts, strategyId)
}

// MaxRebalanceDelta is a free data retrieval call binding the contract method 0xee1f4f4f.
//
// Solidity: function maxRebalanceDelta() view returns(uint16)
func (_Vault *VaultCaller) MaxRebalanceDelta(opts *bind.CallOpts) (uint16, error) {
	var out []interface{}
	err := _Vault.contract.Call(opts, &out, "maxRebalanceDelta")

	if err != nil {
		return *new(uint16), err
	}

	out0 := *abi.ConvertType(out[0], new(uint16)).(*uint16)

	return out0, err

}

// MaxRebalanceDelta is a free data retrieval call binding the contract method 0xee1f4f4f.
//
// Solidity: function maxRebalanceDelta() view returns(uint16)
func (_Vault *VaultSession) MaxRebalanceDelta() (uint16, error) {
	return _Vault.Contract.MaxRebalanceDelta(&_Vault.CallOpts)
}

// MaxRebalanceDelta is a free data retrieval call binding the contract method 0xee1f4f4f.
//
// Solidity: function maxRebalanceDelta() view returns(uint16)
func (_Vault *VaultCallerSession) MaxRebalanceDelta() (uint16, error) {
	return _Vault.Contract.MaxRebalanceDelta(&_Vault.CallOpts)
}

// NextWithdrawRequestId is a free data retrieval call binding the contract method 0x8ca82108.
//
// Solidity: function nextWithdrawRequestId() view returns(uint256)
func (_Vault *VaultCaller) NextWithdrawRequestId(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Vault.contract.Call(opts, &out, "nextWithdrawRequestId")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// NextWithdrawRequestId is a free data retrieval call binding the contract method 0x8ca82108.
//
// Solidity: function nextWithdrawRequestId() view returns(uint256)
func (_Vault *VaultSession) NextWithdrawRequestId() (*big.Int, error) {
	return _Vault.Contract.NextWithdrawRequestId(&_Vault.CallOpts)
}

// NextWithdrawRequestId is a free data retrieval call binding the contract method 0x8ca82108.
//
// Solidity: function nextWithdrawRequestId() view returns(uint256)
func (_Vault *VaultCallerSession) NextWithdrawRequestId() (*big.Int, error) {
	return _Vault.Contract.NextWithdrawRequestId(&_Vault.CallOpts)
}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_Vault *VaultCaller) Paused(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _Vault.contract.Call(opts, &out, "paused")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_Vault *VaultSession) Paused() (bool, error) {
	return _Vault.Contract.Paused(&_Vault.CallOpts)
}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_Vault *VaultCallerSession) Paused() (bool, error) {
	return _Vault.Contract.Paused(&_Vault.CallOpts)
}

// PendingWithdrawShares is a free data retrieval call binding the contract method 0xdd72010b.
//
// Solidity: function pendingWithdrawShares() view returns(uint256)
func (_Vault *VaultCaller) PendingWithdrawShares(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Vault.contract.Call(opts, &out, "pendingWithdrawShares")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// PendingWithdrawShares is a free data retrieval call binding the contract method 0xdd72010b.
//
// Solidity: function pendingWithdrawShares() view returns(uint256)
func (_Vault *VaultSession) PendingWithdrawShares() (*big.Int, error) {
	return _Vault.Contract.PendingWithdrawShares(&_Vault.CallOpts)
}

// PendingWithdrawShares is a free data retrieval call binding the contract method 0xdd72010b.
//
// Solidity: function pendingWithdrawShares() view returns(uint256)
func (_Vault *VaultCallerSession) PendingWithdrawShares() (*big.Int, error) {
	return _Vault.Contract.PendingWithdrawShares(&_Vault.CallOpts)
}

// Strategies is a free data retrieval call binding the contract method 0xd9f9027f.
//
// Solidity: function strategies() view returns(bytes32[])
func (_Vault *VaultCaller) Strategies(opts *bind.CallOpts) ([][32]byte, error) {
	var out []interface{}
	err := _Vault.contract.Call(opts, &out, "strategies")

	if err != nil {
		return *new([][32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([][32]byte)).(*[][32]byte)

	return out0, err

}

// Strategies is a free data retrieval call binding the contract method 0xd9f9027f.
//
// Solidity: function strategies() view returns(bytes32[])
func (_Vault *VaultSession) Strategies() ([][32]byte, error) {
	return _Vault.Contract.Strategies(&_Vault.CallOpts)
}

// Strategies is a free data retrieval call binding the contract method 0xd9f9027f.
//
// Solidity: function strategies() view returns(bytes32[])
func (_Vault *VaultCallerSession) Strategies() ([][32]byte, error) {
	return _Vault.Contract.Strategies(&_Vault.CallOpts)
}

// StrategyCap is a free data retrieval call binding the contract method 0x22caa435.
//
// Solidity: function strategyCap(bytes32 strategyId) view returns(uint16)
func (_Vault *VaultCaller) StrategyCap(opts *bind.CallOpts, strategyId [32]byte) (uint16, error) {
	var out []interface{}
	err := _Vault.contract.Call(opts, &out, "strategyCap", strategyId)

	if err != nil {
		return *new(uint16), err
	}

	out0 := *abi.ConvertType(out[0], new(uint16)).(*uint16)

	return out0, err

}

// StrategyCap is a free data retrieval call binding the contract method 0x22caa435.
//
// Solidity: function strategyCap(bytes32 strategyId) view returns(uint16)
func (_Vault *VaultSession) StrategyCap(strategyId [32]byte) (uint16, error) {
	return _Vault.Contract.StrategyCap(&_Vault.CallOpts, strategyId)
}

// StrategyCap is a free data retrieval call binding the contract method 0x22caa435.
//
// Solidity: function strategyCap(bytes32 strategyId) view returns(uint16)
func (_Vault *VaultCallerSession) StrategyCap(strategyId [32]byte) (uint16, error) {
	return _Vault.Contract.StrategyCap(&_Vault.CallOpts, strategyId)
}

// StrategyTotalAssets is a free data retrieval call binding the contract method 0xad5b130c.
//
// Solidity: function strategyTotalAssets(bytes32 strategyId) view returns(uint256)
func (_Vault *VaultCaller) StrategyTotalAssets(opts *bind.CallOpts, strategyId [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _Vault.contract.Call(opts, &out, "strategyTotalAssets", strategyId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// StrategyTotalAssets is a free data retrieval call binding the contract method 0xad5b130c.
//
// Solidity: function strategyTotalAssets(bytes32 strategyId) view returns(uint256)
func (_Vault *VaultSession) StrategyTotalAssets(strategyId [32]byte) (*big.Int, error) {
	return _Vault.Contract.StrategyTotalAssets(&_Vault.CallOpts, strategyId)
}

// StrategyTotalAssets is a free data retrieval call binding the contract method 0xad5b130c.
//
// Solidity: function strategyTotalAssets(bytes32 strategyId) view returns(uint256)
func (_Vault *VaultCallerSession) StrategyTotalAssets(strategyId [32]byte) (*big.Int, error) {
	return _Vault.Contract.StrategyTotalAssets(&_Vault.CallOpts, strategyId)
}

// TargetAllocation is a free data retrieval call binding the contract method 0x8550c9ea.
//
// Solidity: function targetAllocation(bytes32 strategyId) view returns(uint16)
func (_Vault *VaultCaller) TargetAllocation(opts *bind.CallOpts, strategyId [32]byte) (uint16, error) {
	var out []interface{}
	err := _Vault.contract.Call(opts, &out, "targetAllocation", strategyId)

	if err != nil {
		return *new(uint16), err
	}

	out0 := *abi.ConvertType(out[0], new(uint16)).(*uint16)

	return out0, err

}

// TargetAllocation is a free data retrieval call binding the contract method 0x8550c9ea.
//
// Solidity: function targetAllocation(bytes32 strategyId) view returns(uint16)
func (_Vault *VaultSession) TargetAllocation(strategyId [32]byte) (uint16, error) {
	return _Vault.Contract.TargetAllocation(&_Vault.CallOpts, strategyId)
}

// TargetAllocation is a free data retrieval call binding the contract method 0x8550c9ea.
//
// Solidity: function targetAllocation(bytes32 strategyId) view returns(uint16)
func (_Vault *VaultCallerSession) TargetAllocation(strategyId [32]byte) (uint16, error) {
	return _Vault.Contract.TargetAllocation(&_Vault.CallOpts, strategyId)
}

// TotalAssets is a free data retrieval call binding the contract method 0x01e1d114.
//
// Solidity: function totalAssets() view returns(uint256)
func (_Vault *VaultCaller) TotalAssets(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Vault.contract.Call(opts, &out, "totalAssets")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalAssets is a free data retrieval call binding the contract method 0x01e1d114.
//
// Solidity: function totalAssets() view returns(uint256)
func (_Vault *VaultSession) TotalAssets() (*big.Int, error) {
	return _Vault.Contract.TotalAssets(&_Vault.CallOpts)
}

// TotalAssets is a free data retrieval call binding the contract method 0x01e1d114.
//
// Solidity: function totalAssets() view returns(uint256)
func (_Vault *VaultCallerSession) TotalAssets() (*big.Int, error) {
	return _Vault.Contract.TotalAssets(&_Vault.CallOpts)
}

// WithdrawRequest is a free data retrieval call binding the contract method 0x74899a7e.
//
// Solidity: function withdrawRequest(uint256 id) view returns((address,address,uint256))
func (_Vault *VaultCaller) WithdrawRequest(opts *bind.CallOpts, id *big.Int) (LibWithdrawQueueWithdrawRequest, error) {
	var out []interface{}
	err := _Vault.contract.Call(opts, &out, "withdrawRequest", id)

	if err != nil {
		return *new(LibWithdrawQueueWithdrawRequest), err
	}

	out0 := *abi.ConvertType(out[0], new(LibWithdrawQueueWithdrawRequest)).(*LibWithdrawQueueWithdrawRequest)

	return out0, err

}

// WithdrawRequest is a free data retrieval call binding the contract method 0x74899a7e.
//
// Solidity: function withdrawRequest(uint256 id) view returns((address,address,uint256))
func (_Vault *VaultSession) WithdrawRequest(id *big.Int) (LibWithdrawQueueWithdrawRequest, error) {
	return _Vault.Contract.WithdrawRequest(&_Vault.CallOpts, id)
}

// WithdrawRequest is a free data retrieval call binding the contract method 0x74899a7e.
//
// Solidity: function withdrawRequest(uint256 id) view returns((address,address,uint256))
func (_Vault *VaultCallerSession) WithdrawRequest(id *big.Int) (LibWithdrawQueueWithdrawRequest, error) {
	return _Vault.Contract.WithdrawRequest(&_Vault.CallOpts, id)
}

// FulfillWithdraw is a paid mutator transaction binding the contract method 0x4016d436.
//
// Solidity: function fulfillWithdraw(uint256 id) returns()
func (_Vault *VaultTransactor) FulfillWithdraw(opts *bind.TransactOpts, id *big.Int) (*types.Transaction, error) {
	return _Vault.contract.Transact(opts, "fulfillWithdraw", id)
}

// FulfillWithdraw is a paid mutator transaction binding the contract method 0x4016d436.
//
// Solidity: function fulfillWithdraw(uint256 id) returns()
func (_Vault *VaultSession) FulfillWithdraw(id *big.Int) (*types.Transaction, error) {
	return _Vault.Contract.FulfillWithdraw(&_Vault.TransactOpts, id)
}

// FulfillWithdraw is a paid mutator transaction binding the contract method 0x4016d436.
//
// Solidity: function fulfillWithdraw(uint256 id) returns()
func (_Vault *VaultTransactorSession) FulfillWithdraw(id *big.Int) (*types.Transaction, error) {
	return _Vault.Contract.FulfillWithdraw(&_Vault.TransactOpts, id)
}

// GuardCheckpoint is a paid mutator transaction binding the contract method 0xbe92c2b1.
//
// Solidity: function guardCheckpoint() returns()
func (_Vault *VaultTransactor) GuardCheckpoint(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Vault.contract.Transact(opts, "guardCheckpoint")
}

// GuardCheckpoint is a paid mutator transaction binding the contract method 0xbe92c2b1.
//
// Solidity: function guardCheckpoint() returns()
func (_Vault *VaultSession) GuardCheckpoint() (*types.Transaction, error) {
	return _Vault.Contract.GuardCheckpoint(&_Vault.TransactOpts)
}

// GuardCheckpoint is a paid mutator transaction binding the contract method 0xbe92c2b1.
//
// Solidity: function guardCheckpoint() returns()
func (_Vault *VaultTransactorSession) GuardCheckpoint() (*types.Transaction, error) {
	return _Vault.Contract.GuardCheckpoint(&_Vault.TransactOpts)
}

// HarvestAll is a paid mutator transaction binding the contract method 0x8ed955b9.
//
// Solidity: function harvestAll() returns()
func (_Vault *VaultTransactor) HarvestAll(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Vault.contract.Transact(opts, "harvestAll")
}

// HarvestAll is a paid mutator transaction binding the contract method 0x8ed955b9.
//
// Solidity: function harvestAll() returns()
func (_Vault *VaultSession) HarvestAll() (*types.Transaction, error) {
	return _Vault.Contract.HarvestAll(&_Vault.TransactOpts)
}

// HarvestAll is a paid mutator transaction binding the contract method 0x8ed955b9.
//
// Solidity: function harvestAll() returns()
func (_Vault *VaultTransactorSession) HarvestAll() (*types.Transaction, error) {
	return _Vault.Contract.HarvestAll(&_Vault.TransactOpts)
}

// Rebalance is a paid mutator transaction binding the contract method 0x7d7c2a1c.
//
// Solidity: function rebalance() returns()
func (_Vault *VaultTransactor) Rebalance(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Vault.contract.Transact(opts, "rebalance")
}

// Rebalance is a paid mutator transaction binding the contract method 0x7d7c2a1c.
//
// Solidity: function rebalance() returns()
func (_Vault *VaultSession) Rebalance() (*types.Transaction, error) {
	return _Vault.Contract.Rebalance(&_Vault.TransactOpts)
}

// Rebalance is a paid mutator transaction binding the contract method 0x7d7c2a1c.
//
// Solidity: function rebalance() returns()
func (_Vault *VaultTransactorSession) Rebalance() (*types.Transaction, error) {
	return _Vault.Contract.Rebalance(&_Vault.TransactOpts)
}

// SetAllocation is a paid mutator transaction binding the contract method 0x87e05d21.
//
// Solidity: function setAllocation(bytes32[] strategyIds, uint16[] bps) returns()
func (_Vault *VaultTransactor) SetAllocation(opts *bind.TransactOpts, strategyIds [][32]byte, bps []uint16) (*types.Transaction, error) {
	return _Vault.contract.Transact(opts, "setAllocation", strategyIds, bps)
}

// SetAllocation is a paid mutator transaction binding the contract method 0x87e05d21.
//
// Solidity: function setAllocation(bytes32[] strategyIds, uint16[] bps) returns()
func (_Vault *VaultSession) SetAllocation(strategyIds [][32]byte, bps []uint16) (*types.Transaction, error) {
	return _Vault.Contract.SetAllocation(&_Vault.TransactOpts, strategyIds, bps)
}

// SetAllocation is a paid mutator transaction binding the contract method 0x87e05d21.
//
// Solidity: function setAllocation(bytes32[] strategyIds, uint16[] bps) returns()
func (_Vault *VaultTransactorSession) SetAllocation(strategyIds [][32]byte, bps []uint16) (*types.Transaction, error) {
	return _Vault.Contract.SetAllocation(&_Vault.TransactOpts, strategyIds, bps)
}
