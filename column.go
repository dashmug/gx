package gx

import (
	"cmp"
	"fmt"
	"regexp"
	"unicode/utf8"
)

// OrderedColumn is a typed accessor over an ordered field (integers, floats,
// strings). Build it with Ordered.
type OrderedColumn[T any, V cmp.Ordered] struct {
	name string
	get  func(T) V
}

// Ordered creates a column over an ordered field. name is a report label and
// may differ from the struct field, so computed values are first-class.
func Ordered[T any, V cmp.Ordered](name string, get func(T) V) OrderedColumn[T, V] {
	return OrderedColumn[T, V]{name: name, get: get}
}

// Between asserts lo <= value <= hi (inclusive).
func (c OrderedColumn[T, V]) Between(lo, hi V) Expectation[T] {
	return newCol(
		fmt.Sprintf("%s between [%v,%v]", c.name, lo, hi),
		c.name, c.get,
		func(v V) bool { return v >= lo && v <= hi },
	)
}

// GreaterThan asserts value > bound.
func (c OrderedColumn[T, V]) GreaterThan(bound V) Expectation[T] {
	return newCol(
		fmt.Sprintf("%s > %v", c.name, bound),
		c.name, c.get,
		func(v V) bool { return v > bound },
	)
}

// LessThan asserts value < bound.
func (c OrderedColumn[T, V]) LessThan(bound V) Expectation[T] {
	return newCol(
		fmt.Sprintf("%s < %v", c.name, bound),
		c.name, c.get,
		func(v V) bool { return v < bound },
	)
}

// GreaterOrEqual asserts value >= bound.
func (c OrderedColumn[T, V]) GreaterOrEqual(bound V) Expectation[T] {
	return newCol(
		fmt.Sprintf("%s >= %v", c.name, bound),
		c.name, c.get,
		func(v V) bool { return v >= bound },
	)
}

// LessOrEqual asserts value <= bound.
func (c OrderedColumn[T, V]) LessOrEqual(bound V) Expectation[T] {
	return newCol(
		fmt.Sprintf("%s <= %v", c.name, bound),
		c.name, c.get,
		func(v V) bool { return v <= bound },
	)
}

// In asserts the value is one of vals.
func (c OrderedColumn[T, V]) In(vals ...V) Expectation[T] {
	return inSet(fmt.Sprintf("%s in %v", c.name, vals), c.name, c.get, vals)
}

// NotIn asserts the value is not one of vals.
func (c OrderedColumn[T, V]) NotIn(vals ...V) Expectation[T] {
	return notInSet(fmt.Sprintf("%s not in %v", c.name, vals), c.name, c.get, vals)
}

// NotZero asserts the value is not the zero value of its type.
func (c OrderedColumn[T, V]) NotZero() Expectation[T] {
	var zero V
	return newCol(c.name+" not zero", c.name, c.get, func(v V) bool { return v != zero })
}

// Zero asserts the value is the zero value of its type.
func (c OrderedColumn[T, V]) Zero() Expectation[T] {
	var zero V
	return newCol(c.name+" zero", c.name, c.get, func(v V) bool { return v == zero })
}

// Satisfy asserts the value matches a custom predicate; check names the rule.
func (c OrderedColumn[T, V]) Satisfy(check string, pred func(V) bool) Expectation[T] {
	return newCol(c.name+": "+check, c.name, c.get, pred)
}

// Unique asserts every value in the column is distinct.
func (c OrderedColumn[T, V]) Unique() Expectation[T] {
	return unique(c.name, c.get)
}

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

// NotEmpty asserts the string is non-empty (the string-friendly alias of NotZero).
func (c StringColumn[T]) NotEmpty() Expectation[T] {
	return newCol(c.name+" not empty", c.name, c.get, func(v string) bool { return v != "" })
}

// NotMatchRegex asserts the value does not match re.
func (c StringColumn[T]) NotMatchRegex(re *regexp.Regexp) Expectation[T] {
	return newCol(
		fmt.Sprintf("%s does not match /%s/", c.name, re.String()),
		c.name, c.get,
		func(v string) bool { return !re.MatchString(v) },
	)
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

// ComparableColumn is a typed accessor over a comparable but not necessarily
// ordered field (bools, enums, struct keys). Build it with Comparable.
type ComparableColumn[T any, V comparable] struct {
	name string
	get  func(T) V
}

// Comparable creates a column over a comparable field.
func Comparable[T any, V comparable](name string, get func(T) V) ComparableColumn[T, V] {
	return ComparableColumn[T, V]{name: name, get: get}
}

// In asserts the value is one of vals.
func (c ComparableColumn[T, V]) In(vals ...V) Expectation[T] {
	return inSet(fmt.Sprintf("%s in %v", c.name, vals), c.name, c.get, vals)
}

// NotIn asserts the value is not one of vals.
func (c ComparableColumn[T, V]) NotIn(vals ...V) Expectation[T] {
	return notInSet(fmt.Sprintf("%s not in %v", c.name, vals), c.name, c.get, vals)
}

// NotZero asserts the value is not the zero value of its type.
func (c ComparableColumn[T, V]) NotZero() Expectation[T] {
	var zero V
	return newCol(c.name+" not zero", c.name, c.get, func(v V) bool { return v != zero })
}

// Zero asserts the value is the zero value of its type.
func (c ComparableColumn[T, V]) Zero() Expectation[T] {
	var zero V
	return newCol(c.name+" zero", c.name, c.get, func(v V) bool { return v == zero })
}

// Satisfy asserts the value matches a custom predicate; check names the rule.
func (c ComparableColumn[T, V]) Satisfy(check string, pred func(V) bool) Expectation[T] {
	return newCol(c.name+": "+check, c.name, c.get, pred)
}

// Unique asserts every value in the column is distinct.
func (c ComparableColumn[T, V]) Unique() Expectation[T] {
	return unique(c.name, c.get)
}

// FieldColumn is a typed accessor over any field. It offers only Satisfy — the
// escape hatch for types that are neither ordered nor comparable. Build with Field.
type FieldColumn[T any, V any] struct {
	name string
	get  func(T) V
}

// Field creates a column over any field type.
func Field[T any, V any](name string, get func(T) V) FieldColumn[T, V] {
	return FieldColumn[T, V]{name: name, get: get}
}

// Satisfy asserts the value matches a custom predicate; check names the rule.
func (c FieldColumn[T, V]) Satisfy(check string, pred func(V) bool) Expectation[T] {
	return newCol(c.name+": "+check, c.name, c.get, pred)
}
