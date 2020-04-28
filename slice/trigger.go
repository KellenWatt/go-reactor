package slice

import (
	"sync"
	"errors"

	"github.com/KellenWatt/reactor"
)

var OutOfBoundsError = errors.New("Index out of bounds")

type Index struct {
	Key int
	Value interface{}
}



type SliceTrigger struct {
	Lock sync.Mutex
	value []interface{}

	readCallbacks []reactor.ReadCallback
	writeCallbacks []reactor.WriteCallback

	indexReadCallbacks []reactor.ReadCallback
	indexWriteCallbacks []reactor.WriteCallback

	bindings []reactor.Binding
}

func (s *SliceTrigger) Value() interface{} {
	s.Lock.Lock()
		v := make([]interface{}, len(s.value))
		copy(v, s.value)
	s.Lock.Unlock()

	for _,c := range s.readCallbacks {
		c(v)
	}

	return v
}

func (s *SliceTrigger) SetValue(v interface{}) {
	s.Lock.Lock()
		val := v.([]interface{})
		prev := make([]interface{}, len(s.value))
		copy(prev, s.value)
		s.value = make([]interface{}, len(val))
		copy(s.value, val)
	s.Lock.Unlock()

	for _,c := range s.writeCallbacks {
		c(prev, val)
	}

	for _,b := range s.bindings {
		val := b.F(v)
		if !b.Concurrent {
			b.Binder.SetValue(val)
		}
	}
}

func (s *SliceTrigger) AddBinder(b reactor.Binder, f reactor.BindingFunc, concurrent bool) {
	s.bindings = append(s.bindings, reactor.Binding{s, b, f, concurrent})
}

// runs from SetAt
// funcAddIndexBinder(b reactor.Binder, f reactor.BindingFunc, concurrent bool)

func (s *SliceTrigger) AddReadCallback(r reactor.ReadCallback) {
	s.readCallbacks = append(s.readCallbacks, r)
}

func (s *SliceTrigger) AddWriteCallback(w reactor.WriteCallback) {
	s.writeCallbacks = append(s.writeCallbacks, w)
}

func SliceReadCallback(f func([]interface{})) reactor.ReadCallback {
	return func(v interface{}) {
		val := v.([]interface{})
		f(val)
	}
}

func SliceWriteCallback(f func([]interface{}, []interface{})) reactor.WriteCallback {
	return func(prev, v interface{}) {
		prevVal := prev.([]interface{})
		val := v.([]interface{})
		f(prevVal, val)
	}
}


func (s *SliceTrigger) At(index int) (interface{}, error) {
	s.Lock.Lock()
		valid := index >= 0 && index < len(s.value)
		var v interface{}
		if valid {
			v = s.value[index]
		}
	s.Lock.Unlock()
	if !valid {
		return nil, OutOfBoundsError
	}

	for _,c := range s.indexReadCallbacks {
		c(Index{index, v})
	}

	return v, nil
}

func (s *SliceTrigger) SetAt(index int, v interface{}) error {
	s.Lock.Lock()
		valid := index >= 0 && index < len(s.value)
		var prev interface{}
		if valid {
			prev = s.value[index]
			s.value[index] = v
		}
	s.Lock.Unlock()
	if !valid {
		return OutOfBoundsError
	}

	for _,c := range s.indexWriteCallbacks {
		c(Index{index, prev}, Index{index, v})
	}

	return nil
}

func (s *SliceTrigger) Append(v interface{}) {
	s.Lock.Lock()
		s.value = append(s.value, v)
	s.Lock.Unlock()
	
	for _,c := range s.indexWriteCallbacks {
		c(Index{-1, nil}, Index{len(s.value)-1, v})
	}
}

func (s *SliceTrigger) Pop() (interface{}, error) {
	s.Lock.Lock()
		valid := len(s.value) > 0
		var v interface{}
		if valid {
			v = s.value[len(s.value)-1]
			s.value = s.value[:len(s.value)-1]
		}
	s.Lock.Unlock()
	if !valid {
		// Probably change this to be more accurate
		return nil, OutOfBoundsError
	}

	for _,c := range s.indexWriteCallbacks {
		c(Index{len(s.value), v}, Index{-1, nil})
	}

	return v, nil
}

func (s *SliceTrigger) Slice(from, to int) ([]interface{}, error) {
	s.Lock.Lock()
		valid := from <= to && from >= 0 && from < len(s.value) && to <= len(s.value)
		v := make([]interface{}, to-from)
		if valid {
			copy(v, s.value[from:to])
		}
	s.Lock.Unlock()
	if !valid {
		return nil, OutOfBoundsError
	}

	for _,c := range s.ReadCallbacks {
		c(v)
	}
}


// does not trigger any kind of ReadCallback
func (s *SliceTrigger) Size() int {
	return len(s.value)
}

func (s *SliceTrigger) AddIndexReadCallback(r reactor.ReadCallback) {
	s.indexReadCallbacks = append(s.indexReadCallbacks, r)
}

func (s *SliceTrigger) AddIndexWriteCallback(w reactor.WriteCallback) {
	s.indexWriteCallbacks = append(s.indexWriteCallbacks, w)
}


// Helper methods for creating index callbacks, encapsulates more complex functions
func IndexReadCallback(f func(int,interface{})) reactor.ReadCallback {
	return func(v interface{}) {
		val := v.(Index)
		f(val.Key, val.Value)
	}
}

func IndexWriteCallback(f func(int,interface{},int,interface{})) reactor.WriteCallback {
	return func(prev, v interface{}) {
		prevVal := prev.(Index)
		val := v.(Index)

		f(prevVal.Key, prevVal.Value, val.Key, prevVal.Key)
	}
}
