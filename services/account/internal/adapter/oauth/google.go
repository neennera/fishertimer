// Package oauth is the driven adapter that talks to Google's OAuth 2.0 /
// OpenID Connect endpoints. It is the only place that knows Google exists.
package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/neennera/fishertimer/services/account/internal/domain"
)

// Config mirrors the Google values from config.Config so this adapter does not
// import the service configuration package.
type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	AuthURL      string
	TokenURL     string
	UserInfoURL  string
}

type GoogleProvider struct {
	cfg    Config
	client *http.Client
}

func New(cfg Config) *GoogleProvider {
	return &GoogleProvider{cfg: cfg, client: &http.Client{Timeout: 10 * time.Second}}
}

// AuthCodeURL builds the consent-screen URL (step 1). `state` is echoed back by
// Google and checked in the callback, which is what stops CSRF.
func (p *GoogleProvider) AuthCodeURL(state string) string {
	q := url.Values{}
	q.Set("client_id", p.cfg.ClientID)
	q.Set("redirect_uri", p.cfg.RedirectURL)
	q.Set("response_type", "code")
	q.Set("scope", "openid email profile")
	q.Set("state", state)
	q.Set("prompt", "select_account")
	return p.cfg.AuthURL + "?" + q.Encode()
}

// FetchProfile exchanges the authorization code for an access token (step 2,
// server-to-server with the client secret) and reads the user's profile
// (step 3).
func (p *GoogleProvider) FetchProfile(ctx context.Context, code string) (*domain.GoogleProfile, error) {
	accessToken, err := p.exchangeCode(ctx, code)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.cfg.UserInfoURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	res, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("account: google userinfo request failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("account: google userinfo returned %d", res.StatusCode)
	}

	var payload struct {
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("account: cannot decode google userinfo: %w", err)
	}

	return &domain.GoogleProfile{
		Email:         payload.Email,
		EmailVerified: payload.EmailVerified,
		Name:          payload.Name,
		Picture:       payload.Picture,
	}, nil
}

func (p *GoogleProvider) exchangeCode(ctx context.Context, code string) (string, error) {
	form := url.Values{}
	form.Set("code", code)
	form.Set("client_id", p.cfg.ClientID)
	form.Set("client_secret", p.cfg.ClientSecret)
	form.Set("redirect_uri", p.cfg.RedirectURL)
	form.Set("grant_type", "authorization_code")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.cfg.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("account: google token exchange failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("account: google token endpoint returned %d", res.StatusCode)
	}

	var payload struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("account: cannot decode google token response: %w", err)
	}
	if payload.AccessToken == "" {
		return "", domain.ErrUnauthorized
	}
	return payload.AccessToken, nil
}
