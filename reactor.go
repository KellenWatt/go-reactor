package reactor

import (
	"sync"
)

type conRead struct {
	data Comparable
	f ReadCallback
}

type conWrite struct {
	prev,now Comparable
	f WriteCallback
}

var readQueue chan conRead
var writeQueue chan conWrite

// These aren't complicated, but neither is any other simple event loop

func runReadCallbacks() {
	for r := range readQueue {
		r.f(r.data)
	}
}

func runWriteCallbacks() {
	for w := range writeQueue {
		w.f(w.prev, w.now)
	}
}

func makeConditionalRead(f ReadCallback, condition func(Comparable) bool) ReadCallback {
	return func (c Comparable) {
		if condition(c) {
			f(c)
		}
	}
}

func makeConditionalWrite(f WriteCallback, condition func(Comparable, Comparable) bool) WriteCallback {
	return func (c,d Comparable) {
		if condition(c,d) {
			f(c,d)
		}
	}
}

func makeAsyncRead(f ReadCallback) ReadCallback {
	return func (c Comparable) {
		go f(c)
	}
}

func makeAsyncWrite(f WriteCallback) WriteCallback {
	return func (c,d Comparable) {
		go f(c,d)
	}
}

func makeConcurrentRead(f ReadCallback) ReadCallback {
	// kicks off the read event loop, if it hasn't already
	if readQueue == nil {
		readQueue = make(chan conRead, 100)
		go runReadCallbacks()
	}
	return func (c Comparable) {
		readQueue <- conRead{c, f}
	}
}

func makeConcurrentWrite(f WriteCallback) WriteCallback {
	// kicks off the write event loop, if it hasn't already
	if writeQueue == nil {
		writeQueue = make(chan conWrite, 100)
		go runWriteCallbacks()
	}
	return func (c,d Comparable) {
		writeQueue <- conWrite{c, d, f}
	}
}

// Reactor is a structure that allows for the addition of callbacks that are 
// run, based on the usage and changes to it. 
//
// There are four kinds of callbacks provided for both read and write 
// operations: normal, conditional, asynchronous, and concurrent. See below 
// for specific meanings.
//
// A normal callback is exactly what one would expect of a callback, 
// in that it is run as soon as the instigating event occurs. 
// 
// A conditional callback is nearly similar to a normal callback, except it is 
// only run when a specified condition is met. 
//
// An asynchronous callback is run in a goroutine, and makes no promises about 
// when it is run. 
// 
// A concurrent callback is one that is added to a queue of waiting callbacks 
// and will be run eventually, but it makes no other promise than that. Of the 
// four kinds of callbacks provided, concurrent callbacks are the most 
// similar to traditional event-based designs in that there is a sort of loop 
// running in the background, executing callbacks as they are triggered.
//
// Because Reactor has a potential for asynchronous use, a mutex is included 
// with each instance, which is used to ensure data consistency for callbacks.
// Outside of this, Reactor is not guaranteed to be thread-safe for user 
// operations, nor should it be assumed that it is.
type Reactor struct {
	Lock sync.Mutex
	value Comparable

	readCallbacks []ReadCallback
	writeCallbacks []WriteCallback
	
	watcherLock sync.Mutex
	watchers []Binder
}

// Value returns the value of r, then runs or enqueues all callbacks, as
// appropriate. For asynchronous purposes, the value passed to all functions 
// is the value when Value is called.
func (r Reactor) Value() Comparable {
	r.Lock.Lock()
		val := r.value
	r.Lock.Unlock()
	
	for _,c := range r.readCallbacks {
		c(val)
	}
	
	return val
}

// SetValue sets the value of r, then runs or enqueues all callbacks, as 
// appropriate. It also updates any Binders that are associated with r. For 
// asynchronous purposes, the values passed to all functions are the previous 
// value and the new value when SetValue is called. The new value is set 
// before any callbacks run.
func (r *Reactor) SetValue(v Comparable) {
	r.Lock.Lock()
		prev := r.value
		r.value = v
	r.Lock.Unlock()
	
	for _,c := range r.writeCallbacks {
		c(prev, v)
	}
	for _,w := range r.watchers {
		w.SetValue(v)
	}
}

// AddReadCallback registers a new read callback to r.
func (r *Reactor) AddReadCallback(c ReadCallback) {
	r.readCallbacks = append(r.readCallbacks, c)
}

// AddAsyncReadCallback registers a new asynchronous read callback to r.
func (r *Reactor) AddAsyncReadCallback(c ReadCallback) {
	r.readCallbacks = append(r.readCallbacks, makeAsyncRead(c))
}

// AddConcurrentReadCallback registers a new concurrent read callback to r.
func (r *Reactor) AddConcurrentReadCallback(c ReadCallback) {
	r.readCallbacks = append(r.readCallbacks, makeConcurrentRead(c))
}

// AddConditionalReadCallback registers a new conditional read callback to r.
// This callback is only run if calling f on the current value returns true.
func (r *Reactor) AddConditionalReadCallback(c ReadCallback, f func(Comparable) bool) {
	r.readCallbacks = append(r.readCallbacks, makeConditionalRead(c, f))
} 

// AddWriteCallback registers a new write callback to r.
func (r *Reactor) AddWriteCallback(c WriteCallback) {
	r.writeCallbacks = append(r.writeCallbacks, c)
}

// AddAsyncWriteCallback registers a new asynchronous write callback to r.
func (r *Reactor) AddAsyncWriteCallback(c WriteCallback) {
	r.writeCallbacks = append(r.writeCallbacks, makeAsyncWrite(c))
}

// AddConcurrentWriteCallback registers a new concurrent write callback to r.
func (r *Reactor) AddConcurrentWriteCallback(c WriteCallback) {
	r.writeCallbacks = append(r.writeCallbacks, makeConcurrentWrite(c))
}

// AddConditionalReadCallback registers a new conditional write callback to r.
// This callback is only run if calling f on the previous and current values 
// returns true.
func (r *Reactor) AddConditionalWriteCallback(c WriteCallback, f func(Comparable, Comparable) bool) {
	r.writeCallbacks = append(r.writeCallbacks, makeConditionalWrite(c, f))
} 


func (r *Reactor) addWatcher(b Binder) {
	r.watcherLock.Lock()
	r.watchers = append(r.watchers, b)
	r.watcherLock.Unlock()
}


