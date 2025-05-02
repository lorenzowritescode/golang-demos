package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func reader(sm *SharedMap, wg *sync.WaitGroup) {
	defer wg.Done()
	key := rand.Intn(20)
	_, _ = sm.Get(key)
}

func writer(sm *SharedMap, wg *sync.WaitGroup) {
	defer wg.Done()
	key := rand.Intn(20)
	value := rand.Intn(100)
	sm.Set(key, value)
}

func runSimulation(lockType string, locker MapLocker, numReaders, numWriters int) time.Duration {
	sm := NewSharedMap(locker)
	var wg sync.WaitGroup
	totalOps := numReaders + numWriters
	wg.Add(totalOps)

	start := time.Now()

	for i := 0; i < numReaders; i++ {
		go reader(sm, &wg)
	}

	for i := 0; i < numWriters; i++ {
		go writer(sm, &wg)
	}

	wg.Wait()
	duration := time.Since(start)
	fmt.Printf("%s took: %v\n", lockType, duration)
	return duration
}

func main() {
	numReaders := 1_000_000
	numWriters := 100_000

	fmt.Println("Running simulation with wrapped Mutex...")
	durationMutex := runSimulation("Mutex", &MutexWrapper{}, numReaders, numWriters)

	fmt.Println("\nRunning simulation with RWMutex...")
	durationRWMutex := runSimulation("RWMutex", &sync.RWMutex{}, numReaders, numWriters)

	if durationRWMutex < durationMutex {
		fmt.Printf("\nRWMutex was %.2fx faster than Mutex\n", float64(durationMutex)/float64(durationRWMutex))
	} else {
		fmt.Printf("\nMutex was faster or equal to RWMutex (difference: %v)\n", durationMutex-durationRWMutex)
	}
}
