package gx

import (
	"errors"
	"strings"
	"testing"
)

func TestReportOK(t *testing.T) {
	r := Report{Results: []Result{{Success: true}, {Success: true}}}
	if !r.OK() {
		t.Fatal("all-success report should be OK")
	}
	r.Results[1].Success = false
	if r.OK() {
		t.Fatal("report with a failure should not be OK")
	}
}

func TestReportFailures(t *testing.T) {
	r := Report{Results: []Result{
		{Name: "a", Success: true},
		{Name: "b", Success: false},
	}}
	f := r.Failures()
	if len(f) != 1 || f[0].Name != "b" {
		t.Fatalf("Failures() = %v, want one result named b", f)
	}
}

func TestReportErr(t *testing.T) {
	if err := (Report{Results: []Result{{Success: true}}}).Err(); err != nil {
		t.Fatalf("OK report Err() = %v, want nil", err)
	}
	err := (Report{Results: []Result{{Name: "b", Success: false}}}).Err()
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("Err() = %T, want *ValidationError", err)
	}
	if !strings.Contains(err.Error(), "b") {
		t.Fatalf("Error() = %q, want it to mention failed expectation b", err.Error())
	}
}
