package usecase_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/neennera/fishertimer/services/account/internal/adapter/repository"
	"github.com/neennera/fishertimer/services/account/internal/adapter/token"
	"github.com/neennera/fishertimer/services/account/internal/domain"
	"github.com/neennera/fishertimer/services/account/internal/usecase"
)

type profileEvent struct {
	userID, displayName string
	at                  time.Time
}

type recordingPublisher struct {
	mu     sync.Mutex
	events []profileEvent
}

func (p *recordingPublisher) PublishProfileUpdated(userID, displayName string, at time.Time) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.events = append(p.events, profileEvent{userID, displayName, at})
}

func newServiceWithEvents(p domain.EventPublisher) usecase.Usecase {
	repo := repository.NewInMemory()
	tokens := token.New("test-secret-test-secret-test-secret", "fishertimer-account", time.Hour)
	return usecase.New(repo, &stubProvider{profile: googleProfile()}, tokens, &stubTimerClient{}, &stubRewardClient{}, usecase.WithEvents(p))
}

func TestUpdateProfile_PublishesProfileUpdated(t *testing.T) {
	pub := &recordingPublisher{}
	svc := newServiceWithEvents(pub)
	session := signUp(t, svc, "Fish Lover")

	renamed, err := svc.UpdateProfile(context.Background(), session.Token, "  New Name  ")
	if err != nil {
		t.Fatalf("UpdateProfile: %v", err)
	}
	if len(pub.events) != 1 {
		t.Fatalf("expected exactly one event, got %+v", pub.events)
	}
	ev := pub.events[0]
	if ev.userID != session.User.UserID || ev.displayName != "New Name" || ev.at.IsZero() {
		t.Fatalf("unexpected event %+v", ev)
	}
	if renamed.User.DisplayName != ev.displayName {
		t.Fatalf("the event must carry the stored (trimmed) name, got %q vs %q", ev.displayName, renamed.User.DisplayName)
	}
}

func TestUpdateProfile_NoEventWhenTheRenameFails(t *testing.T) {
	pub := &recordingPublisher{}
	svc := newServiceWithEvents(pub)
	session := signUp(t, svc, "Fish Lover")

	if _, err := svc.UpdateProfile(context.Background(), session.Token, "   "); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
	if _, err := svc.UpdateProfile(context.Background(), "not-a-token", "Name"); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
	if len(pub.events) != 0 {
		t.Fatalf("a failed rename must not be announced: %+v", pub.events)
	}
}

func TestSignUp_IsNotAProfileUpdate(t *testing.T) {
	pub := &recordingPublisher{}
	svc := newServiceWithEvents(pub)
	signUp(t, svc, "Fish Lover")
	if len(pub.events) != 0 {
		t.Fatalf("creating the account is not a rename: %+v", pub.events)
	}
}

func TestUpdateProfile_WorksWithoutAPublisher(t *testing.T) {
	svc, _ := newService(googleProfile()) // built without usecase.WithEvents
	session := signUp(t, svc, "Fish Lover")
	if _, err := svc.UpdateProfile(context.Background(), session.Token, "Other"); err != nil {
		t.Fatalf("a rename must work with no event publisher: %v", err)
	}
}
