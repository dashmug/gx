package gx

import (
	"reflect"
	"testing"
)

type orow struct {
	Age  int
	Name string
}

func TestBetweenCountsIndicesPercent(t *testing.T) {
	rows := []orow{{Age: 10}, {Age: 200}, {Age: -5}, {Age: 30}}
	rep := NewSuite[orow](
		Ordered("age", func(r orow) int { return r.Age }).Between(0, 120),
	).Validate(rows)

	res := rep.Results[0]
	if res.Success {
		t.Fatal("want failure: 200 and -5 are out of [0,120]")
	}
	if res.Name != "age between [0,120]" || res.Column != "age" {
		t.Fatalf("Name=%q Column=%q", res.Name, res.Column)
	}
	if res.Total != 4 || res.FailedCount != 2 {
		t.Fatalf("Total=%d FailedCount=%d, want 4 and 2", res.Total, res.FailedCount)
	}
	if !reflect.DeepEqual(res.FailedIndices, []int{1, 2}) {
		t.Fatalf("FailedIndices=%v, want [1 2]", res.FailedIndices)
	}
	if res.FailedPercent != 50 {
		t.Fatalf("FailedPercent=%v, want 50", res.FailedPercent)
	}
}

func TestInAndNotZero(t *testing.T) {
	rows := []orow{{Age: 1}, {Age: 0}, {Age: 9}}
	rep := NewSuite[orow](
		Ordered("age", func(r orow) int { return r.Age }).In(1, 2, 3),
		Ordered("age", func(r orow) int { return r.Age }).NotZero(),
	).Validate(rows)

	if rep.Results[0].FailedCount != 2 { // 0 and 9 not in {1,2,3}
		t.Fatalf("In failed=%d, want 2", rep.Results[0].FailedCount)
	}
	if rep.Results[1].FailedCount != 1 { // only Age==0
		t.Fatalf("NotZero failed=%d, want 1", rep.Results[1].FailedCount)
	}
}

func TestSampleCapBoundsSamplesNotIndices(t *testing.T) {
	rows := make([]orow, 5) // all Age==0, all fail Between(1,10)
	rep := NewSuite[orow](
		Ordered("age", func(r orow) int { return r.Age }).Between(1, 10),
	).WithSampleCap(2).Validate(rows)

	res := rep.Results[0]
	if len(res.SampleValues) != 2 {
		t.Fatalf("SampleValues len=%d, want capped at 2", len(res.SampleValues))
	}
	if len(res.FailedIndices) != 5 {
		t.Fatalf("FailedIndices len=%d, want complete 5", len(res.FailedIndices))
	}
}

func TestEmptyInputVacuousPass(t *testing.T) {
	rep := NewSuite[orow](
		Ordered("age", func(r orow) int { return r.Age }).Between(0, 120),
	).Validate(nil)

	if !rep.OK() {
		t.Fatal("empty input should pass per-row checks vacuously")
	}
	if rep.Results[0].Total != 0 {
		t.Fatalf("Total=%d, want 0", rep.Results[0].Total)
	}
}

func TestSatisfy(t *testing.T) {
	rows := []orow{{Age: 2}, {Age: 3}, {Age: 4}}
	rep := NewSuite[orow](
		Ordered("age", func(r orow) int { return r.Age }).Satisfy("even", func(v int) bool { return v%2 == 0 }),
	).Validate(rows)

	res := rep.Results[0]
	if res.Name != "age: even" {
		t.Fatalf("Name=%q, want \"age: even\"", res.Name)
	}
	if res.FailedCount != 1 || res.FailedIndices[0] != 1 {
		t.Fatalf("FailedCount=%d FailedIndices=%v, want 1 and [1]", res.FailedCount, res.FailedIndices)
	}
}

func TestNotInForbiddenValues(t *testing.T) {
	rows := []orow{{Age: 1}, {Age: 2}, {Age: 9}}
	rep := NewSuite[orow](
		Ordered("age", func(r orow) int { return r.Age }).NotIn(2, 3),
	).Validate(rows)

	res := rep.Results[0]
	if res.Success {
		t.Fatal("want failure: 2 is forbidden")
	}
	if res.Name != "age not in [2 3]" || res.Column != "age" {
		t.Fatalf("Name=%q Column=%q", res.Name, res.Column)
	}
	if res.FailedCount != 1 || res.FailedIndices[0] != 1 {
		t.Fatalf("FailedCount=%d FailedIndices=%v, want 1 and [1]", res.FailedCount, res.FailedIndices)
	}
}

func TestNotInEmptySetVacuousPass(t *testing.T) {
	rows := []orow{{Age: 0}, {Age: 99}}
	rep := NewSuite[orow](
		Ordered("age", func(r orow) int { return r.Age }).NotIn(),
	).Validate(rows)

	if !rep.OK() {
		t.Fatal("NotIn() with no forbidden values should pass all rows")
	}
}

func TestZeroOnlyNonZeroRowsFail(t *testing.T) {
	// ages [0, 5, 0]: Zero() passes 0, fails 5 → FailedIndices=[1]
	rows := []orow{{Age: 0}, {Age: 5}, {Age: 0}}
	rep := NewSuite[orow](
		Ordered("age", func(r orow) int { return r.Age }).Zero(),
	).Validate(rows)

	res := rep.Results[0]
	if res.Success {
		t.Fatal("want failure: non-zero age should fail Zero()")
	}
	if res.Name != "age zero" || res.Column != "age" {
		t.Fatalf("Name=%q Column=%q", res.Name, res.Column)
	}
	if res.FailedCount != 1 {
		t.Fatalf("FailedCount=%d, want 1", res.FailedCount)
	}
	if !reflect.DeepEqual(res.FailedIndices, []int{1}) {
		t.Fatalf("FailedIndices=%v, want [1]", res.FailedIndices)
	}
}

func TestOrderingMethodsBoundaries(t *testing.T) {
	rows := []orow{{Age: 10}, {Age: 20}, {Age: 30}}

	cases := []struct {
		name      string
		exp       Expectation[orow]
		wantFail  []int
		wantLabel string
	}{
		{
			name:      "gt",
			exp:       Ordered("age", func(r orow) int { return r.Age }).GreaterThan(20),
			wantFail:  []int{0, 1},
			wantLabel: "age > 20",
		},
		{
			name:      "gte",
			exp:       Ordered("age", func(r orow) int { return r.Age }).GreaterOrEqual(20),
			wantFail:  []int{0},
			wantLabel: "age >= 20",
		},
		{
			name:      "lt",
			exp:       Ordered("age", func(r orow) int { return r.Age }).LessThan(20),
			wantFail:  []int{1, 2},
			wantLabel: "age < 20",
		},
		{
			name:      "lte",
			exp:       Ordered("age", func(r orow) int { return r.Age }).LessOrEqual(20),
			wantFail:  []int{2},
			wantLabel: "age <= 20",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rep := NewSuite[orow](tc.exp).Validate(rows)
			res := rep.Results[0]
			if res.Name != tc.wantLabel {
				t.Fatalf("Name=%q, want %q", res.Name, tc.wantLabel)
			}
			if !reflect.DeepEqual(res.FailedIndices, tc.wantFail) {
				t.Fatalf("FailedIndices=%v, want %v", res.FailedIndices, tc.wantFail)
			}
		})
	}
}
