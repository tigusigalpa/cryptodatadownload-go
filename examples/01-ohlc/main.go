package main

// Binance spot OHLCV. Set CDD_API_KEY first:
//   CDD_API_KEY=your-token go run ./examples/01-ohlc

import (
	"context"
	"fmt"
	"os"

	cdd "github.com/tigusigalpa/cryptodatadownload-go"
)

func main() {
	client, err := cdd.NewClient(os.Getenv("CDD_API_KEY"))
	if err != nil {
		panic(err)
	}
	ctx := context.Background()

	// Daily candles (schema defaults), hourly window, or minute bars.
	daily, err := client.BinanceOHLC.Spot(ctx, cdd.SpotParams{Symbol: cdd.Ptr("BTCUSDT")})
	if err != nil {
		panic(err)
	}
	hourly, err := client.BinanceOHLC.Spot(ctx, cdd.SpotParams{
		Symbol:   cdd.Ptr("ETHUSDT"),
		Interval: cdd.Ptr("1h"),
		Limit:    cdd.Ptr(168),
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("BTCUSDT daily rows: %d\n", len(daily))
	fmt.Printf("ETHUSDT hourly rows: %d\n", len(hourly))
	if len(daily) > 0 {
		fmt.Println("latest:", daily[len(daily)-1])
	}
}
