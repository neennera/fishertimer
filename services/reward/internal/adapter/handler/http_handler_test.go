package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/neennera/fishertimer/services/reward/internal/adapter/handler"
	"github.com/neennera/fishertimer/services/reward/internal/domain"
	"github.com/neennera/fishertimer/services/reward/internal/usecase"
)

// fakeUsecase records the AwardInput the handler passes on.
type fakeUsecase struct {
	got    *usecase.AwardInput
	called int
}

func (f *fakeUsecase) AwardReward(ctx context.Context, in usecase.AwardInput) (*usecase.AwardResult, error) {
	f.called++
	f.got = &in
	return &usecase.AwardResult{CycleID: in.CycleID, Rewards: []domain.UnlockedReward{}}, nil
}

func (f *fakeUsecase) GetUserInventory(ctx context.Context, userID string) ([]domain.UnlockedReward, error) {
	return nil, nil
}

func (f *fakeUsecase) ListAllRewards(ctx context.Context) ([]domain.UnlockedReward, error) {
	return nil, nil
}

func (f *fakeUsecase) GetLastUpdate(ctx context.Context) (time.Time, error) {
	return time.Time{}, nil
}

func postAward(t *testing.T, body string) (*httptest.ResponseRecorder, *fakeUsecase) {
	t.Helper()
	uc := &fakeUsecase{}
	mux := http.NewServeMux()
	handler.New(uc).RegisterRoutes(mux)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/reward/award", strings.NewReader(body))
	mux.ServeHTTP(rec, req)
	return rec, uc
}

func TestAwardReward_CycleLength(t *testing.T) {
	cases := []struct {
		name        string
		body        string
		wantStatus  int
		wantMinutes int
		wantMessage string
	}{
		{
			name:        "work_duration only (Study Timer's payload, reason ignored)",
			body:        `{"user_id":"u1","session_id":"s1","cycle_id":"c1","work_duration":45,"participant_count":3,"reason":"CompleteCycle"}`,
			wantStatus:  http.StatusOK,
			wantMinutes: 45,
		},
		{
			name:        "work_minutes only",
			body:        `{"user_id":"u1","cycle_id":"c1","work_minutes":60}`,
			wantStatus:  http.StatusOK,
			wantMinutes: 60,
		},
		{
			name:        "both equal",
			body:        `{"user_id":"u1","cycle_id":"c1","work_minutes":30,"work_duration":30}`,
			wantStatus:  http.StatusOK,
			wantMinutes: 30,
		},
		{
			name:        "both different",
			body:        `{"user_id":"u1","cycle_id":"c1","work_minutes":30,"work_duration":45}`,
			wantStatus:  http.StatusBadRequest,
			wantMessage: "work_minutes and work_duration differ",
		},
		{
			name:        "neither",
			body:        `{"user_id":"u1","cycle_id":"c1","reason":"CompleteCycle"}`,
			wantStatus:  http.StatusBadRequest,
			wantMessage: "user_id, cycle_id and work_minutes (or work_duration)",
		},
		{
			name:        "negative work_duration",
			body:        `{"user_id":"u1","cycle_id":"c1","work_duration":-1}`,
			wantStatus:  http.StatusBadRequest,
			wantMessage: "user_id, cycle_id and work_minutes (or work_duration)",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rec, uc := postAward(t, c.body)
			if rec.Code != c.wantStatus {
				t.Fatalf("status = %d, want %d (body %q)", rec.Code, c.wantStatus, rec.Body.String())
			}
			if c.wantStatus != http.StatusOK {
				if uc.called != 0 {
					t.Fatal("usecase should not be called for a rejected request")
				}
				if !strings.Contains(rec.Body.String(), c.wantMessage) {
					t.Fatalf("body = %q, want it to contain %q", rec.Body.String(), c.wantMessage)
				}
				return
			}
			if uc.called != 1 || uc.got.WorkMinutes != c.wantMinutes {
				t.Fatalf("usecase got %+v after %d call(s), want WorkMinutes %d", uc.got, uc.called, c.wantMinutes)
			}
		})
	}
}
