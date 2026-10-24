package main

// Pre-computed ML features & aggregates — partition reads (required UTC
// date). A 404 means "no partition for that date yet", not a routing failure.
//   CDD_API_KEY=your-token go run ./examples/02-ml-features

import (
	"context"
	"errors"
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

	yesterday := time.Now().UTC().AddDate(0, 0, -1)

	// ~70-column composite in one call; project just the columns you need.
	rows, err := client.Aggregates.FeaturesMlAll(ctx, cdd.FeaturesMlAllParams{
		Symbol: "BTCUSDT",
		Date:   yesterday,
		Fields: []string{"symbol", "date", "vol_rank_z", "momentum_12_1"},
	})
	var apiErr *cdd.ApiError
	switch {
	case err == nil:
		for _, r := range rows {
			fmt.Println(r)
		}
	case errors.As(err, &apiErr) && apiErr.IsNotFound():
		fmt.Println("no partition for yesterday yet — try an earlier date")
	default:
		panic(err)
	}
}
