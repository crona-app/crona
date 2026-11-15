// Package away renders away-mode screens.
package away

import (
	viewtypes "crona.work/tui/internal/tui/views/types"
	viewui "crona.work/tui/internal/tui/views/ui"
)

func Render(theme viewtypes.Theme, state viewtypes.ContentState) string {
	return viewui.NewLayout(theme, state, newView().Render).RenderView()
}
