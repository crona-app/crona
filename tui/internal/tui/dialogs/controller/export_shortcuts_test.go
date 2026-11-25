package controller

import (
	"slices"
	"testing"

	sharedtypes "crona.work/shared/types"
	"crona.work/tui/internal/api"

	tea "github.com/charmbracelet/bubbletea"
)

func TestCommonExportsSubmitFromFirstMenu(t *testing.T) {
	for _, tc := range []struct {
		label  string
		kind   sharedtypes.ExportReportKind
		format sharedtypes.ExportFormat
		mode   sharedtypes.ExportOutputMode
	}{
		{"Daily report: write PDF file", sharedtypes.ExportReportKindDaily, sharedtypes.ExportFormatPDF, sharedtypes.ExportOutputModeFile},
		{"Weekly summary: write Markdown file", sharedtypes.ExportReportKindWeekly, sharedtypes.ExportFormatMarkdown, sharedtypes.ExportOutputModeFile},
		{"Daily report: copy to clipboard", sharedtypes.ExportReportKindDaily, sharedtypes.ExportFormatMarkdown, sharedtypes.ExportOutputModeClipboard},
	} {
		t.Run(tc.label, func(t *testing.T) {
			state := OpenExportDaily(State{}, "2026-11-25", true, nil, nil, nil, nil, nil, 0, nil)
			state.ChoiceCursor = slices.Index(state.ChoiceItems, tc.label)
			if state.ChoiceCursor < 0 {
				t.Fatal("report missing from first menu")
			}
			next, action, status := Update(state, UpdateContext{}, state.CheckInDate, tea.KeyMsg{Type: tea.KeyEnter})
			if status != "" || !next.Processing || action == nil || action.ReportKind != tc.kind ||
				action.ReportFormat != tc.format || action.OutputMode != tc.mode || action.PresetID != "" ||
				action.CheckInDate != state.CheckInDate {
				t.Fatalf("unexpected export action %#v, status %q", action, status)
			}
		})
	}
}

func TestExportBackRestoresScopeAndRangeChoices(t *testing.T) {
	state := OpenExportDaily(State{}, "2026-11-25", false, []api.Repo{{ID: 1, Name: "Work"}, {ID: 2, Name: "Personal"}}, nil, nil, nil, nil, 0, nil)
	state.TemplateAssets = []sharedtypes.ExportTemplateAsset{{
		ReportKind: sharedtypes.ExportReportKindSummaryRange, AssetKind: sharedtypes.ExportAssetKindTemplateMarkdown,
		Customized: true,
	}}
	press := func(state State, key tea.KeyType) State {
		t.Helper()
		next, action, status := Update(state, UpdateContext{}, state.CheckInDate, tea.KeyMsg{Type: key})
		if action != nil || status != "" {
			t.Fatalf("unexpected export while navigating: %#v, %q", action, status)
		}
		return next
	}
	choose := func(state State, label string) State {
		t.Helper()
		state.ChoiceCursor = slices.Index(state.ChoiceItems, label)
		if state.ChoiceCursor < 0 {
			t.Fatalf("missing %q in %v", label, state.ChoiceItems)
		}
		return press(state, tea.KeyEnter)
	}
	repoScope := choose(choose(state, "Project Reports"), "Repo report: write Markdown file")
	repoScope = press(repoScope, tea.KeyDown)
	repoScope = press(press(repoScope, tea.KeyEsc), tea.KeyEnter)
	if repoScope.ChoiceItems[repoScope.ChoiceCursor] != "Personal" {
		t.Fatal("reopening scope lost the selected repo")
	}
	scope := choose(choose(state, "Project Reports"), "Issue rollup: write Markdown file")
	project := press(scope, tea.KeyEsc)
	if project.ChoiceItems[project.ChoiceCursor] != "Issue rollup: write Markdown file" {
		t.Fatal("Back lost the report selection")
	}
	root := press(project, tea.KeyEsc)
	if root.ChoiceItems[root.ChoiceCursor] != "Project Reports" {
		t.Fatal("Back lost the category selection")
	}
	custom := choose(choose(root, "Summary"), "Custom range")
	custom.Inputs[0].SetValue("2026-11-01")
	output := press(custom, tea.KeyEnter)
	output.ChoiceCursor = slices.Index(output.ChoiceItems, "Copy Markdown")
	style, action, status := Update(output, UpdateContext{}, output.CheckInDate,
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	if action != nil || status != "" || style.Kind != "export_preset" {
		t.Fatalf("style key should open the style picker, got %q, %#v, %q", style.Kind, action, status)
	}
	back := press(style, tea.KeyEsc)
	if back.Kind != "export_summary_output" || back.ChoiceItems[back.ChoiceCursor] != "Copy Markdown" {
		t.Fatal("Back from styles lost the chosen output")
	}
	back = press(back, tea.KeyEsc)
	if back.Kind != "export_summary_range" || len(back.Inputs) != 2 || back.Inputs[0].Value() != "2026-11-01" {
		t.Fatal("Back from output lost the custom dates")
	}
}
