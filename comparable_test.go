package gx

import (
	"reflect"
	"testing"
)

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

func TestComparableNotInAndZero(t *testing.T) {
	// active [true, false, true]:
	//   NotIn(false) fails index 1 (false is forbidden)
	//   Zero() passes false (zero value for bool), fails true → FailedIndices=[0, 2]
	rows := []crow{{Active: true}, {Active: false}, {Active: true}}
	rep := NewSuite[crow](
		Comparable("active", func(r crow) bool { return r.Active }).NotIn(false),
		Comparable("active", func(r crow) bool { return r.Active }).Zero(),
	).Validate(rows)

	if rep.Results[0].FailedCount != 1 || rep.Results[0].FailedIndices[0] != 1 {
		t.Fatalf("NotIn(false) FailedCount=%d FailedIndices=%v, want 1 and [1]",
			rep.Results[0].FailedCount, rep.Results[0].FailedIndices)
	}
	if rep.Results[1].FailedCount != 2 || !reflect.DeepEqual(rep.Results[1].FailedIndices, []int{0, 2}) {
		t.Fatalf("Zero() FailedCount=%d FailedIndices=%v, want 2 and [0 2] (true is non-zero for bool)",
			rep.Results[1].FailedCount, rep.Results[1].FailedIndices)
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
