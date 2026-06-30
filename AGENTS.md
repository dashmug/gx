# Repository Guidelines

## Project Overview

`gx` is a Go data-quality library. It validates an in-memory slice of structs
against declarative, type-safe expectations and returns a structured pass/fail
report. Designed for pipeline gates and test assertions; zero runtime
dependencies.

- Module: `github.com/dashmug/gx`
- Subpackage: `github.com/dashmug/gx/gxtest` (test helper)
- Go: 1.23+, stdlib only

## Architecture & Data Flow

```
User code
  └─ NewSuite[T](expectations...) → *Suite[T]
       └─ .Validate(rows []T) → Report
            ├─ calls Expectation[T].Evaluate(rows, EvalOptions) → Result  (one per expectation)
            ├─ normalizes: Result.Err != nil → Success = false
            └─ collects all results (never short-circuits)

Report
  ├─ OK() bool
  ├─ Failures() []Result
  ├─ Err() error  (*ValidationError wrapping Report, or nil)
  └─ String() string  (human-readable, via render.go)
```

Column builders produce `Expectation[T]` values via two internal helpers:

- `evalColumn` — single-pass per-row loop; boxes failing values into
  `SampleValues` (capped); `FailedIndices` is always complete
- `uniqueExpectation` — single-pass with a `seen` map; first occurrence passes,
  later duplicates fail

## Key Directories

```
gx/                   Package root (package gx)
gx/gxtest/            Test-helper subpackage (package gxtest)
```

All source and tests live flat in the package root. No subdirectories except
`gxtest/`.

## Development Commands

```bash
# Full quality gate (fmt-check + vet + test-race)
make check

# Common individual targets
make test-race   # tests with race detector
make test        # tests without race detector
make vet         # go vet
make fmt         # apply gofmt in place
make fmt-check   # verify formatting (exits non-zero if dirty)
make lint        # golangci-lint (requires installation)
make bench       # benchmarks
make cover       # HTML coverage report → coverage.html
make tidy        # go mod tidy
make clean       # remove coverage.out / coverage.html
make help        # list all targets

# Run a specific package or test by name directly
go test ./gxtest/...
go test -run TestBetweenCountsIndicesPercent ./...
```

## Code Conventions & Common Patterns

**Naming**

- Exported types: `PascalCase` — `Result`, `Report`, `Suite`, `OrderedColumn`,
  `ComparableColumn`
- Unexported internals: `camelCase` — `evalColumn`, `colExpectation`,
  `uniqueExpectation`, `rowExpectation`
- Constructor functions mirror type names: `Ordered(...)`, `Str(...)`,
  `Comparable(...)`, `Field(...)`
- Expectation names appear verbatim in `Result.Name` — e.g.
  `"age between [0,120]"`, `"email: non-empty"`

**Generics**

- Every public API is fully generic: `Suite[T]`, `Expectation[T]`,
  `OrderedColumn[T, V]`, etc.
- Type constraints: `V cmp.Ordered` for ordered columns, `V comparable` for
  comparable/unique, `V any` for Field
- No reflection anywhere

**Error handling**

- `Validate` never panics and never returns an error directly; use
  `Report.Err()` for pipeline gates
- `ValidationError` implements `error`; supports `errors.As` to extract the full
  `Report`
- Custom `Expectation` implementations: if `Result.Err != nil`, `Suite.Validate`
  forces `Success = false`

**Collect-all semantics**

- All expectations always run; `Validate` never short-circuits
- Results are in declaration order

**Internal construction pattern**

```go
// Column builders call newCol or inSet, never construct Result directly:
func (c OrderedColumn[T, V]) Between(lo, hi V) Expectation[T] {
    return newCol(
        fmt.Sprintf("%s between [%v,%v]", c.name, lo, hi),
        c.name,
        c.get,
        func(v V) bool { return v >= lo && v <= hi },
    )
}
```

**SampleCap**

- Default: `DefaultSampleCap = 20`
- Override: `suite.WithSampleCap(n)`; zero means no samples; negative values are
  rejected when `Validate` runs
- Only `SampleValues` is capped; `FailedIndices` is capped when
  `WithFailedIndicesCap` is set (default 100; zero means unlimited)

**FailedIndicesCap**

- Default: `DefaultFailedIndicesCap = 100` (unlimited when set to 0 via
  `WithFailedIndicesCap`)
- Override: `suite.WithFailedIndicesCap(n)`; negative values rejected at
  `Validate`
- `FailedCount` and `FailedPercent` stay complete when indices are capped

**String rendering** (`render.go`)

- `Result.String()` → `"✓ name (N rows)"` or `"✓ name"` when `Total == 0`
- `Result.String()` failure → `"✗ name  M/N failed (P%)  e.g. [vals] @ [indices]"`
- `Report.String()` → header line + indented result lines
- `truncList`: `[a b c]` when ≤ cap; `[a b c …]` (space before U+2026) when over
  cap

## Important Files

| File               | Purpose                                                                                                                                                  |
| ------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `report.go`        | `Result`, `Report`, `ValidationError` types                                                                                                              |
| `suite.go`         | `Suite[T]`, `NewSuite`, `WithSampleCap`, `Validate`, `DefaultSampleCap`, `EvalOptions`                                                                   |
| `expectation.go`   | `EvalColumn` (exported helper); `evalColumn`, `colExpectation`, `newCol`, `inSet`, `notInSet`                                                            |
| `ordered.go`       | `OrderedColumn[T, V]` and its builder `Ordered`                                                                                                          |
| `string.go`        | `StringColumn[T]` and its builder `Str`                                                                                                                  |
| `comparable.go`    | `ComparableColumn[T, V]`, `FieldColumn[T, V]` and their builders `Comparable`, `Field`                                                                   |
| `unique.go`        | `uniqueExpectation`; `unique` internal constructor                                                                                                       |
| `row.go`           | `rowExpectation`, `rowCountExpectation`, `rowCountBetweenExpectation`, `rowCountEqualExpectation`; `Row`, `RowCount`, `RowCountBetween`, `RowCountEqual` |
| `render.go`        | `String()` methods on `Result` and `Report`; `truncList` helper                                                                                          |
| `doc.go`           | Package-level godoc with usage example                                                                                                                   |
| `example_test.go`  | Runnable `Example()` (package `gx_test`) — verifies output                                                                                               |
| `gxtest/gxtest.go` | `TestingT` interface, `Check[T]`, `Require[T]`                                                                                                           |

## Runtime/Tooling Preferences

- **Go 1.23+** required (generic type inference, `cmp.Ordered`)
- **Zero runtime dependencies** — `go.mod` has no `require` block
- **Makefile** for common dev operations; `make check` is the full quality gate
- `golangci-lint` for linting (`.golangci.yml` in the repo root); not required
  at runtime

## Testing & QA

**Framework**: standard `testing` package only — no testify, no gomock.

**Test file locations**: flat alongside source (`suite_test.go`,
`ordered_test.go`, `comparable_test.go`, etc.); `gxtest/gxtest_test.go` is
`package gxtest_test` (external test package).

**Pattern**: inline, direct assertion — no table-driven tests, no subtests.

```go
func TestBetweenCountsIndicesPercent(t *testing.T) {
    rows := []struct{ V int }{{1}, {5}, {10}}
    rep := NewSuite(Ordered("v", func(r struct{ V int }) int { return r.V }).Between(2, 8)).Validate(rows)
    res := rep.Results[0]
    if res.FailedCount != 2 { t.Fatalf(...) }
    ...
}
```

**What tests assert**: `Result` fields (`FailedCount`, `FailedIndices`,
`SampleValues`, `FailedPercent`, `Name`, `Success`), `Report.OK()`,
`Report.Err()`.

**Required gate**: `go test -race ./...` must pass on both
`github.com/dashmug/gx` and `github.com/dashmug/gx/gxtest`.

**Coverage expectations**: each new expectation type gets a test covering
failure count, failed index recording, sample capping, and the vacuous-pass
(empty input) case.

**gxtest usage in tests**:

```go
import "github.com/dashmug/gx/gxtest"

gxtest.Require(t, suite, rows)  // fatal on failure
gxtest.Check(t, suite, rows)    // non-fatal, returns bool
```

`TestingT` is satisfied by `*testing.T` and `*testing.B` without any wrapping.
