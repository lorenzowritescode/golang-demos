package main

import (
	"fmt"
	"math/rand"
	"sync"
)

func accessSyncMap(smap *sync.Map, wg *sync.WaitGroup) {
	defer wg.Done()

	key := rand.Intn(10)

	actual, loaded := smap.LoadOrStore(key, 1)
	if loaded {
		if intVal, ok := actual.(int); ok {
			smap.Store(key, intVal*2)
		}
	}
	_, _ = smap.Load(key)
}

func main() {
	var smap sync.Map
	var wg sync.WaitGroup
	numGoroutines := 100

	wg.Add(numGoroutines)

	for range numGoroutines {
		go accessSyncMap(&smap, &wg)
	}

	wg.Wait()

	fmt.Println("Final sync.Map state:")
	smap.Range(func(key, value interface{}) bool {
		fmt.Printf("  %v: %v\n", key, value)
		return true
	})
	fmt.Println("Program finished.")
}
