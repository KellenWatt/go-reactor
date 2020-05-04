// Package dict implements the reactor.Initiator interface for the special
// case of maps. Implementations in this package are guaranteed to, at minimum, 
// work exactly the same as the go standard map, except for iteration, due to 
// restrictions in how range works.
//
// There is no common-sense implementation of reactor.Initiator for structs, 
// so dict.Trigger is the closest approximation of that idea.
package dict

import (
	"sync"

	"github.com/KellenWatt/reactor"
)

// performs a semi-deep copy of the passed map. Used in methods that return 
// the map or provides access to the data (such as callbacks)
func copyMap(m map[interface{}]interface{}) map[interface{}]interface{} {
	n := make(map[interface{}]interface{})
	for k,v := range m {
		n[k] = v
	}
	return n
}

// Pair represents a standard key-value pair composed of interface{} types.
type Pair struct {
	Key interface{}
	Value interface{}
}

// Trigger implements the reactor.Initiator interface for the special case of 
// maps. In addition to the functionality required by reactor.Initiator, 
// Trigger provides methods for accessing individual key-value pairs and 
// callbacks for those events. Methods for obtaining a raw map of the Trigger
// and its size are provided for convenience.
//
// Any maps stored in a retrieved from a Trigger will be copied values. Any
// modification to the source, result, or argument maps will have no effect on the 
// internal value of the Trigger. If the values are pointers, the data being 
// pointed to can be changed without trigger events, but that is beyond the
// scope of this package.
type Trigger struct {
	Lock sync.Mutex
	value map[interface{}]interface{}

	readCallbacks []reactor.ReadCallback
    writeCallbacks []reactor.WriteCallback

    keyReadCallbacks []reactor.ReadCallback
    keyWriteCallbacks []reactor.WriteCallback

    bindings []reactor.Binding
}

// Value returns a copy of the full slice underlying t.
//
// Value calls any Callbacks registerd with AddReadCallback, passing a copy of
// the slice underlying t.
func (t *Trigger) Value() interface{} {
	t.Lock.Lock()
		m := copyMap(t.value)
	t.Lock.Unlock()

	for _,c := range t.readCallbacks {
		c(m)
	}

	return m
}

// SetValue sets the underlying map of t to a copy of v. v is tested to 
// ensure that it is a slice of type map[interface{}]interface{}. If the 
// value is not the proper type, SetValue panics.
//
// SetValue calls any callbacks registered with AddWriteCallback, passing
// a copy of the previous map and v (as an map[interface{}]interface{}).
func (t *Trigger) SetValue(v interface{}) {
	t.Lock.Lock() 
		m := v.(map[interface{}]interface{})
		prev := t.value
		t.value = copyMap(m)
	t.Lock.Unlock()

	for _,c := range t.writeCallbacks {
		c(prev, m)
	}

	for _,b := range t.bindings {
		val := b.F(v)
        if !b.Concurrent {
            b.Binder.SetValue(val)
        }
	}
}

// AddBinder adds a Binder to be executed when the value of t changes. If 
// concurrent is true, b.SetValue will not be called when t changes; instead it
// will be queued for change.
//
// Note that Bindings are not be executed for individual key-value changes. 
// As such, any Binder that is relying on such a change would best be served 
// by using a delayed binding.
//
// AddBinder is largely intended for use in implementing Binders, and its 
// use is heavily discouraged outside of that. If concurrent is true outside 
// of Binder implementations, f may or may not have the desired effect.
// If concurrent is false, this will have exactly the same effect as 
// Binder.AddBinding(), which is the preferred method of creating bindings.
func (t *Trigger) AddBinder(b reactor.Binder, f reactor.BindingFunc, concurrent bool) {
    t.bindings = append(t.bindings, reactor.Binding{t, b, f, concurrent})
}

// AddReadCallback adds a callback that will be run when t is read using Value
func (t *Trigger) AddReadCallback(r reactor.ReadCallback) {
    t.readCallbacks = append(t.readCallbacks, r)
}

// AddWriteCallback adds a callback that will be run when t is written to 
// using SetValue.
func (t *Trigger) AddWriteCallback(w reactor.WriteCallback) {
    t.writeCallbacks = append(t.writeCallbacks, w)
}

// Get returns the value associated with key. If key does not exist in t, 
// Get returns nil (the zero value for interface{}).
//
// If the map in t has not been initalized by another key-level method or 
// SetValue, Get initializes the map.
//
// Any callbacks registered with AddKeyReadCallback will be called, passing
// the resulting key-value pair to the callback as a Pair struct. If the key 
// does not exist in t, the Pair is (kye, <nil>).
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

// GetCheck returns the value associated with key, and whether or not key 
// exists in t. If key does not exist in t, GetCheck returns (nil,false).
//
// If the map in t has not been initalized by another key-level method or 
// SetValue, GetCheck initializes the map.
//
// Any callbacks registered with AddKeyReadCallback will be called, passing
// the resulting key-value pair to the callback as a Pair struct. If the key 
// does not exist in t, the Pair is (kye, <nil>).
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

// Set updates the value associated with key to value. If key does not exist 
// in t, Set creates the key with a value of value.
//
// If the map in t has not been initalized by another key-level method or 
// SetValue, Set initializes the map.
//
// Any callbacks registered with AddKeyWriteCallback will be called, passing
// the previous and resulting key-value pair to the callback as a Pair struct. 
// If the key did not exist previously, the previous Pair is (key, <nil>).
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

// Delete removes key from t. If key does not exist in t, delete changes 
// nothing.
//
// If the map in t has not been initalized by another key-level method or 
// SetValue, Delete initializes the map.
//
// Any callbacks registered with AddKeyWriteCallback will be called, passing
// the previous and resulting key-value pair to the callback as a Pair struct.
// The resulting Pair is always (key, <nil>). If the key did not exist 
// previously, the previous Pair is (key, <nil>).
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

// Keys returns an unordered slice containing all of the keys created for t.
//
// Calls to Keys do not trigger any form of registred ReadCallbacks, general
// or key-level.
func (t *Trigger) Keys() []interface{} {
	t.Lock.Lock()
		keys := make([]interface{}, len(t.value))
		i := 0
		for k,_ := range t.value {
			keys[i] = k
			i++
		}
	t.Lock.Unlock()
	return keys
}

// Values returns an unordered slice containing the values associated with 
// existing keys in t. 
//
// The slice returned by this method is not guaranteed to be in the same order
// as that returned by Keys.
//
// Calls to Values do not trigger any form of registered ReadCallbacks, general
// or key-level.
func (t *Trigger) Values() []interface{} {
	t.Lock.Lock()
		values := make([]interface{}, len(t.value))
		i := 0
		for _,v := range t.value {
			values[i] = v
			i++
		}
	t.Lock.Unlock()
	return values
}

// Size returns the size of the underlying map of t.
func (t *Trigger) Size() int {
	return len(t.value)
}

// AddKeyReadCallback registers a ReadCallback that will be triggered any 
// key-level read events. The value passed to the callback will be a Pair 
// struct containing the key-value pair being read.
func (t *Trigger) AddKeyReadCallback(r reactor.ReadCallback) {
	t.keyReadCallbacks = append(t.keyReadCallbacks, r)
}

// AddKeyWriteCallback registers a WriteCallback that will be triggered any 
// key-level write events. The values passed to the callback will be Pair 
// structs containing the previous and resluting key-value pairs.
func (t *Trigger) AddKeyWriteCallback(w reactor.WriteCallback) {
	t.keyWriteCallbacks = append(t.keyWriteCallbacks, w)
}
