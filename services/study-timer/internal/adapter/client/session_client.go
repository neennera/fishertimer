package client

import (
	"context"

	sessionv1 "github.com/neennera/fishertimer/proto/studysession/v1"
)

// GRPCSessionClient reads room membership from Study Session
// (StudySessionService.GetParticipants, UC-09 step 2).
type GRPCSessionClient struct {
	client sessionv1.StudySessionServiceClient
}

func NewSessionClient(client sessionv1.StudySessionServiceClient) *GRPCSessionClient {
	return &GRPCSessionClient{client: client}
}

func (c *GRPCSessionClient) ParticipantCount(ctx context.Context, sessionID string) (int, error) {
	resp, err := c.client.GetParticipants(ctx, &sessionv1.GetParticipantsRequest{SessionId: sessionID})
	if err != nil {
		return 0, err
	}
	return len(resp.GetParticipants()), nil
}
