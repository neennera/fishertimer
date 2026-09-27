package handler

import (
	"context"
	"errors"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	timerv1 "github.com/neennera/fishertimer/proto/studytimer/v1"
	"github.com/neennera/fishertimer/services/study-timer/internal/domain"
	"github.com/neennera/fishertimer/services/study-timer/internal/usecase"
)

// GRPCHandler is the gRPC driving adapter. It translates StudyTimerService
// RPCs into usecase calls, the same way HTTPHandler does for REST.
type GRPCHandler struct {
	timerv1.UnimplementedStudyTimerServiceServer
	uc usecase.Usecase
}

func NewGRPC(uc usecase.Usecase) *GRPCHandler {
	return &GRPCHandler{uc: uc}
}

func (h *GRPCHandler) StartTimer(ctx context.Context, req *timerv1.StartTimerRequest) (*timerv1.TimerStateResponse, error) {
	return toTimerResponse(h.uc.StartTimer(ctx, req.GetSessionId(), req.GetUserId()))
}

func (h *GRPCHandler) GetTimer(ctx context.Context, req *timerv1.GetTimerRequest) (*timerv1.TimerStateResponse, error) {
	return toTimerResponse(h.uc.GetTimer(ctx, req.GetSessionId(), req.GetUserId()))
}

func (h *GRPCHandler) PauseTimer(ctx context.Context, req *timerv1.PauseTimerRequest) (*timerv1.TimerStateResponse, error) {
	return toTimerResponse(h.uc.PauseTimer(ctx, req.GetSessionId(), req.GetUserId()))
}

func (h *GRPCHandler) ResumeTimer(ctx context.Context, req *timerv1.ResumeTimerRequest) (*timerv1.TimerStateResponse, error) {
	return toTimerResponse(h.uc.ResumeTimer(ctx, req.GetSessionId(), req.GetUserId()))
}

func (h *GRPCHandler) StopTimer(ctx context.Context, req *timerv1.StopTimerRequest) (*timerv1.TimerStateResponse, error) {
	return toTimerResponse(h.uc.StopTimer(ctx, req.GetSessionId(), req.GetUserId()))
}

func (h *GRPCHandler) ResetTimer(ctx context.Context, req *timerv1.ResetTimerRequest) (*timerv1.TimerStateResponse, error) {
	return toTimerResponse(h.uc.ResetTimer(ctx, req.GetSessionId(), req.GetUserId()))
}

func (h *GRPCHandler) CompleteCycle(ctx context.Context, req *timerv1.CompleteCycleRequest) (*timerv1.TimerStateResponse, error) {
	return toTimerResponse(h.uc.CompleteCycle(ctx, req.GetSessionId(), req.GetUserId()))
}

func (h *GRPCHandler) SkipRest(ctx context.Context, req *timerv1.SkipRestRequest) (*timerv1.TimerStateResponse, error) {
	return toTimerResponse(h.uc.SkipRest(ctx, req.GetSessionId(), req.GetUserId()))
}

func (h *GRPCHandler) UpdateTimerSetting(ctx context.Context, req *timerv1.UpdateTimerSettingRequest) (*timerv1.TimerStateResponse, error) {
	return toTimerResponse(h.uc.UpdateTimerSetting(ctx, req.GetSessionId(), req.GetUserId(),
		int(req.GetWorkMinutes()), int(req.GetRestMinutes())))
}

func (h *GRPCHandler) TimerStatistics(ctx context.Context, req *timerv1.TimerStatisticsRequest) (*timerv1.TimerStatisticsResponse, error) {
	stats, err := h.uc.TimerStatistics(ctx, req.GetUserId())
	if err != nil {
		return nil, toStatus(err)
	}
	resp := &timerv1.TimerStatisticsResponse{
		UserId:            stats.UserID,
		TotalSessions:     int32(stats.SessionsJoined),
		TotalFocusMinutes: int32(stats.TotalFocusMinutes),
	}
	if !stats.LastActive.IsZero() {
		resp.LastActive = stats.LastActive.Format(time.RFC3339)
	}
	return resp, nil
}

func toTimerResponse(t *domain.TimerState, err error) (*timerv1.TimerStateResponse, error) {
	if err != nil {
		return nil, toStatus(err)
	}
	now := time.Now().UTC()
	resp := &timerv1.TimerStateResponse{
		SessionId:        t.SessionID,
		UserId:           t.UserID,
		Status:           toProtoStatus(t.Status),
		Phase:            toProtoPhase(t.Phase),
		WorkMinutes:      int32(t.WorkMinutes),
		RestMinutes:      int32(t.RestMinutes),
		CurrentCycle:     int32(t.CurrentCycle),
		DurationSeconds:  int32(t.Duration().Seconds()),
		RemainingSeconds: int32(t.RemainingSeconds(now)),
	}
	if !t.LastUpdated.IsZero() {
		resp.LastUpdated = t.LastUpdated.Format(time.RFC3339)
	}
	return resp, nil
}

func toProtoStatus(s domain.TimerStatus) timerv1.TimerStatus {
	switch s {
	case domain.StatusStopped:
		return timerv1.TimerStatus_TIMER_STATUS_STOPPED
	case domain.StatusRunning:
		return timerv1.TimerStatus_TIMER_STATUS_RUNNING
	case domain.StatusPaused:
		return timerv1.TimerStatus_TIMER_STATUS_PAUSED
	default:
		return timerv1.TimerStatus_TIMER_STATUS_UNSPECIFIED
	}
}

func toProtoPhase(p domain.TimerPhase) timerv1.TimerPhase {
	switch p {
	case domain.PhaseWork:
		return timerv1.TimerPhase_TIMER_PHASE_WORK
	case domain.PhaseRest:
		return timerv1.TimerPhase_TIMER_PHASE_REST
	default:
		return timerv1.TimerPhase_TIMER_PHASE_UNSPECIFIED
	}
}

// toStatus maps a usecase error to the gRPC status code a client can act on.
func toStatus(err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalid):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrInvalidState):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, domain.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
