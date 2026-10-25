# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed

- Renamed `ApiError` and `AsApiError` to `APIError` and `AsAPIError` to follow
  Go initialism conventions. Callers using the previous names must update them.
- Rate-limited sequential and concurrent calls no longer reserve fresh slots
  indefinitely; cancelled waits do not consume capacity.
- The three Deribit symbol catalog methods return `[]string`, matching the
  published API responses. This changes their previous, unusable return type.
- Incomplete 400/401/404 responses and oversized success bodies are not retried;
  retries of incomplete transient responses honor `Retry-After`.
- HTTP-date retry delays round up, and large retry delays cannot overflow.
- Empty enum values, invalid numeric enums and raw formats, invalid rate
  limits and malformed client configuration fail before sending requests.
- Redirects cannot forward the token to another host or downgrade HTTPS;
  echoed keys are redacted from printable API and transport errors.
- Null object responses and oversized bodies after empty reads are rejected.
- CI shell scripts run with Bash on Windows; setup no longer expects a
  `go.sum` file for this dependency-free module.
- Non-finite query numbers, zero optional dates and UTC years outside
  0001–9999 fail locally; optional zero values remain explicit on the wire.
- Permanent transport errors are no longer retried, and arbitrarily large
  Retry-After values saturate safely instead of bypassing the delay.

### Added

- A reproducible OpenAPI snapshot and contract fixtures using official
  response examples, plus regression tests for runtime behavior.
- Streaming CSV/XLSX downloads on all 43 format-capable operations, with
  explicit ownership of the response body and bounded memory usage.
- `Result[T]`, sanitized `Response` metadata and `*WithResponse` companions
  for all 116 typed operations, including HTTP status and attempt count.
- Concrete HTTP and decode error types compatible with existing
  `errors.As` checks, plus independently bounded error diagnostics.
- Option-based construction through `NewClientWithOptions`/`WithToken`,
  explicit query-token authentication, and shareable rate limiters.
- Complete endpoint/parameter/default/enum/operationId manifest checks,
  official hourly-limit notes and a generated README endpoint index.
- Opt-in catalog integration testing, contribution/security policies,
  coverage and pinned Staticcheck CI checks on Go 1.27.

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
