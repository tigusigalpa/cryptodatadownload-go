package cryptodatadownload

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/tigusigalpa/cryptodatadownload-go/internal/clock"
)

// recordedRequest captures a single request seen by the test server.
type recordedRequest struct {
	Method string
	Path   string
	Query  string
	Header http.Header
	Body   []byte
}

// testServer is an httptest.Server that always records requests and returns
// either a canned response or a custom responder.
type testServer struct {
	*httptest.Server
	mu         sync.Mutex
	requests   []recordedRequest
	status     int
	respBody   []byte
	respHeader http.Header
	responder  func(w http.ResponseWriter, r *http.Request)
}

func newTestServer(t *testing.T) *testServer {
	t.Helper()
	ts := &testServer{status: http.StatusOK, respHeader: http.Header{}}
	ts.Server = httptest.NewServer(http.HandlerFunc(ts.serve))
	t.Cleanup(ts.Close)
	return ts
}

func (ts *testServer) serve(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	ts.mu.Lock()
	ts.requests = append(ts.requests, recordedRequest{
		Method: r.Method,
		Path:   r.URL.Path,
		Query:  r.URL.RawQuery,
		Header: r.Header.Clone(),
		Body:   body,
	})
	responder := ts.responder
	status := ts.status
	respBody := ts.respBody
	respHeader := ts.respHeader
	ts.mu.Unlock()

	if responder != nil {
		responder(w, r)
		return
	}
	for k, vs := range respHeader {
		for _, v := range vs {
			w.Header().Set(k, v)
		}
	}
	if w.Header().Get("Content-Type") == "" && len(respBody) > 0 {
		w.Header().Set("Content-Type", "application/json")
	}
	w.WriteHeader(status)
	_, _ = w.Write(respBody)
}

func (ts *testServer) setResponder(f func(w http.ResponseWriter, r *http.Request)) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.responder = f
}

func (ts *testServer) setResponse(status int, body []byte) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.status = status
	ts.respBody = body
	ts.responder = nil
}

func (ts *testServer) lastRequest() recordedRequest {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if len(ts.requests) == 0 {
		return recordedRequest{}
	}
	return ts.requests[len(ts.requests)-1]
}

func (ts *testServer) requestCount() int {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return len(ts.requests)
}

// newTestClient builds a Client pointed at ts with deterministic retry and a
// disabled rate limiter (0 rps) so tests never wait.
func newTestClient(t *testing.T, ts *testServer, opts ...Option) *Client {
	t.Helper()
	all := append([]Option{
		WithBaseURL(ts.URL),
		WithSleeper(&clock.FakeSleeper{}),
		WithClock(clock.NewFakeClock(clock.SystemClock{}.Now())),
		WithRetryMaxAttempts(1),
		WithRequestsPerSecond(0),
	}, opts...)
	c, err := NewClient("test-key", all...)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

// requireToken asserts the Authorization header carries the Token scheme —
// "Token test-key", never "Bearer test-key" — and that no auth_token ever
// leaks into the query string.
func requireToken(t *testing.T, req recordedRequest) {
	t.Helper()
	got := req.Header.Get("Authorization")
	if got != "Token test-key" {
		t.Fatalf("Authorization = %q, want %q", got, "Token test-key")
	}
	if strings.Contains(req.Query, "auth_token=") {
		t.Fatalf("query %q must never contain auth_token", req.Query)
	}
}
