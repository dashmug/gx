package gx

import (
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
