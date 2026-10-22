package usecase_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/neennera/fishertimer/services/study-session/internal/domain"
)

func TestRenameParticipant_UpdatesTheRoomTheUserIsIn(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	room := f.createRoom(t, alice, 3)
	if _, err := f.uc.JoinSession(ctx, room.ID, bob, "Bob"); err != nil {
		t.Fatalf("JoinSession: %v", err)
	}

	if err := f.uc.RenameParticipant(ctx, bob, "  Bobby  "); err != nil {
		t.Fatalf("RenameParticipant: %v", err)
	}
	parts, err := f.uc.GetParticipants(ctx, room.ID)
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]string{}
	for _, p := range parts {
		names[p.UserID] = p.DisplayName
	}
	if names[bob] != "Bobby" || names[alice] != "Creator" {
		t.Fatalf("only Bob's name may change, trimmed: %+v", names)
	}
}

func TestRenameParticipant_UserNotInARoomIsANoOp(t *testing.T) {
	f := newFixture()
	if err := f.uc.RenameParticipant(context.Background(), carol, "Carol"); err != nil {
		t.Fatalf("a user who is in no room is not an error: %v", err)
	}
}

func TestRenameParticipant_LeavesOldStaysUntouched(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	room := f.createRoom(t, alice, 3)
	if _, err := f.uc.JoinSession(ctx, room.ID, bob, "Bob"); err != nil {
		t.Fatal(err)
	}
	if _, err := f.uc.LeaveSession(ctx, room.ID, bob, ""); err != nil {
		t.Fatal(err)
	}
	if err := f.uc.RenameParticipant(ctx, bob, "Bobby"); err != nil {
		t.Fatal(err)
	}
	last, err := f.repo.GetLastParticipation(ctx, room.ID, bob)
	if err != nil {
		t.Fatal(err)
	}
	if last.DisplayName != "Bob" {
		t.Fatalf("a stay that already ended keeps the name it had, got %q", last.DisplayName)
	}
}

func TestRenameParticipant_RejectsBadInput(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	for name, call := range map[string]func() error{
		"not a uuid": func() error { return f.uc.RenameParticipant(ctx, "bob", "Bob") },
		"empty name": func() error { return f.uc.RenameParticipant(ctx, bob, "   ") },
	} {
		if err := call(); !errors.Is(err, domain.ErrInvalid) {
			t.Errorf("%s: expected ErrInvalid, got %v", name, err)
		}
	}
}

func TestRenameParticipant_KeepsTheNameWithinTheColumn(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	room := f.createRoom(t, alice, 3)
	if err := f.uc.RenameParticipant(ctx, alice, strings.Repeat("x", 300)); err != nil {
		t.Fatal(err)
	}
	parts, _ := f.uc.GetParticipants(ctx, room.ID)
	if got := len([]rune(parts[0].DisplayName)); got != 100 {
		t.Fatalf("expected the name cut to the 100-character column, got %d", got)
	}
}
