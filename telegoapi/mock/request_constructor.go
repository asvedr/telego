package mock

import (
	"gitlab.com/asvedr/testify/mock"

	"github.com/mymmrac/telego/telegoapi"
)

// MockRequestConstructor is a mock of [telegoapi.RequestConstructor] interface
type MockRequestConstructor struct {
	mock.Mock
}

// NewMockRequestConstructor creates a new mock instance
func NewMockRequestConstructor() *MockRequestConstructor {
	return &MockRequestConstructor{}
}

// JSONRequest mocks [telegoapi.RequestConstructor.JSONRequest] method
func (m *MockRequestConstructor) JSONRequest(parameters any) (*telegoapi.RequestData, error) {
	return mock.CastErr[*telegoapi.RequestData](m.Called(parameters))
}

// MultipartRequest mocks [telegoapi.RequestConstructor.MultipartRequest] method
func (m *MockRequestConstructor) MultipartRequest(
	parameters map[string]string,
	filesParameters map[string]telegoapi.NamedReader,
) (*telegoapi.RequestData, error) {
	return mock.CastErr[*telegoapi.RequestData](m.Called(parameters, filesParameters))
}
