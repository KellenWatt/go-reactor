package reactor

import (
	"sync"
)

type conEventState struct {
	values []interface{}
	f Callback
}

type conBindState struct {
	value interface{}
	bound Binder
	f BindingFunc
}

var conEvent chan conEventState
var conEventLock sync.Mutex
var conBind chan conBindState
var conBindLock sync.Mutex

func runConcurrentBind() {
	for b := range conBind {
		b.bound.SetValue(b.f(b.value))
	}
}

func runConcurrentEvent() {
	for c := range conEvent {
		c.f(c.values...)
	}
}

func killRead() {
	killEvent()
}

func killWrite() {
	killEvent()
}

func killEvent() {
	conEventLock.Lock()
		if conEvent != nil {
			close(conEvent)
			conEvent = nil
		}
	conEventLock.Unlock()
}

func killBind() {
	conBindLock.Lock()
		if conBind != nil {
			close(conBind)
			conBind = nil
		}
	conBindLock.Unlock()
}

