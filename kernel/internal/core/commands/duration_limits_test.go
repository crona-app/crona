package commands

import (
	"context"
	"math"
	"strings"
	"testing"

	shareddto "crona.work/shared/dto"
	sharedtypes "crona.work/shared/types"
)

func TestCheckInRejectsNonFiniteSleepWithoutWriting(t *testing.T) {
	ctx, _, _ := newTimerTestContext(t, func() string { return "2026-05-24T12:00:00Z" })
	for _, hours := range []float64{math.NaN(), math.Inf(1), 68} {
		if _, err := UpsertDailyCheckIn(context.Background(), ctx, shareddto.DailyCheckInUpsertRequest{Date: "2026-05-24", Mood: 3, Energy: 3, SleepHours: &hours}); err == nil {
			t.Fatalf("accepted %v", hours)
		}
	}
	saved, err := GetDailyCheckIn(context.Background(), ctx, "2026-05-24")
	if err != nil || saved != nil {
		t.Fatalf("invalid write created check-in: %+v %v", saved, err)
	}
}

func TestTimerExtensionRejectsOverflowBeforeMutation(t *testing.T) {
	ctx, timer, issue := newTimerTestContext(t, func() string { return "2026-05-24T12:00:00Z" })
	mustMakeIssuePlanned(t, context.Background(), ctx, issue.ID)
	seconds := 23 * 3600
	_, err := timer.Start(context.Background(), nil, nil, &issue.ID, &shareddto.TimerStartRequest{IssueID: &issue.ID, HardLimitTotalSeconds: &seconds})
	if err != nil {
		t.Fatal(err)
	}
	_, err = timer.Extend(context.Background(), 2*3600)
	if err == nil || !strings.Contains(err.Error(), "24h") {
		t.Fatalf("extension: %v", err)
	}
}

func TestMomentumWorkLimitBeforeNormalization(t *testing.T) {
	for _, value := range []int{-1, 0, 86401, math.MaxInt} {
		_, err := CreateHabitStreakDefinition(context.Background(), nil, sharedtypes.HabitStreakDefinition{Name: "Work", TargetKind: sharedtypes.MomentumTargetKindContext, RequiredCount: value})
		if err == nil {
			t.Fatalf("accepted required work %d", value)
		}
	}
}
