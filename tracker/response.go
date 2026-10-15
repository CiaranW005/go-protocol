package tracker

import (
	"errors"
	"time"

	"github.com/CiaranW005/go-protocol/internal/core"
)

var ErrResponseTimeout = errors.New("response timed out")
var ErrResponseCanceled = errors.New("response canceled")

type ResponseResult struct {
	Message core.Message
	Err     error
}

type PendingResponse struct {
	Message  core.Message
	Deadline time.Time
	Result   chan ResponseResult
}

type ResponseTracker struct {
	*Tracker[PendingResponse]
}

func NewResponseTracker() *ResponseTracker {
	t := &ResponseTracker{
		Tracker: NewTracker[PendingResponse](),
	}
	t.Tracker.checkTimeouts = t.checkTimeouts
	return t
}

func (t *ResponseTracker) Track(message core.Message, timeout time.Duration) (<-chan ResponseResult, error) {
	result := make(chan ResponseResult, 1)

	err := t.Tracker.Track(message.ID, PendingResponse{
		Message:  message,
		Deadline: time.Now().Add(timeout),
		Result:   result,
	})
	if err != nil {
		return nil, err
	}

	return result, err
}

func (t *ResponseTracker) Resolve(message core.Message) bool {
	t.mu.Lock()

	pending, exists := t.pending[message.ReplyTo]
	if exists {
		delete(t.pending, message.ReplyTo)
	}
	t.mu.Unlock()

	if !exists {
		return false
	}

	pending.Result <- ResponseResult{Message: message}
	return true
}

func (t *ResponseTracker) Fail(id core.ID, err error) {
	t.mu.Lock()
	pending, exists := t.pending[id]
	if exists {
		delete(t.pending, id)
	}
	t.mu.Unlock()

	if exists {
		pending.Result <- ResponseResult{Err: err}
	}
}

func (t *ResponseTracker) Cancel(id core.ID) {
	t.Fail(id, ErrResponseCanceled)
}

func (t *ResponseTracker) Clear() {
	t.mu.Lock()

	pending := make([]PendingResponse, 0, len(t.pending))
	for _, entry := range t.pending {
		pending = append(pending, entry)
	}
	clear(t.pending)

	t.mu.Unlock()

	for _, entry := range pending {
		entry.Result <- ResponseResult{Err: ErrResponseCanceled}
	}
}

func (t *ResponseTracker) checkTimeouts() {
	now := time.Now()
	var failures []PendingResponse

	t.mu.Lock()

	for id, pending := range t.pending {
		if !now.Before(pending.Deadline) {
			failures = append(failures, pending)
			delete(t.pending, id)
		}
	}
	t.mu.Unlock()

	for _, pending := range failures {
		pending.Result <- ResponseResult{Err: ErrResponseTimeout}
	}
}
