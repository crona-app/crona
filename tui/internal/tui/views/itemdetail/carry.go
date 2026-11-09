package itemdetail

import (
	"fmt"

	sharedtypes "crona.work/shared/types"
	helperpkg "crona.work/tui/internal/tui/helpers"
	types "crona.work/tui/internal/tui/views/types"
)

// CarryMarker is the compact carry-over marker for issue rows, unstyled so it takes the row's
// style: where a carried issue went and how long it was worked, or where a carried-in issue
// came from. It is "" for issues outside a carry chain.
func CarryMarker(state types.ContentState, issueID int64) string {
	entry := shownEntry(state, issueID)
	if entry == nil {
		return ""
	}
	if entry.Status == sharedtypes.DailyPlanEntryStatusCarried {
		text := "↪"
		if entry.CarriedToDate != nil {
			text += " " + helperpkg.FormatDisplayDate(*entry.CarriedToDate, state.Settings)
		}
		if entry.WorkedSeconds > 0 {
			text += " · " + helperpkg.FormatCompactDurationSeconds(entry.WorkedSeconds)
		}
		return text
	}
	if entry.CarriedFromDate == nil {
		return ""
	}
	text := "↩ from " + helperpkg.FormatDisplayDate(*entry.CarriedFromDate, state.Settings)
	if entry.CarryCount > 1 {
		text = fmt.Sprintf("↩ ×%d", entry.CarryCount)
	}
	if entry.Stalled {
		text += " stalled"
	}
	return text
}

// carryLines explain where the issue sits in its carry chain on the shown day.
func carryLines(theme types.Theme, state types.ContentState, issueID int64, width int) []string {
	entry := shownEntry(state, issueID)
	if entry == nil {
		return nil
	}
	lines := []string{}
	if entry.Status == sharedtypes.DailyPlanEntryStatusCarried {
		to := "the next day"
		if entry.CarriedToDate != nil {
			to = helperpkg.FormatDisplayDate(*entry.CarriedToDate, state.Settings)
		}
		lines = append(lines, label(theme, "Carried to", 14)+theme.StyleNormal.Render(fmt.Sprintf(
			"%s after %s of work", to, helperpkg.FormatCompactDurationSeconds(entry.WorkedSeconds),
		)))
	}
	if entry.CarriedFromDate != nil {
		text := helperpkg.FormatDisplayDate(*entry.CarriedFromDate, state.Settings)
		if entry.CarryCount > 1 {
			text += fmt.Sprintf(" · %d carries in a row", entry.CarryCount)
		}
		lines = append(lines, label(theme, "Carried from", 14)+theme.StyleNormal.Render(text))
	}
	if entry.Stalled && entry.Status != sharedtypes.DailyPlanEntryStatusCarried {
		lines = append(lines, label(theme, "", 14)+theme.StyleError.Render("stalled · consider splitting, re-estimating or abandoning it"))
	}
	if len(lines) == 0 {
		return nil
	}
	return truncateLines(append(heading(theme, "Carry-over"), lines...), width)
}
