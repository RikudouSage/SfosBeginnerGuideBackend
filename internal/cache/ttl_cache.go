package cache

import (
	"sync"
	"time"
)

type Store[T any] interface {
	Get(key string) (T, bool)
	Set(key string, value T)
}

type entry[T any] struct {
	value     T
	expiresAt time.Time
}

type TTLCache[T any] struct {
	mu    sync.RWMutex
	items map[string]entry[T]
	ttl   time.Duration
}

func NewTTL[T any](ttl time.Duration) *TTLCache[T] {
	return &TTLCache[T]{
		items: make(map[string]entry[T]),
		ttl:   ttl,
	}
}

func (receiver *TTLCache[T]) Get(key string) (T, bool) {
	receiver.mu.RLock()
	item, ok := receiver.items[key]
	receiver.mu.RUnlock()
	if !ok {
		var zero T
		return zero, false
	}

	if time.Now().After(item.expiresAt) {
		receiver.mu.Lock()
		if latest, stillPresent := receiver.items[key]; stillPresent && time.Now().After(latest.expiresAt) {
			delete(receiver.items, key)
		}
		receiver.mu.Unlock()
		var zero T
		return zero, false
	}

	return item.value, true
}

func (receiver *TTLCache[T]) Set(key string, value T) {
	receiver.mu.Lock()
	receiver.items[key] = entry[T]{value: value, expiresAt: time.Now().Add(receiver.ttl)}
	receiver.mu.Unlock()
}
