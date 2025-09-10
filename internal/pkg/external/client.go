package external

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/sony/gobreaker"
)

// Client defines the interface for external service clients
// with built-in resilience patterns
//
//go:generate mockery --name=Client --output=./mocks --filename=client_mock.go
type Client interface {
	ExecuteRequest(ctx context.Context, req *Request) (*Response, error)
}

// CircuitBreaker defines the interface for circuit breaker
type CircuitBreaker interface {
	Execute(func() (interface{}, error)) (interface{}, error)
}

// Request represents a service request
type Request struct {
	Method      string
	URL         string
	Headers     map[string]string
	Body        []byte
	IsStreaming bool
}

// Response represents a service response
type Response struct {
	StatusCode int
	Body       []byte
}

// HTTPClient implements Client with resilience patterns
type HTTPClient struct {
	BaseURL  string
	CB       CircuitBreaker
	RetryCfg RetryConfig
	Timeout  time.Duration
}

// RetryConfig defines retry behavior
type RetryConfig struct {
	MaxRetries      int
	InitialInterval time.Duration
	MaxInterval     time.Duration
}

// NewHTTPClient creates a new resilient HTTP client
func NewHTTPClient(baseURL, name string, retryCfg RetryConfig, timeout time.Duration) Client {
	return &HTTPClient{
		BaseURL: baseURL,
		CB: gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:        name,
			MaxRequests: 5,
			Interval:    30 * time.Second,
			Timeout:     10 * time.Second,
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				return counts.ConsecutiveFailures > 3
			},
		}),
		RetryCfg: retryCfg,
		Timeout:  timeout,
	}
}

// ExecuteRequest executes a request with circuit breaker and retry logic
func (c *HTTPClient) ExecuteRequest(ctx context.Context, req *Request) (*Response, error) {
	var resp *Response
	var err error

	// Wrap with circuit breaker
	result, cbErr := c.CB.Execute(func() (interface{}, error) {
		// Wrap with retry logic
		retryErr := backoff.Retry(func() error {
			resp, err = c.doActualRequest(ctx, req)
			if err != nil {
				// Only retry on transient errors
				if isTransientError(err) {
					return err
				}
				return backoff.Permanent(err)
			}
			return nil
		}, c.newExponentialBackoff())

		if retryErr != nil {
			return nil, retryErr
		}
		return resp, nil
	})

	if cbErr != nil {
		return nil, cbErr
	}
	// If the circuit breaker returned a value directly (e.g., in tests), prefer it
	if result != nil {
		if r, ok := result.(*Response); ok && r != nil {
			return r, nil
		}
	}
	return resp, nil
}

// doActualRequest performs the actual HTTP request
func (c *HTTPClient) doActualRequest(ctx context.Context, req *Request) (*Response, error) {
	// Build the full URL robustly:
	// - If req.URL is absolute (starts with http:// or https://), use it as-is
	// - Otherwise, join with BaseURL while handling slashes
	fullURL := req.URL
	if !strings.HasPrefix(req.URL, "http://") && !strings.HasPrefix(req.URL, "https://") {
		if c.BaseURL != "" {
			switch {
			case strings.HasSuffix(c.BaseURL, "/") && strings.HasPrefix(req.URL, "/"):
				fullURL = c.BaseURL + req.URL[1:]
			case !strings.HasSuffix(c.BaseURL, "/") && !strings.HasPrefix(req.URL, "/"):
				fullURL = c.BaseURL + "/" + req.URL
			default:
				fullURL = c.BaseURL + req.URL
			}
		}
	}
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, fullURL, bytes.NewBuffer(req.Body))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}

	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	client := &http.Client{Timeout: c.Timeout}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var body []byte
	if req.IsStreaming {
		// Handle streaming response
		scanner := bufio.NewScanner(resp.Body)
		var accumulatedContent strings.Builder
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "data: ") {
				dataStr := line[6:]
				if dataStr == "[DONE]" {
					break
				}
				var streamResp struct {
					Choices []struct {
						Delta struct {
							Content string `json:"content"`
						} `json:"delta"`
					} `json:"choices"`
				}
				if err := json.Unmarshal([]byte(dataStr), &streamResp); err != nil {
					continue
				}
				if len(streamResp.Choices) > 0 {
					accumulatedContent.WriteString(streamResp.Choices[0].Delta.Content)
				}
			}
		}
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("error reading streaming response: %w", err)
		}
		body = []byte(accumulatedContent.String())
	} else {
		// Handle regular response
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Body:       body,
	}, nil
}

// isTransientError determines if an error is transient
func isTransientError(err error) bool {
	// Treat common network/timeouts as transient
	var nerr net.Error
	if errors.As(err, &nerr) {
		// Timeout is generally transient
		if nerr.Timeout() {
			return true
		}
		// Some net errors may be temporary (historical API)
		type temporary interface{ Temporary() bool }
		if te, ok := any(nerr).(temporary); ok && te.Temporary() {
			return true
		}
	}

	// url.Error also exposes Timeout()
	var uerr *url.Error
	if errors.As(err, &uerr) {
		if uerr.Timeout() {
			return true
		}
	}

	// Context deadline exceeded
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	return false
}

func (c *HTTPClient) newExponentialBackoff() *backoff.ExponentialBackOff {
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = c.RetryCfg.InitialInterval
	b.MaxInterval = c.RetryCfg.MaxInterval
	b.MaxElapsedTime = time.Duration(c.RetryCfg.MaxRetries) * c.RetryCfg.MaxInterval
	return b
}
