// Package slice implements the reactor.Initiator interface for the special 
// case of slices. Slice access for read/write operations is 0-indexed in 
// all cases, unless specifically mentioned otherwise.
package slice

import (
	"sync"
	"errors"

	"github.com/KellenWatt/reactor"
)

// OutOfBoundsError is the error most often returned by Trigger.
var OutOfBoundsError = errors.New("Index out of bounds")

// Index represents a key-value pair that is passed to a callback for and event
// on a single item.
type Index struct {
	Key int
	Value interface{}
}

// Trigger implements reactor.Initiator for the special case of slices. In 
// addition to the functionality ensured by the reactor.Initiator interface, 
// Trigger provides methods for accessing individual indices and callbacks 
// for those events. Also, methods for obtaining a raw slice of the Trigger
// and its size are provided for ease of use.
//
// Any slices stored in or retrieved from a Trigger will be copied values. 
// Any modification to the source slice or result slices will have no impact 
// on the internal value of the Trigger. If the values are pointers, the 
// the data being pointed to can be changed without trigger events, but that 
// is beyond the scope of this package.
type Trigger struct {
	Lock sync.Mutex
	value []interface{}

	readCallbacks []reactor.ReadCallback
	writeCallbacks []reactor.WriteCallback

	indexReadCallbacks []reactor.ReadCallback
	indexWriteCallbacks []reactor.WriteCallback

	bindings []reactor.Binding
}

// Value returns a copy of the full slice underlying s.
//
// Value calls any Callbacks registerd with AddReadCallback, passing a copy of
// the slice underlying s.
func (s *Trigger) Value() interface{} {
	s.Lock.Lock()
		v := make([]interface{}, len(s.value))
		copy(v, s.value)
	s.Lock.Unlock()

	for _,c := range s.readCallbacks {
		c(v)
	}

	return v
}

// SetValue sets the underlying slice of s to a copy of v. v is tested to 
// ensure that it is a slice of type []interface{}. If the value is not the
// proper type, SetValue panics.
//
// SetValue calls any callbacks registered with AddWriteCallback, passing
// a copy of the previous slice and v (as an []interface{}).
func (s *Trigger) SetValue(v interface{}) {
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

// AddBinder adds a Binder to be executed when the value of s changes. If 
// concurrent is true, b.SetValue will not be called when s changes; instead it
// will be queued for change.
//
// Note that Bindings are not be executed for index-level changes. As such, any
// Binder that is relying on such a change would best be served by using a 
// delayed binding.
//
// AddBinder is largely intended for use in implementing Binders, and its 
// use is heavily discouraged outside of that. If concurrent is true outside 
// of Binder implementations, f may or may not have the desired effect.
// If concurrent is false, this will have exactly the same effect as 
// Binder.AddBinding(), which is the preferred method of creating bindings.
func (s *Trigger) AddBinder(b reactor.Binder, f reactor.BindingFunc, concurrent bool) {
	s.bindings = append(s.bindings, reactor.Binding{s, b, f, concurrent})
}

// runs from SetAt. Currently unsupported 
// funcAddIndexBinder(b reactor.Binder, f reactor.BindingFunc, concurrent bool)

// AddReadCallback adds a callback that will be run when s is read using Value
func (s *Trigger) AddReadCallback(r reactor.ReadCallback) {
	s.readCallbacks = append(s.readCallbacks, r)
}

// AddReadCallback adds a callback that will be run when s is written to using 
// SetValue
func (s *Trigger) AddWriteCallback(w reactor.WriteCallback) {
	s.writeCallbacks = append(s.writeCallbacks, w)
}

// At returns the value at index, if index is within range of the underlying 
// slice. If the index is negative or greater than or equal to the length of 
// s, At returns an OutOfBoundsError. Otherwise, the error is nil. 
//
// Any callbacks registered by AddIndexReadCallback will be called, passing 
// index and the value at index to the callback in an Index struct.
func (s *Trigger) At(index int) (interface{}, error) {
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

// SetAt sets the value at index to v, if index is with range of the underlying
// slice. If the index is negative or greater than or equal to the length of s
// SetAt returns an OutOfBoundsError. Otherwise the error is nil.
//
// Any callbacks registered by AddIndexWriteCallback will be called, passing
// the index and the previous and new values, respectively, to the callback
// in two separate Index structs. The index of both will be equal.
func (s *Trigger) SetAt(index int, v interface{}) error {
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

// Append adds v to the end of s.
//
// Any callbacks registered by AddIndexWriteCallback will be called, passing
// two Index structs with the following specification. The first struct will 
// be Index{-1, nil}, and the second will be the new index and value. The 
// first struct is specified as such because there was no value at that index 
// previously, and specifying an always-invalid index provides an unambiguous 
// identifier for the event.
func (s *Trigger) Append(v interface{}) {
	s.Lock.Lock()
		s.value = append(s.value, v)
	s.Lock.Unlock()
	
	for _,c := range s.indexWriteCallbacks {
		c(Index{-1, nil}, Index{len(s.value)-1, v})
	}
}

// Pop removes the highest-index value from the end of s and returns that value.
// If s is empty, Pop returns an OutOfBoundsError (as the closest applicable 
// error). Otherwise, the error is nil.
//
// Any callbacks registered by AddIndexWriteCallback will be called, passing
// two Index structs, with the following specification. The first struct will
// be the former index and value, and the second will be Index{-1, nil}. The 
// second struct is specified as such because there is now no value at the 
// previous index, and specifying an always-invalid index provides an 
// unambiguous identifier for the event.
func (s *Trigger) Pop() (interface{}, error) {
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

// Slice returns a slice of s with bounds of [from, to). This slice will be
// a copy of that underlying s, and modifying it will have no impact on s. If
// to is less than from, or from or to are outside the bounds of 
// [0, len(s.Value())), and OutOfBoundsError will be returned. Otherwise, the 
// error will be nil.
// 
// Any callbacks registered by AddReadCallback will be called, passing a 
// copy of the slice bounded by [from, to).
func (s *Trigger) Slice(from, to int) ([]interface{}, error) {
	s.Lock.Lock()
		valid := from <= to && from >= 0 && to <= len(s.value)
		var v []interface{}
		if valid {
			v = make([]interface{}, to-from)
			copy(v, s.value[from:to])
		}

	s.Lock.Unlock()
	if !valid {
		return nil, OutOfBoundsError
	}

	for _,c := range s.readCallbacks {
		c(v)
	}

	return v, nil
}

// Size returns the size of the slice underlying s. No ReadCallbacks will 
// be triggered.
func (s *Trigger) Size() int {
	return len(s.value)
}

// AddIndexReadCallback adds a ReadCallback that will be triggered by any 
// index-level read events. The value passed to the callback will be 
// an Index struct containing the index and value being read.
func (s *Trigger) AddIndexReadCallback(r reactor.ReadCallback) {
	s.indexReadCallbacks = append(s.indexReadCallbacks, r)
}

// AddIndexWriteCallback adds a WriteCallback that will be triggered by any 
// index-level write events. The values passed to the callback will be twp 
// Index structs containing the index and previous and new values written, 
// respectively.
func (s *Trigger) AddIndexWriteCallback(w reactor.WriteCallback) {
	s.indexWriteCallbacks = append(s.indexWriteCallbacks, w)
}

