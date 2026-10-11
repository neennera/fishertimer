package handler

import (
	"context"
	"errors"
	"log"
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
	phase := domain.PhaseWork
	if req.GetPhase() == timerv1.TimerPhase_TIMER_PHASE_REST {
		phase = domain.PhaseRest
	}
	return toTimerResponse(h.uc.StartTimer(ctx, req.GetSessionId(), req.GetUserId(), phase, int(req.GetDurationMinutes())))
}

func (h *GRPCHandler) GetTimer(ctx context.Context, req *timerv1.GetTimerRequest) (*timerv1.TimerStateResponse, error) {
	return toTimerResponse(h.uc.GetTimer(ctx, req.GetSessionId(), req.GetUserId()))
}

func (h *GRPCHandler) GetRoomTimers(ctx context.Context, req *timerv1.GetRoomTimersRequest) (*timerv1.GetRoomTimersResponse, error) {
	views, err := h.uc.GetRoomTimers(ctx, req.GetSessionId())
	if err != nil {
		return nil, toStatus(err)
	}
	resp := &timerv1.GetRoomTimersResponse{Timers: make([]*timerv1.TimerStateResponse, 0, len(views))}
	for i := range views {
		resp.Timers = append(resp.Timers, ToProto(&views[i]))
	}
	return resp, nil
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
		CyclesCompleted:   int32(stats.CyclesCompleted),
	}
	if !stats.LastActive.IsZero() {
		resp.LastActive = stats.LastActive.Format(time.RFC3339)
	}
	return resp, nil
}

func toTimerResponse(v *domain.View, err error) (*timerv1.TimerStateResponse, error) {
	if err != nil {
		return nil, toStatus(err)
	}
	return ToProto(v), nil
}

// ToProto renders a timer view as the gRPC response.
func ToProto(v *domain.View) *timerv1.TimerStateResponse {
	resp := &timerv1.TimerStateResponse{
		SessionId:            v.SessionID,
		UserId:               v.UserID,
		Status:               toProtoStatus(v.Status),
		Phase:                toProtoPhase(v.Phase),
		WorkMinutes:          int32(v.Settings.WorkMinutes),
		RestMinutes:          int32(v.Settings.RestMinutes),
		CurrentCycle:         int32(v.CompletedWork),
		DurationSeconds:      int32(v.DurationSeconds),
		RemainingSeconds:     int32(v.RemainingSeconds),
		State:                string(v.State),
		CycleId:              v.CycleID,
		PausedTotalSeconds:   int32(v.PausedTotalSeconds),
		FocusSeconds:         int32(v.FocusSeconds),
		LastCompletedCycleId: v.LastCompletedID,
		MinWorkMinutes:       int32(v.Limits.MinWorkMinutes),
		MaxWorkMinutes:       int32(v.Limits.MaxWorkMinutes),
		MinRestMinutes:       int32(v.Limits.MinRestMinutes),
		MaxRestMinutes:       int32(v.Limits.MaxRestMinutes),
		MaxPauseMinutes:      int32(v.Limits.MaxPauseMinutes),
	}
	if !v.LastUpdated.IsZero() {
		resp.LastUpdated = v.LastUpdated.UTC().Format(time.RFC3339)
	}
	if v.StartedAt != nil {
		resp.StartedAt = v.StartedAt.UTC().Format(time.RFC3339)
	}
	if v.LastCompletedAt != nil {
		resp.LastCompletedAt = v.LastCompletedAt.UTC().Format(time.RFC3339)
	}
	return resp
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
	case errors.Is(err, domain.ErrInvalidState), errors.Is(err, domain.ErrNotFinished):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, domain.ErrTimerClosed):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, domain.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	default:
		log.Printf("study-timer: internal error: %v", err)
		return status.Error(codes.Internal, "internal error")
	}
}
