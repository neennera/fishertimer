import fs from 'fs';
import path from 'path';

const rootDir = process.cwd();

const services = [
  {
    name: 'auth',
    port: 8081,
    entityName: 'UserAccount',
    entityFields: `\tID        string    \`json:"id"\`
\tEmail     string    \`json:"email"\`
\tName      string    \`json:"name"\`
\tCreatedAt time.Time \`json:"created_at"\``,
    repoInterface: `\tFindByEmail(ctx context.Context, email string) (*UserAccount, error)
\tCreate(ctx context.Context, user *UserAccount) error`,
    usecaseInterface: `\tAuthenticate(ctx context.Context, email, name string) (*domain.UserAccount, error)`,
    usecaseImpl: `func (s *service) Authenticate(ctx context.Context, email, name string) (*domain.UserAccount, error) {
\tif email == "" {
\t\treturn nil, domain.ErrInvalid
\t}
\tuser := &domain.UserAccount{ID: "usr_test", Email: email, Name: name}
\treturn user, s.repo.Create(ctx, user)
}`,
    repoMethodsImpl: (typePrefix) => `func (r *${typePrefix}) FindByEmail(ctx context.Context, email string) (*domain.UserAccount, error) {
\treturn nil, domain.ErrNotFound
}

func (r *${typePrefix}) Create(ctx context.Context, user *domain.UserAccount) error {
\treturn nil
}`
  },
  {
    name: 'account',
    port: 8082,
    entityName: 'Profile',
    entityFields: `\tUserID      string    \`json:"user_id"\`
\tDisplayName string    \`json:"display_name"\`
\tIsBanned    bool      \`json:"is_banned"\`
\tTotalFocus  int       \`json:"total_focus_minutes"\`
\tUpdatedAt   time.Time \`json:"updated_at"\``,
    repoInterface: `\tGetProfile(ctx context.Context, userID string) (*Profile, error)
\tUpdateBanStatus(ctx context.Context, userID string, banned bool) error`,
    usecaseInterface: `\tGetProfile(ctx context.Context, userID string) (*domain.Profile, error)
\tSetBanStatus(ctx context.Context, userID string, banned bool) error`,
    usecaseImpl: `func (s *service) GetProfile(ctx context.Context, userID string) (*domain.Profile, error) {
\treturn s.repo.GetProfile(ctx, userID)
}

func (s *service) SetBanStatus(ctx context.Context, userID string, banned bool) error {
\treturn s.repo.UpdateBanStatus(ctx, userID, banned)
}`,
    repoMethodsImpl: (typePrefix) => `func (r *${typePrefix}) GetProfile(ctx context.Context, userID string) (*domain.Profile, error) {
\treturn &domain.Profile{UserID: userID, DisplayName: "Fisher"}, nil
}

func (r *${typePrefix}) UpdateBanStatus(ctx context.Context, userID string, banned bool) error {
\treturn nil
}`
  },
  {
    name: 'study-session',
    port: 8083,
    entityName: 'StudySession',
    entityFields: `\tID               string    \`json:"id"\`
\tName             string    \`json:"name"\`
\tCreatorID        string    \`json:"creator_id"\`
\tParticipantLimit int       \`json:"participant_limit"\`
\tStatus           string    \`json:"status"\`
\tCreatedAt        time.Time \`json:"created_at"\``,
    repoInterface: `\tGetSession(ctx context.Context, id string) (*StudySession, error)
\tCreateSession(ctx context.Context, s *StudySession) error`,
    usecaseInterface: `\tCreateRoom(ctx context.Context, name, creatorID string, limit int) (*domain.StudySession, error)
\tGetRoom(ctx context.Context, id string) (*domain.StudySession, error)`,
    usecaseImpl: `func (s *service) CreateRoom(ctx context.Context, name, creatorID string, limit int) (*domain.StudySession, error) {
\tsess := &domain.StudySession{ID: "sess_test", Name: name, CreatorID: creatorID, ParticipantLimit: limit, Status: "ACTIVE"}
\treturn sess, s.repo.CreateSession(ctx, sess)
}

func (s *service) GetRoom(ctx context.Context, id string) (*domain.StudySession, error) {
\treturn s.repo.GetSession(ctx, id)
}`,
    repoMethodsImpl: (typePrefix) => `func (r *${typePrefix}) GetSession(ctx context.Context, id string) (*domain.StudySession, error) {
\treturn &domain.StudySession{ID: id, Name: "Deep Work Room", ParticipantLimit: 10, Status: "ACTIVE"}, nil
}

func (r *${typePrefix}) CreateSession(ctx context.Context, s *domain.StudySession) error {
\treturn nil
}`
  },
  {
    name: 'study-timer',
    port: 8084,
    entityName: 'TimerState',
    entityFields: `\tSessionID    string    \`json:"session_id"\`
\tUserID       string    \`json:"user_id"\`
\tStatus       string    \`json:"status"\`
\tWorkMinutes  int       \`json:"work_minutes"\`
\tRestMinutes  int       \`json:"rest_minutes"\`
\tCurrentCycle int       \`json:"current_cycle"\`
\tLastUpdated  time.Time \`json:"last_updated"\``,
    repoInterface: `\tGetTimer(ctx context.Context, sessionID, userID string) (*TimerState, error)
\tSaveTimer(ctx context.Context, t *TimerState) error`,
    usecaseInterface: `\tStartTimer(ctx context.Context, sessionID, userID string) (*domain.TimerState, error)
\tStopTimer(ctx context.Context, sessionID, userID string) (*domain.TimerState, error)`,
    usecaseImpl: `func (s *service) StartTimer(ctx context.Context, sessionID, userID string) (*domain.TimerState, error) {
\tt := &domain.TimerState{SessionID: sessionID, UserID: userID, Status: "RUNNING", WorkMinutes: 25, RestMinutes: 5}
\treturn t, s.repo.SaveTimer(ctx, t)
}

func (s *service) StopTimer(ctx context.Context, sessionID, userID string) (*domain.TimerState, error) {
\tt := &domain.TimerState{SessionID: sessionID, UserID: userID, Status: "STOPPED"}
\treturn t, s.repo.SaveTimer(ctx, t)
}`,
    repoMethodsImpl: (typePrefix) => `func (r *${typePrefix}) GetTimer(ctx context.Context, sessionID, userID string) (*domain.TimerState, error) {
\treturn &domain.TimerState{SessionID: sessionID, UserID: userID, Status: "STOPPED"}, nil
}

func (r *${typePrefix}) SaveTimer(ctx context.Context, t *domain.TimerState) error {
\treturn nil
}`
  },
  {
    name: 'reward',
    port: 8085,
    entityName: 'FishReward',
    entityFields: `\tID        string    \`json:"id"\`
\tUserID    string    \`json:"user_id"\`
\tSpecies   string    \`json:"species"\`
\tRarity    string    \`json:"rarity"\`
\tAwardedAt time.Time \`json:"awarded_at"\``,
    repoInterface: `\tAward(ctx context.Context, r *FishReward) error
\tListByUser(ctx context.Context, userID string) ([]FishReward, error)`,
    usecaseInterface: `\tGrantReward(ctx context.Context, userID, species, rarity string) (*domain.FishReward, error)
\tGetUserInventory(ctx context.Context, userID string) ([]domain.FishReward, error)`,
    usecaseImpl: `func (s *service) GrantReward(ctx context.Context, userID, species, rarity string) (*domain.FishReward, error) {
\tr := &domain.FishReward{ID: "fish_test", UserID: userID, Species: species, Rarity: rarity}
\treturn r, s.repo.Award(ctx, r)
}

func (s *service) GetUserInventory(ctx context.Context, userID string) ([]domain.FishReward, error) {
\treturn s.repo.ListByUser(ctx, userID)
}`,
    repoMethodsImpl: (typePrefix) => `func (repo *${typePrefix}) Award(ctx context.Context, r *domain.FishReward) error {
\treturn nil
}

func (repo *${typePrefix}) ListByUser(ctx context.Context, userID string) ([]domain.FishReward, error) {
\treturn []domain.FishReward{{ID: "fish_1", UserID: userID, Species: "Golden Salmon", Rarity: "RARE"}}, nil
}`
  },
  {
    name: 'leaderboard',
    port: 8086,
    entityName: 'RankEntry',
    entityFields: `\tUserID       string    \`json:"user_id"\`
\tDisplayName  string    \`json:"display_name"\`
\tRank         int       \`json:"rank"\`
\tRewardCount  int       \`json:"reward_count"\`
\tFocusMinutes int       \`json:"focus_minutes"\`
\tPeriod       string    \`json:"period"\``,
    repoInterface: `\tGetRankings(ctx context.Context, period string) ([]RankEntry, error)
\tUpsertScore(ctx context.Context, entry *RankEntry) error`,
    usecaseInterface: `\tGetTopUsers(ctx context.Context, period string) ([]domain.RankEntry, error)`,
    usecaseImpl: `func (s *service) GetTopUsers(ctx context.Context, period string) ([]domain.RankEntry, error) {
\treturn s.repo.GetRankings(ctx, period)
}`,
    repoMethodsImpl: (typePrefix) => `func (repo *${typePrefix}) GetRankings(ctx context.Context, period string) ([]domain.RankEntry, error) {
\treturn []domain.RankEntry{{UserID: "u1", DisplayName: "TopAngler", Rank: 1, RewardCount: 50, FocusMinutes: 1200, Period: period}}, nil
}

func (repo *${typePrefix}) UpsertScore(ctx context.Context, entry *domain.RankEntry) error {
\treturn nil
}`
  },
  {
    name: 'admin',
    port: 8087,
    entityName: 'ModerationReport',
    entityFields: `\tID         string    \`json:"id"\`
\tReporterID string    \`json:"reporter_id"\`
\tReportedID string    \`json:"reported_id"\`
\tReason     string    \`json:"reason"\`
\tStatus     string    \`json:"status"\`
\tCreatedAt  time.Time \`json:"created_at"\``,
    repoInterface: `\tCreateReport(ctx context.Context, r *ModerationReport) error
\tListPending(ctx context.Context) ([]ModerationReport, error)`,
    usecaseInterface: `\tSubmitReport(ctx context.Context, reporterID, reportedID, reason string) (*domain.ModerationReport, error)
\tGetPendingReports(ctx context.Context) ([]domain.ModerationReport, error)`,
    usecaseImpl: `func (s *service) SubmitReport(ctx context.Context, reporterID, reportedID, reason string) (*domain.ModerationReport, error) {
\tr := &domain.ModerationReport{ID: "rep_test", ReporterID: reporterID, ReportedID: reportedID, Reason: reason, Status: "PENDING"}
\treturn r, s.repo.CreateReport(ctx, r)
}

func (s *service) GetPendingReports(ctx context.Context) ([]domain.ModerationReport, error) {
\treturn s.repo.ListPending(ctx)
}`,
    repoMethodsImpl: (typePrefix) => `func (repo *${typePrefix}) CreateReport(ctx context.Context, r *domain.ModerationReport) error {
\treturn nil
}

func (repo *${typePrefix}) ListPending(ctx context.Context) ([]domain.ModerationReport, error) {
\treturn []domain.ModerationReport{{ID: "rep_1", Reason: "AFK Spam", Status: "PENDING"}}, nil
}`
  }
];

for (const svc of services) {
  const svcDir = path.join(rootDir, 'services', svc.name);
  const pkgMod = `github.com/neennera/fishertimer/services/${svc.name}`;

  // 1. config/config.go
  const configDir = path.join(svcDir, 'config');
  fs.mkdirSync(configDir, { recursive: true });
  fs.writeFileSync(path.join(configDir, 'config.go'), `package config

import (
\t"os"
\t"strconv"
)

type Config struct {
\tPort int
\tEnv  string
}

func Load() *Config {
\tport := ${svc.port}
\tif p := os.Getenv("PORT"); p != "" {
\t\tif val, err := strconv.Atoi(p); err == nil {
\t\t\tport = val
\t\t}
\t}
\tenv := os.Getenv("ENV")
\tif env == "" {
\t\tenv = "development"
\t}
\treturn &Config{
\t\tPort: port,
\t\tEnv:  env,
\t}
}
`);

  // 2. internal/domain/entity.go
  const domainDir = path.join(svcDir, 'internal', 'domain');
  fs.mkdirSync(domainDir, { recursive: true });
  const hasTime = svc.entityFields.includes('time.Time');
  fs.writeFileSync(path.join(domainDir, 'entity.go'), `package domain

import (
\t"errors"${hasTime ? '\n\t"time"' : ''}
)

var (
\tErrNotFound = errors.New("${svc.name}: resource not found")
\tErrInvalid  = errors.New("${svc.name}: invalid input")
)

type ${svc.entityName} struct {
${svc.entityFields}
}
`);

  // 3. internal/domain/repository.go
  fs.writeFileSync(path.join(domainDir, 'repository.go'), `package domain

import (
\t"context"
)

type Repository interface {
${svc.repoInterface}
}
`);

  // 4. internal/usecase/service.go
  const usecaseDir = path.join(svcDir, 'internal', 'usecase');
  fs.mkdirSync(usecaseDir, { recursive: true });
  fs.writeFileSync(path.join(usecaseDir, 'service.go'), `package usecase

import (
\t"context"
\t"${pkgMod}/internal/domain"
)

type Usecase interface {
${svc.usecaseInterface}
}

type service struct {
\trepo domain.Repository
}

func New(repo domain.Repository) Usecase {
\treturn &service{repo: repo}
}

${svc.usecaseImpl}
`);

  // 5. internal/usecase/service_test.go
  fs.writeFileSync(path.join(usecaseDir, 'service_test.go'), `package usecase_test

import (
\t"context"
\t"testing"
\t"${pkgMod}/internal/domain"
\t"${pkgMod}/internal/usecase"
)

type mockRepository struct{}

${svc.repoMethodsImpl('mockRepository')}

func TestUsecase_Success(t *testing.T) {
\trepo := &mockRepository{}
\tsvc := usecase.New(repo)
\tif svc == nil {
\t\tt.Fatal("expected usecase service to be initialized")
\t}
\tctx := context.Background()
\t_ = ctx
}
`);

  // 6. internal/adapter/repository/memory_repo.go
  const repoDir = path.join(svcDir, 'internal', 'adapter', 'repository');
  fs.mkdirSync(repoDir, { recursive: true });
  fs.writeFileSync(path.join(repoDir, 'memory_repo.go'), `package repository

import (
\t"context"
\t"sync"
\t"${pkgMod}/internal/domain"
)

type InMemoryRepository struct {
\tmu sync.RWMutex
}

func NewInMemory() *InMemoryRepository {
\treturn &InMemoryRepository{}
}

${svc.repoMethodsImpl('InMemoryRepository')}
`);

  // 7. internal/adapter/handler/http_handler.go
  const handlerDir = path.join(svcDir, 'internal', 'adapter', 'handler');
  fs.mkdirSync(handlerDir, { recursive: true });
  fs.writeFileSync(path.join(handlerDir, 'http_handler.go'), `package handler

import (
\t"encoding/json"
\t"net/http"
\t"time"
\t"${pkgMod}/internal/usecase"
)

type HTTPHandler struct {
\tuc usecase.Usecase
}

func New(uc usecase.Usecase) *HTTPHandler {
\treturn &HTTPHandler{uc: uc}
}

func (h *HTTPHandler) RegisterRoutes(mux *http.ServeMux) {
\tmux.HandleFunc("/health", h.Health)
\tmux.HandleFunc("/api/v1/${svc.name}/status", h.Status)
}

func (h *HTTPHandler) Health(w http.ResponseWriter, r *http.Request) {
\tw.Header().Set("Content-Type", "application/json")
\tjson.NewEncoder(w).Encode(map[string]any{
\t\t"service":   "${svc.name}",
\t\t"status":    "healthy",
\t\t"port":      ${svc.port},
\t\t"timestamp": time.Now().UTC().Format(time.RFC3339),
\t})
}

func (h *HTTPHandler) Status(w http.ResponseWriter, r *http.Request) {
\tw.Header().Set("Content-Type", "application/json")
\tjson.NewEncoder(w).Encode(map[string]any{
\t\t"service": "${svc.name}",
\t\t"layer":   "adapter.handler",
\t\t"ready":   true,
\t})
}
`);

  // 8. cmd/main.go composition root
  const mainGoPath = path.join(svcDir, 'cmd', 'main.go');
  fs.writeFileSync(mainGoPath, `package main

import (
\t"context"
\t"fmt"
\t"log"
\t"net/http"
\t"os"
\t"os/signal"
\t"syscall"
\t"time"

\t"${pkgMod}/config"
\t"${pkgMod}/internal/adapter/handler"
\t"${pkgMod}/internal/adapter/repository"
\t"${pkgMod}/internal/usecase"
)

func main() {
\tcfg := config.Load()

\t// 1. Instantiate Driven Adapter (Repository)
\trepo := repository.NewInMemory()

\t// 2. Inject into Application Usecase
\tuc := usecase.New(repo)

\t// 3. Inject into Driving Adapter (HTTP Handler)
\th := handler.New(uc)

\t// 4. Setup Router & Server
\tmux := http.NewServeMux()
\th.RegisterRoutes(mux)

\tserver := &http.Server{
\t\tAddr:         fmt.Sprintf(":%d", cfg.Port),
\t\tHandler:      mux,
\t\tReadTimeout:  5 * time.Second,
\t\tWriteTimeout: 10 * time.Second,
\t}

\tgo func() {
\t\tlog.Printf("${svc.name} service listening on port %d [%s]", cfg.Port, cfg.Env)
\t\tif err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
\t\t\tlog.Fatalf("listen error: %s\\n", err)
\t\t}
\t}()

\tquit := make(chan os.Signal, 1)
\tsignal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
\t<-quit

\tlog.Printf("Shutting down ${svc.name} service...")
\tctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
\tdefer cancel()
\tif err := server.Shutdown(ctx); err != nil {
\t\tlog.Fatalf("forced shutdown: %s\\n", err)
\t}
\tlog.Printf("${svc.name} exited cleanly")
}
`);

  console.log(`Updated Clean Architecture template for service: ${svc.name}`);
}

console.log('All services updated successfully!');
