//go:build e2e

package e2e

import (
	"strings"
	"testing"
	"time"

	shareddto "crona.work/shared/dto"
	"crona.work/shared/protocol"
	sharedtypes "crona.work/shared/types"
)

func TestDailyDurationLimitsOverIPC(t *testing.T) {
	kernel := startTestKernel(t)
	defer kernel.close(t)
	today := time.Now().Format(time.DateOnly)
	for _, hours := range []float64{24.01, 68} {
		message := kernel.callError(t, protocol.MethodCheckInUpsert, shareddto.DailyCheckInUpsertRequest{Date: today, Mood: 3, Energy: 3, SleepHours: &hours})
		if !strings.Contains(message, "24h") {
			t.Fatalf("sleep %v: %s", hours, message)
		}
	}
	screen := 1441
	message := kernel.callError(t, protocol.MethodCheckInUpsert, shareddto.DailyCheckInUpsertRequest{Date: today, Mood: 3, Energy: 3, ScreenTimeMinutes: &screen})
	if !strings.Contains(message, "24h") {
		t.Fatalf("screen: %s", message)
	}
	hours := 24.0
	screen = 1440
	var saved sharedtypes.DailyCheckIn
	kernel.call(t, protocol.MethodCheckInUpsert, shareddto.DailyCheckInUpsertRequest{Date: today, Mood: 3, Energy: 3, SleepHours: &hours, ScreenTimeMinutes: &screen}, &saved)
	if saved.SleepHours == nil || *saved.SleepHours != 24 || saved.ScreenTimeMinutes == nil || *saved.ScreenTimeMinutes != 1440 {
		t.Fatalf("boundary values lost: %+v", saved)
	}
	repo := createRepo(t, kernel, "Limits")
	stream := createStream(t, kernel, repo.ID, "main")
	issue := createIssue(t, kernel, stream.ID, "Limits", nil)
	message = kernel.callError(t, protocol.MethodSessionLogManual, shareddto.ManualSessionLogRequest{IssueID: issue.ID, Date: today, WorkDurationSeconds: 86400, BreakDurationSeconds: 1})
	if !strings.Contains(message, "24h") {
		t.Fatalf("manual total: %s", message)
	}
	seconds := 68 * 3600
	message = kernel.callError(t, protocol.MethodTimerStart, shareddto.TimerStartRequest{IssueID: &issue.ID, HardLimitTotalSeconds: &seconds})
	if !strings.Contains(message, "24h") {
		t.Fatalf("timer: %s", message)
	}
}
