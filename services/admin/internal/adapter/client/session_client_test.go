package client_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/neennera/fishertimer/services/admin/internal/adapter/client"
	"github.com/neennera/fishertimer/services/admin/internal/domain"
)

func TestKickParticipant_SendsKickedReason(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if r.URL.Path != "/api/v1/study-session/leave" || body["reason"] != "KICKED" ||
			body["session_id"] != "s1" || body["user_id"] != "u1" {
			t.Errorf("unexpected request %s %v", r.URL.Path, body)
		}
		w.Write([]byte(`{"status":"left","session_ended":true}`))
	}))
	defer srv.Close()

	ended, err := client.NewSessionClient(srv.URL, nil).KickParticipant(context.Background(), "s1", "u1")
	if err != nil || !ended {
		t.Fatalf("expected session_ended, got %v %v", ended, err)
	}
}

func TestKickParticipant_AlreadyLeftIsStateChanged(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"already_left","session_ended":false}`))
	}))
	defer srv.Close()

	_, err := client.NewSessionClient(srv.URL, nil).KickParticipant(context.Background(), "s1", "u1")
	if !errors.Is(err, domain.ErrStateChanged) {
		t.Fatalf("expected ErrStateChanged, got %v", err)
	}
}

func TestCloseSession_MapsStatuses(t *testing.T) {
	cases := []struct {
		status int
		want   error
	}{
		{http.StatusNotFound, domain.ErrNotFound},
		{http.StatusConflict, domain.ErrStateChanged},
		{http.StatusBadRequest, domain.ErrInvalid},
		{http.StatusInternalServerError, domain.ErrUpstream},
	}
	for _, c := range cases {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var body map[string]string
			json.NewDecoder(r.Body).Decode(&body)
			if body["reason"] != "ADMIN_CLOSED" {
				t.Errorf("expected ADMIN_CLOSED reason, got %v", body)
			}
			w.WriteHeader(c.status)
			w.Write([]byte(`{"error":"nope"}`))
		}))
		err := client.NewSessionClient(srv.URL, nil).CloseSession(context.Background(), "s1")
		srv.Close()
		if !errors.Is(err, c.want) {
			t.Errorf("status %d: expected %v, got %v", c.status, c.want, err)
		}
	}
}

func TestCloseSession_Unreachable(t *testing.T) {
	err := client.NewSessionClient("http://127.0.0.1:1", nil).CloseSession(context.Background(), "s1")
	if !errors.Is(err, domain.ErrUpstream) {
		t.Fatalf("expected ErrUpstream, got %v", err)
	}
}
