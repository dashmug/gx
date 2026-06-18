package gx

import (
	"reflect"
	"regexp"
	"testing"
)

type srow struct {
	Name string
}

func TestMatchRegex(t *testing.T) {
	re := regexp.MustCompile(`^\d+$`)
	rows := []srow{{Name: "123"}, {Name: "abc"}, {Name: "45"}}
	rep := NewSuite[srow](
		Str("name", func(r srow) string { return r.Name }).MatchRegex(re),
	).Validate(rows)

	res := rep.Results[0]
	if res.FailedCount != 1 || res.FailedIndices[0] != 1 {
		t.Fatalf("FailedCount=%d FailedIndices=%v, want 1 and [1]", res.FailedCount, res.FailedIndices)
	}
	if res.Name != `name matches /^\d+$/` {
		t.Fatalf("Name=%q", res.Name)
	}
}

func TestNotEmpty(t *testing.T) {
	rows := []srow{{Name: "x"}, {Name: ""}, {Name: "y"}}
	rep := NewSuite[srow](
		Str("name", func(r srow) string { return r.Name }).NotEmpty(),
	).Validate(rows)
	if rep.Results[0].FailedCount != 1 || rep.Results[0].FailedIndices[0] != 1 {
		t.Fatalf("got FailedCount=%d FailedIndices=%v", rep.Results[0].FailedCount, rep.Results[0].FailedIndices)
	}
}

func TestStrInheritsOrderedChecks(t *testing.T) {
	rows := []srow{{Name: "a"}, {Name: "z"}}
	rep := NewSuite[srow](
		Str("name", func(r srow) string { return r.Name }).In("a", "b"),
	).Validate(rows)
	if rep.Results[0].FailedCount != 1 { // "z" not in {"a","b"}
		t.Fatalf("inherited In failed=%d, want 1", rep.Results[0].FailedCount)
	}
}

func TestNotMatchRegex(t *testing.T) {
	re := regexp.MustCompile(`^\d+$`)
	rows := []srow{{Name: "123"}, {Name: "abc"}, {Name: "45"}}
	rep := NewSuite[srow](
		Str("name", func(r srow) string { return r.Name }).NotMatchRegex(re),
	).Validate(rows)

	res := rep.Results[0]
	if res.FailedCount != 2 || !reflect.DeepEqual(res.FailedIndices, []int{0, 2}) {
		t.Fatalf("FailedCount=%d FailedIndices=%v, want 2 and [0 2]", res.FailedCount, res.FailedIndices)
	}
	if res.Name != `name does not match /^\d+$/` {
		t.Fatalf("Name=%q", res.Name)
	}
}

func TestEmptyOnlyNonEmptyRowsFail(t *testing.T) {
	// names ["", "x", ""]: Empty() passes "", fails "x" → FailedIndices=[1]
	rows := []srow{{Name: ""}, {Name: "x"}, {Name: ""}}
	rep := NewSuite[srow](
		Str("name", func(r srow) string { return r.Name }).Empty(),
	).Validate(rows)

	res := rep.Results[0]
	if res.Success {
		t.Fatal("want failure: non-empty name should fail Empty()")
	}
	if res.FailedCount != 1 || res.FailedIndices[0] != 1 {
		t.Fatalf("FailedCount=%d FailedIndices=%v, want 1 and [1]", res.FailedCount, res.FailedIndices)
	}
}

func TestLenBetweenRuneCount(t *testing.T) {
	rows := []srow{{Name: "ab"}, {Name: "abc"}, {Name: "café"}, {Name: "hi👋"}}
	rep := NewSuite[srow](
		Str("name", func(r srow) string { return r.Name }).LenBetween(3, 4),
	).Validate(rows)

	res := rep.Results[0]
	if res.FailedCount != 1 || res.FailedIndices[0] != 0 {
		t.Fatalf("LenBetween(3,4) FailedCount=%d FailedIndices=%v, want 1 and [0] (ab has 2 runes)", res.FailedCount, res.FailedIndices)
	}
	if res.Name != "name length in [3,4]" {
		t.Fatalf("Name=%q", res.Name)
	}
}

func TestLenEqualRuneCount(t *testing.T) {
	rows := []srow{{Name: "café"}, {Name: "abc"}, {Name: "hi👋"}}
	rep := NewSuite[srow](
		Str("name", func(r srow) string { return r.Name }).LenEqual(4),
	).Validate(rows)

	res := rep.Results[0]
	if res.FailedCount != 2 || !reflect.DeepEqual(res.FailedIndices, []int{1, 2}) {
		t.Fatalf("LenEqual(4) FailedCount=%d FailedIndices=%v, want 2 and [1 2]", res.FailedCount, res.FailedIndices)
	}
}

func TestStrInheritsOrdering(t *testing.T) {
	rows := []srow{{Name: "b"}, {Name: "a"}}
	rep := NewSuite[srow](
		Str("name", func(r srow) string { return r.Name }).GreaterThan("a"),
	).Validate(rows)

	res := rep.Results[0]
	if res.FailedCount != 1 || res.FailedIndices[0] != 1 {
		t.Fatalf("inherited GreaterThan FailedCount=%d FailedIndices=%v, want 1 and [1]", res.FailedCount, res.FailedIndices)
	}
}
