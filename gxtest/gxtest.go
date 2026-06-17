// Package gxtest adapts a gx.Suite to Go tests. It mirrors the assert/require
// split: Check reports failures and continues; Require stops the test.
package gxtest

import "github.com/dashmug/gx"

// TestingT is the subset of *testing.T / *testing.B used by this package; both
// satisfy it implicitly, so callers pass their *testing.T unchanged.
type TestingT interface {
	Helper()
	Errorf(format string, args ...any)
	Fatalf(format string, args ...any)
}

// Check validates rows against the suite. On failure it calls t.Errorf with the
// rendered report and the test continues. It returns true when all pass.
func Check[T any](t TestingT, s *gx.Suite[T], rows []T) bool {
	t.Helper()
	rep := s.Validate(rows)
	if rep.OK() {
		return true
	}
	t.Errorf("gx: data quality failed\n%s", rep)
	return false
}

// Require validates rows against the suite. On failure it calls t.Fatalf with
// the rendered report, stopping the test.
func Require[T any](t TestingT, s *gx.Suite[T], rows []T) {
	t.Helper()
	rep := s.Validate(rows)
	if !rep.OK() {
		t.Fatalf("gx: data quality failed\n%s", rep)
	}
}
