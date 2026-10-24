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
// API key is sent as "Authorization: Token <key>" and never appears in URLs,
// error messages or logs. All services are safe for concurrent use.
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
// and options. It performs no network calls. The key is sent as a Token
// header; it is never logged or placed in a query string.
func NewClient(apiKey string, options ...Option) (*Client, error) {
	apiKey = strings.TrimSpace(apiKey)

	cfg := defaultConfig()
	for _, opt := range options {
		if opt == nil {
			return nil, errors.New("cryptodatadownload: nil client option")
		}
		opt(&cfg)
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
	if cfg.requestsPerSecond < 0 {
		return nil, fmt.Errorf("cryptodatadownload: requests per second must not be negative: %v", cfg.requestsPerSecond)
	}
	if cfg.maxResponseBodyBytes <= 0 {
		return nil, fmt.Errorf("cryptodatadownload: max response body bytes must be positive, got %d", cfg.maxResponseBodyBytes)
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

// BaseURL returns the REST base URL in use.
func (c *Client) BaseURL() string { return c.cfg.baseURL }

// UserAgent returns the User-Agent header sent on requests.
func (c *Client) UserAgent() string { return c.cfg.userAgent }

func validateBaseURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("cryptodatadownload: invalid base URL %q", raw)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("cryptodatadownload: invalid base URL scheme %q", u.Scheme)
	}
	if u.RawQuery != "" || u.Fragment != "" {
		return fmt.Errorf("cryptodatadownload: base URL must not contain a query or fragment")
	}
	return nil
}
