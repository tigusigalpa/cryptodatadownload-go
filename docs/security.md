# Security notes

## Credential handling

- The API key is sent as `Authorization: Token <key>` by default. The
  `WithQueryTokenAuth(true)` option explicitly selects the API's query-token
  mode. SDK metadata, retained error bodies, response headers and printable
  errors redact the token, including its URL-encoded form.
- Why: keys in URLs leak into access logs, `Referer` headers, browser history
  and APM telemetry. Query authentication cannot protect against upstream
  access logs; prefer the default header mode.
- Store the key in `CDD_API_KEY` / your secret manager; never commit it.

## Transport

- HTTPS only in the default base URL; `WithBaseURL` exists for test stubs and
  approved proxies — keep TLS verification on for any real endpoint.
- Redirects are followed only within the original scheme and host, preventing
  token forwarding to subdomains or an HTTP downgrade. A supplied
  `CheckRedirect` can further restrict these redirects.
- Buffered success bodies are bounded by `WithMaxResponseBodyBytes` (16 MiB
  default); error diagnostics use `WithMaxErrorBodyBytes` (64 KiB default).
  Streaming success bodies are read on demand and must be closed by the caller.
- Unwrapped transport errors retain their original cause for `errors.Is/As`.
  Treat these low-level diagnostics as sensitive if a custom transport may
  echo credentials. Printable SDK errors are redacted.

## Downloads

- `*Raw` CSV/XLSX variants return unverified bytes — treat them as untrusted
  input. XLSX is a ZIP container: parse with a dedicated library
  (`xuri/excelize`, `tealeg/xlsx`), never unzip untrusted archives to disk.

## Rate limits are contract, not courtesy

- Burst 60 req/min, sustained 10 000 req/24h, and 100 req/h on "heavy"
  endpoints are enforced server-side. Exceeding them yields 429s and can get
  a token throttled or suspended. Keep the built-in limiter on; budget daily
  usage and heavy endpoints in your own scheduler. The default 1 req/s
  ceiling alone does not enforce 10,000 requests per day.

## Reporting

Report vulnerabilities privately to the maintainer (see go.mod module path /
GitHub) rather than opening a public issue.
