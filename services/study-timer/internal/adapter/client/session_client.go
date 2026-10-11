package client

import (
	"context"

	sessionv1 "github.com/neennera/fishertimer/proto/studysession/v1"
	"github.com/neennera/fishertimer/services/study-timer/internal/domain"
)

// GRPCSessionClient reads room membership from Study Session
// (StudySessionService.GetParticipants, UC-09 step 2).
type GRPCSessionClient struct {
	client sessionv1.StudySessionServiceClient
}

func NewSessionClient(client sessionv1.StudySessionServiceClient) *GRPCSessionClient {
	return &GRPCSessionClient{client: client}
}

// Room counts the room's active participants and finds the user's display
// name among them.
func (c *GRPCSessionClient) Room(ctx context.Context, sessionID, userID string) (domain.RoomSnapshot, error) {
	resp, err := c.client.GetParticipants(ctx, &sessionv1.GetParticipantsRequest{SessionId: sessionID})
	if err != nil {
		return domain.RoomSnapshot{}, err
	}
	room := domain.RoomSnapshot{ParticipantCount: len(resp.GetParticipants())}
	for _, p := range resp.GetParticipants() {
		if p.GetUserId() == userID {
			room.DisplayName = p.GetDisplayName()
			break
		}
	}
	return room, nil
}
