# gx Documentation

Welcome to the `gx` documentation. `gx` is a type-safe, declarative data quality
validation library for Go.

## Table of Contents

- [Core Concepts](concepts/) - Learn what `gx` is and how to use it
- [Built-in Expectations](expectations/) - Complete reference of all validation
  functions
- [Custom Expectations](custom/) - How to extend `gx` with your own validation
  logic
- [Comparisons](comparison/) - How `gx` differs from other validation approaches

- [Cookbook](cookbook/) - Practical examples and patterns for real-world usage

- [API Reference](reference/) - Complete reference for all public types and
  functions

## Quick Start

```go
suite := gx.NewSuite[User](
    gx.Ordered("age", func(u User) int { return u.Age }).Between(0, 120),
    gx.Str("email", func(u User) string { return u.Email }).MatchRegex(emailRE),
    gx.Comparable("id", func(u User) string { return u.ID }).Unique(),
)

report := suite.Validate(users)
if !report.OK() {
    // Handle validation failures
}
```

## Key Features

- **Type-safe** - Field access is checked by the compiler
- **Collect-all** - All expectations run, all failures reported at once
- **Actionable results** - Complete list of failing row indices
- **Zero dependencies** - Standard library only
- **Test-friendly** - Integration with Go's testing package

## Next Steps

1. Read [Core Concepts](concepts/) to understand how `gx` works
2. Browse [Built-in Expectations](expectations/) to see what validations are
   available
3. Check [Custom Expectations](custom/) if you need domain-specific validations
4. Review [Comparisons](comparison/) to understand when to use `gx` vs
   alternatives
