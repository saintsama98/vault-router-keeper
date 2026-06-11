// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package aavedata

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

// AaveDataMetaData contains all meta data concerning the AaveData contract.
var AaveDataMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"stateMutability\":\"view\",\"name\":\"getReserveData\",\"inputs\":[{\"name\":\"asset\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"unbacked\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"accruedToTreasuryScaled\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"totalAToken\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"totalStableDebt\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"totalVariableDebt\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"liquidityRate\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"variableBorrowRate\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"stableBorrowRate\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"averageStableBorrowRate\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"liquidityIndex\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"variableBorrowIndex\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"lastUpdateTimestamp\",\"type\":\"uint40\",\"internalType\":\"uint40\"}]},{\"type\":\"function\",\"stateMutability\":\"view\",\"name\":\"getReserveConfigurationData\",\"inputs\":[{\"name\":\"asset\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"decimals\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"ltv\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"liquidationThreshold\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"liquidationBonus\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"reserveFactor\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"usageAsCollateralEnabled\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"borrowingEnabled\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"stableBorrowRateEnabled\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"isActive\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"isFrozen\",\"type\":\"bool\",\"internalType\":\"bool\"}]},{\"type\":\"function\",\"stateMutability\":\"view\",\"name\":\"getReserveCaps\",\"inputs\":[{\"name\":\"asset\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"borrowCap\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"supplyCap\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"function\",\"stateMutability\":\"view\",\"name\":\"getPaused\",\"inputs\":[{\"name\":\"asset\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"isPaused\",\"type\":\"bool\",\"internalType\":\"bool\"}]}]",
}

// AaveDataABI is the input ABI used to generate the binding from.
// Deprecated: Use AaveDataMetaData.ABI instead.
var AaveDataABI = AaveDataMetaData.ABI

// AaveData is an auto generated Go binding around an Ethereum contract.
type AaveData struct {
	AaveDataCaller     // Read-only binding to the contract
	AaveDataTransactor // Write-only binding to the contract
	AaveDataFilterer   // Log filterer for contract events
}

// AaveDataCaller is an auto generated read-only Go binding around an Ethereum contract.
type AaveDataCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AaveDataTransactor is an auto generated write-only Go binding around an Ethereum contract.
type AaveDataTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AaveDataFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type AaveDataFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AaveDataSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type AaveDataSession struct {
	Contract     *AaveData         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// AaveDataCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type AaveDataCallerSession struct {
	Contract *AaveDataCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// AaveDataTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type AaveDataTransactorSession struct {
	Contract     *AaveDataTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// AaveDataRaw is an auto generated low-level Go binding around an Ethereum contract.
type AaveDataRaw struct {
	Contract *AaveData // Generic contract binding to access the raw methods on
}

// AaveDataCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type AaveDataCallerRaw struct {
	Contract *AaveDataCaller // Generic read-only contract binding to access the raw methods on
}

// AaveDataTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type AaveDataTransactorRaw struct {
	Contract *AaveDataTransactor // Generic write-only contract binding to access the raw methods on
}

// NewAaveData creates a new instance of AaveData, bound to a specific deployed contract.
func NewAaveData(address common.Address, backend bind.ContractBackend) (*AaveData, error) {
	contract, err := bindAaveData(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &AaveData{AaveDataCaller: AaveDataCaller{contract: contract}, AaveDataTransactor: AaveDataTransactor{contract: contract}, AaveDataFilterer: AaveDataFilterer{contract: contract}}, nil
}

// NewAaveDataCaller creates a new read-only instance of AaveData, bound to a specific deployed contract.
func NewAaveDataCaller(address common.Address, caller bind.ContractCaller) (*AaveDataCaller, error) {
	contract, err := bindAaveData(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &AaveDataCaller{contract: contract}, nil
}

// NewAaveDataTransactor creates a new write-only instance of AaveData, bound to a specific deployed contract.
func NewAaveDataTransactor(address common.Address, transactor bind.ContractTransactor) (*AaveDataTransactor, error) {
	contract, err := bindAaveData(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &AaveDataTransactor{contract: contract}, nil
}

// NewAaveDataFilterer creates a new log filterer instance of AaveData, bound to a specific deployed contract.
func NewAaveDataFilterer(address common.Address, filterer bind.ContractFilterer) (*AaveDataFilterer, error) {
	contract, err := bindAaveData(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &AaveDataFilterer{contract: contract}, nil
}

// bindAaveData binds a generic wrapper to an already deployed contract.
func bindAaveData(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := AaveDataMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AaveData *AaveDataRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AaveData.Contract.AaveDataCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AaveData *AaveDataRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AaveData.Contract.AaveDataTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AaveData *AaveDataRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AaveData.Contract.AaveDataTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AaveData *AaveDataCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AaveData.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AaveData *AaveDataTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AaveData.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AaveData *AaveDataTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AaveData.Contract.contract.Transact(opts, method, params...)
}

// GetPaused is a free data retrieval call binding the contract method 0xb55d9904.
//
// Solidity: function getPaused(address asset) view returns(bool isPaused)
func (_AaveData *AaveDataCaller) GetPaused(opts *bind.CallOpts, asset common.Address) (bool, error) {
	var out []interface{}
	err := _AaveData.contract.Call(opts, &out, "getPaused", asset)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// GetPaused is a free data retrieval call binding the contract method 0xb55d9904.
//
// Solidity: function getPaused(address asset) view returns(bool isPaused)
func (_AaveData *AaveDataSession) GetPaused(asset common.Address) (bool, error) {
	return _AaveData.Contract.GetPaused(&_AaveData.CallOpts, asset)
}

// GetPaused is a free data retrieval call binding the contract method 0xb55d9904.
//
// Solidity: function getPaused(address asset) view returns(bool isPaused)
func (_AaveData *AaveDataCallerSession) GetPaused(asset common.Address) (bool, error) {
	return _AaveData.Contract.GetPaused(&_AaveData.CallOpts, asset)
}

// GetReserveCaps is a free data retrieval call binding the contract method 0x46fbe558.
//
// Solidity: function getReserveCaps(address asset) view returns(uint256 borrowCap, uint256 supplyCap)
func (_AaveData *AaveDataCaller) GetReserveCaps(opts *bind.CallOpts, asset common.Address) (struct {
	BorrowCap *big.Int
	SupplyCap *big.Int
}, error) {
	var out []interface{}
	err := _AaveData.contract.Call(opts, &out, "getReserveCaps", asset)

	outstruct := new(struct {
		BorrowCap *big.Int
		SupplyCap *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.BorrowCap = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.SupplyCap = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// GetReserveCaps is a free data retrieval call binding the contract method 0x46fbe558.
//
// Solidity: function getReserveCaps(address asset) view returns(uint256 borrowCap, uint256 supplyCap)
func (_AaveData *AaveDataSession) GetReserveCaps(asset common.Address) (struct {
	BorrowCap *big.Int
	SupplyCap *big.Int
}, error) {
	return _AaveData.Contract.GetReserveCaps(&_AaveData.CallOpts, asset)
}

// GetReserveCaps is a free data retrieval call binding the contract method 0x46fbe558.
//
// Solidity: function getReserveCaps(address asset) view returns(uint256 borrowCap, uint256 supplyCap)
func (_AaveData *AaveDataCallerSession) GetReserveCaps(asset common.Address) (struct {
	BorrowCap *big.Int
	SupplyCap *big.Int
}, error) {
	return _AaveData.Contract.GetReserveCaps(&_AaveData.CallOpts, asset)
}

// GetReserveConfigurationData is a free data retrieval call binding the contract method 0x3e150141.
//
// Solidity: function getReserveConfigurationData(address asset) view returns(uint256 decimals, uint256 ltv, uint256 liquidationThreshold, uint256 liquidationBonus, uint256 reserveFactor, bool usageAsCollateralEnabled, bool borrowingEnabled, bool stableBorrowRateEnabled, bool isActive, bool isFrozen)
func (_AaveData *AaveDataCaller) GetReserveConfigurationData(opts *bind.CallOpts, asset common.Address) (struct {
	Decimals                 *big.Int
	Ltv                      *big.Int
	LiquidationThreshold     *big.Int
	LiquidationBonus         *big.Int
	ReserveFactor            *big.Int
	UsageAsCollateralEnabled bool
	BorrowingEnabled         bool
	StableBorrowRateEnabled  bool
	IsActive                 bool
	IsFrozen                 bool
}, error) {
	var out []interface{}
	err := _AaveData.contract.Call(opts, &out, "getReserveConfigurationData", asset)

	outstruct := new(struct {
		Decimals                 *big.Int
		Ltv                      *big.Int
		LiquidationThreshold     *big.Int
		LiquidationBonus         *big.Int
		ReserveFactor            *big.Int
		UsageAsCollateralEnabled bool
		BorrowingEnabled         bool
		StableBorrowRateEnabled  bool
		IsActive                 bool
		IsFrozen                 bool
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Decimals = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.Ltv = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.LiquidationThreshold = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.LiquidationBonus = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.ReserveFactor = *abi.ConvertType(out[4], new(*big.Int)).(**big.Int)
	outstruct.UsageAsCollateralEnabled = *abi.ConvertType(out[5], new(bool)).(*bool)
	outstruct.BorrowingEnabled = *abi.ConvertType(out[6], new(bool)).(*bool)
	outstruct.StableBorrowRateEnabled = *abi.ConvertType(out[7], new(bool)).(*bool)
	outstruct.IsActive = *abi.ConvertType(out[8], new(bool)).(*bool)
	outstruct.IsFrozen = *abi.ConvertType(out[9], new(bool)).(*bool)

	return *outstruct, err

}

// GetReserveConfigurationData is a free data retrieval call binding the contract method 0x3e150141.
//
// Solidity: function getReserveConfigurationData(address asset) view returns(uint256 decimals, uint256 ltv, uint256 liquidationThreshold, uint256 liquidationBonus, uint256 reserveFactor, bool usageAsCollateralEnabled, bool borrowingEnabled, bool stableBorrowRateEnabled, bool isActive, bool isFrozen)
func (_AaveData *AaveDataSession) GetReserveConfigurationData(asset common.Address) (struct {
	Decimals                 *big.Int
	Ltv                      *big.Int
	LiquidationThreshold     *big.Int
	LiquidationBonus         *big.Int
	ReserveFactor            *big.Int
	UsageAsCollateralEnabled bool
	BorrowingEnabled         bool
	StableBorrowRateEnabled  bool
	IsActive                 bool
	IsFrozen                 bool
}, error) {
	return _AaveData.Contract.GetReserveConfigurationData(&_AaveData.CallOpts, asset)
}

// GetReserveConfigurationData is a free data retrieval call binding the contract method 0x3e150141.
//
// Solidity: function getReserveConfigurationData(address asset) view returns(uint256 decimals, uint256 ltv, uint256 liquidationThreshold, uint256 liquidationBonus, uint256 reserveFactor, bool usageAsCollateralEnabled, bool borrowingEnabled, bool stableBorrowRateEnabled, bool isActive, bool isFrozen)
func (_AaveData *AaveDataCallerSession) GetReserveConfigurationData(asset common.Address) (struct {
	Decimals                 *big.Int
	Ltv                      *big.Int
	LiquidationThreshold     *big.Int
	LiquidationBonus         *big.Int
	ReserveFactor            *big.Int
	UsageAsCollateralEnabled bool
	BorrowingEnabled         bool
	StableBorrowRateEnabled  bool
	IsActive                 bool
	IsFrozen                 bool
}, error) {
	return _AaveData.Contract.GetReserveConfigurationData(&_AaveData.CallOpts, asset)
}

// GetReserveData is a free data retrieval call binding the contract method 0x35ea6a75.
//
// Solidity: function getReserveData(address asset) view returns(uint256 unbacked, uint256 accruedToTreasuryScaled, uint256 totalAToken, uint256 totalStableDebt, uint256 totalVariableDebt, uint256 liquidityRate, uint256 variableBorrowRate, uint256 stableBorrowRate, uint256 averageStableBorrowRate, uint256 liquidityIndex, uint256 variableBorrowIndex, uint40 lastUpdateTimestamp)
func (_AaveData *AaveDataCaller) GetReserveData(opts *bind.CallOpts, asset common.Address) (struct {
	Unbacked                *big.Int
	AccruedToTreasuryScaled *big.Int
	TotalAToken             *big.Int
	TotalStableDebt         *big.Int
	TotalVariableDebt       *big.Int
	LiquidityRate           *big.Int
	VariableBorrowRate      *big.Int
	StableBorrowRate        *big.Int
	AverageStableBorrowRate *big.Int
	LiquidityIndex          *big.Int
	VariableBorrowIndex     *big.Int
	LastUpdateTimestamp     *big.Int
}, error) {
	var out []interface{}
	err := _AaveData.contract.Call(opts, &out, "getReserveData", asset)

	outstruct := new(struct {
		Unbacked                *big.Int
		AccruedToTreasuryScaled *big.Int
		TotalAToken             *big.Int
		TotalStableDebt         *big.Int
		TotalVariableDebt       *big.Int
		LiquidityRate           *big.Int
		VariableBorrowRate      *big.Int
		StableBorrowRate        *big.Int
		AverageStableBorrowRate *big.Int
		LiquidityIndex          *big.Int
		VariableBorrowIndex     *big.Int
		LastUpdateTimestamp     *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Unbacked = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.AccruedToTreasuryScaled = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.TotalAToken = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.TotalStableDebt = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.TotalVariableDebt = *abi.ConvertType(out[4], new(*big.Int)).(**big.Int)
	outstruct.LiquidityRate = *abi.ConvertType(out[5], new(*big.Int)).(**big.Int)
	outstruct.VariableBorrowRate = *abi.ConvertType(out[6], new(*big.Int)).(**big.Int)
	outstruct.StableBorrowRate = *abi.ConvertType(out[7], new(*big.Int)).(**big.Int)
	outstruct.AverageStableBorrowRate = *abi.ConvertType(out[8], new(*big.Int)).(**big.Int)
	outstruct.LiquidityIndex = *abi.ConvertType(out[9], new(*big.Int)).(**big.Int)
	outstruct.VariableBorrowIndex = *abi.ConvertType(out[10], new(*big.Int)).(**big.Int)
	outstruct.LastUpdateTimestamp = *abi.ConvertType(out[11], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// GetReserveData is a free data retrieval call binding the contract method 0x35ea6a75.
//
// Solidity: function getReserveData(address asset) view returns(uint256 unbacked, uint256 accruedToTreasuryScaled, uint256 totalAToken, uint256 totalStableDebt, uint256 totalVariableDebt, uint256 liquidityRate, uint256 variableBorrowRate, uint256 stableBorrowRate, uint256 averageStableBorrowRate, uint256 liquidityIndex, uint256 variableBorrowIndex, uint40 lastUpdateTimestamp)
func (_AaveData *AaveDataSession) GetReserveData(asset common.Address) (struct {
	Unbacked                *big.Int
	AccruedToTreasuryScaled *big.Int
	TotalAToken             *big.Int
	TotalStableDebt         *big.Int
	TotalVariableDebt       *big.Int
	LiquidityRate           *big.Int
	VariableBorrowRate      *big.Int
	StableBorrowRate        *big.Int
	AverageStableBorrowRate *big.Int
	LiquidityIndex          *big.Int
	VariableBorrowIndex     *big.Int
	LastUpdateTimestamp     *big.Int
}, error) {
	return _AaveData.Contract.GetReserveData(&_AaveData.CallOpts, asset)
}

// GetReserveData is a free data retrieval call binding the contract method 0x35ea6a75.
//
// Solidity: function getReserveData(address asset) view returns(uint256 unbacked, uint256 accruedToTreasuryScaled, uint256 totalAToken, uint256 totalStableDebt, uint256 totalVariableDebt, uint256 liquidityRate, uint256 variableBorrowRate, uint256 stableBorrowRate, uint256 averageStableBorrowRate, uint256 liquidityIndex, uint256 variableBorrowIndex, uint40 lastUpdateTimestamp)
func (_AaveData *AaveDataCallerSession) GetReserveData(asset common.Address) (struct {
	Unbacked                *big.Int
	AccruedToTreasuryScaled *big.Int
	TotalAToken             *big.Int
	TotalStableDebt         *big.Int
	TotalVariableDebt       *big.Int
	LiquidityRate           *big.Int
	VariableBorrowRate      *big.Int
	StableBorrowRate        *big.Int
	AverageStableBorrowRate *big.Int
	LiquidityIndex          *big.Int
	VariableBorrowIndex     *big.Int
	LastUpdateTimestamp     *big.Int
}, error) {
	return _AaveData.Contract.GetReserveData(&_AaveData.CallOpts, asset)
}
