package issues

import (
	"strings"
	"testing"

	types "crona.work/tui/internal/tui/views/types"
	"github.com/charmbracelet/x/ansi"
)

func TestWorkIssueDetailKeepsContextualActionsAndFooter(t *testing.T) {
	state := issuesState(200, 40)
	state.DetailSessionsIssueID = state.DefaultIssues[0].ID
	text := ansi.Strip(renderDetail(types.Theme{}, state, 76, 30))
	for _, want := range []string{"[enter] expand", "[s] status", "[d] due", "Effort/1h", "Sessions", "[m] log", "[e] edit", "[D] delete"} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %q:\n%s", want, text)
		}
	}
	lines := strings.Split(text, "\n")
	if !strings.Contains(lines[len(lines)-1], "[D] delete") {
		t.Fatal("general issue actions are not at the bottom")
	}
	state.Pane = "scope"
	text = ansi.Strip(renderDetail(types.Theme{}, state, 76, 30))
	if strings.Contains(text, "[s] status") || strings.Contains(text, "[e] edit") {
		t.Fatal("scope focus advertised actions targeting the issue")
	}
}

func TestVisibleDetailsOwnActionsWithoutDuplicatingCardHints(t *testing.T) {
	text := ansi.Strip(Render(types.Theme{}, issuesState(120, 40)))
	if count := strings.Count(text, "[f] focus"); count != 1 {
		t.Errorf("focus action count = %d, want one in the detail footer", count)
	}
	if count := strings.Count(text, "[s] status"); count != 1 {
		t.Errorf("status action count = %d, want one beside the detail status", count)
	}
}
