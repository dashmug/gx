package gx

import (
	"fmt"
	"math"
	"sort"
)

// Number is the constraint for numeric aggregate column accessors.
type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64
}

// NumericStats holds aggregate statistics over a numeric column.
type NumericStats struct {
	Count            int
	Sum              float64
	Average          float64
	Median           float64
	StdDevPopulation float64
	StdDevSample     float64
}

// NumericColumn is a typed accessor over a numeric field. Build it with Numeric.
type NumericColumn[T any, V Number] struct {
	name string
	get  func(T) V
}

// Numeric creates a column over a numeric field. name is a report label and may
// differ from the struct field, so computed values are first-class.
func Numeric[T any, V Number](name string, get func(T) V) NumericColumn[T, V] {
	return NumericColumn[T, V]{name: name, get: get}
}

// AverageBetween asserts lo <= average <= hi (inclusive) over the column.
func (c NumericColumn[T, V]) AverageBetween(lo, hi float64) Expectation[T] {
	return numericBetweenExpectation[T, V]{
		column: c,
		label:  fmt.Sprintf("%s average in [%g,%g]", c.name, lo, hi),
		lo:     lo,
		hi:     hi,
		observe: func(s NumericStats) float64 {
			return s.Average
		},
	}
}

// MedianBetween asserts lo <= median <= hi (inclusive) over the column.
func (c NumericColumn[T, V]) MedianBetween(lo, hi float64) Expectation[T] {
	return numericBetweenExpectation[T, V]{
		column: c,
		label:  fmt.Sprintf("%s median in [%g,%g]", c.name, lo, hi),
		lo:     lo,
		hi:     hi,
		observe: func(s NumericStats) float64 {
			return s.Median
		},
	}
}

// StdDevBetween asserts lo <= population standard deviation <= hi (inclusive).
func (c NumericColumn[T, V]) StdDevBetween(lo, hi float64) Expectation[T] {
	return numericBetweenExpectation[T, V]{
		column: c,
		label:  fmt.Sprintf("%s standard deviation in [%g,%g]", c.name, lo, hi),
		lo:     lo,
		hi:     hi,
		observe: func(s NumericStats) float64 {
			return s.StdDevPopulation
		},
	}
}

// SatisfyAggregate asserts a custom predicate over aggregate statistics; check
// names the rule in reports.
func (c NumericColumn[T, V]) SatisfyAggregate(check string, pred func(NumericStats) bool) Expectation[T] {
	return numericSatisfyAggregateExpectation[T, V]{
		column: c,
		check:  check,
		pred:   pred,
	}
}

type numericBetweenExpectation[T any, V Number] struct {
	column  NumericColumn[T, V]
	label   string
	lo, hi  float64
	observe func(NumericStats) float64
}

func (e numericBetweenExpectation[T, V]) Name() string { return e.label }

func (e numericBetweenExpectation[T, V]) Evaluate(rows []T, _ EvalOptions) Result {
	res := tableLevelNumericResult(e.column.name, e.label, true)

	if len(rows) == 0 {
		return res
	}

	values, nonFinite := extractNumericValues(rows, e.column.get)
	if nonFinite {
		res.Name = e.label + ": got non-finite value"
		res.Success = false
		return res
	}

	stats := numericStatsFromValues(values)
	observed := e.observe(stats)
	res.Name = fmt.Sprintf("%s: got %g", e.label, observed)
	res.Success = observed >= e.lo && observed <= e.hi
	return res
}

type numericSatisfyAggregateExpectation[T any, V Number] struct {
	column NumericColumn[T, V]
	check  string
	pred   func(NumericStats) bool
}

func (e numericSatisfyAggregateExpectation[T, V]) Name() string {
	return fmt.Sprintf("%s: %s", e.column.name, e.check)
}

func (e numericSatisfyAggregateExpectation[T, V]) Evaluate(rows []T, _ EvalOptions) Result {
	label := e.Name()
	res := tableLevelNumericResult(e.column.name, label, true)

	if len(rows) == 0 {
		return res
	}

	values, nonFinite := extractNumericValues(rows, e.column.get)
	if nonFinite {
		res.Name = label + ": got non-finite value"
		res.Success = false
		return res
	}

	stats := numericStatsFromValues(values)
	res.Success = e.pred(stats)
	return res
}

func tableLevelNumericResult(column, name string, success bool) Result {
	return Result{
		Name:    name,
		Column:  column,
		Success: success,
	}
}

func extractNumericValues[T any, V Number](rows []T, get func(T) V) ([]float64, bool) {
	values := make([]float64, 0, len(rows))
	for _, row := range rows {
		f := float64(get(row))
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return nil, true
		}
		values = append(values, f)
	}
	return values, false
}

func numericStatsFromValues(values []float64) NumericStats {
	n := len(values)
	if n == 0 {
		return NumericStats{}
	}

	var sum float64
	for _, v := range values {
		sum += v
	}
	avg := sum / float64(n)

	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)

	var median float64
	mid := n / 2
	if n%2 == 1 {
		median = sorted[mid]
	} else {
		median = (sorted[mid-1] + sorted[mid]) / 2
	}

	var sqDiffSum float64
	for _, v := range values {
		d := v - avg
		sqDiffSum += d * d
	}

	popStd := math.Sqrt(sqDiffSum / float64(n))
	var sampleStd float64
	if n > 1 {
		sampleStd = math.Sqrt(sqDiffSum / float64(n-1))
	}

	return NumericStats{
		Count:            n,
		Sum:              sum,
		Average:          avg,
		Median:           median,
		StdDevPopulation: popStd,
		StdDevSample:     sampleStd,
	}
}
