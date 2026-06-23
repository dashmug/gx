# Built-in Expectations

`gx` provides several built-in expectation types for common data validation
scenarios. Each expectation type is accessed through a typed column builder that
ensures type safety at compile time.

## Ordered Column Expectations

For ordered types like integers, floats, and strings. Accessed through
`gx.Ordered()`.

### Between

Asserts that `lo <= value <= hi` (inclusive).

```go
gx.Ordered("age", func(u User) int { return u.Age }).Between(0, 120)
```

Use when you need to constrain values within a range.

### GreaterThan

Asserts that `value > bound`.

```go
gx.Ordered("score", func(s Student) float64 { return s.Score }).GreaterThan(0.0)
```

Use for minimum thresholds.

### LessThan

Asserts that `value < bound`.

```go
gx.Ordered("discount", func(p Product) float64 { return p.Discount }).LessThan(1.0)
```

Use for maximum limits.

### GreaterOrEqual

Asserts that `value >= bound`.

```go
gx.Ordered("quantity", func(i Item) int { return i.Quantity }).GreaterOrEqual(0)
```

Use when zero is valid but negative values are not.

### LessOrEqual

Asserts that `value <= bound`.

```go
gx.Ordered("rating", func(r Review) float64 { return r.Rating }).LessOrEqual(5.0)
```

Use for upper bounds.

### In

Asserts that the value is one of the listed values.

```go
gx.Ordered("status", func(o Order) int { return o.Status }).In(1, 2, 3)
```

Use for enumerated numeric values.

### NotIn

Asserts that the value is not one of the listed values.

```go
gx.Ordered("code", func(e Error) int { return e.Code }).NotIn(500, 502, 503)
```

Use to exclude specific values.

### NotZero

Asserts that the value is not the zero value of its type.

```go
gx.Ordered("id", func(u User) int { return u.ID }).NotZero()
```

Use to require presence of values.

### Zero

Asserts that the value is the zero value of its type.

```go
gx.Ordered("deleted_at", func(r Record) int64 { return r.DeletedAt }).Zero()
```

Use to check for unset values.

### Satisfy

Asserts that the value matches a custom predicate.

```go
gx.Ordered("number", func(d Data) int { return d.Number }).Satisfy("even", func(v int) bool { return v%2 == 0 })
```

Use for custom validation logic.

### Unique

Asserts that every value in the column is distinct.

```go
gx.Ordered("email", func(u User) string { return u.Email }).Unique()
```

Use for enforcing uniqueness constraints.

## String Column Expectations

Specialized expectations for string fields. Includes all Ordered expectations
plus string-specific ones. Accessed through `gx.Str()`.

### MatchRegex

Asserts that the value matches a regular expression.

```go
emailRE := regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
gx.Str("email", func(u User) string { return u.Email }).MatchRegex(emailRE)
```

Use for pattern validation like emails, phone numbers, etc.

### NotMatchRegex

Asserts that the value does not match a regular expression.

```go
profanityRE := regexp.MustCompile(`(?i)(badword|offensive)`)
gx.Str("comment", func(c Comment) string { return c.Text }).NotMatchRegex(profanityRE)
```

Use for content filtering.

### NotEmpty

Asserts that the string is not empty.

```go
gx.Str("name", func(u User) string { return u.Name }).NotEmpty()
```

Use to require non-empty strings.

### Empty

Asserts that the string is empty.

```go
gx.Str("middle_name", func(u User) string { return u.MiddleName }).Empty()
```

Use to check for unset string values.

### LenBetween

Asserts that `lo <= rune count <= hi` (inclusive).

```go
gx.Str("password", func(u User) string { return u.Password }).LenBetween(8, 128)
```

Use for length constraints on text fields.

### LenEqual

Asserts that rune count equals n.

```go
gx.Str("country_code", func(a Address) string { return a.CountryCode }).LenEqual(2)
```

Use for fixed-length codes or identifiers.

## Comparable Column Expectations

For comparable but not necessarily ordered types like booleans, enums, and
structs. Accessed through `gx.Comparable()`.

### In

Asserts that the value is one of the listed values.

```go
gx.Comparable("status", func(o Order) Status { return o.Status }).In(Active, Pending)
```

Use for enum validation.

### NotIn

Asserts that the value is not one of the listed values.

```go
gx.Comparable("status", func(o Order) Status { return o.Status }).NotIn(Cancelled, Expired)
```

Use to exclude specific enum values.

### NotZero

Asserts that the value is not the zero value.

```go
gx.Comparable("category", func(p Product) Category { return p.Category }).NotZero()
```

Use to require presence of comparable values.

### Zero

Asserts that the value is the zero value.

```go
gx.Comparable("deleted_category", func(p Product) Category { return p.DeletedCategory }).Zero()
```

Use to check for unset comparable values.

### Satisfy

Asserts that the value matches a custom predicate.

```go
gx.Comparable("role", func(u User) Role { return u.Role }).Satisfy("privileged", func(r Role) bool {
    return r == Admin || r == Manager
})
```

Use for complex enum logic.

### Unique

Same as for Ordered columns.

## Field Column Expectations

Escape hatch for non-comparable types. Accessed through `gx.Field()`.

### Satisfy

Only method available for Field columns.

```go
gx.Field("tags", func(p Post) []string { return p.Tags }).Satisfy("non-empty", func(tags []string) bool {
    return len(tags) > 0
})
```

Use for custom validation on complex types.

## Row-Level Expectations

Cross-field validations that operate on entire rows. Accessed through
`gx.Row()`.

```go
gx.Row("ship date after order date", func(o Order) bool {
    return !o.ShipDate.Before(o.OrderDate)
})
```

Use when validation depends on relationships between fields.

## Numeric Column Aggregate Expectations

For integer and floating-point fields. Accessed through `gx.Numeric()`. Unlike
`Ordered`, these validate dataset-level statistics rather than individual rows.
They share table-level `Result` shape with `RowCount*`: `Total` is `0`, per-row
fields stay empty, and `Success` carries the verdict — but `Result.Column` is
the numeric accessor label.

After evaluation, `AverageBetween`, `MedianBetween`, and `StdDevBetween` append
the observed statistic to `Result.Name` (for example
`amount average in [10,100]: got 55`). `SatisfyAggregate` uses
`"<column>: <check>"` and does not append a `got` value. Any extracted `NaN` or
infinite float causes built-in aggregates to fail with `Success=false`,
`Err=nil`, and `Name` containing `got non-finite value`. Empty or nil input
passes vacuously; built-in aggregates do not call `SatisfyAggregate` predicates
on empty input.

### AverageBetween

Asserts that `lo <= average <= hi` (inclusive) over extracted numeric values.

```go
gx.Numeric("amount", func(o Order) float64 { return o.Amount }).AverageBetween(10, 100)
```

Use for batch-level mean checks such as average order value.

### MedianBetween

Asserts that `lo <= median <= hi` (inclusive). Even-length datasets use the
average of the two middle values.

```go
gx.Numeric("amount", func(o Order) float64 { return o.Amount }).MedianBetween(10, 100)
```

Use when the typical value matters more than the mean.

### StdDevBetween

Asserts that population standard deviation is in `[lo, hi]` (inclusive).

```go
gx.Numeric("amount", func(o Order) float64 { return o.Amount }).StdDevBetween(0, 25)
```

Use to bound spread. `NumericStats.StdDevSample` is available only in
`SatisfyAggregate` callbacks.

### SatisfyAggregate

Asserts that computed statistics match a custom predicate.

```go
gx.Numeric("amount", func(o Order) float64 { return o.Amount }).SatisfyAggregate(
    "coefficient of variation <= 0.25",
    func(s gx.NumericStats) bool {
        return s.Average == 0 || s.StdDevPopulation/s.Average <= 0.25
    },
)
```

Use for aggregate logic that built-in range checks cannot express. Empty or nil
input passes vacuously and does not call the predicate.

## Table-Level Expectations

Validations about the dataset as a whole. `RowCount*` helpers check `len(rows)`
only. They use table-level `Result` shape: `Column` is `""`, `Total` is `0`, and
per-row fields stay empty. `RowCountBetween`, `RowCountEqual`, and the threshold
helpers append the observed count to `Result.Name` (for example
`row count in [1,1000]: got 42`). Custom `RowCount(name, pred)` keeps the
caller-provided name.

`gx.Numeric()` aggregate expectations (see above) are also table-level but set
`Result.Column` to the accessor label.

### RowCountGreaterThan

Asserts that `len(rows) > bound`.

```go
gx.RowCountGreaterThan[User](0)
```

Use to require a non-empty dataset.

### RowCountGreaterOrEqual

Asserts that `len(rows) >= bound`.

```go
gx.RowCountGreaterOrEqual[User](1)
```

Use for minimum batch sizes.

### RowCountLessThan

Asserts that `len(rows) < bound`.

```go
gx.RowCountLessThan[Batch](10000)
```

Use for maximum batch sizes (exclusive).

### RowCountLessOrEqual

Asserts that `len(rows) <= bound`.

```go
gx.RowCountLessOrEqual[Batch](1000)
```

Use for maximum batch sizes (inclusive).

### RowCountBetween

Asserts that `lo <= len(rows) <= hi`.

```go
gx.RowCountBetween[User](1, 1000)
```

Use to validate batch size limits.

### RowCountEqual

Asserts that `len(rows) == want`.

```go
gx.RowCountEqual[Order](42)
```

Use when exact count is required.

### RowCount

Custom table-level validation with a predicate over row count.

```go
gx.RowCount[Transaction]("at least one", func(n int) bool { return n > 0 })
```

Use for custom row count validations.
