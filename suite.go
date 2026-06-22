package gx

import "fmt"

// DefaultSampleCap is the maximum number of offending sample values a Result
// retains by default. Override per suite with WithSampleCap.
const DefaultSampleCap = 20

// unnamedExpectation is used when an expectation's Name cannot be obtained safely.
const unnamedExpectation = "<unnamed expectation>"

// EvalOptions is passed by the suite into each expectation's Evaluate. It is a
// struct so new knobs can be added without breaking the Expectation interface.
type EvalOptions struct {
	SampleCap int
}

// Expectation is the unit of validation over rows of type T. Implement it for
// fully-custom checks; most users build expectations via the column helpers.
type Expectation[T any] interface {
	Name() string
	Evaluate(rows []T, opts EvalOptions) Result
}

// Suite is an ordered set of expectations over rows of type T.
type Suite[T any] struct {
	expectations []Expectation[T]
	sampleCap    int
}

// NewSuite builds a suite from the given expectations, evaluated in order.
func NewSuite[T any](exps ...Expectation[T]) *Suite[T] {
	return &Suite[T]{expectations: exps, sampleCap: DefaultSampleCap}
}

// WithSampleCap sets the maximum sample values retained per Result and returns
// the suite for chaining.
func (s *Suite[T]) WithSampleCap(n int) *Suite[T] {
	s.sampleCap = n
	return s
}

// Validate runs every expectation in declaration order (collect-all, never
// fail-fast) and returns the aggregated Report.
func (s *Suite[T]) Validate(rows []T) Report {
	opts := EvalOptions{SampleCap: s.sampleCap}
	results := make([]Result, len(s.expectations))
	for i, e := range s.expectations {
		res := evaluateExpectation(e, rows, opts)
		if res.Err != nil {
			res.Success = false
		}
		results[i] = res
	}
	return Report{Results: results}
}

func safeExpectationName[T any](e Expectation[T]) (name string, ok bool) {
	defer func() {
		if recover() != nil {
			name = ""
			ok = false
		}
	}()
	if e == nil {
		return "", false
	}
	return e.Name(), true
}

func evaluateExpectation[T any](e Expectation[T], rows []T, opts EvalOptions) (res Result) {
	name, hasName := safeExpectationName(e)
	defer func() {
		if r := recover(); r != nil {
			failName := unnamedExpectation
			if hasName {
				failName = name
			} else if n, ok := safeExpectationName(e); ok {
				failName = n
			}
			res = Result{
				Name:    failName,
				Success: false,
				Err:     fmt.Errorf("panic during validation: %v", r),
			}
		}
	}()
	return e.Evaluate(rows, opts)
}
