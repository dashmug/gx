package gx

import (
	"strings"
	"testing"
)

func TestRowReportsCrossFieldFailures(t *testing.T) {
	type ord struct{ A, B int }
	rows := []ord{{A: 2, B: 1}, {A: 1, B: 2}, {A: 3, B: 3}}
	rep := NewSuite[ord](
		Row("a>=b", func(o ord) bool { return o.A >= o.B }),
	).Validate(rows)

	res := rep.Results[0]
	if res.Name != "a>=b" || res.Column != "" {
		t.Fatalf("Name=%q Column=%q", res.Name, res.Column)
	}
	if res.FailedCount != 1 || len(res.FailedIndices) != 1 || res.FailedIndices[0] != 1 {
		t.Fatalf("FailedCount=%d FailedIndices=%v, want 1 and [1]", res.FailedCount, res.FailedIndices)
	}
}

func TestRowCountBetween(t *testing.T) {
	rep := NewSuite[int](RowCountBetween[int](1, 2)).Validate([]int{1, 2, 3})
	res := rep.Results[0]
	if res.Success {
		t.Fatal("3 rows is not in [1,2]")
	}
	if !strings.Contains(res.Name, "got 3") {
		t.Fatalf("Name=%q, want it to report the observed count", res.Name)
	}
}

func TestRowCountEqualPasses(t *testing.T) {
	rep := NewSuite[int](RowCountEqual[int](3)).Validate([]int{1, 2, 3})
	if !rep.OK() {
		t.Fatalf("RowCountEqual(3) over 3 rows should pass; report: %v", rep.Results[0])
	}
}

// Guard against nil FailedIndices on success
func TestRowVacuousPass(t *testing.T) {
	rep := NewSuite[int](Row("all pass", func(v int) bool { return true })).Validate([]int{1, 2, 3})
	if !rep.OK() {
		t.Fatal("expected ok")
	}
}

// Len guard before indexing FailedIndices in cross-field test
func TestRowCrossField(t *testing.T) {
	type pair struct{ A, B int }
	rows := []pair{{1, 2}, {3, 3}}
	rep := NewSuite[pair](
		Row("A!=B", func(p pair) bool { return p.A != p.B }),
	).Validate(rows)
	res := rep.Results[0]
	if res.FailedCount != 1 {
		t.Fatalf("FailedCount=%d want 1", res.FailedCount)
	}
	if len(res.FailedIndices) == 0 {
		t.Fatal("expected FailedIndices to be non-empty")
	}
	if res.FailedIndices[0] != 1 {
		t.Fatalf("FailedIndices[0]=%d want 1", res.FailedIndices[0])
	}
}

func TestRowCountOnEmptySlice(t *testing.T) {
	rep := NewSuite[int](RowCount[int]("no rows", func(n int) bool { return n == 0 })).Validate(nil)
	if !rep.OK() {
		t.Fatal("RowCount on empty slice should pass when predicate matches")
	}
}

func TestRowCountPerRowFieldsEmpty(t *testing.T) {
	rep := NewSuite[int](RowCount[int]("count", func(n int) bool { return n > 0 })).Validate([]int{1, 2})
	res := rep.Results[0]
	if res.FailedCount != 0 || len(res.FailedIndices) != 0 || len(res.SampleValues) != 0 {
		t.Fatal("RowCount result should have no per-row failure fields")
	}
}

func TestRowCountNameStability(t *testing.T) {
	exp := RowCount[int]("count check", func(n int) bool { return n > 0 })
	if exp.Name() != "count check" {
		t.Fatalf("Name()=%q want \"count check\"", exp.Name())
	}
}

func TestRowCountCustomPredicateFailure(t *testing.T) {
	rep := NewSuite[int](RowCount[int]("even row count", func(n int) bool { return n%2 == 0 })).Validate([]int{1, 2, 3})
	res := rep.Results[0]
	if res.Success {
		t.Fatal("expected custom RowCount predicate to fail")
	}
	if res.Name != "even row count" {
		t.Fatalf("Name = %q, want %q", res.Name, "even row count")
	}
	if res.Column != "" {
		t.Fatalf("Column = %q, want empty", res.Column)
	}
	if res.Total != 0 {
		t.Fatalf("Total = %d, want 0 for table-level RowCount", res.Total)
	}
	if res.FailedCount != 0 {
		t.Fatalf("FailedCount = %d, want 0 for table-level RowCount", res.FailedCount)
	}
	if len(res.FailedIndices) != 0 {
		t.Fatalf("FailedIndices = %v, want empty", res.FailedIndices)
	}
	if len(res.SampleValues) != 0 {
		t.Fatalf("SampleValues = %v, want empty", res.SampleValues)
	}
}

func TestRowCountCustomPredicatePassesOnNilRows(t *testing.T) {
	rep := NewSuite[int](RowCount[int]("empty ok", func(n int) bool { return n == 0 })).Validate(nil)
	if !rep.OK() {
		t.Fatalf("expected RowCount predicate over nil rows to pass; report: %v", rep.Results[0])
	}
}
