// Code generated by mockery v2.43.2. DO NOT EDIT.

package mocks

import (
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	mock "github.com/stretchr/testify/mock"

	proto "github.com/cosmos/gogoproto/proto"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"

	txsigning "github.com/cosmos/cosmos-sdk/types/tx/signing"

	types "github.com/cosmos/cosmos-sdk/types"
)

// MockTx is an autogenerated mock type for the Tx type
type MockTx struct {
	mock.Mock
}

type MockTx_Expecter struct {
	mock *mock.Mock
}

func (_m *MockTx) EXPECT() *MockTx_Expecter {
	return &MockTx_Expecter{mock: &_m.Mock}
}

// FeeGranter provides a mock function with given fields:
func (_m *MockTx) FeeGranter() []byte {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for FeeGranter")
	}

	var r0 []byte
	if rf, ok := ret.Get(0).(func() []byte); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	return r0
}

// MockTx_FeeGranter_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FeeGranter'
type MockTx_FeeGranter_Call struct {
	*mock.Call
}

// FeeGranter is a helper method to define mock.On call
func (_e *MockTx_Expecter) FeeGranter() *MockTx_FeeGranter_Call {
	return &MockTx_FeeGranter_Call{Call: _e.mock.On("FeeGranter")}
}

func (_c *MockTx_FeeGranter_Call) Run(run func()) *MockTx_FeeGranter_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockTx_FeeGranter_Call) Return(_a0 []byte) *MockTx_FeeGranter_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockTx_FeeGranter_Call) RunAndReturn(run func() []byte) *MockTx_FeeGranter_Call {
	_c.Call.Return(run)
	return _c
}

// FeePayer provides a mock function with given fields:
func (_m *MockTx) FeePayer() []byte {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for FeePayer")
	}

	var r0 []byte
	if rf, ok := ret.Get(0).(func() []byte); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	return r0
}

// MockTx_FeePayer_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FeePayer'
type MockTx_FeePayer_Call struct {
	*mock.Call
}

// FeePayer is a helper method to define mock.On call
func (_e *MockTx_Expecter) FeePayer() *MockTx_FeePayer_Call {
	return &MockTx_FeePayer_Call{Call: _e.mock.On("FeePayer")}
}

func (_c *MockTx_FeePayer_Call) Run(run func()) *MockTx_FeePayer_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockTx_FeePayer_Call) Return(_a0 []byte) *MockTx_FeePayer_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockTx_FeePayer_Call) RunAndReturn(run func() []byte) *MockTx_FeePayer_Call {
	_c.Call.Return(run)
	return _c
}

// GetFee provides a mock function with given fields:
func (_m *MockTx) GetFee() types.Coins {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetFee")
	}

	var r0 types.Coins
	if rf, ok := ret.Get(0).(func() types.Coins); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(types.Coins)
		}
	}

	return r0
}

// MockTx_GetFee_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetFee'
type MockTx_GetFee_Call struct {
	*mock.Call
}

// GetFee is a helper method to define mock.On call
func (_e *MockTx_Expecter) GetFee() *MockTx_GetFee_Call {
	return &MockTx_GetFee_Call{Call: _e.mock.On("GetFee")}
}

func (_c *MockTx_GetFee_Call) Run(run func()) *MockTx_GetFee_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockTx_GetFee_Call) Return(_a0 types.Coins) *MockTx_GetFee_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockTx_GetFee_Call) RunAndReturn(run func() types.Coins) *MockTx_GetFee_Call {
	_c.Call.Return(run)
	return _c
}

// GetGas provides a mock function with given fields:
func (_m *MockTx) GetGas() uint64 {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetGas")
	}

	var r0 uint64
	if rf, ok := ret.Get(0).(func() uint64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(uint64)
	}

	return r0
}

// MockTx_GetGas_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetGas'
type MockTx_GetGas_Call struct {
	*mock.Call
}

// GetGas is a helper method to define mock.On call
func (_e *MockTx_Expecter) GetGas() *MockTx_GetGas_Call {
	return &MockTx_GetGas_Call{Call: _e.mock.On("GetGas")}
}

func (_c *MockTx_GetGas_Call) Run(run func()) *MockTx_GetGas_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockTx_GetGas_Call) Return(_a0 uint64) *MockTx_GetGas_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockTx_GetGas_Call) RunAndReturn(run func() uint64) *MockTx_GetGas_Call {
	_c.Call.Return(run)
	return _c
}

// GetMemo provides a mock function with given fields:
func (_m *MockTx) GetMemo() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetMemo")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockTx_GetMemo_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetMemo'
type MockTx_GetMemo_Call struct {
	*mock.Call
}

// GetMemo is a helper method to define mock.On call
func (_e *MockTx_Expecter) GetMemo() *MockTx_GetMemo_Call {
	return &MockTx_GetMemo_Call{Call: _e.mock.On("GetMemo")}
}

func (_c *MockTx_GetMemo_Call) Run(run func()) *MockTx_GetMemo_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockTx_GetMemo_Call) Return(_a0 string) *MockTx_GetMemo_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockTx_GetMemo_Call) RunAndReturn(run func() string) *MockTx_GetMemo_Call {
	_c.Call.Return(run)
	return _c
}

// GetMsgs provides a mock function with given fields:
func (_m *MockTx) GetMsgs() []proto.Message {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetMsgs")
	}

	var r0 []proto.Message
	if rf, ok := ret.Get(0).(func() []proto.Message); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]proto.Message)
		}
	}

	return r0
}

// MockTx_GetMsgs_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetMsgs'
type MockTx_GetMsgs_Call struct {
	*mock.Call
}

// GetMsgs is a helper method to define mock.On call
func (_e *MockTx_Expecter) GetMsgs() *MockTx_GetMsgs_Call {
	return &MockTx_GetMsgs_Call{Call: _e.mock.On("GetMsgs")}
}

func (_c *MockTx_GetMsgs_Call) Run(run func()) *MockTx_GetMsgs_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockTx_GetMsgs_Call) Return(_a0 []proto.Message) *MockTx_GetMsgs_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockTx_GetMsgs_Call) RunAndReturn(run func() []proto.Message) *MockTx_GetMsgs_Call {
	_c.Call.Return(run)
	return _c
}

// GetMsgsV2 provides a mock function with given fields:
func (_m *MockTx) GetMsgsV2() ([]protoreflect.ProtoMessage, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetMsgsV2")
	}

	var r0 []protoreflect.ProtoMessage
	var r1 error
	if rf, ok := ret.Get(0).(func() ([]protoreflect.ProtoMessage, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() []protoreflect.ProtoMessage); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]protoreflect.ProtoMessage)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockTx_GetMsgsV2_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetMsgsV2'
type MockTx_GetMsgsV2_Call struct {
	*mock.Call
}

// GetMsgsV2 is a helper method to define mock.On call
func (_e *MockTx_Expecter) GetMsgsV2() *MockTx_GetMsgsV2_Call {
	return &MockTx_GetMsgsV2_Call{Call: _e.mock.On("GetMsgsV2")}
}

func (_c *MockTx_GetMsgsV2_Call) Run(run func()) *MockTx_GetMsgsV2_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockTx_GetMsgsV2_Call) Return(_a0 []protoreflect.ProtoMessage, _a1 error) *MockTx_GetMsgsV2_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockTx_GetMsgsV2_Call) RunAndReturn(run func() ([]protoreflect.ProtoMessage, error)) *MockTx_GetMsgsV2_Call {
	_c.Call.Return(run)
	return _c
}

// GetPubKeys provides a mock function with given fields:
func (_m *MockTx) GetPubKeys() ([]cryptotypes.PubKey, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetPubKeys")
	}

	var r0 []cryptotypes.PubKey
	var r1 error
	if rf, ok := ret.Get(0).(func() ([]cryptotypes.PubKey, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() []cryptotypes.PubKey); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]cryptotypes.PubKey)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockTx_GetPubKeys_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetPubKeys'
type MockTx_GetPubKeys_Call struct {
	*mock.Call
}

// GetPubKeys is a helper method to define mock.On call
func (_e *MockTx_Expecter) GetPubKeys() *MockTx_GetPubKeys_Call {
	return &MockTx_GetPubKeys_Call{Call: _e.mock.On("GetPubKeys")}
}

func (_c *MockTx_GetPubKeys_Call) Run(run func()) *MockTx_GetPubKeys_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockTx_GetPubKeys_Call) Return(_a0 []cryptotypes.PubKey, _a1 error) *MockTx_GetPubKeys_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockTx_GetPubKeys_Call) RunAndReturn(run func() ([]cryptotypes.PubKey, error)) *MockTx_GetPubKeys_Call {
	_c.Call.Return(run)
	return _c
}

// GetSignaturesV2 provides a mock function with given fields:
func (_m *MockTx) GetSignaturesV2() ([]txsigning.SignatureV2, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetSignaturesV2")
	}

	var r0 []txsigning.SignatureV2
	var r1 error
	if rf, ok := ret.Get(0).(func() ([]txsigning.SignatureV2, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() []txsigning.SignatureV2); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]txsigning.SignatureV2)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockTx_GetSignaturesV2_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetSignaturesV2'
type MockTx_GetSignaturesV2_Call struct {
	*mock.Call
}

// GetSignaturesV2 is a helper method to define mock.On call
func (_e *MockTx_Expecter) GetSignaturesV2() *MockTx_GetSignaturesV2_Call {
	return &MockTx_GetSignaturesV2_Call{Call: _e.mock.On("GetSignaturesV2")}
}

func (_c *MockTx_GetSignaturesV2_Call) Run(run func()) *MockTx_GetSignaturesV2_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockTx_GetSignaturesV2_Call) Return(_a0 []txsigning.SignatureV2, _a1 error) *MockTx_GetSignaturesV2_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockTx_GetSignaturesV2_Call) RunAndReturn(run func() ([]txsigning.SignatureV2, error)) *MockTx_GetSignaturesV2_Call {
	_c.Call.Return(run)
	return _c
}

// GetSigners provides a mock function with given fields:
func (_m *MockTx) GetSigners() ([][]byte, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetSigners")
	}

	var r0 [][]byte
	var r1 error
	if rf, ok := ret.Get(0).(func() ([][]byte, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() [][]byte); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([][]byte)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockTx_GetSigners_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetSigners'
type MockTx_GetSigners_Call struct {
	*mock.Call
}

// GetSigners is a helper method to define mock.On call
func (_e *MockTx_Expecter) GetSigners() *MockTx_GetSigners_Call {
	return &MockTx_GetSigners_Call{Call: _e.mock.On("GetSigners")}
}

func (_c *MockTx_GetSigners_Call) Run(run func()) *MockTx_GetSigners_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockTx_GetSigners_Call) Return(_a0 [][]byte, _a1 error) *MockTx_GetSigners_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockTx_GetSigners_Call) RunAndReturn(run func() ([][]byte, error)) *MockTx_GetSigners_Call {
	_c.Call.Return(run)
	return _c
}

// GetTimeoutHeight provides a mock function with given fields:
func (_m *MockTx) GetTimeoutHeight() uint64 {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetTimeoutHeight")
	}

	var r0 uint64
	if rf, ok := ret.Get(0).(func() uint64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(uint64)
	}

	return r0
}

// MockTx_GetTimeoutHeight_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetTimeoutHeight'
type MockTx_GetTimeoutHeight_Call struct {
	*mock.Call
}

// GetTimeoutHeight is a helper method to define mock.On call
func (_e *MockTx_Expecter) GetTimeoutHeight() *MockTx_GetTimeoutHeight_Call {
	return &MockTx_GetTimeoutHeight_Call{Call: _e.mock.On("GetTimeoutHeight")}
}

func (_c *MockTx_GetTimeoutHeight_Call) Run(run func()) *MockTx_GetTimeoutHeight_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockTx_GetTimeoutHeight_Call) Return(_a0 uint64) *MockTx_GetTimeoutHeight_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockTx_GetTimeoutHeight_Call) RunAndReturn(run func() uint64) *MockTx_GetTimeoutHeight_Call {
	_c.Call.Return(run)
	return _c
}

// ValidateBasic provides a mock function with given fields:
func (_m *MockTx) ValidateBasic() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for ValidateBasic")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockTx_ValidateBasic_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ValidateBasic'
type MockTx_ValidateBasic_Call struct {
	*mock.Call
}

// ValidateBasic is a helper method to define mock.On call
func (_e *MockTx_Expecter) ValidateBasic() *MockTx_ValidateBasic_Call {
	return &MockTx_ValidateBasic_Call{Call: _e.mock.On("ValidateBasic")}
}

func (_c *MockTx_ValidateBasic_Call) Run(run func()) *MockTx_ValidateBasic_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockTx_ValidateBasic_Call) Return(_a0 error) *MockTx_ValidateBasic_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockTx_ValidateBasic_Call) RunAndReturn(run func() error) *MockTx_ValidateBasic_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockTx creates a new instance of MockTx. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockTx(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockTx {
	mock := &MockTx{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
