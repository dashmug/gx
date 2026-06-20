# Custom Expectations

<!-- omit in toc -->

How to extend gx with your own validation logic.

- [Implementing the Expectation Interface](#implementing-the-expectation-interface)
- [Using the evalColumn Helper](#using-the-evcolumn-helper)
- [Best Practices](#best-practices)
- [Examples](#examples)

## Implementing the Expectation Interface

The core extension point for custom validations is the `Expectation[T]`
interface:

```go
type Expectation[T any] interface {
    Name() string
    Evaluate(rows []T, opts EvalOptions) Result
}
```

Your custom expectation must implement both methods:

- `Name() string` - Returns a descriptive name for the expectation that appears
  in reports
- `Evaluate(rows []T, opts EvalOptions) Result` - Performs the validation logic
  and returns a `Result`

Here's a basic example of a custom expectation that checks if user emails are
from a trusted domain:

```go
type TrustedDomainExpectation[T any] struct {
    name   string
    column string
    get    func(T) string  // accessor function for email field
    domain string
}

func (e TrustedDomainExpectation[T]) Name() string {
    return e.name
}

func (e TrustedDomainExpectation[T]) Evaluate(rows []T, opts gx.EvalOptions) gx.Result {
    return gx.EvalColumn(e.name, e.column, rows, e.get, func(email string) bool {
        return strings.HasSuffix(email, "@"+e.domain)
    }, opts)
}
```

## Using the evalColumn Helper

Most custom expectations will leverage the `evalColumn` helper function, which
handles row iteration, result aggregation, and common reporting:

```go
func evalColumn[T, V any](name, column string, rows []T,
    get func(T) V, pred func(V) bool, opts EvalOptions) Result
```

Arguments:

- `name` - Name for the expectation report
- `column` - Column label for the expectation (appears in reports)
- `rows` - Data to evaluate
- `get` - Accessor function that extracts the value to check from each row
- `pred` - Predicate function that returns true for valid values, false for
  invalid ones
- `opts` - Evaluation options (sample caps, etc.)

The helper:

- Iterates through all rows
- Applies the getter function to extract values
- Applies the predicate to check validity
- Collects indices of failed rows in `FailedIndices`
- Samples failed values in `SampleValues`
- Calculates statistics (counts, percentages)

## Best Practices

### 1. Use Descriptive Names

Choose names that clearly explain what the expectation checks:

```go
// Good
name := fmt.Sprintf("email from trusted domains (%s)", domainList)

// Avoid
name := "email check"
```

### 2. Implement Row-Level, Not Table-Level Logic

`Expectation[T]` is designed for row-by-row validation. For table-level checks,
use the row count expectations or implement a simple expectation:

```go
// Row level - check each user individually
func ValidUserExpectation(get func(User) bool) gx.Expectation[User] {
    return gx.Row("valid user", get)
}

// Table level - check total row count
func HasMinRows(min int) gx.Expectation[User] {
    return gx.RowCountBetween[User](min, math.MaxInt)
}
```

### 3. Handle Edge Cases Gracefully

Account for empty inputs, nil values, and other edge cases:

```go
func (e CustomExpectation[T]) Evaluate(rows []T, opts gx.EvalOptions) gx.Result {
    // Handle empty inputs gracefully
    if len(rows) == 0 {
        return gx.Result{
            Name:    e.name,
            Column:  e.column,
            Success: true,  // Vacuous success
            Total:   0,
        }
    }

    // Your validation logic here
}
```

### 4. Provide Meaningful Error Messages

Set the `Err` field in the `Result` when your evaluation encounters issues:

```go
func (e CustomExpectation[T]) Evaluate(rows []T, opts gx.EvalOptions) (res gx.Result) {
    defer func() {
        if r := recover(); r != nil {
            res.Err = fmt.Errorf("panic during validation: %v", r)
        }
    }()

    // Validation logic that might panic

    return res
}
```

### 5. Leverage Existing Builders When Possible

Before implementing a custom expectation, see if you can compose existing ones:

```go
// Instead of a custom expectation
func IsReasonableAge(user User) bool {
    return user.Age >= 0 && user.Age <= 120
}

// Compose existing expectations
gx.Ordered("age", func(u User) int { return u.Age }).Between(0, 120)
```

## Examples

### Example 1: Email Domain Validation

Validates that all email addresses come from a specified domain:

```go
type emailDomainExpectation[T any] struct {
    name   string
    column string
    get    func(T) string
    domain string
}

func (e emailDomainExpectation[T]) Name() string {
    return e.name
}

func (e emailDomainExpectation[T]) Evaluate(rows []T, opts gx.EvalOptions) gx.Result {
    return gx.EvalColumn(
        e.name,
        e.column,
        rows,
        e.get,
        func(email string) bool {
            return strings.HasSuffix(strings.ToLower(email), "@"+strings.ToLower(e.domain))
        },
        opts,
    )
}

func EmailDomain[T any](name string, get func(T) string, domain string) gx.Expectation[T] {
    return emailDomainExpectation[T]{
        name:   fmt.Sprintf("%s is from domain %s", name, domain),
        column: name,
        get:    get,
        domain: domain,
    }
}

// Usage
suite := gx.NewSuite[User](
    EmailDomain("email", func(u User) string { return u.Email }, "company.com"),
)
```

### Example 2: Password Strength Validation

Verifies password complexity requirements:

```go
type passwordStrengthExpectation[T any] struct {
    name   string
    column string
    get    func(T) string
    minLen int
}

func (e passwordStrengthExpectation[T]) Name() string {
    return e.name
}

func (e passwordStrengthExpectation[T]) Evaluate(rows []T, opts gx.EvalOptions) gx.Result {
    return gx.EvalColumn(
        e.name,
        e.column,
        rows,
        e.get,
        func(password string) bool {
            if len(password) < e.minLen {
                return false
            }

            hasUpper := strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
            hasLower := strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyz")
            hasDigit := strings.ContainsAny(password, "0123456789")
            hasSpecial := strings.ContainsAny(password, "!@#$%^&*()_+-=[]{}|;:,.<>?")

            return hasUpper && hasLower && hasDigit && hasSpecial
        },
        opts,
    )
}

func StrongPassword[T any](name string, get func(T) string, minLen int) gx.Expectation[T] {
    return passwordStrengthExpectation[T]{
        name:   fmt.Sprintf("%s is strong (min %d chars)", name, minLen),
        column: name,
        get:    get,
        minLen: minLen,
    }
}

// Usage
suite := gx.NewSuite[User](
    StrongPassword("password", func(u User) string { return u.Password }, 8),
)
```

### Example 3: Business Rule Validation

Enforces a complex business rule with cross-field validation:

```go
type orderValidationExpectation[T any] struct {
    name string
    get  func(T) Order  // Assuming Order is a struct with Amount and Discount fields
}

func (e orderValidationExpectation[T]) Name() string {
    return e.name
}

func (e orderValidationExpectation[T]) Evaluate(rows []T, opts gx.EvalOptions) gx.Result {
    return gx.EvalColumn(
        e.name,
        "",  // No specific column for cross-field validations
        rows,
        e.get,
        func(order Order) bool {
            // Business rule: discount cannot exceed 50% of amount
            if order.Amount <= 0 {
                return false
            }
            maxDiscount := order.Amount * 0.5
            return order.Discount <= maxDiscount
        },
        opts,
    )
}

func ValidDiscount[T any](name string, get func(T) Order) gx.Expectation[T] {
    return orderValidationExpectation[T]{
        name: fmt.Sprintf("%s has valid discount", name),
        get:  get,
    }
}

// Usage
suite := gx.NewSuite[Transaction](
    ValidDiscount("order", func(t Transaction) Order { return t.OrderDetails }),
)
```

### Example 4: Custom Error Handling

Shows how to properly handle and report validation errors:

```go
type parsingExpectation[T any] struct {
    name   string
    column string
    get    func(T) string
}

func (e parsingExpectation[T]) Name() string {
    return e.name
}

func (e parsingExpectation[T]) Evaluate(rows []T, opts gx.EvalOptions) (res gx.Result) {
    // Capture panics and convert them to errors
    defer func() {
        if r := recover(); r != nil {
            res.Err = fmt.Errorf("failed to parse value: %v", r)
        }
    }()

    return gx.EvalColumn(
        e.name,
        e.column,
        rows,
        e.get,
        func(value string) bool {
            _, err := strconv.ParseFloat(value, 64)
            return err == nil
        },
        opts,
    )
}

func ParsableNumber[T any](name string, get func(T) string) gx.Expectation[T] {
    return parsingExpectation[T]{
        name:   fmt.Sprintf("%s is a valid number", name),
        column: name,
        get:    get,
    }
}

// Usage
suite := gx.NewSuite[Record](
    ParsableNumber("price", func(r Record) string { return r.PriceString }),
)
```

Remember that while the framework guarantees that every expectation is called,
it's your responsibility to ensure your implementation is correct, performant,
and integrates well with the gx ecosystem.
