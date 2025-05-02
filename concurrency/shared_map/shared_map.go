package main

import (
	"fmt"
	"math/rand"
	"sync"
)

func accessMap(m map[int]int, wg *sync.WaitGroup) {
	defer wg.Done()

	key := rand.Intn(10)

	if val, ok := m[key]; ok {
		m[key] = val * 2
	} else {
		m[key] = 1
	}
	_ = m[key]
}

func main() {
	sharedMap := make(map[int]int)
	var wg sync.WaitGroup
	numGoroutines := 100

	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go accessMap(sharedMap, &wg)
	}

	wg.Wait()

	fmt.Println("Final map state (likely unreachable due to panic):", sharedMap)
	fmt.Println("Program finished (if no panic occurred).")

}
