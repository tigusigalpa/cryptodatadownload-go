package main

// Raw CSV/XLSX download — only on the 43 legacy operations that declare a
// `return` parameter. *Raw returns the response body verbatim.
//   CDD_API_KEY=your-token go run ./examples/03-csv

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

	data, err := client.BinanceOHLC.SpotRaw(context.Background(), cdd.SpotParams{
		Symbol: cdd.Ptr("BTCUSDT"),
		Limit:  cdd.Ptr(500),
	}, cdd.FormatCSV)
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile("btcusdt_daily.csv", data, 0o644); err != nil {
		panic(err)
	}
	fmt.Printf("wrote btcusdt_daily.csv (%d bytes)\n", len(data))
}
