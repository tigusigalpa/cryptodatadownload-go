package main

// Error handling, retries and rate limiting.
//   CDD_API_KEY=your-token go run ./examples/04-errors

import (
	"context"
	"errors"
	"fmt"
	"os"

	cdd "github.com/tigusigalpa/cryptodatadownload-go"
)

func main() {
	client, err := cdd.NewClient(os.Getenv("CDD_API_KEY"),
		cdd.WithRequestsPerSecond(0.5), // 30 req/min steady-state
		cdd.WithRetryMaxAttempts(4),
	)
	if err != nil {
		panic(err)
	}

	_, err = client.Catalog.OhlcBinanceAllAvailable(context.Background(), cdd.OhlcBinanceAllAvailableParams{})

	var apiErr *cdd.APIError
	var tErr *cdd.TransportError
	var argErr *cdd.ArgumentError
	switch {
	case err == nil:
		fmt.Println("ok")
	case errors.As(err, &apiErr):
		switch {
		case apiErr.IsAuthentication():
			fmt.Println("401 — check CDD_API_KEY")
		case apiErr.IsNotFound():
			fmt.Println("404 — partition not found / no such resource")
		case apiErr.IsRateLimited():
			secs, _ := apiErr.RetryAfter()
			fmt.Printf("429 — back off; Retry-After: %ds\n", secs)
		case apiErr.IsValidation():
			fmt.Printf("400 — bad params: %s\n", apiErr.Message)
		case apiErr.IsServerError():
			fmt.Printf("%d — upstream error\n", apiErr.Status)
		}
	case errors.As(err, &tErr):
		fmt.Println("transport failure:", tErr)
	case errors.As(err, &argErr):
		fmt.Println("invalid argument:", argErr)
	}
}
