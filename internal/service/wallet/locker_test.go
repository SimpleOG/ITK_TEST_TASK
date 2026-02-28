package wallet

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestWalletLocker_BasicLockUnlock(t *testing.T) {
	locker := newWalletLocker()
	id := uuid.New()

	unlock := locker.Lock(id)
	unlock()

	locker.mu.Lock()
	_, exists := locker.wallets[id]
	locker.mu.Unlock()

	assert.False(t, exists, "запись должна быть удалена после unlock")
}

func TestWalletLocker_DifferentWalletsDontBlock(t *testing.T) {
	locker := newWalletLocker()
	id1 := uuid.New()
	id2 := uuid.New()

	unlock1 := locker.Lock(id1)
	defer unlock1()

	done := make(chan struct{})
	go func() {
		unlock2 := locker.Lock(id2)
		unlock2()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("разные кошельки заблокировали друг друга")
	}
}

func TestWalletLocker_SameWalletIsSequential(t *testing.T) {
	locker := newWalletLocker()
	id := uuid.New()

	var counter int64
	var maxConcurrent int64

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			unlock := locker.Lock(id)
			defer unlock()

			current := atomic.AddInt64(&counter, 1)
			if current > atomic.LoadInt64(&maxConcurrent) {
				atomic.StoreInt64(&maxConcurrent, current)
			}

			time.Sleep(time.Millisecond)

			atomic.AddInt64(&counter, -1)
		}()
	}

	wg.Wait()

	assert.Equal(t, int64(1), maxConcurrent, "одновременно должна работать только 1 горутина на кошелёк")
}

// Проверяем что разные кошельки работают параллельно
func TestWalletLocker_DifferentWalletsAreParallel(t *testing.T) {
	locker := newWalletLocker()

	walletCount := 10
	var wg sync.WaitGroup
	start := time.Now()

	for i := 0; i < walletCount; i++ {
		wg.Add(1)
		id := uuid.New()
		go func(walletID uuid.UUID) {
			defer wg.Done()
			unlock := locker.Lock(walletID)
			defer unlock()
			time.Sleep(50 * time.Millisecond)
		}(id)
	}

	wg.Wait()
	elapsed := time.Since(start)

	assert.Less(t, elapsed, 200*time.Millisecond, "разные кошельки должны работать параллельно")
}

func TestWalletLocker_CorrectCounterUnderConcurrency(t *testing.T) {
	locker := newWalletLocker()
	id := uuid.New()

	balance := 0
	goroutines := 500

	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			unlock := locker.Lock(id)
			defer unlock()
			balance++
		}()
	}

	wg.Wait()

	assert.Equal(t, goroutines, balance, "все операции должны быть учтены без race condition")
}

func TestWalletLocker_NoMemoryLeak(t *testing.T) {
	locker := newWalletLocker()

	for i := 0; i < 1000; i++ {
		id := uuid.New()
		unlock := locker.Lock(id)
		unlock()
	}

	locker.mu.Lock()
	size := len(locker.wallets)
	locker.mu.Unlock()

	assert.Equal(t, 0, size, "map должна быть пустой после всех unlock")
}
