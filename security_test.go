package cryptodatadownload

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestRedirectsKeepCredentialsOnOriginalOrigin(t *testing.T) {
	for _, target := range []string{
		"http://api.example.com/redirected",
		"https://child.api.example.com/redirected",
		"https://other.example.com/redirected",
	} {
		t.Run(target, func(t *testing.T) {
			attempts := 0
			c := transportTestClient(t, roundTripperFunc(func(r *http.Request) (*http.Response, error) {
				attempts++
				if attempts == 1 {
					return &http.Response{StatusCode: 302, Header: http.Header{"Location": {target}}, Body: io.NopCloser(strings.NewReader("")), Request: r}, nil
				}
				return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"result":[]}`)), Request: r}, nil
			}), WithBaseURL("https://api.example.com/v1"))
			_, err := c.BinanceOHLC.Spot(context.Background(), SpotParams{})
			if apiErr := AsAPIError(err); apiErr == nil || apiErr.Status != 302 || attempts != 1 {
				t.Fatalf("error = %v, requests = %d; want blocked redirect", err, attempts)
			}
		})
	}
}

func TestSameOriginRedirectHonorsCustomPolicy(t *testing.T) {
	requests, redirects := 0, 0
	custom := &http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			requests++
			return &http.Response{StatusCode: 302, Header: http.Header{"Location": {"/redirected"}}, Body: io.NopCloser(strings.NewReader("")), Request: r}, nil
		}),
		CheckRedirect: func(*http.Request, []*http.Request) error {
			redirects++
			return http.ErrUseLastResponse
		},
	}
	c, err := NewClient("test-key", WithHTTPClient(custom), WithRequestsPerSecond(0))
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.BinanceOHLC.Spot(context.Background(), SpotParams{})
	if AsAPIError(err) == nil || redirects != 1 || requests != 1 {
		t.Fatalf("error = %v, redirects = %d, requests = %d", err, redirects, requests)
	}
	if err := custom.CheckRedirect(nil, nil); err != http.ErrUseLastResponse {
		t.Fatal("custom HTTP client was mutated")
	}
}

func TestSameOriginRedirectPreservesToken(t *testing.T) {
	requests := 0
	c := transportTestClient(t, roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		requests++
		if r.Header.Get("Authorization") != "Token test-key" {
			t.Error("same-origin request lost authentication")
		}
		if requests == 1 {
			if r.URL.Path != "/v1/data/ohlc/binance/spot/" {
				t.Errorf("unexpected joined path: %s", r.URL.Path)
			}
			return &http.Response{StatusCode: 302, Header: http.Header{"Location": {"/redirected"}}, Body: io.NopCloser(strings.NewReader("")), Request: r}, nil
		}
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"result":[]}`)), Request: r}, nil
	}), WithBaseURL("https://api.example.com/v1///"))
	if _, err := c.BinanceOHLC.Spot(context.Background(), SpotParams{}); err != nil || requests != 2 {
		t.Fatalf("error = %v, requests = %d; want successful same-origin redirect", err, requests)
	}
}

func TestErrorsRedactEchoedAPIKey(t *testing.T) {
	underlying := errors.New("transport failure containing test-key")
	c := transportTestClient(t, roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return nil, underlying
	}))
	_, err := c.BinanceOHLC.Spot(context.Background(), SpotParams{})
	if err == nil || strings.Contains(fmt.Sprint(err), "test-key") || !errors.Is(err, underlying) {
		t.Fatalf("transport error is not safely wrapped: %v", err)
	}
	c = transportTestClient(t, roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 401, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"detail":"invalid token test-key"}`))}, nil
	}))
	_, err = c.BinanceOHLC.Spot(context.Background(), SpotParams{})
	if err == nil || strings.Contains(fmt.Sprint(err), "test-key") || AsAPIError(err) == nil {
		t.Fatalf("API error leaked echoed key: %v", err)
	}
}

func TestRejectsUnsafeClientInput(t *testing.T) {
	for _, base := range []string{"https://user:secret@example.com/v1", "https://example.com/v1?", "https://example.com/v1#"} {
		if _, err := NewClient("test-key", WithBaseURL(base)); err == nil {
			t.Errorf("base URL %q was accepted", base)
		}
	}
	for _, key := range []string{"test\r\nkey", "test\x00key", "test\x7fkey"} {
		if _, err := NewClient(key); err == nil || strings.Contains(err.Error(), key) {
			t.Error("invalid header token was accepted or exposed in an error")
		}
	}
}

func TestQueryTokenAuthIsOptInAndRedacted(t *testing.T) {
	const key = "test+key/with?special=characters"
	c, err := NewClientWithOptions(WithToken(key), WithQueryTokenAuth(true), WithRequestsPerSecond(0), WithRetryMaxAttempts(1), WithHTTPClient(&http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			if r.URL.Query().Get("auth_token") != key || r.Header.Get("Authorization") != "" {
				t.Error("query-token mode was not applied explicitly")
			}
			return &http.Response{StatusCode: 401, Header: http.Header{"X-Echo": {key}}, Body: io.NopCloser(strings.NewReader("invalid token " + key))}, nil
		}),
	}))
	if err != nil {
		t.Fatal(err)
	}
	_, response, err := c.BinanceOHLC.SpotWithResponse(context.Background(), SpotParams{})
	apiErr := AsAPIError(err)
	if apiErr == nil || response == nil {
		t.Fatalf("missing API error or metadata: %v", err)
	}
	for _, value := range []string{err.Error(), string(apiErr.RawBody), response.URL, response.Headers.Get("X-Echo"), apiErr.Headers.Get("X-Echo")} {
		if strings.Contains(value, key) || strings.Contains(value, url.QueryEscape(key)) {
			t.Fatalf("query token leaked into diagnostics: %q", value)
		}
	}
	safeURL, err := url.Parse(response.URL)
	if err != nil || safeURL.Query().Get("auth_token") != "[REDACTED]" {
		t.Fatal("query token was not removed from the metadata URL")
	}
	c.cfg.httpClient.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return nil, errors.New("failed request " + r.URL.String())
	})
	_, err = c.BinanceOHLC.Spot(context.Background(), SpotParams{})
	if err == nil || strings.Contains(err.Error(), url.QueryEscape(key)) {
		t.Fatal("encoded query token leaked into a transport error")
	}
}
