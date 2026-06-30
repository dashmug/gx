package gx

import (
	"cmp"
	"fmt"
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

// In asserts the value is one of vals. An empty vals list fails every row
// (nothing is in the empty set). gxsql rejects empty In at configuration time.
func (c OrderedColumn[T, V]) In(vals ...V) Expectation[T] {
	return inSet(fmt.Sprintf("%s in %v", c.name, vals), c.name, c.get, vals)
}

// NotIn asserts the value is not one of vals. An empty vals list passes every
// row vacuously (no forbidden values). gxsql rejects empty NotIn at
// configuration time.
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
