package reactor

import (
	"sync"
)

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
		val := b.f(v)
		if !b.concurrent {
			b.binder.SetValue(val)
		}
	}
}


func (t *Trigger) AddBinder(b Binder, f BindingFunc, concurrent bool) {
	t.bindings = append(t.bindings, binding{t, b, f, concurrent})
}


func (t *Trigger) AddReadCallback(r ReadCallback) {
	t.readCallbacks = append(t.readCallbacks, r)
}

func (t *Trigger) AddAsyncReadCallback(r ReadCallback) {
	t.readCallbacks = append(t.readCallbacks, makeAsyncRead(r))
}

func (t *Trigger) AddConcurrentReadCallback(r ReadCallback) {
	conReadLock.Lock()
		if conRead == nil {
			conRead = make(chan readConState, 100)
			go runConcurrentRead()
		}
	conReadLock.Unlock()
	t.readCallbacks = append(t.readCallbacks, makeConcurrentRead(r))
}

func (t *Trigger) AddConditionalReadCallback(r ReadCallback, f func(interface{})bool) {
	t.readCallbacks = append(t.readCallbacks, makeConditionalRead(r, f))
}

func (t *Trigger) AddWriteCallback(w WriteCallback) {
	t.writeCallbacks = append(t.writeCallbacks, w)
}

func (t *Trigger) AddAsyncWriteCallback(w WriteCallback) {
	t.writeCallbacks = append(t.writeCallbacks, makeAsyncWrite(w))
}

func (t *Trigger) AddConcurrentWriteCallback(w WriteCallback) {
	conWriteLock.Lock()
		if conWrite == nil {
			conWrite = make(chan writeConState, 100)
			go runConcurrentWrite()
		}
	conWriteLock.Unlock()
	t.writeCallbacks = append(t.writeCallbacks, makeConcurrentWrite(w))
}

func (t *Trigger) AddConditionalWriteCallback(w WriteCallback, f func(interface{}, interface{})bool) {
	t.writeCallbacks = append(t.writeCallbacks, makeConditionalWrite(w, f))
}


