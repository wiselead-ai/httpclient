package httpclient

import (
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"

	"net/http/httptest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {
	t.Parallel()

	t.Run("Default client", func(t *testing.T) {
		t.Parallel()

		httpClient := New()
		transport := httpClient.Transport.(*http.Transport)
		require.NotNil(t, httpClient)

		assert.Equal(t, 15*time.Second, httpClient.Timeout)
		assert.Equal(t, 90*time.Second, transport.IdleConnTimeout)
	})

	t.Run("WithTimeout", func(t *testing.T) {
		t.Parallel()

		httpClient := New(WithTimeout(1 * time.Second))
		assert.Equal(t, 1*time.Second, httpClient.Timeout)
	})

	t.Run("WithTLSHandshakeTimeout", func(t *testing.T) {
		t.Parallel()

		httpClient := New(WithTLSHandshakeTimeout(1 * time.Second))
		transport := httpClient.Transport.(*http.Transport)
		assert.Equal(t, 1*time.Second, transport.TLSHandshakeTimeout)
	})

	t.Run("WithResponseHeaderTimeout", func(t *testing.T) {
		t.Parallel()

		httpClient := New(WithResponseHeaderTimeout(1 * time.Second))
		transport := httpClient.Transport.(*http.Transport)
		assert.Equal(t, 1*time.Second, transport.ResponseHeaderTimeout)
	})

	t.Run("WithIdleConnTimeout", func(t *testing.T) {
		t.Parallel()

		httpClient := New(WithIdleConnTimeout(1 * time.Second))
		transport := httpClient.Transport.(*http.Transport)
		assert.Equal(t, 1*time.Second, transport.IdleConnTimeout)
	})

	t.Run("WithMaxIdleConns", func(t *testing.T) {
		t.Parallel()

		httpClient := New(WithMaxIdleConns(100))
		transport := httpClient.Transport.(*http.Transport)
		assert.Equal(t, 100, transport.MaxIdleConns)
	})

	t.Run("WithMaxIdleConnsPerHost", func(t *testing.T) {
		t.Parallel()

		httpClient := New(WithMaxIdleConnsPerHost(100))
		transport := httpClient.Transport.(*http.Transport)
		assert.Equal(t, 100, transport.MaxIdleConnsPerHost)
	})

	t.Run("WithForceHTTP2Disabled", func(t *testing.T) {
		t.Parallel()

		httpClient := New(WithForceHTTP2Disabled())
		transport := httpClient.Transport.(*http.Transport)
		assert.Equal(t, false, transport.ForceAttemptHTTP2)
	})

	t.Run("WithCustomTransport", func(t *testing.T) {
		t.Parallel()

		customTransport := http.Transport{
			MaxIdleConns:    200,
			IdleConnTimeout: 30 * time.Second,
		}

		httpClient := New(WithTransport(&customTransport))
		assert.Equal(t, &customTransport, httpClient.Transport)
	})

	t.Run("WithExpectContinueTimeout", func(t *testing.T) {
		t.Parallel()

		httpClient := New(WithExpectContinueTimeout(2 * time.Second))
		transport := httpClient.Transport.(*http.Transport)
		assert.Equal(t, 2*time.Second, transport.ExpectContinueTimeout)
	})

	t.Run("WithProxy", func(t *testing.T) {
		t.Parallel()

		proxyFunc := func(req *http.Request) (*url.URL, error) {
			return &url.URL{Host: "localhost:8080"}, nil
		}

		httpClient := New(WithProxy(proxyFunc))
		transport := httpClient.Transport.(*http.Transport)

		proxiedURL, err := transport.Proxy(&http.Request{})
		require.NoError(t, err)
		assert.Equal(t, "localhost:8080", proxiedURL.Host)
	})

	t.Run("WithDialerTimeout", func(t *testing.T) {
		t.Parallel()

		expectedTimeout := 1 * time.Second
		httpClient := New(WithDialerTimeout(expectedTimeout))
		_ = httpClient.Transport.(*http.Transport)

		expectedDialer := &net.Dialer{
			Timeout: expectedTimeout,
		}
		assert.Equal(t, expectedDialer.Timeout, expectedTimeout)
	})

	t.Run("WithDialerKeepAlive", func(t *testing.T) {
		t.Parallel()

		expectedKeepAlive := 1 * time.Second
		httpClient := New(WithDialerKeepAlive(expectedKeepAlive))
		_ = httpClient.Transport.(*http.Transport) // type assertion to ensure transport is valid

		expectedDialer := &net.Dialer{
			KeepAlive: expectedKeepAlive,
		}
		assert.Equal(t, expectedDialer.KeepAlive, expectedKeepAlive)
	})

	t.Run("DoWithRetry success", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := New()
		req, err := http.NewRequest("GET", server.URL, nil)
		require.NoError(t, err)

		resp, doErr := DoWithRetry(client, req)
		require.NoError(t, doErr)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("DoWithRetry fail after all retries", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "some error", http.StatusInternalServerError)
		}))
		defer server.Close()

		client := New(WithTimeout(2 * time.Second))
		req, err := http.NewRequest("GET", server.URL, nil)
		require.NoError(t, err)

		resp, doErr := DoWithRetry(client, req)
		require.Error(t, doErr)
		assert.Nil(t, resp)
	})
}
