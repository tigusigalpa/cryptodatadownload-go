package cryptodatadownload

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"syscall"
	"testing"
	"time"
)

type observedBody struct {
	io.Reader
	reads  int
	closed bool
}

func (b *observedBody) Read(p []byte) (int, error) { b.reads++; return b.Reader.Read(p) }
func (b *observedBody) Close() error               { b.closed = true; return nil }

func TestStreamingDownloadsDoNotBufferOrApplySuccessLimit(t *testing.T) {
	for _, format := range []ReturnFormat{FormatCSV, FormatXLSX} {
		t.Run(string(format), func(t *testing.T) {
			payload := bytes.Repeat([]byte{'P', 'K', 0, 255}, 1024)
			body := &observedBody{Reader: bytes.NewReader(payload)}
			c := transportTestClient(t, roundTripperFunc(func(r *http.Request) (*http.Response, error) {
				if r.URL.Query().Get("return") != string(format) {
					t.Error("stream format was not sent")
				}
				return &http.Response{StatusCode: 200, Header: http.Header{"Content-Disposition": {"attachment; filename=data"}}, Body: body}, nil
			}), WithMaxResponseBodyBytes(8))
			download, err := c.BinanceOHLC.SpotStream(context.Background(), SpotParams{}, format)
			if err != nil {
				t.Fatal(err)
			}
			if body.reads != 0 || body.closed || download.StatusCode != 200 || download.Headers.Get("Content-Disposition") == "" {
				t.Fatal("stream was read or closed before the caller took ownership")
			}
			var output bytes.Buffer
			if n, err := download.WriteTo(&output); err != nil || n != int64(len(payload)) || !bytes.Equal(output.Bytes(), payload) {
				t.Fatalf("stream copy = %d bytes, %v", n, err)
			}
			if err := download.Close(); err != nil || !body.closed {
				t.Fatal("stream body was not closed")
			}
		})
	}
}

func TestStreamingDownloadHonorsCancellation(t *testing.T) {
	ts := newTestServer(t)
	ts.setResponder(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/csv")
		w.WriteHeader(200)
		w.(http.Flusher).Flush()
		<-r.Context().Done()
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c := newTestClient(t, ts)
	download, err := c.BinanceOHLC.SpotStream(ctx, SpotParams{}, FormatCSV)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := download.Close(); err != nil {
			t.Errorf("close download: %v", err)
		}
	}()
	cancel()
	if _, err := io.ReadAll(download.Body); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled stream read = %v", err)
	}
}

func TestResponseMetadataPreservesAttemptsHeadersAndPrecision(t *testing.T) {
	attempts := 0
	headers := http.Header{"Content-Type": {"application/json"}, "X-Source": {"original"}}
	c := transportTestClient(t, roundTripperFunc(func(*http.Request) (*http.Response, error) {
		attempts++
		if attempts == 1 {
			return &http.Response{StatusCode: 503, Header: make(http.Header), Body: io.NopCloser(strings.NewReader("temporary"))}, nil
		}
		return &http.Response{StatusCode: 200, Header: headers, Body: io.NopCloser(strings.NewReader(`{"result":[{"price":0.123456789123456789,"volume":"123456789.123456789"}]}`))}, nil
	}), WithRetryMaxAttempts(2), WithSleeper(sleeperFunc(func(context.Context, time.Duration) error { return nil })))
	result, response, err := c.BinanceOHLC.SpotWithResponse(context.Background(), SpotParams{})
	if err != nil {
		t.Fatal(err)
	}
	if response.Attempts != 2 || response.StatusCode != 200 || response.Method != "GET" || response.Operation != "binanceOHLC.spot" {
		t.Fatalf("metadata = %+v", response)
	}
	headers.Set("X-Source", "changed")
	if response.Headers.Get("X-Source") != "original" {
		t.Fatal("response headers share backing storage with the transport")
	}
	if result.Result[0]["price"] != json.Number("0.123456789123456789") || result.Result[0]["volume"] != "123456789.123456789" {
		t.Fatalf("numeric precision changed: %v", result.Result)
	}
}

func TestTypedAPIErrorsKeepBoundedDiagnosticsAndMetadata(t *testing.T) {
	for _, status := range []int{400, 401, 403, 404, 429, 500, 501, 503} {
		c := transportTestClient(t, roundTripperFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: status, Header: http.Header{"Retry-After": {"3"}}, Body: io.NopCloser(strings.NewReader(strings.Repeat("x", 4096)))}, nil
		}), WithMaxErrorBodyBytes(32))
		_, response, err := c.BinanceOHLC.SpotWithResponse(context.Background(), SpotParams{})
		apiErr := AsAPIError(err)
		if response == nil || response.StatusCode != status || apiErr == nil || apiErr.StatusCode != status || len(apiErr.RawBody) > 32 || !apiErr.BodyTruncated {
			t.Fatalf("status %d: response=%+v, error=%v", status, response, err)
		}
		var badRequest *BadRequestError
		var authentication *AuthenticationError
		var notFound *NotFoundError
		var rateLimit *RateLimitError
		var server *ServerError
		matched := status == 403 || status == 400 && errors.As(err, &badRequest) || status == 401 && errors.As(err, &authentication) || status == 404 && errors.As(err, &notFound) || status == 429 && errors.As(err, &rateLimit) || status >= 500 && errors.As(err, &server)
		if !matched {
			t.Errorf("status %d: missing concrete error type: %T", status, err)
		}
	}
}

func TestDecodeErrorsKeepResponseMetadata(t *testing.T) {
	for _, tc := range []struct{ contentType, body string }{
		{"text/csv", "a,b\n1,2"},
		{"application/json", `{"result":`},
		{"application/json", `{"result":null}`},
	} {
		c := transportTestClient(t, roundTripperFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {tc.contentType}}, Body: io.NopCloser(strings.NewReader(tc.body))}, nil
		}))
		_, response, err := c.BinanceOHLC.SpotWithResponse(context.Background(), SpotParams{})
		var decodeErr *DecodeError
		if !errors.As(err, &decodeErr) || AsTransport(err) == nil || response == nil || decodeErr.StatusCode != 200 {
			t.Errorf("body %q: metadata=%+v, error=%v", tc.body, response, err)
		}
	}
}

func TestOnlyTransientTransportFailuresAreRetried(t *testing.T) {
	for _, tc := range []struct {
		err      error
		attempts int
	}{
		{errors.New("invalid transport configuration"), 1},
		{context.Canceled, 1},
		{&net.DNSError{Err: "no such host", IsNotFound: true}, 1},
		{&net.DNSError{Err: "temporary DNS failure", IsTemporary: true}, 3},
		{io.ErrUnexpectedEOF, 3},
		{syscall.ECONNRESET, 3},
	} {
		attempts := 0
		c := transportTestClient(t, roundTripperFunc(func(*http.Request) (*http.Response, error) {
			attempts++
			return nil, tc.err
		}), WithRetryMaxAttempts(3), WithSleeper(sleeperFunc(func(context.Context, time.Duration) error { return nil })))
		_, err := c.BinanceOHLC.Spot(context.Background(), SpotParams{})
		if !errors.Is(err, tc.err) || attempts != tc.attempts {
			t.Errorf("%v: attempts=%d, error=%v", tc.err, attempts, err)
		}
	}
}
