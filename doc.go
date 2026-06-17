// Package gx validates an in-memory slice of structs against declarative,
// type-safe expectations and returns a rich pass/fail report.
//
// Build a suite from typed column accessors and run it over your data:
//
//	suite := gx.NewSuite[User](
//		gx.Ordered("age", func(u User) int { return u.Age }).Between(0, 120),
//		gx.Str("email", func(u User) string { return u.Email }).MatchRegex(emailRE),
//		gx.Comparable("id", func(u User) string { return u.ID }).Unique(),
//	)
//	if err := suite.Validate(users).Err(); err != nil {
//		// gate the pipeline; err is a *gx.ValidationError carrying the Report
//	}
//
// Validation is collect-all: every expectation runs and the Report holds a
// Result per expectation, including failing row indices for quarantining bad
// records. For tests, use the gxtest sub-package.
package gx
