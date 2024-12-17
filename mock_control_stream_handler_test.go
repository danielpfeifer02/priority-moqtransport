// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/danielpfeifer02/priority-moqtransport (interfaces: ControlStreamHandler)
//
// Generated by this command:
//
//	mockgen -build_flags=-tags=gomock -package moqtransport -self_package github.com/danielpfeifer02/priority-moqtransport -destination mock_control_stream_handler_test.go github.com/danielpfeifer02/priority-moqtransport ControlStreamHandler
//
// Package moqtransport is a generated GoMock package.
package moqtransport

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockControlStreamHandler is a mock of ControlStreamHandler interface.
type MockControlStreamHandler struct {
	ctrl     *gomock.Controller
	recorder *MockControlStreamHandlerMockRecorder
}

// MockControlStreamHandlerMockRecorder is the mock recorder for MockControlStreamHandler.
type MockControlStreamHandlerMockRecorder struct {
	mock *MockControlStreamHandler
}

// NewMockControlStreamHandler creates a new mock instance.
func NewMockControlStreamHandler(ctrl *gomock.Controller) *MockControlStreamHandler {
	mock := &MockControlStreamHandler{ctrl: ctrl}
	mock.recorder = &MockControlStreamHandlerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockControlStreamHandler) EXPECT() *MockControlStreamHandlerMockRecorder {
	return m.recorder
}

// Read mocks base method.
func (m *MockControlStreamHandler) Read(arg0 []byte) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Read", arg0)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Read indicates an expected call of Read.
func (mr *MockControlStreamHandlerMockRecorder) Read(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Read", reflect.TypeOf((*MockControlStreamHandler)(nil).Read), arg0)
}

// readMessages mocks base method.
func (m *MockControlStreamHandler) readMessages(arg0 messageHandler) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "readMessages", arg0)
}

// readMessages indicates an expected call of readMessages.
func (mr *MockControlStreamHandlerMockRecorder) readMessages(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "readMessages", reflect.TypeOf((*MockControlStreamHandler)(nil).readMessages), arg0)
}

// send mocks base method.
func (m *MockControlStreamHandler) send(arg0 message) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "send", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// send indicates an expected call of send.
func (mr *MockControlStreamHandlerMockRecorder) send(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "send", reflect.TypeOf((*MockControlStreamHandler)(nil).send), arg0)
}