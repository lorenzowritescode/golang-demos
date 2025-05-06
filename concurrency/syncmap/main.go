package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Result struct {
	Ops      int64
	IsReadOp bool // Flag to distinguish read/write results
}

func reader(smap *sync.Map, results chan<- Result, done <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	var count int64
	for {
		select {
		case <-done:
			results <- Result{Ops: count, IsReadOp: true}
			return
		default:
			key := rand.Intn(20)
			_, _ = smap.Load(key)
			count++
		}
	}
}

func writer(smap *sync.Map, results chan<- Result, done <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	var count int64
	for {
		select {
		case <-done:
			results <- Result{Ops: count, IsReadOp: false}
			return
		default:
			key := rand.Intn(20)
			value := rand.Intn(100)
			smap.Store(key, value)
			count++
		}
	}
}

func runSimulation(numReaders, numWriters int, duration time.Duration) (readsPerSec, writesPerSec float64) {
	var smap sync.Map
	var wg sync.WaitGroup
	done := make(chan struct{})
	totalRoutines := numReaders + numWriters
	results := make(chan Result, totalRoutines)

	wg.Add(numReaders)
	for range numReaders {
		go reader(&smap, results, done, &wg)
	}

	wg.Add(numWriters)
	for range numWriters {
		go writer(&smap, results, done, &wg)
	}

	time.Sleep(duration)
	close(done)
	wg.Wait()
	close(results)

	var finalReadOps, finalWriteOps int64
	for result := range results {
		if result.IsReadOp {
			finalReadOps += result.Ops
		} else {
			finalWriteOps += result.Ops
		}
	}

	simSeconds := duration.Seconds()

	readsPerSec = float64(finalReadOps) / simSeconds
	writesPerSec = float64(finalWriteOps) / simSeconds

	fmt.Println("Sync.Map Results:")
	fmt.Printf("  Total Reads: %d\n", finalReadOps)
	fmt.Printf("  Total Writes: %d\n", finalWriteOps)
	fmt.Printf("  Reads/sec: %.2f\n", readsPerSec)
	fmt.Printf("  Writes/sec: %.2f\n", writesPerSec)

	return readsPerSec, writesPerSec
}

func main() {
	numReaders := 1000
	numWriters := 10
	duration := 5 * time.Second

	fmt.Printf("Running sync.Map simulation for %v with %d readers and %d writers...\n\n", duration, numReaders, numWriters)

	runSimulation(numReaders, numWriters, duration)

	// No comparison within this file, just running the sync.Map simulation.
}
