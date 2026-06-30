package gx

// EvalColumn runs a typed predicate over every row, recording failures. It is
// the shared per-row loop for column-value checks: SampleValues and
// FailedIndices are capped when the corresponding EvalOptions limits are set;
// FailedCount stays complete.
func EvalColumn[T, V any](name, column string, rows []T,
	get func(T) V, pred func(V) bool, opts EvalOptions) Result {

	res := Result{Name: name, Column: column, Total: len(rows), Success: true}
	for i, row := range rows {
		v := get(row)
		if pred(v) {
			continue
		}
		res.Success = false
		res.FailedCount++
		res.FailedIndices = appendFailedIndex(res.FailedIndices, opts.FailedIndicesCap, i)
		if len(res.SampleValues) < opts.SampleCap {
			res.SampleValues = append(res.SampleValues, v)
		}
	}
	if res.Total > 0 {
		res.FailedPercent = float64(res.FailedCount) / float64(res.Total) * 100
	}
	return res
}

func appendFailedIndex(indices []int, cap int, i int) []int {
	if cap > 0 && len(indices) >= cap {
		return indices
	}
	return append(indices, i)
}

func evalColumn[T, V any](name, column string, rows []T,
	get func(T) V, pred func(V) bool, opts EvalOptions) Result {
	return EvalColumn(name, column, rows, get, pred, opts)
}

// colExpectation adapts a typed accessor + predicate to the Expectation interface.
type colExpectation[T, V any] struct {
	name, column string
	get          func(T) V
	pred         func(V) bool
}

func (e colExpectation[T, V]) Name() string { return e.name }

func (e colExpectation[T, V]) Evaluate(rows []T, opts EvalOptions) Result {
	return evalColumn(e.name, e.column, rows, e.get, e.pred, opts)
}

// newCol is the shared constructor used by every column builder.
func newCol[T, V any](name, column string, get func(T) V, pred func(V) bool) colExpectation[T, V] {
	return colExpectation[T, V]{name: name, column: column, get: get, pred: pred}
}

// inSet builds a set-membership expectation over any comparable value type.
func inSet[T any, V comparable](name, column string, get func(T) V, vals []V) Expectation[T] {
	set := make(map[V]struct{}, len(vals))
	for _, v := range vals {
		set[v] = struct{}{}
	}
	return newCol(name, column, get, func(v V) bool {
		_, ok := set[v]
		return ok
	})
}

// notInSet builds a set-exclusion expectation: value must NOT be in vals.
func notInSet[T any, V comparable](name, column string, get func(T) V, vals []V) Expectation[T] {
	set := make(map[V]struct{}, len(vals))
	for _, v := range vals {
		set[v] = struct{}{}
	}
	return newCol(name, column, get, func(v V) bool {
		_, ok := set[v]
		return !ok
	})
}
