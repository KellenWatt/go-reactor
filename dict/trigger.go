package dict

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
	Lock sync.Mutex
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

func (t *Trigger) AddReadCallback(r reactor.ReadCallback) {
    t.readCallbacks = append(t.readCallbacks, r)
}

func (t *Trigger) AddWriteCallback(w reactor.WriteCallback) {
    t.writeCallbacks = append(t.writeCallbacks, w)
}

func (t *Trigger) Get(key interface{}) interface{} {
	t.Lock.Lock()
		if t.value == nil {
			t.value = make(map[interface{}]interface{})
		}
		v := t.value[key]
	t.Lock.Unlock()

	for _,c := range t.keyReadCallbacks {
		c(Pair{key,v})
	}

	return v
}

func (t *Trigger) GetCheck(key interface{}) (interface{}, bool) {
	t.Lock.Lock()
		if t.value == nil {
			t.value = make(map[interface{}]interface{})
		}
		v,exists := t.value[key]
	t.Lock.Unlock()

	for _,c := range t.keyReadCallbacks {
		c(Pair{key, v})
	}

	return v, exists
}

// Create map if non existant
func (t *Trigger) Set(key, value interface{}) {
	t.Lock.Lock()
		if t.value == nil {
			t.value = make(map[interface{}]interface{})
		}
		prev := t.value[key]
		t.value[key] = value
	t.Lock.Unlock()

	for _,c := range t.keyWriteCallbacks {
		c(Pair{key, prev}, Pair{key, value})
	}
}

func (t *Trigger) Delete(key interface{}) {
	t.Lock.Lock()
		if t.value == nil {
			t.value = make(map[interface{}]interface{})
		}
		prev := t.value[key]
		delete(t.value, key)
	t.Lock.Unlock()

	for _,c := range t.keyWriteCallbacks {
		c(Pair{key, prev}, Pair{key, nil})
	}
}

func (t *Trigger) Size() int {
	return len(t.value)
}

func (t *Trigger) AddKeyReadCallback(r reactor.ReadCallback) {
	t.keyReadCallbacks = append(t.keyReadCallbacks, r)
}

func (t *Trigger) AddKeyWriteCallback(w reactor.WriteCallback) {
	t.keyWriteCallbacks = append(t.keyWriteCallbacks, w)
}


