package gx

import "fmt"

// rowExpectation checks a whole-row predicate (cross-field rules). Offending
// rows are boxed into SampleValues (capped); Column is "".
type rowExpectation[T any] struct {
	name string
	pred func(T) bool
}

func (e rowExpectation[T]) Name() string { return e.name }

func (e rowExpectation[T]) Evaluate(rows []T, opts EvalOptions) Result {
	res := Result{Name: e.name, Total: len(rows), Success: true}
	for i, row := range rows {
		if e.pred(row) {
			continue
		}
		res.Success = false
		res.FailedCount++
		res.FailedIndices = appendFailedIndex(res.FailedIndices, opts.FailedIndicesCap, i)
		if len(res.SampleValues) < opts.SampleCap {
			res.SampleValues = append(res.SampleValues, row)
		}
	}
	if res.Total > 0 {
		res.FailedPercent = float64(res.FailedCount) / float64(res.Total) * 100
	}
	return res
}

// Row builds a row-level expectation from a cross-field predicate.
func Row[T any](name string, pred func(T) bool) Expectation[T] {
	return rowExpectation[T]{name: name, pred: pred}
}

// rowCountExpectation checks the number of rows. Per-row Result fields are left
// empty; Total stays 0 because the check is table-level, not per-row.
type rowCountExpectation[T any] struct {
	label     string
	check     func(n int) bool
	reportGot bool
}

func (e rowCountExpectation[T]) Name() string { return e.label }

func (e rowCountExpectation[T]) Evaluate(rows []T, _ EvalOptions) Result {
	n := len(rows)
	name := e.label
	if e.reportGot {
		name = fmt.Sprintf("%s: got %d", e.label, n)
	}
	return Result{
		Name:    name,
		Total:   0,
		Success: e.check(n),
	}
}

// RowCount is the custom table-level escape hatch: it checks len(rows) with a
// predicate and leaves per-row Result fields empty (Total stays 0).
func RowCount[T any](name string, pred func(int) bool) Expectation[T] {
	return rowCountExpectation[T]{label: name, check: pred}
}

type rowCountBetweenExpectation[T any] struct {
	lo, hi int
}

func (e rowCountBetweenExpectation[T]) Name() string {
	return fmt.Sprintf("row count in [%d,%d]", e.lo, e.hi)
}

func (e rowCountBetweenExpectation[T]) Evaluate(rows []T, _ EvalOptions) Result {
	n := len(rows)
	return Result{
		Name:    fmt.Sprintf("row count in [%d,%d]: got %d", e.lo, e.hi, n),
		Total:   0,
		Success: n >= e.lo && n <= e.hi,
	}
}

// RowCountBetween asserts lo <= len(rows) <= hi (inclusive).
func RowCountBetween[T any](lo, hi int) Expectation[T] {
	return rowCountBetweenExpectation[T]{lo: lo, hi: hi}
}

type rowCountEqualExpectation[T any] struct {
	want int
}

func (e rowCountEqualExpectation[T]) Name() string {
	return fmt.Sprintf("row count == %d", e.want)
}

func (e rowCountEqualExpectation[T]) Evaluate(rows []T, _ EvalOptions) Result {
	n := len(rows)
	return Result{
		Name:    fmt.Sprintf("row count == %d: got %d", e.want, n),
		Total:   0,
		Success: n == e.want,
	}
}

// RowCountEqual asserts len(rows) == want.
func RowCountEqual[T any](want int) Expectation[T] {
	return rowCountEqualExpectation[T]{want: want}
}

// RowCountGreaterThan asserts len(rows) > bound.
func RowCountGreaterThan[T any](bound int) Expectation[T] {
	return rowCountExpectation[T]{
		label:     fmt.Sprintf("row count > %d", bound),
		check:     func(n int) bool { return n > bound },
		reportGot: true,
	}
}

// RowCountGreaterOrEqual asserts len(rows) >= bound.
func RowCountGreaterOrEqual[T any](bound int) Expectation[T] {
	return rowCountExpectation[T]{
		label:     fmt.Sprintf("row count >= %d", bound),
		check:     func(n int) bool { return n >= bound },
		reportGot: true,
	}
}

// RowCountLessThan asserts len(rows) < bound.
func RowCountLessThan[T any](bound int) Expectation[T] {
	return rowCountExpectation[T]{
		label:     fmt.Sprintf("row count < %d", bound),
		check:     func(n int) bool { return n < bound },
		reportGot: true,
	}
}

// RowCountLessOrEqual asserts len(rows) <= bound.
func RowCountLessOrEqual[T any](bound int) Expectation[T] {
	return rowCountExpectation[T]{
		label:     fmt.Sprintf("row count <= %d", bound),
		check:     func(n int) bool { return n <= bound },
		reportGot: true,
	}
}
