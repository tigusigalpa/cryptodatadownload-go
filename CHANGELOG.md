# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-10-24

### Added

- Initial release covering all **116 GET operations** of the CryptoDataDownload
  REST API (`https://api.cryptodatadownload.com/v1`), grouped into 21 service
  clients matching the official OpenAPI tags. Zero external dependencies.
- `context.Context` on every call; injectable and reused `*http.Client`.
- `Token`-scheme authentication (`Authorization: Token <key>`) — never
  `Bearer`, never `auth_token` in the query string.
- Typed `ApiError` with status predicates and `Retry-After` parsing for 429,
  `TransportError` for failures below the API contract, and `ArgumentError`
  for client-side parameter validation.
- Built-in minimum-interval rate limiter (default 1 req/s) plus exponential
  backoff with full jitter; `Retry-After` takes precedence and `ctx`
  cancellation is honored during waits.
- Per-operation `*Params` structs (required fields, pointer optionals),
  `ReturnFormat` + `*Raw` variants on the 43 legacy operations that declare a
  `return` query parameter.
- `internal/clock` injectable Clock/Sleeper for deterministic tests;
  generated contract test covering all 116 operations against a stub server.
