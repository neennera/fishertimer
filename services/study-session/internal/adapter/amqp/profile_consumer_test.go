package amqp_test

import (
	"context"
	"errors"
	"testing"

	sessionamqp "github.com/neennera/fishertimer/services/study-session/internal/adapter/amqp"
	"github.com/neennera/fishertimer/services/study-session/internal/domain"
)

type fakeRenamer struct {
	userID, name string
	calls        int
	err          error
}

func (f *fakeRenamer) RenameParticipant(_ context.Context, userID, displayName string) error {
	f.calls++
	f.userID, f.name = userID, displayName
	return f.err
}

const goodBody = `{"event_id":"e1","occurred_at":"2026-10-20T10:00:00Z","user_id":"aaaaaaaa-0000-4000-8000-000000000001","display_name":"Bobby"}`

func TestHandle_AppliesAProfileUpdate(t *testing.T) {
	r := &fakeRenamer{}
	if v := sessionamqp.Handle(context.Background(), r, "profile.updated", []byte(goodBody)); v != sessionamqp.Ack {
		t.Fatalf("expected Ack, got %v", v)
	}
	if r.calls != 1 || r.name != "Bobby" || r.userID != "aaaaaaaa-0000-4000-8000-000000000001" {
		t.Fatalf("unexpected call: %+v", r)
	}
}

func TestHandle_BadPayloadsAreDeadLettered(t *testing.T) {
	for name, body := range map[string]string{
		"not json":   `{nope`,
		"no user":    `{"display_name":"Bobby"}`,
		"no name":    `{"user_id":"aaaaaaaa-0000-4000-8000-000000000001"}`,
		"empty body": ``,
	} {
		r := &fakeRenamer{}
		if v := sessionamqp.Handle(context.Background(), r, "profile.updated", []byte(body)); v != sessionamqp.Reject || r.calls != 0 {
			t.Errorf("%s: expected Reject without a rename, got %v (calls %d)", name, v, r.calls)
		}
	}
}

func TestHandle_InvalidIsDeadLetteredButOutagesAreRetried(t *testing.T) {
	invalid := &fakeRenamer{err: domain.ErrInvalid}
	if v := sessionamqp.Handle(context.Background(), invalid, "profile.updated", []byte(goodBody)); v != sessionamqp.Reject {
		t.Fatalf("an invalid id can never succeed: expected Reject, got %v", v)
	}
	down := &fakeRenamer{err: errors.New("session_db is down")}
	if v := sessionamqp.Handle(context.Background(), down, "profile.updated", []byte(goodBody)); v != sessionamqp.Requeue {
		t.Fatalf("a transient failure must be retried: expected Requeue, got %v", v)
	}
}

func TestHandle_OtherRoutingKeysAreIgnored(t *testing.T) {
	r := &fakeRenamer{}
	if v := sessionamqp.Handle(context.Background(), r, "something.else", []byte(goodBody)); v != sessionamqp.Ack || r.calls != 0 {
		t.Fatalf("expected Ack without a rename, got %v (calls %d)", v, r.calls)
	}
}
