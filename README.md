# gx

Type-safe, declarative data quality for Go.

`gx` validates a `[]T` directly: build a suite with `NewSuite(...)` and run
`suite.Validate(rows)` to get a rich pass/fail report — including the complete
list of failing row indices in `FailedIndices`, so you can quarantine bad
records instead of just learning that _something_ failed.

Think [Great Expectations](https://greatexpectations.io/), but Go-native,
generic, and with zero runtime dependencies.

```go
suite := gx.NewSuite[User](
    gx.Ordered("age", func(u User) int { return u.Age }).Between(0, 120),
    gx.Str("email", func(u User) string { return u.Email }).MatchRegex(emailRE),
    gx.Comparable("id", func(u User) string { return u.ID }).Unique(),
)

if err := suite.Validate(users).Err(); err != nil {
    // err is a *gx.ValidationError carrying the full report
    log.Fatal(err)
}
```

## Why gx

- **Type-safe.** Field access is a closure (`func(u User) int`), checked by the
  compiler. No reflection, no stringly-typed field paths, no runtime type
  errors.
- **Collect-all.** Every expectation runs on every `Validate` call. You get
  _all_ failures at once, not just the first — essential for triaging a dirty
  dataset.
- **Actionable results.** Each failure reports the count, percentage, a sample
  of offending values, and the **complete** list of failing row indices.
- **Zero dependencies.** Standard library only.
- **Test-friendly.** The `gxtest` sub-package is a thin adapter over the runtime
  API for `*testing.T`.

## Installation

```bash
go get github.com/dashmug/gx
```

Requires **Go 1.23+** (uses generics and `cmp.Ordered`).

## Quick Start

```go
package main

import (
    "fmt"
    "regexp"

    "github.com/dashmug/gx"
)

func main() {
    type User struct {
        Age   int
        Email string
    }
    users := []User{
        {Age: 30, Email: "a@example.com"},
        {Age: 200, Email: "bad"},
    }
    emailRE := regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

    suite := gx.NewSuite[User](
        gx.Ordered("age", func(u User) int { return u.Age }).Between(0, 120),
        gx.Str("email", func(u User) string { return u.Email }).MatchRegex(emailRE),
    )

    report := suite.Validate(users)
    fmt.Println(report.OK())
    for _, r := range report.Failures() {
        fmt.Printf("%s: %d/%d failed at %v\n", r.Name, r.FailedCount, r.Total, r.FailedIndices)
    }
}
```

Output:

```
false
age between [0,120]: 1/2 failed at [1]
email matches /^[^@\s]+@[^@\s]+\.[^@\s]+$/: 1/2 failed at [1]
```

Row `1` (the `{200, "bad"}` user) failed both checks, and `gx` tells you exactly
which row.

## Core Concepts

A **`Suite[T]`** is an ordered set of **expectations** over rows of type `T`.
Calling `Validate(rows)` runs all of them and returns a **`Report`**.

```go
suite := gx.NewSuite[T](expectation1, expectation2, ...)
report := suite.Validate(rows)
```

**Typed column builders** (`Ordered`, `Str`, `Comparable`, `Field`) are the
primary authoring path — you rarely implement `Expectation` directly. Each
builder takes a label and an accessor function:

```go
gx.Ordered("age", func(u User) int { return u.Age })  // a column
    .Between(0, 120)                                   // an expectation
```

The label (`"age"`) is purely for the report — it can differ from the struct
field, so computed values are first-class:

```go
gx.Ordered("name length", func(u User) int { return len(u.Name) }).Between(1, 100)
```

## Column Types

Pick the builder that matches your field's type. Each exposes the checks that
make sense for it.

### `Ordered` — integers, floats, ordered values

```go
col := gx.Ordered("age", func(u User) int { return u.Age })

col.Between(0, 120)             // lo <= v <= hi (inclusive)
col.In(1, 2, 3)                 // v is one of the listed values
col.NotIn(4, 5)                 // v not in listed values
col.NotZero()                   // v != zero value
col.Zero()                      // v == zero value (matching rows pass)
col.GreaterThan(0)              // v > bound
col.LessThan(100)               // v < bound
col.GreaterOrEqual(0)           // v >= bound
col.LessOrEqual(100)            // v <= bound
col.Satisfy("even", isEven)     // custom predicate func(V) bool
col.Unique()                    // every value distinct
```

### `Str` — strings

Embeds `Ordered`, so all of the above work too, plus:

```go
col := gx.Str("email", func(u User) string { return u.Email })

col.MatchRegex(emailRE)         // matches a *regexp.Regexp
col.NotMatchRegex(emailRE)      // does not match *regexp.Regexp
col.NotEmpty()                  // non-empty string
col.Empty()                     // v == "" (matching rows pass)
col.LenBetween(1, 100)          // rune count in [lo, hi]
col.LenEqual(10)                // rune count == n
```

`LenBetween`/`LenEqual` count Unicode code points via `utf8.RuneCountInString`,
not bytes. `Zero`/`Empty` pass rows that match; they are complements of
`NotZero`/`NotEmpty`.

### `Comparable` — bools, enums, struct keys

For values that are comparable but not ordered:

```go
col := gx.Comparable("status", func(o Order) Status { return o.Status })

col.In(Active, Pending)         // v is one of the listed values
col.NotIn(Closed, Cancelled)    // v is not one of the listed values
col.NotZero()                   // v != zero value
col.Zero()                      // v == zero value (matching rows pass)
col.Satisfy("terminal", isDone) // custom predicate func(V) bool
col.Unique()                    // every value distinct
```

### `Numeric` — integers, floats (aggregates)

For dataset-level statistics over a numeric column:

```go
col := gx.Numeric("amount", func(o Order) float64 { return o.Amount })

col.AverageBetween(10, 100)     // lo <= average <= hi
col.MedianBetween(10, 100)      // lo <= median <= hi
col.StdDevBetween(0, 25)        // population std dev in [lo, hi]
col.SatisfyAggregate("CV <= 0.25", func(s gx.NumericStats) bool {
    return s.Average == 0 || s.StdDevPopulation/s.Average <= 0.25
})
```

Unlike `Ordered`, `Numeric` checks dataset-level statistics, not individual
rows. Results use table-level reporting: `Result.Column` is the accessor label,
`Total` is `0`, and per-row fields (`FailedCount`, `FailedIndices`,
`SampleValues`) stay empty — `Success` carries the pass/fail verdict.

After evaluation, `AverageBetween`, `MedianBetween`, and `StdDevBetween` append
the observed statistic to `Result.Name` (for example
`amount average in [10,100]: got 55`). `SatisfyAggregate` uses
`"<column>: <check>"` and does not append a `got` value. If any extracted
`float32`/`float64` is `NaN` or infinite, built-in aggregates fail with
`Success=false`, `Err=nil`, and a `Name` containing `got non-finite value`.

Empty or nil input passes vacuously for built-in aggregates and does not call
`SatisfyAggregate` predicates; use `RowCountGreaterThan[T](0)` when you need
non-empty data. `StdDevBetween` uses population standard deviation;
`NumericStats.StdDevSample` is available in `SatisfyAggregate` callbacks.

### `Field` — any type (escape hatch)

For types that are neither ordered nor comparable (slices, maps, structs):

```go
gx.Field("tags", func(p Post) []string { return p.Tags }).
    Satisfy("non-empty", func(v []string) bool { return len(v) > 0 })
```

## Row & Table-Level Checks

### Cross-field rules — `Row`

When a rule spans multiple fields of the same row:

```go
gx.Row("ship date after order date", func(o Order) bool {
    return !o.ShipDate.Before(o.OrderDate)
})
```

### Row-count rules — `RowCount`

Assertions about the _number_ of rows rather than their contents.
`RowCountGreaterThan` / `RowCountLessThan` and friends cover common thresholds;
`RowCountBetween` and `RowCountEqual` cover ranges and exact counts;
`RowCount(name, pred)` is the escape hatch for custom rules:

```go
gx.RowCountGreaterThan[User](0)      // len(rows) > bound
gx.RowCountGreaterOrEqual[User](1)   // len(rows) >= bound
gx.RowCountLessThan[User](10000)     // len(rows) < bound
gx.RowCountLessOrEqual[User](1000)   // len(rows) <= bound
gx.RowCountBetween[User](1, 1000)    // 1 <= len(rows) <= 1000
gx.RowCountEqual[User](42)           // len(rows) == 42
gx.RowCount[User]("even batch", func(n int) bool { return n%2 == 0 })
```

`RowCount*` checks are also table-level: `Column` is `""`, `Total` is `0`, and
per-row fields stay empty. `RowCountBetween`, `RowCountEqual`, and the threshold
helpers (`RowCountGreaterThan`, …) append the observed count to `Result.Name`
(for example `row count > 2: got 3`). Custom `RowCount(name, pred)` keeps the
caller-provided `Name` unchanged.

## Working with Results

`Validate` returns a `Report`. How you consume it depends on the context.

### Gate a pipeline

```go
if err := suite.Validate(rows).Err(); err != nil {
    return err   // *gx.ValidationError, names every failed expectation
}
```

Recover the full report from the error with `errors.As`:

```go
var verr *gx.ValidationError
if errors.As(err, &verr) {
    for _, r := range verr.Report.Failures() {
        quarantine(rows, r.FailedIndices)
    }
}
```

### Inspect every result

```go
report := suite.Validate(rows)

report.OK()           // true if all expectations passed
report.Failures()     // []Result, only the failures
report.Results        // []Result, all of them, in declaration order
```

### A `Result`

```go
type Result struct {
    Name          string    // human-readable label; table-level checks may append ": got …"
    Column        string    // accessor label for per-row column and Numeric aggregate checks; "" for gx.Row and RowCount*
    Success       bool
    Total         int       // len(rows) for per-row checks; 0 for table-level checks
    FailedCount   int
    FailedPercent float64   // FailedCount/Total*100; 0 when Total==0
    SampleValues  []any     // capped sample of offending values
    FailedIndices []int     // indices into the slice; complete (uncapped)
    Err           error     // set only by custom expectations that error
}
```

Per-row column and `gx.Row` checks populate `Total`, `FailedCount`,
`FailedIndices`, and `SampleValues`. Table-level `RowCount*` and `Numeric`
aggregate checks leave those fields at zero — inspect `Success` and `Name`
instead. `Result.String()` renders table-level failures without row counts when
`FailedCount == 0`.

`FailedIndices` is **never truncated** in the field — it always lists every
failing row — but `Result.String()` may truncate its displayed samples/indices
for readability. `SampleValues` _is_ capped (default 20) for readable reports.

### Human-readable output

Both `Report` and `Result` implement `String()`:

```go
fmt.Println(suite.Validate(users))
```

```
gx report: 0/2 expectations passed
  ✗ age between [0,120]  1/2 failed (50.0%)  e.g. [200] @ [1]
  ✗ email matches /^[^@\s]+@[^@\s]+\.[^@\s]+$/  1/2 failed (50.0%)  e.g. [bad] @ [1]
```

## Use in Tests

The `gxtest` sub-package is a thin adapter over the runtime API, mirroring the
assert/require split:

```go
import (
    "testing"

    "github.com/dashmug/gx"
    "github.com/dashmug/gx/gxtest"
)

func TestUserData(t *testing.T) {
    suite := gx.NewSuite[User](
        gx.Ordered("age", func(u User) int { return u.Age }).Between(0, 120),
    )

    // Require: stops the test on failure (t.Fatalf)
    gxtest.Require(t, suite, loadUsers())

    // Check: reports failure but continues (t.Errorf), returns bool
    if !gxtest.Check(t, suite, loadUsers()) {
        // ...
    }
}
```

`gxtest` works with any `TestingT` (`*testing.T` and `*testing.B` satisfy it out
of the box):

```go
type TestingT interface {
    Helper()
    Errorf(format string, args ...any)
    Fatalf(format string, args ...any)
}
```

## Tuning

### Sample cap

Limit how many offending sample values each `Result` retains (default
`gx.DefaultSampleCap`, which is 20):

```go
suite := gx.NewSuite[User](...).WithSampleCap(5)
```

Only `SampleValues` is affected; `FailedIndices` is always complete.

## Custom Expectations

For logic the column helpers can't express, implement `Expectation[T]` directly:

```go
type Expectation[T any] interface {
    Name() string
    Evaluate(rows []T, opts gx.EvalOptions) gx.Result
}
```

Set `Result.Err` if evaluation itself fails — the suite normalizes a non-nil
`Err` to `Success = false`, so a broken check can never silently pass.

## Behavior Notes

- **Empty input passes vacuously.** A per-row column check over zero rows
  succeeds (`Total == 0`). Built-in numeric aggregates and `SatisfyAggregate`
  also pass on empty or nil input; compose `RowCountGreaterThan[T](0)` when
  non-empty data is required.
- **Non-finite floats fail aggregates.** If a `Numeric` accessor returns `NaN`
  or ±`Inf`, built-in aggregate expectations fail as data (`Success=false`,
  `Err=nil`) with `Name` containing `got non-finite value`.
- **`Validate` never panics** and never returns an error directly — gate via
  `Report.Err()`.
- **`Unique`**: the first occurrence of a value passes; every later duplicate
  fails.
- **Results preserve declaration order.**

## Documentation

See the [full documentation](docs/) for:

- [Concepts](docs/concepts/) - Core concepts and usage patterns
- [Built-in Expectations](docs/expectations/) - Complete reference of all
  validation functions
- [Custom Expectations](docs/custom/) - How to extend `gx` with your own
  validation logic
- [Comparisons](docs/comparison/) - How `gx` differs from other validation
  approaches
- [Cookbook](docs/cookbook/) - Practical examples and patterns for real-world
  usage
- [API Reference](docs/reference/) - Complete reference for all public types and
  functions
