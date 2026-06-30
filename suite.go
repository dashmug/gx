package gx

import "fmt"

// DefaultSampleCap is the maximum number of offending sample values a Result
// retains by default. Override per suite with WithSampleCap.
const DefaultSampleCap = 20

// DefaultFailedIndicesCap is the default maximum failing row indices retained
// per Result. Use WithFailedIndicesCap(0) for unlimited retention.
const DefaultFailedIndicesCap = 100

// unnamedExpectation is used when an expectation's Name cannot be obtained safely.
const unnamedExpectation = "<unnamed expectation>"

// EvalOptions is passed by the suite into each expectation's Evaluate. It is a
// struct so new knobs can be added without breaking the Expectation interface.
type EvalOptions struct {
	SampleCap        int
	FailedIndicesCap int // 0 means unlimited; default DefaultFailedIndicesCap
}

// Expectation is the unit of validation over rows of type T. Implement it for
// fully-custom checks; most users build expectations via the column helpers.
type Expectation[T any] interface {
	Name() string
	Evaluate(rows []T, opts EvalOptions) Result
}

// Suite is an ordered set of expectations over rows of type T.
type Suite[T any] struct {
	expectations     []Expectation[T]
	sampleCap        int
	failedIndicesCap int
}

// NewSuite builds a suite from the given expectations, evaluated in order.
func NewSuite[T any](exps ...Expectation[T]) *Suite[T] {
	return &Suite[T]{
		expectations:     exps,
		sampleCap:        DefaultSampleCap,
		failedIndicesCap: DefaultFailedIndicesCap,
	}
}

// WithSampleCap sets the maximum sample values retained per Result and returns
// the suite for chaining. Zero means no samples are collected; negative values
// are rejected when Validate runs.
func (s *Suite[T]) WithSampleCap(n int) *Suite[T] {
	s.sampleCap = n
	return s
}

// WithFailedIndicesCap sets the maximum failing row indices retained per Result
// and returns the suite for chaining. Zero means unlimited. FailedCount and
// FailedPercent remain complete when indices are capped.
func (s *Suite[T]) WithFailedIndicesCap(n int) *Suite[T] {
	s.failedIndicesCap = n
	return s
}

// Validate runs every expectation in declaration order (collect-all, never
// fail-fast) and returns the aggregated Report.
func (s *Suite[T]) Validate(rows []T) Report {
	if s.sampleCap < 0 {
		return Report{Results: []Result{{
			Name:    "<configuration error>",
			Success: false,
			Err:     fmt.Errorf("gx: sample cap must be non-negative"),
		}}}
	}
	if s.failedIndicesCap < 0 {
		return Report{Results: []Result{{
			Name:    "<configuration error>",
			Success: false,
			Err:     fmt.Errorf("gx: failed indices cap must be non-negative"),
		}}}
	}
	opts := EvalOptions{SampleCap: s.sampleCap, FailedIndicesCap: s.failedIndicesCap}
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
