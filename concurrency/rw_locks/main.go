package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Result struct {
	Ops int64
}

func reader(sm *SharedMap, results chan<- Result, done <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	var count int64
	for {
		select {
		case <-done:
			results <- Result{Ops: count}
			return
		default:
			key := rand.Intn(20)
			_, _ = sm.Get(key)
			count++
		}
	}
}

func writer(sm *SharedMap, results chan<- Result, done <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	var count int64
	for {
		select {
		case <-done:
			results <- Result{Ops: count}
			return
		default:
			key := rand.Intn(20)
			value := rand.Intn(100)
			sm.Set(key, value)
			count++
		}
	}
}

func runSimulation(lockType string, locker MapLocker, numReaders, numWriters int, duration time.Duration) (readsPerSec, writesPerSec float64) {
	sm := NewSharedMap(locker)
	var wg sync.WaitGroup
	done := make(chan struct{})
	totalRoutines := numReaders + numWriters
	results := make(chan Result, totalRoutines)

	wg.Add(numReaders)
	for range numReaders {
		go reader(sm, results, done, &wg)
	}

	wg.Add(numWriters)
	for range numWriters {
		go writer(sm, results, done, &wg)
	}

	time.Sleep(duration)
	close(done)
	wg.Wait()
	close(results)

	var finalReadOps, finalWriteOps int64
	allResults := make([]Result, 0, totalRoutines)
	for result := range results {
		allResults = append(allResults, result)
	}

	var totalOps int64
	for _, r := range allResults {
		totalOps += r.Ops
	}

	finalReadOps = int64(float64(totalOps) * (float64(numReaders) / float64(totalRoutines)))
	finalWriteOps = totalOps - finalReadOps

	simSeconds := duration.Seconds()

	readsPerSec = float64(finalReadOps) / simSeconds
	writesPerSec = float64(finalWriteOps) / simSeconds

	fmt.Printf("%s Results:\n", lockType)
	fmt.Printf("  Total Reads: %d\n", finalReadOps)
	fmt.Printf("  Total Writes: %d\n", finalWriteOps)
	fmt.Printf("  Reads/sec: %.2f\n", readsPerSec)
	fmt.Printf("  Writes/sec: %.2f\n", writesPerSec)

	return readsPerSec, writesPerSec
}

func main() {
	numReaders := 1000 // Reduced for clarity, increase for more pronounced differences
	numWriters := 10
	duration := 5 * time.Second // Reduced for faster execution

	fmt.Printf("Running simulation for %v with %d readers and %d writers...\n\n", duration, numReaders, numWriters)

	fmt.Println("Running simulation with wrapped Mutex...")
	mutexReadsPerSec, mutexWritesPerSec := runSimulation("Mutex", &MutexWrapper{}, numReaders, numWriters, duration)

	fmt.Println("\nRunning simulation with RWMutex...")
	rwMutexReadsPerSec, rwMutexWritesPerSec := runSimulation("RWMutex", &sync.RWMutex{}, numReaders, numWriters, duration)

	fmt.Println("\n--- Comparison ---")
	fmt.Printf("Read Performance: RWMutex was %.2fx faster than Mutex\n", rwMutexReadsPerSec/mutexReadsPerSec)
	fmt.Printf("Write Performance: RWMutex was %.2fx faster than Mutex\n", rwMutexWritesPerSec/mutexWritesPerSec) // Note: Writes might be slower with RWMutex due to overhead
}
