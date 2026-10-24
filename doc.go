// Package cryptodatadownload provides a typed Go client for the
// CryptoDataDownload REST API — Deribit options analytics, Binance OHLCV with
// multi-year depth, cross-venue funding rates, on-chain signals, pre-computed
// ML features and risk metrics.
//
// The client covers all 116 GET operations of
// https://api.cryptodatadownload.com/v1, generated from and verified against
// the live OpenAPI schema at
// https://api.cryptodatadownload.com/?format=openapi.
//
// Authentication uses a single header — "Authorization: Token <key>" — on
// every request. The key is never placed in the query string (the optional
// auth_token parameter the API technically accepts is deliberately
// unsupported) and never appears in errors or logs.
//
// Create one Client with NewClient and reuse it for the lifetime of your
// application; all services are safe for concurrent use.
package cryptodatadownload
