package amqp_test

import (
	"testing"
	"time"

	accountamqp "github.com/neennera/fishertimer/services/account/internal/adapter/amqp"
)

// With no broker listening, queueing an event must not block or fail the caller:
// a rename is already saved by then.
func TestPublishProfileUpdated_NeverBlocksWithoutABroker(t *testing.T) {
	p := accountamqp.NewPublisher("amqp://nobody:nobody@127.0.0.1:1/")
	defer p.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 1000; i++ { // more than the queue holds: the surplus is dropped, not blocked on
			p.PublishProfileUpdated("11111111-1111-4111-8111-111111111111", "Name", time.Now())
		}
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("PublishProfileUpdated blocked while the broker was unreachable")
	}
}

func TestClose_StopsTheWorkerPromptly(t *testing.T) {
	p := accountamqp.NewPublisher("amqp://nobody:nobody@127.0.0.1:1/")
	p.PublishProfileUpdated("11111111-1111-4111-8111-111111111111", "Name", time.Now())

	closed := make(chan struct{})
	go func() { p.Close(); close(closed) }()
	select {
	case <-closed:
	case <-time.After(3 * time.Second):
		t.Fatal("Close did not return while the worker was retrying")
	}
}
