package main

import (
	"fmt"
	"math/rand"
	"sync"
)

type SafeMap struct {
	mu sync.Mutex
	m  map[int]int
}

func (sm *SafeMap) accessMap(wg *sync.WaitGroup) {
	defer wg.Done()

	key := rand.Intn(10)

	sm.mu.Lock()
	defer sm.mu.Unlock()

	if val, ok := sm.m[key]; ok {
		sm.m[key] = val * 2
	} else {
		sm.m[key] = 1
	}
	
	_ = sm.m[key]
}

func main() {
	safeMap := SafeMap{m: make(map[int]int)}
	var wg sync.WaitGroup
	numGoroutines := 100

	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go safeMap.accessMap(&wg)
	}

	wg.Wait()
	fmt.Println("Final map state:", safeMap.m)
	fmt.Println("Program finished successfully.")
}
