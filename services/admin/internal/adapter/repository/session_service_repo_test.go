package repository_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/neennera/fishertimer/services/admin/internal/adapter/repository"
	"github.com/neennera/fishertimer/services/admin/internal/domain"
)

func TestListActiveSessions_MapsRooms(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/study-session/active" {
			t.Errorf("unexpected call to %s", r.URL.Path)
		}
		w.Write([]byte(`[
			{"id":"new","name":"New Room","creator_id":"u2","participant_limit":5,"participant_count":3,"status":"ACTIVE","created_at":"2026-10-06T09:00:00Z"}
		]`))
	}))
	defer srv.Close()

	got, err := repository.NewSessionService(srv.URL, nil).ListActiveSessions(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 session, got %d", len(got))
	}
	s := got[0]
	if s.SessionID != "new" || s.Title != "New Room" || s.HostID != "u2" ||
		s.MaxParticipants != 5 || s.ParticipantCount != 3 || !s.IsActive || s.CreatedAt.IsZero() {
		t.Fatalf("unexpected session: %+v", s)
	}
}

func TestGetActiveSession_UnknownRoomIsNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{"id":"room","status":"ACTIVE"}]`))
	}))
	defer srv.Close()

	_, err := repository.NewSessionService(srv.URL, nil).GetActiveSession(context.Background(), "ghost")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestListParticipants_MapsDisplayName(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("session_id") != "room" {
			t.Errorf("unexpected session_id %q", r.URL.Query().Get("session_id"))
		}
		w.Write([]byte(`[{"user_id":"u1","display_name":"Fisher","joined_at":"2026-10-06T09:00:00Z"},{"user_id":"u2","display_name":""}]`))
	}))
	defer srv.Close()

	got, err := repository.NewSessionService(srv.URL, nil).ListParticipants(context.Background(), "room")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 || got[0].DisplayName == nil || *got[0].DisplayName != "Fisher" ||
		got[1].DisplayName != nil || got[0].SessionID != "room" {
		t.Fatalf("unexpected participants: %+v", got)
	}
}

func TestListActiveSessions_UpstreamFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	_, err := repository.NewSessionService(srv.URL, nil).ListActiveSessions(context.Background())
	if !errors.Is(err, domain.ErrUpstream) {
		t.Fatalf("expected ErrUpstream, got %v", err)
	}
}

// The 400 text reaches the admin's screen, so it must not carry the internal
// request path.
func TestListParticipants_RejectedIDMessageHasNoInternalPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()

	_, err := repository.NewSessionService(srv.URL, nil).ListParticipants(context.Background(), "not-a-uuid")
	if !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
	if strings.Contains(err.Error(), "/api/") || strings.Contains(err.Error(), "GET") {
		t.Fatalf("internal path leaked into the message: %q", err.Error())
	}
}
