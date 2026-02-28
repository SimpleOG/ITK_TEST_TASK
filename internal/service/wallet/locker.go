package wallet

import (
	"sync"

	"github.com/google/uuid"
)

type walletLocker struct {
	mu      sync.Mutex
	wallets map[uuid.UUID]*entry
}

type entry struct {
	mu    sync.Mutex
	count int
}

func newWalletLocker() *walletLocker {
	return &walletLocker{
		wallets: make(map[uuid.UUID]*entry),
	}
}

func (l *walletLocker) Lock(id uuid.UUID) func() {
	l.mu.Lock()
	e, ok := l.wallets[id]
	if !ok {
		e = &entry{}
		l.wallets[id] = e
	}
	e.count++
	l.mu.Unlock()

	e.mu.Lock()

	return func() {
		l.mu.Lock()
		e.count--
		if e.count == 0 {
			delete(l.wallets, id)
		}
		l.mu.Unlock()
		e.mu.Unlock()
	}
}
