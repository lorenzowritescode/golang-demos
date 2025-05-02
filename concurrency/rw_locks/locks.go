package main

import "sync"

type MapLocker interface {
	sync.Locker
	RLock()
	RUnlock()
}

type MutexWrapper struct {
	sync.Mutex
}

func (mw *MutexWrapper) RLock() {
	mw.Lock()
}

func (mw *MutexWrapper) RUnlock() {
	mw.Unlock()
}