package protocol

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrAckTimeout  = errors.New("acknowledgment timed out")
	ErrDuplicateID = errors.New("message ID already tracked")
)

type RetryPolicy struct {
	Timeout    time.Duration
	MaxRetries int
}

type PendingMesssage struct {
	Message  Message
	Deadline time.Time
	Retries  int
	Policy   RetryPolicy
}

type TrackerCallbacks struct {
	Retry  func(Message) error
	Failed func(Message, error)
	Acked  func(Message, Acknowledgement)
}

type AckTracker struct {
	mu        sync.Mutex
	pending   map[ID]PendingMesssage
	callbacks TrackerCallbacks
}

func NewAckTracker(callbacks TrackerCallbacks) *AckTracker {
	return &AckTracker{
		pending:   make(map[ID]PendingMesssage),
		callbacks: callbacks,
	}
}

func (t *AckTracker) Track(
	message Message,
	policy RetryPolicy,
) error {
	if message.ID == "" {
		return errors.New("cannot track message without ID")
	}

	if policy.Timeout <= 0 {
		return errors.New("acknowledgement timeout must be positive")
	}

	if policy.MaxRetries < 0 {
		return errors.New("maximum retries cannot be negative")
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	if _, exists := t.pending[message.ID]; exists {
		return ErrDuplicateID
	}

	t.pending[message.ID] = PendingMesssage{
		Message:  message,
		Deadline: time.Now().Add(policy.Timeout),
		Policy:   policy,
	}
	return nil
}

func (t *AckTracker) Acknowledgement(ack Acknowledgement) bool {
	t.mu.Lock()

	pending, exists := t.pending[ack.ID]
	if exists {
		delete(t.pending, ack.ID)
	}

	t.mu.Unlock()

	if !exists {
		return false
	}

	if t.callbacks.Acked != nil {
		t.callbacks.Acked(pending.Message, ack)
	}

	return true
}

func (t *AckTracker) Cancel(id ID) {
	t.mu.Lock()
	defer t.mu.Unlock()

	delete(t.pending, id)
}

func (t *AckTracker) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()

	clear(t.pending)
}

func (t *AckTracker) Run(ctx context.Context) {
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

func (t *AckTracker) checkTimeouts() {
	now := time.Now()

	var retries []Message
	var failures []Message

	t.mu.Lock()

	for id, pending := range t.pending {
		if now.Before(pending.Deadline) {
			continue
		}

		if pending.Retries >= pending.Policy.MaxRetries {
			failures = append(failures, pending.Message)
			delete(t.pending, id)
			continue
		}

		pending.Retries++
		pending.Deadline = now.Add(pending.Policy.Timeout)

		t.pending[id] = pending
		retries = append(retries, pending.Message)
	}

	t.mu.Unlock()

	// callbacks have to be run outside the mutex (deadlocks)
	for _, message := range retries {
		// an ACK couldve arrived after the timeout scan
		if !t.IsPending(message.ID) {
			continue
		}

		if t.callbacks.Retry != nil {
			if err := t.callbacks.Retry(message); err != nil {
				t.fail(message.ID, err)
			}
		} else {
			t.fail(message.ID, errors.New("retry callback not configured"))
		}
	}

	for _, message := range failures {
		if t.callbacks.Failed != nil {
			t.callbacks.Failed(message, ErrAckTimeout)
		}
	}
}

func (t *AckTracker) IsPending(id ID) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	_, exists := t.pending[id]
	return exists
}

func (t *AckTracker) fail(id ID, err error) {
	t.mu.Lock()

	pending, exists := t.pending[id]
	if exists {
		delete(t.pending, id)
	}

	t.mu.Unlock()

	if exists && t.callbacks.Failed != nil {
		t.callbacks.Failed(pending.Message, err)
	}
}
