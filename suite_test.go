package gx

import (
	"fmt"
	"testing"
)

// stubExp is a test-only Expectation that records the options it received.
type stubExp struct {
	name   string
	res    Result
	gotCap int
}

func (s *stubExp) Name() string { return s.name }
func (s *stubExp) Evaluate(rows []int, opts EvalOptions) Result {
	s.gotCap = opts.SampleCap
	return s.res
}

func TestValidateCollectAllInOrder(t *testing.T) {
	a := &stubExp{name: "a", res: Result{Name: "a", Success: true}}
	b := &stubExp{name: "b", res: Result{Name: "b", Success: false}}
	rep := NewSuite[int](a, b).Validate([]int{1, 2, 3})

	if len(rep.Results) != 2 {
		t.Fatalf("got %d results, want 2", len(rep.Results))
	}
	if rep.Results[0].Name != "a" || rep.Results[1].Name != "b" {
		t.Fatalf("results out of declaration order: %v", rep.Results)
	}
	if rep.OK() {
		t.Fatal("report with a failing expectation should not be OK")
	}
}

func TestWithSampleCapThreaded(t *testing.T) {
	a := &stubExp{name: "a", res: Result{Success: true}}
	NewSuite[int](a).WithSampleCap(7).Validate(nil)
	if a.gotCap != 7 {
		t.Fatalf("expectation saw SampleCap=%d, want 7", a.gotCap)
	}
}

func TestDefaultSampleCap(t *testing.T) {
	a := &stubExp{name: "a", res: Result{Success: true}}
	NewSuite[int](a).Validate(nil)
	if a.gotCap != DefaultSampleCap {
		t.Fatalf("expectation saw SampleCap=%d, want DefaultSampleCap=%d", a.gotCap, DefaultSampleCap)
	}
}

type errExp struct{}

func (errExp) Name() string { return "always-err" }
func (errExp) Evaluate(rows []int, opts EvalOptions) Result {
	return Result{Success: true, Err: fmt.Errorf("eval error"), Name: "always-err"}
}

func TestValidateNormalizesResultErr(t *testing.T) {
	rep := NewSuite[int](errExp{}).Validate(nil)
	if rep.OK() {
		t.Fatal("Result.Err should normalize Success to false")
	}
	if rep.Err() == nil {
		t.Fatal("Report.Err() should be non-nil when a result has Err set")
	}
}
