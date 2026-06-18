package gx_test

import (
	"fmt"

	"github.com/dashmug/gx"
)

func Example() {
	type User struct {
		Age int
	}

	users := []User{{Age: 10}, {Age: 200}}
	suite := gx.NewSuite(
		gx.Ordered("age", func(u User) int { return u.Age }).Between(0, 120),
		gx.RowCount[User]("at least two rows", func(n int) bool { return n >= 2 }),
	)

	report := suite.Validate(users)
	fmt.Println(report.OK())
	for _, r := range report.Failures() {
		fmt.Println(r.Name)
	}

	// Output:
	// false
	// age between [0,120]
}
