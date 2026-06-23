package gx

import (
	"math"
	"testing"
)

type nrow struct {
	Amount float64
}

func amountCol() NumericColumn[nrow, float64] {
	return Numeric("amount", func(r nrow) float64 { return r.Amount })
}

func TestNumericAverageBetween(t *testing.T) {
	passRows := []nrow{{5}, {10}, {15}}
	passRep := NewSuite(amountCol().AverageBetween(10, 10)).Validate(passRows)
	passRes := passRep.Results[0]
	if !passRes.Success {
		t.Fatal("average 10 should pass [10,10]")
	}
	if passRes.Column != "amount" {
		t.Fatalf("Column=%q, want amount", passRes.Column)
	}
	if passRes.Name != "amount average in [10,10]: got 10" {
		t.Fatalf("Name=%q", passRes.Name)
	}

	failRows := []nrow{{5}, {10}}
	failRep := NewSuite(amountCol().AverageBetween(9, 9)).Validate(failRows)
	failRes := failRep.Results[0]
	if failRes.Success {
		t.Fatal("average 7.5 should fail [9,9]")
	}
	if failRes.Column != "amount" {
		t.Fatalf("Column=%q, want amount", failRes.Column)
	}
	if failRes.Name != "amount average in [9,9]: got 7.5" {
		t.Fatalf("Name=%q", failRes.Name)
	}
}

func TestNumericMedianBetweenOddAndEven(t *testing.T) {
	oddRows := []nrow{{1}, {9}, {5}}
	oddRep := NewSuite(amountCol().MedianBetween(5, 5)).Validate(oddRows)
	if !oddRep.Results[0].Success {
		t.Fatalf("odd median should pass: Name=%q", oddRep.Results[0].Name)
	}
	if oddRep.Results[0].Name != "amount median in [5,5]: got 5" {
		t.Fatalf("Name=%q", oddRep.Results[0].Name)
	}

	evenRows := []nrow{{1}, {3}, {7}, {9}}
	evenRep := NewSuite(amountCol().MedianBetween(5, 5)).Validate(evenRows)
	if !evenRep.Results[0].Success {
		t.Fatalf("even median should pass: Name=%q", evenRep.Results[0].Name)
	}
	if evenRep.Results[0].Name != "amount median in [5,5]: got 5" {
		t.Fatalf("Name=%q", evenRep.Results[0].Name)
	}
}

func TestNumericStdDevBetweenPopulation(t *testing.T) {
	rows := []nrow{{2}, {4}, {4}, {4}, {5}, {5}, {7}, {9}}
	rep := NewSuite(amountCol().StdDevBetween(2, 2)).Validate(rows)
	res := rep.Results[0]
	if !res.Success {
		t.Fatalf("population stddev 2 should pass [2,2]: Name=%q", res.Name)
	}
	if res.Name != "amount standard deviation in [2,2]: got 2" {
		t.Fatalf("Name=%q", res.Name)
	}
}

func TestNumericSatisfyAggregate(t *testing.T) {
	rows := []nrow{{2}, {4}, {4}, {4}, {5}, {5}, {7}, {9}}

	var got NumericStats
	passRep := NewSuite(amountCol().SatisfyAggregate("stats match", func(s NumericStats) bool {
		got = s
		return s.Count == 8 &&
			s.Sum == 40 &&
			s.Average == 5 &&
			s.Median == 4.5 &&
			s.StdDevPopulation == 2 &&
			s.StdDevSample == math.Sqrt(32.0/7.0)
	})).Validate(rows)
	if !passRep.Results[0].Success {
		t.Fatalf("predicate should pass: Name=%q", passRep.Results[0].Name)
	}
	if got.Count != 8 {
		t.Fatalf("callback stats not recorded: %+v", got)
	}
	if passRep.Results[0].Name != "amount: stats match" {
		t.Fatalf("Name=%q", passRep.Results[0].Name)
	}

	failRep := NewSuite(amountCol().SatisfyAggregate("always false", func(NumericStats) bool {
		return false
	})).Validate(rows)
	if failRep.Results[0].Success {
		t.Fatal("false predicate should fail")
	}
	if failRep.Results[0].Name != "amount: always false" {
		t.Fatalf("Name=%q", failRep.Results[0].Name)
	}
}

func TestNumericExpectationNameStability(t *testing.T) {
	cases := []struct {
		name string
		exp  Expectation[nrow]
		want string
	}{
		{"average", amountCol().AverageBetween(10, 100), "amount average in [10,100]"},
		{"median", amountCol().MedianBetween(10, 100), "amount median in [10,100]"},
		{"stddev", amountCol().StdDevBetween(0, 25), "amount standard deviation in [0,25]"},
		{"satisfy", amountCol().SatisfyAggregate("CV <= 0.25", func(NumericStats) bool { return true }), "amount: CV <= 0.25"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.exp.Name(); got != tc.want {
				t.Fatalf("Name()=%q want %q", got, tc.want)
			}
		})
	}
}

func TestNumericEvaluateResultNameSuffix(t *testing.T) {
	vacExp := amountCol().AverageBetween(10, 100)
	vacRes := NewSuite(vacExp).Validate(nil).Results[0]
	if vacRes.Name != vacExp.Name() {
		t.Fatalf("vacuous Name=%q want %q", vacRes.Name, vacExp.Name())
	}

	passRes := NewSuite(amountCol().AverageBetween(10, 10)).Validate([]nrow{{10}}).Results[0]
	if passRes.Name != "amount average in [10,10]: got 10" {
		t.Fatalf("Name=%q", passRes.Name)
	}

	satExp := amountCol().SatisfyAggregate("ok", func(NumericStats) bool { return true })
	satRes := NewSuite(satExp).Validate([]nrow{{1}}).Results[0]
	if satRes.Name != satExp.Name() {
		t.Fatalf("Name=%q want %q", satRes.Name, satExp.Name())
	}
}

type irow struct {
	Qty int
}

func qtyCol() NumericColumn[irow, int] {
	return Numeric("qty", func(r irow) int { return r.Qty })
}

func TestNumericIntegerColumn(t *testing.T) {
	rows := []irow{{10}, {20}, {30}}
	rep := NewSuite(qtyCol().AverageBetween(20, 20)).Validate(rows)
	res := rep.Results[0]
	if !res.Success {
		t.Fatalf("integer average should pass: Name=%q", res.Name)
	}
	if res.Name != "qty average in [20,20]: got 20" {
		t.Fatalf("Name=%q", res.Name)
	}
	if res.Column != "qty" {
		t.Fatalf("Column=%q, want qty", res.Column)
	}
}

func TestNumericAggregatesEmptyInputVacuousPass(t *testing.T) {
	satisfyExp := amountCol().SatisfyAggregate("unused", func(NumericStats) bool { return false })
	wantNames := []string{
		"amount average in [10,100]",
		"amount median in [10,100]",
		"amount standard deviation in [0,25]",
		satisfyExp.Name(),
	}

	for _, rows := range [][]nrow{nil, {}} {
		label := "nil"
		if rows != nil {
			label = "empty"
		}
		t.Run(label, func(t *testing.T) {
			predicateCalled := false
			rep := NewSuite[nrow](
				amountCol().AverageBetween(10, 100),
				amountCol().MedianBetween(10, 100),
				amountCol().StdDevBetween(0, 25),
				amountCol().SatisfyAggregate("unused", func(NumericStats) bool {
					predicateCalled = true
					return false
				}),
			).Validate(rows)

			if !rep.OK() {
				t.Fatal("empty input should pass numeric aggregates vacuously")
			}
			if predicateCalled {
				t.Fatal("SatisfyAggregate predicate should not run on empty input")
			}
			for i, res := range rep.Results {
				if !res.Success {
					t.Fatalf("result[%d] should pass: Name=%q", i, res.Name)
				}
				if res.Name != wantNames[i] {
					t.Fatalf("result[%d] Name=%q, want %q", i, res.Name, wantNames[i])
				}
				if res.Total != 0 || res.FailedCount != 0 {
					t.Fatalf("result[%d] Total=%d FailedCount=%d, want 0 and 0", i, res.Total, res.FailedCount)
				}
			}
		})
	}
}

func TestNumericAggregateNonFiniteFailsWithoutPanic(t *testing.T) {
	cases := []struct {
		name string
		rows []nrow
	}{
		{name: "nan", rows: []nrow{{math.NaN()}}},
		{name: "inf", rows: []nrow{{math.Inf(1)}}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			satisfyCheck := "aggregate ok"
			rep := NewSuite(
				amountCol().AverageBetween(0, 100),
				amountCol().MedianBetween(0, 100),
				amountCol().StdDevBetween(0, 100),
				amountCol().SatisfyAggregate(satisfyCheck, func(NumericStats) bool { return true }),
			).Validate(tc.rows)

			wantNames := []string{
				"amount average in [0,100]: got non-finite value",
				"amount median in [0,100]: got non-finite value",
				"amount standard deviation in [0,100]: got non-finite value",
				"amount: aggregate ok: got non-finite value",
			}

			for i, res := range rep.Results {
				if res.Success {
					t.Fatalf("result[%d] should fail on non-finite input", i)
				}
				if res.Err != nil {
					t.Fatalf("result[%d] Err=%v, want nil", i, res.Err)
				}
				if res.Name != wantNames[i] {
					t.Fatalf("result[%d] Name=%q, want %q", i, res.Name, wantNames[i])
				}
			}
		})
	}
}

func TestNumericAggregateResultFieldsAreTableLevel(t *testing.T) {
	rows := []nrow{{5}, {10}}
	rep := NewSuite(amountCol().AverageBetween(9, 9)).Validate(rows)
	res := rep.Results[0]
	if res.Success {
		t.Fatal("want failing aggregate")
	}
	if res.Total != 0 || res.FailedCount != 0 || res.FailedPercent != 0 {
		t.Fatalf("Total=%d FailedCount=%d FailedPercent=%v, want zero table-level fields",
			res.Total, res.FailedCount, res.FailedPercent)
	}
	if len(res.FailedIndices) != 0 || len(res.SampleValues) != 0 {
		t.Fatalf("FailedIndices=%v SampleValues=%v, want empty", res.FailedIndices, res.SampleValues)
	}
}
