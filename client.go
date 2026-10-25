package cryptodatadownload

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tigusigalpa/cryptodatadownload-go/internal/clock"
)

// ErrAPIKeyRequired is returned when an operation is attempted without an
// API key. Every CryptoDataDownload operation requires a token.
var ErrAPIKeyRequired = errors.New("cryptodatadownload: API key is required (see https://www.cryptodatadownload.com/api/)")

// Client is the top-level CryptoDataDownload client. It exposes grouped
// service clients covering all 116 GET operations of
// https://api.cryptodatadownload.com/v1.
//
// Create one client and reuse it for the lifetime of your application. The
// API key is sent as "Authorization: Token <key>" by default. Query-based
// authentication requires WithQueryTokenAuth. Diagnostics redact the key.
// All services are safe for concurrent use.
type Client struct {
	apiKey string
	cfg    config
	gate   *rateGate

	// Aggregates provides pre-joined composite rows.
	Aggregates *AggregatesService
	// BinanceOHLC provides Binance spot/futures/options candles and tick data.
	BinanceOHLC *BinanceOHLCService
	// BinanceSummary provides Binance trade and funding summaries.
	BinanceSummary *BinanceSummaryService
	// DeribitOHLC provides Deribit futures/options candles and volatility.
	DeribitOHLC *DeribitOHLCService
	// DeribitSummary provides Deribit options summaries: greeks, flows, ratios.
	DeribitSummary *DeribitSummaryService
	// Catalog lists available symbols, dates, and contracts.
	Catalog *CatalogService
	// Breadth provides market breadth indicators.
	Breadth *BreadthService
	// OnChain provides BTC/ETH on-chain network metrics.
	OnChain *OnChainService
	// OptionsAnalytics provides volatility-surface and skew analytics.
	OptionsAnalytics *OptionsAnalyticsService
	// OptionsPricing provides BSM option pricing tools.
	OptionsPricing *OptionsPricingService
	// Funding provides cross-venue perpetual funding rates.
	Funding *FundingService
	// MLCrossSectional provides cross-sectional ML features.
	MLCrossSectional *MLCrossSectionalService
	// MLFutures provides futures-positioning ML features.
	MLFutures *MLFuturesService
	// MLMacro provides macro-signal ML features.
	MLMacro *MLMacroService
	// MLMarket provides market-basics ML features.
	MLMarket *MLMarketService
	// MLMicrostructure provides microstructure ML features.
	MLMicrostructure *MLMicrostructureService
	// MLOptions provides options-implied ML features.
	MLOptions *MLOptionsService
	// MLRegime provides regime & anomaly ML features.
	MLRegime *MLRegimeService
	// Risk provides risk & volatility metrics.
	Risk *RiskService
	// Technicals provides technical-analysis indicators.
	Technicals *TechnicalsService
	// Govt provides US government data: CFTC COT and Treasury curves.
	Govt *GovtService
}

// NewClient creates a CryptoDataDownload client with the supplied API key
// and options. It performs no network calls. Authentication uses a Token
// header unless WithQueryTokenAuth is explicitly enabled. Diagnostics redact
// the key in either mode.
func NewClient(apiKey string, options ...Option) (*Client, error) {
	cfg := defaultConfig()
	cfg.token = apiKey
	for _, opt := range options {
		if opt == nil {
			return nil, errors.New("cryptodatadownload: nil client option")
		}
		opt(&cfg)
	}
	apiKey = strings.TrimSpace(cfg.token)
	cfg.token = ""
	for i := 0; i < len(apiKey); i++ {
		if apiKey[i] < 0x20 || apiKey[i] == 0x7f {
			return nil, errors.New("cryptodatadownload: API key contains an invalid header character")
		}
	}
	if err := validateBaseURL(cfg.baseURL); err != nil {
		return nil, err
	}
	if cfg.httpClient == nil {
		cfg.httpClient = &http.Client{Timeout: 60 * time.Second}
	} else {
		clone := *cfg.httpClient
		cfg.httpClient = &clone
	}
	checkRedirect := cfg.httpClient.CheckRedirect
	cfg.httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) > 0 && (!strings.EqualFold(req.URL.Scheme, via[0].URL.Scheme) || !strings.EqualFold(req.URL.Host, via[0].URL.Host)) {
			return http.ErrUseLastResponse
		}
		if checkRedirect != nil {
			return checkRedirect(req, via)
		}
		if len(via) >= 10 {
			return errors.New("stopped after 10 redirects")
		}
		return nil
	}
	if cfg.httpTimeout != nil {
		if *cfg.httpTimeout < 0 {
			return nil, fmt.Errorf("cryptodatadownload: timeout must not be negative: %s", *cfg.httpTimeout)
		}
		cfg.httpClient.Timeout = *cfg.httpTimeout
	}
	if cfg.clock == nil {
		cfg.clock = clock.SystemClock{}
	}
	if cfg.sleeper == nil {
		cfg.sleeper = clock.SystemSleeper{}
	}
	if cfg.retry.MaxAttempts == 0 {
		cfg.retry.MaxAttempts = DefaultRetryPolicy().MaxAttempts
	} else if cfg.retry.MaxAttempts < 0 {
		return nil, fmt.Errorf("cryptodatadownload: retry max attempts must not be negative: %d", cfg.retry.MaxAttempts)
	}
	if cfg.retry.InitialBackoff < 0 || cfg.retry.MaxBackoff < 0 {
		return nil, errors.New("cryptodatadownload: retry backoff durations must not be negative")
	}
	if cfg.retry.MaxAttempts < 1 {
		cfg.retry.MaxAttempts = 1
	}
	if err := validateRequestsPerSecond(cfg.requestsPerSecond); err != nil {
		return nil, err
	}
	if cfg.maxResponseBodyBytes <= 0 {
		return nil, fmt.Errorf("cryptodatadownload: max response body bytes must be positive, got %d", cfg.maxResponseBodyBytes)
	}
	if cfg.maxErrorBodyBytes <= 0 {
		return nil, errors.New("cryptodatadownload: max error body bytes must be positive")
	}

	c := &Client{
		apiKey: apiKey,
		cfg:    cfg,
	}
	c.gate = newRateGate(cfg.requestsPerSecond, cfg.clock, cfg.sleeper)

	c.Aggregates = &AggregatesService{client: c}
	c.BinanceOHLC = &BinanceOHLCService{client: c}
	c.BinanceSummary = &BinanceSummaryService{client: c}
	c.DeribitOHLC = &DeribitOHLCService{client: c}
	c.DeribitSummary = &DeribitSummaryService{client: c}
	c.Catalog = &CatalogService{client: c}
	c.Breadth = &BreadthService{client: c}
	c.OnChain = &OnChainService{client: c}
	c.OptionsAnalytics = &OptionsAnalyticsService{client: c}
	c.OptionsPricing = &OptionsPricingService{client: c}
	c.Funding = &FundingService{client: c}
	c.MLCrossSectional = &MLCrossSectionalService{client: c}
	c.MLFutures = &MLFuturesService{client: c}
	c.MLMacro = &MLMacroService{client: c}
	c.MLMarket = &MLMarketService{client: c}
	c.MLMicrostructure = &MLMicrostructureService{client: c}
	c.MLOptions = &MLOptionsService{client: c}
	c.MLRegime = &MLRegimeService{client: c}
	c.Risk = &RiskService{client: c}
	c.Technicals = &TechnicalsService{client: c}
	c.Govt = &GovtService{client: c}

	return c, nil
}

// NewClientWithOptions creates a client using only options, including
// WithToken. NewClient retains its positional-token signature for compatibility.
func NewClientWithOptions(options ...Option) (*Client, error) {
	return NewClient("", options...)
}

// BaseURL returns the REST base URL in use.
func (c *Client) BaseURL() string { return c.cfg.baseURL }

// UserAgent returns the User-Agent header sent on requests.
func (c *Client) UserAgent() string { return c.cfg.userAgent }

func validateBaseURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return errors.New("cryptodatadownload: invalid base URL; expected an absolute HTTP(S) URL")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("cryptodatadownload: invalid base URL scheme %q", u.Scheme)
	}
	if u.User != nil {
		return errors.New("cryptodatadownload: base URL must not contain user credentials")
	}
	if u.RawQuery != "" || u.ForceQuery || strings.Contains(raw, "#") {
		return fmt.Errorf("cryptodatadownload: base URL must not contain a query or fragment")
	}
	return nil
}
