package viewchrome

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestDateColumnKeepsSummaryBesideDate(t *testing.T) {
	rows := []DayRow{{Date: "2026-11-15", Line: "Entry"}}
	for _, summary := range []string{"2 · 1h30m", "3 done · 1 failed"} {
		for _, width := range []int{24, 60, 120} {
			lines := DateColumnList(Theme{}, rows, 0, width, 5, func(string) string { return summary })
			heading := ansi.Strip(lines[0])
			want := ansi.Truncate(" SUN 15 NOV · "+summary, width, "…")
			if heading != want {
				t.Errorf("width %d: date heading = %q, want %q", width, heading, want)
			}
			if strings.Contains(heading, "   ") {
				t.Errorf("date metadata has a wide gap: %q", heading)
			}
		}
	}
}
