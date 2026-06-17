package gx

import (
	"strings"
	"testing"
)

func TestResultStringPassAndFail(t *testing.T) {
	pass := Result{Name: "id unique", Success: true, Total: 100}
	if got := pass.String(); !strings.HasPrefix(got, "✓ id unique") {
		t.Fatalf("pass render = %q", got)
	}

	fail := Result{
		Name: "email matches", Success: false, Total: 100,
		FailedCount: 12, FailedPercent: 12.0,
		SampleValues:  []any{"bad", "x@"},
		FailedIndices: []int{3, 9, 41},
	}
	got := fail.String()
	if !strings.HasPrefix(got, "✗ email matches") {
		t.Fatalf("fail render missing mark/name: %q", got)
	}
	if !strings.Contains(got, "12/100 failed (12.0%)") {
		t.Fatalf("fail render missing counts: %q", got)
	}
}

func TestResultStringTruncates(t *testing.T) {
	idx := make([]int, 15)
	for i := range idx {
		idx[i] = i
	}
	r := Result{Name: "x", Success: false, Total: 15, FailedCount: 15, FailedIndices: idx}
	if !strings.Contains(r.String(), "…") {
		t.Fatalf("long index list should be truncated with …: %q", r.String())
	}
}

func TestReportStringHeader(t *testing.T) {
	rep := Report{Results: []Result{
		{Name: "a", Success: true, Total: 1},
		{Name: "b", Success: false, Total: 1, FailedCount: 1, FailedPercent: 100},
	}}
	got := rep.String()
	if !strings.HasPrefix(got, "gx report: 1/2 expectations passed") {
		t.Fatalf("report header = %q", got)
	}
	if strings.Count(got, "\n") != 2 {
		t.Fatalf("want one line per result, got %q", got)
	}
}

func TestTruncListExact(t *testing.T) {
	got := truncList([]int{1, 2, 3, 4}, 3)
	want := "[1 2 3 \u2026]" // U+2026 HORIZONTAL ELLIPSIS, space before it
	if got != want {
		t.Fatalf("truncList: got %q want %q", got, want)
	}
}
