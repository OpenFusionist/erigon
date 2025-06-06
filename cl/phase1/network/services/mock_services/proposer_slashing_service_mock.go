// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/erigontech/erigon/cl/phase1/network/services (interfaces: ProposerSlashingService)
//
// Generated by this command:
//
//	mockgen -typed=true -destination=./mock_services/proposer_slashing_service_mock.go -package=mock_services . ProposerSlashingService
//

// Package mock_services is a generated GoMock package.
package mock_services

import (
	context "context"
	reflect "reflect"

	cltypes "github.com/erigontech/erigon/cl/cltypes"
	gomock "go.uber.org/mock/gomock"
)

// MockProposerSlashingService is a mock of ProposerSlashingService interface.
type MockProposerSlashingService struct {
	ctrl     *gomock.Controller
	recorder *MockProposerSlashingServiceMockRecorder
	isgomock struct{}
}

// MockProposerSlashingServiceMockRecorder is the mock recorder for MockProposerSlashingService.
type MockProposerSlashingServiceMockRecorder struct {
	mock *MockProposerSlashingService
}

// NewMockProposerSlashingService creates a new mock instance.
func NewMockProposerSlashingService(ctrl *gomock.Controller) *MockProposerSlashingService {
	mock := &MockProposerSlashingService{ctrl: ctrl}
	mock.recorder = &MockProposerSlashingServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockProposerSlashingService) EXPECT() *MockProposerSlashingServiceMockRecorder {
	return m.recorder
}

// ProcessMessage mocks base method.
func (m *MockProposerSlashingService) ProcessMessage(ctx context.Context, subnet *uint64, msg *cltypes.ProposerSlashing) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProcessMessage", ctx, subnet, msg)
	ret0, _ := ret[0].(error)
	return ret0
}

// ProcessMessage indicates an expected call of ProcessMessage.
func (mr *MockProposerSlashingServiceMockRecorder) ProcessMessage(ctx, subnet, msg any) *MockProposerSlashingServiceProcessMessageCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProcessMessage", reflect.TypeOf((*MockProposerSlashingService)(nil).ProcessMessage), ctx, subnet, msg)
	return &MockProposerSlashingServiceProcessMessageCall{Call: call}
}

// MockProposerSlashingServiceProcessMessageCall wrap *gomock.Call
type MockProposerSlashingServiceProcessMessageCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockProposerSlashingServiceProcessMessageCall) Return(arg0 error) *MockProposerSlashingServiceProcessMessageCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockProposerSlashingServiceProcessMessageCall) Do(f func(context.Context, *uint64, *cltypes.ProposerSlashing) error) *MockProposerSlashingServiceProcessMessageCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockProposerSlashingServiceProcessMessageCall) DoAndReturn(f func(context.Context, *uint64, *cltypes.ProposerSlashing) error) *MockProposerSlashingServiceProcessMessageCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}
