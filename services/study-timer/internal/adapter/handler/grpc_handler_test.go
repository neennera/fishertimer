package handler_test

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	timerv1 "github.com/neennera/fishertimer/proto/studytimer/v1"
	"github.com/neennera/fishertimer/services/study-timer/internal/adapter/handler"
	"github.com/neennera/fishertimer/services/study-timer/internal/adapter/repository"
	"github.com/neennera/fishertimer/services/study-timer/internal/usecase"
)

const (
	demoRoom = "00000000-0000-0000-0000-000000000001"
	demoUser = "00000000-0000-0000-0000-000000000002"
)

// newClient serves the gRPC handler over an in-memory connection.
func newClient(t *testing.T) timerv1.StudyTimerServiceClient {
	t.Helper()
	lis := bufconn.Listen(1 << 20)
	server := grpc.NewServer()
	uc := usecase.New(repository.NewInMemory(), nil)
	timerv1.RegisterStudyTimerServiceServer(server, handler.NewGRPC(uc))
	go func() { _ = server.Serve(lis) }()
	t.Cleanup(server.Stop)

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return lis.DialContext(ctx) }),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return timerv1.NewStudyTimerServiceClient(conn)
}

func TestGRPC_StartThenGet(t *testing.T) {
	client := newClient(t)
	ctx := context.Background()

	if _, err := client.StartTimer(ctx, &timerv1.StartTimerRequest{SessionId: demoRoom, UserId: demoUser}); err != nil {
		t.Fatalf("start: %v", err)
	}
	got, err := client.GetTimer(ctx, &timerv1.GetTimerRequest{SessionId: demoRoom, UserId: demoUser})
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.GetStatus() != timerv1.TimerStatus_TIMER_STATUS_RUNNING || got.GetPhase() != timerv1.TimerPhase_TIMER_PHASE_WORK {
		t.Fatalf("got status %v phase %v, want RUNNING WORK", got.GetStatus(), got.GetPhase())
	}
	if got.GetDurationSeconds() != 25*60 || got.GetRemainingSeconds() <= 0 {
		t.Fatalf("got duration %d remaining %d", got.GetDurationSeconds(), got.GetRemainingSeconds())
	}
}

func TestGRPC_ErrorCodes(t *testing.T) {
	client := newClient(t)
	ctx := context.Background()

	_, err := client.PauseTimer(ctx, &timerv1.PauseTimerRequest{SessionId: demoRoom, UserId: demoUser})
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("pause stopped timer: code = %v, want FailedPrecondition", status.Code(err))
	}

	_, err = client.StartTimer(ctx, &timerv1.StartTimerRequest{UserId: demoUser})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("start without session: code = %v, want InvalidArgument", status.Code(err))
	}
}
