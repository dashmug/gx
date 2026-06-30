# Core Concepts

`gx` is a data quality validation library for Go that helps you validate
collections of data with comprehensive reporting. Unlike traditional assertion
libraries or schema validators, `gx` is designed specifically for validating
batches of data and reporting all issues at once.

## What is gx?

`gx` stands for "Go eXpectations" and provides type-safe, declarative data
quality validation for Go applications. It validates a `[]T` directly and gives
you a rich pass/fail report, including the complete list of failing row indices
so you can quarantine bad records instead of just learning that _something_
failed.

## Key Concepts

### Dataset

A collection of homogeneous data records represented as a `[]T` slice in Go.

### Row

A single element of type `T` within a Dataset.

### Suite

A validated set of expectations applied to a Dataset. Created with
`gx.NewSuite[T](...)`.

### Expectation

A single validation rule that can be applied to a Dataset. Expectations are
composable and can be column-based, row-based, or table-based validations.

### Report

The result of validating a Dataset against a Suite of expectations. Contains
detailed information about which expectations passed or failed.

### Result

Details about a single expectation's validation outcome, including success
status, counts, sample values, and indices of all failing rows.

## Basic Usage

```go
suite := gx.NewSuite[User](
    gx.Ordered("age", func(u User) int { return u.Age }).Between(0, 120),
    gx.Str("email", func(u User) string { return u.Email }).MatchRegex(emailRE),
    gx.Comparable("id", func(u User) string { return u.ID }).Unique(),
)

report := suite.Validate(users)
if !report.OK() {
    // Handle validation failures
    for _, result := range report.Failures() {
        fmt.Printf("%s: %d/%d failed\n", result.Name, result.FailedCount, result.Total)
    }
}
```

## When to Use gx

Use `gx` when you need to:

- Validate batches of data for quality issues
- Get comprehensive reporting on all validation failures
- Identify exactly which records have problems
- Quarantine or process bad data records separately
- Perform data quality audits

Don't use `gx` when you need to:

- Validate a single object in a request handler (use a schema validator)
- Assert conditions in unit tests (use testify or standard library)
- Validate complex nested document structures (use JSON Schema)

## Operational Limits

`gx` validates data entirely in process. Large datasets and widespread failures
can consume significant memory when full failure metadata is retained.

| Control | Default | Effect |
| ------- | ------- | ------ |
| `WithSampleCap(n)` | 20 | Caps offending values in `SampleValues`; zero collects none |
| `WithFailedIndicesCap(n)` | 100 | Caps `FailedIndices`; pass zero for unlimited |

For production pipeline gates on large batches, the default caps bound memory.
Pass `WithFailedIndicesCap(0)` when every failing row index is required for
remediation.

`Result.String()` and `Report.String()` embed sample values. Redact or avoid
logging reports when column values may contain PII or secrets.

For database-backed validation at scale, use the sibling [`gxsql`](../../gxsql/docs/)
library instead.
