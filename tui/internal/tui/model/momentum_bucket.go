package model

import (
	"fmt"

	"crona.work/tui/internal/api"
	"crona.work/tui/internal/tui/commands"

	tea "github.com/charmbracelet/bubbletea"
)

// clampMomentumBucket keeps a bucket offset within the selected momentum's series.
func (m Model) clampMomentumBucket(offset int) int {
	card, ok := m.selectedMomentumCard()
	if !ok || len(card.Series) == 0 {
		return 0
	}
	return min(max(0, offset), len(card.Series)-1)
}

// momentumBucketDate is a date inside the chosen bucket, no later than the shown date, which
// the daemon resolves to that bucket.
func (m Model) momentumBucketDate(card api.MomentumCard) string {
	date := m.currentDate()
	if len(card.Series) == 0 {
		return date
	}
	point := card.Series[len(card.Series)-1-min(m.momentumBucket, len(card.Series)-1)]
	end := point.EndDate
	if end == "" {
		end = point.BucketKey
	}
	if end != "" && end < date {
		return end
	}
	return date
}

func (m Model) momentumBucketKeyFor(card api.MomentumCard) string {
	return fmt.Sprintf("%s|%s|%d", card.Definition.ID, m.momentumBucketDate(card), m.currentMomentumWindowDays())
}

// ensureMomentumBucket loads the chosen bucket's contributors on Momentum when they are not
// what was last asked for.
func (m Model) ensureMomentumBucket() (Model, tea.Cmd) {
	if m.client == nil || m.view != ViewMomentum {
		return m, nil
	}
	card, ok := m.selectedMomentumCard()
	if !ok {
		return m, nil
	}
	key := m.momentumBucketKeyFor(*card)
	if key == m.momentumBucketWant {
		return m, nil
	}
	m.momentumBucketWant = key
	return m, commands.LoadMomentumBucket(m.client, key, card.Definition.ID, m.momentumBucketDate(*card), m.currentMomentumWindowDays())
}

// momentumBucketShown is the contributors when they answer the current bucket, else nil.
func (m Model) momentumBucketShown() *api.MomentumDetail {
	card, ok := m.selectedMomentumCard()
	if !ok || m.momentumBucketKey != m.momentumBucketKeyFor(*card) {
		return nil
	}
	return m.momentumBucketDetail
}
