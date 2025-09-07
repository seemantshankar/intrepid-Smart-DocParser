package external

import "time"

// ClientConfig holds configuration for service clients
type ClientConfig struct {
	Name        string
	BaseURL     string
	Timeout     time.Duration
	Credentials Credentials
	Retry       RetryConfig
}

// Credentials represents authentication credentials
type Credentials struct {
	APIKey     string
	SecretKey  string
	Token      string
}

// DefaultRetryConfig returns a sensible default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:      3,
		InitialInterval: 500 * time.Millisecond,
		MaxInterval:     5 * time.Second,
	}
}
