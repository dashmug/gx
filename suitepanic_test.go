package gx

import "testing"

type panicExpectation struct {
	name string
}

func (p panicExpectation) Name() string { return p.name }

func (p panicExpectation) Evaluate(rows []int, opts EvalOptions) Result {
	panic("boom")
}

type recordingExpectation struct {
	called bool
}

func (r *recordingExpectation) Name() string { return "later" }

func (r *recordingExpectation) Evaluate(rows []int, opts EvalOptions) Result {
	r.called = true
	return Result{Name: "later", Success: true, Total: len(rows)}
}

type panicNameExpectation struct{}

func (panicNameExpectation) Name() string { panic("name panic") }

func (panicNameExpectation) Evaluate(rows []int, opts EvalOptions) Result {
	panic("eval panic")
}

func TestValidateRecoversPanicsAndContinues(t *testing.T) {
	later := &recordingExpectation{}
	rep := NewSuite[int](panicExpectation{name: "panic"}, later).Validate([]int{1, 2})

	if !later.called {
		t.Fatal("later expectation should still run")
	}
	if len(rep.Results) != 2 {
		t.Fatalf("len(Results)=%d, want 2", len(rep.Results))
	}
	if rep.Results[0].Name != "panic" {
		t.Fatalf("first result Name=%q, want panic", rep.Results[0].Name)
	}
	if rep.Results[0].Success {
		t.Fatal("panicking expectation should fail")
	}
	if rep.Results[0].Err == nil {
		t.Fatal("panicking expectation should set Err")
	}
	if !rep.Results[1].Success || rep.Results[1].Name != "later" {
		t.Fatalf("later result=%+v, want successful later result", rep.Results[1])
	}
	if rep.Err() == nil {
		t.Fatal("Report.Err() should be non-nil when an expectation panics")
	}
}

func TestValidateRecoversWhenNamePanics(t *testing.T) {
	later := &recordingExpectation{}
	rep := NewSuite[int](panicNameExpectation{}, later).Validate([]int{1})

	if !later.called {
		t.Fatal("later expectation should still run")
	}
	if len(rep.Results) != 2 {
		t.Fatalf("len(Results)=%d, want 2", len(rep.Results))
	}
	if rep.Results[0].Name != unnamedExpectation {
		t.Fatalf("first result Name=%q, want %q", rep.Results[0].Name, unnamedExpectation)
	}
	if rep.Results[0].Success || rep.Results[0].Err == nil {
		t.Fatalf("panicking expectation should fail with Err set: %+v", rep.Results[0])
	}
	if !rep.Results[1].Success || rep.Results[1].Name != "later" {
		t.Fatalf("later result=%+v, want successful later result", rep.Results[1])
	}
	if rep.Err() == nil {
		t.Fatal("Report.Err() should be non-nil when an expectation panics")
	}
}

func TestValidateRecoversNilExpectation(t *testing.T) {
	later := &recordingExpectation{}
	rep := NewSuite[int](nil, later).Validate([]int{1})

	if !later.called {
		t.Fatal("later expectation should still run")
	}
	if len(rep.Results) != 2 {
		t.Fatalf("len(Results)=%d, want 2", len(rep.Results))
	}
	if rep.Results[0].Name != unnamedExpectation {
		t.Fatalf("nil expectation Name=%q, want %q", rep.Results[0].Name, unnamedExpectation)
	}
	if rep.Results[0].Success {
		t.Fatal("nil expectation should fail")
	}
	if rep.Results[0].Err == nil {
		t.Fatal("nil expectation should set Err")
	}
	if !rep.Results[1].Success {
		t.Fatalf("later result=%+v, want success", rep.Results[1])
	}
	if rep.Err() == nil {
		t.Fatal("Report.Err() should be non-nil when an expectation is nil")
	}
}
