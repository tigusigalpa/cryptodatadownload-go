# Contributing

Use Go 1.22 or newer. Keep runtime dependencies in the standard library and
preserve existing public call signatures when adding functionality.

The repository's `.gitattributes` keeps text files in LF form on all
platforms. This also applies when Git on Windows uses `core.autocrlf=true`,
so a fresh checkout agrees with `gofmt` without rewriting source files in CI.

Run the offline checks before proposing a change:

```bash
gofmt -l .
go vet ./...
go test ./...
go test -race ./...
go test -cover ./...
go run honnef.co/go/tools/cmd/staticcheck@v0.8.2 ./...
go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.12
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 run --timeout=5m
```

The pinned Staticcheck development tool requires Go 1.26 or newer; the SDK
itself continues to target Go 1.22. CI tests the minimum version and Go 1.27.
No API token is needed for these checks. Integration tests require both
`CDD_RUN_INTEGRATION_TESTS=1` and `CDD_API_TOKEN` and make one catalog request.

Use Go 1.27 for the pinned development tools to match CI. The workflow files
have separate responsibilities:

- `ci.yml`: build, formatting, vet, tests and race tests on Linux, macOS and
  Windows with Go 1.22 and 1.27. Go 1.22 on newer macOS uses the system linker
  to include the `LC_UUID` metadata required by the OS.
- `test.yml`: pinned Actionlint, Staticcheck and GolangCI-Lint checks.
- `coverage.yml`: race-enabled coverage with an artifact available on GitHub.
  Upload to Codecov is optional and runs when `CODECOV_TOKEN` is configured.
- `codeql.yml`: Go 1.27 setup followed by explicit CodeQL initialization,
  build and analysis.

Workflows run for pull requests, pushes to `main`, `master` or `develop`, and
manual dispatch. Missing external coverage credentials do not prevent the
test suite or the coverage artifact from being produced.

The generated services, endpoint manifest, README coverage index and contract
tests come from the versioned fixtures in `testdata/`. PHP 8.0+ is needed only
for regeneration, never for consuming the SDK:

```bash
php tools/generate.php
gofmt -w .
```

When refreshing fixtures, download the official OpenAPI and CSV catalog,
record their retrieval date, inspect the diff, and regenerate. OpenAPI is
authoritative for operation paths, parameters, defaults and enums; the CSV
catalog supplies additional rate-limit notes. Do not infer new operations or
formats from marketing examples. Preserve the exact Token authentication
prefix and legacy parameter spellings.

`TestManifestPreservesPublishedContract` verifies all 116 operation mappings,
query structures, defaults, enums, raw/stream support and documentation URLs.
Changes to the baseline count or public return types need explicit review and
a changelog entry. Do not regenerate from the network during normal tests.
