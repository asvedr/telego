// Package mock contains mocks of telegoapi interfaces
package mock

import (
	"context"

	"gitlab.com/asvedr/testify/mock"

	"github.com/mymmrac/telego/telegoapi"
)

// MockCaller is a mock of [telegoapi.Caller] interface
type MockCaller struct {
	mock.Mock
}

// NewMockCaller creates a new mock instance
func NewMockCaller() *MockCaller {
	return &MockCaller{}
}

// Call mocks [telegoapi.Caller.Call] method
func (m *MockCaller) Call(ctx context.Context, url string, data *telegoapi.RequestData) (*telegoapi.Response, error) {
	return mock.CastErr[*telegoapi.Response](m.Called(ctx, url, data))
}
