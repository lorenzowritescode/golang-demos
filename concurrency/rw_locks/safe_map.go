package main

type SharedMap struct {
	items map[int]int
	mu    MapLocker
}

func (sm *SharedMap) Get(key int) (int, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	val, exists := sm.items[key]
	return val, exists
}

func (sm *SharedMap) Set(key, value int) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.items[key] = value
}

func NewSharedMap(locker MapLocker) *SharedMap {
	return &SharedMap{
		items: make(map[int]int),
		mu:    locker,
	}
}