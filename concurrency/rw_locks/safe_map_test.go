package main

import (
	"math/rand"
	"sync"
	"testing"
)

// This test gives us the illusion of coverage
// we don't have any guarantees that the goroutines are scheduled
// in a way that would cause a race condition
func TestSharedMap_ConcurrentReadWrite(t *testing.T) {
	sm := NewSharedMap(&MutexWrapper{})
	const key = 42
	const initialValue = 123
	sm.Set(key, initialValue)

	const numReaders = 5
	const numWriters = 5
	const operationsPerGoroutine = 10

	var wg sync.WaitGroup

	// Start reader goroutines
	for range numReaders {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range operationsPerGoroutine {
				_, exists := sm.Get(key)
				if !exists {
					t.Errorf("Key %d should exist", key)
				}
			}
		}()
	}

	// Start writer goroutines
	for i := range numWriters {
		wg.Add(1)
		go func(writerID int) {
			defer wg.Done()
			for range operationsPerGoroutine {
				sm.Set(key, rand.Intn(1000)) // Writers set a random value
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Final check for existence is still useful
	_, exists := sm.Get(key)
	if !exists {
		t.Errorf("Key %d should exist at the end", key)
	}
}

// pausableLocker allows pausing a goroutine after it acquires a lock,
// and before its Lock() method returns. It mimics MutexWrapper's behavior
// where RLock is also an exclusive lock.
// It is intended for use in specific tests to control goroutine interleaving.
type pausableLocker struct {
	mu                           sync.Mutex
	lockAcquired_Signal          chan struct{} // Goroutine signals on this after mu.Lock()
	proceedAfterLockAquired_Wait chan struct{} // Goroutine waits on this before Lock() returns
}

func newPausableLocker() *pausableLocker {
	return &pausableLocker{
		// Buffered channel allows the signaling goroutine to not block if the receiver isn't immediately ready,
		// which can simplify orchestration slightly by preventing deadlocks if signals are slightly misordered initially.
		lockAcquired_Signal:          make(chan struct{}, 1),
		proceedAfterLockAquired_Wait: make(chan struct{}),
	}
}

func (pl *pausableLocker) Lock() {
	pl.mu.Lock()
	pl.lockAcquired_Signal <- struct{}{}
	<-pl.proceedAfterLockAquired_Wait
}

func (pl *pausableLocker) Unlock() {
	pl.mu.Unlock()
}

func (pl *pausableLocker) RLock() {
	// Mimic MutexWrapper: RLock is an exclusive lock
	pl.Lock()
}

func (pl *pausableLocker) RUnlock() {
	// Mimic MutexWrapper
	pl.Unlock()
}

func TestSharedMap_ReadAttemptedDuringWrite_Mutex(t *testing.T) {
	pl := newPausableLocker()
	sm := NewSharedMap(pl) // Assumes NewSharedMap is available in package main

	key := 42
	initialValue := 100
	writeValue := 200

	// Perform an initial Set to put the map in a known state.
	// The main goroutine orchestrates the pausableLocker for this initial Set.
	var initialSetWg sync.WaitGroup
	initialSetWg.Add(1)
	go func() {
		defer initialSetWg.Done()
		sm.Set(key, initialValue)
	}()
	// Wait for initial Set's Lock() to acquire mu and signal
	<-pl.lockAcquired_Signal
	// Allow initial Set's Lock() to return, so Set can proceed
	pl.proceedAfterLockAquired_Wait <- struct{}{}
	initialSetWg.Wait() // Ensure the initial Set operation (including Unlock) is complete

	var testWg sync.WaitGroup
	readValueChan := make(chan interface{}, 1)

	testWg.Add(2) // For writer and reader goroutines

	// Writer Goroutine: Attempts to Set a new value
	go func() {
		defer testWg.Done()
		t.Log("Writer: Attempting sm.Set()...")
		sm.Set(key, writeValue)
		// Set will call pl.Lock(), which acquires mu, signals lockAcquired_Signal,
		// then waits on proceedAfterLockAquired_Wait before its Lock() returns.
		t.Log("Writer: sm.Set() completed.")
	}()

	// Reader Goroutine: Attempts to Get the value
	go func() {
		defer testWg.Done()
		t.Log("Reader: Waiting for Writer to signal lock acquisition...")
		// Wait for the Writer's Set->Lock() to acquire mu and send the signal.
		<-pl.lockAcquired_Signal
		t.Log("Reader: Writer has signaled lock acquisition. Attempting sm.Get()...")
		// Now, Writer holds pl.mu and is paused inside its pl.Lock(), waiting on proceedAfterLockAquired_Wait.
		// Reader's Get->RLock->Lock will attempt pl.mu.Lock() and should block.
		val, exists := sm.Get(key)
		// When Get() unblocks, its own Lock() will have signaled and waited.
		t.Log("Reader: sm.Get() completed.")

		if !exists {
			t.Errorf("Reader: Key %d was expected to exist", key)
		}
		readValueChan <- val
	}()

	// Orchestration by the main test goroutine:

	// 1. Allow Writer's Lock() to return, so sm.Set() can proceed to modify data and then Unlock.
	t.Log("Main: Allowing Writer's Lock() to proceed...")
	pl.proceedAfterLockAquired_Wait <- struct{}{}

	// 2. Writer's Set completes, releasing pl.mu. Reader's Get()->Lock() now acquires pl.mu,
	//    signals on pl.lockAcquired_Signal (from its own Lock()), and waits.
	t.Log("Main: Waiting for Reader to signal lock acquisition...")
	<-pl.lockAcquired_Signal // This signal is from the Reader's Lock()
	t.Log("Main: Reader has signaled lock acquisition. Allowing Reader's Lock() to proceed...")
	pl.proceedAfterLockAquired_Wait <- struct{}{}

	// Wait for reader to finish and send the value
	finalReadValue := <-readValueChan

	if finalReadValue != writeValue {
		t.Errorf("Reader: Expected to read value '%v', but got '%v'", writeValue, finalReadValue)
	}

	testWg.Wait() // Wait for both writer and reader goroutines to finish
	t.Log("TestSharedMap_ReadAttemptedDuringWrite_Mutex finished.")
}
