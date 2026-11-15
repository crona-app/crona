package viewchrome

import (
	"strings"
	"testing"

	"crona.work/tui/internal/api"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"
)

func TestAwayDayStripFitsAndRequiresExplicitAwayDate(t *testing.T) {
	settings := &api.CoreSettings{AwayDates: []string{"2026-10-14"}, RestSpecificDates: []string{"2026-10-15"}}
	for _, width := range []int{30, 60, 120} {
		line := AwayDayStrip(Theme{ColorGreen: "10"}, "2026-10-14", settings, width)
		if ansi.StringWidth(line) != width || !strings.Contains(ansi.Strip(line), "Away day") || !strings.Contains(ansi.Strip(line), "2026-10-14") {
			t.Fatalf("width %d: invalid away strip %q", width, line)
		}
	}
	for _, date := range []string{"2026-10-15", "2026-10-16"} {
		if got := AwayDayStrip(Theme{}, date, settings, 80); got != "" {
			t.Fatalf("date %s: unexpected away strip %q", date, got)
		}
	}
}

func TestAwayBandPreservesForegroundAndGlyph(t *testing.T) {
	previous := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(previous)
	glyph := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff00ff")).Render("⣿")
	got := AwayBand(glyph)
	if ansi.Strip(got) != "⣿" || !strings.Contains(got, "48;2;") || !strings.Contains(got, "38;2;255;0;255") {
		t.Fatalf("band lost background, foreground, or glyph: %q", got)
	}
}
