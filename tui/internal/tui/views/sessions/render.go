// Package sessions renders active sessions and session history screens.
package sessions

import (
	types "crona.work/tui/internal/tui/views/types"
	viewui "crona.work/tui/internal/tui/views/ui"
)

func Render(theme types.Theme, state types.ContentState) string {
	return viewui.NewLayout(theme, state, newView().Render).RenderView()
}
