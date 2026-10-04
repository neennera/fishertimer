package handler

import (
	"context"
	"errors"
	"log"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sessionv1 "github.com/neennera/fishertimer/proto/studysession/v1"
	"github.com/neennera/fishertimer/services/study-session/internal/domain"
	"github.com/neennera/fishertimer/services/study-session/internal/usecase"
)

// GRPCHandler is the gRPC driving adapter used by the API Gateway and the
// Admin service. It translates StudySessionService RPCs into usecase calls.
type GRPCHandler struct {
	sessionv1.UnimplementedStudySessionServiceServer
	uc usecase.Usecase
}

func NewGRPC(uc usecase.Usecase) *GRPCHandler {
	return &GRPCHandler{uc: uc}
}

func (h *GRPCHandler) CreateSession(ctx context.Context, req *sessionv1.CreateSessionRequest) (*sessionv1.StudySessionResponse, error) {
	sess, err := h.uc.CreateSession(ctx, req.GetName(), req.GetCreatorId(), req.GetCreatorDisplayName(), int(req.GetParticipantLimit()))
	if err != nil {
		return nil, toStatus(err)
	}
	return toSessionResponse(sess), nil
}

func (h *GRPCHandler) JoinSession(ctx context.Context, req *sessionv1.JoinSessionRequest) (*sessionv1.JoinSessionResponse, error) {
	sess, err := h.uc.JoinSession(ctx, req.GetSessionId(), req.GetUserId(), req.GetDisplayName())
	if err != nil {
		return nil, toStatus(err)
	}
	return &sessionv1.JoinSessionResponse{Success: true, Session: toSessionResponse(sess)}, nil
}

func (h *GRPCHandler) LeaveSession(ctx context.Context, req *sessionv1.LeaveSessionRequest) (*sessionv1.LeaveSessionResponse, error) {
	res, err := h.uc.LeaveSession(ctx, req.GetSessionId(), req.GetUserId(), req.GetReason())
	if err != nil {
		return nil, toStatus(err)
	}
	resp := &sessionv1.LeaveSessionResponse{
		Success:      true,
		Left:         res.Left,
		SessionEnded: res.SessionEnded,
		SessionId:    req.GetSessionId(),
		UserId:       req.GetUserId(),
	}
	if p := res.Participant; p != nil {
		resp.JoinedAt = formatTime(p.JoinedAt)
		if p.LeftAt != nil {
			resp.LeftAt = formatTime(*p.LeftAt)
		}
	}
	return resp, nil
}

func (h *GRPCHandler) EndSession(ctx context.Context, req *sessionv1.EndSessionRequest) (*sessionv1.StudySessionResponse, error) {
	sess, err := h.uc.EndSession(ctx, req.GetSessionId(), req.GetReason())
	if err != nil {
		return nil, toStatus(err)
	}
	return toSessionResponse(sess), nil
}

func (h *GRPCHandler) ListActiveSession(ctx context.Context, _ *sessionv1.ListActiveSessionRequest) (*sessionv1.ListActiveSessionResponse, error) {
	sessions, err := h.uc.ListActiveSession(ctx)
	if err != nil {
		return nil, toStatus(err)
	}
	resp := &sessionv1.ListActiveSessionResponse{Sessions: make([]*sessionv1.StudySessionResponse, 0, len(sessions))}
	for i := range sessions {
		resp.Sessions = append(resp.Sessions, toSessionResponse(&sessions[i]))
	}
	return resp, nil
}

func (h *GRPCHandler) GetParticipants(ctx context.Context, req *sessionv1.GetParticipantsRequest) (*sessionv1.GetParticipantsResponse, error) {
	participants, err := h.uc.GetParticipants(ctx, req.GetSessionId())
	if err != nil {
		return nil, toStatus(err)
	}
	resp := &sessionv1.GetParticipantsResponse{Participants: make([]*sessionv1.ParticipantInfo, 0, len(participants))}
	for i := range participants {
		resp.Participants = append(resp.Participants, toParticipantInfo(&participants[i]))
	}
	return resp, nil
}

func (h *GRPCHandler) GetSession(ctx context.Context, req *sessionv1.GetSessionRequest) (*sessionv1.StudySessionResponse, error) {
	sess, err := h.uc.GetSession(ctx, req.GetSessionId())
	if err != nil {
		return nil, toStatus(err)
	}
	return toSessionResponse(sess), nil
}

func (h *GRPCHandler) GetMySession(ctx context.Context, req *sessionv1.GetMySessionRequest) (*sessionv1.GetMySessionResponse, error) {
	sess, p, err := h.uc.GetMySession(ctx, req.GetUserId())
	if err != nil {
		return nil, toStatus(err)
	}
	if sess == nil {
		return &sessionv1.GetMySessionResponse{InSession: false}, nil
	}
	return &sessionv1.GetMySessionResponse{
		InSession:   true,
		Session:     toSessionResponse(sess),
		Participant: toParticipantInfo(p),
	}, nil
}

func (h *GRPCHandler) Heartbeat(ctx context.Context, req *sessionv1.HeartbeatRequest) (*sessionv1.HeartbeatResponse, error) {
	active, reason, err := h.uc.Heartbeat(ctx, req.GetSessionId(), req.GetUserId())
	if err != nil {
		return nil, toStatus(err)
	}
	return &sessionv1.HeartbeatResponse{Active: active, Reason: reason}, nil
}

func toSessionResponse(s *domain.StudySession) *sessionv1.StudySessionResponse {
	resp := &sessionv1.StudySessionResponse{
		Id:               s.ID,
		Name:             s.Name,
		CreatorId:        s.CreatorID,
		ParticipantLimit: int32(s.ParticipantLimit),
		ParticipantCount: int32(s.ParticipantCount),
		Status:           string(s.Status),
		CreatedAt:        formatTime(s.CreatedAt),
		EndReason:        s.EndReason,
	}
	if s.EndedAt != nil {
		resp.EndedAt = formatTime(*s.EndedAt)
	}
	return resp
}

func toParticipantInfo(p *domain.Participant) *sessionv1.ParticipantInfo {
	return &sessionv1.ParticipantInfo{
		SessionId:   p.SessionID,
		UserId:      p.UserID,
		JoinedAt:    formatTime(p.JoinedAt),
		DisplayName: p.DisplayName,
		LastSeenAt:  formatTime(p.LastSeenAt),
	}
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

// toStatus maps domain errors onto the gRPC codes documented in
// session.proto, so callers can branch on the code alone.
func toStatus(err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalid):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrNotFound):
		return status.Error(codes.NotFound, "room not found")
	case errors.Is(err, domain.ErrSessionEnded):
		return status.Error(codes.FailedPrecondition, "this room has ended")
	case errors.Is(err, domain.ErrSessionFull):
		return status.Error(codes.ResourceExhausted, "this room is full")
	case errors.Is(err, domain.ErrAlreadyInSession):
		return status.Error(codes.AlreadyExists, "you are already in a room")
	default:
		log.Printf("study-session: internal error: %v", err)
		return status.Error(codes.Internal, "internal error")
	}
}
