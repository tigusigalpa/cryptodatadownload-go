# CryptoDataDownload Go SDK

Typed Go client for the [CryptoDataDownload](https://www.cryptodatadownload.com/api/)
REST API — Deribit options analytics, Binance OHLCV with multi-year depth,
cross-venue funding rates, on-chain signals, pre-computed ML features and
risk metrics. Covers **all 116 GET operations** of
`https://api.cryptodatadownload.com/v1`, generated from and verified against
the [live OpenAPI schema](https://api.cryptodatadownload.com/?format=openapi).

- Zero dependencies — stdlib only (`net/http`, `context`, `encoding/json`)
- `context.Context` first argument on every call; injectable `*http.Client`
- Token auth, built-in 1 rps rate limiting, retries with `Retry-After`, typed errors
- `*Raw` CSV/XLSX downloads on the 43 operations that declare `return`

## Install

```bash
go get github.com/tigusigalpa/cryptodatadownload-go
```

## Quick start

```go
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	cdd "github.com/tigusigalpa/cryptodatadownload-go"
)

func main() {
	client, err := cdd.NewClient(os.Getenv("CDD_API_KEY"))
	if err != nil {
		panic(err)
	}
	ctx := context.Background()

	// Binance spot OHLCV (optional fields are pointers — nil means "server default")
	rows, err := client.BinanceOHLC.Spot(ctx, cdd.SpotParams{
		Symbol:   cdd.Ptr("BTCUSDT"),
		Interval: cdd.Ptr("1h"),
		Limit:    cdd.Ptr(500),
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(len(rows), "rows")

	// Partition read — required params are plain fields
	features, err := client.Aggregates.FeaturesMlAll(ctx, cdd.FeaturesMlAllParams{
		Symbol: "BTCUSDT",
		Date:   time.Now().UTC().AddDate(0, 0, -1),
		Fields: []string{"symbol", "date", "vol_rank_z"},
	})
	_ = features
}
```

Every method returns the unwrapped `result` payload — `[]map[string]any` for
row lists, `map[string]any` for objects and the non-envelope legacy endpoints.
Required params are plain struct fields; optional params are pointers (nil →
server default); `date` params take `time.Time` normalized to UTC
`YYYY-MM-DD`.

## Services

21 service groups mirroring the official OpenAPI tags. Full method list with
endpoints, parameters and doc links:
[`docs/api-coverage.md`](docs/api-coverage.md) (116 operations).

| Field | Coverage |
|---|---|
| `client.Aggregates` | Composite rows: ML features, on-chain & options dashboards, risk panel |
| `client.BinanceOHLC` / `BinanceSummary` | Spot/futures/options candles, tick data, trade & funding summaries |
| `client.DeribitOHLC` / `DeribitSummary` | Futures/options OHLC, funding, greeks, flows, ratios, large trades |
| `client.OptionsAnalytics` / `OptionsPricing` | Vol surface, skew, term structure, BSM pricing |
| `client.Funding` | Cross-venue perpetual funding rates |
| `client.OnChain` | BTC/ETH network metrics, mempool, whales, NVT |
| `client.Breadth` | Advance/decline, 52-week highs/lows, MA tracking |
| `client.Technicals` / `Risk` | TA indicators, realized vol, VaR, drawdowns, correlations |
| `client.MLMarket` / `MLFutures` / `MLMacro` / `MLCrossSectional` / `MLMicrostructure` / `MLOptions` / `MLRegime` | Pre-computed ML feature sets |
| `client.Govt` | CFTC COT, Treasury par-yield curves |
| `client.Catalog` | Available symbols/dates/contracts discovery |

## Options

```go
client, err := cdd.NewClient(key,
	cdd.WithHTTPClient(custom),        // injected & reused — net/http clients are safe for concurrent use
	cdd.WithBaseURL("http://stub"),    // httptest / proxies
	cdd.WithRequestsPerSecond(0.5),    // default 1.0; 0 disables
	cdd.WithRetryMaxAttempts(5),
	cdd.WithTimeout(30*time.Second),
	cdd.WithMaxResponseBodyBytes(32<<20),
	cdd.WithUserAgent("myapp/1.0"),
	cdd.WithLogger(log.Default()),
)
```

## Authentication

The SDK sends exactly one header — `Authorization: Token <key>` — on every
request. The OpenAPI security scheme is *named* `Bearer`, but the official
[authentication guide](https://www.cryptodatadownload.com/blog/posts/how-to-authenticate-cryptodatadownload-api-token/)
and the schema itself specify the `Token` prefix; this package does not emit
`Bearer`.

The API also accepts `?auth_token=<key>`. This SDK never uses it and does not
expose it: keys in URLs leak into logs, referers and telemetry.

## Rate limits & retries

Per token: 60 req/min burst, 10 000 req per rolling 24 h, plus **100 req/h on
"heavy" endpoints** (all `interval=1m` pulls and the endpoints flagged in
[api-coverage](docs/api-coverage.md)). Violations return `429` with
`Retry-After`.

The client applies a minimum-interval limiter at **1 req/s** by default and
retries up to 3 times on 429/5xx/transport errors with exponential backoff +
full jitter; `Retry-After` wins over computed backoff, and `ctx` cancellation
is honored during waits. 400/401/404 are never retried.

## Errors

```go
rows, err := client.MLMarket.SpotReturns(ctx, cdd.SpotReturnsParams{
	Symbol: "BTCUSDT", Date: someDay,
})
var apiErr *cdd.ApiError
if errors.As(err, &apiErr) {
	switch {
	case apiErr.IsNotFound():      // 404 — partition for that date doesn't exist yet
	case apiErr.IsRateLimited():
		secs, _ := apiErr.RetryAfter()
	case apiErr.IsAuthentication(): // 401 — bad/expired token
	}
}
var tErr *cdd.TransportError
if errors.As(err, &tErr) { /* network/decode failure below the API contract */ }
var argErr *cdd.ArgumentError
if errors.As(err, &argErr) { /* client-side param validation, no request was sent */ }
```

`404` on partition endpoints is a **data state** ("no partition yet"), not a
routing failure — treat it accordingly.

## CSV/XLSX downloads

43 legacy operations declare a `return` query param (`JSON` default,
`CSV`/`XLSX` supported). Each has a `*Raw` variant returning the body bytes:

```go
data, err := client.BinanceOHLC.SpotRaw(ctx, cdd.SpotParams{
	Symbol: cdd.Ptr("BTCUSDT"),
}, cdd.FormatCSV)
_ = os.WriteFile("btc.csv", data, 0o644)
```

No operation accepts a `format` parameter (it appears only in marketing
quick-starts), so the SDK doesn't have one — `return` is exposed only where
the schema declares it.

## Regenerating

The service files (`zz_gen_*.go`), contract test and manifest are generated
from the OpenAPI schema:

```bash
php tools/generate.php path/to/openapi.json
```

## Testing

```bash
go test ./...        # unit tests; httptest only, no live calls
go test -race ./...
go vet ./...
```

`TestGeneratedContractCoverage` exercises all 116 operations against a stub
server and asserts method, path, required params, and the `Token` header.

## Notes & known discrepancies

- Marketing pages say "112+ endpoints" — the live schema publishes **116**
  GET operations; this SDK covers the live schema.
- `Authorization: Token` is correct even though the security scheme is named
  `Bearer` (verified against the official auth guide and schema).
- 401/429 are not declared in the schema but are part of the real HTTP
  contract — they are always handled.
- The vendor publishes a single Swagger UI with no per-route pages. Every
  method's godoc and every row of `docs/api-coverage.md` links to the shared
  UI plus a JSON-Pointer anchor into the published schema — those are the
  only two real doc URLs.

## License

MIT. See [LICENSE](LICENSE).
