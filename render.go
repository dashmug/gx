package gx

import (
	"fmt"
	"strings"
)

// maxDisplay caps how many sample values / indices String() shows. The Result
// fields still retain everything -- this is presentation only.
const maxDisplay = 10

// String renders one result as a single line, prefixed ✓ or ✗.
func (r Result) String() string {
	if r.Success {
		return fmt.Sprintf("✓ %s (%d rows)", r.Name, r.Total)
	}
	if r.FailedCount == 0 { // table-level style: Success carries the verdict
		return fmt.Sprintf("✗ %s", r.Name)
	}
	return fmt.Sprintf("✗ %s  %d/%d failed (%.1f%%)  e.g. %s @ %s",
		r.Name,
		r.FailedCount,
		r.Total,
		r.FailedPercent,
		truncList(r.SampleValues, maxDisplay),
		truncList(r.FailedIndices, maxDisplay),
	)
}

// String renders the whole report: a summary header plus one line per result.
func (r Report) String() string {
	passed := 0
	for _, res := range r.Results {
		if res.Success {
			passed++
		}
	}

	var b strings.Builder
	fmt.Fprintf(&b, "gx report: %d/%d expectations passed", passed, len(r.Results))
	for _, res := range r.Results {
		b.WriteString("\n  ")
		b.WriteString(res.String())
	}
	return b.String()
}

// truncList renders a bracketed list, truncating after max items with an ellipsis.
func truncList[E any](xs []E, max int) string {
	var b strings.Builder
	b.WriteByte('[')
	for i, x := range xs {
		if i >= max {
			b.WriteString(" …")
			break
		}
		if i > 0 {
			b.WriteByte(' ')
		}
		fmt.Fprintf(&b, "%v", x)
	}
	b.WriteByte(']')
	return b.String()
}
