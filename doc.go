// Package gx validates an in-memory slice of structs against declarative,
// type-safe expectations and returns a rich pass/fail report.
//
// See https://github.com/dashmug/gx/blob/main/docs/README.md for guides,
// built-in expectation reference, and API details.
//
// Build a suite from typed column accessors and run it over your data:
//
//	suite := gx.NewSuite[User](
//		gx.Ordered("age", func(u User) int { return u.Age }).Between(0, 120),
//		gx.Str("email", func(u User) string { return u.Email }).MatchRegex(emailRE),
//	)
//	if err := suite.Validate(users).Err(); err != nil {
//		// gate the pipeline; err is a *gx.ValidationError carrying the Report
//	}
//
// Validation is collect-all: every expectation runs and the Report holds a
// Result per expectation. Per-row checks record the complete list of failing
// row indices in FailedIndices. Table-level RowCount* and Numeric aggregate
// expectations use Total=0 and leave per-row fields empty; see Result and the
// docs for naming and reporting semantics. Tests use the same suite via the
// gxtest sub-package.
package gx
