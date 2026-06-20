# API Reference

Complete reference for the `gx` library public API.

## Table of Contents

- [Suite](#suite)
- [Column Builders](#column-builders)
- [Expectations](#expectations)
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

### Suite[T].Validate(rows []T) \*Report

Runs all expectations in the suite against the provided data and returns a
report.

```go
report := suite.Validate(users)
```

### Suite[T].WithSampleCap(cap int) \*Suite[T]

Returns a new suite with the specified sample cap for result reporting.

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

### Table-Level Expectations

Functions for dataset-level validations:

- `RowCount[T any](name string, pred func(int) bool) Expectation[T]` - Custom
  row count predicate
- `RowCountBetween[T any](lo, hi int) Expectation[T]` - Row count between bounds
- `RowCountEqual[T any](want int) Expectation[T]` - Exact row count

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

- `Name string` - Expectation name
- `Column string` - Column label (empty for row/table-level)
- `Success bool` - Whether the expectation passed
- `Total int` - Total rows evaluated
- `FailedCount int` - Number of rows that failed
- `FailedPercent float64` - Percentage of rows that failed
- `SampleValues []any` - Sample of failing values (capped)
- `FailedIndices []int` - Indices of all failing rows (complete)
- `Err error` - Error during evaluation (if any)

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
