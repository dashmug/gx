package gx

import (
	"reflect"
	"testing"
)

type evalColumnRow struct {
	Age int
}

func TestEvalColumnRecordsFailuresAndSamples(t *testing.T) {
	rows := []evalColumnRow{{Age: 1}, {Age: -2}, {Age: 0}, {Age: 3}}

	res := EvalColumn(
		"age positive",
		"age",
		rows,
		func(r evalColumnRow) int { return r.Age },
		func(v int) bool { return v > 0 },
		EvalOptions{SampleCap: 1},
	)

	if res.Success {
		t.Fatal("want failure")
	}
	if res.Name != "age positive" || res.Column != "age" {
		t.Fatalf("Name=%q Column=%q", res.Name, res.Column)
	}
	if res.Total != 4 || res.FailedCount != 2 {
		t.Fatalf("Total=%d FailedCount=%d, want 4 and 2", res.Total, res.FailedCount)
	}
	if !reflect.DeepEqual(res.FailedIndices, []int{1, 2}) {
		t.Fatalf("FailedIndices=%v, want [1 2]", res.FailedIndices)
	}
	if !reflect.DeepEqual(res.SampleValues, []any{-2}) {
		t.Fatalf("SampleValues=%v, want [-2]", res.SampleValues)
	}
	if res.FailedPercent != 50 {
		t.Fatalf("FailedPercent=%v, want 50", res.FailedPercent)
	}
}

func TestEvalColumnEmptyInputVacuousPass(t *testing.T) {
	res := EvalColumn(
		"empty",
		"age",
		nil,
		func(r evalColumnRow) int { return r.Age },
		func(int) bool { return false },
		EvalOptions{SampleCap: 1},
	)

	if !res.Success {
		t.Fatal("empty input should pass vacuously")
	}
	if res.Total != 0 || res.FailedCount != 0 {
		t.Fatalf("Total=%d FailedCount=%d, want 0 and 0", res.Total, res.FailedCount)
	}
	if len(res.FailedIndices) != 0 {
		t.Fatalf("FailedIndices=%v, want empty", res.FailedIndices)
	}
	if len(res.SampleValues) != 0 {
		t.Fatalf("SampleValues=%v, want empty", res.SampleValues)
	}
	if res.FailedPercent != 0 {
		t.Fatalf("FailedPercent=%v, want 0", res.FailedPercent)
	}
}
