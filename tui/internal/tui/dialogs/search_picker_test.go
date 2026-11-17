package dialogs

import (
	"github.com/charmbracelet/x/ansi"
	"strings"
	"testing"
)

func TestSearchPickerShowsListWithoutQuery(t *testing.T) {
	options := []string{"First repo", "Second repo", "Third repo"}
	rows := SearchPicker(testTheme(), PickerSpec{Label: "Repo", Column: 6, Focused: true, Options: options, Input: "type to filter"}, 46)
	text := ansi.Strip(strings.Join(rows, "\n"))
	for _, name := range options {
		if !strings.Contains(text, name) {
			t.Fatalf("missing %s: %s", name, text)
		}
	}
	if !strings.Contains(text, "▶ First repo") || strings.Contains(text, "filter:") || strings.Contains(text, "‹") {
		t.Fatalf("unexpected picker: %s", text)
	}
	fewer := SearchPicker(testTheme(), PickerSpec{Focused: true, Options: options[:1]}, 46)
	empty := SearchPicker(testTheme(), PickerSpec{Focused: true}, 46)
	if len(rows) != len(fewer) || len(rows) != len(empty) {
		t.Fatal("filtering changed picker height")
	}
}

func TestSearchPickerScrollsAndWrapsSelectedName(t *testing.T) {
	name := "A repository name long enough to need its own wrapped readout"
	options := []string{"One", "Two", "Three", "Four", "Five", "Six", name}
	rows := SearchPicker(testTheme(), PickerSpec{Label: "Repo", Column: 6, Focused: true, Options: options, Selected: 6}, 42)
	text := ansi.Strip(strings.Join(rows, "\n"))
	if strings.Contains(text, "  One") || !strings.Contains(text, "▶ A repository") || !strings.Contains(text, "wrapped readout") {
		t.Fatalf("list did not scroll or wrap: %s", text)
	}
	for _, row := range rows {
		if ansi.StringWidth(row) > 42 {
			t.Fatalf("row exceeds width: %q", row)
		}
	}
}

func TestSearchPickerDoesNotRepeatVisibleCreateOption(t *testing.T) {
	for _, option := range []string{"Create New Repo: example", "Create New Stream: example", "Existing repo"} {
		rows := SearchPicker(testTheme(), PickerSpec{Focused: true, Options: []string{option}}, 60)
		text := ansi.Strip(strings.Join(rows, "\n"))
		if count := strings.Count(text, option); count != 1 {
			t.Fatalf("option %q appears %d times: %s", option, count, text)
		}
	}
}
