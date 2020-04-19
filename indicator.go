package reactor

import (
	"sync"
)



type Indicator struct {
	Lock sync.Mutex
	value interface{}

	readCallbacks []ReadCallback
	writeCallbacks []WriteCallback

	bindings []binding // dependent binders
	delayedBindings []binding // supervisor reactors
}

func (n *Indicator) Value() interface{} {
	var v interface{}
	for _,b := range n.delayedBindings {
		v = b.f(b.source.Value())
	}

	if len(n.delayedBindings) > 0 {
		n.SetValue(v)
	}

	for _.c := range n.readCallbacks {
		c(v)
	}

	return n.value
}

func (n *Indicator) SetValue(v interfacE{}) {
	n.Lock.Lock()
		prev := n.value
		n.value = v
	n.Lock.Unlock()

	for _,c := range n.writeCallbacks {
		c(prev, v)
	}

	for _,b := range n.bindings {
		val := b.f(v)
		if !b.concourrent {
			b.binder.SetValue(val)
		}
	}
}

func (n *Indicator) AddBinder(b Binder, f BindingFunc) {
	n.bindings = append(n.bindings, binding{n, b, f, false})
}

func (n *Indicator) AddReadCallback(r ReadCallback) {
    n.readCallbacks = append(n.readCallbacks, r)
}

func (n *Indicator) AddAxyncReadCallback(r ReadCallback) {
    n.readCallbacks = append(n.readCallbacks, makeAsyncRead(r))
}

func (n *Indicator) AddConcurrentReadCallback(r ReadCallback) {
    n.Lock.Lock()
        if conRead == nil {
            conRead = make(chan readConState, 100)
        }
    n.Lock.Unlock()
    n.readCallbacks = append(n.readCallbacks, makeConcurrentRead(r))
}

func (n *Indicator) AddConditionalReadCallback(r ReadCallback) {
    n.readCallbacks = append(n.readCallbacks, makeConditionalRead(r))
}

func (n *Indicator) AddWriteCallback(w WriteCallback) {
    n.writeCallbacks = append(n.writeCallbacks, w)
}

func (n *Indicator) AddAxyncWriteCallback(w WriteCallback) {
    n.writeCallbacks = append(n.writeCallbacks, makeAsyncWrite(w))
}

func (n *Indicator) AddConcurrentWriteCallback(w WriteCallback) {
    n.Lock.Lock()
        if conWrite == nil {
            conWrite = make(chan writeConState, 100)
        }
    n.Lock.Unlock()
    n.writeCallbacks = append(n.writeCallbacks, makeConcurrentWrite(w))
}


func (n *Indicator) AddBinding(i Initiator, f BindingFunc) {
	i.AddBinder(n, f, false)
}

func (n *Indicator) AddSimpleBinding(i Initiator) {
	i.AddBinder(n, func(v interface{})interface{}{return v}, false)
}

func (n *Indicator) AddDelayedBinding(i Initiator, f BindingFunc) {
	n.delayedBindings = append(n.delayedBindings, binder{i, n, f, false})
}

func (n *Indicator) AddConcurrentBinding(i Initiator, f BindingFunc) {
	n.Lock.Lock()
		if conBind == nil {
			conBind = make(chan binding, 100)
		}
	n.Lock.Unlock()

	c := func(v interface{}) interface{} {
		conBind <- conBindState{v, f}
	}
	i.AddBinder(n, c, true)
}

