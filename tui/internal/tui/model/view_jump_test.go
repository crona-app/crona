package model

import (
	"testing"

	dialogctl "crona.work/tui/internal/tui/dialogs/controller"
	uistate "crona.work/tui/internal/tui/state"
)

func TestViewJumpGroupsSurviveTheModelRoundTrip(t *testing.T) {
	opened := dialogctl.OpenViewJumpFrom(dialogctl.State{}, uistate.ViewOrder(), uistate.ViewReview)
	got := Model{}.withDialogState(opened).dialogState()
	if len(got.ChoiceGroups) != len(opened.ChoiceGroups) || got.ChoiceGroups[0] == "" || got.ChoiceCurrent != "review" {
		t.Fatalf("expected the jump dialog's groups and current view kept, got groups=%q current=%q", got.ChoiceGroups, got.ChoiceCurrent)
	}
}
