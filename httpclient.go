package httpclient

import (
	"fmt"
	"math"
	"net"
	"net/http"
	"net/url"
	"time"
)

const (
	defaultKeepAlive   = 15 * time.Second
	defaultTLSTimeout  = 5 * time.Second
	defaultTimeout     = 15 * time.Second
	defaultDialTimeout = 5 * time.Second
	maxRetries         = 5
	baseRetryDelay     = 2 * time.Second
	maxRetryDelay      = 30 * time.Second
)

type Option func(*http.Client)

// WithTimeout sets client timeout
func WithTimeout(t time.Duration) Option {
	return func(c *http.Client) {
		c.Timeout = t
	}
}

// WithTransport sets a custom transport
func WithTransport(t *http.Transport) Option {
	return func(c *http.Client) {
		c.Transport = t
	}
}

// WithTLSHandshakeTimeout sets TLS handshake timeout
func WithTLSHandshakeTimeout(t time.Duration) Option {
	return func(c *http.Client) {
		if transport, ok := c.Transport.(*http.Transport); ok {
			transport.TLSHandshakeTimeout = t
		}
	}
}

// WithResponseHeaderTimeout sets response header timeout
func WithResponseHeaderTimeout(t time.Duration) Option {
	return func(c *http.Client) {
		if transport, ok := c.Transport.(*http.Transport); ok {
			transport.ResponseHeaderTimeout = t
		}
	}
}

// WithIdleConnTimeout sets idle connection timeout
func WithIdleConnTimeout(t time.Duration) Option {
	return func(c *http.Client) {
		if transport, ok := c.Transport.(*http.Transport); ok {
			transport.IdleConnTimeout = t
		}
	}
}

// WithMaxIdleConns sets maximum idle connections
func WithMaxIdleConns(n int) Option {
	return func(c *http.Client) {
		if transport, ok := c.Transport.(*http.Transport); ok {
			transport.MaxIdleConns = n
		}
	}
}

// WithMaxIdleConnsPerHost sets maximum idle connections per host
func WithMaxIdleConnsPerHost(n int) Option {
	return func(c *http.Client) {
		if transport, ok := c.Transport.(*http.Transport); ok {
			transport.MaxIdleConnsPerHost = n
		}
	}
}

// WithForceHTTP2Disabled disables HTTP/2
func WithForceHTTP2Disabled() Option {
	return func(c *http.Client) {
		if transport, ok := c.Transport.(*http.Transport); ok {
			transport.ForceAttemptHTTP2 = false
		}
	}
}

// WithExpectContinueTimeout sets expect continue timeout
func WithExpectContinueTimeout(t time.Duration) Option {
	return func(c *http.Client) {
		if transport, ok := c.Transport.(*http.Transport); ok {
			transport.ExpectContinueTimeout = t
		}
	}
}

// WithProxy sets the proxy function
func WithProxy(proxy func(*http.Request) (*url.URL, error)) Option {
	return func(c *http.Client) {
		if transport, ok := c.Transport.(*http.Transport); ok {
			transport.Proxy = proxy
		}
	}
}

// WithDialerTimeout sets dialer timeout
func WithDialerTimeout(t time.Duration) Option {
	return func(c *http.Client) {
		if transport, ok := c.Transport.(*http.Transport); ok {
			transport.DialContext = (&net.Dialer{
				Timeout:   t,
				KeepAlive: defaultKeepAlive,
			}).DialContext
		}
	}
}

// WithDialerKeepAlive sets dialer keep alive duration
func WithDialerKeepAlive(t time.Duration) Option {
	return func(c *http.Client) {
		if transport, ok := c.Transport.(*http.Transport); ok {
			transport.DialContext = (&net.Dialer{
				Timeout:   defaultDialTimeout,
				KeepAlive: t,
			}).DialContext
		}
	}
}

// New creates a new HTTP client with default transport settings and applies options
func New(opts ...Option) *http.Client {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   defaultDialTimeout,
			KeepAlive: defaultKeepAlive,
		}).DialContext,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   defaultTLSTimeout,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: defaultTimeout,
	}

	client := &http.Client{
		Timeout:   defaultTimeout,
		Transport: transport,
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

// DoWithRetry performs an HTTP request with retries
func DoWithRetry(client *http.Client, req *http.Request) (*http.Response, error) {
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		resp, err := client.Do(req)
		if err == nil {
			return resp, nil
		}

		lastErr = err
		delay := time.Duration(float64(baseRetryDelay) * math.Pow(2, float64(attempt)))
		if delay > maxRetryDelay {
			delay = maxRetryDelay
		}

		select {
		case <-req.Context().Done():
			return nil, fmt.Errorf("request cancelled or timed out: %w", req.Context().Err())
		case <-time.After(delay):
			continue
		}
	}
	return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

func isTemporaryError(err error) bool {
	if netErr, ok := err.(net.Error); ok {
		return netErr.Timeout()
	}
	return true
}
