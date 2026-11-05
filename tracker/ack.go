package tracker

import (
	"errors"
	"time"

	"github.com/CiaranW005/go-protocol/internal/core"
)

// Timeout is the wait for each acknowledgement attempt.
// MaxRetries counts additional sends after the initial send;
// zero means wait for acknowledgement without retrying.
type RetryPolicy struct {
	Timeout    time.Duration
	MaxRetries int
}

type PendingMessage struct {
	Message  core.Message
	Deadline time.Time
	Retries  int
	Policy   RetryPolicy
}

type TrackerCallbacks struct {
	Retry  func(core.Message) error
	Failed func(core.Message, error)
	Acked  func(core.Message, core.Acknowledgement)
}

type AckTracker struct {
	*Tracker[PendingMessage]
	callbacks TrackerCallbacks
}

func NewAckTracker(callbacks TrackerCallbacks) *AckTracker {
	t := &AckTracker{
		Tracker:   NewTracker[PendingMessage](),
		callbacks: callbacks,
	}
	t.Tracker.checkTimeouts = t.checkTimeouts
	return t
}

// Track starts waiting for acknowledgement but does not send the message.
// The caller must perform the initial send.
func (t *AckTracker) Track(message core.Message, policy RetryPolicy) error {
	return t.Tracker.Track(message.ID, PendingMessage{
		Message:  message,
		Deadline: time.Now().Add(policy.Timeout),
		Policy:   policy,
	})
}

func (t *AckTracker) Acknowledgement(ack core.Acknowledgement) bool {
	t.mu.Lock()

	pending, exists := t.pending[ack.ID]
	if exists {
		delete(t.pending, ack.ID)
	}

	t.mu.Unlock()

	if !exists {
		return false
	}

	// Any matching acknowledgement stops retries, including error statuses.
	// The Acked callback decides how to handle the reported status.
	if t.callbacks.Acked != nil {
		t.callbacks.Acked(pending.Message, ack)
	}

	return true
}

func (t *AckTracker) fail(id core.ID, err error) {
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

func (t *AckTracker) checkTimeouts() {
	now := time.Now()

	var retries []core.Message
	var failures []core.Message

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

	// Release the lock before calling external code because a callback may
	// need to access this tracker again.
	for _, message := range retries {
		// Skip this retry if the message was acknowledged or canceled after the scan.
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
