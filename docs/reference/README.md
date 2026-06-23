# API Reference

Complete reference for the `gx` library public API.

## Table of Contents

- [Suite](#suite)
- [Column Builders](#column-builders)
- [Expectations](#expectations)
- [EvalColumn](#evalcolumn)
- [Report and Result](#report-and-result)
- [Testing Integration](#testing-integration)

## Suite

The central type in `gx` that groups expectations together.

### NewSuite[T any](expectations ...Expectation[T]) \*Suite[T]

Creates a new validation suite with the provided expectations.

```go
suite := gx.NewSuite[User](
    gx.Ordered("age", func(u User) int { return u.Age }).Between(0, 120),
    gx.Str("email", func(u User) string { return u.Email }).MatchRegex(emailRE),
)
```

### Suite[T].Validate(rows []T) Report

Runs all expectations in the suite against the provided data and returns a
report.

```go
report := suite.Validate(users)
```

### Suite[T].WithSampleCap(cap int) \*Suite[T]

Sets the maximum sample values retained per result and returns the same suite
for chaining.

```go
suite := gx.NewSuite[User](/* expectations */).WithSampleCap(5)
```

## Column Builders

Functions that create typed column accessors for defining field-level
expectations.

### Ordered[T any, V cmp.Ordered](name string, get func(T) V) OrderedColumn[T, V]

Creates a column accessor for ordered types (integers, floats, strings).

```go
ageCol := gx.Ordered("age", func(u User) int { return u.Age })
```

### Str[T any](name string, get func(T) string) StringColumn[T]

Creates a column accessor for string fields.

```go
emailCol := gx.Str("email", func(u User) string { return u.Email })
```

### Comparable[T any, V comparable](name string, get func(T) V) ComparableColumn[T, V]

Creates a column accessor for comparable but not necessarily ordered types.

```go
statusCol := gx.Comparable("status", func(o Order) Status { return o.Status })
```

### Field[T any, V any](name string, get func(T) V) FieldColumn[T, V]

Creates a column accessor for any type (escape hatch for complex types).

```go
tagsCol := gx.Field("tags", func(p Post) []string { return p.Tags })
```

### Numeric[T any, V Number](name string, get func(T) V) NumericColumn[T, V]

Creates a column accessor for numeric aggregate expectations over signed and
unsigned integers, plus floating-point values.

```go
amountCol := gx.Numeric("amount", func(o Order) float64 { return o.Amount })
```

### Row[T any](name string, pred func(T) bool) Expectation[T]

Creates a row-level expectation for cross-field validations.

```go
gx.Row("ship date after order date", func(o Order) bool {
    return !o.ShipDate.Before(o.OrderDate)
})
```

## Expectations

All expectation types and their methods.

### OrderedColumn[T, V] Methods

- `Between(lo, hi V) Expectation[T]` - Value between bounds (inclusive)
- `GreaterThan(bound V) Expectation[T]` - Value greater than bound
- `LessThan(bound V) Expectation[T]` - Value less than bound
- `GreaterOrEqual(bound V) Expectation[T]` - Value greater than or equal to
  bound
- `LessOrEqual(bound V) Expectation[T]` - Value less than or equal to bound
- `In(vals ...V) Expectation[T]` - Value is one of the listed values
- `NotIn(vals ...V) Expectation[T]` - Value is not one of the listed values
- `NotZero() Expectation[T]` - Value is not the zero value
- `Zero() Expectation[T]` - Value is the zero value
- `Satisfy(check string, pred func(V) bool) Expectation[T]` - Custom predicate
- `Unique() Expectation[T]` - All values are distinct

### StringColumn[T] Methods

(Inherits all OrderedColumn methods plus:)

- `MatchRegex(re *regexp.Regexp) Expectation[T]` - Matches regular expression
- `NotMatchRegex(re *regexp.Regexp) Expectation[T]` - Does not match regex
- `NotEmpty() Expectation[T]` - String is not empty
- `Empty() Expectation[T]` - String is empty
- `LenBetween(lo, hi int) Expectation[T]` - Rune count between bounds
- `LenEqual(n int) Expectation[T]` - Rune count equals n

### ComparableColumn[T, V] Methods

- `In(vals ...V) Expectation[T]` - Value is one of the listed values
- `NotIn(vals ...V) Expectation[T]` - Value is not one of the listed values
- `NotZero() Expectation[T]` - Value is not the zero value
- `Zero() Expectation[T]` - Value is the zero value
- `Satisfy(check string, pred func(V) bool) Expectation[T]` - Custom predicate
- `Unique() Expectation[T]` - All values are distinct

### FieldColumn[T, V] Methods

- `Satisfy(check string, pred func(V) bool) Expectation[T]` - Custom predicate

### NumericColumn[T, V] Methods

- `AverageBetween(lo, hi float64) Expectation[T]` - Average in range (inclusive)
- `MedianBetween(lo, hi float64) Expectation[T]` - Median in range (inclusive)
- `StdDevBetween(lo, hi float64) Expectation[T]` - Population standard deviation
  in range (inclusive)
- `SatisfyAggregate(check string, pred func(NumericStats) bool) Expectation[T]` -
  Custom aggregate predicate

### NumericStats

Statistics computed for numeric aggregate expectations:

- `Count int` - Number of extracted values
- `Sum float64` - Sum of values
- `Average float64` - Arithmetic mean
- `Median float64` - Median (even count: average of middle pair)
- `StdDevPopulation float64` - Population standard deviation
- `StdDevSample float64` - Sample standard deviation (`n-1` denominator when
  `n > 1`)

Aggregate expectations set `Result.Column` to the numeric label and leave
per-row fields (`Total`, `FailedCount`, `FailedIndices`, `SampleValues`) at zero
values, matching table-level `RowCount*` semantics. `Success` carries the
verdict.

After evaluation:

- `AverageBetween`, `MedianBetween`, and `StdDevBetween` append the observed
  statistic to `Result.Name` (for example `amount average in [10,100]: got 55`).
- `SatisfyAggregate` uses `"<column>: <check>"` and does not append a `got`
  value.
- Any extracted `NaN` or infinite float fails with `Success=false`, `Err=nil`,
  and `Name` containing `got non-finite value`.
- Empty or nil input passes vacuously; `SatisfyAggregate` predicates are not
  called on empty input.

### Table-Level Expectations

Functions that validate the dataset as a whole. `RowCount*` helpers check
`len(rows)` only. They use table-level `Result` shape: `Column` is `""`, `Total`
is `0`, and per-row fields stay empty. `RowCountBetween`, `RowCountEqual`, and
the threshold helpers append the observed count to `Result.Name` (for example
`row count > 2: got 3`). Custom `RowCount(name, pred)` keeps the caller-provided
name.

- `RowCountGreaterThan[T any](bound int) Expectation[T]` - Row count greater
  than bound
- `RowCountGreaterOrEqual[T any](bound int) Expectation[T]` - Row count greater
  than or equal to bound
- `RowCountLessThan[T any](bound int) Expectation[T]` - Row count less than
  bound
- `RowCountLessOrEqual[T any](bound int) Expectation[T]` - Row count less than
  or equal to bound
- `RowCount[T any](name string, pred func(int) bool) Expectation[T]` - Custom
  row count predicate
- `RowCountBetween[T any](lo, hi int) Expectation[T]` - Row count between bounds
- `RowCountEqual[T any](want int) Expectation[T]` - Exact row count

## EvalColumn

Exported helper for implementing custom `Expectation[T]` types. It runs the
shared per-row loop used by the built-in column builders: extract a value with
`get`, test it with `pred`, and aggregate failures into a `Result`.

### EvalColumn[T, V any](name, column string, rows []T, get func(T) V, pred func(V) bool, opts EvalOptions) Result

```go
func EvalColumn[T, V any](name, column string, rows []T,
    get func(T) V, pred func(V) bool, opts EvalOptions) Result
```

Parameters:

- `name` - expectation name shown in reports (`Result.Name`)
- `column` - column label shown in reports (`Result.Column`; use `""` for
  cross-field checks)
- `rows` - data to evaluate
- `get` - accessor that extracts the value to check from each row
- `pred` - returns `true` when the value is valid, `false` when it fails
- `opts` - evaluation options from the suite (see `EvalOptions` below)

Result semantics:

- `Total` - `len(rows)` (number of rows evaluated)
- `FailedCount` - rows where `pred(get(row))` returned `false`
- `FailedPercent` - `FailedCount / Total * 100` when `Total > 0`, otherwise `0`
- `FailedIndices` - complete list of failing row indices (not capped)
- `SampleValues` - failing values boxed as `[]any`, capped at `opts.SampleCap`
  (first failures retained)
- `Success` - `true` when `FailedCount == 0`
- Empty input (`len(rows) == 0`) passes vacuously: `Success` is `true`, `Total`
  and `FailedCount` are `0`, and `FailedIndices` / `SampleValues` are empty

When a suite runs, `opts.SampleCap` comes from `Suite.WithSampleCap` (default
`DefaultSampleCap`, currently 20). Call `EvalColumn` from your type's `Evaluate`
method and pass through the `opts` argument unchanged.

```go
type trustedDomainExpectation[T any] struct {
    name   string
    column string
    get    func(T) string
    domain string
}

func (e trustedDomainExpectation[T]) Name() string { return e.name }

func (e trustedDomainExpectation[T]) Evaluate(rows []T, opts gx.EvalOptions) gx.Result {
    return gx.EvalColumn(e.name, e.column, rows, e.get, func(email string) bool {
        return strings.HasSuffix(email, "@"+e.domain)
    }, opts)
}
```

### EvalOptions

Options passed by `Suite.Validate` into each expectation's `Evaluate` method.

- `SampleCap int` - maximum failing values stored in `Result.SampleValues`

## Report and Result

Types for validation results.

### Report

Contains results from validating a suite against data.

#### Report.OK() bool

Returns true if all expectations passed.

```go
if report.OK() {
    // All validations passed
}
```

#### Report.Failures() []Result

Returns only the failed results.

```go
failures := report.Failures()
```

#### Report.Results []Result

Slice of all results (both passing and failing).

### Result

Details about a single expectation's validation outcome.

#### Fields

- `Name string` - Human-readable label; table-level checks may append `: got …`
- `Column string` - Accessor label for per-row column and Numeric aggregate
  checks; empty for `gx.Row` and `RowCount*`
- `Success bool` - Whether the expectation passed
- `Total int` - `len(rows)` for per-row checks; `0` for table-level checks
- `FailedCount int` - Number of rows that failed (per-row checks only)
- `FailedPercent float64` - `FailedCount/Total*100`; `0` when `Total==0`
- `SampleValues []any` - Sample of failing values (capped; per-row checks only)
- `FailedIndices []int` - Indices of all failing rows (complete; per-row checks
  only)
- `Err error` - Evaluation error from custom expectations; built-in validation
  failures leave this `nil`

#### Semantics by expectation family

| Family                                  | `Column`       | `Total`     | Per-row fields       | `Name` after `Evaluate`                                                                                                        |
| --------------------------------------- | -------------- | ----------- | -------------------- | ------------------------------------------------------------------------------------------------------------------------------ |
| Per-row column (`Ordered`, `Str`, …)    | accessor label | `len(rows)` | populated on failure | stable label (for example `age between [0,120]`)                                                                               |
| `gx.Row`                                | `""`           | `len(rows)` | populated on failure | stable caller-provided name                                                                                                    |
| `RowCount*` threshold / between / equal | `""`           | `0`         | empty                | appends `: got <n>` (custom `RowCount` keeps caller name)                                                                      |
| `Numeric` aggregates                    | accessor label | `0`         | empty                | range helpers append `: got <stat>`; `SatisfyAggregate` uses `<column>: <check>`; non-finite floats use `got non-finite value` |

`Result.String()` omits row-count details for table-level failures when
`FailedCount == 0`.

## Testing Integration

The `gxtest` package provides integration with Go's testing framework.

### gxtest.Check(t TestingT, suite \*Suite[T], rows []T) bool

Reports failures using `t.Errorf` but continues test execution.

```go
func TestUserData(t *testing.T) {
    suite := gx.NewSuite[User](/* expectations */)
    users := loadTestData()

    gxtest.Check(t, suite, users)
    // Test continues even if validation fails
}
```

### gxtest.Require(t TestingT, suite \*Suite[T], rows []T)

Reports failures using `t.Fatalf` and stops test execution.

```go
func TestUserData(t *testing.T) {
    suite := gx.NewSuite[User](/* expectations */)
    users := loadTestData()

    gxtest.Require(t, suite, users)
    // Test stops if validation fails
}
```

### TestingT Interface

Interface satisfied by `*testing.T` and `*testing.B`:

```go
type TestingT interface {
    Helper()
    Errorf(format string, args ...any)
    Fatalf(format string, args ...any)
}
```
