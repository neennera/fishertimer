package usecase

import (
	"context"
	"log"
	"time"

	"github.com/neennera/fishertimer/services/study-session/internal/domain"
)

// OutboxRelay publishes queued outbox events to the broker, oldest first.
//
// Room changes write their events into the outbox in the same transaction
// (see domain.Repository), so the relay is the only path to RabbitMQ. If
// the broker is down the events simply wait in session_db and go out once
// it is back: Study Timer never misses a leave, so no timer is orphaned.
// Delivery is at-least-once (a crash between publish and mark re-sends),
// which the Timer consumer absorbs by deduplicating on event_id.
type OutboxRelay struct {
	repo      domain.Repository
	publisher domain.EventPublisher
	interval  time.Duration
	batch     int
	now       func() time.Time
}

func NewOutboxRelay(repo domain.Repository, publisher domain.EventPublisher, interval time.Duration) *OutboxRelay {
	return &OutboxRelay{
		repo:      repo,
		publisher: publisher,
		interval:  interval,
		batch:     50,
		now:       func() time.Time { return time.Now().UTC() },
	}
}

// Run drains the outbox every interval until ctx is cancelled.
func (r *OutboxRelay) Run(ctx context.Context) {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()
	for {
		if _, err := r.Drain(ctx); err != nil && ctx.Err() == nil {
			log.Printf("study-session: outbox relay: %v", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// Drain publishes pending events until the outbox is empty or a publish
// fails. It stops at the first failure so events keep their order: a left
// event never overtakes the joined event before it. Returns how many events
// were published.
func (r *OutboxRelay) Drain(ctx context.Context) (int, error) {
	published := 0
	for {
		pending, err := r.repo.PendingEvents(ctx, r.batch)
		if err != nil {
			return published, err
		}
		if len(pending) == 0 {
			return published, nil
		}
		for _, ev := range pending {
			if err := r.publisher.Publish(ctx, ev.RoutingKey, ev.EventID, ev.Payload); err != nil {
				if markErr := r.repo.MarkEventFailed(ctx, ev.ID, err.Error()); markErr != nil {
					log.Printf("study-session: outbox: record failure of %s: %v", ev.EventID, markErr)
				}
				return published, err
			}
			if err := r.repo.MarkEventPublished(ctx, ev.ID, r.now()); err != nil {
				return published, err
			}
			published++
			log.Printf("study-session: published %s (event_id=%s)", ev.RoutingKey, ev.EventID)
		}
		if len(pending) < r.batch {
			return published, nil
		}
	}
}
