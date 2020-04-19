package reactor

import (
	"sync"
)

type readConState struct {
    value interface{}
    f ReadCallback
}

type writeConState struct {
    prev, value interface{}
    f WriteCallback
}

type bindConState struct {
	value interface{}
	bound Binder
	f BindingFunc
}

var conRead chan readConState
var conReadLock sync.Mutex
var conWrite chan writeConState
var conWriteLock sync.Mutex
var conBind chan bindConState
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
		close(conRead)
		conRead = nil
	conReadLock.Unlock()
}

func killWrite() {
	conWriteLock.Lock()
		close(conWrite)
		conWrite = nil
	conWriteLock.Unlock()
}

func killBind() {
	conBindLock.Lock()
		close(conBind)
		conBind = nil
	conBindLock.Unlock()
}

func makeAsyncRead(r ReadCallback) ReadCallback {
	return func(v interface{}) {
		go r(v)
	}
}

func makeAsyncWrite(w WriteCallback) WriteCallback {
    return func(p, v interface{}) {
        go w(p,v)
    }
}

func makeConcurrentRead(r ReadCallback) ReadCallback {
    return func(v interface{}) {
        conRead <- readConState{v, r}
    }
}

func makeConcurrentWrite(w WriteCallback) WriteCallback {
    return func(p, v interface{}) {
        conWrite <- writeConState{p,v,w}
    }
}

func makeConditionalRead(r ReadCallback, f func(interface{})bool) ReadCallback {
    return func(v interface{}) {
        if f(v) {
            r(v)
        }
    }
}

func makeConditionalWrite(w WriteCallback, f func(interface{}, interface{})bool) WriteCallback {
    return func(p, v interface{}) {
        if f(p,v) {
            w(p,v)
        }
    }
}
