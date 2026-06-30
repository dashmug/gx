package gx

import (
	"fmt"
	"testing"
)

// stubExp is a test-only Expectation that records the options it received.
type stubExp struct {
	name          string
	res           Result
	gotCap        int
	gotIndicesCap int
}

func (s *stubExp) Name() string { return s.name }
func (s *stubExp) Evaluate(rows []int, opts EvalOptions) Result {
	s.gotCap = opts.SampleCap
	s.gotIndicesCap = opts.FailedIndicesCap
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

func TestDefaultFailedIndicesCap(t *testing.T) {
	a := &stubExp{name: "a", res: Result{Success: true}}
	NewSuite[int](a).Validate(nil)
	if a.gotIndicesCap != DefaultFailedIndicesCap {
		t.Fatalf("expectation saw FailedIndicesCap=%d, want DefaultFailedIndicesCap=%d", a.gotIndicesCap, DefaultFailedIndicesCap)
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

func TestValidateRejectsNegativeSampleCap(t *testing.T) {
	a := &stubExp{name: "a", res: Result{Name: "a", Success: true}}
	rep := NewSuite[int](a).WithSampleCap(-1).Validate([]int{1})

	if rep.OK() {
		t.Fatal("negative sample cap should fail validation")
	}
	if len(rep.Results) != 1 || rep.Results[0].Err == nil {
		t.Fatalf("want configuration error result, got %#v", rep.Results)
	}
	if a.gotCap != 0 {
		t.Fatal("expectations should not run when sample cap is invalid")
	}
}

func TestWithSampleCapZeroCollectsNoSamples(t *testing.T) {
	rows := []orow{{Age: 1}, {Age: 99}}
	rep := NewSuite[orow](
		Ordered("age", func(r orow) int { return r.Age }).Between(0, 50),
	).WithSampleCap(0).Validate(rows)

	res := rep.Results[0]
	if res.Success {
		t.Fatal("want failure")
	}
	if len(res.SampleValues) != 0 {
		t.Fatalf("SampleValues = %#v, want empty when cap is 0", res.SampleValues)
	}
	if res.FailedCount != 1 {
		t.Fatalf("FailedCount = %d, want 1", res.FailedCount)
	}
}

func TestWithFailedIndicesCapThreaded(t *testing.T) {
	a := &stubExp{name: "a", res: Result{Success: true}}
	NewSuite[int](a).WithFailedIndicesCap(50).Validate(nil)
	if a.gotIndicesCap != 50 {
		t.Fatalf("expectation saw FailedIndicesCap=%d, want 50", a.gotIndicesCap)
	}
}

func TestValidateRejectsNegativeFailedIndicesCap(t *testing.T) {
	a := &stubExp{name: "a", res: Result{Name: "a", Success: true}}
	rep := NewSuite[int](a).WithFailedIndicesCap(-1).Validate([]int{1})

	if rep.OK() {
		t.Fatal("negative failed indices cap should fail validation")
	}
	if len(rep.Results) != 1 || rep.Results[0].Err == nil {
		t.Fatalf("want configuration error result, got %#v", rep.Results)
	}
	if a.gotIndicesCap != 0 {
		t.Fatal("expectations should not run when failed indices cap is invalid")
	}
}

func TestFailedIndicesCapLimitsReturnedIndices(t *testing.T) {
	rows := []orow{{Age: 1}, {Age: 200}, {Age: 300}}
	rep := NewSuite[orow](
		Ordered("age", func(r orow) int { return r.Age }).Between(0, 120),
	).WithFailedIndicesCap(1).Validate(rows)

	res := rep.Results[0]
	if res.FailedCount != 2 {
		t.Fatalf("FailedCount = %d, want 2", res.FailedCount)
	}
	if len(res.FailedIndices) != 1 {
		t.Fatalf("FailedIndices len = %d, want 1", len(res.FailedIndices))
	}
}

func TestWithFailedIndicesCapZeroUnlimited(t *testing.T) {
	rows := make([]orow, 105)
	for i := range rows {
		rows[i] = orow{Age: 200}
	}
	rep := NewSuite[orow](
		Ordered("age", func(r orow) int { return r.Age }).Between(0, 120),
	).WithFailedIndicesCap(0).Validate(rows)

	res := rep.Results[0]
	if res.FailedCount != 105 {
		t.Fatalf("FailedCount = %d, want 105", res.FailedCount)
	}
	if len(res.FailedIndices) != 105 {
		t.Fatalf("FailedIndices len = %d, want 105 (unlimited)", len(res.FailedIndices))
	}
}

func TestDefaultFailedIndicesCapBoundsLargeFailureSets(t *testing.T) {
	rows := make([]orow, 105)
	for i := range rows {
		rows[i] = orow{Age: 200}
	}
	rep := NewSuite[orow](
		Ordered("age", func(r orow) int { return r.Age }).Between(0, 120),
	).Validate(rows)

	res := rep.Results[0]
	if res.FailedCount != 105 {
		t.Fatalf("FailedCount = %d, want 105", res.FailedCount)
	}
	if len(res.FailedIndices) != DefaultFailedIndicesCap {
		t.Fatalf("FailedIndices len = %d, want default cap %d", len(res.FailedIndices), DefaultFailedIndicesCap)
	}
}
