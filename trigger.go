package reactor

import (
	"sync"
)

type readConState struct {
	value interface{}
	f ReadCallback
}

type writeConState struct {
	value interface{}
	f WriteCallback
}

type conRead chan readConState
type conWrite chan writeConState

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
		conWrite <- writeConState{p,v w}
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


type Trigger struct {
	Lock sync.Mutex
	value interface{}

	readCallbacks []ReadCallback
	writeCallbacks []WriteCallback

	bindings []binding
}

func (t *Trigger) Value() interface{} {
	t.Lock.Lock()
		v := t.value
	t.Lock.Unlock()
	
	for _,c := range t.readCallbacks {
		c(v)
	}

	return v
}

func (t *Trigger) SetValue(v interface{}) {
	t.Lock.Lock()
		prev := t.value
		t.value = v
	t.Lock.Unlock()

	for _,c := range t.writeCallbacks {
		c(prev, v)
	}

	for _,b := range t.bindings {
		b.binder.SetValue(b.f(v))
	}
}


func (t *Trigger) AddBinder(b Binder, f BindingFunc) {
	t.bindings = append(t.bindings, binding{t, b, f})
}


func (t *Trigger) AddReadCallback(r ReadCallback) {
	t.readCallbacks = append(t.readCallbacks, r)
}

func (t *Trigger) AddAxyncReadCallback(r ReadCallback) {
	t.readCallbacks = append(t.readCallbacks, makeAsyncRead(r))
}

func (t *Trigger) AddConcurrentReadCallback(r ReadCallback) {
	t.Lock.Lock()
		if conRead == nil {
			conRead = make(chan readConState, 100)
		}
		t.readCallbacks = append(t.readCallbacks, makeConcurrentRead(r))
	t.Lock.Unlock()
}

func (t *Trigger) AddConditionalReadCallback(r ReadCallback) {
	t.readCallbacks = append(t.readCallbacks, makeConditionalRead(r))
}

func (t *Trigger) AddWriteCallback(w WriteCallback) {
	t.writeCallbacks = append(t.writeCallbacks, w)
}

func (t *Trigger) AddAxyncWriteCallback(w WriteCallback) {
	t.writeCallbacks = append(t.writeCallbacks, makeAsyncWrite(w))
}

func (t *Trigger) AddConcurrentWriteCallback(w WriteCallback) {
	t.Lock.Lock()
		if conWrite == nil {
			conWrite = make(chan writeConState, 100)
		}
		t.writeCallbacks = append(t.writeCallbacks, makeConcurrentWrite(w))
	t.Lock.Unlock()
}

func (t *Trigger) AddConditionalWriteCallback(w WriteCallback) {
	t.writeCallbacks = append(t.writeCallbacks, makeConditionalWrite(w))
}


