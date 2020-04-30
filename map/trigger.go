package map

import (
	"sync"

	"github.com/KellenWatt/reactor"
)

func copyMap(m map[interface{}]interface{}) map[interface{}]interface{} {
	n := make(map[interface{}]interface{})
	for k,v := range m {
		n[k] = v
	}
	return n
}


type Pair struct {
	Key interface{}
	Value interface{}
}

type Trigger struct {
	var Lock sync.Mutex
	value map[interface{}]interface{}

	readCallbacks []reactor.ReadCallback
    writeCallbacks []reactor.WriteCallback

    keyReadCallbacks []reactor.ReadCallback
    keyWriteCallbacks []reactor.WriteCallback

    bindings []reactor.Binding
}

func (t *Trigger) Value() interface{} {
	t.Lock.Lock()
		m := copyMap(t.value)
	t.Lock.Unlock()

	for _,c := range t.readCallbacks {
		c(m)
	}

	return m
}

func (t *Trigger) SetValue(v interface{}) {
	t.Lock.Lock() 
		m := v.(map[interface{}]interface{})
		prev := t.value
		t.value = copyMap(m)
	t.Lock.Unlock()

	for _,c := range t.writeCallback {
		c(prev, m)
	}

	for _,b := range t.bindings {
		val := b.F(v)
        if !b.Concurrent {
            b.Binder.SetValue(val)
        }
	}
}

func (t *Trigger) AddBinder(b reactor.Binder, f reactor.BindingFunc, concurrent bool) {
    s.bindings = append(s.bindings, reactor.Binding{s, b, f, concurrent})
}




