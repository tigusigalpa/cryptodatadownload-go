package cryptodatadownload

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestIntegrationCatalog(t *testing.T) {
	if os.Getenv("CDD_RUN_INTEGRATION_TESTS") != "1" || os.Getenv("CDD_API_TOKEN") == "" {
		t.Skip("requires CDD_RUN_INTEGRATION_TESTS=1 and CDD_API_TOKEN")
	}
	c, err := NewClient(os.Getenv("CDD_API_TOKEN"), WithRetryMaxAttempts(1))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	result, response, err := c.Catalog.OhlcBinanceAllAvailableWithResponse(ctx, OhlcBinanceAllAvailableParams{})
	if err != nil {
		t.Fatal(err)
	}
	if result == nil || response.StatusCode != 200 || response.Attempts != 1 {
		t.Fatal("catalog did not return a successful single-attempt response")
	}
}
