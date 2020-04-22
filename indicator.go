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

func (n *Indicator) AddBinder(b Binder, f BindingFunc, concurrent bool) {
	n.bindings = append(n.bindings, binding{n, b, f, concurrent})
}

func (n *Indicator) AddReadCallback(r ReadCallback) {
    n.readCallbacks = append(n.readCallbacks, r)
}

func (n *Indicator) AddAsyncReadCallback(r ReadCallback) {
    n.readCallbacks = append(n.readCallbacks, makeAsyncRead(r))
}

func (n *Indicator) AddConcurrentReadCallback(r ReadCallback) {
    conReadLock.Lock()
        if conRead == nil {
            conRead = make(chan readConState, 100)
			go runConcurrentRead()
        }
    conReadLock.Unlock()
    n.readCallbacks = append(n.readCallbacks, makeConcurrentRead(r))
}

func (n *Indicator) AddConditionalReadCallback(r ReadCallback, f func(interface{})bool) {
    n.readCallbacks = append(n.readCallbacks, makeConditionalRead(r, f))
}

func (n *Indicator) AddWriteCallback(w WriteCallback) {
    n.writeCallbacks = append(n.writeCallbacks, w)
}

func (n *Indicator) AddAsyncWriteCallback(w WriteCallback) {
    n.writeCallbacks = append(n.writeCallbacks, makeAsyncWrite(w))
}

func (n *Indicator) AddConcurrentWriteCallback(w WriteCallback) {
    conWriteLock.Lock()
        if conWrite == nil {
            conWrite = make(chan writeConState, 100)
			go runConcurrentWrite()
        }
    conWriteLock.Unlock()
    n.writeCallbacks = append(n.writeCallbacks, makeConcurrentWrite(w))
}

func (n *Indicator) AddConditionalWriteCallback(w WriteCallback, f func(interface{}, interface{})bool) {
	n.writeCallbacks = append(n.writeCallbacks, makeConditionalWrite(w, f))
}


func (n *Indicator) AddBinding(i Initiator, f BindingFunc) {
	i.AddBinder(n, f, false)
}

func (n *Indicator) AddTrivialBinding(i Initiator) {
	i.AddBinder(n, func(v interface{})interface{}{return v}, false)
}

func (n *Indicator) AddDelayedBinding(i Initiator, f BindingFunc) {
	n.delayedBindings = append(n.delayedBindings, binding{i, n, f, false})
}


func (n *Indicator) AddConcurrentBinding(i Initiator, f BindingFunc) {
	conBindLock.Lock()
		if conBind == nil {
			conBind = make(chan bindConState, 100)
			go runConcurrentBind()
		}
	conBindLock.Unlock()

	c := func(v interface{}) interface{} {
		conBind <- bindConState{v, n, f}
		return nil
	}
	i.AddBinder(n, c, true)
}

