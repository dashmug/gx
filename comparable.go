package gx

import "fmt"

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

// In asserts the value is one of vals. An empty vals list fails every row
// (nothing is in the empty set). gxsql rejects empty In at configuration time.
func (c ComparableColumn[T, V]) In(vals ...V) Expectation[T] {
	return inSet(fmt.Sprintf("%s in %v", c.name, vals), c.name, c.get, vals)
}

// NotIn asserts the value is not one of vals. An empty vals list passes every
// row vacuously (no forbidden values). gxsql rejects empty NotIn at
// configuration time.
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
