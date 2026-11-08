package viewchrome

import (
	"strings"
	"time"

	"github.com/charmbracelet/x/ansi"
)

// DateColumnList draws the cursor's window of rows grouped by day, laid out like Settings: a
// blank line between days, the day as a quiet heading with its summary dim at the right, and
// the day's rows indented under it. Rows off the cursor are faint.
func DateColumnList(
	theme Theme,
	rows []DayRow,
	cursor, width, height int,
	summary func(date string) string,
) []string {
	if len(rows) == 0 || height < 1 {
		return nil
	}
	inner := max(1, width-2)
	lines := make([]dayLine, 0, len(rows)*2)
	for i, row := range rows {
		if i == 0 || row.Date != rows[i-1].Date {
			if i > 0 {
				lines = append(lines, dayLine{row: -1})
			}
			heading := " " + theme.StyleDim.Bold(true).Render(strings.ToUpper(dateColumnLabel(row.Date)))
			if summary != nil {
				if tail := summary(row.Date); tail != "" {
					gap := inner + 2 - ansi.StringWidth(heading) - ansi.StringWidth(tail)
					if gap >= 2 {
						heading += strings.Repeat(" ", gap) + theme.StyleDim.Render(tail)
					}
				}
			}
			lines = append(lines, dayLine{text: heading, row: -1})
		}
		line := " " + row.Line
		if i != cursor {
			line = Faint(line)
		}
		lines = append(lines, dayLine{text: line, row: i})
	}
	return dividedList(theme, lines, cursor, width, height)
}

func dateColumnLabel(date string) string {
	if day, err := time.Parse("2006-01-02", date); err == nil {
		return day.Format("Mon 2 Jan")
	}
	return date
}
