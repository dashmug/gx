package gx

import (
	"reflect"
	"testing"
)

func TestUniqueFlagsLaterDuplicates(t *testing.T) {
	rows := []srow{{Name: "a"}, {Name: "b"}, {Name: "a"}, {Name: "a"}}
	rep := NewSuite[srow](
		Str("name", func(r srow) string { return r.Name }).Unique(),
	).Validate(rows)

	res := rep.Results[0]
	if res.Success {
		t.Fatal("want failure: 'a' is duplicated")
	}
	if res.Name != "name unique" {
		t.Fatalf("Name=%q", res.Name)
	}
	if !reflect.DeepEqual(res.FailedIndices, []int{2, 3}) { // first 'a' at 0 passes
		t.Fatalf("FailedIndices=%v, want [2 3]", res.FailedIndices)
	}
}

func TestUniqueAllDistinctPasses(t *testing.T) {
	rows := []orow{{Age: 1}, {Age: 2}, {Age: 3}}
	rep := NewSuite[orow](
		Ordered("age", func(r orow) int { return r.Age }).Unique(),
	).Validate(rows)
	if !rep.OK() {
		t.Fatal("all-distinct column should pass Unique")
	}
}

func TestUniqueComparableColumn(t *testing.T) {
	rows := []crow{{Active: true}, {Active: false}, {Active: true}}
	rep := NewSuite[crow](
		Comparable("active", func(r crow) bool { return r.Active }).Unique(),
	).Validate(rows)

	res := rep.Results[0]
	if res.Success {
		t.Fatal("want failure: true is duplicated")
	}
	if !reflect.DeepEqual(res.FailedIndices, []int{2}) {
		t.Fatalf("FailedIndices=%v, want [2]", res.FailedIndices)
	}
}

func TestUniqueEmptyVacuous(t *testing.T) {
	rep := NewSuite[orow](
		Ordered("age", func(r orow) int { return r.Age }).Unique(),
	).Validate(nil)
	if !rep.OK() {
		t.Fatal("empty input should pass Unique vacuously")
	}
}

func TestUniqueSampleValuesAndPercent(t *testing.T) {
	// Create rows with enough duplicates to exceed SampleCap
	rows := make([]struct{ V int }, 25)
	for i := range rows {
		rows[i].V = i % 3 // values 0,1,2 repeat — many duplicates
	}
	col := Ordered("v", func(r struct{ V int }) int { return r.V })
	rep := NewSuite(col.Unique()).WithSampleCap(5).Validate(rows)
	res := rep.Results[0]
	if res.Success {
		t.Fatal("expected failure")
	}
	if len(res.SampleValues) > 5 {
		t.Fatalf("SampleValues len=%d, want <=5", len(res.SampleValues))
	}
	if res.FailedPercent <= 0 || res.FailedPercent > 100 {
		t.Fatalf("FailedPercent=%f, want (0,100]", res.FailedPercent)
	}
}
