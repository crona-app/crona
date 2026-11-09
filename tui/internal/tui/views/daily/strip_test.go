package daily

import (
	"fmt"
	"strings"
	"testing"

	helperpkg "crona.work/tui/internal/tui/helpers"
	viewcalendar "crona.work/tui/internal/tui/views/calendar"

	"github.com/charmbracelet/x/ansi"
)

// fieldsOf collapses runs of spaces so assertions ignore column padding.
func fieldsOf(row string) string {
	return strings.Join(strings.Fields(row), " ")
}

// cellIndex is the display column of needle in row, or -1.

func TestSubTabMetaShowsDateWeekAndKeys(t *testing.T) {
	state := dailySummaryTestState()
	date, _ := viewcalendar.ParseDate("2026-05-27")
	got := ansi.Strip(SubTabMeta(testTheme(), state, 80))
	for _, want := range []string{
		helperpkg.FormatDisplayDate("2026-05-27", nil),
		fmt.Sprintf("W%02d", viewcalendar.WeekNumber(date, state.WeekStart)),
		"‹ ,",
		". ›",
		"g today",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected sub-tab meta to contain %q, got %q", want, got)
		}
	}
}

func TestSubTabMetaLeadsWithToday(t *testing.T) {
	got := ansi.Strip(SubTabMeta(testTheme(), dailySummaryTestState(), 80))
	if !strings.HasPrefix(got, "‹ ,") || !strings.HasSuffix(got, "g today") {
		t.Fatalf("expected the stepper first and g today last, got %q", got)
	}
}

func TestSubTabMetaShortensToFit(t *testing.T) {
	state := dailySummaryTestState()
	date := helperpkg.FormatDisplayDate("2026-05-27", nil)
	for _, room := range []int{60, 40, 30, 20} {
		got := ansi.Strip(SubTabMeta(testTheme(), state, room))
		if ansi.StringWidth(got) > room || !strings.Contains(got, date) {
			t.Fatalf("room %d: expected the date within the room, got %q", room, got)
		}
	}
	if got := ansi.Strip(SubTabMeta(testTheme(), state, 30)); !strings.Contains(got, "‹ ,") ||
		!strings.Contains(got, ". ›") {
		t.Fatalf("expected the keys kept before the words drop out, got %q", got)
	}
}
