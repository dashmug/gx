package gxtest_test

import (
	"fmt"
	"testing"

	"github.com/dashmug/gx"
	"github.com/dashmug/gx/gxtest"
)

// fakeT records calls so we can assert on Check/Require behavior without
// failing the real test.
type fakeT struct {
	helpers int
	errorfs int
	fatalfs int
	last    string
}

func (f *fakeT) Helper() { f.helpers++ }
func (f *fakeT) Errorf(format string, args ...any) {
	f.errorfs++
	f.last = fmt.Sprintf(format, args...)
}
func (f *fakeT) Fatalf(format string, args ...any) {
	f.fatalfs++
	f.last = fmt.Sprintf(format, args...)
}

func TestCheckPasses(t *testing.T) {
	f := &fakeT{}
	s := gx.NewSuite[int](gx.RowCountEqual[int](2))
	if !gxtest.Check(f, s, []int{1, 2}) {
		t.Fatal("Check should return true when the suite passes")
	}
	if f.errorfs != 0 {
		t.Fatalf("Errorf called %d times on success, want 0", f.errorfs)
	}
}

func TestCheckFails(t *testing.T) {
	f := &fakeT{}
	s := gx.NewSuite[int](gx.RowCountEqual[int](5))
	if gxtest.Check(f, s, []int{1, 2}) {
		t.Fatal("Check should return false when the suite fails")
	}
	if f.errorfs != 1 {
		t.Fatalf("Errorf called %d times, want 1", f.errorfs)
	}
	if f.helpers == 0 {
		t.Fatal("Check should call t.Helper()")
	}
}

func TestRequireFails(t *testing.T) {
	f := &fakeT{}
	gxtest.Require(f, gx.NewSuite[int](gx.RowCountEqual[int](5)), []int{1, 2})
	if f.fatalfs != 1 {
		t.Fatalf("Fatalf called %d times, want 1", f.fatalfs)
	}
	if f.helpers == 0 {
		t.Fatal("Require should call t.Helper()")
	}
}
