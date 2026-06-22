package gx

// uniqueExpectation flags duplicate values in a column: the first occurrence of
// a value passes; every later occurrence fails.
type uniqueExpectation[T any, V comparable] struct {
	name, column string
	get          func(T) V
}

func (e uniqueExpectation[T, V]) Name() string { return e.name }

func (e uniqueExpectation[T, V]) Evaluate(rows []T, opts EvalOptions) Result {
	res := Result{Name: e.name, Column: e.column, Total: len(rows), Success: true}
	seen := make(map[V]struct{}, len(rows))
	for i, row := range rows {
		v := e.get(row)
		if _, dup := seen[v]; dup {
			res.Success = false
			res.FailedCount++
			res.FailedIndices = append(res.FailedIndices, i)
			if len(res.SampleValues) < opts.SampleCap {
				res.SampleValues = append(res.SampleValues, v)
			}
			continue
		}
		seen[v] = struct{}{}
	}
	if res.Total > 0 {
		res.FailedPercent = float64(res.FailedCount) / float64(res.Total) * 100
	}
	return res
}

// unique builds a uniqueness expectation for a comparable column.
func unique[T any, V comparable](label string, get func(T) V) Expectation[T] {
	return uniqueExpectation[T, V]{name: label + " unique", column: label, get: get}
}
