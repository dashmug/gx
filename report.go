package gx

import (
	"fmt"
	"strings"
)

// Result is the outcome of one expectation over a single Validate run.
type Result struct {
	Name          string // e.g. "age between [0,120]"
	Column        string // column label; "" for row-level and table-level checks
	Success       bool
	Total         int // rows evaluated
	FailedCount   int
	FailedPercent float64 // FailedCount/Total*100; 0 when Total==0
	SampleValues  []any   // capped sample of offending values
	FailedIndices []int   // indices into the validated slice; complete (uncapped)
	Err           error   // custom-expectation evaluation error; nil for built-ins
}

// Report aggregates the results of every expectation in a suite.
type Report struct {
	Results []Result
}

// OK reports whether every expectation passed.
func (r Report) OK() bool {
	for _, res := range r.Results {
		if !res.Success {
			return false
		}
	}
	return true
}

// Failures returns only the results that did not pass.
func (r Report) Failures() []Result {
	var out []Result
	for _, res := range r.Results {
		if !res.Success {
			out = append(out, res)
		}
	}
	return out
}

// Err returns nil when the report is OK, otherwise a *ValidationError carrying
// the full report for gating and inspection.
func (r Report) Err() error {
	if r.OK() {
		return nil
	}
	return &ValidationError{Report: r}
}

// ValidationError wraps a failed Report as an error for runtime gating.
// Recover the full report via errors.As and the Report field.
type ValidationError struct {
	Report Report
}

func (e *ValidationError) Error() string {
	failures := e.Report.Failures()
	names := make([]string, len(failures))
	for i, res := range failures {
		names[i] = res.Name
	}
	return fmt.Sprintf("gx: %d expectation(s) failed: %s",
		len(failures), strings.Join(names, "; "))
}
