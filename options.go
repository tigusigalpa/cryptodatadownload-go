package cryptodatadownload

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/tigusigalpa/cryptodatadownload-go/internal/clock"
)

// DefaultBaseURL is the CryptoDataDownload REST base URL.
const DefaultBaseURL = "https://api.cryptodatadownload.com/v1"

// DefaultUserAgent is the default User-Agent header for requests.
const DefaultUserAgent = "cryptodatadownload-go/1.0.0"

// DefaultMaxResponseBodyBytes is the largest response body read into memory
// by default, for JSON and raw CSV/XLSX downloads alike.
const DefaultMaxResponseBodyBytes int64 = 16 << 20 // 16 MiB

// DefaultMaxErrorBodyBytes caps diagnostics retained from an unsuccessful
// response, independently of the size of a successful download.
const DefaultMaxErrorBodyBytes int64 = 64 << 10 // 64 KiB

// DefaultRequestsPerSecond is the built-in steady-state request ceiling.
// The API allows 60 req/min burst and 10 000 req per rolling 24 h. A 1 rps
// ceiling paces short bursts; applications must separately budget daily
// usage and heavy endpoints (100 req/h).
const DefaultRequestsPerSecond = 1.0

// RetryPolicy controls automatic retry of failed requests.
//
// Every CryptoDataDownload operation is an idempotent GET, so all calls are
// retried on transient transport errors and statuses 429, 500, 502, 503 and 504
// using capped exponential backoff with jitter and honouring Retry-After.
// 400/401/404 are permanent and never retried.
type RetryPolicy struct {
	// MaxAttempts is the total number of attempts including the first. A value
	// of 1 disables retry. Zero means use the default.
	MaxAttempts int
	// InitialBackoff is the first backoff delay. Zero means use the default.
	InitialBackoff time.Duration
	// MaxBackoff caps the backoff delay. Zero means use the default.
	MaxBackoff time.Duration
}

// DefaultRetryPolicy returns the default retry policy: up to 3 attempts,
// 500 ms initial backoff, 30 s max backoff.
func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxAttempts:    3,
		InitialBackoff: 500 * time.Millisecond,
		MaxBackoff:     30 * time.Second,
	}
}

// config holds resolved client configuration after applying options.
type config struct {
	token                string
	baseURL              string
	userAgent            string
	httpClient           *http.Client
	httpTimeout          *time.Duration
	retry                RetryPolicy
	logger               *log.Logger
	clock                Clock
	sleeper              Sleeper
	requestsPerSecond    float64
	maxResponseBodyBytes int64
	maxErrorBodyBytes    int64
	rateLimiter          RateLimiter
	queryTokenAuth       bool
}

// Option configures a Client.
type Option func(*config)

// WithToken supplies the API token when using NewClientWithOptions. When
// combined with NewClient it overrides the positional token.
func WithToken(token string) Option { return func(c *config) { c.token = token } }

// Clock supplies the current time for Retry-After and rate-limit
// calculations. Most callers should use the system clock configured by
// default; the interface exists for deterministic tests.
type Clock interface {
	Now() time.Time
}

// Sleeper waits between retry attempts and rate-limit slots while respecting
// cancellation. Most callers should use the context-aware system sleeper
// configured by default.
type Sleeper interface {
	Sleep(context.Context, time.Duration) error
}

// RateLimiter coordinates request attempts. A single implementation can be
// shared across clients using the same API token. Wait must respect context
// cancellation and be safe for concurrent use.
type RateLimiter interface {
	Wait(context.Context) error
}

// WithRateLimiter replaces the built-in per-client limiter. Sharing a limiter
// coordinates multiple clients in this process; cross-process budgets require
// an implementation backed by shared storage. Nil uses the built-in limiter.
func WithRateLimiter(limiter RateLimiter) Option {
	return func(c *config) { c.rateLimiter = limiter }
}

// WithQueryTokenAuth explicitly opts into sending the token in auth_token
// instead of Authorization. URLs can expose credentials in upstream access
// logs and proxies. SDK response metadata and printable errors redact it.
func WithQueryTokenAuth(enabled bool) Option {
	return func(c *config) { c.queryTokenAuth = enabled }
}

// WithDebugLogger enables redacted retry diagnostics. Logging is off by default.
func WithDebugLogger(logger *log.Logger) Option { return WithLogger(logger) }

// WithBaseURL overrides the REST base URL, e.g. for httptest or a proxy.
func WithBaseURL(url string) Option {
	return func(c *config) { c.baseURL = url }
}

// WithHTTPClient supplies a custom *http.Client (transport, proxy, timeouts).
// The client is reused for every request; it is safe for concurrent use and
// designed to be shared. It is never mutated: when combined with WithTimeout
// a shallow copy is configured instead. Redirects are restricted to the
// original scheme and host, then the supplied CheckRedirect policy is applied.
func WithHTTPClient(h *http.Client) Option {
	return func(c *config) { c.httpClient = h }
}

// WithTimeout sets the per-request timeout.
func WithTimeout(d time.Duration) Option {
	return func(c *config) {
		c.httpTimeout = &d
	}
}

// WithUserAgent overrides the User-Agent header.
func WithUserAgent(ua string) Option {
	return func(c *config) { c.userAgent = ua }
}

// WithRetryPolicy overrides the retry policy.
func WithRetryPolicy(p RetryPolicy) Option {
	return func(c *config) { c.retry = p }
}

// WithRetryMaxAttempts is a convenience option setting the retry attempt
// count; 1 disables retry.
func WithRetryMaxAttempts(n int) Option {
	return func(c *config) { c.retry.MaxAttempts = n }
}

// WithLogger sets the logger used for redacted diagnostic messages.
func WithLogger(l *log.Logger) Option {
	return func(c *config) { c.logger = l }
}

// WithClock injects a Clock for deterministic tests.
func WithClock(cl Clock) Option {
	return func(c *config) { c.clock = cl }
}

// WithSleeper injects a Sleeper for deterministic retry/rate-limit tests.
func WithSleeper(s Sleeper) Option {
	return func(c *config) { c.sleeper = s }
}

// WithRequestsPerSecond sets the client-side steady-state rate limit in
// requests per second. The default is 1.0, which keeps callers under the
// 60 req/min burst limit. Set 0 to disable the built-in limiter. Positive
// values must be finite and produce an interval representable by time.Duration.
func WithRequestsPerSecond(rps float64) Option {
	return func(c *config) { c.requestsPerSecond = rps }
}

// WithMaxResponseBodyBytes limits buffered successful responses.
// The limit applies to JSON and raw CSV/XLSX downloads. Values must be
// positive. The default is 16 MiB. Streaming methods do not buffer success
// bodies and are not subject to this limit.
func WithMaxResponseBodyBytes(n int64) Option {
	return func(c *config) { c.maxResponseBodyBytes = n }
}

// WithMaxErrorBodyBytes limits retained error diagnostics. A larger error
// response is truncated while its HTTP status remains available as an APIError.
func WithMaxErrorBodyBytes(n int64) Option {
	return func(c *config) { c.maxErrorBodyBytes = n }
}

func defaultConfig() config {
	return config{
		baseURL:              DefaultBaseURL,
		userAgent:            DefaultUserAgent,
		retry:                DefaultRetryPolicy(),
		clock:                clock.SystemClock{},
		sleeper:              clock.SystemSleeper{},
		requestsPerSecond:    DefaultRequestsPerSecond,
		maxResponseBodyBytes: DefaultMaxResponseBodyBytes,
		maxErrorBodyBytes:    DefaultMaxErrorBodyBytes,
	}
}
