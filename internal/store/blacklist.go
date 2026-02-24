package store

import (
	"sync"
	"time"
)

// TokenBlacklist menyimpan token yang sudah di-revoke (logout).
// Token yang ada di blacklist tidak boleh dipakai lagi sampai kadaluarsa.
type TokenBlacklist interface {
	Add(token string, expiresAt time.Time)
	IsBlacklisted(token string) bool
}

type memoryBlacklist struct {
	mu   sync.RWMutex
	data map[string]time.Time
}

// NewMemoryBlacklist membuat blacklist in-memory (single instance).
func NewMemoryBlacklist() TokenBlacklist {
	return &memoryBlacklist{
		data: make(map[string]time.Time),
	}
}

func (b *memoryBlacklist) Add(token string, expiresAt time.Time) {
	if token == "" || time.Now().After(expiresAt) {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.data[token] = expiresAt
}

func (b *memoryBlacklist) IsBlacklisted(token string) bool {
	if token == "" {
		return false
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	expiresAt, ok := b.data[token]
	if !ok {
		return false
	}
	if time.Now().After(expiresAt) {
		delete(b.data, token)
		return false
	}
	return true
}
