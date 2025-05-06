package main

import (
	"math/rand"
	"time"
)

type SharedMap struct {
	items map[int]int
	mu    MapLocker
}

const maxReadLatencyMS = 10
const maxWriteLatencyMS = 50

func sleepRandom(maxMS int) {
	time.Sleep(time.Duration(rand.Intn(maxMS)) * time.Millisecond)
}

func (sm *SharedMap) Get(key int) (int, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	sleepRandom(maxReadLatencyMS)
	val, exists := sm.items[key]
	return val, exists
}

func (sm *SharedMap) Set(key, value int) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sleepRandom(maxWriteLatencyMS)
	sm.items[key] = value
}

func NewSharedMap(locker MapLocker) *SharedMap {
	return &SharedMap{
		items: make(map[int]int),
		mu:    locker,
	}
}
