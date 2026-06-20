# Comparing gx with Other Validation Approaches

`gx` occupies a unique space in the validation landscape. Understanding how it
differs from other approaches will help you choose the right tool for each job.

## vs. Test Assertion Libraries (testify, etc.)

### Test Assertions

```go
// testify approach
assert.Equal(t, expected, actual)
assert.NoError(t, err)
```

**Characteristics:**

- Designed for unit testing single values
- Fail-fast behavior (stop on first failure)
- Tight integration with testing frameworks
- Primarily focused on program correctness, not data quality

### gx Approach

```go
// gx approach
suite := gx.NewSuite[User](
    gx.Ordered("age", func(u User) int { return u.Age }).Between(0, 120),
)
report := suite.Validate(users)
// All validations run, complete failure report provided
```

**Characteristics:**

- Designed for batch data validation
- Collect-all behavior (all validations run)
- Comprehensive failure reporting with row indices
- Focused on data quality assessment

### When to Use Each

**Use Test Assertions When:**

- Writing unit tests for functions and components
- Validating single values or small data structures
- You need integration with testing frameworks
- Program correctness is the primary concern

**Use gx When:**

- Validating batches of data from files, APIs, or databases
- You need to identify all data quality issues at once
- You want to quarantine or process bad records separately
- Performing data quality audits or analysis
- Building data processing pipelines

## vs. Schema Validation Libraries (jsonschema, etc.)

### Schema Validation

```json
{
  "type": "object",
  "properties": {
    "age": {
      "type": "integer",
      "minimum": 0,
      "maximum": 120
    }
  },
  "required": ["age"]
}
```

**Characteristics:**

- Declarative schema definition (often JSON/YAML)
- Validates structure and basic types
- Language-agnostic schemas
- Good for API request/response validation
- Typically validates one document at a time

### gx Approach

```go
gx.NewSuite[User](
    gx.Ordered("age", func(u User) int { return u.Age }).Between(0, 120),
    gx.Str("email", func(u User) string { return u.Email }).MatchRegex(emailRE),
    gx.Row("age-appropriate content", func(u User) bool {
        return u.Age >= 18 || !u.HasAdultContent
    }),
)
```

**Characteristics:**

- Programmatic validation definition (Go code)
- Validates business rules and cross-field relationships
- Go type system ensures compile-time safety
- Validates batches of homogeneous data
- Provides detailed failure analysis with row indices

### When to Use Each

**Use Schema Validation When:**

- Defining API contracts
- Validating JSON/YAML documents
- Need language-agnostic validation rules
- Validating individual document structure
- Working with non-Go systems

**Use gx When:**

- Validating Go structs in batches
- Need type-safe validation with compile-time checking
- Validating business rules and cross-field constraints
- Want comprehensive reporting on data quality issues
- Processing datasets where you need to identify bad records

## vs. Manual Validation Loops

### Manual Validation

```go
var errors []error
for i, user := range users {
    if user.Age < 0 || user.Age > 120 {
        errors = append(errors, fmt.Errorf("user[%d]: age out of range", i))
    }
    if !isValidEmail(user.Email) {
        errors = append(errors, fmt.Errorf("user[%d]: invalid email", i))
    }
}
if len(errors) > 0 {
    // Handle errors
}
```

**Characteristics:**

- Complete control over validation logic
- Ad-hoc error reporting format
- Easy to implement simple cases
- Difficult to maintain consistency
- No standardized reporting

### gx Approach

```go
suite := gx.NewSuite[User](
    gx.Ordered("age", func(u User) int { return u.Age }).Between(0, 120),
    gx.Str("email", func(u User) string { return u.Email }).MatchRegex(emailRE),
)
report := suite.Validate(users)
// Standardized report format with complete indices
```

**Characteristics:**

- Standardized validation framework
- Consistent, comprehensive reporting
- Separation of validation rules from execution
- Extensible with custom expectations
- Built-in handling of edge cases

### When to Use Each

**Use Manual Validation When:**

- Very simple, one-off validation needs
- Performance is absolutely critical
- You can't add dependencies
- Validation logic is truly ad-hoc

**Use gx When:**

- Consistent validation across multiple datasets
- You want standardized reporting format
- Need to identify specific bad records in large datasets
- Want reusable validation rules
- Building data quality tools or pipelines

## Decision Matrix

| Scenario                        | Recommended Approach          | Why                                      |
| ------------------------------- | ----------------------------- | ---------------------------------------- |
| Unit testing function output    | Test assertions               | Integrated with testing frameworks       |
| API request validation          | Schema validation             | Language-agnostic, standard contracts    |
| Batch data quality audit        | `gx`                          | Comprehensive reporting, row indices     |
| Cross-field business rules      | `gx`                          | Native support for row-level validations |
| Large dataset processing        | `gx`                          | Actionable failure indices               |
| Simple single-record validation | Schema validation or manual   | Overhead not justified                   |
| Complex domain validation       | `gx` with custom expectations | Extensible, type-safe                    |

## Hybrid Approaches

Many applications benefit from using multiple validation approaches:

1. **Schema validation at API boundaries** - Ensure basic structure and types
2. **gx for data processing pipelines** - Comprehensive quality checks with
   actionable results
3. **Test assertions for unit tests** - Verify component behavior

Example hybrid architecture:

```go
// 1. API layer - schema validation
func handleUserUpload(w http.ResponseWriter, r *http.Request) {
    var users []User
    if err := decodeAndValidateJSON(r.Body, &users); err != nil {
        http.Error(w, err.Error(), 400)
        return
    }

    // 2. Processing layer - gx validation
    suite := gx.NewSuite[User](/* ... expectations ... */)
    report := suite.Validate(users)
    if !report.OK() {
        // Log issues, quarantine bad records
        quarantineBadRecords(users, report)
    }

    // 3. Process only good records
    goodRecords := getGoodRecords(users, report)
    processRecords(goodRecords)
}
```
