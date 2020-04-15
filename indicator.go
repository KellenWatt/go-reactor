package reactor

import (
	"sync"
)

type conBind struct {
	data Comparable
	bind binding
}

var bindQueue chan conBind

func runConcurrentBindings() {
	for b := range bindQueue {
		b.bind.bound.SetValue(b.bind.bindingFunc(b.data))
	}
}

// SetValue will be called once, whenever a binding is resolved. The delayed 
// binding is an exception to this. For more specific details, see AddDelayedBinding.
//
// Warning in advance, any bindings should not create a cycle, otherwise, there's a very real possibility of 
// causing a stack overflow by infinitely recursive updates. 
type Indicator struct {
	Lock sync.Mutex
	value Comparable

	readCallbacks []ReadCallback
	writeCallbacks []WriteCallback

	watcherLock sync.Mutex
	watchers []binding

	bindingLock sync.Mutex
	delayedBindings []binding
}

// Value returns the internal value of n. If n has any read callbacks 
// registered, they will be resolved here. 
//
// If n has one or more delayed bindings attached to it, the new value is 
// calculated and assigned now, using SetValue. Of note is that SetValue is 
// only called once, for the final result after all bindings are processed.
//
// Of note is that Value is called on the ReadWriteReactor that is bound to,
// as the most recent value of that variable is required to update properly.
func (n *Indicator) Value() Comparable {
	var v Comparable
	for _,b := range n.delayedBindings {
		v = b.bindingFunc(b.bound.Value())
	}

	if len(n.delayedBindings) > 0 {
		n.SetValue(v)
	}
	
	for _,c := range n.readCallbacks {
		c(v)
	}

	return n.value
}

// SetValue sets the internal value of n to v. If n is bound to by any other 
// Binders, they will be processed here, and any write callbacks will be 
// resolved.
func (n *Indicator) SetValue(v Comparable) {
	n.Lock.Lock()
		prev := n.value
		n.value = v
	n.Lock.Unlock()

	for _,c := range n.writeCallbacks {
		c(prev, v)
	}
	
	for _,w := range n.watchers {
		val := w.bindingFunc(v)
		if !w.concurrent {
			w.bound.SetValue(val)
		}
	}
}

// AddReadCallback registers a new read callback to n.
func (n *Indicator) AddReadCallback(c ReadCallback) {
    n.readCallbacks = append(n.readCallbacks, c)
}

// AddAsyncReadCallback registers a new asynchronous read callback to n.
func (n *Indicator) AddAsyncReadCallback(c ReadCallback) {
    n.readCallbacks = append(n.readCallbacks, makeAsyncRead(c))
}

// AddConcurrentReadCallback registers a new concurrent read callback to n.
func (n *Indicator) AddConcurrentReadCallback(c ReadCallback) {
	n.Lock.Lock()
		n.readCallbacks = append(n.readCallbacks, makeConcurrentRead(c))
	n.Lock.Unlock()
}

// AddConditionalReadCallback registers a new conditional read callback to n.
// This callback is only run if calling f on the current value returns true.
func (n *Indicator) AddConditionalReadCallback(c ReadCallback, f func(Comparable) bool) {
    n.readCallbacks = append(n.readCallbacks, makeConditionalRead(c, f))
}

// AddWriteCallback registers a new write callback to n.
func (n *Indicator) AddWriteCallback(c WriteCallback) {
    n.writeCallbacks = append(n.writeCallbacks, c)
}

// AddAsyncWriteCallback registers a new asynchronous write callback to n.
func (n *Indicator) AddAsyncWriteCallback(c WriteCallback) {
    n.writeCallbacks = append(n.writeCallbacks, makeAsyncWrite(c))
}

// AddConcurrentWriteCallback registers a new concurrent write callback to n.
func (n *Indicator) AddConcurrentWriteCallback(c WriteCallback) {
	n.Lock.Lock()
		n.writeCallbacks = append(n.writeCallbacks, makeConcurrentWrite(c))
	n.Lock.Unlock()
}

// AddConditionalReadCallback registers a new conditional write callback to n.
// This callback is only run if calling f on the previous and current values 
// returns true.
func (n *Indicator) AddConditionalWriteCallback(c WriteCallback, f func(Comparable, Comparable) bool) {
    n.writeCallbacks = append(n.writeCallbacks, makeConditionalWrite(c, f))
}

func (n *Indicator) addWatcher(b binding) {
    n.watcherLock.Lock()
		n.watchers = append(n.watchers, b)
    n.watcherLock.Unlock()
}

// AddBinding binds n to the state of r. The value of n is determined by the 
// result of calling f on the value of r.
func (n *Indicator) AddBinding(r ReadWriteReactor, f BindingFunc) {
	b := binding{n, f, false}
	r.addWatcher(b)
}

// AddSimpleBinding binds n to the state of r. The value of n will be changed 
// to the value of r whenever r changes
func (n *Indicator) AddSimpleBinding(r ReadWriteReactor, f BindingFunc) {
	b := binding{n, func(c Comparable)Comparable {return c}, false}
	r.addWatcher(b)
}

// AddDelayedBinding binds n to the state of r. Due to the nature of a delayed 
// binding, only the result of the final binding will be stored, though all 
// bindings will stay registered. 
//
// An example of when this behaviour might be useful is when n is bound to 
// multiple different ReadWriteReactors, and should only set to a different 
// value when one of the ReadWriteReactors has a particular state, and the 
// others are guaranteed to not be in the desired state. More generally, 
// this is useful when two bound variables are disjoint, and you want to
// represent that in some way.
//
// This type of binding can also be useful for expensive or sensitive 
// operations that should be executed only when absolutely necessary.
func (n *Indicator) AddDelayedBinding(r ReadWriteReactor, f BindingFunc) {
	b := binding{r, f, false}
	n.delayedBindings = append(n.delayedBindings, b)
}

// AddConcurrentBinding binds n to the state of r. This type of binding is most
// useful when eventual consistency is a desired system behaviour, or when the 
// value of n doesn't need to be perfectly accurate at any given moment. 
//
// The only promise made by a concurrent binding is that the value of n will 
// eventually be correct, relative to r.
func (n *Indicator) AddConcurrentBinding(r ReadWriteReactor, f BindingFunc) {
	n.Lock.Lock()
		if bindQueue == nil {
			bindQueue = make(chan conBind, 100)
			go runConcurrentBindings()
		}
		b := binding{n, func(v Comparable) Comparable {
			bindQueue <- conBind{v, binding{n, f, false}}
			return n.value
		}, true}
		r.addWatcher(b)
	n.Lock.Unlock()
}

