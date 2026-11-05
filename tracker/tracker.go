package tracker

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/CiaranW005/go-protocol/internal/core"
)

var (
	ErrAckTimeout  = errors.New("acknowledgment timed out")
	ErrDuplicateID = errors.New("message ID already tracked")
)

// Tracker stores pending entries shared by the acknowledgement and response
// trackers. Each supplies its own timeout-checking function.
type Tracker[T any] struct {
	mu            sync.Mutex
	pending       map[core.ID]T
	checkTimeouts func()
}

func NewTracker[T any]() *Tracker[T] {
	return &Tracker[T]{
		pending: make(map[core.ID]T),
	}
}

// Run checks deadlines every 100 milliseconds until the context is canceled.
// Stopping the loop does not clear pending entries or notify their callers.
func (t *Tracker[T]) Run(ctx context.Context) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			t.checkTimeouts()

		case <-ctx.Done():
			return
		}
	}
}

func (t *Tracker[T]) Track(id core.ID, pending T) error {
	if id == "" {
		return errors.New("cannot track message without ID")
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	if _, exists := t.pending[id]; exists {
		return ErrDuplicateID
	}

	t.pending[id] = pending
	return nil
}

func (t *Tracker[T]) Cancel(id core.ID) {
	t.mu.Lock()
	defer t.mu.Unlock()

	delete(t.pending, id)
}

func (t *Tracker[T]) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()

	clear(t.pending)
}

func (t *Tracker[T]) IsPending(id core.ID) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	_, exists := t.pending[id]
	return exists
}
