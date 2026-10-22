package amqp_test

import (
	"context"
	"errors"
	"testing"

	lbamqp "github.com/neennera/fishertimer/services/leaderboard/internal/adapter/amqp"
	"github.com/neennera/fishertimer/services/leaderboard/internal/domain"
)

type fakeRenamer struct {
	userID, name string
	calls        int
	err          error
}

func (f *fakeRenamer) RenameUser(_ context.Context, userID, displayName string) error {
	f.calls++
	f.userID, f.name = userID, displayName
	return f.err
}

const goodBody = `{"event_id":"e1","occurred_at":"2026-10-20T10:00:00Z","user_id":"user1","display_name":"Bobby"}`

func TestHandle_AppliesAProfileUpdate(t *testing.T) {
	r := &fakeRenamer{}
	if v := lbamqp.Handle(context.Background(), r, "profile.updated", []byte(goodBody)); v != lbamqp.Ack {
		t.Fatalf("expected Ack, got %v", v)
	}
	if r.calls != 1 || r.userID != "user1" || r.name != "Bobby" {
		t.Fatalf("unexpected call: %+v", r)
	}
}

func TestHandle_BadPayloadsAreDeadLettered(t *testing.T) {
	for name, body := range map[string]string{
		"not json": `{nope`,
		"no user":  `{"display_name":"Bobby"}`,
		"no name":  `{"user_id":"user1"}`,
		"empty":    ``,
	} {
		r := &fakeRenamer{}
		if v := lbamqp.Handle(context.Background(), r, "profile.updated", []byte(body)); v != lbamqp.Reject || r.calls != 0 {
			t.Errorf("%s: expected Reject without a rename, got %v (calls %d)", name, v, r.calls)
		}
	}
}

func TestHandle_InvalidIsDeadLetteredButOutagesAreRetried(t *testing.T) {
	if v := lbamqp.Handle(context.Background(), &fakeRenamer{err: domain.ErrInvalid}, "profile.updated", []byte(goodBody)); v != lbamqp.Reject {
		t.Fatalf("expected Reject, got %v", v)
	}
	if v := lbamqp.Handle(context.Background(), &fakeRenamer{err: errors.New("redis down")}, "profile.updated", []byte(goodBody)); v != lbamqp.Requeue {
		t.Fatalf("expected Requeue, got %v", v)
	}
}

func TestHandle_OtherRoutingKeysAreIgnored(t *testing.T) {
	r := &fakeRenamer{}
	if v := lbamqp.Handle(context.Background(), r, "something.else", []byte(goodBody)); v != lbamqp.Ack || r.calls != 0 {
		t.Fatalf("expected Ack without a rename, got %v (calls %d)", v, r.calls)
	}
}
