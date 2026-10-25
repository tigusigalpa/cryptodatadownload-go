package cryptodatadownload

import (
	"context"
	"io"
	"math"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestInvalidArgumentsDoNotSendRequests(t *testing.T) {
	requests := 0
	c := transportTestClient(t, roundTripperFunc(func(*http.Request) (*http.Response, error) {
		requests++
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"result":[]}`))}, nil
	}))
	day := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		name string
		call func() error
	}{
		{"missing required enum", func() error {
			_, err := c.Aggregates.OnchainDashboard(context.Background(), OnchainDashboardParams{Date: day})
			return err
		}},
		{"empty optional enum", func() error {
			_, err := c.OptionsAnalytics.OptionsStrikeHeatmap(context.Background(), OptionsStrikeHeatmapParams{Underlying: "BTC", Metric: Ptr("")})
			return err
		}},
		{"invalid numeric enum", func() error {
			_, err := c.Risk.RiskCorrelationsRolling(context.Background(), RiskCorrelationsRollingParams{Window: 7, Date: day})
			return err
		}},
		{"invalid raw format", func() error {
			_, err := c.BinanceOHLC.SpotRaw(context.Background(), SpotParams{}, ReturnFormat("PDF"))
			return err
		}},
		{"negative limit", func() error {
			_, err := c.BinanceOHLC.Spot(context.Background(), SpotParams{Limit: Ptr(-1)})
			return err
		}},
		{"limit exceeds documented maximum", func() error {
			_, err := c.BinanceOHLC.Spot(context.Background(), SpotParams{Limit: Ptr(5001)})
			return err
		}},
		{"negative offset", func() error {
			_, err := c.BinanceSummary.SpotLargeTrades(context.Background(), SpotLargeTradesParams{Offset: Ptr(-1)})
			return err
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.call(); AsArgument(err) == nil {
				t.Fatalf("error = %v, want ArgumentError", err)
			}
		})
	}
	if requests != 0 {
		t.Errorf("sent %d requests with invalid arguments", requests)
	}
}

func TestRawDownloadPreservesBytes(t *testing.T) {
	want := "symbol,close\nBTCUSDT,123.45\n"
	c := transportTestClient(t, roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Query().Get("return") != "CSV" || r.URL.Query().Get("symbol") != "BTCUSDT" {
			t.Errorf("unexpected query: %v", r.URL.Query())
		}
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(want))}, nil
	}))
	got, err := c.BinanceOHLC.SpotRaw(context.Background(), SpotParams{Symbol: Ptr("BTCUSDT")}, FormatCSV)
	if err != nil || string(got) != want {
		t.Fatalf("raw download = %q, %v", got, err)
	}
}

func TestInvalidDatesAndNonFiniteNumbersDoNotSendRequests(t *testing.T) {
	requests := 0
	c := transportTestClient(t, roundTripperFunc(func(*http.Request) (*http.Response, error) {
		requests++
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"result":[]}`))}, nil
	}))
	for _, date := range []time.Time{{}, time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC)} {
		if _, err := c.Aggregates.FeaturesMlAll(context.Background(), FeaturesMlAllParams{Symbol: "BTCUSDT", Date: date}); AsArgument(err) == nil {
			t.Errorf("required date %v: error = %v", date, err)
		}
		if _, err := c.Breadth.N52wkHighsLows(context.Background(), N52wkHighsLowsParams{Date: Ptr(date)}); AsArgument(err) == nil {
			t.Errorf("optional date %v: error = %v", date, err)
		}
	}
	for _, value := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		if _, err := c.BinanceSummary.SpotLargeTrades(context.Background(), SpotLargeTradesParams{MinUSD: Ptr(value)}); AsArgument(err) == nil {
			t.Errorf("typed min_usd=%v: error = %v", value, err)
		}
		if _, err := c.BinanceSummary.SpotLargeTradesRaw(context.Background(), SpotLargeTradesParams{MinUSD: Ptr(value)}, FormatCSV); AsArgument(err) == nil {
			t.Errorf("raw min_usd=%v: error = %v", value, err)
		}
	}
	if requests != 0 {
		t.Errorf("sent %d requests with invalid arguments", requests)
	}
}

func TestOptionalParametersKeepZeroValuesAndNormalizeDates(t *testing.T) {
	date := time.Date(2026, 1, 2, 1, 0, 0, 0, time.FixedZone("UTC+3", 3*60*60))
	c := transportTestClient(t, roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		query := r.URL.Query()
		if query.Get("date") != "2026-01-01" || query.Get("offset") != "0" || query.Has("limit") {
			t.Errorf("optional query lost UTC conversion, explicit zero or omitted default: %v", query)
		}
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"result":[]}`))}, nil
	}))
	if _, err := c.DeribitOHLC.Options(context.Background(), DeribitOHLCServiceOptionsParams{Date: Ptr(date), Offset: Ptr(0)}); err != nil {
		t.Fatal(err)
	}
}

func TestParseDateUsesExactUTCFormat(t *testing.T) {
	date, err := ParseDate("2026-10-25")
	if err != nil || date.Location() != time.UTC || dateString(date) != "2026-10-25" {
		t.Fatalf("parsed date = %v, %v", date, err)
	}
	for _, value := range []string{"", "2026-1-2", "2026-02-30", "2026-10-25T00:00:00Z", "0000-01-01"} {
		if _, err := ParseDate(value); AsArgument(err) == nil {
			t.Errorf("date %q accepted", value)
		}
	}
}
