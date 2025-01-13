# HTTPCLIENT

A batteries-included HTTP client for Go that provides sane defaults and handles crucial configurations often overlooked in Go's default client.

## Why Use This?

- **Default Configuration**: Pre-configured with timeout values and connection settings
  - Connection timeout: 5s
  - TLS handshake timeout: 5s
  - Response timeout: 15s
  - Keep-alive: 15s
  - Idle connection timeout: 90s

- **Automatic Retry Mechanism**: Built-in exponential backoff retry for failed requests
  - Maximum retries: 5 attempts
  - Initial delay: 500ms
  - Maximum delay: 3s
  - Exponential backoff with factor of 2
  - Retries on network errors and 4xx/5xx responses
  - Respects context cancellation
  - Smart backoff strategy with exponential increase

- **Comprehensive Configuration**: Simple interface for all important HTTP client settings that are typically forgotten

## Usage

```go
// Default client with production-ready settings
client := httpclient.New()

// Custom settings when needed
client := httpclient.New(
    httpclient.WithTimeout(10 * time.Second),
    httpclient.WithMaxIdleConns(100),
)

// Automatic retry handling
req, _ := http.NewRequest(http.MethodGet, "https://api.example.com", nil)
resp, err := httpclient.DoWithRetry(client, req)
```

## Available Options

- `WithTimeout`: Set client timeout
- `WithTLSHandshakeTimeout`: Set TLS handshake timeout
- `WithResponseHeaderTimeout`: Set response header timeout
- `WithIdleConnTimeout`: Set idle connection timeout
- `WithMaxIdleConns`: Set maximum idle connections
- `WithMaxIdleConnsPerHost`: Set maximum idle connections per host
- `WithForceHTTP2Disabled`: Disable HTTP/2
- `WithProxy`: Set proxy function
- `WithDialerTimeout`: Set dialer timeout
- `WithDialerKeepAlive`: Set keep-alive duration

## Retry Behavior

The client implements a sophisticated retry mechanism with the following characteristics:

- Retries on both network errors and HTTP status codes >= 400
- Uses exponential backoff starting at 500ms
- Each retry doubles the delay time
- Maximum delay capped at 3 seconds
- Maximum of 5 retry attempts
- Honors context cancellation for graceful shutdown
- Automatically closes response bodies between retries

Example with context:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

req, _ := http.NewRequest(http.MethodGet, "https://api.example.com", nil)
req = req.WithContext(ctx)

resp, err := httpclient.DoWithRetry(client, req)
```

## Testing

```bash
go test -v ./...
```

## TODO

- [ ] Add support for custom retry configuration
- [ ] Add support for custom retry conditions

## License

MIT
