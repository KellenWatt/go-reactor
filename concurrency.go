package reactor

import (
	"sync"
)

type conReadState struct {
	value interface{}
	f ReadCallback
}

type conWriteState struct {
	prev, value interface{}
	f WriteCallback
}

type conBindState struct {
	value interface{}
	bound Binder
	f BindingFunc
}

var conRead chan conReadState
var conReadLock sync.Mutex
var conWrite chan conWriteState
var conWriteLock sync.Mutex
var conBind chan conBindState
var conBindLock sync.Mutex

func runConcurrentRead() {
	for c := range conRead {
		c.f(c.value)
	}
}

func runConcurrentWrite() {
	for c := range conWrite {
		c.f(c.prev, c.value)
	}
}

func runConcurrentBind() {
	for b := range conBind {
		b.bound.SetValue(b.f(b.value))
	}
}

func killRead() {
	conReadLock.Lock()
		if conRead != nil {
			close(conRead)
			conRead = nil
		}
	conReadLock.Unlock()
}

func killWrite() {
	conWriteLock.Lock()
		if conWrite != nil {
			close(conWrite)
			conWrite = nil
		}
	conWriteLock.Unlock()
}

func killBind() {
	conBindLock.Lock()
		if conBind != nil {
			close(conBind)
			conBind = nil
		}
	conBindLock.Unlock()
}

