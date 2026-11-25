package dialogruntime

import (
	"errors"
	"testing"

	"crona.work/tui/internal/tui/commands"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTaggedEditStopsWhenMutationFails(t *testing.T) {
	failure := errors.New("save rejected")
	tags := []string{"urgent"}
	tagCalls := 0
	cmd := thenSetTags(Deps{SetItemTags: func(string, int64, []string) tea.Cmd { return func() tea.Msg { tagCalls++; return nil } }}, "issue", 7, &tags, func() tea.Msg { return commands.ErrMsg{Err: failure} })
	result, ok := cmd().(commands.ErrMsg)
	if !ok || !errors.Is(result.Err, failure) {
		t.Fatalf("edit did not expose its save failure: %#v", result)
	}
	if tagCalls != 0 {
		t.Fatal("tag update ran after failed save")
	}
}

func TestTaggedEditWaitsForTagsBeforeReportingSuccess(t *testing.T) {
	failure := errors.New("tags rejected")
	tags := []string{"urgent"}
	cmd := thenSetTags(Deps{SetItemTags: func(string, int64, []string) tea.Cmd { return func() tea.Msg { return commands.ErrMsg{Err: failure} } }}, "issue", 7, &tags, func() tea.Msg { return commands.ReposLoadedMsg{} })
	result, ok := cmd().(commands.ErrMsg)
	if !ok || !errors.Is(result.Err, failure) {
		t.Fatalf("edit reported success before tags were saved: %#v", result)
	}
}

func TestSuccessfulMutationsWithoutRefreshMessages(t *testing.T) {
	cmd := thenSetTags(Deps{SetItemTags: func(string, int64, []string) tea.Cmd { return func() tea.Msg { return nil } }}, "issue", 7, &[]string{"urgent"}, func() tea.Msg { return nil })
	if msg := cmd(); msg != nil {
		t.Fatalf("unexpected save result: %#v", msg)
	}
}
