# CryptoDataDownload Go SDK

Fetch market data from [CryptoDataDownload](https://www.cryptodatadownload.com/api/)
in your Go application. Start with a few candles, look up available contracts,
or load funding rates, on-chain metrics and ML features for a particular day.
Each operation has its own query struct, so your editor can help you find the
parameters you need.

The SDK covers the **116 GET operations** in the bundled OpenAPI snapshot and
uses only the Go standard library. It includes request timeouts, retries,
rate limiting, HTTP metadata and CSV/XLSX downloads. You need **Go 1.22+** and
a CryptoDataDownload API token to make live requests.

- [Install and make your first request](#your-first-request)
- [Understand queries and returned data](#working-with-queries-and-results)
- [Load features, funding, on-chain and risk data](#more-things-to-build)
- [Download CSV or XLSX](#download-csv-or-xlsx)
- [Handle errors and missing partitions](#handle-errors)
- [Configure the client](#configure-your-client)
- [Test your application without an API token](#test-without-a-live-api)
- [Browse every endpoint](#endpoint-coverage)

## Your first request

Inside your application's Go module, install the package:

```bash
go get github.com/tigusigalpa/cryptodatadownload-go
```

Save this as `main.go` in your application. It asks for 24 hourly BTCUSDT
candles and prints the dates and closing prices returned by the API.

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	cdd "github.com/tigusigalpa/cryptodatadownload-go"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	client, err := cdd.NewClient(os.Getenv("CDD_API_KEY"))
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	rows, err := client.BinanceOHLC.Spot(ctx, cdd.SpotParams{
		Symbol:   cdd.Ptr("BTCUSDT"),
		Interval: cdd.Ptr("1h"),
		Limit:    cdd.Ptr(24),
	})
	if err != nil {
		return fmt.Errorf("load candles: %w", err)
	}

	fmt.Printf("Received %d candles\n", len(rows))
	for _, row := range rows {
		fmt.Printf("%v  close=%v\n", row["date"], row["close"])
	}
	return nil
}
```

Set your token and run the program. In Bash or Zsh:

```bash
export CDD_API_KEY="your-api-token"
go run .
```

In PowerShell:

```powershell
$env:CDD_API_KEY = "your-api-token"
go run .
```

Keep the token in your environment or secret manager. The client sends it
in `Authorization: Token <key>` by default; logging is off unless you enable it.
An empty token is accepted by the constructor, but a request will return
`ErrAPIKeyRequired`.

If you have cloned this repository, you can also run the
[existing examples](examples) with `go run ./examples/01-ohlc`,
`go run ./examples/02-ml-features`, `go run ./examples/03-csv` or
`go run ./examples/04-errors`. They use the same `CDD_API_KEY` variable.

## Working with queries and results

Create one client and reuse it. The client and its services support
concurrent calls; every request takes a `context.Context` so you can set a
deadline or cancel work when the caller goes away.

The examples below are small helpers you can put in your application.
They take an existing client and, where appropriate, the caller's context.

### Required fields, optional fields and defaults

A required parameter is a plain field. An optional parameter is usually a
pointer: `nil` leaves it out of the request, while `cdd.Ptr(value)` sends it
explicitly. The `Fields` projection parameter uses a slice of strings.

For example, setting only `Symbol` leaves the candle interval and row limit
to the API's defaults:

```go
import (
	"context"

	cdd "github.com/tigusigalpa/cryptodatadownload-go"
)

func loadDailyCandles(ctx context.Context, client *cdd.Client, symbol string) ([]map[string]any, error) {
	return client.BinanceOHLC.Spot(ctx, cdd.SpotParams{
		Symbol: cdd.Ptr(symbol),
	})
}
```

An explicit zero is still a supplied value: `Offset: cdd.Ptr(0)` sends
`offset=0`. Missing required values, invalid declared enums, negative
pagination values and limits above a documented maximum return an
`ArgumentError` before any request is sent.

Wire names stay specific to each operation. You use `StartDate` or `EndDate`
in Go, and the encoder chooses the route's exact `startdate`, `start_date`,
`enddate` or `end_date` spelling.

### Dates are UTC calendar days

Use `cdd.ParseDate("2026-10-24")` when you already have a `YYYY-MM-DD` string.
It returns a UTC `time.Time` and rejects invalid dates. Fields typed as
`time.Time` are converted to UTC before the client formats the day, so a
timestamp shortly after midnight in your local timezone can belong to the
previous UTC day.

Some older parameters, such as the candle `EndDate`, remain `*string` because
that is how their query model is exposed. Pass the date spelling documented
for that field; these legacy string fields do not go through `ParseDate`
automatically.

### Choose the result style you need

| Method style | You receive | Use it when |
|---|---|---|
| `Spot(ctx, params)` | Decoded payload and an error | You want to work with rows directly |
| `SpotWithResponse(ctx, params)` | `*Result[T]`, `*Response` and an error | You need status, headers or attempt count too |
| `SpotRaw(ctx, params, format)` | Buffered `[]byte` and an error | The download is small and you want the whole document |
| `SpotStream(ctx, params, format)` | `*RawResponse` and an error | You want to copy or parse a download as it arrives |

Every typed operation has a `WithResponse` companion. `Raw` and `Stream`
companions exist on the 43 operations whose OpenAPI definition declares
the `return` parameter.

Ordinary methods unwrap the API's `result` envelope. Most return
`[]map[string]any`; object responses return `map[string]any`. The three
Deribit symbol catalogs return `[]string`:
`OhlcDeribitFuturesAvailable`, `OhlcDeribitFuturesFundingAvailable` and
`OhlcDeribitOptionsAvailable`.

### Keep decimal values intact

Prices and volumes can arrive as decimal strings. The SDK leaves those
strings alone, and decodes JSON numeric values as `json.Number`. Avoid
assuming every numeric-looking field is a `float64`.

This helper gives you the original decimal text in either case:

```go
import (
	"encoding/json"
	"fmt"
)

func decimalText(value any) (string, error) {
	switch value := value.(type) {
	case string:
		return value, nil
	case json.Number:
		return value.String(), nil
	default:
		return "", fmt.Errorf("expected a decimal string or number, got %T", value)
	}
}
```

## More things to build

### Load a day's ML features

Partitioned datasets take a date. This helper loads a small selection of
columns from the combined ML feature endpoint. Call it with a symbol and
date string, for example `"BTCUSDT"` and `"2026-10-24"`.

```go
import (
	"context"

	cdd "github.com/tigusigalpa/cryptodatadownload-go"
)

func loadFeatureSnapshot(ctx context.Context, client *cdd.Client, symbol, date string) ([]map[string]any, error) {
	day, err := cdd.ParseDate(date)
	if err != nil {
		return nil, err
	}

	return client.Aggregates.FeaturesMlAll(ctx, cdd.FeaturesMlAllParams{
		Symbol: symbol,
		Date:   day,
		Fields: []string{"symbol", "date", "vol_rank_z", "momentum_12_1"},
	})
}
```

Choose a completed UTC day for historical work. A valid date does not
guarantee that its partition is available: a missing partition returns
`NotFoundError`. See [error handling](#handle-errors) for deciding what your
application should do next.

### Fetch aligned funding rates

Use the funding service when you want a cross-venue dataset for a symbol
and day. You can pass a day from `ParseDate` to this helper:

```go
import (
	"context"
	"time"

	cdd "github.com/tigusigalpa/cryptodatadownload-go"
)

func loadFunding(ctx context.Context, client *cdd.Client, symbol string, day time.Time) ([]map[string]any, error) {
	return client.Funding.FundingAligned(ctx, cdd.FundingAlignedParams{
		Symbol: symbol,
		Date:   day,
	})
}
```

### Read an on-chain metric

The on-chain service follows the same pattern. Here is Bitcoin NVT for a
UTC day; the declared `chain` enum accepts `"BTC"` or `"ETH"`.

```go
import (
	"context"
	"time"

	cdd "github.com/tigusigalpa/cryptodatadownload-go"
)

func loadBitcoinNVT(ctx context.Context, client *cdd.Client, day time.Time) ([]map[string]any, error) {
	return client.OnChain.FeaturesOnchainNvt(ctx, cdd.FeaturesOnchainNvtParams{
		Chain: "BTC",
		Date:  day,
	})
}
```

### Load rolling correlations

Some parameters have a small set of allowed values. The rolling correlation
window is one of `30`, `60`, `90` or `180` days; the client checks it locally.

```go
import (
	"context"
	"time"

	cdd "github.com/tigusigalpa/cryptodatadownload-go"
)

func loadThirtyDayCorrelations(ctx context.Context, client *cdd.Client, day time.Time) ([]map[string]any, error) {
	return client.Risk.RiskCorrelationsRolling(ctx, cdd.RiskCorrelationsRollingParams{
		Window: 30,
		Date:   day,
	})
}
```

Enum fields use strings or integers in the public query structs. Use the
values declared for that particular operation; similar-looking parameters
on another endpoint can have a different allowed set.

### Look up contracts and inspect the HTTP response

Catalog methods help you discover available symbols and contracts. The
`WithResponse` form is useful when you also want request diagnostics:

```go
import (
	"context"
	"fmt"

	cdd "github.com/tigusigalpa/cryptodatadownload-go"
)

func inspectDeribitCatalog(ctx context.Context, client *cdd.Client) error {
	result, response, err := client.Catalog.OhlcDeribitFuturesAvailableWithResponse(
		ctx, cdd.OhlcDeribitFuturesAvailableParams{},
	)
	if response != nil {
		fmt.Printf("HTTP %d, attempts=%d\n", response.StatusCode, response.Attempts)
	}
	if err != nil {
		return err
	}

	fmt.Printf("Available contracts: %d\n", len(result.Result))
	for _, symbol := range result.Result {
		fmt.Println(symbol)
	}
	return nil
}
```

`response.Headers` is a cloned, redacted header map, and `response.URL` is
safe SDK metadata with credentials redacted. HTTP and decode errors retain
available response metadata. A failure before receiving an HTTP response,
such as an invalid argument or a connection failure, can leave it `nil`.

### Request a page of large trades

Pagination belongs to the operations that declare it. For example, this
helper fetches one page of large Binance spot trades:

```go
import (
	"context"

	cdd "github.com/tigusigalpa/cryptodatadownload-go"
)

func loadLargeTradesPage(ctx context.Context, client *cdd.Client, symbol string, offset int) ([]map[string]any, error) {
	return client.BinanceSummary.SpotLargeTrades(ctx, cdd.SpotLargeTradesParams{
		Symbol: cdd.Ptr(symbol),
		MinUSD: cdd.Ptr(100000.0),
		Limit:  cdd.Ptr(250),
		Offset: cdd.Ptr(offset),
	})
}
```

Start with offset `0` and choose subsequent offsets in your application.
Each invocation fetches the page you requested. Check the operation's date
filters when paging through a dataset that can change between requests.

## Download CSV or XLSX

Use a buffered download when the complete file is convenient to keep in
memory. This example lets the caller choose where to save a CSV document
and checks both the API request and the file write:

```go
import (
	"context"
	"os"

	cdd "github.com/tigusigalpa/cryptodatadownload-go"
)

func saveCandlesCSV(ctx context.Context, client *cdd.Client, symbol, filename string) error {
	data, err := client.BinanceOHLC.SpotRaw(ctx, cdd.SpotParams{
		Symbol: cdd.Ptr(symbol),
		Limit:  cdd.Ptr(500),
	}, cdd.FormatCSV)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0o644)
}
```

### Stream a larger workbook to a writer

For a larger download, use `Stream` and give it an `io.Writer`. The writer
can be a file opened by your application, an HTTP response or another
destination. The SDK opens the response without buffering the success body.

```go
import (
	"context"
	"io"

	cdd "github.com/tigusigalpa/cryptodatadownload-go"
)

func writeCandlesXLSX(ctx context.Context, client *cdd.Client, symbol string, destination io.Writer) error {
	download, err := client.BinanceOHLC.SpotStream(ctx, cdd.SpotParams{
		Symbol: cdd.Ptr(symbol),
	}, cdd.FormatXLSX)
	if err != nil {
		return err
	}
	defer download.Close()

	_, err = download.WriteTo(destination)
	return err
}
```

The caller owns the returned body and must close it. `WriteTo` leaves it
open. Keep the request's context alive until the copy is finished; its
deadline and the HTTP client's timeout also apply while reading the stream.
Once reading begins, a failure is returned to you without replaying bytes
that may already have reached the destination.

### Process CSV records as they arrive

You can also parse a CSV stream directly with `encoding/csv`. This helper
passes each record to your callback, including the header record, without
collecting the entire file in a slice:

```go
import (
	"context"
	"encoding/csv"
	"errors"
	"io"

	cdd "github.com/tigusigalpa/cryptodatadownload-go"
)

func visitCandleCSV(ctx context.Context, client *cdd.Client, symbol string, visit func([]string) error) error {
	download, err := client.BinanceOHLC.SpotStream(ctx, cdd.SpotParams{
		Symbol: cdd.Ptr(symbol),
	}, cdd.FormatCSV)
	if err != nil {
		return err
	}
	defer download.Close()

	reader := csv.NewReader(download.Body)
	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		if err := visit(record); err != nil {
			return err
		}
	}
}
```

Buffered success bodies have a 16 MiB default limit, adjustable through
`WithMaxResponseBodyBytes`. Streaming success bodies are not subject to that
buffer limit. Error diagnostics use a separate 64 KiB default limit.

Formats are selected with `FormatJSON`, `FormatCSV` or `FormatXLSX` and sent
as the operation's `return` query parameter. Format support is listed per
operation in [API coverage](docs/api-coverage.md).

## Handle errors

Start with ordinary Go error handling. Use `errors.Is` for cancellation and
`errors.As` when your application needs to distinguish a missing partition,
an authentication problem or a response it could not decode.

For example, this helper treats a missing feature partition as an explicit
"not available" result. Other failures still reach the caller:

```go
import (
	"context"
	"errors"

	cdd "github.com/tigusigalpa/cryptodatadownload-go"
)

func featuresIfAvailable(ctx context.Context, client *cdd.Client, symbol, date string) ([]map[string]any, bool, error) {
	rows, err := loadFeatureSnapshot(ctx, client, symbol, date)
	var missing *cdd.NotFoundError
	if errors.As(err, &missing) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return rows, true, nil
}
```

On partitioned endpoints, HTTP 404 means that the partition was not found.
The SDK returns that error; the helper above makes an application-level
choice about how to represent it.

For a broader error report, you can use a helper like this:

```go
import (
	"context"
	"errors"
	"fmt"

	cdd "github.com/tigusigalpa/cryptodatadownload-go"
)

func describeRequestError(err error) string {
	if err == nil {
		return "request completed"
	}

	var argument *cdd.ArgumentError
	var decode *cdd.DecodeError
	var api *cdd.APIError
	var transport *cdd.TransportError

	switch {
	case errors.Is(err, context.Canceled):
		return "request canceled by its caller"
	case errors.Is(err, context.DeadlineExceeded):
		return "request deadline exceeded"
	case errors.As(err, &argument):
		return fmt.Sprintf("check parameter %s: %s", argument.Param, argument.Msg)
	case errors.As(err, &decode):
		return fmt.Sprintf("could not decode the response: %v", decode)
	case errors.As(err, &api):
		if api.IsRateLimited() {
			if seconds, ok := api.RetryAfter(); ok {
				return fmt.Sprintf("rate limited; server recommends waiting %d seconds", seconds)
			}
		}
		return fmt.Sprintf("API returned HTTP %d: %s", api.StatusCode, api.Message)
	case errors.As(err, &transport):
		return fmt.Sprintf("request transport failed: %v", transport)
	default:
		return err.Error()
	}
}
```

`APIError` is the common HTTP error type. Its concrete wrappers are
`BadRequestError` (400), `AuthenticationError` (401), `NotFoundError` (404),
`RateLimitError` (429) and `ServerError` (5xx). You can match either the
wrapper or the common type with `errors.As`; predicates also cover HTTP 403.
`AsAPIError(err)` is a convenience helper when you only need the common type.

Check `DecodeError` before `TransportError` when handling both: decode errors
also unwrap to `TransportError`. `APIError.RawBody` contains bounded,
redacted diagnostics, and `BodyTruncated` tells you whether they are incomplete.

## Configure your client

The defaults work for a small application making occasional requests.
Change the settings that matter to your workload:

| Setting | Default |
|---|---|
| Base URL | `https://api.cryptodatadownload.com/v1` |
| Timeout on the default HTTP client | 60 seconds |
| Retry attempts | 3 total, including the initial request |
| Backoff | Full jitter, starting with a 500 ms bound and capped at 30 seconds |
| Per-client pacing | 1 request per second |
| Buffered success body limit | 16 MiB |
| Retained error body limit | 64 KiB |
| Logging | Off |

For example, supply your own HTTP client and tune timeouts and response limits:

```go
import (
	"net/http"
	"time"

	cdd "github.com/tigusigalpa/cryptodatadownload-go"
)

func newConfiguredClient(token string) (*cdd.Client, error) {
	return cdd.NewClient(token,
		cdd.WithHTTPClient(&http.Client{Timeout: 30 * time.Second}),
		cdd.WithRetryMaxAttempts(3),
		cdd.WithMaxResponseBodyBytes(32<<20),
		cdd.WithMaxErrorBodyBytes(32<<10),
		cdd.WithUserAgent("my-market-data-app/1.0"),
	)
}
```

The SDK copies the supplied `http.Client` configuration and reuses its
transport. `WithTimeout` can override the timeout on that copy. A caller's
earlier context deadline still takes effect. Redirects stay on the original
scheme and host, and your own `CheckRedirect` can further restrict them.

If you prefer to configure everything with options:

```go
import (
	"time"

	cdd "github.com/tigusigalpa/cryptodatadownload-go"
)

func newClientFromOptions(token string) (*cdd.Client, error) {
	return cdd.NewClientWithOptions(
		cdd.WithToken(token),
		cdd.WithTimeout(20*time.Second),
	)
}
```

Use a constructor rather than a zero-value `Client`. Invalid URLs, nil options
and invalid client limits are rejected during construction. To investigate
retries, opt into `WithDebugLogger(log.Default())` or `WithLogger`; retry
diagnostics redact credentials.

### Retries and request budgets

The client retries temporary transport failures and HTTP **429, 500, 502,
503 and 504**. It honors `Retry-After` in seconds or HTTP-date form and uses
exponential backoff with full jitter. A server-requested delay can exceed the
configured backoff cap. Context cancellation interrupts waits.

HTTP 400, 401, 403 and 404, permanent transport failures and oversized
buffered success bodies are returned without another attempt.
Set `WithRetryMaxAttempts(1)` to disable retries, or use `WithRetryPolicy`
to choose the attempt count and backoff bounds. A zero policy field uses
its default; one attempt means one HTTP attempt.

Plan your request budget separately from how quickly requests are spaced:

- The published per-token limits are **60 requests/minute** and **10,000
  requests per rolling 24 hours**.
- The official catalog adds **100/hour** limits for the specific heavy
  operations and `interval=1m` routes marked in
  [API coverage](docs/api-coverage.md).
- Continuous traffic at the default 1 request/second would make 86,400
  requests/day. The per-second limiter does not track the daily budget.

Usually, sharing one client is enough. If separate parts of your application
need separate clients with the same token, they can share one limiter:

```go
import cdd "github.com/tigusigalpa/cryptodatadownload-go"

func clientsWithSharedLimit(token string) (*cdd.Client, *cdd.Client, error) {
	limiter, err := cdd.NewRateLimiter(1)
	if err != nil {
		return nil, nil, err
	}
	first, err := cdd.NewClient(token, cdd.WithRateLimiter(limiter))
	if err != nil {
		return nil, nil, err
	}
	second, err := cdd.NewClient(token, cdd.WithRateLimiter(limiter))
	if err != nil {
		return nil, nil, err
	}
	return first, second, nil
}
```

A shared limiter coordinates this process. For multiple processes or
machines, provide a concurrency-safe `RateLimiter` implementation backed
by shared state; its `Wait` method must honor context cancellation.
`WithRequestsPerSecond(0)` disables the built-in pacing;
use it for local test servers or when another limiter already owns the budget.

### Authentication details

The default header is `Authorization: Token <key>`. Although the OpenAPI
security scheme is named `Bearer`, its description and the official
[authentication guide](https://www.cryptodatadownload.com/blog/posts/how-to-authenticate-cryptodatadownload-api-token/)
use the `Token` prefix.

For integrations that specifically require query authentication,
`WithQueryTokenAuth(true)` opts into `auth_token` instead of the header.
SDK metadata and printable errors redact it. URLs can still reach upstream
access logs, so prefer the default header mode. More details are in the
[security notes](docs/security.md).

## Find the right service

The 21 service groups follow the published OpenAPI tags. Start with the group
closest to the dataset you need, then use its query struct to choose a symbol,
date or filter.

| Client field | What you can fetch |
|---|---|
| `BinanceOHLC`, `BinanceSummary` | Binance candles, tick bars, trade summaries and positioning data |
| `DeribitOHLC`, `DeribitSummary` | Deribit candles, funding and options summaries |
| `Catalog` | Available symbols, contracts and dates |
| `Aggregates` | Combined ML features, on-chain and options dashboards, risk panels |
| `Funding` | Venue-specific and aligned perpetual funding rates |
| `OnChain` | BTC/ETH network metrics and derived on-chain features |
| `OptionsAnalytics`, `OptionsPricing` | Volatility analytics and option pricing tools |
| `Risk`, `Technicals` | Risk metrics, correlations and technical indicators |
| `Breadth` | Market breadth indicators |
| `Govt` | CFTC COT data and Treasury yield curves |
| `MLMarket`, `MLFutures`, `MLMacro` | Market, futures-positioning and macro features |
| `MLCrossSectional`, `MLMicrostructure`, `MLOptions`, `MLRegime` | Cross-sectional, microstructure, options and regime features |

## Test without a live API

Point `WithBaseURL` at an `httptest.Server` to exercise your application's
request flow with a fake token and a response you control. Put this example
in `marketdata_test.go` in an application using `package main` (or adjust
the package name to match yours):

```go
package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	cdd "github.com/tigusigalpa/cryptodatadownload-go"
)

func TestLoadCandles(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/data/ohlc/binance/spot/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Token test-key" {
			t.Error("missing Token authentication")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"result":[{"symbol":"BTCUSDT","close":"60000.10"}]}`)
	}))
	defer server.Close()

	client, err := cdd.NewClient("test-key",
		cdd.WithBaseURL(server.URL+"/v1"),
		cdd.WithRequestsPerSecond(0),
		cdd.WithRetryMaxAttempts(1),
	)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := loadDailyCandles(context.Background(), client, "BTCUSDT")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0]["close"] != "60000.10" {
		t.Fatalf("unexpected rows: %v", rows)
	}
}
```

You can also inject an `http.Client` with a custom `RoundTripper` through
`WithHTTPClient`.

From the SDK repository, run the offline checks with:

```bash
go test ./...
go test -race ./...
go test -cover ./...
go vet ./...
golint -set_exit_status ./... # if golint is installed
```

CI covers Go 1.22 and Go 1.27 and runs pinned Staticcheck checks. See
[CONTRIBUTING.md](CONTRIBUTING.md) for the complete development workflow.
The optional live catalog test runs only when both
`CDD_RUN_INTEGRATION_TESTS=1` and `CDD_API_TOKEN` are set; ordinary tests
need no paid token.

## Endpoint coverage

The table below maps every operation and its method variants to the exact
route, production URL and official documentation. For required parameters,
defaults, enums and additional hourly limits, use the
[full API coverage reference](docs/api-coverage.md).

<!-- BEGIN GENERATED ENDPOINT COVERAGE -->
<details>
<summary>All 116 operations and method variants</summary>

| SDK methods | Method / exact path | Production URL | Official documentation |
|---|---|---|---|
| `Aggregates.FeaturesMlAll`<br>`Aggregates.FeaturesMlAllWithResponse` | `GET /v1/data/aggregates/features/ml/all/` | [production](https://api.cryptodatadownload.com/v1/data/aggregates/features/ml/all/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1aggregates~1features~1ml~1all~1/get) |
| `Aggregates.OnchainDashboard`<br>`Aggregates.OnchainDashboardWithResponse` | `GET /v1/data/aggregates/onchain/dashboard/` | [production](https://api.cryptodatadownload.com/v1/data/aggregates/onchain/dashboard/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1aggregates~1onchain~1dashboard~1/get) |
| `Aggregates.OptionsDashboard`<br>`Aggregates.OptionsDashboardWithResponse` | `GET /v1/data/aggregates/options/dashboard/` | [production](https://api.cryptodatadownload.com/v1/data/aggregates/options/dashboard/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1aggregates~1options~1dashboard~1/get) |
| `Aggregates.RiskPanel`<br>`Aggregates.RiskPanelWithResponse` | `GET /v1/data/aggregates/risk/panel/` | [production](https://api.cryptodatadownload.com/v1/data/aggregates/risk/panel/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1aggregates~1risk~1panel~1/get) |
| `Breadth.N52wkHighsLows`<br>`Breadth.N52wkHighsLowsWithResponse` | `GET /v1/data/breadth/52wk-highs-lows/` | [production](https://api.cryptodatadownload.com/v1/data/breadth/52wk-highs-lows/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1breadth~152wk-highs-lows~1/get) |
| `Breadth.AdvanceDecline`<br>`Breadth.AdvanceDeclineWithResponse`<br>`Breadth.AdvanceDeclineRaw`<br>`Breadth.AdvanceDeclineStream` | `GET /v1/data/breadth/advance-decline/` | [production](https://api.cryptodatadownload.com/v1/data/breadth/advance-decline/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1breadth~1advance-decline~1/get) |
| `Breadth.MovingAverageTracking`<br>`Breadth.MovingAverageTrackingWithResponse` | `GET /v1/data/breadth/moving-average-tracking/` | [production](https://api.cryptodatadownload.com/v1/data/breadth/moving-average-tracking/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1breadth~1moving-average-tracking~1/get) |
| `Govt.CftcCot`<br>`Govt.CftcCotWithResponse`<br>`Govt.CftcCotRaw`<br>`Govt.CftcCotStream` | `GET /v1/data/cftc/cot/` | [production](https://api.cryptodatadownload.com/v1/data/cftc/cot/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1cftc~1cot~1/get) |
| `Govt.TreasuryParYieldsReal`<br>`Govt.TreasuryParYieldsRealWithResponse`<br>`Govt.TreasuryParYieldsRealRaw`<br>`Govt.TreasuryParYieldsRealStream` | `GET /v1/data/treasury/par-yields-real/` | [production](https://api.cryptodatadownload.com/v1/data/treasury/par-yields-real/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1treasury~1par-yields-real~1/get) |
| `Govt.TreasuryParYields`<br>`Govt.TreasuryParYieldsWithResponse`<br>`Govt.TreasuryParYieldsRaw`<br>`Govt.TreasuryParYieldsStream` | `GET /v1/data/treasury/par-yields/` | [production](https://api.cryptodatadownload.com/v1/data/treasury/par-yields/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1treasury~1par-yields~1/get) |
| `Catalog.CftcCotAvailable`<br>`Catalog.CftcCotAvailableWithResponse` | `GET /v1/data/cftc/cot/available/` | [production](https://api.cryptodatadownload.com/v1/data/cftc/cot/available/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1cftc~1cot~1available~1/get) |
| `Catalog.OhlcBinanceAllAvailable`<br>`Catalog.OhlcBinanceAllAvailableWithResponse` | `GET /v1/data/ohlc/binance/all/available/` | [production](https://api.cryptodatadownload.com/v1/data/ohlc/binance/all/available/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1ohlc~1binance~1all~1available~1/get) |
| `Catalog.OhlcDeribitFuturesAvailable`<br>`Catalog.OhlcDeribitFuturesAvailableWithResponse` | `GET /v1/data/ohlc/deribit/futures/available/` | [production](https://api.cryptodatadownload.com/v1/data/ohlc/deribit/futures/available/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1ohlc~1deribit~1futures~1available~1/get) |
| `Catalog.OhlcDeribitFuturesFundingAvailable`<br>`Catalog.OhlcDeribitFuturesFundingAvailableWithResponse` | `GET /v1/data/ohlc/deribit/futures/funding/available/` | [production](https://api.cryptodatadownload.com/v1/data/ohlc/deribit/futures/funding/available/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1ohlc~1deribit~1futures~1funding~1available~1/get) |
| `Catalog.OhlcDeribitOptionsAvailable`<br>`Catalog.OhlcDeribitOptionsAvailableWithResponse` | `GET /v1/data/ohlc/deribit/options/available/` | [production](https://api.cryptodatadownload.com/v1/data/ohlc/deribit/options/available/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1ohlc~1deribit~1options~1available~1/get) |
| `Catalog.SummaryBinanceFuturesFundingAvailable`<br>`Catalog.SummaryBinanceFuturesFundingAvailableWithResponse` | `GET /v1/data/summary/binance/futures/funding/available/` | [production](https://api.cryptodatadownload.com/v1/data/summary/binance/futures/funding/available/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1binance~1futures~1funding~1available~1/get) |
| `Catalog.SummaryBinanceSpotTransactionalAvailable`<br>`Catalog.SummaryBinanceSpotTransactionalAvailableWithResponse` | `GET /v1/data/summary/binance/spot/transactional/available/` | [production](https://api.cryptodatadownload.com/v1/data/summary/binance/spot/transactional/available/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1binance~1spot~1transactional~1available~1/get) |
| `Catalog.SummaryDeribitOptionsGreeksAvailable`<br>`Catalog.SummaryDeribitOptionsGreeksAvailableWithResponse` | `GET /v1/data/summary/deribit/options/greeks/available/` | [production](https://api.cryptodatadownload.com/v1/data/summary/deribit/options/greeks/available/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1deribit~1options~1greeks~1available~1/get) |
| `Catalog.TransactionsDeribitOptionsAvailable`<br>`Catalog.TransactionsDeribitOptionsAvailableWithResponse` | `GET /v1/data/transactions/deribit/options/available/` | [production](https://api.cryptodatadownload.com/v1/data/transactions/deribit/options/available/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1transactions~1deribit~1options~1available~1/get) |
| `MLMarket.FuturesBasisTerm`<br>`MLMarket.FuturesBasisTermWithResponse` | `GET /v1/data/features/futures/basis-term/` | [production](https://api.cryptodatadownload.com/v1/data/features/futures/basis-term/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1futures~1basis-term~1/get) |
| `MLMarket.MacroYieldSpreads`<br>`MLMarket.MacroYieldSpreadsWithResponse` | `GET /v1/data/features/macro/yield-spreads/` | [production](https://api.cryptodatadownload.com/v1/data/features/macro/yield-spreads/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1macro~1yield-spreads~1/get) |
| `MLMarket.SpotReturns`<br>`MLMarket.SpotReturnsWithResponse` | `GET /v1/data/features/spot/returns/` | [production](https://api.cryptodatadownload.com/v1/data/features/spot/returns/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1spot~1returns~1/get) |
| `MLMarket.SpotVwap`<br>`MLMarket.SpotVwapWithResponse` | `GET /v1/data/features/spot/vwap/` | [production](https://api.cryptodatadownload.com/v1/data/features/spot/vwap/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1spot~1vwap~1/get) |
| `MLFutures.FundingRegime`<br>`MLFutures.FundingRegimeWithResponse` | `GET /v1/data/features/futures/funding-regime/` | [production](https://api.cryptodatadownload.com/v1/data/features/futures/funding-regime/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1futures~1funding-regime~1/get) |
| `MLFutures.OiChange`<br>`MLFutures.OiChangeWithResponse` | `GET /v1/data/features/futures/oi-change/` | [production](https://api.cryptodatadownload.com/v1/data/features/futures/oi-change/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1futures~1oi-change~1/get) |
| `MLFutures.PerpBasisTs`<br>`MLFutures.PerpBasisTsWithResponse` | `GET /v1/data/features/futures/perp-basis-ts/` | [production](https://api.cryptodatadownload.com/v1/data/features/futures/perp-basis-ts/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1futures~1perp-basis-ts~1/get) |
| `MLMacro.CotPositioning`<br>`MLMacro.CotPositioningWithResponse` | `GET /v1/data/features/macro/cot-positioning/` | [production](https://api.cryptodatadownload.com/v1/data/features/macro/cot-positioning/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1macro~1cot-positioning~1/get) |
| `MLMacro.RealRateRegime`<br>`MLMacro.RealRateRegimeWithResponse` | `GET /v1/data/features/macro/real-rate-regime/` | [production](https://api.cryptodatadownload.com/v1/data/features/macro/real-rate-regime/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1macro~1real-rate-regime~1/get) |
| `MLMacro.YieldCurvePca`<br>`MLMacro.YieldCurvePcaWithResponse` | `GET /v1/data/features/macro/yield-curve-pca/` | [production](https://api.cryptodatadownload.com/v1/data/features/macro/yield-curve-pca/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1macro~1yield-curve-pca~1/get) |
| `MLCrossSectional.BetaBtc`<br>`MLCrossSectional.BetaBtcWithResponse` | `GET /v1/data/features/ml/beta-btc/` | [production](https://api.cryptodatadownload.com/v1/data/features/ml/beta-btc/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1ml~1beta-btc~1/get) |
| `MLCrossSectional.Momentum`<br>`MLCrossSectional.MomentumWithResponse` | `GET /v1/data/features/ml/momentum/` | [production](https://api.cryptodatadownload.com/v1/data/features/ml/momentum/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1ml~1momentum~1/get) |
| `MLCrossSectional.ReturnRank`<br>`MLCrossSectional.ReturnRankWithResponse` | `GET /v1/data/features/ml/return-rank/` | [production](https://api.cryptodatadownload.com/v1/data/features/ml/return-rank/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1ml~1return-rank~1/get) |
| `MLCrossSectional.Reversal`<br>`MLCrossSectional.ReversalWithResponse` | `GET /v1/data/features/ml/reversal/` | [production](https://api.cryptodatadownload.com/v1/data/features/ml/reversal/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1ml~1reversal~1/get) |
| `MLCrossSectional.VolRank`<br>`MLCrossSectional.VolRankWithResponse` | `GET /v1/data/features/ml/vol-rank/` | [production](https://api.cryptodatadownload.com/v1/data/features/ml/vol-rank/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1ml~1vol-rank~1/get) |
| `MLMicrostructure.LargeTradeShare`<br>`MLMicrostructure.LargeTradeShareWithResponse` | `GET /v1/data/features/ml/large-trade-share/` | [production](https://api.cryptodatadownload.com/v1/data/features/ml/large-trade-share/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1ml~1large-trade-share~1/get) |
| `MLMicrostructure.OrderFlowImbalance`<br>`MLMicrostructure.OrderFlowImbalanceWithResponse` | `GET /v1/data/features/ml/order-flow-imbalance/` | [production](https://api.cryptodatadownload.com/v1/data/features/ml/order-flow-imbalance/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1ml~1order-flow-imbalance~1/get) |
| `MLMicrostructure.VolumeProfile`<br>`MLMicrostructure.VolumeProfileWithResponse` | `GET /v1/data/features/ml/volume-profile/` | [production](https://api.cryptodatadownload.com/v1/data/features/ml/volume-profile/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1ml~1volume-profile~1/get) |
| `MLMicrostructure.XexchangeSpread`<br>`MLMicrostructure.XexchangeSpreadWithResponse` | `GET /v1/data/features/ml/xexchange-spread/` | [production](https://api.cryptodatadownload.com/v1/data/features/ml/xexchange-spread/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1ml~1xexchange-spread~1/get) |
| `OnChain.FeaturesOnchainAddressVelocity`<br>`OnChain.FeaturesOnchainAddressVelocityWithResponse` | `GET /v1/data/features/onchain/address-velocity/` | [production](https://api.cryptodatadownload.com/v1/data/features/onchain/address-velocity/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1onchain~1address-velocity~1/get) |
| `OnChain.FeaturesOnchainContractGrowth`<br>`OnChain.FeaturesOnchainContractGrowthWithResponse` | `GET /v1/data/features/onchain/contract-growth/` | [production](https://api.cryptodatadownload.com/v1/data/features/onchain/contract-growth/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1onchain~1contract-growth~1/get) |
| `OnChain.FeaturesOnchainNetworkCongestion`<br>`OnChain.FeaturesOnchainNetworkCongestionWithResponse` | `GET /v1/data/features/onchain/network-congestion/` | [production](https://api.cryptodatadownload.com/v1/data/features/onchain/network-congestion/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1onchain~1network-congestion~1/get) |
| `OnChain.FeaturesOnchainNvt`<br>`OnChain.FeaturesOnchainNvtWithResponse` | `GET /v1/data/features/onchain/nvt/` | [production](https://api.cryptodatadownload.com/v1/data/features/onchain/nvt/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1onchain~1nvt~1/get) |
| `OnChain.FeaturesOnchainSupplyVelocity`<br>`OnChain.FeaturesOnchainSupplyVelocityWithResponse` | `GET /v1/data/features/onchain/supply-velocity/` | [production](https://api.cryptodatadownload.com/v1/data/features/onchain/supply-velocity/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1onchain~1supply-velocity~1/get) |
| `OnChain.FeaturesOnchainWhale`<br>`OnChain.FeaturesOnchainWhaleWithResponse` | `GET /v1/data/features/onchain/whale/` | [production](https://api.cryptodatadownload.com/v1/data/features/onchain/whale/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1onchain~1whale~1/get) |
| `OnChain.OnchainBtcMempool`<br>`OnChain.OnchainBtcMempoolWithResponse` | `GET /v1/data/onchain/btc/mempool/` | [production](https://api.cryptodatadownload.com/v1/data/onchain/btc/mempool/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1onchain~1btc~1mempool~1/get) |
| `OnChain.OnchainEthExchangeFlows`<br>`OnChain.OnchainEthExchangeFlowsWithResponse` | `GET /v1/data/onchain/eth/exchange-flows/` | [production](https://api.cryptodatadownload.com/v1/data/onchain/eth/exchange-flows/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1onchain~1eth~1exchange-flows~1/get) |
| `OnChain.OnchainEthExchangeReserves`<br>`OnChain.OnchainEthExchangeReservesWithResponse` | `GET /v1/data/onchain/eth/exchange-reserves/` | [production](https://api.cryptodatadownload.com/v1/data/onchain/eth/exchange-reserves/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1onchain~1eth~1exchange-reserves~1/get) |
| `OnChain.OnchainEthGas`<br>`OnChain.OnchainEthGasWithResponse` | `GET /v1/data/onchain/eth/gas/` | [production](https://api.cryptodatadownload.com/v1/data/onchain/eth/gas/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1onchain~1eth~1gas~1/get) |
| `OnChain.OnchainEthTransactions`<br>`OnChain.OnchainEthTransactionsWithResponse` | `GET /v1/data/onchain/eth/transactions/` | [production](https://api.cryptodatadownload.com/v1/data/onchain/eth/transactions/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1onchain~1eth~1transactions~1/get) |
| `OnChain.SummaryBitcoinMarketcap`<br>`OnChain.SummaryBitcoinMarketcapWithResponse`<br>`OnChain.SummaryBitcoinMarketcapRaw`<br>`OnChain.SummaryBitcoinMarketcapStream` | `GET /v1/data/summary/bitcoin/marketcap/` | [production](https://api.cryptodatadownload.com/v1/data/summary/bitcoin/marketcap/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1bitcoin~1marketcap~1/get) |
| `OnChain.SummaryBlockchainBchain`<br>`OnChain.SummaryBlockchainBchainWithResponse`<br>`OnChain.SummaryBlockchainBchainRaw`<br>`OnChain.SummaryBlockchainBchainStream` | `GET /v1/data/summary/blockchain/bchain/` | [production](https://api.cryptodatadownload.com/v1/data/summary/blockchain/bchain/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1blockchain~1bchain~1/get) |
| `OnChain.SummaryBlockchainBlocks`<br>`OnChain.SummaryBlockchainBlocksWithResponse`<br>`OnChain.SummaryBlockchainBlocksRaw`<br>`OnChain.SummaryBlockchainBlocksStream` | `GET /v1/data/summary/blockchain/blocks/` | [production](https://api.cryptodatadownload.com/v1/data/summary/blockchain/blocks/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1blockchain~1blocks~1/get) |
| `OnChain.SummaryBlockchainEthContracts`<br>`OnChain.SummaryBlockchainEthContractsWithResponse`<br>`OnChain.SummaryBlockchainEthContractsRaw`<br>`OnChain.SummaryBlockchainEthContractsStream` | `GET /v1/data/summary/blockchain/eth/contracts/` | [production](https://api.cryptodatadownload.com/v1/data/summary/blockchain/eth/contracts/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1blockchain~1eth~1contracts~1/get) |
| `OnChain.SummaryBlockchainTransactions`<br>`OnChain.SummaryBlockchainTransactionsWithResponse`<br>`OnChain.SummaryBlockchainTransactionsRaw`<br>`OnChain.SummaryBlockchainTransactionsStream` | `GET /v1/data/summary/blockchain/transactions/` | [production](https://api.cryptodatadownload.com/v1/data/summary/blockchain/transactions/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1blockchain~1transactions~1/get) |
| `OnChain.SummaryBlockchainTransactionsCoinbase`<br>`OnChain.SummaryBlockchainTransactionsCoinbaseWithResponse`<br>`OnChain.SummaryBlockchainTransactionsCoinbaseRaw`<br>`OnChain.SummaryBlockchainTransactionsCoinbaseStream` | `GET /v1/data/summary/blockchain/transactions/coinbase/` | [production](https://api.cryptodatadownload.com/v1/data/summary/blockchain/transactions/coinbase/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1blockchain~1transactions~1coinbase~1/get) |
| `OnChain.SummaryBlockchainTransactionsLargest`<br>`OnChain.SummaryBlockchainTransactionsLargestWithResponse`<br>`OnChain.SummaryBlockchainTransactionsLargestRaw`<br>`OnChain.SummaryBlockchainTransactionsLargestStream` | `GET /v1/data/summary/blockchain/transactions/largest/` | [production](https://api.cryptodatadownload.com/v1/data/summary/blockchain/transactions/largest/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1blockchain~1transactions~1largest~1/get) |
| `MLOptions.ImpliedMove`<br>`MLOptions.ImpliedMoveWithResponse` | `GET /v1/data/features/options/implied-move/` | [production](https://api.cryptodatadownload.com/v1/data/features/options/implied-move/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1options~1implied-move~1/get) |
| `MLOptions.OiConcentration`<br>`MLOptions.OiConcentrationWithResponse` | `GET /v1/data/features/options/oi-concentration/` | [production](https://api.cryptodatadownload.com/v1/data/features/options/oi-concentration/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1options~1oi-concentration~1/get) |
| `MLOptions.SkewSignal`<br>`MLOptions.SkewSignalWithResponse` | `GET /v1/data/features/options/skew-signal/` | [production](https://api.cryptodatadownload.com/v1/data/features/options/skew-signal/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1options~1skew-signal~1/get) |
| `MLOptions.SmileCurvature`<br>`MLOptions.SmileCurvatureWithResponse` | `GET /v1/data/features/options/smile-curvature/` | [production](https://api.cryptodatadownload.com/v1/data/features/options/smile-curvature/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1options~1smile-curvature~1/get) |
| `MLRegime.AnomalyScore`<br>`MLRegime.AnomalyScoreWithResponse` | `GET /v1/data/features/regime/anomaly-score/` | [production](https://api.cryptodatadownload.com/v1/data/features/regime/anomaly-score/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1regime~1anomaly-score~1/get) |
| `MLRegime.ClusterMembership`<br>`MLRegime.ClusterMembershipWithResponse` | `GET /v1/data/features/regime/cluster-membership/` | [production](https://api.cryptodatadownload.com/v1/data/features/regime/cluster-membership/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1regime~1cluster-membership~1/get) |
| `MLRegime.RegimeState`<br>`MLRegime.RegimeStateWithResponse` | `GET /v1/data/features/regime/regime-state/` | [production](https://api.cryptodatadownload.com/v1/data/features/regime/regime-state/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1regime~1regime-state~1/get) |
| `Funding.FundingAligned`<br>`Funding.FundingAlignedWithResponse` | `GET /v1/data/funding/aligned/` | [production](https://api.cryptodatadownload.com/v1/data/funding/aligned/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1funding~1aligned~1/get) |
| `Funding.FundingBinance`<br>`Funding.FundingBinanceWithResponse` | `GET /v1/data/funding/binance/` | [production](https://api.cryptodatadownload.com/v1/data/funding/binance/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1funding~1binance~1/get) |
| `Funding.FundingBybit`<br>`Funding.FundingBybitWithResponse` | `GET /v1/data/funding/bybit/` | [production](https://api.cryptodatadownload.com/v1/data/funding/bybit/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1funding~1bybit~1/get) |
| `Funding.FundingOkx`<br>`Funding.FundingOkxWithResponse` | `GET /v1/data/funding/okx/` | [production](https://api.cryptodatadownload.com/v1/data/funding/okx/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1funding~1okx~1/get) |
| `Funding.SummaryBinanceFuturesFunding`<br>`Funding.SummaryBinanceFuturesFundingWithResponse`<br>`Funding.SummaryBinanceFuturesFundingRaw`<br>`Funding.SummaryBinanceFuturesFundingStream` | `GET /v1/data/summary/binance/futures/funding/` | [production](https://api.cryptodatadownload.com/v1/data/summary/binance/futures/funding/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1binance~1futures~1funding~1/get) |
| `Funding.SummaryFundingComparison`<br>`Funding.SummaryFundingComparisonWithResponse` | `GET /v1/data/summary/funding/comparison/` | [production](https://api.cryptodatadownload.com/v1/data/summary/funding/comparison/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1funding~1comparison~1/get) |
| `BinanceOHLC.FuturesCm`<br>`BinanceOHLC.FuturesCmWithResponse`<br>`BinanceOHLC.FuturesCmRaw`<br>`BinanceOHLC.FuturesCmStream` | `GET /v1/data/ohlc/binance/futures/cm/` | [production](https://api.cryptodatadownload.com/v1/data/ohlc/binance/futures/cm/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1ohlc~1binance~1futures~1cm~1/get) |
| `BinanceOHLC.FuturesUm`<br>`BinanceOHLC.FuturesUmWithResponse`<br>`BinanceOHLC.FuturesUmRaw`<br>`BinanceOHLC.FuturesUmStream` | `GET /v1/data/ohlc/binance/futures/um/` | [production](https://api.cryptodatadownload.com/v1/data/ohlc/binance/futures/um/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1ohlc~1binance~1futures~1um~1/get) |
| `BinanceOHLC.Options`<br>`BinanceOHLC.OptionsWithResponse`<br>`BinanceOHLC.OptionsRaw`<br>`BinanceOHLC.OptionsStream` | `GET /v1/data/ohlc/binance/options/` | [production](https://api.cryptodatadownload.com/v1/data/ohlc/binance/options/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1ohlc~1binance~1options~1/get) |
| `BinanceOHLC.Spot`<br>`BinanceOHLC.SpotWithResponse`<br>`BinanceOHLC.SpotRaw`<br>`BinanceOHLC.SpotStream` | `GET /v1/data/ohlc/binance/spot/` | [production](https://api.cryptodatadownload.com/v1/data/ohlc/binance/spot/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1ohlc~1binance~1spot~1/get) |
| `BinanceOHLC.Tick`<br>`BinanceOHLC.TickWithResponse`<br>`BinanceOHLC.TickRaw`<br>`BinanceOHLC.TickStream` | `GET /v1/data/ohlc/binance/tick/` | [production](https://api.cryptodatadownload.com/v1/data/ohlc/binance/tick/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1ohlc~1binance~1tick~1/get) |
| `DeribitOHLC.Futures`<br>`DeribitOHLC.FuturesWithResponse`<br>`DeribitOHLC.FuturesRaw`<br>`DeribitOHLC.FuturesStream` | `GET /v1/data/ohlc/deribit/futures/` | [production](https://api.cryptodatadownload.com/v1/data/ohlc/deribit/futures/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1ohlc~1deribit~1futures~1/get) |
| `DeribitOHLC.FuturesFunding`<br>`DeribitOHLC.FuturesFundingWithResponse`<br>`DeribitOHLC.FuturesFundingRaw`<br>`DeribitOHLC.FuturesFundingStream` | `GET /v1/data/ohlc/deribit/futures/funding/` | [production](https://api.cryptodatadownload.com/v1/data/ohlc/deribit/futures/funding/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1ohlc~1deribit~1futures~1funding~1/get) |
| `DeribitOHLC.Options`<br>`DeribitOHLC.OptionsWithResponse`<br>`DeribitOHLC.OptionsRaw`<br>`DeribitOHLC.OptionsStream` | `GET /v1/data/ohlc/deribit/options/` | [production](https://api.cryptodatadownload.com/v1/data/ohlc/deribit/options/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1ohlc~1deribit~1options~1/get) |
| `DeribitOHLC.Volatility`<br>`DeribitOHLC.VolatilityWithResponse`<br>`DeribitOHLC.VolatilityRaw`<br>`DeribitOHLC.VolatilityStream` | `GET /v1/data/ohlc/deribit/volatility/` | [production](https://api.cryptodatadownload.com/v1/data/ohlc/deribit/volatility/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1ohlc~1deribit~1volatility~1/get) |
| `OptionsAnalytics.OptionsStrikeHeatmap`<br>`OptionsAnalytics.OptionsStrikeHeatmapWithResponse` | `GET /v1/data/options/strike-heatmap/` | [production](https://api.cryptodatadownload.com/v1/data/options/strike-heatmap/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1options~1strike-heatmap~1/get) |
| `OptionsAnalytics.SummaryDeribitOptionsGreeksHigherOrder`<br>`OptionsAnalytics.SummaryDeribitOptionsGreeksHigherOrderWithResponse` | `GET /v1/data/summary/deribit/options/greeks/higher-order/` | [production](https://api.cryptodatadownload.com/v1/data/summary/deribit/options/greeks/higher-order/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1deribit~1options~1greeks~1higher-order~1/get) |
| `OptionsAnalytics.SummaryDeribitOptionsPinCandidate`<br>`OptionsAnalytics.SummaryDeribitOptionsPinCandidateWithResponse` | `GET /v1/data/summary/deribit/options/pin-candidate/` | [production](https://api.cryptodatadownload.com/v1/data/summary/deribit/options/pin-candidate/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1deribit~1options~1pin-candidate~1/get) |
| `OptionsAnalytics.SummaryDeribitOptionsTerm`<br>`OptionsAnalytics.SummaryDeribitOptionsTermWithResponse` | `GET /v1/data/summary/deribit/options/term/` | [production](https://api.cryptodatadownload.com/v1/data/summary/deribit/options/term/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1deribit~1options~1term~1/get) |
| `OptionsAnalytics.SummaryOptionsVwapPerStrike`<br>`OptionsAnalytics.SummaryOptionsVwapPerStrikeWithResponse` | `GET /v1/data/summary/options/vwap-per-strike/` | [production](https://api.cryptodatadownload.com/v1/data/summary/options/vwap-per-strike/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1options~1vwap-per-strike~1/get) |
| `OptionsAnalytics.VolatilityCone`<br>`OptionsAnalytics.VolatilityConeWithResponse` | `GET /v1/data/volatility/cone/` | [production](https://api.cryptodatadownload.com/v1/data/volatility/cone/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1volatility~1cone~1/get) |
| `OptionsAnalytics.VolatilityIvRvSpread`<br>`OptionsAnalytics.VolatilityIvRvSpreadWithResponse` | `GET /v1/data/volatility/iv-rv-spread/` | [production](https://api.cryptodatadownload.com/v1/data/volatility/iv-rv-spread/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1volatility~1iv-rv-spread~1/get) |
| `OptionsAnalytics.VolatilitySurface`<br>`OptionsAnalytics.VolatilitySurfaceWithResponse` | `GET /v1/data/volatility/surface/` | [production](https://api.cryptodatadownload.com/v1/data/volatility/surface/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1volatility~1surface~1/get) |
| `Risk.RiskCorrelations`<br>`Risk.RiskCorrelationsWithResponse`<br>`Risk.RiskCorrelationsRaw`<br>`Risk.RiskCorrelationsStream` | `GET /v1/data/risk/correlations/` | [production](https://api.cryptodatadownload.com/v1/data/risk/correlations/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1risk~1correlations~1/get) |
| `Risk.RiskCorrelationsRolling`<br>`Risk.RiskCorrelationsRollingWithResponse` | `GET /v1/data/risk/correlations/rolling/` | [production](https://api.cryptodatadownload.com/v1/data/risk/correlations/rolling/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1risk~1correlations~1rolling~1/get) |
| `Risk.RiskDrawdown`<br>`Risk.RiskDrawdownWithResponse` | `GET /v1/data/risk/drawdown/` | [production](https://api.cryptodatadownload.com/v1/data/risk/drawdown/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1risk~1drawdown~1/get) |
| `Risk.RiskDrawdownUnderwater`<br>`Risk.RiskDrawdownUnderwaterWithResponse` | `GET /v1/data/risk/drawdown/underwater/` | [production](https://api.cryptodatadownload.com/v1/data/risk/drawdown/underwater/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1risk~1drawdown~1underwater~1/get) |
| `Risk.RiskRegime`<br>`Risk.RiskRegimeWithResponse` | `GET /v1/data/risk/regime/` | [production](https://api.cryptodatadownload.com/v1/data/risk/regime/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1risk~1regime~1/get) |
| `Risk.RiskVarByMethod`<br>`Risk.RiskVarByMethodWithResponse` | `GET /v1/data/risk/var/by-method/` | [production](https://api.cryptodatadownload.com/v1/data/risk/var/by-method/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1risk~1var~1by-method~1/get) |
| `Risk.RiskVarPortfolio`<br>`Risk.RiskVarPortfolioWithResponse` | `GET /v1/data/risk/var/portfolio/` | [production](https://api.cryptodatadownload.com/v1/data/risk/var/portfolio/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1risk~1var~1portfolio~1/get) |
| `Risk.RiskVarStandalone`<br>`Risk.RiskVarStandaloneWithResponse`<br>`Risk.RiskVarStandaloneRaw`<br>`Risk.RiskVarStandaloneStream` | `GET /v1/data/risk/var/standalone/` | [production](https://api.cryptodatadownload.com/v1/data/risk/var/standalone/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1risk~1var~1standalone~1/get) |
| `Risk.VolatilityRealized`<br>`Risk.VolatilityRealizedWithResponse` | `GET /v1/data/volatility/realized/` | [production](https://api.cryptodatadownload.com/v1/data/volatility/realized/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1volatility~1realized~1/get) |
| `BinanceSummary.Basis`<br>`BinanceSummary.BasisWithResponse`<br>`BinanceSummary.BasisRaw`<br>`BinanceSummary.BasisStream` | `GET /v1/data/summary/binance/basis/` | [production](https://api.cryptodatadownload.com/v1/data/summary/binance/basis/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1binance~1basis~1/get) |
| `BinanceSummary.FuturesMetrics`<br>`BinanceSummary.FuturesMetricsWithResponse`<br>`BinanceSummary.FuturesMetricsRaw`<br>`BinanceSummary.FuturesMetricsStream` | `GET /v1/data/summary/binance/futures/metrics/` | [production](https://api.cryptodatadownload.com/v1/data/summary/binance/futures/metrics/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1binance~1futures~1metrics~1/get) |
| `BinanceSummary.OptionsOi`<br>`BinanceSummary.OptionsOiWithResponse`<br>`BinanceSummary.OptionsOiRaw`<br>`BinanceSummary.OptionsOiStream` | `GET /v1/data/summary/binance/options/oi/` | [production](https://api.cryptodatadownload.com/v1/data/summary/binance/options/oi/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1binance~1options~1oi~1/get) |
| `BinanceSummary.OptionsOiInstrument`<br>`BinanceSummary.OptionsOiInstrumentWithResponse`<br>`BinanceSummary.OptionsOiInstrumentRaw`<br>`BinanceSummary.OptionsOiInstrumentStream` | `GET /v1/data/summary/binance/options/oi/instrument/` | [production](https://api.cryptodatadownload.com/v1/data/summary/binance/options/oi/instrument/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1binance~1options~1oi~1instrument~1/get) |
| `BinanceSummary.OptionsOiMaturity`<br>`BinanceSummary.OptionsOiMaturityWithResponse`<br>`BinanceSummary.OptionsOiMaturityRaw`<br>`BinanceSummary.OptionsOiMaturityStream` | `GET /v1/data/summary/binance/options/oi/maturity/` | [production](https://api.cryptodatadownload.com/v1/data/summary/binance/options/oi/maturity/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1binance~1options~1oi~1maturity~1/get) |
| `BinanceSummary.SpotFlow`<br>`BinanceSummary.SpotFlowWithResponse`<br>`BinanceSummary.SpotFlowRaw`<br>`BinanceSummary.SpotFlowStream` | `GET /v1/data/summary/binance/spot/flow/` | [production](https://api.cryptodatadownload.com/v1/data/summary/binance/spot/flow/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1binance~1spot~1flow~1/get) |
| `BinanceSummary.SpotLargeTrades`<br>`BinanceSummary.SpotLargeTradesWithResponse`<br>`BinanceSummary.SpotLargeTradesRaw`<br>`BinanceSummary.SpotLargeTradesStream` | `GET /v1/data/summary/binance/spot/large-trades/` | [production](https://api.cryptodatadownload.com/v1/data/summary/binance/spot/large-trades/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1binance~1spot~1large-trades~1/get) |
| `BinanceSummary.SpotTransactional`<br>`BinanceSummary.SpotTransactionalWithResponse`<br>`BinanceSummary.SpotTransactionalRaw`<br>`BinanceSummary.SpotTransactionalStream` | `GET /v1/data/summary/binance/spot/transactional/` | [production](https://api.cryptodatadownload.com/v1/data/summary/binance/spot/transactional/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1binance~1spot~1transactional~1/get) |
| `BinanceSummary.SpotTransactionalDaily`<br>`BinanceSummary.SpotTransactionalDailyWithResponse`<br>`BinanceSummary.SpotTransactionalDailyRaw`<br>`BinanceSummary.SpotTransactionalDailyStream` | `GET /v1/data/summary/binance/spot/transactional/daily/` | [production](https://api.cryptodatadownload.com/v1/data/summary/binance/spot/transactional/daily/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1binance~1spot~1transactional~1daily~1/get) |
| `BinanceSummary.SpotTransactionalHourly`<br>`BinanceSummary.SpotTransactionalHourlyWithResponse`<br>`BinanceSummary.SpotTransactionalHourlyRaw`<br>`BinanceSummary.SpotTransactionalHourlyStream` | `GET /v1/data/summary/binance/spot/transactional/hourly/` | [production](https://api.cryptodatadownload.com/v1/data/summary/binance/spot/transactional/hourly/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1binance~1spot~1transactional~1hourly~1/get) |
| `BinanceSummary.SpotVolumeProfile`<br>`BinanceSummary.SpotVolumeProfileWithResponse`<br>`BinanceSummary.SpotVolumeProfileRaw`<br>`BinanceSummary.SpotVolumeProfileStream` | `GET /v1/data/summary/binance/spot/volume-profile/` | [production](https://api.cryptodatadownload.com/v1/data/summary/binance/spot/volume-profile/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1binance~1spot~1volume-profile~1/get) |
| `DeribitSummary.SummaryDeribitOptionsFlow`<br>`DeribitSummary.SummaryDeribitOptionsFlowWithResponse`<br>`DeribitSummary.SummaryDeribitOptionsFlowRaw`<br>`DeribitSummary.SummaryDeribitOptionsFlowStream` | `GET /v1/data/summary/deribit/options/flow/` | [production](https://api.cryptodatadownload.com/v1/data/summary/deribit/options/flow/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1deribit~1options~1flow~1/get) |
| `DeribitSummary.SummaryDeribitOptionsGreeks`<br>`DeribitSummary.SummaryDeribitOptionsGreeksWithResponse`<br>`DeribitSummary.SummaryDeribitOptionsGreeksRaw`<br>`DeribitSummary.SummaryDeribitOptionsGreeksStream` | `GET /v1/data/summary/deribit/options/greeks/` | [production](https://api.cryptodatadownload.com/v1/data/summary/deribit/options/greeks/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1deribit~1options~1greeks~1/get) |
| `DeribitSummary.SummaryDeribitOptionsGreeksMaturities`<br>`DeribitSummary.SummaryDeribitOptionsGreeksMaturitiesWithResponse`<br>`DeribitSummary.SummaryDeribitOptionsGreeksMaturitiesRaw`<br>`DeribitSummary.SummaryDeribitOptionsGreeksMaturitiesStream` | `GET /v1/data/summary/deribit/options/greeks/maturities/` | [production](https://api.cryptodatadownload.com/v1/data/summary/deribit/options/greeks/maturities/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1deribit~1options~1greeks~1maturities~1/get) |
| `DeribitSummary.SummaryDeribitOptionsLargeTrades`<br>`DeribitSummary.SummaryDeribitOptionsLargeTradesWithResponse`<br>`DeribitSummary.SummaryDeribitOptionsLargeTradesRaw`<br>`DeribitSummary.SummaryDeribitOptionsLargeTradesStream` | `GET /v1/data/summary/deribit/options/large-trades/` | [production](https://api.cryptodatadownload.com/v1/data/summary/deribit/options/large-trades/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1deribit~1options~1large-trades~1/get) |
| `DeribitSummary.SummaryDeribitOptionsRatios`<br>`DeribitSummary.SummaryDeribitOptionsRatiosWithResponse`<br>`DeribitSummary.SummaryDeribitOptionsRatiosRaw`<br>`DeribitSummary.SummaryDeribitOptionsRatiosStream` | `GET /v1/data/summary/deribit/options/ratios/` | [production](https://api.cryptodatadownload.com/v1/data/summary/deribit/options/ratios/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1deribit~1options~1ratios~1/get) |
| `DeribitSummary.SummaryDeribitOptionsTransactionsLargest`<br>`DeribitSummary.SummaryDeribitOptionsTransactionsLargestWithResponse`<br>`DeribitSummary.SummaryDeribitOptionsTransactionsLargestRaw`<br>`DeribitSummary.SummaryDeribitOptionsTransactionsLargestStream` | `GET /v1/data/summary/deribit/options/transactions/largest/` | [production](https://api.cryptodatadownload.com/v1/data/summary/deribit/options/transactions/largest/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1deribit~1options~1transactions~1largest~1/get) |
| `DeribitSummary.SummaryDeribitOptionsVolume`<br>`DeribitSummary.SummaryDeribitOptionsVolumeWithResponse`<br>`DeribitSummary.SummaryDeribitOptionsVolumeRaw`<br>`DeribitSummary.SummaryDeribitOptionsVolumeStream` | `GET /v1/data/summary/deribit/options/volume/` | [production](https://api.cryptodatadownload.com/v1/data/summary/deribit/options/volume/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1deribit~1options~1volume~1/get) |
| `DeribitSummary.TransactionsDeribitOptions`<br>`DeribitSummary.TransactionsDeribitOptionsWithResponse`<br>`DeribitSummary.TransactionsDeribitOptionsRaw`<br>`DeribitSummary.TransactionsDeribitOptionsStream` | `GET /v1/data/transactions/deribit/options/` | [production](https://api.cryptodatadownload.com/v1/data/transactions/deribit/options/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1transactions~1deribit~1options~1/get) |
| `Technicals.TechnicalsTA`<br>`Technicals.TechnicalsTAWithResponse`<br>`Technicals.TechnicalsTARaw`<br>`Technicals.TechnicalsTAStream` | `GET /v1/data/technicals/TA/` | [production](https://api.cryptodatadownload.com/v1/data/technicals/TA/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1technicals~1TA~1/get) |
| `OptionsPricing.PricingGreeks`<br>`OptionsPricing.PricingGreeksWithResponse` | `GET /v1/risk/options/pricing/greeks/` | [production](https://api.cryptodatadownload.com/v1/risk/options/pricing/greeks/) | [Swagger UI](https://api.cryptodatadownload.com/) · [OpenAPI record](https://api.cryptodatadownload.com/?format=openapi#/paths/~1risk~1options~1pricing~1greeks~1/get) |

</details>
<!-- END GENERATED ENDPOINT COVERAGE -->

### A few details about the API contract

- Coverage is based on the **116 operations** in `testdata/openapi.json`.
  The marketing page's endpoint count may differ.
- Each path keeps its trailing slash. The encoder preserves the operation's
  exact query names.
- CSV/XLSX support follows the operation's `return` parameter. The SDK does
  not expose the marketing quick-start's `format=csv` spelling.
- HTTP 401 and 429 are handled even where the OpenAPI response list omits them.
- The provider publishes one shared Swagger UI. The OpenAPI links point to
  the operation's JSON Pointer; a browser may display the entire schema.

## Regenerate the API reference

Consumers can use the package directly. Maintainers can regenerate the
services, manifest, contract fixtures and README endpoint index from the
versioned snapshots with PHP 8.0+ and Go:

```bash
php tools/generate.php
gofmt -w .
```

`testdata/openapi.json` and `testdata/endpoints.csv` were retrieved from the
official sources on **2026-10-25**. The OpenAPI snapshot defines operations
and query parameters; the CSV catalog supplies additional rate-limit notes.
Review fixture updates before regenerating. Generation replaces only the
marked endpoint index in this README, leaving the guide and examples intact.

The contract tests cover all 116 routes and compare the manifest's
operation IDs, parameters, defaults and enums with the saved schema. Normal
tests use local fixtures and do not refresh the contract from the network.

## License

[MIT](LICENSE). Copyright Igor Sazonov.
