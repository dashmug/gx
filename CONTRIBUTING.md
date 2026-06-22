# Contributing to gx

Thanks for your interest in improving `gx`. This guide covers the development
workflow, design principles, and conventions the project follows.

## Getting Started

```bash
git clone https://github.com/dashmug/gx
cd gx
make check          # fmt-check + vet + test-race
```

That's the whole setup. `gx` has **zero runtime dependencies**. The `Makefile`
wraps standard `go` tooling — run `make help` to list all available targets.

Requires **Go 1.23+**.

## Project Layout

````
gx/
├── report.go        Result, Report, ValidationError — the data model
├── suite.go         Suite[T], NewSuite, Validate, EvalOptions, Expectation interface
├── expectation.go   Core eval helpers: EvalColumn, colExpectation, newCol, inSet, notInSet
├── ordered.go       OrderedColumn[T, V] and Ordered builder
├── string.go        StringColumn[T] and Str builder
├── comparable.go    ComparableColumn[T, V], FieldColumn[T, V] and their builders
├── unique.go        uniqueExpectation (Unique checks)
├── row.go           Row, RowCount, RowCountBetween, RowCountEqual
├── render.go        String() methods for Result and Report
├── doc.go           Package-level documentation
├── *_test.go        Tests (one file per source file)
├── example_test.go  Runnable Example() (package gx_test)
└── gxtest/          Test-helper sub-package (Check, Require)

Source and tests live flat in the package root. The only sub-package is
`gxtest/`.

## Design Principles

These are non-negotiable. A change that violates one of them will not be
accepted without first revisiting the principle itself.

1. **Zero runtime dependencies.** The library imports only the Go standard
   library. `go.mod` has no `require` block, and it stays that way.

2. **Type-safe, no reflection.** Field access is always a typed closure
   (`func(T) V`) resolved at compile time. We never use `reflect` or
   stringly-typed field paths. If a check can't be expressed with generics, it
   belongs in a custom `Expectation`, not in reflection.

3. **Collect-all, never fail-fast.** `Validate` runs _every_ expectation and
   returns _every_ result. Short-circuiting on the first failure is a
   non-starter — surfacing all problems at once is the point of the library.

4. **Never panic.** `Validate` must not panic on any input, including `nil`
   slices and empty data. Validation failures are data (`Result`), not
   exceptions.

5. **Actionable results.** `FailedIndices` is always complete (never truncated)
   so callers can quarantine every bad row. Only `SampleValues` is capped, and
   only for readable reports.

6. **Empty input passes vacuously.** A check over zero rows succeeds. This keeps
   suites composable across datasets of any size.

## Adding a New Expectation

Most contributions add a check. There are two ways, depending on the check.

### A new column method (most common)

If the check operates on a single column value, add a method to the relevant
column type in the relevant file (`ordered.go`, `string.go`, or `comparable.go`). It returns an `Expectation[T]` built via the shared
`newCol` / `inSet` helpers — never construct a `Result` by hand.

```go
// Positive asserts the value is greater than zero.
func (c OrderedColumn[T, V]) Positive() Expectation[T] {
    var zero V
    return newCol(c.name+" positive", c.name, c.get, func(v V) bool { return v > zero })
}
````

The first `newCol` argument becomes `Result.Name` — make it read naturally in a
report (`"age positive"`, not `"PositiveCheck"`).

### A custom `Expectation`

For checks the column helpers can't express (aggregates, cross-row logic),
implement the `Expectation[T]` interface directly in a new file or in `row.go`
if row-level, `unique.go` if aggregate:

```go
type Expectation[T any] interface {
    Name() string
    Evaluate(rows []T, opts EvalOptions) Result
}
```

- Honor `opts.SampleCap` when populating `SampleValues`.
- Keep `FailedIndices` complete (uncapped).
- Set `Result.Err` if evaluation itself fails — the suite normalizes a non-nil
  `Err` to `Success = false`, so a broken check can never silently pass.
- A single pass over `rows` is the norm; avoid extra allocations on the hot
  path.

## Testing

`gx` follows **test-driven development**. For any change:

1. Write a failing test that encodes the intended behavior.
2. Confirm it fails for the right reason.
3. Implement the minimum to make it pass.
4. Confirm green.

### Conventions

- **Standard library only.** Use the `testing` package — no testify, no mocking
  frameworks.
- **Inline, direct assertions.** No table-driven tests, no subtests. Keep each
  test focused on one behavior.
- Test files mirror source files (`ordered.go` → `ordered_test.go`, `row.go` →
  `row_test.go`, etc.). Tests are `package gx`; `gxtest` external tests are
  `package gxtest_test`.

### What to assert

Test behavior, not plumbing. For an expectation, cover:

- The **failure count** and that `Success` is correct.
- The **failed indices** — assert the exact indices, e.g.
  `FailedIndices[0] == 1`.
- **Sample capping** — that `SampleValues` respects `SampleCap` while
  `FailedIndices` stays complete.
- The **vacuous-pass** case — empty/`nil` input succeeds.

```go
func TestPositiveFlagsNonPositive(t *testing.T) {
    rows := []struct{ V int }{{1}, {-2}, {0}}
    rep := NewSuite(Ordered("v", func(r struct{ V int }) int { return r.V }).Positive()).Validate(rows)

    res := rep.Results[0]
    if res.FailedCount != 2 {
        t.Fatalf("FailedCount = %d, want 2", res.FailedCount)
    }
    if len(res.FailedIndices) != 2 || res.FailedIndices[0] != 1 || res.FailedIndices[1] != 2 {
        t.Fatalf("FailedIndices = %v, want [1 2]", res.FailedIndices)
    }
}
```

Guard slice indexing with a length check before asserting `FailedIndices[0]`, so
a regression that empties the slice fails with a clear message instead of a
panic.

### Required gate

Before submitting, the full gate must pass:

```bash
make check   # equivalent to: gofmt -l . && go vet ./... && go test -race ./...
```

Or run the steps individually:

```bash
make fmt-check   # must print nothing
make vet
make test-race
```

## Code Conventions

- **Formatting:** `gofmt`. No exceptions, no custom style.
- **Naming:** exported types are `PascalCase`; constructor functions mirror the
  type they build (`Ordered` → `OrderedColumn`). Unexported helpers are
  `camelCase`.
- **Comments explain _why_, not _what_.** Document non-obvious intent and
  invariants; don't restate the code. Keep godoc on every exported symbol.
- **No dead code.** No commented-out blocks, no `TODO`s left for later — git has
  the history, and unfinished work shouldn't ship.

## Commit & PR Conventions

- **Conventional Commits** for messages: `feat:`, `fix:`, `docs:`, `test:`,
  `refactor:`, `chore:`. Keep the subject under ~72 characters and in the
  imperative mood.

  ```
  feat: add Positive check to OrderedColumn
  fix: cap SampleValues in uniqueExpectation
  ```

- One logical change per commit; keep the TDD red→green cycle within a single,
  reviewable PR.
- PRs should describe _what_ changed and _why_, and note any new public API.

### Before you submit

- [ ] New behavior is covered by a test that fails without your change.
- [ ] `go test -race ./...` passes both packages.
- [ ] `gofmt -l .` prints nothing; `go vet ./...` is clean.
- [ ] Every new exported symbol has a godoc comment.
- [ ] No dead code, leftover `TODO`s, or commented-out blocks.
- [ ] Public API changes are reflected in `README.md`.
