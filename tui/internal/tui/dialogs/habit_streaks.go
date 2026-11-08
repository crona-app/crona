package dialogs

import (
	"fmt"
	"strings"

	sharedtypes "crona.work/shared/types"
	sharedutils "crona.work/shared/utils"
	"crona.work/tui/internal/api"
)

func habitStreakWizardSteps() []string {
	return []string{"What counts", "Goal"}
}

func habitStreakPeriodLabel(period sharedtypes.HabitStreakPeriod) string {
	switch sharedtypes.NormalizeHabitStreakPeriod(period) {
	case sharedtypes.HabitStreakPeriodWeek:
		return "Weekly"
	case sharedtypes.HabitStreakPeriodMonth:
		return "Monthly"
	default:
		return "Daily"
	}
}

func habitStreakHabitSummary(ids []int64, habits []sharedtypes.HabitWithMeta) string {
	if len(ids) == 0 {
		return "None"
	}
	labels := make([]string, 0, len(ids))
	for _, id := range ids {
		for _, habit := range habits {
			if habit.ID == id {
				labels = append(labels, habit.Name)
				break
			}
		}
	}
	if len(labels) == 0 {
		return fmt.Sprintf("%d habits", len(ids))
	}
	if len(labels) <= 3 {
		return strings.Join(labels, ", ")
	}
	return fmt.Sprintf("%s, %s, %s +%d", labels[0], labels[1], labels[2], len(labels)-3)
}

func habitStreakPeriodUnitLabel(period sharedtypes.HabitStreakPeriod) string {
	switch sharedtypes.NormalizeHabitStreakPeriod(period) {
	case sharedtypes.HabitStreakPeriodWeek:
		return "week"
	case sharedtypes.HabitStreakPeriodMonth:
		return "month"
	default:
		return "day"
	}
}

func habitStreakTargetSelectionSummary(
	def sharedtypes.HabitStreakDefinition,
	repos []api.Repo,
	allIssues []api.IssueWithMeta,
	streams []api.Stream,
	habits []sharedtypes.HabitWithMeta,
) string {
	switch sharedtypes.NormalizeMomentumTargetKind(def.TargetKind) {
	case sharedtypes.MomentumTargetKindContext:
		if len(def.Contexts) == 0 {
			return "None"
		}
		labels := make([]string, 0, len(def.Contexts))
		for _, contextItem := range def.Contexts {
			labels = append(labels, momentumContextLabel(contextItem, repos, allIssues, streams))
		}
		if len(labels) <= 3 {
			return strings.Join(labels, ", ")
		}
		return fmt.Sprintf("%s, %s, %s +%d", labels[0], labels[1], labels[2], len(labels)-3)
	default:
		return habitStreakHabitSummary(def.HabitIDs, habits)
	}
}

func momentumContextLabel(
	contextItem sharedtypes.MomentumContext,
	repos []api.Repo,
	allIssues []api.IssueWithMeta,
	streams []api.Stream,
) string {
	repoLabel := ""
	for _, repo := range repos {
		if repo.ID == contextItem.RepoID {
			repoLabel = strings.TrimSpace(repo.Name)
			break
		}
	}
	if repoLabel == "" {
		repoLabel = fmt.Sprintf("%d", contextItem.RepoID)
	}
	if contextItem.StreamID == nil {
		return repoLabel + " / Any stream"
	}
	streamID := *contextItem.StreamID
	streamLabel := ""
	for _, stream := range streams {
		if stream.ID == streamID {
			streamLabel = strings.TrimSpace(stream.Name)
			break
		}
	}
	if streamLabel == "" {
		for _, issue := range allIssues {
			if issue.RepoID == contextItem.RepoID && issue.StreamID == streamID {
				streamLabel = strings.TrimSpace(issue.StreamName)
				if streamLabel != "" {
					break
				}
			}
		}
	}
	if streamLabel == "" {
		streamLabel = fmt.Sprintf("%d", streamID)
	}
	return repoLabel + " / " + streamLabel
}

func momentumContextRedundancyWarnings(
	contexts []sharedtypes.MomentumContext,
	repos []api.Repo,
	allIssues []api.IssueWithMeta,
	streams []api.Stream,
) []string {
	redundancies := sharedtypes.MomentumContextRedundancies(contexts)
	if len(redundancies) == 0 {
		return nil
	}
	out := make([]string, 0, len(redundancies))
	for _, redundancy := range redundancies {
		repoWideLabel := momentumContextLabel(redundancy.RepoWideContext, repos, allIssues, streams)
		redundantLabels := make([]string, 0, len(redundancy.RedundantContexts))
		for _, contextItem := range redundancy.RedundantContexts {
			redundantLabels = append(
				redundantLabels,
				momentumContextLabel(contextItem, repos, allIssues, streams),
			)
		}
		label := ""
		switch len(redundantLabels) {
		case 0:
			continue
		case 1:
			label = redundantLabels[0]
		case 2:
			label = redundantLabels[0] + " and " + redundantLabels[1]
		default:
			label = fmt.Sprintf("%d specific stream selections", len(redundantLabels))
		}
		out = append(out, fmt.Sprintf("Warning: %s already covers %s", repoWideLabel, label))
	}
	return out
}

func habitMomentumViewCapacity(
	def sharedtypes.HabitStreakDefinition,
	habits []sharedtypes.HabitWithMeta,
) sharedutils.HabitMomentumCapacity {
	habitMap := make(map[int64]sharedtypes.Habit, len(habits))
	for _, habit := range habits {
		habitMap[habit.ID] = habit.Habit
	}
	selected := make([]sharedtypes.Habit, 0, len(def.HabitIDs))
	for _, habitID := range def.HabitIDs {
		habit, ok := habitMap[habitID]
		if !ok {
			return sharedutils.HabitMomentumCapacity{
				Reason: "selected habits are no longer available",
			}
		}
		selected = append(selected, habit)
	}
	return sharedutils.HabitMomentumCapacityForSelection(selected, def.Period, def.MatchMode)
}
