package gx

import (
	"fmt"
	"regexp"
	"unicode/utf8"
)

// StringColumn is a typed accessor over a string field. It embeds
// OrderedColumn[T, string], so Between, In, NotZero and Satisfy are available
// too, plus the string-specific checks below. Build it with Str.
type StringColumn[T any] struct {
	OrderedColumn[T, string]
}

// Str creates a column over a string field.
func Str[T any](name string, get func(T) string) StringColumn[T] {
	return StringColumn[T]{OrderedColumn[T, string]{name: name, get: get}}
}

// MatchRegex asserts the value matches re.
func (c StringColumn[T]) MatchRegex(re *regexp.Regexp) Expectation[T] {
	return newCol(
		fmt.Sprintf("%s matches /%s/", c.name, re.String()),
		c.name, c.get,
		func(v string) bool { return re.MatchString(v) },
	)
}

// NotMatchRegex asserts the value does not match re.
func (c StringColumn[T]) NotMatchRegex(re *regexp.Regexp) Expectation[T] {
	return newCol(
		fmt.Sprintf("%s does not match /%s/", c.name, re.String()),
		c.name, c.get,
		func(v string) bool { return !re.MatchString(v) },
	)
}

// NotEmpty asserts the string is non-empty (the string-friendly alias of NotZero).
func (c StringColumn[T]) NotEmpty() Expectation[T] {
	return newCol(c.name+" not empty", c.name, c.get, func(v string) bool { return v != "" })
}

// Empty asserts the string is empty (complement of NotEmpty).
func (c StringColumn[T]) Empty() Expectation[T] {
	return newCol(c.name+" empty", c.name, c.get, func(v string) bool { return v == "" })
}

// LenBetween asserts lo <= rune count <= hi (inclusive).
func (c StringColumn[T]) LenBetween(lo, hi int) Expectation[T] {
	return newCol(
		fmt.Sprintf("%s length in [%d,%d]", c.name, lo, hi),
		c.name, c.get,
		func(v string) bool {
			l := utf8.RuneCountInString(v)
			return l >= lo && l <= hi
		},
	)
}

// LenEqual asserts rune count == n.
func (c StringColumn[T]) LenEqual(n int) Expectation[T] {
	return newCol(
		fmt.Sprintf("%s length == %d", c.name, n),
		c.name, c.get,
		func(v string) bool { return utf8.RuneCountInString(v) == n },
	)
}
