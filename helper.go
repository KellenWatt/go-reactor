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
	f bindingFunc
}

type conRead chan readConState
var conReadLock sync.Mutex
type conWrite chan writeConState
var conWriteLock sync.Mutex
type conBind chan bindConState
var conBindLock sync.Mutex

func runConcurrentRead() {
    for _,c := range conRead {
        c.f(c.value)
    }
}

func runConcurrentWrite() {
    for _,c := range conWrite {
        c.f(c.value)
    }
}

func runConcurrentBind() {
	for _,b := range conBind {
		b.f(b.binder.SetValue(b.source.Value()))
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

func makeConditionalRead(r ReadCallback, f func(interface{})bool) ReadCallbcak {
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
