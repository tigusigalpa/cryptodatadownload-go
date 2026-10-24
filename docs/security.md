# Security notes

## Credential handling

- The API key is sent exclusively as `Authorization: Token <key>`. The SDK
  never puts it in the URL, never emits the optional `auth_token` query
  parameter the API technically allows, and never includes it in `Error()`
  output or log lines.
- Why: keys in URLs leak into access logs, `Referer` headers, browser history
  and APM telemetry. If a downstream system truly cannot set headers, build
  the URL yourself — the SDK does not expose `auth_token` as an option.
- Store the key in `CDD_API_KEY` / your secret manager; never commit it.

## Transport

- HTTPS only in the default base URL; `WithBaseURL` exists for test stubs and
  approved proxies — keep TLS verification on for any real endpoint.
- Redirects are not followed explicitly; the default `http.Client` follows
  standard Go redirect policy — supply `WithHTTPClient` with a restrictive
  `CheckRedirect` if your threat model requires it.
- Response bodies are bounded by `WithMaxResponseBodyBytes` (16 MiB default)
  to prevent unbounded memory growth.

## Downloads

- `*Raw` CSV/XLSX variants return unverified bytes — treat them as untrusted
  input. XLSX is a ZIP container: parse with a dedicated library
  (`xuri/excelize`, `tealeg/xlsx`), never unzip untrusted archives to disk.

## Rate limits are contract, not courtesy

- Burst 60 req/min, sustained 10 000 req/24h, and 100 req/h on "heavy"
  endpoints are enforced server-side. Exceeding them yields 429s and can get
  a token throttled or suspended. Keep the built-in limiter on; budget heavy
  endpoints in your own scheduler.

## Reporting

Report vulnerabilities privately to the maintainer (see go.mod module path /
GitHub) rather than opening a public issue.
