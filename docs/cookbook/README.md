# Cookbook

Practical examples and patterns for using `gx` in real-world scenarios.

## Table of Contents

- [Data Quality Audits](#data-quality-audits)
- [ETL Pipeline Validation](#etl-pipeline-validation)
- [API Data Validation](#api-data-validation)
- [Database Migration Checks](#database-migration-checks)
- [Custom Business Rules](#custom-business-rules)

## Data Quality Audits

Use `gx` to perform comprehensive data quality audits on your datasets:

```go
func auditUserData(users []User) *gx.Report {
    emailRE := regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

    suite := gx.NewSuite[User](
        // Basic presence checks
        gx.Str("id", func(u User) string { return u.ID }).NotEmpty(),
        gx.Str("email", func(u User) string { return u.Email }).NotEmpty(),
        gx.Str("name", func(u User) string { return u.Name }).NotEmpty(),

        // Format validations
        gx.Str("email", func(u User) string { return u.Email }).MatchRegex(emailRE),

        // Range checks
        gx.Ordered("age", func(u User) int { return u.Age }).Between(0, 120),

        // Cross-field validations
        gx.Row("minor with parental consent", func(u User) bool {
            if u.Age < 18 {
                return u.HasParentalConsent
            }
            return true
        }),

        // Uniqueness constraints
        gx.Str("email", func(u User) string { return u.Email }).Unique(),
        gx.Str("id", func(u User) string { return u.ID }).Unique(),
    )

    return suite.Validate(users)
}

// Usage
func main() {
    users := loadUserData()
    report := auditUserData(users)

    if !report.OK() {
        fmt.Printf("Found %d data quality issues:\n", len(report.Failures()))
        for _, result := range report.Failures() {
            fmt.Printf("- %s: %d records failed\n", result.Name, result.FailedCount)
            if len(result.SampleValues) > 0 {
                fmt.Printf("  Sample: %v\n", result.SampleValues[:min(3, len(result.SampleValues))])
            }
        }

        // Identify specific bad records
        for _, result := range report.Failures() {
            fmt.Printf("\nRecords failing '%s':\n", result.Name)
            for _, idx := range result.FailedIndices {
                fmt.Printf("  Row %d: %+v\n", idx, users[idx])
            }
        }
    }
}
```

## ETL Pipeline Validation

Validate data at each stage of your ETL pipeline:

```go
type ETLPipeline struct {
    validator *gx.Suite[Record]
}

func NewETLPipeline() *ETLPipeline {
    suite := gx.NewSuite[Record](
        // Source validation
        gx.Str("source_id", func(r Record) string { return r.SourceID }).NotEmpty(),

        // Transformation validation
        gx.Ordered("amount", func(r Record) float64 { return r.Amount }).GreaterOrEqual(0),
        gx.Str("currency", func(r Record) string { return r.Currency }).In("USD", "EUR", "GBP"),

        // Destination validation
        gx.Row("consistent totals", func(r Record) bool {
            return math.Abs(r.Debit-r.Credit-r.Adjustment) < 0.01
        }),
    )

    return &ETLPipeline{validator: suite}
}

func (p *ETLPipeline) ProcessBatch(records []Record) ([]Record, []int, error) {
    report := p.validator.Validate(records)

    if !report.OK() {
        // Log issues
        for _, failure := range report.Failures() {
            log.Printf("ETL validation failed: %s (%d failures)",
                failure.Name, failure.FailedCount)
        }

        // Return bad indices for further processing
        var allBadIndices []int
        for _, failure := range report.Failures() {
            allBadIndices = append(allBadIndices, failure.FailedIndices...)
        }

        // Remove duplicates and sort
        sort.Ints(allBadIndices)
        uniqueBad := removeDuplicates(allBadIndices)

        return records, uniqueBad, fmt.Errorf("validation failed")
    }

    return records, nil, nil
}
```

## API Data Validation

Validate incoming data from external APIs:

```go
func validateAPIResponse(data []ExternalUser) (*gx.Report, []ExternalUser, []ExternalUser) {
    phoneRE := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)

    suite := gx.NewSuite[ExternalUser](
        // Required fields
        gx.Str("user_id", func(eu ExternalUser) string { return eu.UserID }).NotEmpty(),
        gx.Str("email", func(eu ExternalUser) string { return eu.Email }).NotEmpty(),

        // Format validations
        gx.Str("email", func(eu ExternalUser) string { return eu.Email }).MatchRegex(emailRE),
        gx.Str("phone", func(eu ExternalUser) string { return eu.Phone }).MatchRegex(phoneRE),

        // Business rule validations
        gx.Ordered("created_at", func(eu ExternalUser) time.Time { return eu.CreatedAt }).
            LessOrEqual(time.Now()),
    )

    report := suite.Validate(data)

    // Separate good and bad records
    var goodRecords, badRecords []ExternalUser
    badIndices := make(map[int]bool)

    for _, failure := range report.Failures() {
        for _, idx := range failure.FailedIndices {
            badIndices[idx] = true
        }
    }

    for i, record := range data {
        if badIndices[i] {
            badRecords = append(badRecords, record)
        } else {
            goodRecords = append(goodRecords, record)
        }
    }

    return report, goodRecords, badRecords
}
```

## Database Migration Checks

Verify data integrity after database migrations:

```go
func validateMigration(oldData, newData []User) *gx.Report {
    suite := gx.NewSuite[MigrationCheck](
        // Ensure no data loss
        gx.RowCountEqual[MigrationCheck](len(oldData)),

        // Check critical field preservation
        gx.Row("ids preserved", func(mc MigrationCheck) bool {
            return mc.Old.ID == mc.New.ID
        }),

        gx.Row("emails preserved", func(mc MigrationCheck) bool {
            return mc.Old.Email == mc.New.Email
        }),

        // Check data transformations
        gx.Row("status mapping correct", func(mc MigrationCheck) bool {
            expectedStatus := mapOldToNewStatus(mc.Old.Status)
            return expectedStatus == mc.New.Status
        }),
    )

    // Create paired data for comparison
    pairedData := make([]MigrationCheck, len(oldData))
    for i := range oldData {
        pairedData[i] = MigrationCheck{
            Old: oldData[i],
            New: newData[i],
        }
    }

    return suite.Validate(pairedData)
}

type MigrationCheck struct {
    Old User
    New User
}
```

## Custom Business Rules

Implement domain-specific validation logic:

```go
// Custom expectation for credit score validation
type creditScoreExpectation[T any] struct {
    name   string
    get    func(T) Applicant
}

func (e creditScoreExpectation[T]) Name() string {
    return e.name
}

func (e creditScoreExpectation[T]) Evaluate(rows []T, opts gx.EvalOptions) gx.Result {
    return gx.EvalColumn(
        e.name,
        "credit_applicant",
        rows,
        e.get,
        func(applicant Applicant) bool {
            // Complex business logic for credit approval
            if applicant.CreditScore >= 700 {
                return true  // Automatic approval
            }

            if applicant.CreditScore >= 650 && applicant.Income > 50000 {
                return true  // Approval with income check
            }

            if applicant.CreditScore >= 600 && applicant.Income > 75000 &&
               applicant.DebtToIncomeRatio < 0.3 {
                return true  // Approval with additional checks
            }

            return false  // Reject
        },
        opts,
    )
}

func CreditApproval[T any](name string, get func(T) Applicant) gx.Expectation[T] {
    return creditScoreExpectation[T]{name: name, get: get}
}

// Usage in loan processing
func validateLoanApplications(applications []LoanApplication) *gx.Report {
    suite := gx.NewSuite[LoanApplication](
        gx.Ordered("loan_amount", func(la LoanApplication) float64 {
            return la.LoanAmount
        }).Between(1000, 1000000),

        CreditApproval("applicant approval", func(la LoanApplication) Applicant {
            return la.Applicant
        }),

        gx.Row("loan purpose consistency", func(la LoanApplication) bool {
            // Business rules about loan purposes
            switch la.Purpose {
            case "mortgage":
                return la.LoanAmount >= 50000
            case "auto":
                return la.LoanAmount >= 5000 && la.LoanAmount <= 100000
            case "personal":
                return la.LoanAmount >= 1000 && la.LoanAmount <= 50000
            default:
                return false
            }
        }),
    )

    return suite.Validate(applications)
}
```

## Best Practices from Examples

1. **Group related validations** in suites for better organization
2. **Use descriptive names** that clearly indicate what each expectation checks
3. **Separate good and bad records** for further processing
4. **Log validation failures** for monitoring and debugging
5. **Combine built-in and custom expectations** for comprehensive validation
6. **Validate at process boundaries** to catch issues early
7. **Preserve row indices** to enable actionable error reporting
