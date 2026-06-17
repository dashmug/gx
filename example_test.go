package gx_test

import (
	"fmt"
	"regexp"

	"github.com/dashmug/gx"
)

func Example() {
	type User struct {
		Age   int
		Email string
	}
	users := []User{
		{Age: 30, Email: "a@example.com"},
		{Age: 200, Email: "bad"},
	}
	emailRE := regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

	suite := gx.NewSuite[User](
		gx.Ordered("age", func(u User) int { return u.Age }).Between(0, 120),
		gx.Str("email", func(u User) string { return u.Email }).MatchRegex(emailRE),
	)

	report := suite.Validate(users)
	fmt.Println(report.OK())
	for _, r := range report.Failures() {
		fmt.Printf("%s: %d/%d failed at %v\n", r.Name, r.FailedCount, r.Total, r.FailedIndices)
	}
	// Output:
	// false
	// age between [0,120]: 1/2 failed at [1]
	// email matches /^[^@\s]+@[^@\s]+\.[^@\s]+$/: 1/2 failed at [1]
}
