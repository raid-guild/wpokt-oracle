// Code generated by mockery v2.43.2. DO NOT EDIT.

package client

import (
	big "math/big"

	common "github.com/ethereum/go-ethereum/common"

	context "context"

	ethereum "github.com/ethereum/go-ethereum"

	mock "github.com/stretchr/testify/mock"

	types "github.com/ethereum/go-ethereum/core/types"
)

// MockEthclientClient is an autogenerated mock type for the EthclientClient type
type MockEthclientClient struct {
	mock.Mock
}

type MockEthclientClient_Expecter struct {
	mock *mock.Mock
}

func (_m *MockEthclientClient) EXPECT() *MockEthclientClient_Expecter {
	return &MockEthclientClient_Expecter{mock: &_m.Mock}
}

// BlockNumber provides a mock function with given fields: ctx
func (_m *MockEthclientClient) BlockNumber(ctx context.Context) (uint64, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for BlockNumber")
	}

	var r0 uint64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (uint64, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) uint64); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Get(0).(uint64)
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockEthclientClient_BlockNumber_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'BlockNumber'
type MockEthclientClient_BlockNumber_Call struct {
	*mock.Call
}

// BlockNumber is a helper method to define mock.On call
//   - ctx context.Context
func (_e *MockEthclientClient_Expecter) BlockNumber(ctx interface{}) *MockEthclientClient_BlockNumber_Call {
	return &MockEthclientClient_BlockNumber_Call{Call: _e.mock.On("BlockNumber", ctx)}
}

func (_c *MockEthclientClient_BlockNumber_Call) Run(run func(ctx context.Context)) *MockEthclientClient_BlockNumber_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *MockEthclientClient_BlockNumber_Call) Return(_a0 uint64, _a1 error) *MockEthclientClient_BlockNumber_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockEthclientClient_BlockNumber_Call) RunAndReturn(run func(context.Context) (uint64, error)) *MockEthclientClient_BlockNumber_Call {
	_c.Call.Return(run)
	return _c
}

// CallContract provides a mock function with given fields: ctx, call, blockNumber
func (_m *MockEthclientClient) CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	ret := _m.Called(ctx, call, blockNumber)

	if len(ret) == 0 {
		panic("no return value specified for CallContract")
	}

	var r0 []byte
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, ethereum.CallMsg, *big.Int) ([]byte, error)); ok {
		return rf(ctx, call, blockNumber)
	}
	if rf, ok := ret.Get(0).(func(context.Context, ethereum.CallMsg, *big.Int) []byte); ok {
		r0 = rf(ctx, call, blockNumber)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, ethereum.CallMsg, *big.Int) error); ok {
		r1 = rf(ctx, call, blockNumber)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockEthclientClient_CallContract_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CallContract'
type MockEthclientClient_CallContract_Call struct {
	*mock.Call
}

// CallContract is a helper method to define mock.On call
//   - ctx context.Context
//   - call ethereum.CallMsg
//   - blockNumber *big.Int
func (_e *MockEthclientClient_Expecter) CallContract(ctx interface{}, call interface{}, blockNumber interface{}) *MockEthclientClient_CallContract_Call {
	return &MockEthclientClient_CallContract_Call{Call: _e.mock.On("CallContract", ctx, call, blockNumber)}
}

func (_c *MockEthclientClient_CallContract_Call) Run(run func(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int)) *MockEthclientClient_CallContract_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(ethereum.CallMsg), args[2].(*big.Int))
	})
	return _c
}

func (_c *MockEthclientClient_CallContract_Call) Return(_a0 []byte, _a1 error) *MockEthclientClient_CallContract_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockEthclientClient_CallContract_Call) RunAndReturn(run func(context.Context, ethereum.CallMsg, *big.Int) ([]byte, error)) *MockEthclientClient_CallContract_Call {
	_c.Call.Return(run)
	return _c
}

// ChainID provides a mock function with given fields: ctx
func (_m *MockEthclientClient) ChainID(ctx context.Context) (*big.Int, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for ChainID")
	}

	var r0 *big.Int
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (*big.Int, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) *big.Int); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*big.Int)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockEthclientClient_ChainID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ChainID'
type MockEthclientClient_ChainID_Call struct {
	*mock.Call
}

// ChainID is a helper method to define mock.On call
//   - ctx context.Context
func (_e *MockEthclientClient_Expecter) ChainID(ctx interface{}) *MockEthclientClient_ChainID_Call {
	return &MockEthclientClient_ChainID_Call{Call: _e.mock.On("ChainID", ctx)}
}

func (_c *MockEthclientClient_ChainID_Call) Run(run func(ctx context.Context)) *MockEthclientClient_ChainID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *MockEthclientClient_ChainID_Call) Return(_a0 *big.Int, _a1 error) *MockEthclientClient_ChainID_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockEthclientClient_ChainID_Call) RunAndReturn(run func(context.Context) (*big.Int, error)) *MockEthclientClient_ChainID_Call {
	_c.Call.Return(run)
	return _c
}

// CodeAt provides a mock function with given fields: ctx, contract, blockNumber
func (_m *MockEthclientClient) CodeAt(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error) {
	ret := _m.Called(ctx, contract, blockNumber)

	if len(ret) == 0 {
		panic("no return value specified for CodeAt")
	}

	var r0 []byte
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, common.Address, *big.Int) ([]byte, error)); ok {
		return rf(ctx, contract, blockNumber)
	}
	if rf, ok := ret.Get(0).(func(context.Context, common.Address, *big.Int) []byte); ok {
		r0 = rf(ctx, contract, blockNumber)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, common.Address, *big.Int) error); ok {
		r1 = rf(ctx, contract, blockNumber)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockEthclientClient_CodeAt_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CodeAt'
type MockEthclientClient_CodeAt_Call struct {
	*mock.Call
}

// CodeAt is a helper method to define mock.On call
//   - ctx context.Context
//   - contract common.Address
//   - blockNumber *big.Int
func (_e *MockEthclientClient_Expecter) CodeAt(ctx interface{}, contract interface{}, blockNumber interface{}) *MockEthclientClient_CodeAt_Call {
	return &MockEthclientClient_CodeAt_Call{Call: _e.mock.On("CodeAt", ctx, contract, blockNumber)}
}

func (_c *MockEthclientClient_CodeAt_Call) Run(run func(ctx context.Context, contract common.Address, blockNumber *big.Int)) *MockEthclientClient_CodeAt_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(common.Address), args[2].(*big.Int))
	})
	return _c
}

func (_c *MockEthclientClient_CodeAt_Call) Return(_a0 []byte, _a1 error) *MockEthclientClient_CodeAt_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockEthclientClient_CodeAt_Call) RunAndReturn(run func(context.Context, common.Address, *big.Int) ([]byte, error)) *MockEthclientClient_CodeAt_Call {
	_c.Call.Return(run)
	return _c
}

// EstimateGas provides a mock function with given fields: ctx, call
func (_m *MockEthclientClient) EstimateGas(ctx context.Context, call ethereum.CallMsg) (uint64, error) {
	ret := _m.Called(ctx, call)

	if len(ret) == 0 {
		panic("no return value specified for EstimateGas")
	}

	var r0 uint64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, ethereum.CallMsg) (uint64, error)); ok {
		return rf(ctx, call)
	}
	if rf, ok := ret.Get(0).(func(context.Context, ethereum.CallMsg) uint64); ok {
		r0 = rf(ctx, call)
	} else {
		r0 = ret.Get(0).(uint64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, ethereum.CallMsg) error); ok {
		r1 = rf(ctx, call)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockEthclientClient_EstimateGas_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'EstimateGas'
type MockEthclientClient_EstimateGas_Call struct {
	*mock.Call
}

// EstimateGas is a helper method to define mock.On call
//   - ctx context.Context
//   - call ethereum.CallMsg
func (_e *MockEthclientClient_Expecter) EstimateGas(ctx interface{}, call interface{}) *MockEthclientClient_EstimateGas_Call {
	return &MockEthclientClient_EstimateGas_Call{Call: _e.mock.On("EstimateGas", ctx, call)}
}

func (_c *MockEthclientClient_EstimateGas_Call) Run(run func(ctx context.Context, call ethereum.CallMsg)) *MockEthclientClient_EstimateGas_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(ethereum.CallMsg))
	})
	return _c
}

func (_c *MockEthclientClient_EstimateGas_Call) Return(_a0 uint64, _a1 error) *MockEthclientClient_EstimateGas_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockEthclientClient_EstimateGas_Call) RunAndReturn(run func(context.Context, ethereum.CallMsg) (uint64, error)) *MockEthclientClient_EstimateGas_Call {
	_c.Call.Return(run)
	return _c
}

// FilterLogs provides a mock function with given fields: ctx, q
func (_m *MockEthclientClient) FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error) {
	ret := _m.Called(ctx, q)

	if len(ret) == 0 {
		panic("no return value specified for FilterLogs")
	}

	var r0 []types.Log
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, ethereum.FilterQuery) ([]types.Log, error)); ok {
		return rf(ctx, q)
	}
	if rf, ok := ret.Get(0).(func(context.Context, ethereum.FilterQuery) []types.Log); ok {
		r0 = rf(ctx, q)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]types.Log)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, ethereum.FilterQuery) error); ok {
		r1 = rf(ctx, q)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockEthclientClient_FilterLogs_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FilterLogs'
type MockEthclientClient_FilterLogs_Call struct {
	*mock.Call
}

// FilterLogs is a helper method to define mock.On call
//   - ctx context.Context
//   - q ethereum.FilterQuery
func (_e *MockEthclientClient_Expecter) FilterLogs(ctx interface{}, q interface{}) *MockEthclientClient_FilterLogs_Call {
	return &MockEthclientClient_FilterLogs_Call{Call: _e.mock.On("FilterLogs", ctx, q)}
}

func (_c *MockEthclientClient_FilterLogs_Call) Run(run func(ctx context.Context, q ethereum.FilterQuery)) *MockEthclientClient_FilterLogs_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(ethereum.FilterQuery))
	})
	return _c
}

func (_c *MockEthclientClient_FilterLogs_Call) Return(_a0 []types.Log, _a1 error) *MockEthclientClient_FilterLogs_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockEthclientClient_FilterLogs_Call) RunAndReturn(run func(context.Context, ethereum.FilterQuery) ([]types.Log, error)) *MockEthclientClient_FilterLogs_Call {
	_c.Call.Return(run)
	return _c
}

// HeaderByNumber provides a mock function with given fields: ctx, number
func (_m *MockEthclientClient) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	ret := _m.Called(ctx, number)

	if len(ret) == 0 {
		panic("no return value specified for HeaderByNumber")
	}

	var r0 *types.Header
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *big.Int) (*types.Header, error)); ok {
		return rf(ctx, number)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *big.Int) *types.Header); ok {
		r0 = rf(ctx, number)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.Header)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *big.Int) error); ok {
		r1 = rf(ctx, number)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockEthclientClient_HeaderByNumber_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'HeaderByNumber'
type MockEthclientClient_HeaderByNumber_Call struct {
	*mock.Call
}

// HeaderByNumber is a helper method to define mock.On call
//   - ctx context.Context
//   - number *big.Int
func (_e *MockEthclientClient_Expecter) HeaderByNumber(ctx interface{}, number interface{}) *MockEthclientClient_HeaderByNumber_Call {
	return &MockEthclientClient_HeaderByNumber_Call{Call: _e.mock.On("HeaderByNumber", ctx, number)}
}

func (_c *MockEthclientClient_HeaderByNumber_Call) Run(run func(ctx context.Context, number *big.Int)) *MockEthclientClient_HeaderByNumber_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*big.Int))
	})
	return _c
}

func (_c *MockEthclientClient_HeaderByNumber_Call) Return(_a0 *types.Header, _a1 error) *MockEthclientClient_HeaderByNumber_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockEthclientClient_HeaderByNumber_Call) RunAndReturn(run func(context.Context, *big.Int) (*types.Header, error)) *MockEthclientClient_HeaderByNumber_Call {
	_c.Call.Return(run)
	return _c
}

// PendingCodeAt provides a mock function with given fields: ctx, account
func (_m *MockEthclientClient) PendingCodeAt(ctx context.Context, account common.Address) ([]byte, error) {
	ret := _m.Called(ctx, account)

	if len(ret) == 0 {
		panic("no return value specified for PendingCodeAt")
	}

	var r0 []byte
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, common.Address) ([]byte, error)); ok {
		return rf(ctx, account)
	}
	if rf, ok := ret.Get(0).(func(context.Context, common.Address) []byte); ok {
		r0 = rf(ctx, account)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, common.Address) error); ok {
		r1 = rf(ctx, account)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockEthclientClient_PendingCodeAt_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'PendingCodeAt'
type MockEthclientClient_PendingCodeAt_Call struct {
	*mock.Call
}

// PendingCodeAt is a helper method to define mock.On call
//   - ctx context.Context
//   - account common.Address
func (_e *MockEthclientClient_Expecter) PendingCodeAt(ctx interface{}, account interface{}) *MockEthclientClient_PendingCodeAt_Call {
	return &MockEthclientClient_PendingCodeAt_Call{Call: _e.mock.On("PendingCodeAt", ctx, account)}
}

func (_c *MockEthclientClient_PendingCodeAt_Call) Run(run func(ctx context.Context, account common.Address)) *MockEthclientClient_PendingCodeAt_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(common.Address))
	})
	return _c
}

func (_c *MockEthclientClient_PendingCodeAt_Call) Return(_a0 []byte, _a1 error) *MockEthclientClient_PendingCodeAt_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockEthclientClient_PendingCodeAt_Call) RunAndReturn(run func(context.Context, common.Address) ([]byte, error)) *MockEthclientClient_PendingCodeAt_Call {
	_c.Call.Return(run)
	return _c
}

// PendingNonceAt provides a mock function with given fields: ctx, account
func (_m *MockEthclientClient) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	ret := _m.Called(ctx, account)

	if len(ret) == 0 {
		panic("no return value specified for PendingNonceAt")
	}

	var r0 uint64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, common.Address) (uint64, error)); ok {
		return rf(ctx, account)
	}
	if rf, ok := ret.Get(0).(func(context.Context, common.Address) uint64); ok {
		r0 = rf(ctx, account)
	} else {
		r0 = ret.Get(0).(uint64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, common.Address) error); ok {
		r1 = rf(ctx, account)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockEthclientClient_PendingNonceAt_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'PendingNonceAt'
type MockEthclientClient_PendingNonceAt_Call struct {
	*mock.Call
}

// PendingNonceAt is a helper method to define mock.On call
//   - ctx context.Context
//   - account common.Address
func (_e *MockEthclientClient_Expecter) PendingNonceAt(ctx interface{}, account interface{}) *MockEthclientClient_PendingNonceAt_Call {
	return &MockEthclientClient_PendingNonceAt_Call{Call: _e.mock.On("PendingNonceAt", ctx, account)}
}

func (_c *MockEthclientClient_PendingNonceAt_Call) Run(run func(ctx context.Context, account common.Address)) *MockEthclientClient_PendingNonceAt_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(common.Address))
	})
	return _c
}

func (_c *MockEthclientClient_PendingNonceAt_Call) Return(_a0 uint64, _a1 error) *MockEthclientClient_PendingNonceAt_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockEthclientClient_PendingNonceAt_Call) RunAndReturn(run func(context.Context, common.Address) (uint64, error)) *MockEthclientClient_PendingNonceAt_Call {
	_c.Call.Return(run)
	return _c
}

// SendTransaction provides a mock function with given fields: ctx, tx
func (_m *MockEthclientClient) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	ret := _m.Called(ctx, tx)

	if len(ret) == 0 {
		panic("no return value specified for SendTransaction")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *types.Transaction) error); ok {
		r0 = rf(ctx, tx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockEthclientClient_SendTransaction_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SendTransaction'
type MockEthclientClient_SendTransaction_Call struct {
	*mock.Call
}

// SendTransaction is a helper method to define mock.On call
//   - ctx context.Context
//   - tx *types.Transaction
func (_e *MockEthclientClient_Expecter) SendTransaction(ctx interface{}, tx interface{}) *MockEthclientClient_SendTransaction_Call {
	return &MockEthclientClient_SendTransaction_Call{Call: _e.mock.On("SendTransaction", ctx, tx)}
}

func (_c *MockEthclientClient_SendTransaction_Call) Run(run func(ctx context.Context, tx *types.Transaction)) *MockEthclientClient_SendTransaction_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*types.Transaction))
	})
	return _c
}

func (_c *MockEthclientClient_SendTransaction_Call) Return(_a0 error) *MockEthclientClient_SendTransaction_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockEthclientClient_SendTransaction_Call) RunAndReturn(run func(context.Context, *types.Transaction) error) *MockEthclientClient_SendTransaction_Call {
	_c.Call.Return(run)
	return _c
}

// SubscribeFilterLogs provides a mock function with given fields: ctx, q, ch
func (_m *MockEthclientClient) SubscribeFilterLogs(ctx context.Context, q ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error) {
	ret := _m.Called(ctx, q, ch)

	if len(ret) == 0 {
		panic("no return value specified for SubscribeFilterLogs")
	}

	var r0 ethereum.Subscription
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, ethereum.FilterQuery, chan<- types.Log) (ethereum.Subscription, error)); ok {
		return rf(ctx, q, ch)
	}
	if rf, ok := ret.Get(0).(func(context.Context, ethereum.FilterQuery, chan<- types.Log) ethereum.Subscription); ok {
		r0 = rf(ctx, q, ch)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(ethereum.Subscription)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, ethereum.FilterQuery, chan<- types.Log) error); ok {
		r1 = rf(ctx, q, ch)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockEthclientClient_SubscribeFilterLogs_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SubscribeFilterLogs'
type MockEthclientClient_SubscribeFilterLogs_Call struct {
	*mock.Call
}

// SubscribeFilterLogs is a helper method to define mock.On call
//   - ctx context.Context
//   - q ethereum.FilterQuery
//   - ch chan<- types.Log
func (_e *MockEthclientClient_Expecter) SubscribeFilterLogs(ctx interface{}, q interface{}, ch interface{}) *MockEthclientClient_SubscribeFilterLogs_Call {
	return &MockEthclientClient_SubscribeFilterLogs_Call{Call: _e.mock.On("SubscribeFilterLogs", ctx, q, ch)}
}

func (_c *MockEthclientClient_SubscribeFilterLogs_Call) Run(run func(ctx context.Context, q ethereum.FilterQuery, ch chan<- types.Log)) *MockEthclientClient_SubscribeFilterLogs_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(ethereum.FilterQuery), args[2].(chan<- types.Log))
	})
	return _c
}

func (_c *MockEthclientClient_SubscribeFilterLogs_Call) Return(_a0 ethereum.Subscription, _a1 error) *MockEthclientClient_SubscribeFilterLogs_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockEthclientClient_SubscribeFilterLogs_Call) RunAndReturn(run func(context.Context, ethereum.FilterQuery, chan<- types.Log) (ethereum.Subscription, error)) *MockEthclientClient_SubscribeFilterLogs_Call {
	_c.Call.Return(run)
	return _c
}

// SuggestGasPrice provides a mock function with given fields: ctx
func (_m *MockEthclientClient) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for SuggestGasPrice")
	}

	var r0 *big.Int
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (*big.Int, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) *big.Int); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*big.Int)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockEthclientClient_SuggestGasPrice_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SuggestGasPrice'
type MockEthclientClient_SuggestGasPrice_Call struct {
	*mock.Call
}

// SuggestGasPrice is a helper method to define mock.On call
//   - ctx context.Context
func (_e *MockEthclientClient_Expecter) SuggestGasPrice(ctx interface{}) *MockEthclientClient_SuggestGasPrice_Call {
	return &MockEthclientClient_SuggestGasPrice_Call{Call: _e.mock.On("SuggestGasPrice", ctx)}
}

func (_c *MockEthclientClient_SuggestGasPrice_Call) Run(run func(ctx context.Context)) *MockEthclientClient_SuggestGasPrice_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *MockEthclientClient_SuggestGasPrice_Call) Return(_a0 *big.Int, _a1 error) *MockEthclientClient_SuggestGasPrice_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockEthclientClient_SuggestGasPrice_Call) RunAndReturn(run func(context.Context) (*big.Int, error)) *MockEthclientClient_SuggestGasPrice_Call {
	_c.Call.Return(run)
	return _c
}

// SuggestGasTipCap provides a mock function with given fields: ctx
func (_m *MockEthclientClient) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for SuggestGasTipCap")
	}

	var r0 *big.Int
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (*big.Int, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) *big.Int); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*big.Int)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockEthclientClient_SuggestGasTipCap_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SuggestGasTipCap'
type MockEthclientClient_SuggestGasTipCap_Call struct {
	*mock.Call
}

// SuggestGasTipCap is a helper method to define mock.On call
//   - ctx context.Context
func (_e *MockEthclientClient_Expecter) SuggestGasTipCap(ctx interface{}) *MockEthclientClient_SuggestGasTipCap_Call {
	return &MockEthclientClient_SuggestGasTipCap_Call{Call: _e.mock.On("SuggestGasTipCap", ctx)}
}

func (_c *MockEthclientClient_SuggestGasTipCap_Call) Run(run func(ctx context.Context)) *MockEthclientClient_SuggestGasTipCap_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *MockEthclientClient_SuggestGasTipCap_Call) Return(_a0 *big.Int, _a1 error) *MockEthclientClient_SuggestGasTipCap_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockEthclientClient_SuggestGasTipCap_Call) RunAndReturn(run func(context.Context) (*big.Int, error)) *MockEthclientClient_SuggestGasTipCap_Call {
	_c.Call.Return(run)
	return _c
}

// TransactionByHash provides a mock function with given fields: ctx, hash
func (_m *MockEthclientClient) TransactionByHash(ctx context.Context, hash common.Hash) (*types.Transaction, bool, error) {
	ret := _m.Called(ctx, hash)

	if len(ret) == 0 {
		panic("no return value specified for TransactionByHash")
	}

	var r0 *types.Transaction
	var r1 bool
	var r2 error
	if rf, ok := ret.Get(0).(func(context.Context, common.Hash) (*types.Transaction, bool, error)); ok {
		return rf(ctx, hash)
	}
	if rf, ok := ret.Get(0).(func(context.Context, common.Hash) *types.Transaction); ok {
		r0 = rf(ctx, hash)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.Transaction)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, common.Hash) bool); ok {
		r1 = rf(ctx, hash)
	} else {
		r1 = ret.Get(1).(bool)
	}

	if rf, ok := ret.Get(2).(func(context.Context, common.Hash) error); ok {
		r2 = rf(ctx, hash)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// MockEthclientClient_TransactionByHash_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'TransactionByHash'
type MockEthclientClient_TransactionByHash_Call struct {
	*mock.Call
}

// TransactionByHash is a helper method to define mock.On call
//   - ctx context.Context
//   - hash common.Hash
func (_e *MockEthclientClient_Expecter) TransactionByHash(ctx interface{}, hash interface{}) *MockEthclientClient_TransactionByHash_Call {
	return &MockEthclientClient_TransactionByHash_Call{Call: _e.mock.On("TransactionByHash", ctx, hash)}
}

func (_c *MockEthclientClient_TransactionByHash_Call) Run(run func(ctx context.Context, hash common.Hash)) *MockEthclientClient_TransactionByHash_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(common.Hash))
	})
	return _c
}

func (_c *MockEthclientClient_TransactionByHash_Call) Return(_a0 *types.Transaction, _a1 bool, _a2 error) *MockEthclientClient_TransactionByHash_Call {
	_c.Call.Return(_a0, _a1, _a2)
	return _c
}

func (_c *MockEthclientClient_TransactionByHash_Call) RunAndReturn(run func(context.Context, common.Hash) (*types.Transaction, bool, error)) *MockEthclientClient_TransactionByHash_Call {
	_c.Call.Return(run)
	return _c
}

// TransactionReceipt provides a mock function with given fields: ctx, txHash
func (_m *MockEthclientClient) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	ret := _m.Called(ctx, txHash)

	if len(ret) == 0 {
		panic("no return value specified for TransactionReceipt")
	}

	var r0 *types.Receipt
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, common.Hash) (*types.Receipt, error)); ok {
		return rf(ctx, txHash)
	}
	if rf, ok := ret.Get(0).(func(context.Context, common.Hash) *types.Receipt); ok {
		r0 = rf(ctx, txHash)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.Receipt)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, common.Hash) error); ok {
		r1 = rf(ctx, txHash)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockEthclientClient_TransactionReceipt_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'TransactionReceipt'
type MockEthclientClient_TransactionReceipt_Call struct {
	*mock.Call
}

// TransactionReceipt is a helper method to define mock.On call
//   - ctx context.Context
//   - txHash common.Hash
func (_e *MockEthclientClient_Expecter) TransactionReceipt(ctx interface{}, txHash interface{}) *MockEthclientClient_TransactionReceipt_Call {
	return &MockEthclientClient_TransactionReceipt_Call{Call: _e.mock.On("TransactionReceipt", ctx, txHash)}
}

func (_c *MockEthclientClient_TransactionReceipt_Call) Run(run func(ctx context.Context, txHash common.Hash)) *MockEthclientClient_TransactionReceipt_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(common.Hash))
	})
	return _c
}

func (_c *MockEthclientClient_TransactionReceipt_Call) Return(_a0 *types.Receipt, _a1 error) *MockEthclientClient_TransactionReceipt_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockEthclientClient_TransactionReceipt_Call) RunAndReturn(run func(context.Context, common.Hash) (*types.Receipt, error)) *MockEthclientClient_TransactionReceipt_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockEthclientClient creates a new instance of MockEthclientClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockEthclientClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockEthclientClient {
	mock := &MockEthclientClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
