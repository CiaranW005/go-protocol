package protocol

import (
	"sync"
	"time"
)

type Deduplicator struct {
	mu       sync.Mutex
	seen     map[ID]time.Time
	lifetime time.Duration
}

func NewDeduplicator(lifetime time.Duration) *Deduplicator {
	return &Deduplicator{
		seen:     make(map[ID]time.Time),
		lifetime: lifetime,
	}
}

func (d *Deduplicator) IsDuplicate(id ID) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()

	if expires, exists := d.seen[id]; exists && now.Before(expires) {
		return true
	}

	d.seen[id] = now.Add(d.lifetime)
	return false
}
