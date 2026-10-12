package repository

import (
	"context"
	"crypto/rand"
	"fmt"
	"sync"
	"time"

	"github.com/neennera/fishertimer/services/admin/internal/domain"
)

// InMemoryRepository is a non-durable AuditLogRepository for unit tests.
type InMemoryRepository struct {
	mu   sync.RWMutex
	logs []domain.AdminLog
}

func NewInMemory() *InMemoryRepository {
	return &InMemoryRepository{}
}

func (r *InMemoryRepository) Record(ctx context.Context, entry *domain.AdminLog, apply func(ctx context.Context) error) error {
	if err := apply(ctx); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	entry.LogID = newUUID()
	entry.CreatedAt = time.Now().UTC()
	r.logs = append(r.logs, *entry)
	return nil
}

func (r *InMemoryRepository) List(ctx context.Context, action domain.AdminAction, limit int) ([]domain.AdminLog, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []domain.AdminLog{}
	for i := len(r.logs) - 1; i >= 0 && len(out) < limit; i-- {
		if action == "" || r.logs[i].Action == action {
			out = append(out, r.logs[i])
		}
	}
	return out, nil
}

func newUUID() string {
	var b [16]byte
	rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
