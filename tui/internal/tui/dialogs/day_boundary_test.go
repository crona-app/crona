package dialogs

import (
	"strings"
	"testing"

	sharedtypes "crona.work/shared/types"
	controllerpkg "crona.work/tui/internal/tui/dialogs/controller"

	"github.com/charmbracelet/x/ansi"
)

func TestDayBoundaryRendersEveryDayInWeekStartOrder(t *testing.T) {
	state := controllerpkg.OpenEditDayBoundary(controllerpkg.State{DueDateWeekStart: sharedtypes.WeekStartSunday},
		sharedtypes.CoreSettingsKeyEndOfDay,
		sharedtypes.DayBoundarySchedule{Enabled: true, DefaultTime: "18:00", WeekdayOverrides: map[int]string{1: "17:00"}},
		"UTC")
	out := ansi.Strip(renderDayBoundaryDialog(testTheme(), sized(state, 120, 40)))
	if !strings.Contains(out, "End of Day") || !strings.Contains(out, "17:00   custom") ||
		!strings.Contains(out, "end-of-day prompts") {
		t.Fatalf("schedule render:\n%s", out)
	}
	if strings.Index(out, "Su ") > strings.Index(out, "Mo ") {
		t.Fatalf("a Sunday week should list Su first:\n%s", out)
	}
}
