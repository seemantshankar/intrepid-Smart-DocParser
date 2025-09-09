package external

import (
	"context"
	"testing"
	"time"

	"github.com/sony/gobreaker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHTTPClient_ExecuteRequest(t *testing.T) {
	// Test cases
	t.Run("successful request", func(t *testing.T) {
		// fresh mock and client per subtest
		mockCB := new(MockCircuitBreaker)
		client := &HTTPClient{
			CB: mockCB,
			RetryCfg: RetryConfig{
				MaxRetries:      3,
				InitialInterval: 10 * time.Millisecond,
				MaxInterval:     50 * time.Millisecond,
			},
		}
		mockCB.On("Execute", mock.Anything).Return(&Response{StatusCode: 200}, nil)

		resp, err := client.ExecuteRequest(context.Background(), &Request{})

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		mockCB.AssertExpectations(t)
	})

	t.Run("circuit breaker open", func(t *testing.T) {
		mockCB := new(MockCircuitBreaker)
		client := &HTTPClient{
			CB: mockCB,
			RetryCfg: RetryConfig{
				MaxRetries:      3,
				InitialInterval: 10 * time.Millisecond,
				MaxInterval:     50 * time.Millisecond,
			},
		}
		mockCB.On("Execute", mock.Anything).Return(nil, gobreaker.ErrOpenState)

		_, err := client.ExecuteRequest(context.Background(), &Request{})

		assert.ErrorIs(t, err, gobreaker.ErrOpenState)
		mockCB.AssertExpectations(t)
	})
}

// MockCircuitBreaker implements a mock circuit breaker
type MockCircuitBreaker struct {
	mock.Mock
}

func (m *MockCircuitBreaker) Execute(req func() (interface{}, error)) (interface{}, error) {
	args := m.Called(req)
	return args.Get(0), args.Error(1)
}
