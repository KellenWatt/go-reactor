package reactor

import (
	"sync"
)

// Trigger implements the Initiator interface. Trigger provides a mutex, Lock, 
// as a convenience for handling shared resources in asynchronous and 
// concurrent callbacks.
type Trigger struct {
	Lock sync.Mutex
	value interface{}

	readCallbacks []ReadCallback
	writeCallbacks []WriteCallback

	bindings []binding
}

// Value returns the value underlying t and runs any callbacks associated with 
// reading. A type assertion will probably be needed to meaningfully use the 
// returned value. If t has never been set, Value returns nil and any callbacks 
// will be passed nil.
func (t *Trigger) Value() interface{} {
	t.Lock.Lock()
		v := t.value
	t.Lock.Unlock()
	
	for _,c := range t.readCallbacks {
		c(v)
	}

	return v
}

// SetValue sets the value underlying t and runs any callbacks associated 
// with writing. If the current value is nil (for example, if t has not been 
// set yet), the previous value in callbacks will be nil.
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

// AddBinder adds a Binder to be executed when the value of t changes. If 
// concurrent is true, b.SetValue will not be called when t changes; instead it
// will be queued for change.
//
// AddBinder is largely intended for use in implementing Binders, and its 
// use is heavily discouraged outside of that. If concurrent is true outside 
// of Binder implementations, f may or may not have the desired effect.
// If concurrent is false, this will have exactly the same effect as 
// Binder.AddBinding(), which is the preferred method of creating bindings.
func (t *Trigger) AddBinder(b Binder, f BindingFunc, concurrent bool) {
	t.bindings = append(t.bindings, binding{t, b, f, concurrent})
}

// AddReadCallback adds a callback that will be run when t is read using Value.
func (t *Trigger) AddReadCallback(r ReadCallback) {
	t.readCallbacks = append(t.readCallbacks, r)
}

// AddAsyncReadCallback adds a callback that will be run in a goroutine when t 
// is read using Value.
func (t *Trigger) AddAsyncReadCallback(r ReadCallback) {
	t.readCallbacks = append(t.readCallbacks, makeAsyncRead(r))
}

// AddConcurrentReadCallback adds a callback that will be run concurrently
// when t is read using Value.
func (t *Trigger) AddConcurrentReadCallback(r ReadCallback) {
	conReadLock.Lock()
		if conRead == nil {
			conRead = make(chan readConState, 100)
			go runConcurrentRead()
		}
	conReadLock.Unlock()
	t.readCallbacks = append(t.readCallbacks, makeConcurrentRead(r))
}

// AddConditionalReadCallback adds a callback that will be run when t is read 
// using Value only if f evaluates to true when passed the value of t.
func (t *Trigger) AddConditionalReadCallback(r ReadCallback, f func(interface{})bool) {
	t.readCallbacks = append(t.readCallbacks, makeConditionalRead(r, f))
}

// AddWriteCallback adds a callback that will be run when t is written to 
// using SetValue.
func (t *Trigger) AddWriteCallback(w WriteCallback) {
	t.writeCallbacks = append(t.writeCallbacks, w)
}

// AddAsyncWriteCallback adds a callback that will be run in a goroutine when t 
// is written to using SetValue.
func (t *Trigger) AddAsyncWriteCallback(w WriteCallback) {
	t.writeCallbacks = append(t.writeCallbacks, makeAsyncWrite(w))
}

// AddConcurentWriteCallback adds a callback that will run concurrently when
// t is written to using SetValue.
func (t *Trigger) AddConcurrentWriteCallback(w WriteCallback) {
	conWriteLock.Lock()
		if conWrite == nil {
			conWrite = make(chan writeConState, 100)
			go runConcurrentWrite()
		}
	conWriteLock.Unlock()
	t.writeCallbacks = append(t.writeCallbacks, makeConcurrentWrite(w))
}

// AddConditionalWriteCallback adds a callback that will be run  when t is 
// written to using SetValue only if f evaluates to true when passed the
// previous and new value of t.
func (t *Trigger) AddConditionalWriteCallback(w WriteCallback, f func(interface{}, interface{})bool) {
	t.writeCallbacks = append(t.writeCallbacks, makeConditionalWrite(w, f))
}


