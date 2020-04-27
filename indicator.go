package reactor

import (
	"sync"
)

// Indicator implements the Binder interface. Indicator provides a mutex, Lock,
// as a convenience for handling shared resources in asynchronous and 
// concurrent callbacks.
type Indicator struct {
	Lock sync.Mutex
	value interface{}

	readCallbacks []Callback
	writeCallbacks []Callback

	bindings []binding // dependent binders
	delayedBindings []binding // supervisor reactors
}

// Value returns the value underlying n and runs any callbacks associated with
// reading. Additionaly, delayed bindings associated with n will be evaluated. 
// A type assertion will probably be needed to meaningfully use the returned 
// value. If n has never been set, Value returns nil and any callbacks will be 
// passed nil.
func (n *Indicator) Value() interface{} {
	n.Lock.Lock()
		v := n.value
	n.Lock.Unlock()
	for _,b := range n.delayedBindings {
		v = b.f(b.source.Value())
	}

	if len(n.delayedBindings) > 0 {
		n.SetValue(v)
	}

	for _,c := range n.readCallbacks {
		c(v)
	}

	return v
}

// SetValue sets the value underlying n and runs any callbacks associated 
// with writing. If the current value is nil (for example, if n has not been 
// set yet), the previous value in callbacks will be nil.
func (n *Indicator) SetValue(v interface{}) {
	n.Lock.Lock()
		prev := n.value
		n.value = v
	n.Lock.Unlock()

	for _,c := range n.writeCallbacks {
		c(prev, v)
	}

	for _,b := range n.bindings {
		val := b.f(v)
		if !b.concurrent {
			b.binder.SetValue(val)
		}
	}
}

// AddBinder adds a Binder to be executed when the value of n changes. If 
// concurrent is true, b.SetValue will not be called when n changes; instead it
// will be queued for change.
//
// AddBinder is largely intended for use in implementing Binders, and its 
// use is heavily discouraged outside of that. If concurrent is true outside 
// of Binder implementations, f may or may not have the desired effect.
// If concurrent is false, this will have exactly the same effect as 
// Binder.AddBinding(), which is the preferred method of creating bindings.
func (n *Indicator) AddBinder(b Binder, f BindingFunc, concurrent bool) {
	n.bindings = append(n.bindings, binding{n, b, f, concurrent})
}

// AddReadCallback adds a callback that will be run when n is read using Value.
func (n *Indicator) AddReadCallback(r Callback) {
    n.readCallbacks = append(n.readCallbacks, r)
}

// AddWriteCallback adds a callback that will be run when n is written to 
// using SetValue.
func (n *Indicator) AddWriteCallback(w Callback) {
    n.writeCallbacks = append(n.writeCallbacks, w)
}

// AddBinding binds n to i, with the value of n being determined by calling f 
// with the value of i.
func (n *Indicator) AddBinding(i Initiator, f BindingFunc) {
	i.AddBinder(n, f, false)
}

// TrivialBinding is a BindingFunc that returns the value passed to it.
// This functions is provided as a convenience
var TrivialBinding = BindingFunc(func(v interface{}) interface{} {return v})

// AddDelayedBinding binds n to i, but the value of n is only determined when 
// Value is called. Because of this, f is only called at the last possible 
// moment. Consequently, this binding behaves differently from the others, 
// and any side effects will be affected as such.
func (n *Indicator) AddDelayedBinding(i Initiator, f BindingFunc) {
	n.delayedBindings = append(n.delayedBindings, binding{i, n, f, false})
}

// AddConcurrentBinding bind n to i, with the value of n being eventually 
// determined by i. The value passed to f when it is eventually called is
// the value of i immediately after it triggers the binding, to ensure 
// consistency.
func (n *Indicator) AddConcurrentBinding(i Initiator, f BindingFunc) {
	conBindLock.Lock()
		if conBind == nil {
			conBind = make(chan conBindState, 100)
			go runConcurrentBind()
		}
	conBindLock.Unlock()

	c := func(v interface{}) interface{} {
		conBind <- conBindState{v, n, f}
		return nil
	}
	i.AddBinder(n, c, true)
}

