package gx

import "testing"

type crow struct {
	Active bool
	Tags   []string
}

func TestComparableInAndNotZero(t *testing.T) {
	rows := []crow{{Active: true}, {Active: false}}
	rep := NewSuite[crow](
		Comparable("active", func(r crow) bool { return r.Active }).In(true),
		Comparable("active", func(r crow) bool { return r.Active }).NotZero(),
	).Validate(rows)

	if rep.Results[0].FailedCount != 1 { // false not in {true}
		t.Fatalf("In failed=%d, want 1", rep.Results[0].FailedCount)
	}
	if rep.Results[1].FailedCount != 1 { // false is the zero value
		t.Fatalf("NotZero failed=%d, want 1", rep.Results[1].FailedCount)
	}
}

func TestFieldSatisfy(t *testing.T) {
	rows := []crow{{Tags: []string{"a"}}, {Tags: nil}}
	rep := NewSuite[crow](
		Field("tags", func(r crow) []string { return r.Tags }).
			Satisfy("non-empty", func(v []string) bool { return len(v) > 0 }),
	).Validate(rows)

	res := rep.Results[0]
	if res.Name != "tags: non-empty" {
		t.Fatalf("Name=%q", res.Name)
	}
	if res.FailedCount != 1 || res.FailedIndices[0] != 1 {
		t.Fatalf("FailedCount=%d FailedIndices=%v, want 1 and [1]", res.FailedCount, res.FailedIndices)
	}
}
