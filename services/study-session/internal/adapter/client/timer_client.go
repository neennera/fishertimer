package client

import (
	"context"
	"strings"
	"time"

	timerv1 "github.com/neennera/fishertimer/proto/studytimer/v1"
	"github.com/neennera/fishertimer/services/study-session/internal/domain"
)

// timerCallTimeout bounds one call to Study Timer.
const timerCallTimeout = 3 * time.Second

// TimerClient reads room timers from Study Timer over gRPC
// (StudyTimerService.GetRoomTimers).
type TimerClient struct {
	client timerv1.StudyTimerServiceClient
}

func NewTimerClient(client timerv1.StudyTimerServiceClient) *TimerClient {
	return &TimerClient{client: client}
}

func (c *TimerClient) RoomTimers(ctx context.Context, sessionID string) ([]domain.TimerActivity, error) {
	ctx, cancel := context.WithTimeout(ctx, timerCallTimeout)
	defer cancel()
	resp, err := c.client.GetRoomTimers(ctx, &timerv1.GetRoomTimersRequest{SessionId: sessionID})
	if err != nil {
		return nil, err
	}
	activity := make([]domain.TimerActivity, 0, len(resp.GetTimers()))
	for _, t := range resp.GetTimers() {
		lastUpdated, _ := time.Parse(time.RFC3339, t.GetLastUpdated())
		activity = append(activity, domain.TimerActivity{
			UserID:           t.GetUserId(),
			Status:           strings.TrimPrefix(t.GetStatus().String(), "TIMER_STATUS_"),
			Phase:            strings.TrimPrefix(t.GetPhase().String(), "TIMER_PHASE_"),
			LastUpdated:      lastUpdated,
			DurationSeconds:  int(t.GetDurationSeconds()),
			RemainingSeconds: int(t.GetRemainingSeconds()),
		})
	}
	return activity, nil
}
